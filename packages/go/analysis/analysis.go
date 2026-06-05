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

type analysisOperation func(context.Context, database.Database, graph.Database) (operationStatus, []error)

// name is an optional field used for logging pipeline steps,
// if an analysisPipelineStep does not have a related analysisStep.
// Otherwise the analysisStep's name is used.
// If doNotSkip is true, when the pipeline is run, that step is never skipped.
type analysisPipelineStep struct {
	name string
	// analysisStep links this pipeline step to a selectable analysis bit.
	// If analysisStep is zero, the step is not selectable and always runs.
	analysisStep                model.AnalysisStep
	operation                   analysisOperation
	countsTowardCompleteFailure bool // used to determine the pipeline failure status

	status operationStatus
}

func (s analysisPipelineStep) shouldRun(analysisSteps model.AnalysisSteps) bool {
	if s.analysisStep == 0 {
		return true
	}

	return analysisSteps.Has(s.analysisStep)
}

type analysisPipeline []analysisPipelineStep

// Err() should return ErrAnalysisFailed if:
// adFailed && azureFailed && agiFailed && dataQualityFailed.
// Err() should return ErrAnalysisPartiallyCompleted if:
// adFailed || azureFailed || agiFailed || agtPartiallyFailed || dataQualityFailed
//
// Pipeline failed if all countsTowardCompleteFailure steps fail.
// Pipeline partial success if not all countsTowardCompleteFailure fail or there was a partial failure.
func (s analysisPipeline) Err() error {
	var hadSuccessfulOrOnlyPartialFailedMainStep = false
	var hadAnyFailuresOrPartialFailures = false

	for _, pipelineStep := range s {
		if pipelineStep.countsTowardCompleteFailure && pipelineStep.status != operationStatusCompleteFailure {
			hadSuccessfulOrOnlyPartialFailedMainStep = true
		}

		if pipelineStep.status != operationStatusSuccess {
			hadAnyFailuresOrPartialFailures = true
		}
	}

	if !hadSuccessfulOrOnlyPartialFailedMainStep {
		return ErrAnalysisFailed
	} else if hadAnyFailuresOrPartialFailures {
		return ErrAnalysisPartiallyCompleted
	}

	return nil
}

func (s analysisPipeline) String() string {
	steps := make([]string, 0, len(s))

	for _, pipelineStep := range s {
		name, present := model.AnalysisStepName(pipelineStep.analysisStep)

		if !present && pipelineStep.name != "" {
			name = pipelineStep.name
		}

		steps = append(steps, fmt.Sprintf("%s:%s", name, pipelineStep.status))
	}

	return strings.Join(steps, ",")
}

// Modifies the state of the pipeline by setting success/failure/partial failure/skipped status for each step
func (s *analysisPipeline) dispatchAnalysisSteps(ctx context.Context, db database.Database, graphDB graph.Database, analysisSteps model.AnalysisSteps) []error {
	var collectedErrors []error

	for i := range *s {
		pipelineStep := &(*s)[i]

		if pipelineStep.shouldRun(analysisSteps) {
			if pipelineStep.operation != nil {
				status, errs := pipelineStep.operation(ctx, db, graphDB)
				pipelineStep.status = status
				collectedErrors = append(collectedErrors, errs...)
			}
		} else {
			pipelineStep.status = operationStatusSkipped
		}
	}

	return collectedErrors
}

func adPostProcessingOperation(ctx context.Context, db database.Database, graphDB graph.Database) (operationStatus, []error) {
	var collectedErrors []error

	if ntlmFlag, err := db.GetFlagByKey(ctx, appcfg.FeatureNTLMPostProcessing); err != nil {
		collectedErrors = append(collectedErrors, fmt.Errorf("error retrieving NTLM Post Processing feature flag: %w", err))
		return operationStatusCompleteFailure, collectedErrors
	} else if stats, err := adAnalysis.Post(ctx, graphDB, appcfg.GetCitrixRDPSupport(ctx, db), ntlmFlag.Enabled); err != nil {
		collectedErrors = append(collectedErrors, fmt.Errorf("error during ad post: %w", err))
		return operationStatusCompleteFailure, collectedErrors
	} else {
		stats.LogStats()
	}

	return operationStatusSuccess, collectedErrors
}

func azurePostProcessingOperation(ctx context.Context, db database.Database, graphDB graph.Database) (operationStatus, []error) {
	var collectedErrors []error

	if stats, err := azureAnalysis.Post(ctx, graphDB); err != nil {
		collectedErrors = append(collectedErrors, fmt.Errorf("error during azure post: %w", err))
		return operationStatusCompleteFailure, collectedErrors
	} else {
		stats.LogStats()
	}

	return operationStatusSuccess, collectedErrors
}

// TODO Cleanup tieringEnabled after Tiering GA
func taggingOperation(ctx context.Context, db database.Database, graphDB graph.Database) (operationStatus, []error) {
	var (
		collectedErrors []error
		status          operationStatus = operationStatusSuccess
		tieringEnabled                  = appcfg.GetTieringEnabled(ctx, db)
	)

	if errs := TagAssetGroupsAndTierZero(ctx, db, graphDB); len(errs) > 0 {
		for _, err := range errs {
			collectedErrors = append(collectedErrors, fmt.Errorf("tagging asset groups and tier zero failed: %w", err))
		}

		if ContainsOnlyCypherSelectorErrors(errs) {
			status = operationStatusPartialFailure
		}
	}

	if !tieringEnabled {
		if err := agi.RunAssetGroupIsolationCollections(ctx, db, graphDB, graphschema.GetNodeKindDisplayLabel); err != nil {
			collectedErrors = append(collectedErrors, fmt.Errorf("asset group isolation collection failed: %w", err))
			return operationStatusCompleteFailure, collectedErrors
		}
	}

	return status, collectedErrors
}

func dataQualityOperation(ctx context.Context, db database.Database, graphDB graph.Database) (operationStatus, []error) {
	var collectedErrors []error

	if err := dataquality.SaveDataQuality(ctx, db, graphDB); err != nil {
		collectedErrors = append(collectedErrors, fmt.Errorf("error saving data quality stat: %v", err))
		return operationStatusCompleteFailure, collectedErrors
	}

	return operationStatusSuccess, collectedErrors
}

const DataQuality = "data_quality"

// The definition of our analysis pipeline
func newPipeline() analysisPipeline {
	return analysisPipeline{
		{
			analysisStep:                model.AnalysisStepADPostProcessing(),
			countsTowardCompleteFailure: true,
			operation:                   adPostProcessingOperation,
		},
		{
			analysisStep:                model.AnalysisStepAzurePostProcessing(),
			countsTowardCompleteFailure: true,
			operation:                   azurePostProcessingOperation,
		},
		{
			analysisStep:                model.AnalysisStepTagging(),
			countsTowardCompleteFailure: false,
			operation:                   taggingOperation,
		},
		{
			name:                        DataQuality,
			countsTowardCompleteFailure: true,
			operation:                   dataQualityOperation,
		},
	}
}

func RunAnalysisOperations(ctx context.Context, db database.Database, graphDB graph.Database, _ config.Configuration, analysisSteps model.AnalysisSteps) error {
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
	)

	collectedErrors := pipeline.dispatchAnalysisSteps(ctx, db, graphDB, analysisSteps)

	slog.InfoContext(
		ctx,
		"Finished Running Analysis Operations",
		slog.String("namespace", "analysis"),
		slog.String("fn", "RunAnalysisOperations"),
		slog.String("pipeline_status", pipeline.String()),
	)

	if len(collectedErrors) > 0 {
		for _, err := range collectedErrors {
			slog.ErrorContext(ctx, "Analysis error encountered", attr.Error(err))
		}
	}

	if err := pipeline.Err(); err != nil {
		return err
	}

	return nil
}
