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

package hybrid

import (
	"context"
	"errors"
	"strings"

	graphschemaAD "github.com/specterops/bloodhound/packages/go/graphschema/ad"
	"github.com/specterops/bloodhound/packages/go/graphschema/azure"
	"github.com/specterops/dawgs/graph"
	"github.com/specterops/dawgs/ops"
	"github.com/specterops/dawgs/query"
)

// GetAddEntraDSGroupMemberEdgeComposition reconstructs the paths that compose an AddEntraDSGroupMember edge. The
// edge's start node is the AZUser and its end node is the AD Group whose membership it can modify. The composition is:
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

		syncedUserPaths, err := ops.FetchPathSet(tx.Relationships().Filter(query.And(
			query.Equals(query.StartID(), azUser.ID),
			query.Kind(query.Relationship(), azure.SyncedToEntraDSUser),
		)))
		if err != nil {
			return err
		}

		syncedGroupPaths, err := ops.FetchPathSet(tx.Relationships().Filter(query.And(
			query.Equals(query.EndID(), targetGroup.ID),
			query.Kind(query.Relationship(), azure.SyncedToEntraDSGroup),
		)))
		if err != nil {
			return err
		}

		if syncedUserPaths.Len() == 0 || syncedGroupPaths.Len() == 0 {
			return nil
		}

		for _, syncedGroupPath := range syncedGroupPaths {
			azGroup := syncedGroupPath.Root()
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

		if finalPaths.Len() > 0 {
			finalPaths.AddPathSet(syncedUserPaths)
		}
		return nil
	}); err != nil {
		return graph.NewPathSet(), err
	}

	return finalPaths, nil
}

// GetManageEntraDSSyncEdgeComposition reconstructs the relationships that compose a ManageEntraDSSync edge:
//
//	p1 = (source)-[:AZManageEntraDS]->(domainService:AZEntraDS)
//	p2 = (domainService)-[:EntraDSFor]->(domain:Domain)
//	p3 = (domain)-[:Contains*1..]->(targetGroup:Group)
func GetManageEntraDSSyncEdgeComposition(ctx context.Context, db graph.Database, edge *graph.Relationship) (graph.PathSet, error) {
	finalPaths := graph.NewPathSet()

	if err := db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		source, targetGroup, err := ops.FetchRelationshipNodes(tx, edge)
		if err != nil {
			return err
		}

		targetDomainSID, err := targetGroup.Properties.Get(graphschemaAD.DomainSID.String()).String()
		if errors.Is(err, graph.ErrPropertyNotFound) || strings.TrimSpace(targetDomainSID) == "" {
			return nil
		} else if err != nil {
			return err
		}

		managementPaths, err := ops.FetchPathSet(tx.Relationships().Filter(query.And(
			query.Equals(query.StartID(), source.ID),
			query.Kind(query.Relationship(), azure.ManageEntraDS),
		)))
		if err != nil {
			return err
		}

		for _, managementPath := range managementPaths {
			domainService := managementPath.Terminal()
			correlationPaths, err := ops.FetchPathSet(tx.Relationships().Filter(query.And(
				query.Equals(query.StartID(), domainService.ID),
				query.Kind(query.Relationship(), azure.EntraDSFor),
			)))
			if err != nil {
				return err
			}

			for _, correlationPath := range correlationPaths {
				domain := correlationPath.Terminal()
				domainSID, err := domain.Properties.Get(graphschemaAD.DomainSID.String()).String()
				if errors.Is(err, graph.ErrPropertyNotFound) {
					continue
				} else if err != nil {
					return err
				} else if !strings.EqualFold(strings.TrimSpace(domainSID), strings.TrimSpace(targetDomainSID)) {
					continue
				}

				containmentPaths, err := ops.TraversePaths(tx, ops.TraversalPlan{
					Root:      domain,
					Direction: graph.DirectionOutbound,
					BranchQuery: func() graph.Criteria {
						return query.Kind(query.Relationship(), graphschemaAD.Contains)
					},
					PathFilter: func(_ *ops.TraversalContext, segment *graph.PathSegment) bool {
						return segment.Node.ID == targetGroup.ID
					},
				})
				if errors.Is(err, graph.ErrNoResultsFound) {
					continue
				} else if err != nil {
					return err
				} else if containmentPaths.Len() == 0 {
					continue
				}

				finalPaths.AddPath(managementPath)
				finalPaths.AddPath(correlationPath)
				finalPaths.AddPathSet(containmentPaths)
			}
		}

		return nil
	}); err != nil {
		return graph.NewPathSet(), err
	}

	return finalPaths, nil
}
