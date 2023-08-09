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

package ad

import (
	"context"
	"fmt"

	"github.com/specterops/bloodhound/analysis"
	adAnalysis "github.com/specterops/bloodhound/analysis/ad"
	"github.com/specterops/bloodhound/analysis/impact"
	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/dawgs/util/channels"
	"github.com/specterops/bloodhound/graphschema/ad"
	"github.com/specterops/bloodhound/log"
	"github.com/specterops/bloodhound/src/model/appcfg"
)

func AnalyzedRelationships(localGroupAnalysisFlag appcfg.FeatureFlag) []graph.Kind {
	if localGroupAnalysisFlag.Enabled {
		return []graph.Kind{
			ad.DCSync,
			ad.SyncLAPSPassword,
			ad.CanRDP,
			ad.AdminTo,
			ad.CanPSRemote,
			ad.ExecuteDCOM,
		}
	}

	return []graph.Kind{
		ad.SyncLAPSPassword,
		ad.DCSync,
	}
}

func AnalyzeLocalGroups(ctx context.Context, db graph.Database) (*analysis.AtomicAnalysisStats, error) {
	var (
		adminGroupSuffix    = "-544"
		psRemoteGroupSuffix = "-580"
		dcomGroupSuffix     = "-562"
	)

	if localGroupExpansions, err := adAnalysis.ExpandAllRDPLocalGroups(ctx, db); err != nil {
		return &analysis.AtomicAnalysisStats{}, err
	} else if computers, err := adAnalysis.FetchComputers(ctx, db); err != nil {
		return &analysis.AtomicAnalysisStats{}, err
	} else {
		var (
			threadSafeLocalGroupExpansions = impact.NewThreadSafeAggregator(localGroupExpansions)
			operation                      = analysis.NewAnalysisRelationshipOperation(ctx, db, "LocalGroup Analysis")
		)

		for idx, computer := range computers.ToArray() {
			computerID := graph.ID(computer)

			if idx > 0 && idx%10000 == 0 {
				log.Infof("Analyzed %d active directory computers", idx)
			}

			if err := operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreateAnalysisRelationshipJob) error {
				if entities, err := adAnalysis.FetchLocalGroupBitmapForComputer(tx, computerID, dcomGroupSuffix); err != nil {
					return err
				} else {
					for _, admin := range entities.Slice() {
						nextJob := analysis.CreateAnalysisRelationshipJob{
							FromID: graph.ID(admin),
							ToID:   computerID,
							Kind:   ad.ExecuteDCOM,
						}

						if !channels.Submit(ctx, outC, nextJob) {
							return nil
						}
					}

					return nil
				}
			}); err != nil {
				return &analysis.AtomicAnalysisStats{}, fmt.Errorf("failed submitting reader for operation involving computer %d: %w", computerID, err)
			}

			if err := operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreateAnalysisRelationshipJob) error {
				if entities, err := adAnalysis.FetchLocalGroupBitmapForComputer(tx, computerID, psRemoteGroupSuffix); err != nil {
					return err
				} else {
					for _, admin := range entities.Slice() {
						nextJob := analysis.CreateAnalysisRelationshipJob{
							FromID: graph.ID(admin),
							ToID:   computerID,
							Kind:   ad.CanPSRemote,
						}

						if !channels.Submit(ctx, outC, nextJob) {
							return nil
						}
					}

					return nil
				}
			}); err != nil {
				return &analysis.AtomicAnalysisStats{}, fmt.Errorf("failed submitting reader for operation involving computer %d: %w", computerID, err)
			}

			if err := operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreateAnalysisRelationshipJob) error {
				if entities, err := adAnalysis.FetchLocalGroupBitmapForComputer(tx, computerID, adminGroupSuffix); err != nil {
					return err
				} else {
					for _, admin := range entities.Slice() {
						nextJob := analysis.CreateAnalysisRelationshipJob{
							FromID: graph.ID(admin),
							ToID:   computerID,
							Kind:   ad.AdminTo,
						}

						if !channels.Submit(ctx, outC, nextJob) {
							return nil
						}
					}

					return nil
				}
			}); err != nil {
				return &analysis.AtomicAnalysisStats{}, fmt.Errorf("failed submitting reader for operation involving computer %d: %w", computerID, err)
			}

			if err := operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreateAnalysisRelationshipJob) error {
				if entities, err := adAnalysis.FetchRDPEntityBitmapForComputerWithUnenforcedURA(tx, computerID, threadSafeLocalGroupExpansions); err != nil {
					return err
				} else {
					for _, rdp := range entities.Slice() {
						nextJob := analysis.CreateAnalysisRelationshipJob{
							FromID: graph.ID(rdp),
							ToID:   computerID,
							Kind:   ad.CanRDP,
						}

						if !channels.Submit(ctx, outC, nextJob) {
							return nil
						}
					}
				}

				return nil
			}); err != nil {
				return &analysis.AtomicAnalysisStats{}, fmt.Errorf("failed submitting reader for operation involving computer %d: %w", computerID, err)
			}
		}

		log.Infof("Finished analyzing %d active directory computers", computers.GetCardinality())
		return &operation.Stats, operation.Done()
	}
}

func Analyze(ctx context.Context, db graph.Database) (*analysis.AtomicAnalysisStats, error) {
	aggregateStats := analysis.NewAtomicAnalysisStats()
	if stats, err := analysis.DeleteTransitEdges(ctx, db, ad.Entity, ad.Entity, adAnalysis.AnalyzedRelationships()...); err != nil {
		return &aggregateStats, err
	} else if dcSyncStats, err := adAnalysis.AnalyzeDCSync(ctx, db); err != nil {
		return &aggregateStats, err
	} else if syncLAPSStats, err := adAnalysis.AnalyzeSyncLAPSPassword(ctx, db); err != nil {
		return &aggregateStats, err
	} else if localGroupStats, err := AnalyzeLocalGroups(ctx, db); err != nil {
		return &aggregateStats, err
	} else {
		aggregateStats.Merge(stats)
		aggregateStats.Merge(syncLAPSStats)
		aggregateStats.Merge(dcSyncStats)
		aggregateStats.Merge(localGroupStats)
		return &aggregateStats, nil
	}
}
