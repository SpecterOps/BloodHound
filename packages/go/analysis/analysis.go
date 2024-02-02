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
	"fmt"
	"slices"
	"strings"

	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/dawgs/ops"
	"github.com/specterops/bloodhound/dawgs/query"
	"github.com/specterops/bloodhound/graphschema/ad"
	"github.com/specterops/bloodhound/graphschema/azure"
	"github.com/specterops/bloodhound/graphschema/common"
	"github.com/specterops/bloodhound/log"
	"github.com/specterops/bloodhound/slicesext"
)

const (
	NodeKindUnknown                = "Unknown"
	MaximumDatabaseParallelWorkers = 6
)

var (
	metaKind       = graph.StringKind("Meta")
	metaDetailKind = graph.StringKind("MetaDetail")
)

func AllTaggedNodesFilter(additionalFilter graph.Criteria) graph.Criteria {
	var (
		filters = []graph.Criteria{
			query.IsNotNull(query.NodeProperty(common.SystemTags.String())),
		}
	)

	if additionalFilter != nil {
		filters = append(filters, additionalFilter)
	}

	return query.And(filters...)
}

func GetNodeKindDisplayLabel(node *graph.Node) string {
	return GetNodeKind(node).String()
}

func GetNodeKind(node *graph.Node) graph.Kind {
	var (
		resultKind = graph.StringKind(NodeKindUnknown)
		baseKind   = resultKind
	)

	for _, kind := range node.Kinds {
		// If this is a BHE meta kind, return early
		if kind.Is(metaKind, metaDetailKind) {
			return metaKind
		} else if kind.Is(ad.Entity, azure.Entity) {
			baseKind = kind
		} else if kind.Is(ad.LocalGroup) {
			// Allow ad.LocalGroup to overwrite NodeKindUnknown, but nothing else
			if resultKind.String() == NodeKindUnknown {
				resultKind = kind
			}
		} else if slices.Contains(ValidKinds(), kind) {
			resultKind = kind
		}
	}

	if resultKind.String() == NodeKindUnknown {
		return baseKind
	} else {
		return resultKind
	}
}

func ClearSystemTags(ctx context.Context, db graph.Database) error {
	defer log.Measure(log.LevelInfo, "ClearSystemTagsIncludeMeta")()

	var (
		props = graph.NewProperties()
	)

	props.Delete(common.SystemTags.String())

	return db.WriteTransaction(ctx, func(tx graph.Transaction) error {
		if ids, err := ops.FetchNodeIDs(tx.Nodes().Filter(AllTaggedNodesFilter(nil))); err != nil {
			return err
		} else {
			return tx.Nodes().Filterf(func() graph.Criteria {
				return query.InIDs(query.NodeID(), ids...)
			}).Update(props)
		}
	})
}

func ValidKinds() []graph.Kind {
	var (
		metaKinds = []graph.Kind{metaKind, metaDetailKind}
	)

	return slicesext.Concat(ad.Nodes(), ad.Relationships(), azure.NodeKinds(), azure.Relationships(), metaKinds)
}

func ParseKind(rawKind string) (graph.Kind, error) {
	for _, kind := range ValidKinds() {
		if kind.String() == rawKind {
			return kind, nil
		}
	}

	return nil, fmt.Errorf("unknown kind %s", rawKind)
}

func ParseKinds(rawKinds ...string) (graph.Kinds, error) {
	if len(rawKinds) == 0 {
		return graph.Kinds{ad.Entity, azure.Entity}, nil
	}

	return slicesext.MapWithErr(rawKinds, ParseKind)
}

func nodeByIndexedKindProperty(property, value string, kind graph.Kind) graph.Criteria {
	return query.And(
		query.Equals(query.NodeProperty(property), value),
		query.Kind(query.Node(), kind),
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

func ExpandGroupMembership(tx graph.Transaction, candidates graph.NodeSet) (graph.NodeSet, error) {
	if paths, err := ExpandGroupMembershipPaths(tx, candidates); err != nil {
		return nil, err
	} else {
		return paths.AllNodes(), nil
	}
}

func GetLAPSSyncers(tx graph.Transaction, domain *graph.Node) ([]*graph.Node, error) {
	var (
		getChangesQuery         = fromEntityToEntityWithRelationshipKind(tx, domain, ad.GetChanges, false)
		getChangesFilteredQuery = fromEntityToEntityWithRelationshipKind(tx, domain, ad.GetChangesInFilteredSet, false)
	)

	if getChangesNodes, err := ops.FetchStartNodes(getChangesQuery); err != nil {
		return nil, err
	} else if getChangesNodeMembers, err := ExpandGroupMembership(tx, getChangesNodes); err != nil {
		return nil, err
	} else if getChangesFilteredNodes, err := ops.FetchStartNodes(getChangesFilteredQuery); err != nil {
		return nil, err
	} else if getChangesFilteredNodeMembers, err := ExpandGroupMembership(tx, getChangesFilteredNodes); err != nil {
		return nil, err
	} else {
		// Collect and filter the bitmap
		getChangesNodes.AddSet(getChangesNodeMembers)
		getChangesFilteredNodes.AddSet(getChangesFilteredNodeMembers)

		syncerBitmap := graph.NodeSetToBitmap(getChangesNodes)
		syncerBitmap.And(graph.NodeSetToBitmap(getChangesFilteredNodes))

		var (
			nodeIDs = syncerBitmap.ToArray()
			nodes   = make([]*graph.Node, len(nodeIDs))
		)

		for idx, rawID := range syncerBitmap.ToArray() {
			// Since the bitmap is an intersection of both node sets each set is guaranteed to have a valid reference
			// to the node
			nodes[idx] = getChangesNodes.Get(graph.ID(int64(rawID)))
		}

		return nodes, nil
	}
}

func GetDCSyncers(tx graph.Transaction, domain *graph.Node, filterTierZero bool) ([]*graph.Node, error) {
	var (
		getChangesQuery    = fromEntityToEntityWithRelationshipKind(tx, domain, ad.GetChanges, filterTierZero)
		getChangesAllQuery = fromEntityToEntityWithRelationshipKind(tx, domain, ad.GetChangesAll, filterTierZero)
	)

	if getChangesNodes, err := ops.FetchStartNodes(getChangesQuery); err != nil {
		return nil, err
	} else if getChangesNodeMembers, err := ExpandGroupMembership(tx, getChangesNodes); err != nil {
		return nil, err
	} else if getChangesAllNodes, err := ops.FetchStartNodes(getChangesAllQuery); err != nil {
		return nil, err
	} else if getChangesAllNodeMembers, err := ExpandGroupMembership(tx, getChangesAllNodes); err != nil {
		return nil, err
	} else {
		// Collect and filter the bitmap
		getChangesNodes.AddSet(getChangesNodeMembers)
		getChangesAllNodes.AddSet(getChangesAllNodeMembers)

		if filterTierZero {
			//Do a second pass to filter out T0 nodes that might have ended up through group membership
			for _, node := range getChangesNodes {
				if systemTags, err := node.Properties.Get(common.SystemTags.String()).String(); err != nil {
					if graph.IsErrPropertyNotFound(err) {
						continue
					}

					return nil, err
				} else if strings.Contains(systemTags, ad.AdminTierZero) {
					getChangesNodes.Remove(node.ID)
				}
			}

			for _, node := range getChangesAllNodes {
				if systemTags, err := node.Properties.Get(common.SystemTags.String()).String(); err != nil {
					if graph.IsErrPropertyNotFound(err) {
						continue
					}

					return nil, err
				} else if strings.Contains(systemTags, ad.AdminTierZero) {
					getChangesNodes.Remove(node.ID)
				}
			}
		}

		dcSyncerBitmap := graph.NodeSetToBitmap(getChangesNodes)
		dcSyncerBitmap.And(graph.NodeSetToBitmap(getChangesAllNodes))

		var (
			nodeIDs = dcSyncerBitmap.ToArray()
			nodes   = make([]*graph.Node, len(nodeIDs))
		)

		for idx, rawID := range dcSyncerBitmap.ToArray() {
			// Since the bitmap is an intersection of both node sets each set is guaranteed to have a valid reference
			// to the node
			nodes[idx] = getChangesNodes.Get(graph.ID(int64(rawID)))
		}

		return nodes, nil
	}
}

func fromEntityToEntityWithRelationshipKind(tx graph.Transaction, target *graph.Node, relKind graph.Kind, filterTierZero bool) graph.RelationshipQuery {
	return tx.Relationships().Filterf(func() graph.Criteria {
		filters := []graph.Criteria{
			query.Kind(query.Start(), ad.Entity),
			query.Kind(query.Relationship(), relKind),
			query.Equals(query.EndID(), target.ID),
		}

		if filterTierZero {
			filters = append(filters, query.Not(
				query.StringContains(query.StartProperty(common.SystemTags.String()), ad.AdminTierZero),
			))
		}

		return query.And(filters...)
	})
}

type PathDelegate = func(tx graph.Transaction, node *graph.Node) (graph.PathSet, error)
type ListDelegate = func(tx graph.Transaction, node *graph.Node, skip, limit int) (graph.NodeSet, error)

type ParallelPathDelegate = func(ctx context.Context, db graph.Database, node *graph.Node) (graph.PathSet, error)
type ParallelListDelegate = func(ctx context.Context, db graph.Database, node *graph.Node, skip int, limit int) (graph.NodeSet, error)
