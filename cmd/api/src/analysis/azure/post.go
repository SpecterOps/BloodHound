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

	"github.com/specterops/bloodhound/analysis"
	azureAnalysis "github.com/specterops/bloodhound/analysis/azure"
	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/graphschema/azure"
)

func Post(ctx context.Context, db graph.Database) (*analysis.AtomicPostProcessingStats, error) {
	aggregateStats := analysis.NewAtomicPostProcessingStats()
	if stats, err := analysis.DeleteTransitEdges(ctx, db, azure.Entity, azure.Entity, azureAnalysis.AzurePostProcessedRelationships()...); err != nil {
		return &aggregateStats, err
	} else if userRoleStats, err := azureAnalysis.UserRoleAssignments(ctx, db); err != nil {
		return &aggregateStats, err
	} else if addSecretStats, err := azureAnalysis.AddSecret(ctx, db); err != nil {
		return &aggregateStats, err
	} else if executeCommandStats, err := azureAnalysis.ExecuteCommand(ctx, db); err != nil {
		return &aggregateStats, err
	} else if appRoleAssignmentStats, err := azureAnalysis.AppRoleAssignments(ctx, db); err != nil {
		return &aggregateStats, err
	} else {
		aggregateStats.Merge(stats)
		aggregateStats.Merge(userRoleStats)
		aggregateStats.Merge(addSecretStats)
		aggregateStats.Merge(executeCommandStats)
		aggregateStats.Merge(appRoleAssignmentStats)
		return &aggregateStats, nil
	}
}
