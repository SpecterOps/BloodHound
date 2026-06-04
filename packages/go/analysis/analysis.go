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

func runAnalysisStepOperation(operation func()) {
	if operation != nil {
		operation()
	}
}

func dispatchAnalysisSteps(analysisSteps model.AnalysisSteps, pipeline analysisPipeline) {
	for _, pipelineStep := range pipeline {
		if analysisSteps.Has(pipelineStep.analysisStep) || pipelineStep.alwaysRun {
			runAnalysisStepOperation(pipelineStep.operation)
		}
	}
}

type analysisPipelineStep struct {
	analysisStep model.AnalysisStep
	operation    func()
	name         string
	alwaysRun    bool
}

type analysisPipeline []analysisPipelineStep

func (s analysisPipeline) String() string {
	stepNames := make([]string, 0, len(s))

	for _, pipelineStep := range s {
		name, present := model.AnalysisStepName(pipelineStep.analysisStep)

		if !present && pipelineStep.name != "" {
			stepNames = append(stepNames, pipelineStep.name)
		} else {
			stepNames = append(stepNames, name)
		}

	}

	return strings.Join(stepNames, ",")
}

// TODO Cleanup tieringEnabled after Tiering GA
func RunAnalysisOperations(ctx context.Context, db database.Database, graphDB graph.Database, _ config.Configuration, analysisSteps model.AnalysisSteps) error {
	var (
		collectedErrors []error
		tieringEnabled  = appcfg.GetTieringEnabled(ctx, db)
	)

	if !appcfg.GetVariableAnalysisModeEnabled(ctx, db) {
		analysisSteps = model.AnalysisStepsFull()
	}

	var (
		adFailed           = false
		azureFailed        = false
		agiFailed          = false
		agtPartiallyFailed = false
		dataQualityFailed  = false
	)

	pipeline := analysisPipeline{
		{
			analysisStep: model.AnalysisStepADPostProcessing(),
			operation: func() {
				if ntlmFlag, err := db.GetFlagByKey(ctx, appcfg.FeatureNTLMPostProcessing); err != nil {
					collectedErrors = append(collectedErrors, fmt.Errorf("error retrieving NTLM Post Processing feature flag: %w", err))
				} else if stats, err := adAnalysis.Post(ctx, graphDB, appcfg.GetCitrixRDPSupport(ctx, db), ntlmFlag.Enabled); err != nil {
					collectedErrors = append(collectedErrors, fmt.Errorf("error during ad post: %w", err))
					adFailed = true
				} else {
					stats.LogStats()
				}
			},
		},
		{
			analysisStep: model.AnalysisStepAzurePostProcessing(),
			operation: func() {
				if stats, err := azureAnalysis.Post(ctx, graphDB); err != nil {
					collectedErrors = append(collectedErrors, fmt.Errorf("error during azure post: %w", err))
					azureFailed = true
				} else {
					stats.LogStats()
				}
			},
		},
		{
			analysisStep: model.AnalysisStepAzurePostProcessing(),
			operation: func() {
				if errs := TagAssetGroupsAndTierZero(ctx, db, graphDB); len(errs) > 0 {
					for _, err := range errs {
						collectedErrors = append(collectedErrors, fmt.Errorf("tagging asset groups and tier zero failed: %w", err))
					}

					if ContainsOnlyCypherSelectorErrors(errs) {
						agtPartiallyFailed = true
					}
				}

				if !tieringEnabled {
					if err := agi.RunAssetGroupIsolationCollections(ctx, db, graphDB, graphschema.GetNodeKindDisplayLabel); err != nil {
						collectedErrors = append(collectedErrors, fmt.Errorf("asset group isolation collection failed: %w", err))
						agiFailed = true
					}
				}
			},
		},
		{
			name:      "data_quality", // TODO: move to constant
			alwaysRun: true,
			operation: func() {
				if err := dataquality.SaveDataQuality(ctx, db, graphDB); err != nil {
					collectedErrors = append(collectedErrors, fmt.Errorf("error saving data quality stat: %v", err))
					dataQualityFailed = true
				}
			},
		},
	}

	slog.InfoContext(
		ctx,
		"Running Analysis Operations",
		slog.String("namespace", "analysis"),
		slog.String("fn", "RunAnalysisOperations"),
		slog.String("pipeline_steps", pipeline.String()),
	)

	dispatchAnalysisSteps(analysisSteps, pipeline)

	if len(collectedErrors) > 0 {
		for _, err := range collectedErrors {
			slog.ErrorContext(ctx, "Analysis error encountered", attr.Error(err))
		}
	}

	if adFailed && azureFailed && agiFailed && dataQualityFailed {
		return ErrAnalysisFailed
	} else if adFailed || azureFailed || agiFailed || agtPartiallyFailed || dataQualityFailed {
		return ErrAnalysisPartiallyCompleted
	}

	return nil
}
