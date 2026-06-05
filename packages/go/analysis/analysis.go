// Copyright 2023 Specter Ops, Inc.
//
// Licensed under the Apache License, Version 2.0
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

package analysis

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/specterops/bloodhound/cmd/api/src/config"
	"github.com/specterops/bloodhound/cmd/api/src/database"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/cmd/api/src/model/appcfg"
	"github.com/specterops/bloodhound/cmd/api/src/services/agi"
	"github.com/specterops/bloodhound/cmd/api/src/services/dataquality"
	adAnalysis "github.com/specterops/bloodhound/packages/go/analysis/ad"
	azureAnalysis "github.com/specterops/bloodhound/packages/go/analysis/azure"
	"github.com/specterops/bloodhound/packages/go/bhlog/attr"
	"github.com/specterops/bloodhound/packages/go/graphschema"
	"github.com/specterops/bloodhound/packages/go/graphschema/ad"
	"github.com/specterops/bloodhound/packages/go/graphschema/azure"
	"github.com/specterops/bloodhound/packages/go/graphschema/common"
	"github.com/specterops/dawgs/graph"
	"github.com/specterops/dawgs/ops"
	"github.com/specterops/dawgs/query"
)

const (
	NodeKindUnknown = "Unknown"
)

func nodeByIndexedKindProperty(property, value string, kind graph.Kind) graph.Criteria {
	return query.And(
		query.Equals(query.NodeProperty(property), value),
		query.Kind(query.Node(), kind),
	)
}

func openGraphNodeByIndexedKindProperty(property, value string) graph.Criteria {
	return query.And(
		query.Equals(query.NodeProperty(property), value),
		query.Not(query.Kind(query.Node(), ad.Entity, azure.Entity)),
	)
}

// FetchNodeByObjectID will search for a node given its object ID. This function may run more than one query to ensure
// that label indexes are correctly exercised. The turnaround time of multiple indexed queries is an order of magnitude
// faster in larger environments than allowing Neo4j to perform a table scan of unindexed node properties.
func FetchNodeByObjectID(tx graph.Transaction, objectID string) (*graph.Node, error) {
	if node, err := tx.Nodes().Filter(nodeByIndexedKindProperty(common.ObjectID.String(), objectID, ad.Entity)).First(); err != nil {
		if !graph.IsErrNotFound(err) {
			return nil, err
		}
	} else {
		return node, nil
	}

	return tx.Nodes().Filter(nodeByIndexedKindProperty(common.ObjectID.String(), objectID, azure.Entity)).First()
}

func FetchNodeByObjectIDIncludeOpenGraph(tx graph.Transaction, objectID string) (*graph.Node, error) {
	if node, err := FetchNodeByObjectID(tx, objectID); err != nil {
		if !graph.IsErrNotFound(err) {
			return nil, err
		}
	} else {
		return node, nil
	}
	return tx.Nodes().Filter(openGraphNodeByIndexedKindProperty(common.ObjectID.String(), objectID)).First()
}

func FetchEdgeByStartAndEnd(ctx context.Context, graphDB graph.Database, start, end graph.ID, edgeKind graph.Kind) (*graph.Relationship, error) {
	var result *graph.Relationship
	return result, graphDB.ReadTransaction(ctx, func(tx graph.Transaction) error {
		if rel, err := tx.Relationships().Filter(query.And(
			query.Equals(query.StartID(), start),
			query.Equals(query.EndID(), end),
			query.Kind(query.Relationship(), edgeKind),
		)).First(); err != nil {
			return err
		} else {
			result = rel
			return nil
		}
	})
}

func ExpandGroupMembershipPaths(tx graph.Transaction, candidates graph.NodeSet) (graph.PathSet, error) {
	groupMemberPaths := graph.NewPathSet()

	for _, candidate := range candidates {
		if candidate.Kinds.ContainsOneOf(ad.Group) {
			if membershipPaths, err := ops.TraversePaths(tx, ops.TraversalPlan{
				Root:      candidate,
				Direction: graph.DirectionInbound,
				BranchQuery: func() graph.Criteria {
					return query.Kind(query.Relationship(), ad.MemberOf)
				},
			}); err != nil {
				return nil, err
			} else {
				groupMemberPaths.AddPathSet(membershipPaths)
			}
		}
	}

	return groupMemberPaths, nil
}

var (
	ErrAnalysisFailed             = errors.New("analysis failed")
	ErrAnalysisPartiallyCompleted = errors.New("analysis partially completed")
)

type operationStatus string

const (
	operationStatusPartialFailure  = "partial_failure"
	operationStatusCompleteFailure = "complete_failure"
	operationStatusSuccess         = "success"
	operationStatusSkipped         = "skipped"
)

type analysisErrors struct {
	adPost      bool
	azurePost   bool
	agi         bool
	agtPartial  bool
	dataQuality bool
}

func (s *analysisErrors) evaluateErrors() error {
	if s.adPost && s.azurePost && s.agi && s.dataQuality {
		return ErrAnalysisFailed
	} else if s.adPost || s.azurePost || s.agi || s.agtPartial || s.dataQuality {
		return ErrAnalysisPartiallyCompleted
	}

	return nil
}

type analysisOperation func(analysisPipelineRun) (operationStatus, []error)

// name is an optional field used for logging pipeline steps
// if an analysisPipelineStep does not have a related analysisStep.
// Otherwise the analysisStep's name is used.
type analysisPipelineStep struct {
	name string
	// analysisStep links this pipeline step to a selectable analysis bit.
	// If analysisStep is zero, the step is not selectable and always runs.
	analysisStep model.AnalysisStep
	operation    analysisOperation
}

func (s analysisPipelineStep) shouldRun(analysisSteps model.AnalysisSteps) bool {
	if s.analysisStep == 0 {
		return true
	}

	return analysisSteps.Has(s.analysisStep)
}

func (s analysisPipelineStep) String() string {
	name, present := model.AnalysisStepName(s.analysisStep)

	if !present && s.name != "" {
		name = s.name
	}

	return name
}

type analysisPipeline []analysisPipelineStep

func (s analysisPipeline) String() string {
	steps := make([]string, 0, len(s))

	for _, pipelineStep := range s {
		steps = append(steps, pipelineStep.String())
	}

	return strings.Join(steps, ",")
}

type analysisPipelineRun struct {
	ctx           context.Context
	db            database.Database
	graphDB       graph.Database
	analysisSteps model.AnalysisSteps
	analysisErrs  *analysisErrors
}

type analysisPipelineStepResult struct {
	name   string
	status operationStatus
	errors []error
}

func (s analysisPipelineStepResult) String() string {
	return fmt.Sprintf("%s:%s", s.name, s.status)
}

type analysisPipelineResult []analysisPipelineStepResult

func (s analysisPipelineResult) Errors() []error {
	var collectedErrors []error

	for _, pipelineStepResult := range s {
		collectedErrors = append(collectedErrors, pipelineStepResult.errors...)
	}

	return collectedErrors
}

func (s analysisPipelineResult) String() string {
	steps := make([]string, 0, len(s))

	for _, pipelineStepResult := range s {
		steps = append(steps, pipelineStepResult.String())
	}

	return strings.Join(steps, ",")
}

func (s analysisPipeline) dispatchAnalysisSteps(run analysisPipelineRun) analysisPipelineResult {
	result := make(analysisPipelineResult, 0, len(s))

	for _, pipelineStep := range s {
		pipelineStepResult := analysisPipelineStepResult{
			name:   pipelineStep.String(),
			status: operationStatusSkipped,
		}

		if pipelineStep.shouldRun(run.analysisSteps) {
			if pipelineStep.operation != nil {
				status, errs := pipelineStep.operation(run)
				pipelineStepResult.status = status
				pipelineStepResult.errors = errs
			}
		}

		result = append(result, pipelineStepResult)
	}

	return result
}

func adPostProcessingOperation(run analysisPipelineRun) (operationStatus, []error) {
	var collectedErrors []error

	if ntlmFlag, err := run.db.GetFlagByKey(run.ctx, appcfg.FeatureNTLMPostProcessing); err != nil {
		collectedErrors = append(collectedErrors, fmt.Errorf("error retrieving NTLM Post Processing feature flag: %w", err))
		run.analysisErrs.adPost = true
		return operationStatusCompleteFailure, collectedErrors
	} else if stats, err := adAnalysis.Post(run.ctx, run.graphDB, appcfg.GetCitrixRDPSupport(run.ctx, run.db), ntlmFlag.Enabled); err != nil {
		collectedErrors = append(collectedErrors, fmt.Errorf("error during ad post: %w", err))
		run.analysisErrs.adPost = true
		return operationStatusCompleteFailure, collectedErrors
	} else {
		stats.LogStats()
	}

	return operationStatusSuccess, collectedErrors
}

func azurePostProcessingOperation(run analysisPipelineRun) (operationStatus, []error) {
	var collectedErrors []error

	if stats, err := azureAnalysis.Post(run.ctx, run.graphDB); err != nil {
		collectedErrors = append(collectedErrors, fmt.Errorf("error during azure post: %w", err))
		run.analysisErrs.azurePost = true
		return operationStatusCompleteFailure, collectedErrors
	} else {
		stats.LogStats()
	}

	return operationStatusSuccess, collectedErrors
}

// TODO Cleanup tieringEnabled after Tiering GA
func taggingOperation(run analysisPipelineRun) (operationStatus, []error) {
	var (
		collectedErrors []error
		status          operationStatus = operationStatusSuccess
		tieringEnabled                  = appcfg.GetTieringEnabled(run.ctx, run.db)
	)

	if errs := TagAssetGroupsAndTierZero(run.ctx, run.db, run.graphDB); len(errs) > 0 {
		for _, err := range errs {
			collectedErrors = append(collectedErrors, fmt.Errorf("tagging asset groups and tier zero failed: %w", err))
		}

		if ContainsOnlyCypherSelectorErrors(errs) {
			run.analysisErrs.agtPartial = true
			status = operationStatusPartialFailure
		}
	}

	if !tieringEnabled {
		if err := agi.RunAssetGroupIsolationCollections(run.ctx, run.db, run.graphDB, graphschema.GetNodeKindDisplayLabel); err != nil {
			collectedErrors = append(collectedErrors, fmt.Errorf("asset group isolation collection failed: %w", err))
			run.analysisErrs.agi = true
			return operationStatusCompleteFailure, collectedErrors
		}
	}

	return status, collectedErrors
}

func dataQualityOperation(run analysisPipelineRun) (operationStatus, []error) {
	var collectedErrors []error

	if err := dataquality.SaveDataQuality(run.ctx, run.db, run.graphDB); err != nil {
		collectedErrors = append(collectedErrors, fmt.Errorf("error saving data quality stat: %v", err))
		run.analysisErrs.dataQuality = true
		return operationStatusCompleteFailure, collectedErrors
	}

	return operationStatusSuccess, collectedErrors
}

const DataQuality = "data_quality"

// The definition of our analysis pipeline
func newPipeline() analysisPipeline {
	return analysisPipeline{
		{
			analysisStep: model.AnalysisStepADPostProcessing(),
			operation:    adPostProcessingOperation,
		},
		{
			analysisStep: model.AnalysisStepAzurePostProcessing(),
			operation:    azurePostProcessingOperation,
		},
		{
			analysisStep: model.AnalysisStepTagging(),
			operation:    taggingOperation,
		},
		{
			name:      DataQuality,
			operation: dataQualityOperation,
		},
	}
}

func RunAnalysisOperations(ctx context.Context, db database.Database, graphDB graph.Database, _ config.Configuration, analysisSteps model.AnalysisSteps) error {
	analysisErrs := &analysisErrors{}

	if !appcfg.GetVariableAnalysisModeEnabled(ctx, db) {
		analysisSteps = model.AnalysisStepsFull()
	}

	pipeline := newPipeline()

	slog.InfoContext(
		ctx,
		"Running Analysis Operations",
		slog.String("namespace", "analysis"),
		slog.String("fn", "RunAnalysisOperations"),
		slog.Int("analysis_steps_bits", analysisSteps.Bits()),
		slog.String("pipeline_steps", pipeline.String()),
	)

	pipelineResult := pipeline.dispatchAnalysisSteps(analysisPipelineRun{
		ctx:           ctx,
		db:            db,
		graphDB:       graphDB,
		analysisSteps: analysisSteps,
		analysisErrs:  analysisErrs,
	})

	slog.InfoContext(
		ctx,
		"Finished Running Analysis Operations",
		slog.String("namespace", "analysis"),
		slog.String("fn", "RunAnalysisOperations"),
		slog.String("pipeline_status", pipelineResult.String()),
	)

	collectedErrors := pipelineResult.Errors()
	if len(collectedErrors) > 0 {
		for _, err := range collectedErrors {
			slog.ErrorContext(ctx, "Analysis error encountered", attr.Error(err))
		}
	}

	if err := analysisErrs.evaluateErrors(); err != nil {
		return err
	}

	return nil
}
