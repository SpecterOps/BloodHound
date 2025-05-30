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
	"sync/atomic"

	"github.com/specterops/bloodhound/analysis"
	"github.com/specterops/bloodhound/bhlog/measure"
	"github.com/specterops/bloodhound/dawgs/cardinality"
	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/dawgs/ops"
	"github.com/specterops/bloodhound/dawgs/query"
	"github.com/specterops/bloodhound/dawgs/traversal"
	"github.com/specterops/bloodhound/dawgs/util/channels"
	"github.com/specterops/bloodhound/graphschema/ad"
	"github.com/specterops/bloodhound/graphschema/azure"
	"github.com/specterops/bloodhound/graphschema/common"
	"github.com/specterops/bloodhound/src/database"
	"github.com/specterops/bloodhound/src/database/types/null"
	"github.com/specterops/bloodhound/src/model"
	"github.com/specterops/bloodhound/src/model/appcfg"
)

// This is a bespoke result set to contain a dedupe'd node with source info
type nodeWithSource struct {
	*graph.Node
	Source model.AssetGroupSelectorNodeSource
}

type nodeWithSourceSet map[graph.ID]*nodeWithSource

func (s nodeWithSourceSet) LimitReached(limit int) bool {
	if limit <= 0 {
		return false
	}

	return len(s) >= limit

}

func (s nodeWithSourceSet) AddIfNotExists(node *nodeWithSource) bool {
	if _, exists := s[node.ID]; exists {
		return false
	}
	s[node.ID] = node
	return true
}

// FetchNodesFromSeeds fetches all seed nodes along with any child or parent nodes via known expansion paths
func FetchNodesFromSeeds(ctx context.Context, graphDb graph.Database, seeds []model.SelectorSeed, expansionMethod model.AssetGroupExpansionMethod, limit int) nodeWithSourceSet {
	var (
		seedNodes = make(nodeWithSourceSet)
		result    = make(nodeWithSourceSet)
	)

	_ = graphDb.ReadTransaction(ctx, func(tx graph.Transaction) error {
		// Then we grab the nodes that should be selected
		for _, seed := range seeds {
			switch seed.Type {
			case model.SelectorTypeObjectId:
				if node, err := tx.Nodes().Filter(query.Equals(query.NodeProperty(common.ObjectID.String()), seed.Value)).First(); err != nil {
					slog.WarnContext(ctx, "AGT: Fetch Object ID Err", "objectId", seed.Value, "error", err)
				} else {
					nodeWithSrc := &nodeWithSource{Source: model.AssetGroupSelectorNodeSourceSeed, Node: node}
					if result.AddIfNotExists(nodeWithSrc) {
						if result.LimitReached(limit) {
							return nil
						}
					}
					seedNodes.AddIfNotExists(nodeWithSrc)
				}
			case model.SelectorTypeCypher:
				if nodes, err := ops.FetchNodesByQuery(tx, seed.Value, limit); err != nil {
					slog.WarnContext(ctx, "AGT: Fetch Cypher Err", "cypherQuery", seed.Value, "error", err)
				} else {
					for _, node := range nodes {
						nodeWithSrc := &nodeWithSource{Source: model.AssetGroupSelectorNodeSourceSeed, Node: node}
						if result.AddIfNotExists(nodeWithSrc) {
							if result.LimitReached(limit) {
								return nil
							}
						}
						seedNodes.AddIfNotExists(nodeWithSrc)
					}
				}
			default:
				slog.WarnContext(ctx, fmt.Sprintf("AGT: Unsupported selector type: %d", seed.Type))
			}
		}
		return nil
	})

	if expansionMethod == model.AssetGroupExpansionMethodNone || result.LimitReached(limit) || len(result) == 0 {
		return result
	}

	if expansionMethod == model.AssetGroupExpansionMethodAll || expansionMethod == model.AssetGroupExpansionMethodChildren {
		collected := fetchAllChildNodes(ctx, graphDb, seedNodes, result, limit)
		if result.LimitReached(limit) {
			return result
		}

		// Add any newly collected child nodes to seeds for optional parent expansion below
		for _, node := range collected {
			seedNodes.AddIfNotExists(node)
		}
	}

	if expansionMethod == model.AssetGroupExpansionMethodAll || expansionMethod == model.AssetGroupExpansionMethodParents {
		fetchParentNodes(ctx, graphDb, seedNodes, result, limit)
	}

	return result
}

// fetchChildNodes - fetches all children for a single node and submits any found to supplied collector ch
func fetchChildNodes(ctx context.Context, tx traversal.Traversal, node *graph.Node, ch chan<- *nodeWithSource) error {
	var pattern traversal.PatternContinuation

	switch {
	case node.Kinds.ContainsOneOf(ad.Group):
		// MATCH (n:Group)<-[:MemberOf*..]-(m:Base) RETURN m
		pattern = traversal.NewPattern().InboundWithDepth(0, 0, query.And(
			query.KindIn(query.Relationship(), ad.MemberOf),
			query.KindIn(query.End(), ad.Entity),
		))
	case node.Kinds.ContainsOneOf(ad.OU, ad.Container):
		// MATCH (n:Container)-[:Contains*..]->(m:Base) RETURN m
		// MATCH (n:OU)-[:Contains*..]->(m:Base) RETURN m
		pattern = traversal.NewPattern().OutboundWithDepth(0, 0, query.And(
			query.KindIn(query.Relationship(), ad.Contains),
			query.KindIn(query.End(), ad.Entity),
		))
	case node.Kinds.ContainsOneOf(azure.Group):
		// MATCH (n:AZGroup)<-[:AZMemberOf*..]-(m:AZBase) RETURN m
		pattern = traversal.NewPattern().InboundWithDepth(0, 0, query.And(
			query.KindIn(query.Relationship(), azure.MemberOf),
			query.KindIn(query.End(), azure.Entity),
		))
	case node.Kinds.ContainsOneOf(azure.ResourceGroup, azure.ManagementGroup, azure.Subscription):
		// MATCH (n:AZResourceGroup)-[:AZContains*..]->(m:AZBase) RETURN m
		// MATCH (n:AZManagementGroup)-[:AZContains*..]->(m:AZBase) RETURN m
		// MATCH (n:AZSubscription)-[:AZContains*..]->(m:AZBase) RETURN m
		pattern = traversal.NewPattern().OutboundWithDepth(0, 0, query.And(
			query.KindIn(query.Relationship(), azure.Contains),
			query.KindIn(query.End(), azure.Entity),
		))
	case node.Kinds.ContainsOneOf(azure.Role):
		// MATCH (n:AZRole)<-[:AZHasRole]-(m:AZBase) RETURN m
		pattern = traversal.NewPattern().InboundWithDepth(0, 1, query.And(
			query.KindIn(query.Relationship(), azure.HasRole),
			query.KindIn(query.End(), azure.Entity),
		))
	default:
		// Skip any that do not need expanding
		return nil
	}

	if err := tx.BreadthFirst(ctx, traversal.Plan{
		Root: node,
		Driver: pattern.Do(func(path *graph.PathSegment) error {
			path.WalkReverse(func(nextSegment *graph.PathSegment) bool {
				return channels.Submit(ctx, ch, &nodeWithSource{Source: model.AssetGroupSelectorNodeSourceChild, Node: nextSegment.Node})
			})
			return nil
		})}); err != nil {

		return err
	}

	return nil
}

// fetchAllChildNodes - concurrently fetches all seeds + their children until no additional children are found
func fetchAllChildNodes(ctx context.Context, db graph.Database, seedNodes nodeWithSourceSet, result nodeWithSourceSet, limit int) []*nodeWithSource {
	var (
		wg              = sync.WaitGroup{}
		queueLen        = &atomic.Int64{}
		chCtx, doneFunc = context.WithCancel(ctx)

		sendCh, getCh = channels.BufferedPipe[*nodeWithSource](chCtx)
		collectorCh   = make(chan *nodeWithSource)

		traversalInst = traversal.New(db, analysis.MaximumDatabaseParallelWorkers)
		collected     []*nodeWithSource
	)
	defer doneFunc()
	// Close the send channel to the buffered pipe
	defer close(sendCh)

	// Spin out some workers, at least 1 per seed node
	for range len(seedNodes) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				// Block here until we receive a node to fetch child nodes
				if node, ok := channels.Receive(chCtx, getCh); !ok {
					return
				} else {
					// Fetch child nodes for this node and send any collected to the collector
					if err := fetchChildNodes(chCtx, traversalInst, node.Node, collectorCh); err != nil {
						slog.ErrorContext(ctx, "AGT: error fetching child nodes", "node", node.ID, "err", err)
					}

					// Fire to collector to signal job done to synchronously decr queue
					if !channels.Submit(chCtx, collectorCh, nil) {
						break
					}
				}
			}
		}()
	}

	// Spin out a collector
	go func() {
		for {
			// Synchronously fill the collected child nodes and queue up more child checks
			// Keep track of queue length and trigger finish when queue is empty
			if nodeWithSrc, ok := channels.Receive(chCtx, collectorCh); !ok {
				break
			} else {
				if nodeWithSrc != nil && result.AddIfNotExists(nodeWithSrc) {
					// As new child nodes are found, fill collection and queue up a child check for that child as well
					collected = append(collected, nodeWithSrc)
					if result.LimitReached(limit) {
						doneFunc()
						break
					}
					queueLen.Add(1)
					if !channels.Submit(chCtx, sendCh, nodeWithSrc) {
						break
					}
				} else if nodeWithSrc == nil {
					// Decr and check if we need to stop
					queueLen.Add(-1)
					if queueLen.Load() == 0 {
						// Once queue hits 0, we fire ctx cancel to exit all workers
						doneFunc()
						break
					}
				}
			}
		}
	}()

	// Start off with seed nodes
	for _, node := range seedNodes {
		queueLen.Add(1)
		channels.Submit(chCtx, sendCh, node)
	}

	wg.Wait() // Wait for workers to process all nodes
	return collected
}

// fetchADParentNodes -  fetches all parents for a single active directory node and submits any found to supplied collector ch
func fetchADParentNodes(ctx context.Context, tx traversal.Traversal, node *graph.Node, ch chan<- *nodeWithSource) error {
	// MATCH (n:OU)-[:Contains*..]->(m:Base) RETURN n
	// MATCH (n:GPO)-[:GPLink]->(m:Base) WHERE (m:Domain) OR (m:OU) RETURN n
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
			path.WalkReverse(func(nextSegment *graph.PathSegment) bool {
				return channels.Submit(ctx, ch, &nodeWithSource{Source: model.AssetGroupSelectorNodeSourceParent, Node: nextSegment.Node})
			})
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
				path.WalkReverse(func(nextSegment *graph.PathSegment) bool {
					return channels.Submit(ctx, ch, &nodeWithSource{Source: model.AssetGroupSelectorNodeSourceParent, Node: nextSegment.Node})
				})
				return nil
			})}); err != nil {
			return err
		}
	}
	return nil
}

// fetchAzureParentNodes -  fetches all parents for a single azure node and submits any found to supplied collector ch
func fetchAzureParentNodes(ctx context.Context, tx traversal.Traversal, node *graph.Node, ch chan<- *nodeWithSource) error {
	// MATCH (n:AZBase)-[:AZContains*..]->(m:AZBase) WHERE (n:Subscription) OR (n:ResourceGroup) OR (n:ManagementGroup) RETURN n
	if err := tx.BreadthFirst(ctx, traversal.Plan{
		Root: node,
		Driver: traversal.NewPattern().InboundWithDepth(0, 0, query.And(
			query.KindIn(query.Relationship(), azure.Contains),
			query.KindIn(query.Start(), azure.Entity),
		)).Do(func(path *graph.PathSegment) error {
			path.WalkReverse(func(nextSegment *graph.PathSegment) bool {
				if nextSegment.Node.Kinds.ContainsOneOf(azure.Subscription, azure.ResourceGroup, azure.ManagementGroup) {
					return channels.Submit(ctx, ch, &nodeWithSource{Source: model.AssetGroupSelectorNodeSourceParent, Node: nextSegment.Node})
				}
				return true
			})
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
				path.WalkReverse(func(nextSegment *graph.PathSegment) bool {
					return channels.Submit(ctx, ch, &nodeWithSource{Source: model.AssetGroupSelectorNodeSourceParent, Node: nextSegment.Node})
				})
				return nil
			})}); err != nil {
			return err
		}
	}
	return nil
}

// fetchParentNodes - concurrently fetches all parents for seed nodes (which may also contain their children)
func fetchParentNodes(ctx context.Context, db graph.Database, seedNodes nodeWithSourceSet, result nodeWithSourceSet, limit int) {
	// Expand to parent nodes as needed
	var (
		wg                      = sync.WaitGroup{}
		ch                      = make(chan *nodeWithSource)
		ctxWithCancel, doneFunc = context.WithCancel(ctx)
		traversalInst           = traversal.New(db, analysis.MaximumDatabaseParallelWorkers)
	)
	defer doneFunc()
	// Spin out a job per node -> may be just seeds or seeds + children here
	for _, node := range seedNodes {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if node.Kinds.ContainsOneOf(ad.Entity) {
				if err := fetchADParentNodes(ctxWithCancel, traversalInst, node.Node, ch); err != nil {
					slog.ErrorContext(ctx, "AGT: error fetching active directory parent nodes", "node", node.ID, "err", err)
				}
			} else if node.Kinds.ContainsOneOf(azure.Entity) {
				if err := fetchAzureParentNodes(ctxWithCancel, traversalInst, node.Node, ch); err != nil {
					slog.ErrorContext(ctx, "AGT: error fetching azure parent nodes", "node", node.ID, "err", err)
				}
			}
		}()
	}

	// This will wait to close the channel and release the below for loop until all jobs are done
	go func() {
		wg.Wait()
		close(ch)
	}()

	// This will block and collect all parent nodes until channel is closed
	for nodeWithSrc := range ch {
		if result.AddIfNotExists(nodeWithSrc) && result.LimitReached(limit) {
			doneFunc()
		}
	}

}

// fetchOldSelectedNodes - fetches the currently selected nodes and assembles a map lookup for minimal memory footprint
func fetchOldSelectedNodes(ctx context.Context, db database.Database, selectorId int) (map[graph.ID]model.AssetGroupCertification, error) {
	oldSelectedNodesMap := make(map[graph.ID]model.AssetGroupCertification)
	if oldSelectedNodes, err := db.GetSelectorNodesBySelectorIds(ctx, selectorId); err != nil {
		return oldSelectedNodesMap, err
	} else {
		for _, node := range oldSelectedNodes {
			oldSelectedNodesMap[node.NodeId] = node.Certified
		}
		return oldSelectedNodesMap, nil
	}
}

// SelectNodes - selects all nodes for a given selector and diffs previous db state for minimal db updates
func SelectNodes(ctx context.Context, db database.Database, graphDb graph.Database, selector model.AssetGroupTagSelector, expansionMethod model.AssetGroupExpansionMethod) error {
	var (
		countInserted int

		certified   = model.AssetGroupCertificationNone
		certifiedBy null.String

		nodeIdsToDelete []graph.ID
		nodeIdsToUpdate []graph.ID
	)

	if selector.AutoCertify.Bool {
		certified = model.AssetGroupCertificationAuto
		certifiedBy = null.StringFrom(model.AssetGroupActorSystem)
	}

	// 1. Grab the graph nodes
	nodesWithSrcSet := FetchNodesFromSeeds(ctx, graphDb, selector.Seeds, expansionMethod, -1)
	// 2. Grab the already selected nodes
	if oldSelectedNodesByNodeId, err := fetchOldSelectedNodes(ctx, db, selector.ID); err != nil {
		return err
	} else {
		// 3. Range the graph nodes and insert any that haven't been inserted yet, mark for update any that need updating, pare down the existing map for future deleting
		for id, node := range nodesWithSrcSet {
			// Missing, insert the record
			if oldCert, ok := oldSelectedNodesByNodeId[id]; !ok {
				if err = db.InsertSelectorNode(ctx, selector.ID, id, certified, certifiedBy, node.Source); err != nil {
					return err
				}
				countInserted++
				// Auto certify is enabled but this node hasn't been certified, certify it
			} else if selector.AutoCertify.Bool && oldCert == model.AssetGroupCertificationNone {
				nodeIdsToUpdate = append(nodeIdsToUpdate, id)
				delete(oldSelectedNodesByNodeId, id)
			} else {
				delete(oldSelectedNodesByNodeId, id)
			}
		}

		// Update the selected nodes that need updating
		if len(nodeIdsToUpdate) > 0 {
			for _, nodeId := range nodeIdsToUpdate {
				if err = db.UpdateSelectorNodesByNodeId(ctx, selector.ID, certified, certifiedBy, nodeId); err != nil {
					return err
				}
			}
		}

		// Delete the selected nodes that need to be deleted
		if len(oldSelectedNodesByNodeId) > 0 {
			for nodeId := range oldSelectedNodesByNodeId {
				if err = db.DeleteSelectorNodesByNodeId(ctx, selector.ID, nodeId); err != nil {
					return err
				}
			}
		}

		slog.Info("AGT: Completed selecting", "selector", selector.Name, "countTotal", len(nodesWithSrcSet), "countInserted", countInserted, "countUpdated", len(nodeIdsToUpdate), "countDeleted", len(nodeIdsToDelete))
	}
	return nil
}

// selectAssetGroupNodes - concurrently selects all nodes for all tags
func selectAssetGroupNodes(ctx context.Context, db database.Database, graphDb graph.Database) error {
	defer measure.ContextMeasure(ctx, slog.LevelInfo, "Finished selecting asset group nodes via new selectors")()

	if tags, err := db.GetAssetGroupTagForSelection(ctx); err != nil {
		return err
	} else {
		for _, tag := range tags {
			if selectors, err := db.GetAssetGroupTagSelectorsByTagId(ctx, tag.ID, model.SQLFilter{}, model.SQLFilter{}); err != nil {
				return err
			} else {
				var (
					disabledSelectorIds []int
					wg                  = sync.WaitGroup{}
				)

				// Spawn N (# of selectors) goroutines for each tag for maximum speed.
				// We are relying on connection pools to negotiate any contention here.
				for _, selector := range selectors {
					if !selector.DisabledAt.Time.IsZero() {
						disabledSelectorIds = append(disabledSelectorIds, selector.ID)
						continue
					}

					// Parallelize the selection of nodes
					wg.Add(1)
					go func() {
						defer wg.Done()
						if err = SelectNodes(ctx, db, graphDb, selector, tag.GetExpansionMethod()); err != nil {
							slog.Error("AGT: Error selecting nodes", "selector", selector, "err", err)
						}
					}()
				}
				wg.Wait()
				// Remove any disabled selector nodes
				if len(disabledSelectorIds) > 0 {
					err = db.DeleteSelectorNodesBySelectorIds(ctx, disabledSelectorIds...)
				}
			}
		}
	}
	return nil
}

// tagAssetGroupNodesForTag - tags all nodes for a given tag and diffs previous db state for minimal db updates
func tagAssetGroupNodesForTag(ctx context.Context, db database.Database, graphDb graph.Database, tag model.AssetGroupTag, additionalFilters ...graph.Criteria) error {
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

		for _, selector := range selectors {
			selectorIds = append(selectorIds, selector.ID)
		}

		// 1. Fetch the selected nodes for this label
		if selectedNodes, err = db.GetSelectorNodesBySelectorIds(ctx, selectorIds...); err != nil {
			return err
		} else if err = graphDb.WriteTransaction(ctx, func(tx graph.Transaction) error {
			filters := []graph.Criteria{query.Kind(query.Node(), tagKind)}
			if additionalFilters != nil {
				filters = append(filters, additionalFilters...)
			}

			// 2. Fetch already tagged nodes
			if oldTaggedNodeSet, err := ops.FetchNodeSet(tx.Nodes().Filter(query.And(filters...))); err != nil {
				return err
			} else {
				oldTaggedNodes = oldTaggedNodeSet.IDBitmap()
			}

			// 3. Diff the sets filling the respective sets for later db updates
			for _, nodeDb := range selectedNodes {
				if !nodesSeen.Contains(nodeDb.NodeId.Uint64()) {
					// Skip any that are not certified when tag requires certification or are selected by disabled selectors
					if tag.RequireCertify.Bool && nodeDb.Certified <= 0 {
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

// tagAssetGroupNodes - concurrently tags all nodes for all tags
func tagAssetGroupNodes(ctx context.Context, db database.Database, graphDb graph.Database, additionalFilters ...graph.Criteria) error {
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
			wg.Add(1)
			go func() {
				defer wg.Done()
				if err = tagAssetGroupNodesForTag(ctx, db, graphDb, tag, additionalFilters...); err != nil {
					slog.Error("AGT: Error tagging nodes", tag.ToType(), tag, "err", err)
				}
			}()
		}

		// Process the tier tagging synchronously
		for _, tier := range tiersOrdered {
			if err := tagAssetGroupNodesForTag(ctx, db, graphDb, tier, additionalFilters...); err != nil {
				slog.Error("AGT: Error tagging nodes", "tier", tier, "err", err)
			}
		}

		// Wait for labels to finish
		wg.Wait()
	}
	return nil
}

func clearAssetGroupTags(ctx context.Context, db database.Database, graphDb graph.Database) error {
	if tags, err := db.GetAssetGroupTags(ctx, model.SQLFilter{}); err != nil {
		return err
	} else {
		for _, tag := range tags {
			tagKind := tag.ToKind()
			if err = graphDb.WriteTransaction(ctx, func(tx graph.Transaction) error {
				if taggedNodeSet, err := ops.FetchNodeSet(tx.Nodes().Filter(query.Kind(query.Node(), tagKind))); err != nil {
					return err
				} else {
					for _, node := range taggedNodeSet {
						node.DeleteKinds(tagKind)
						if err := tx.UpdateNode(node); err != nil {
							slog.WarnContext(ctx, "AGT: Error cleaning node", slog.String("nodeId", node.ID.String()), slog.String("err", err.Error()))
						}
					}
				}

				return nil
			}); err != nil {
				return err
			}
		}
	}
	return nil
}

// TODO Cleanup tieringEnabled after Tiering GA
func TagAssetGroupsAndTierZero(ctx context.Context, db database.Database, graphDb graph.Database, additionalFilters ...graph.Criteria) []error {
	var errors []error

	if appcfg.GetTieringEnabled(ctx, db) {
		// Tiering enabled, we don't want system tags present
		if err := clearSystemTags(ctx, graphDb, additionalFilters...); err != nil {
			slog.Error(fmt.Sprintf("AGT: wiping old system tags: %v", err))
			errors = append(errors, err)
		}
		if err := selectAssetGroupNodes(ctx, db, graphDb); err != nil {
			slog.Error(fmt.Sprintf("AGT: selecting failed: %v", err))
			errors = append(errors, err)
		}

		if err := tagAssetGroupNodes(ctx, db, graphDb, additionalFilters...); err != nil {
			slog.Error(fmt.Sprintf("AGT: tagging failed: %v", err))
			errors = append(errors, err)
		}
	} else {
		// Tiering disabled, we don't want nodes with tagged kinds
		if err := clearAssetGroupTags(ctx, db, graphDb); err != nil {
			slog.Error(fmt.Sprintf("AGT: clearing tags failed: %v", err))
			errors = append(errors, err)
		}

		if err := clearSystemTags(ctx, graphDb, additionalFilters...); err != nil {
			slog.Error(fmt.Sprintf("Failed clearing system tags: %v", err))
			errors = append(errors, err)
		} else if err := updateAssetGroupIsolationTags(ctx, db, graphDb); err != nil {
			slog.Error(fmt.Sprintf("Failed updating asset group isolation tags: %v", err))
			errors = append(errors, err)
		}

		if err := tagActiveDirectoryTierZero(ctx, db, graphDb); err != nil {
			slog.Error(fmt.Sprintf("Failed tagging Active Directory attack path roots: %v", err))
			errors = append(errors, err)
		}

		if err := parallelTagAzureTierZero(ctx, graphDb); err != nil {
			slog.Error(fmt.Sprintf("Failed tagging Azure attack path roots: %v", err))
			errors = append(errors, err)
		}
	}

	return errors
}
