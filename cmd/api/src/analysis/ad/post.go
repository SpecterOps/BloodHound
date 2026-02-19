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
	"log/slog"

	"github.com/specterops/bloodhound/packages/go/analysis"
	adAnalysis "github.com/specterops/bloodhound/packages/go/analysis/ad"
	"github.com/specterops/bloodhound/packages/go/bhlog/attr"
	"github.com/specterops/bloodhound/packages/go/bhlog/measure"
	"github.com/specterops/bloodhound/packages/go/graphschema/ad"
	"github.com/specterops/bloodhound/packages/go/graphschema/azure"
	"github.com/specterops/dawgs/graph"
)

func Post(ctx context.Context, db graph.Database, adcsEnabled, citrixEnabled, ntlmEnabled bool, compositionCounter *analysis.CompositionCounter) (*analysis.AtomicPostProcessingStats, error) {
	defer measure.ContextMeasure(
		ctx,
		slog.LevelInfo,
		"Active Directory Post Processing",
		attr.Namespace("analysis"),
		attr.Function("Post"),
		attr.Scope("step"),
	)()

	aggregateStats := analysis.NewAtomicPostProcessingStats()

	if err := adAnalysis.FixWellKnownNodeTypes(ctx, db); err != nil {
		return &aggregateStats, err
	} else if err := adAnalysis.RunDomainAssociations(ctx, db); err != nil {
		return &aggregateStats, err
	} else if err := adAnalysis.LinkWellKnownNodes(ctx, db); err != nil {
		return &aggregateStats, err
	} else if deleteTransitEdgesStats, err := analysis.DeleteTransitEdges(ctx, db, graph.Kinds{ad.Entity, azure.Entity}, ad.PostProcessedRelationships()); err != nil {
		return &aggregateStats, err
	} else if localGroupData, err := adAnalysis.FetchLocalGroupData(ctx, db); err != nil {
		return &aggregateStats, err
	} else if dcSyncStats, err := adAnalysis.PostDCSync(ctx, db, localGroupData); err != nil {
		return &aggregateStats, err
	} else if protectAdminGroupsStats, err := adAnalysis.PostProtectAdminGroups(ctx, db); err != nil {
		return &aggregateStats, err
	} else if syncLAPSStats, err := adAnalysis.PostSyncLAPSPassword(ctx, db, localGroupData); err != nil {
		return &aggregateStats, err
	} else if hasTrustKeyStats, err := adAnalysis.PostHasTrustKeys(ctx, db); err != nil {
		return &aggregateStats, err
	} else if localGroupStats, err := adAnalysis.PostLocalGroups(ctx, db, localGroupData); err != nil {
		return &aggregateStats, err
	} else if canRDPStats, err := adAnalysis.PostCanRDP(ctx, db, localGroupData, true, citrixEnabled); err != nil {
		return &aggregateStats, err
	} else if adcsStats, adcsCache, err := adAnalysis.PostADCS(ctx, db, localGroupData, adcsEnabled); err != nil {
		return &aggregateStats, err
	} else if ownsStats, err := adAnalysis.PostOwnsAndWriteOwner(ctx, db, localGroupData); err != nil {
		return &aggregateStats, err
	} else if ntlmStats, err := adAnalysis.PostNTLM(ctx, db, localGroupData, adcsCache, ntlmEnabled, compositionCounter); err != nil {
		return &aggregateStats, err
	} else {
		aggregateStats.Merge(deleteTransitEdgesStats)
		aggregateStats.Merge(syncLAPSStats)
		aggregateStats.Merge(hasTrustKeyStats)
		aggregateStats.Merge(dcSyncStats)
		aggregateStats.Merge(protectAdminGroupsStats)
		aggregateStats.Merge(localGroupStats)
		aggregateStats.Merge(canRDPStats)
		aggregateStats.Merge(adcsStats)
		aggregateStats.Merge(ownsStats)
		aggregateStats.Merge(ntlmStats)

		return &aggregateStats, nil
	}
}
