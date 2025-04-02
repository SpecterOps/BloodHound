// Copyright 2025 Specter Ops, Inc.
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

package datapipe

import (
	"context"
	"fmt"
	"log/slog"
	"slices"
	"sync"

	"github.com/specterops/bloodhound/analysis"
	"github.com/specterops/bloodhound/bhlog/measure"
	"github.com/specterops/bloodhound/dawgs/cardinality"
	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/dawgs/ops"
	"github.com/specterops/bloodhound/dawgs/query"
	"github.com/specterops/bloodhound/dawgs/traversal"
	"github.com/specterops/bloodhound/graphschema/ad"
	"github.com/specterops/bloodhound/graphschema/azure"
	"github.com/specterops/bloodhound/graphschema/common"
	"github.com/specterops/bloodhound/src/database"
	"github.com/specterops/bloodhound/src/database/types/null"
	"github.com/specterops/bloodhound/src/model"
)

// Below is a thread safe way to track the source of why a node was selected by a selector
type srcSet struct {
	srcMap map[graph.ID]model.AssetGroupSelectorNodeSource
	lock   *sync.Mutex
}

func (s srcSet) AddSet(set graph.NodeSet, src model.AssetGroupSelectorNodeSource) {
	s.lock.Lock()
	defer s.lock.Unlock()
	for id := range set {
		if _, ok := s.srcMap[id]; !ok {
			s.srcMap[id] = src
		}
	}
}

func (s srcSet) AddIfNotExists(id graph.ID, src model.AssetGroupSelectorNodeSource) {
	s.lock.Lock()
	defer s.lock.Unlock()
	if _, ok := s.srcMap[id]; !ok {
		s.srcMap[id] = src
	}
}

func (s srcSet) GetSrc(id graph.ID) model.AssetGroupSelectorNodeSource {
	s.lock.Lock()
	defer s.lock.Unlock()

	return s.srcMap[id]
}

func newSrcSet() *srcSet {
	lock := sync.Mutex{}
	return &srcSet{
		srcMap: make(map[graph.ID]model.AssetGroupSelectorNodeSource),
		lock:   &lock,
	}
}

func fetchNodesFromSeeds(ctx context.Context, graphDb graph.Database, seeds []model.SelectorSeed, expansionMethod model.AssetGroupExpansionMethod) (graph.ThreadSafeNodeSet, srcSet, error) {
	var (
		result       = graph.NewThreadSafeNodeSet(graph.NodeSet{})
		resultSrcSet = newSrcSet()
		err          error
	)

	if err = graphDb.ReadTransaction(ctx, func(tx graph.Transaction) error {
		// Then we grab the nodes that should be selected
		for _, seed := range seeds {
			switch seed.Type {
			case model.SelectorTypeObjectId:
				if seedNodes, err := ops.FetchNodeSet(tx.Nodes().Filter(query.Equals(query.NodeProperty(common.ObjectID.String()), seed.Value))); err != nil {
					return err
				} else {
					resultSrcSet.AddSet(seedNodes, model.AssetGroupSelectorNodeSourceSeed)
					result.AddSet(seedNodes)
				}
			case model.SelectorTypeCypher:
				if seedNodes, err := ops.FetchNodesByQuery(tx, seed.Value); err != nil {
					return err
				} else {
					resultSrcSet.AddSet(seedNodes, model.AssetGroupSelectorNodeSourceSeed)
					result.AddSet(seedNodes)
				}
			default:
				slog.WarnContext(ctx, fmt.Sprintf("AGT: Unsupported selector type: %d", seed.Type))
			}
		}
		return nil
	}); err != nil {
		return *result, *resultSrcSet, err
	}

	if expansionMethod == model.AssetGroupExpansionMethodNone {
		return *result, *resultSrcSet, nil
	}

	traversalInst := traversal.New(graphDb, analysis.MaximumDatabaseParallelWorkers)

	if expansionMethod == model.AssetGroupExpansionMethodAll || expansionMethod == model.AssetGroupExpansionMethodChildren {
		// Expand to child nodes recursively as needed based on kind
		for _, node := range result.Copy().Slice() {
			if err = fetchChildNodes(ctx, traversalInst, node, result, resultSrcSet); err != nil {
				return *result, *resultSrcSet, err
			}
		}
	}

	if expansionMethod == model.AssetGroupExpansionMethodAll || expansionMethod == model.AssetGroupExpansionMethodParents {
		// Expand to parent nodes as needed
		for _, node := range result.Copy().Slice() {
			if err = fetchParentNodes(ctx, traversalInst, node, result, resultSrcSet); err != nil {
				return *result, *resultSrcSet, err
			}
		}
	}

	return *result, *resultSrcSet, err
}

func fetchParentNodes(ctx context.Context, tx traversal.Traversal, node *graph.Node, result *graph.ThreadSafeNodeSet, resultSrcSet *srcSet) error {
	if node.Kinds.ContainsOneOf(ad.Entity) {
		// MATCH (n:OU)-[:Contains*..]->(m:Base) RETURN n
		if err := tx.BreadthFirst(ctx, traversal.Plan{
			Root: node,
			Driver: traversal.NewPattern().InboundWithDepth(0, 0,
				query.And(
					query.Kind(query.Relationship(), ad.Contains),
					query.Kind(query.Start(), ad.OU),
				)).InboundWithDepth(0, 1,
				query.And(
					query.Kind(query.Relationship(), ad.GPLink),
					query.Kind(query.Start(), ad.GPO),
				)).Do(func(path *graph.PathSegment) error {
				if path.Trunk != nil {
					if result.AddIfNotExists(path.Trunk.Node) {
						resultSrcSet.AddIfNotExists(path.Trunk.Node.ID, model.AssetGroupSelectorNodeSourceParent)
					}
				}
				if result.AddIfNotExists(path.Node) {
					resultSrcSet.AddIfNotExists(path.Node.ID, model.AssetGroupSelectorNodeSourceParent)
				}

				return nil
			})}); err != nil {
			return err
		}

		// MATCH (n:Container)-[:Contains*..]->(m:Base) AND m.isaclprotected = False RETURN n
		if isAclProtected, err := node.Properties.Get(ad.IsACLProtected.String()).Bool(); err == nil && !isAclProtected {
			if err := tx.BreadthFirst(ctx, traversal.Plan{
				Root: node,
				Driver: traversal.NewPattern().InboundWithDepth(0, 0, query.And(
					query.Kind(query.Relationship(), ad.Contains),
					query.Kind(query.Start(), ad.Container),
				)).Do(func(path *graph.PathSegment) error {
					if path.Trunk != nil {
						if result.AddIfNotExists(path.Trunk.Node) {
							resultSrcSet.AddIfNotExists(path.Trunk.Node.ID, model.AssetGroupSelectorNodeSourceParent)
						}
					}
					if result.AddIfNotExists(path.Node) {
						resultSrcSet.AddIfNotExists(path.Node.ID, model.AssetGroupSelectorNodeSourceParent)
					}
					return nil
				})}); err != nil {
				return err
			}
		}

		// MATCH (n:GPO)-[:GPLink]->(m:Base) WHERE (m:Domain) OR (m:OU) RETURN n
		if node.Kinds.ContainsOneOf(ad.OU, ad.Domain) {
			if err := tx.BreadthFirst(ctx, traversal.Plan{
				Root: node,
				Driver: traversal.NewPattern().InboundWithDepth(0, 1, query.And(
					query.Kind(query.Relationship(), ad.GPLink),
					query.Kind(query.Start(), ad.GPO),
				)).Do(func(path *graph.PathSegment) error {
					if path.Trunk != nil {
						if result.AddIfNotExists(path.Trunk.Node) {
							resultSrcSet.AddIfNotExists(path.Trunk.Node.ID, model.AssetGroupSelectorNodeSourceParent)
						}
					}
					if result.AddIfNotExists(path.Node) {
						resultSrcSet.AddIfNotExists(path.Node.ID, model.AssetGroupSelectorNodeSourceParent)
					}
					return nil
				})}); err != nil {
				return err
			}
		}
	} else if node.Kinds.ContainsOneOf(azure.Entity) {
		// MATCH (n:AZBase)-[:AZContains*..]->(m:AZBase) WHERE (n:Subscription) OR (n:ResourceGroup) OR (n:ManagementGroup) RETURN n
		if err := tx.BreadthFirst(ctx, traversal.Plan{
			Root: node,
			Driver: traversal.NewPattern().InboundWithDepth(0, 0, query.And(
				query.KindIn(query.Relationship(), azure.Contains),
				query.KindIn(query.Start(), azure.Subscription, azure.ResourceGroup, azure.ManagementGroup),
			)).Do(func(path *graph.PathSegment) error {
				if path.Trunk != nil {
					if result.AddIfNotExists(path.Trunk.Node) {
						resultSrcSet.AddIfNotExists(path.Trunk.Node.ID, model.AssetGroupSelectorNodeSourceParent)
					}
				}
				if result.AddIfNotExists(path.Node) {
					resultSrcSet.AddIfNotExists(path.Node.ID, model.AssetGroupSelectorNodeSourceParent)
				}
				return nil
			})}); err != nil {
			return err
		}

		if node.Kinds.ContainsOneOf(azure.ServicePrincipal) {
			// MATCH (n:AZApp)-[:AZRunsAs]->(m:AZServicePrincipal) RETURN n
			if err := tx.BreadthFirst(ctx, traversal.Plan{
				Root: node,
				Driver: traversal.NewPattern().InboundWithDepth(0, 1, query.And(
					query.Kind(query.Relationship(), azure.RunsAs),
					query.Kind(query.Start(), azure.App),
				)).Do(func(path *graph.PathSegment) error {
					if path.Trunk != nil {
						if result.AddIfNotExists(path.Trunk.Node) {
							resultSrcSet.AddIfNotExists(path.Trunk.Node.ID, model.AssetGroupSelectorNodeSourceParent)
						}
					}
					if result.AddIfNotExists(path.Node) {
						resultSrcSet.AddIfNotExists(path.Node.ID, model.AssetGroupSelectorNodeSourceParent)
					}
					return nil
				})}); err != nil {
				return err
			}
		}
	}

	return nil
}

func fetchChildNodes(ctx context.Context, tx traversal.Traversal, node *graph.Node, result *graph.ThreadSafeNodeSet, resultSrcSet *srcSet) error {
	var pattern traversal.PatternContinuation

	switch {
	case node.Kinds.ContainsOneOf(ad.Group, azure.Group):
		// MATCH (m:Group)<-[:MemberOf*..]-(n:Base) RETURN n
		// MATCH (m:AZGroup)<-[:AZMemberOf*..]-(n:AZBase)RETURN n
		pattern = traversal.NewPattern().InboundWithDepth(0, 0, query.And(
			query.KindIn(query.Relationship(), ad.MemberOf, azure.MemberOf),
			query.KindIn(query.End(), ad.Entity, azure.Entity),
		))
	case node.Kinds.ContainsOneOf(ad.OU, ad.Container, azure.ResourceGroup, azure.ManagementGroup, azure.Subscription):
		// MATCH (n:Container)-[:Contains*..]->(m:Base) RETURN m
		// MATCH (n:OU)-[:Contains*..]->(m:Base) RETURN m
		// MATCH (n:AZResourceGroup)-[:AZContains*..]->(m:AZBase) RETURN m
		// MATCH (n:AZManagementGroup)-[:AZContains*..]->(m:AZBase) RETURN m
		// MATCH (n:AZSubscription)-[:AZContains*..]->(m:AZBase) RETURN m
		pattern = traversal.NewPattern().OutboundWithDepth(0, 0, query.And(
			query.KindIn(query.Relationship(), ad.Contains, azure.Contains),
			query.KindIn(query.End(), ad.Entity, azure.Entity),
		))
	case node.Kinds.ContainsOneOf(azure.Role):
		// MATCH (m:AZRole)<-[:AZHasRole]-(n:AZBase) RETURN n
		pattern = traversal.NewPattern().InboundWithDepth(0, 0, query.And(
			query.KindIn(query.Relationship(), azure.HasRole),
			query.KindIn(query.End(), azure.Entity),
		))
	default:
		// Skip any that do not need expanding
		return nil
	}

	addedNodes := graph.NewThreadSafeNodeSet(graph.NodeSet{})
	if err := tx.BreadthFirst(ctx, traversal.Plan{
		Root: node,
		Driver: pattern.Do(func(path *graph.PathSegment) error {
			if path.Trunk != nil {
				if result.AddIfNotExists(path.Trunk.Node) {
					resultSrcSet.AddIfNotExists(path.Trunk.Node.ID, model.AssetGroupSelectorNodeSourceExpand)
					addedNodes.Add(path.Trunk.Node)
				}
			}
			if result.AddIfNotExists(path.Node) {
				resultSrcSet.AddIfNotExists(path.Node.ID, model.AssetGroupSelectorNodeSourceExpand)
				addedNodes.Add(path.Node)
			}

			return nil
		})}); err != nil {
		return err
	}

	if addedNodes != nil && addedNodes.Len() > 0 {
		for _, node := range addedNodes.Slice() {
			resultSrcSet.AddIfNotExists(node.ID, model.AssetGroupSelectorNodeSourceExpand)
			// Expand to child nodes as needed based on kind
			if err := fetchChildNodes(ctx, tx, node, result, resultSrcSet); err != nil {
				return err
			}
		}
	}

	return nil
}

// TODO Batching?
func SelectNodes(ctx context.Context, db database.Database, graphDb graph.Database, selector model.AssetGroupTagSelector, expansionMethod model.AssetGroupExpansionMethod) error {
	var (
		countInserted int

		certified   = model.AssetGroupCertificationNone
		certifiedBy null.String

		oldSelectedNodes []model.AssetGroupSelectorNode

		nodeIdsToDelete []graph.ID
		nodeIdsToUpdate []graph.ID
	)
	if selector.AutoCertify {
		certified = model.AssetGroupCertificationAuto
		certifiedBy = null.StringFrom(model.AssetGroupActorSystem)
	}

	// 1. Grab the graph nodes
	if nodes, srcSet, err := fetchNodesFromSeeds(ctx, graphDb, selector.Seeds, expansionMethod); err != nil {
		return err
		// 2. Grab the already selected nodes
	} else if oldSelectedNodes, err = db.GetSelectorNodesBySelectorIds(ctx, selector.ID); err != nil {
		return err
	} else {
		oldSelectedNodesByNodeId := make(map[graph.ID]*model.AssetGroupSelectorNode)
		for _, node := range oldSelectedNodes {
			oldSelectedNodesByNodeId[node.NodeId] = &node
		}

		// 3. Range the graph nodes and insert any that haven't been inserted yet, mark for update any that need updating, pare down the existing map for future deleting
		for _, id := range nodes.IDs() {
			// Missing, insert the record
			if oldSelectedNodesByNodeId[id] == nil {
				if err = db.InsertSelectorNode(ctx, selector.ID, id, certified, certifiedBy, srcSet.GetSrc(id)); err != nil {
					return err
				}
				countInserted++
				// Auto certify is enabled but this node hasn't been certified, certify it
			} else if selector.AutoCertify && oldSelectedNodesByNodeId[id].Certified == model.AssetGroupCertificationNone {
				nodeIdsToUpdate = append(nodeIdsToUpdate, id)
				delete(oldSelectedNodesByNodeId, id)
			} else {
				delete(oldSelectedNodesByNodeId, id)
			}
		}

		// Update the selected nodes that need updating
		if len(nodeIdsToUpdate) > 0 {
			if err = db.UpdateSelectorNodesByNodeId(ctx, selector.ID, certified, certifiedBy, nodeIdsToUpdate...); err != nil {
				return err
			}
		}

		// Delete the selected nodes that need to be deleted
		if len(oldSelectedNodesByNodeId) > 0 {
			for nodeId := range oldSelectedNodesByNodeId {
				nodeIdsToDelete = append(nodeIdsToDelete, nodeId)
			}
			if err = db.DeleteSelectorNodesByNodeId(ctx, selector.ID, nodeIdsToDelete...); err != nil {
				return err
			}
		}

		slog.Info("AGT: Completed selecting", "selector", selector.Name, "countTotal", nodes.Len(), "countInserted", countInserted, "countUpdated", len(nodeIdsToUpdate), "countDeleted", len(nodeIdsToDelete))
	}
	return nil
}

func SelectAssetGroupNodes(ctx context.Context, db database.Database, graphDb graph.Database) error {
	defer measure.ContextMeasure(ctx, slog.LevelInfo, "Finished selecting asset group nodes via new selectors")()

	if tags, err := db.GetAssetGroupTagForSelection(ctx); err != nil {
		return err
	} else {
		for _, tag := range tags {
			if selectors, err := db.GetAssetGroupTagSelectorsByTagId(ctx, tag.ID, model.SQLFilter{}, model.SQLFilter{}); err != nil {
				return err
			} else {
				wg := sync.WaitGroup{}
				for _, selector := range selectors {
					if !selector.DisabledAt.IsZero() {
						continue
					}
					// Parallelize the selection of nodes
					go func() {
						defer wg.Done()
						if err = SelectNodes(ctx, db, graphDb, selector, tag.GetExpansionMethod()); err != nil {
							slog.Error("AGT: Error selecting nodes", "selector", selector, "err", err)
						}
					}()
					wg.Add(1)
				}
				wg.Wait()
			}
		}
	}
	return nil
}

// TODO Batching?
func tagAssetGroupNodes(ctx context.Context, db database.Database, graphDb graph.Database, tag model.AssetGroupTag) error {
	if selectors, err := db.GetAssetGroupTagSelectorsByTagId(ctx, tag.ID, model.SQLFilter{}, model.SQLFilter{}); err != nil {
		return err
	} else {
		var (
			countTagged, countUntagged, countTotal int
			selectorIds                            []int
			selectedNodes                          []model.AssetGroupSelectorNode

			tagKind = tag.ToKind()

			nodesSeen      = cardinality.NewBitmap64()
			oldTaggedNodes = cardinality.NewBitmap64()
			newTaggedNodes = cardinality.NewBitmap64()
		)

		disabledSelectors := cardinality.NewBitmap32()
		for _, selector := range selectors {
			if !selector.DisabledAt.IsZero() {
				disabledSelectors.Add(uint32(selector.ID))
			}
			selectorIds = append(selectorIds, selector.ID)
		}

		// 1. Fetch the selected nodes for this label
		if selectedNodes, err = db.GetSelectorNodesBySelectorIds(ctx, selectorIds...); err != nil {
			return err
		} else if err = graphDb.WriteTransaction(ctx, func(tx graph.Transaction) error {
			// 2. Fetch already tagged nodes
			if oldTaggedNodeSet, err := ops.FetchNodeSet(tx.Nodes().Filter(query.Kind(query.Node(), tagKind))); err != nil {
				return err
			} else {
				oldTaggedNodes = oldTaggedNodeSet.IDBitmap()
			}

			// 3. Diff the sets filling the respective sets for later db updates
			for _, nodeDb := range selectedNodes {
				if !nodesSeen.Contains(nodeDb.NodeId.Uint64()) {
					// Skip any that are not certified when tag requires certification or are selected by disabled selectors
					if tag.RequireCertify.Bool && nodeDb.Certified <= 0 || disabledSelectors.Contains(uint32(nodeDb.SelectorId)) {
						continue
					}

					// If the id is not present, we must queue it for tagging
					if !oldTaggedNodes.Contains(nodeDb.NodeId.Uint64()) {
						newTaggedNodes.Add(nodeDb.NodeId.Uint64())
					} else {
						// If it is present, we don't need to update anything and will remove tags from any nodes left in this bitmap
						oldTaggedNodes.Remove(nodeDb.NodeId.Uint64())
					}
					// Once a node is processed, we can skip future duplicates that might be selected by other selectors
					nodesSeen.Add(nodeDb.NodeId.Uint64())
					countTotal++
				}
			}

			// 4. Tag the new nodes
			newTaggedNodes.Each(func(nodeId uint64) bool {
				node := &graph.Node{ID: graph.ID(nodeId), Properties: &graph.Properties{}}
				node.AddKinds(tagKind)
				err = tx.UpdateNode(node)
				countTagged++
				return err == nil
			})
			if err != nil {
				return err
			}

			// 5. Remove the old nodes
			oldTaggedNodes.Each(func(nodeId uint64) bool {
				node := &graph.Node{ID: graph.ID(nodeId), Properties: &graph.Properties{}}
				node.DeleteKinds(tagKind)
				err = tx.UpdateNode(node)
				countUntagged++
				return err == nil
			})
			if err != nil {
				return err
			}

			return nil
		}); err != nil {
			return err
		}

		slog.Info("AGT: Completed tagging", tag.ToType(), tag.Name, "total", countTotal, "tagged", countTagged, "untagged", countUntagged)
	}
	return nil
}

func TagAssetGroupNodes(ctx context.Context, db database.Database, graphDb graph.Database) error {
	defer measure.ContextMeasure(ctx, slog.LevelInfo, "Finished tagging asset group nodes")()

	if tags, err := db.GetAssetGroupTagForSelection(ctx); err != nil {
		return err
	} else {
		// Tiers are hierarchical and must be handled synchronously while labels can be tagged in parallel
		var (
			labelsOrOwned []model.AssetGroupTag
			tiersOrdered  []model.AssetGroupTag
		)
		for _, tag := range tags {
			switch tag.Type {
			case model.AssetGroupTagTypeTier:
				tiersOrdered = append(tiersOrdered, tag)
			case model.AssetGroupTagTypeLabel, model.AssetGroupTagTypeOwned:
				labelsOrOwned = append(labelsOrOwned, tag)
			default:
				slog.WarnContext(ctx, fmt.Sprintf("AGT: Tag type %d is not supported", tag.Type), "tag", tag)
			}
		}

		// Order the tiers by position
		slices.SortFunc(tiersOrdered, func(i, j model.AssetGroupTag) int {
			return int(i.Position.Int32 - j.Position.Int32)
		})

		// Fire off the label tagging
		wg := sync.WaitGroup{}
		for _, tag := range labelsOrOwned {
			// Parallelize the tagging of label nodes
			go func() {
				defer wg.Done()
				if err = tagAssetGroupNodes(ctx, db, graphDb, tag); err != nil {
					slog.Error("AGT: Error tagging nodes", tag.ToType(), tag, "err", err)
				}
			}()
			wg.Add(1)
		}

		// Process the tier tagging synchronously
		for _, tier := range tiersOrdered {
			if err := tagAssetGroupNodes(ctx, db, graphDb, tier); err != nil {
				slog.Error("AGT: Error tagging nodes", "tier", tier, "err", err)
			}
		}

		// Wait for labels to finish
		wg.Wait()
	}
	return nil
}
