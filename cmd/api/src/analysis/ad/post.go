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

	"github.com/specterops/bloodhound/packages/go/analysis"
	adAnalysis "github.com/specterops/bloodhound/packages/go/analysis/ad"
	"github.com/specterops/bloodhound/packages/go/graphschema/ad"
	"github.com/specterops/bloodhound/packages/go/graphschema/azure"
	"github.com/specterops/dawgs/graph"
)

func Post(ctx context.Context, db graph.Database, adcsEnabled, citrixEnabled, ntlmEnabled bool, compositionCounter *analysis.CompositionCounter) (*analysis.AtomicPostProcessingStats, error) {
	aggregateStats := analysis.NewAtomicPostProcessingStats()
	if stats, err := analysis.DeleteTransitEdges(ctx, db, graph.Kinds{ad.Entity, azure.Entity}, adAnalysis.PostProcessedRelationships()...); err != nil {
		return &aggregateStats, err
	} else if groupExpansions, err := adAnalysis.ExpandAllRDPLocalGroups(ctx, db); err != nil {
		return &aggregateStats, err
	} else if dcSyncStats, err := adAnalysis.PostDCSync(ctx, db, groupExpansions); err != nil {
		return &aggregateStats, err
	} else if syncLAPSStats, err := adAnalysis.PostSyncLAPSPassword(ctx, db, groupExpansions); err != nil {
		return &aggregateStats, err
	} else if hasTrustKeyStats, err := adAnalysis.PostHasTrustKeys(ctx, db); err != nil {
		return &aggregateStats, err
	} else if localGroupStats, err := adAnalysis.PostLocalGroups(ctx, db, groupExpansions, false, citrixEnabled); err != nil {
		return &aggregateStats, err
	} else if adcsStats, adcsCache, err := adAnalysis.PostADCS(ctx, db, groupExpansions, adcsEnabled); err != nil {
		return &aggregateStats, err
	} else if ownsStats, err := adAnalysis.PostOwnsAndWriteOwner(ctx, db, groupExpansions); err != nil {
		return &aggregateStats, err
	} else if ntlmStats, err := adAnalysis.PostNTLM(ctx, db, groupExpansions, adcsCache, ntlmEnabled, compositionCounter); err != nil {
		return &aggregateStats, err
	} else {
		aggregateStats.Merge(stats)
		aggregateStats.Merge(syncLAPSStats)
		aggregateStats.Merge(hasTrustKeyStats)
		aggregateStats.Merge(dcSyncStats)
		aggregateStats.Merge(localGroupStats)
		aggregateStats.Merge(adcsStats)
		aggregateStats.Merge(ownsStats)
		aggregateStats.Merge(ntlmStats)
		return &aggregateStats, nil
	}
}
