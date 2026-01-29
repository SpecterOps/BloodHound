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

package azure

import (
	"context"
	"log/slog"

	"github.com/specterops/bloodhound/packages/go/analysis"
	azureAnalysis "github.com/specterops/bloodhound/packages/go/analysis/azure"
	"github.com/specterops/bloodhound/packages/go/analysis/hybrid"
	"github.com/specterops/bloodhound/packages/go/bhlog/attr"
	"github.com/specterops/bloodhound/packages/go/graphschema/ad"
	"github.com/specterops/bloodhound/packages/go/graphschema/azure"
	"github.com/specterops/dawgs/graph"
)

func Post(ctx context.Context, db graph.Database) (*analysis.AtomicPostProcessingStats, error) {
	aggregateStats := analysis.NewAtomicPostProcessingStats()
	if err := azureAnalysis.FixManagementGroupNames(ctx, db); err != nil {
		slog.WarnContext(ctx, "Error fixing management group names", attr.Error(err))
	}
	if stats, err := analysis.DeleteTransitEdges(ctx, db, "Delete Azure Post-Processed Relationships", graph.Kinds{ad.Entity, azure.Entity}, azure.PostProcessedRelationships()...); err != nil {
		return &aggregateStats, err
	} else if userRoleStats, err := azureAnalysis.UserRoleAssignments(ctx, db); err != nil {
		return &aggregateStats, err
	} else if executeCommandStats, err := azureAnalysis.ExecuteCommand(ctx, db); err != nil {
		return &aggregateStats, err
	} else if appRoleAssignmentStats, err := azureAnalysis.AppRoleAssignments(ctx, db); err != nil {
		return &aggregateStats, err
	} else if hybridStats, err := hybrid.PostHybrid(ctx, db); err != nil {
		return &aggregateStats, err
	} else if pimRolesStats, err := azureAnalysis.CreateAZRoleApproverEdge(ctx, db); err != nil {
		return &aggregateStats, err
	} else {
		aggregateStats.Merge(stats)
		aggregateStats.Merge(userRoleStats)
		aggregateStats.Merge(executeCommandStats)
		aggregateStats.Merge(appRoleAssignmentStats)
		aggregateStats.Merge(hybridStats)
		aggregateStats.Merge(pimRolesStats)
		return &aggregateStats, nil
	}
}
