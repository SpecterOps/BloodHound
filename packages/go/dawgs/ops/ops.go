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

package ops

import (
	"context"
	"fmt"

	"github.com/RoaringBitmap/roaring/roaring64"
	"github.com/specterops/bloodhound/dawgs/cardinality"
	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/dawgs/query"
	"github.com/specterops/bloodhound/dawgs/util/size"
)

func FetchAllNodeProperties(tx graph.Transaction, nodes graph.NodeSet) error {
	return tx.Nodes().Filter(
		query.InIDs(query.NodeID(), nodes.IDs()...),
	).Fetch(func(cursor graph.Cursor[*graph.Node]) error {
		for next := range cursor.Chan() {
			nodes[next.ID] = next
		}

		return cursor.Error()
	})
}

func FetchNodeProperties(tx graph.Transaction, nodes graph.NodeSet, propertyNames []string) error {
	returningCriteria := make([]graph.Criteria, len(propertyNames)+1)
	returningCriteria[0] = query.NodeID()

	for idx, propertyName := range propertyNames {
		returningCriteria[idx+1] = query.NodeProperty(propertyName)
	}

	return tx.Nodes().Filter(
		query.InIDs(query.NodeID(), nodes.IDs()...),
	).Query(func(results graph.Result) error {
		var nodeID graph.ID

		for results.Next() {
			if values, err := results.Values(); err != nil {
				return err
			} else {
				nodeProperties := map[string]any{}

				// Map the node ID first
				if err := values.Map(&nodeID); err != nil {
					return err
				}

				// Map requested properties next by matching the name index
				for idx := 0; idx < len(propertyNames); idx++ {
					if next, err := values.Next(); err != nil {
						return err
					} else {
						nodeProperties[propertyNames[idx]] = next
					}
				}

				// Update the node in the node set
				nodes[nodeID].Properties = graph.AsProperties(nodeProperties)
			}
		}

		return nil
	}, query.Returning(
		returningCriteria...,
	))
}

func DeleteNodes(tx graph.Transaction, nodeIDs ...graph.ID) error {
	return tx.Nodes().Filterf(func() graph.Criteria {
		return query.InIDs(query.NodeID(), nodeIDs...)
	}).Delete()
}

func DeleteRelationshipsQuery(relationshipIDs ...graph.ID) graph.CriteriaProvider {
	return func() graph.Criteria {
		return query.InIDs(query.RelationshipID(), relationshipIDs...)
	}
}

func DeleteRelationships(tx graph.Transaction, relationshipIDs ...graph.ID) error {
	return tx.Relationships().Filterf(DeleteRelationshipsQuery(relationshipIDs...)).Delete()
}

func FetchNodeSet(query graph.NodeQuery) (graph.NodeSet, error) {
	nodes := graph.NewNodeSet()

	return nodes, query.Fetch(func(cursor graph.Cursor[*graph.Node]) error {
		for node := range cursor.Chan() {
			nodes.Add(node)
		}

		return cursor.Error()
	})
}

func DBFetchNodesByIDBitmap(ctx context.Context, db graph.Database, nodeIDs cardinality.Duplex[uint32]) ([]*graph.Node, error) {
	var nodes []*graph.Node

	return nodes, db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		if fetchedNodes, err := TXFetchNodesByIDBitmap(tx, nodeIDs); err != nil {
			return err
		} else {
			nodes = fetchedNodes
		}

		return nil
	})
}

func TXFetchNodesByIDBitmap(tx graph.Transaction, nodeIDs cardinality.Duplex[uint32]) ([]*graph.Node, error) {
	return FetchNodes(tx.Nodes().Filter(query.InIDs(query.NodeID(), graph.Uint32SliceToIDs(nodeIDs.Slice())...)))
}

func FetchNodeIDsOfKindFromBitmap(tx graph.Transaction, nodeIDs cardinality.Duplex[uint32], kinds ...graph.Kind) ([]graph.ID, error) {
	return FetchNodeIDs(tx.Nodes().Filter(
		query.And(
			query.InIDs(query.NodeID(), graph.Uint32SliceToIDs(nodeIDs.Slice())...),
			query.KindIn(query.Node(), kinds...),
		)))
}

func FetchNodes(query graph.NodeQuery) ([]*graph.Node, error) {
	var nodes []*graph.Node

	return nodes, query.Fetch(func(cursor graph.Cursor[*graph.Node]) error {
		for node := range cursor.Chan() {
			nodes = append(nodes, node)
		}

		return cursor.Error()
	})
}

func FetchNodeIDs(query graph.NodeQuery) ([]graph.ID, error) {
	var ids []graph.ID

	return ids, query.FetchIDs(func(cursor graph.Cursor[graph.ID]) error {
		for id := range cursor.Chan() {
			ids = append(ids, id)
		}

		return cursor.Error()
	})
}

func FetchPathSetByQuery(tx graph.Transaction, query string) (graph.PathSet, error) {
	var (
		currentPath graph.Path
		pathSet     graph.PathSet
	)

	if result := tx.Query(query, map[string]any{}); result.Error() != nil {
		return pathSet, result.Error()
	} else {
		defer result.Close()

		for result.Next() {
			var (
				relationship graph.Relationship
				node         graph.Node
				path         graph.Path
			)

			if values, err := result.Values(); err != nil {
				return pathSet, err
			} else if mapped, err := values.MapOptions(&relationship, &node, &path); err != nil {
				return pathSet, err
			} else {
				switch typedMapped := mapped.(type) {
				case *graph.Relationship:
					currentPath.Edges = append(currentPath.Edges, typedMapped)

				case *graph.Node:
					currentPath.Nodes = append(currentPath.Nodes, typedMapped)

				case *graph.Path:
					pathSet = append(pathSet, *typedMapped)
				}

				currentPathSize := size.OfSlice(currentPath.Edges) + size.OfSlice(currentPath.Nodes)
				pathSetSize := size.Of(pathSet)

				if currentPathSize > tx.TraversalMemoryLimit() || pathSetSize > tx.TraversalMemoryLimit() {
					return pathSet, fmt.Errorf("%s - Limit: %.2f MB", "query required more memory than allowed", tx.TraversalMemoryLimit().Mebibytes())
				}
			}
		}

		// If there were elements added to the current path ensure that it's added to the path set before returning
		if len(currentPath.Nodes) > 0 || len(currentPath.Edges) > 0 {
			pathSet = append(pathSet, currentPath)
		}

		return pathSet, result.Error()
	}
}

func FetchNode(tx graph.Transaction, id graph.ID) (*graph.Node, error) {
	return tx.Nodes().Filterf(func() graph.Criteria {
		return query.Equals(query.NodeID(), id)
	}).First()
}

func FetchRelationship(tx graph.Transaction, id graph.ID) (*graph.Relationship, error) {
	return tx.Relationships().Filterf(func() graph.Criteria {
		return query.Equals(query.RelationshipID(), id)
	}).First()
}

func FetchNodeRelationships(tx graph.Transaction, root *graph.Node, direction graph.Direction) ([]*graph.Relationship, error) {
	var queryCriteria graph.Criteria

	switch direction {
	case graph.DirectionInbound:
		queryCriteria = query.InIDs(query.EndID(), root.ID)

	case graph.DirectionOutbound:
		queryCriteria = query.InIDs(query.StartID(), root.ID)

	default:
		return nil, fmt.Errorf("unexpected direction %d", direction)
	}

	return FetchRelationships(tx.Relationships().Filterf(func() graph.Criteria {
		return queryCriteria
	}))
}

func CollectNodeIDs(relationshipQuery graph.RelationshipQuery) (RelationshipNodes, error) {
	var relNodeIDs RelationshipNodes

	return relNodeIDs, relationshipQuery.Fetch(func(cursor graph.Cursor[*graph.Relationship]) error {
		for relationship := range cursor.Chan() {
			relNodeIDs.Add(relationship)
		}

		return cursor.Error()
	})
}

func ForEachNodeID(tx graph.Transaction, ids []graph.ID, delegate func(node *graph.Node) error) error {
	return tx.Nodes().Filterf(func() graph.Criteria {
		return query.InIDs(query.NodeID(), ids...)
	}).Fetch(func(cursor graph.Cursor[*graph.Node]) error {
		for node := range cursor.Chan() {
			if err := delegate(node); err != nil {
				return err
			}
		}

		return cursor.Error()
	})
}

func ForEachStartNode(relationshipQuery graph.RelationshipQuery, delegate func(relationship *graph.Relationship, node *graph.Node) error) error {
	return relationshipQuery.FetchDirection(graph.DirectionOutbound, func(cursor graph.Cursor[graph.DirectionalResult]) error {
		for result := range cursor.Chan() {
			if err := delegate(result.Relationship, result.Node); err != nil {
				return err
			}
		}

		return cursor.Error()
	})
}

func ForEachEndNode(relationshipQuery graph.RelationshipQuery, delegate func(relationship *graph.Relationship, node *graph.Node) error) error {
	return relationshipQuery.FetchDirection(graph.DirectionInbound, func(cursor graph.Cursor[graph.DirectionalResult]) error {
		for result := range cursor.Chan() {
			if err := delegate(result.Relationship, result.Node); err != nil {
				return err
			}
		}

		return cursor.Error()
	})
}

type RelationshipNodes struct {
	Start []graph.ID
	End   []graph.ID
}

func (s *RelationshipNodes) Add(relationship *graph.Relationship) {
	s.Start = append(s.Start, relationship.StartID)
	s.End = append(s.End, relationship.EndID)
}

func FetchStartNodeIDs(query graph.RelationshipQuery) ([]graph.ID, error) {
	var ids []graph.ID

	return ids, ForEachStartNode(query, func(_ *graph.Relationship, node *graph.Node) error {
		ids = append(ids, node.ID)
		return nil
	})
}

func FetchStartNodes(query graph.RelationshipQuery) (graph.NodeSet, error) {
	nodes := graph.NewNodeSet()

	return nodes, ForEachStartNode(query, func(_ *graph.Relationship, node *graph.Node) error {
		nodes.Add(node)
		return nil
	})
}

func FetchEndNodes(query graph.RelationshipQuery) (graph.NodeSet, error) {
	nodes := graph.NewNodeSet()

	return nodes, ForEachEndNode(query, func(_ *graph.Relationship, node *graph.Node) error {
		nodes.Add(node)
		return nil
	})
}

func FetchRelationships(query graph.RelationshipQuery) ([]*graph.Relationship, error) {
	var relationships []*graph.Relationship

	return relationships, query.Fetch(func(cursor graph.Cursor[*graph.Relationship]) error {
		for relationship := range cursor.Chan() {
			relationships = append(relationships, relationship)
		}

		return cursor.Error()
	})
}

func FetchRelationshipIDs(query graph.RelationshipQuery) ([]graph.ID, error) {
	var relationshipIDs []graph.ID

	return relationshipIDs, query.FetchIDs(func(cursor graph.Cursor[graph.ID]) error {
		for relationshipID := range cursor.Chan() {
			relationshipIDs = append(relationshipIDs, relationshipID)
		}

		return cursor.Error()
	})
}

func FetchPathSet(queryInst graph.RelationshipQuery) (graph.PathSet, error) {
	pathSet := graph.NewPathSet()

	return pathSet, queryInst.Query(func(results graph.Result) error {
		defer results.Close()

		for results.Next() {
			var (
				start, end graph.Node
				edge       graph.Relationship
			)

			if err := results.Scan(&start, &edge, &end); err != nil {
				return err
			} else {
				pathSet.AddPath(graph.Path{
					Nodes: []*graph.Node{&start, &end},
					Edges: []*graph.Relationship{&edge},
				})
			}
		}

		return results.Error()
	}, query.Returning(
		query.Start(), query.Relationship(), query.End(),
	))
}

func FetchRelationshipNodes(tx graph.Transaction, relationship *graph.Relationship) (*graph.Node, *graph.Node, error) {
	if nodes, err := FetchNodes(tx.Nodes().Filterf(func() graph.Criteria {
		return query.InIDs(query.NodeID(), relationship.StartID, relationship.EndID)
	})); err != nil {
		return nil, nil, err
	} else if len(nodes) != 2 {
		return nil, nil, graph.ErrNoResultsFound
	} else {
		var (
			startNode = nodes[0]
			endNode   = nodes[1]
		)

		if startNode.ID != relationship.StartID {
			startNode = nodes[1]
			endNode = nodes[0]
		}

		return startNode, endNode, nil
	}
}

func NodeQueryToIDBitmap(query graph.NodeQuery) (*roaring64.Bitmap, error) {
	bitmap := roaring64.NewBitmap()

	return bitmap, query.FetchIDs(func(cursor graph.Cursor[graph.ID]) error {
		for nextID := range cursor.Chan() {
			bitmap.Add(nextID.Uint64())
		}

		return cursor.Error()
	})
}
