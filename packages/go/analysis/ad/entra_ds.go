// Copyright 2026 Specter Ops, Inc.
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

	"github.com/specterops/bloodhound/packages/go/graphschema/azure"
	"github.com/specterops/dawgs/graph"
	"github.com/specterops/dawgs/ops"
	"github.com/specterops/dawgs/query"
)

// GetAddEntraDSGroupMemberEdgeComposition reconstructs the paths that compose an AddEntraDSGroupMember edge. The
// edge's start node is the AZUser and its end node is the on-prem Group it can add a member to. The composition is:
//
//	p1 = (azUser)-[:SyncedToEntraDSUser]->(:User)
//	p2 = (azUser)-[:AZOwns|AZAddMembers]->(azGroup:AZGroup)
//	p3 = (azGroup)-[:SyncedToEntraDSGroup]->(targetGroup)
func GetAddEntraDSGroupMemberEdgeComposition(ctx context.Context, db graph.Database, edge *graph.Relationship) (graph.PathSet, error) {
	finalPaths := graph.NewPathSet()

	if err := db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		azUser, targetGroup, err := ops.FetchRelationshipNodes(tx, edge)
		if err != nil {
			return err
		}

		// p1: the SyncedToEntraDSUser edge originating at the AZUser
		syncedUserPaths, err := ops.FetchPathSet(tx.Relationships().Filter(query.And(
			query.Equals(query.StartID(), azUser.ID),
			query.Kind(query.Relationship(), azure.SyncedToEntraDSUser),
		)))
		if err != nil {
			return err
		}

		// p3: the SyncedToEntraDSGroup edges terminating at the target on-prem Group; the start node of each is an AZGroup
		syncedGroupPaths, err := ops.FetchPathSet(tx.Relationships().Filter(query.And(
			query.Equals(query.EndID(), targetGroup.ID),
			query.Kind(query.Relationship(), azure.SyncedToEntraDSGroup),
		)))
		if err != nil {
			return err
		}

		// Without both the AZUser being synced and the target group being synced from an AZGroup there is no valid composition
		if syncedUserPaths.Len() == 0 || syncedGroupPaths.Len() == 0 {
			return nil
		}

		for _, syncedGroupPath := range syncedGroupPaths {
			azGroup := syncedGroupPath.Root()

			// p2: the AZOwns / AZAddMembers control edge(s) from the AZUser to this AZGroup
			controlPaths, err := ops.FetchPathSet(tx.Relationships().Filter(query.And(
				query.Equals(query.StartID(), azUser.ID),
				query.Equals(query.EndID(), azGroup.ID),
				query.KindIn(query.Relationship(), azure.AddMembers, azure.Owns),
			)))
			if err != nil {
				return err
			} else if controlPaths.Len() == 0 {
				continue
			}

			finalPaths.AddPathSet(controlPaths)
			finalPaths.AddPath(syncedGroupPath)
		}

		// Only include the user sync path (p1) if at least one complete composition was found
		if finalPaths.Len() > 0 {
			finalPaths.AddPathSet(syncedUserPaths)
		}

		return nil
	}); err != nil {
		return graph.NewPathSet(), err
	}

	return finalPaths, nil
}
