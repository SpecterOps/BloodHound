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
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/specterops/bloodhound/cmd/api/src/database"
	"github.com/specterops/bloodhound/cmd/api/src/database/types/null"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/cmd/api/src/model/appcfg"
	"github.com/specterops/bloodhound/packages/go/bhlog/attr"
	"github.com/specterops/bloodhound/packages/go/bhlog/measure"
	"github.com/specterops/bloodhound/packages/go/graphschema/ad"
	"github.com/specterops/bloodhound/packages/go/graphschema/azure"
	"github.com/specterops/bloodhound/packages/go/graphschema/common"
	"github.com/specterops/dawgs/cardinality"
	"github.com/specterops/dawgs/graph"
	"github.com/specterops/dawgs/ops"
	"github.com/specterops/dawgs/query"
	"github.com/specterops/dawgs/traversal"
	"github.com/specterops/dawgs/util/channels"
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

type errorsWithLock struct {
	errs []error
	lock *sync.Mutex
}

func newErrorsWithLock() *errorsWithLock {
	return &errorsWithLock{
		errs: []error{},
		lock: &sync.Mutex{},
	}
}

func (s *errorsWithLock) Append(err ...error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.errs = append(s.errs, err...)
}

func (s *errorsWithLock) Errors() []error {
	s.lock.Lock()
	defer s.lock.Unlock()

	return s.errs
}

// FetchNodesFromSeeds fetches all seed nodes along with any child or parent nodes via known expansion paths
func FetchNodesFromSeeds(ctx context.Context, agtParameters appcfg.AGTParameters, graphDb graph.Database, seeds []model.SelectorSeed, expansionMethod model.AssetGroupExpansionMethod, limit int) (nodeWithSourceSet, []error) {
	var (
		seedNodes = make(nodeWithSourceSet)
		result    = make(nodeWithSourceSet)
		errs      []error
	)
	// Then we grab the nodes that should be selected
	for _, seed := range seeds {
		_ = graphDb.ReadTransaction(ctx, func(tx graph.Transaction) error {
			switch seed.Type {
			case model.SelectorTypeObjectId:
				if node, err := tx.Nodes().Filter(query.Equals(query.NodeProperty(common.ObjectID.String()), seed.Value)).First(); err != nil {
					slog.WarnContext(
						ctx,
						"AGT: Fetch Object ID Err",
						slog.String("objectid", seed.Value),
						attr.Error(err),
					)
					if !graph.IsErrNotFound(err) { // Don't halt analysis for not found objectids
						errs = append(errs, err)
					}
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
					slog.WarnContext(
						ctx,
						"AGT: Fetch Cypher Err",
						slog.String("cypher_query", seed.Value),
						attr.Error(err),
					)
					errs = append(errs, err)
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
				slog.WarnContext(ctx, "AGT: Unsupported selector type", slog.Int("type", int(seed.Type)))
			}
			return nil
		})
	}

	if expansionMethod == model.AssetGroupExpansionMethodNone || result.LimitReached(limit) || len(result) == 0 {
		return result, errs
	}

	if expansionMethod == model.AssetGroupExpansionMethodAll || expansionMethod == model.AssetGroupExpansionMethodChildren {
		if collected, fetchErrs := fetchAllChildNodes(ctx, agtParameters, graphDb, seedNodes, result, limit); len(fetchErrs) > 0 {
			errs = append(errs, fetchErrs...)
		} else {
			if result.LimitReached(limit) {
				return result, errs
			}

			// Add any newly collected child nodes to seeds for optional parent expansion below
			for _, node := range collected {
				seedNodes.AddIfNotExists(node)
			}
		}
	}

	if expansionMethod == model.AssetGroupExpansionMethodAll || expansionMethod == model.AssetGroupExpansionMethodParents {
		if fetchErrs := fetchParentNodes(ctx, agtParameters, graphDb, seedNodes, result, limit); len(fetchErrs) > 0 {
			errs = append(errs, fetchErrs...)
		}
	}

	return result, errs
}

// fetchChildNodes - fetches all children for a single node and submits any found to supplied collector ch
func fetchChildNodes(ctx context.Context, tx traversal.Traversal, node *graph.Node, ch chan<- *nodeWithSource) error {
	var pattern traversal.PatternContinuation

	// This protects any nodes that are not AD or Azure from expansion
	if !node.Kinds.ContainsOneOf(ad.Entity, azure.Entity) {
		return nil
	}

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
		// MATCH (n:AZRole)<-[:AZHasRole|AZRoleEligible]-(m:AZBase) RETURN m
		pattern = traversal.NewPattern().InboundWithDepth(0, 1, query.And(
			query.KindIn(query.Relationship(), azure.HasRole, azure.AZRoleEligible),
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
func fetchAllChildNodes(ctx context.Context, agtParameters appcfg.AGTParameters, db graph.Database, seedNodes nodeWithSourceSet, result nodeWithSourceSet, limit int) ([]*nodeWithSource, []error) {
	var (
		wg              = sync.WaitGroup{}
		queueLen        = &atomic.Int64{}
		chCtx, doneFunc = context.WithCancel(ctx)

		sendCh, getCh = channels.BufferedPipe[*nodeWithSource](chCtx)
		collectorCh   = make(chan *nodeWithSource)

		traversalInst = traversal.New(db, agtParameters.DAWGsWorkerLimit)
		collected     []*nodeWithSource

		// Due to concurrency, to keep track of errors, mutex is required
		errs = newErrorsWithLock()
	)
	defer doneFunc()
	// Close the send channel to the buffered pipe
	defer close(sendCh)

	// Spin out some workers, capped to prevent exhausting pg connection pool
	for range agtParameters.ExpansionWorkerLimit {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				// Block here until we receive a node to fetch child nodes
				if nodeToExpand, ok := channels.Receive(chCtx, getCh); !ok {
					return
				} else {
					// Fetch child nodes for this node and send any collected to the collector
					if err := fetchChildNodes(chCtx, traversalInst, nodeToExpand.Node, collectorCh); err != nil {
						slog.ErrorContext(
							ctx,
							"AGT: error fetching child nodes",
							slog.Uint64("node", nodeToExpand.ID.Uint64()),
							attr.Error(err),
						)
						errs.Append(err)
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
	for _, seedNode := range seedNodes {
		queueLen.Add(1)
		channels.Submit(chCtx, sendCh, seedNode)
	}

	wg.Wait() // Wait for workers to process all nodes

	return collected, errs.Errors()
}

// fetchADParentNodes -  fetches all parents for a single active directory node and submits any found to supplied collector ch
func fetchADParentNodes(ctx context.Context, tx traversal.Traversal, node *graph.Node, ch chan<- *nodeWithSource) error {
	// This protects any nodes that are not AD from expansion
	if !node.Kinds.ContainsOneOf(ad.Entity) {
		return nil
	}

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
	// This protects any nodes that are not Azure from expansion
	if !node.Kinds.ContainsOneOf(azure.Entity) {
		return nil
	}

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
func fetchParentNodes(ctx context.Context, agtParameters appcfg.AGTParameters, db graph.Database, seedNodes nodeWithSourceSet, result nodeWithSourceSet, limit int) []error {
	var (
		wg = sync.WaitGroup{}

		ctxWithCancel, doneFunc = context.WithCancel(ctx)
		sendCh, getCh           = channels.BufferedPipe[*nodeWithSource](ctxWithCancel)
		collectorCh             = make(chan *nodeWithSource)

		// Due to concurrency, to keep track of errors, mutex is required
		errs = newErrorsWithLock()

		traversalInst = traversal.New(db, agtParameters.DAWGsWorkerLimit)
	)
	// Expand to parent nodes as needed
	defer doneFunc()

	// Spin out some workers, capped to prevent exhausting pg connection pool
	for range agtParameters.ExpansionWorkerLimit {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				// Block here until we receive a node to fetch parent nodes
				if nodeToExpand, ok := channels.Receive(ctxWithCancel, getCh); !ok {
					return
				} else {
					if nodeToExpand.Kinds.ContainsOneOf(ad.Entity) {
						if err := fetchADParentNodes(ctxWithCancel, traversalInst, nodeToExpand.Node, collectorCh); err != nil {
							slog.ErrorContext(
								ctx,
								"AGT: error fetching active directory parent nodes",
								slog.Uint64("node", nodeToExpand.ID.Uint64()),
								attr.Error(err),
							)
							errs.Append(err)
						}
					} else if nodeToExpand.Kinds.ContainsOneOf(azure.Entity) {
						if err := fetchAzureParentNodes(ctxWithCancel, traversalInst, nodeToExpand.Node, collectorCh); err != nil {
							slog.ErrorContext(
								ctx,
								"AGT: error fetching azure parent nodes",
								slog.Uint64("node", nodeToExpand.ID.Uint64()),
								attr.Error(err),
							)
							errs.Append(err)
						}
					}
				}
			}
		}()
	}

	go func() {
		// This will wait to close the collector channel and release the below blocking for loop once the workers have finished
		wg.Wait()
		close(collectorCh)
	}()

	// Fill queue with seed nodes
	for _, seedNode := range seedNodes {
		channels.Submit(ctxWithCancel, sendCh, seedNode)
	}
	// Close the queue channel once filled, this will cause the worker goroutines to finish once the queue is emptied
	close(sendCh)

	// This will block and collect all parent nodes until channel is closed
	for nodeWithSrc := range collectorCh {
		if result.AddIfNotExists(nodeWithSrc) && result.LimitReached(limit) {
			doneFunc()
		}
	}

	return errs.Errors()
}

// fetchOldSelectedNodes - fetches the currently selected nodes and assembles a map lookup for minimal memory footprint
func fetchOldSelectedNodes(ctx context.Context, db database.Database, selectorId int) (map[graph.ID]model.AssetGroupSelectorNode, error) {
	oldSelectedNodesMap := make(map[graph.ID]model.AssetGroupSelectorNode)
	if oldSelectedNodes, err := db.GetSelectorNodesBySelectorIds(ctx, selectorId); err != nil {
		return oldSelectedNodesMap, err
	} else {
		for _, node := range oldSelectedNodes {
			oldSelectedNodesMap[node.NodeId] = node
		}
		return oldSelectedNodesMap, nil
	}
}

// SelectNodes - selects all nodes for a given selector and diffs previous db state for minimal db updates
func SelectNodes(ctx context.Context, db database.Database, agtParameters appcfg.AGTParameters, graphDb graph.Database, selector model.AssetGroupTagSelector, expansionMethod model.AssetGroupExpansionMethod) []error {
	defer measure.ContextMeasure(ctx, slog.LevelDebug, "Finished selecting nodes", slog.String("selector", strconv.Itoa(selector.ID)))()

	var (
		countInserted int
		nodesToUpdate []model.AssetGroupSelectorNode
		errs          []error
	)

	// 1. Grab the graph nodes
	if nodesWithSrcSet, fetchErrs := FetchNodesFromSeeds(ctx, agtParameters, graphDb, selector.Seeds, expansionMethod, -1); len(fetchErrs) > 0 {
		for _, err := range fetchErrs {
			errs = append(errs, fmt.Errorf("selector %d - %s fetch failure: %w", selector.ID, selector.Name, err))
		}
	} else if oldSelectedNodesByNodeId, err := fetchOldSelectedNodes(ctx, db, selector.ID); err != nil {
		// 2. Grab the already selected nodes
		errs = append(errs, err)
	} else {
		// 3. Range the graph nodes and insert any that haven't been inserted yet, mark for update any that need updating, pare down the existing map for future deleting
		for id, node := range nodesWithSrcSet {
			var (
				certified                                 = model.AssetGroupCertificationPending
				certifiedBy                               null.String
				primaryKind, displayName, objectId, envId = model.GetAssetGroupMemberProperties(node.Node)
			)

			if (selector.AutoCertify == model.SelectorAutoCertifyMethodSeedsOnly && node.Source == model.AssetGroupSelectorNodeSourceSeed) || selector.AutoCertify == model.SelectorAutoCertifyMethodAllMembers {
				certified = model.AssetGroupCertificationAuto
				certifiedBy = null.StringFrom(model.AssetGroupActorBloodHound)
			}

			// Missing, insert the record
			if oldNode, ok := oldSelectedNodesByNodeId[id]; !ok {
				if err = db.InsertSelectorNode(ctx, selector.AssetGroupTagId, selector.ID, id, certified, certifiedBy, node.Source, primaryKind, envId, objectId, displayName); err != nil {
					errs = append(errs, err)
				}
				countInserted++
				// Auto certify is enabled but this node hasn't been certified, certify it. Further - update any out of sync node properties
			} else if ((oldNode.Certified != model.AssetGroupCertificationRevoked && oldNode.Certified != model.AssetGroupCertificationManual) && certified != oldNode.Certified) ||
				oldNode.NodeName != displayName ||
				oldNode.NodePrimaryKind != primaryKind ||
				oldNode.NodeEnvironmentId != envId ||
				oldNode.NodeObjectId != objectId {
				nodesToUpdate = append(nodesToUpdate, oldNode)
				delete(oldSelectedNodesByNodeId, id)
			} else {
				delete(oldSelectedNodesByNodeId, id)
			}
		}

		// Update the selected nodes that need updating
		if len(nodesToUpdate) > 0 {
			for _, oldSelectorNode := range nodesToUpdate {
				var (
					certified   = model.AssetGroupCertificationPending
					certifiedBy null.String
				)

				// Protect property updates from overwriting existing manual certifications
				if oldSelectorNode.Certified == model.AssetGroupCertificationRevoked || oldSelectorNode.Certified == model.AssetGroupCertificationManual {
					certified = oldSelectorNode.Certified
					certifiedBy = oldSelectorNode.CertifiedBy
				} else if oldSelectorNode.Certified == model.AssetGroupCertificationPending && ((selector.AutoCertify == model.SelectorAutoCertifyMethodSeedsOnly && oldSelectorNode.Source == model.AssetGroupSelectorNodeSourceSeed) || selector.AutoCertify == model.SelectorAutoCertifyMethodAllMembers) {
					certified = model.AssetGroupCertificationAuto
					certifiedBy = null.StringFrom(model.AssetGroupActorBloodHound)
				}

				if graphNode, ok := nodesWithSrcSet[oldSelectorNode.NodeId]; !ok {
					// todo: maybe grab it from graph manually in this case?
					slog.WarnContext(ctx, "AGT: selected node for update missing graph node...skipping update to protect data integrity", slog.Uint64("node_id", oldSelectorNode.NodeId.Uint64()))
				} else {
					primaryKind, displayName, objectId, envId := model.GetAssetGroupMemberProperties(graphNode.Node)
					if err = db.UpdateSelectorNodesByNodeId(ctx, selector.AssetGroupTagId, selector.ID, oldSelectorNode.NodeId, certified, certifiedBy, primaryKind, envId, objectId, displayName); err != nil {
						errs = append(errs, err)
					}
				}
			}
		}

		// Delete the selected nodes that need to be deleted
		if len(oldSelectedNodesByNodeId) > 0 {
			for nodeId := range oldSelectedNodesByNodeId {
				if err = db.DeleteSelectorNodesByNodeId(ctx, selector.ID, nodeId); err != nil {
					errs = append(errs, err)
				}
			}
		}

		slog.InfoContext(
			ctx,
			"AGT: Completed selecting",
			slog.String("selector", selector.Name),
			slog.Int("count_total", len(nodesWithSrcSet)),
			slog.Int("count_inserted", countInserted),
			slog.Int("count_updated", len(nodesToUpdate)),
			slog.Int("count_deleted", len(oldSelectedNodesByNodeId)),
		)
	}

	return errs
}

// selectAssetGroupNodes - concurrently selects all nodes for all tags
func selectAssetGroupNodes(ctx context.Context, db database.Database, graphDb graph.Database) []error {
	defer measure.ContextMeasure(
		ctx,
		slog.LevelInfo,
		"Finished selecting agt nodes",
		attr.Namespace("analysis"),
		attr.Function("selectAssetGroupNodes"),
		attr.Scope("process"),
	)()

	// Due to concurrency, to keep track of errors, mutex is required
	errs := newErrorsWithLock()

	if tags, err := db.GetAssetGroupTagForSelection(ctx); err != nil {
		errs.Append(err)
	} else {
		agtParameters := appcfg.GetAGTParameters(ctx, db)
		slog.InfoContext(ctx,
			"AGT: Pooling parameters",
			slog.String("selector_worker_limit", strconv.Itoa(agtParameters.SelectorWorkerLimit)),
			slog.String("expansion_worker_limit", strconv.Itoa(agtParameters.ExpansionWorkerLimit)),
			slog.String("dawgs_worker_limit", strconv.Itoa(agtParameters.DAWGsWorkerLimit)),
			slog.String("agt_max_conn", strconv.Itoa(agtParameters.SelectorWorkerLimit*agtParameters.ExpansionWorkerLimit*agtParameters.DAWGsWorkerLimit)),
		)

		var (
			disabledSelectorIds []int
			sendCh, getCh       = channels.BufferedPipe[model.AssetGroupTagSelector](ctx)
			wg                  = sync.WaitGroup{}
			expansionByTagId    = make(map[int]model.AssetGroupExpansionMethod)
		)

		// Build expansion map
		for _, tag := range tags {
			expansionByTagId[tag.ID] = tag.GetExpansionMethod()
		}

		// Parallelize the selection of nodes
		// Spin out some workers, capped to prevent exhausting pg connection pool
		for range agtParameters.SelectorWorkerLimit {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for {
					// Block here until we receive a selector
					if selector, ok := channels.Receive(ctx, getCh); !ok {
						return
					} else {
						if selectNodeErrors := SelectNodes(ctx, db, agtParameters, graphDb, selector, expansionByTagId[selector.AssetGroupTagId]); len(selectNodeErrors) > 0 {
							errs.Append(selectNodeErrors...)
						}
					}
				}
			}()
		}

		for _, tag := range tags {
			if selectors, _, err := db.GetAssetGroupTagSelectorsByTagId(ctx, tag.ID); err != nil {
				errs.Append(err)
			} else {
				// Fill worker queue with selectors
				for _, selector := range selectors {
					if !selector.DisabledAt.Time.IsZero() {
						disabledSelectorIds = append(disabledSelectorIds, selector.ID)
						continue
					}
					channels.Submit(ctx, sendCh, selector)
				}
			}
		}

		// Close the queue channel once filled, this will cause the worker goroutines to finish once the queue is emptied
		close(sendCh)

		// Remove any disabled selector nodes while waiting for selectors to select
		if len(disabledSelectorIds) > 0 {
			if err = db.DeleteSelectorNodesBySelectorIds(ctx, disabledSelectorIds...); err != nil {
				errs.Append(err)
			}
		}

		// Wait for selection to finish
		wg.Wait()
	}

	return errs.Errors()
}

// tagAssetGroupNodesForTag - tags all nodes for a given tag and diffs previous db state for minimal db updates
func tagAssetGroupNodesForTag(ctx context.Context, db database.Database, graphDb graph.Database, tag model.AssetGroupTag, nodesSeen cardinality.Duplex[uint64], additionalFilters ...graph.Criteria) error {
	if selectors, _, err := db.GetAssetGroupTagSelectorsByTagId(ctx, tag.ID); err != nil {
		return err
	} else {
		var (
			countTotal    int
			selectorIds   []int
			selectedNodes []model.AssetGroupSelectorNode

			tagKind = tag.ToKind()

			oldTaggedNodes         = cardinality.NewBitmap64()
			newTaggedNodes         = cardinality.NewBitmap64()
			missingSystemTagsNodes = cardinality.NewBitmap64()
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

				// 3. Diff the sets filling the respective sets for later db updates
				for _, nodeDb := range selectedNodes {
					if !nodesSeen.Contains(nodeDb.NodeId.Uint64()) {
						// Skip any that are not certified when tag requires certification or are selected by disabled selectors
						if tag.RequireCertify.Bool && nodeDb.Certified <= model.AssetGroupCertificationRevoked {
							continue
						}

						// If the id is not present, we must queue it for tagging
						if !oldTaggedNodes.Contains(nodeDb.NodeId.Uint64()) {
							newTaggedNodes.Add(nodeDb.NodeId.Uint64())
						} else {
							// TODO Cleanup system tagging after Tiering GA
							if tag.Type == model.AssetGroupTagTypeTier && tag.Position.ValueOrZero() == model.AssetGroupTierZeroPosition && oldTaggedNodeSet.Get(nodeDb.NodeId).Properties.Get(common.SystemTags.String()).IsNil() {
								missingSystemTagsNodes.Add(nodeDb.NodeId.Uint64())
							}

							// If it is present, we don't need to update anything and will remove tags from any nodes left in this bitmap
							oldTaggedNodes.Remove(nodeDb.NodeId.Uint64())
						}
						// Once a node is processed, we can skip future duplicates that might be selected by other selectors
						nodesSeen.Add(nodeDb.NodeId.Uint64())
						countTotal++
					}
				}
			}

			// 4. Tag the new nodes
			newTaggedNodes.Each(func(nodeId uint64) bool {
				node := &graph.Node{ID: graph.ID(nodeId), Properties: graph.NewProperties()}
				// Temporarily include this for backwards compatibility with old asset group system
				if tag.Type == model.AssetGroupTagTypeTier && tag.Position.ValueOrZero() == model.AssetGroupTierZeroPosition {
					node.Properties.Set(common.SystemTags.String(), ad.AdminTierZero)
				}

				node.AddKinds(tagKind)
				err = tx.UpdateNode(node)
				return err == nil
			})
			if err != nil {
				return err
			}
			/// TODO Cleanup system tagging after Tiering GA
			// 4.5 Update already tagged nodes missing system tags
			missingSystemTagsNodes.Each(func(nodeId uint64) bool {
				node := &graph.Node{ID: graph.ID(nodeId), Properties: graph.NewProperties()}
				node.Properties.Set(common.SystemTags.String(), ad.AdminTierZero)

				err = tx.UpdateNode(node)
				return err == nil
			})
			if err != nil {
				return err
			}

			// 5. Remove the old nodes
			oldTaggedNodes.Each(func(nodeId uint64) bool {
				node := &graph.Node{ID: graph.ID(nodeId), Properties: graph.NewProperties()}
				node.DeleteKinds(tagKind)
				err = tx.UpdateNode(node)
				return err == nil
			})
			if err != nil {
				return err
			}

			return nil
		}); err != nil {
			return err
		}

		slog.InfoContext(
			ctx,
			"AGT: Completed tagging",
			slog.String("tag_type", tag.ToType()),
			slog.String("tag_name", tag.Name),
			slog.Int("total", countTotal),
			slog.Uint64("tagged", newTaggedNodes.Cardinality()),
			slog.Uint64("untagged", oldTaggedNodes.Cardinality()),
		)
	}
	return nil
}

// tagAssetGroupNodes - concurrently tags all nodes for all tags
func tagAssetGroupNodes(ctx context.Context, db database.Database, graphDb graph.Database, additionalFilters ...graph.Criteria) []error {
	defer measure.ContextMeasure(
		ctx,
		slog.LevelInfo,
		"Finished tagging asset group nodes",
		attr.Namespace("analysis"),
		attr.Function("tagAssetGroupNodes"),
		attr.Scope("process"),
	)()

	// Due to concurrency, to keep track of errors, mutex is required
	errs := newErrorsWithLock()

	if tags, err := db.GetAssetGroupTagForSelection(ctx); err != nil {
		errs.Append(err)
	} else {
		// Tiers are hierarchical and must be handled synchronously while labels can be tagged in parallel
		var (
			labelsOrOwned []model.AssetGroupTag
			tiersOrdered  []model.AssetGroupTag
			nodesSeen     = cardinality.NewBitmap64()
		)
		for _, tag := range tags {
			switch tag.Type {
			case model.AssetGroupTagTypeTier:
				tiersOrdered = append(tiersOrdered, tag)
			case model.AssetGroupTagTypeLabel, model.AssetGroupTagTypeOwned:
				labelsOrOwned = append(labelsOrOwned, tag)
			default:
				slog.WarnContext(
					ctx,
					"AGT: Tag type is not supported",
					slog.Int("tag_type", int(tag.Type)),
					slog.Any("tag", tag),
				)
			}
		}

		// Fire off the label tagging
		wg := sync.WaitGroup{}
		for _, tag := range labelsOrOwned {
			// Parallelize the tagging of label nodes
			wg.Add(1)
			go func() {
				defer wg.Done()
				// Nodes can contain multiple labels therefore there is no need to exclude here
				if err := tagAssetGroupNodesForTag(ctx, db, graphDb, tag, cardinality.NewBitmap64(), additionalFilters...); err != nil {
					errs.Append(err)
				}
			}()
		}

		// Process the tier tagging synchronously
		for _, tier := range tiersOrdered {
			// Nodes cannot contain multiple tiers therefore the nodesSeen serves as a running exclusion bitmap
			if err := tagAssetGroupNodesForTag(ctx, db, graphDb, tier, nodesSeen, additionalFilters...); err != nil {
				errs.Append(err)
			}
		}

		// Wait for labels to finish
		wg.Wait()
	}

	return errs.Errors()
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
							slog.WarnContext(
								ctx,
								"AGT: Error cleaning node",
								slog.String("node_id", node.ID.String()),
								attr.Error(err),
							)
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

func ClearAssetGroupTagNodeSet(ctx context.Context, graphDb graph.Database, assetGroupTag model.AssetGroupTag) error {
	tagKind := assetGroupTag.ToKind()
	if err := graphDb.WriteTransaction(ctx, func(tx graph.Transaction) error {
		if taggedNodeSet, err := ops.FetchNodeSet(tx.Nodes().Filter(query.Kind(query.Node(), tagKind))); err != nil {
			return err
		} else {
			for _, node := range taggedNodeSet {
				node.DeleteKinds(tagKind)
				if err = tx.UpdateNode(node); err != nil {
					return err
				}
			}
		}

		return nil
	}); err != nil {
		return err
	}

	return nil
}

// ClearAssetGroupHistoryRecords Truncate the asset group history table to the rolling window
func ClearAssetGroupHistoryRecords(ctx context.Context, db database.Database) {
	defer measure.ContextMeasure(
		ctx,
		slog.LevelInfo,
		"Cleared Asset Group History",
		attr.Namespace("analysis"),
		attr.Function("ClearAssetGroupHistoryRecords"),
		attr.Scope("step"),
	)()
	if recordsDeletedCount, err := db.DeleteAssetGroupHistoryRecordsByCreatedDate(ctx, time.Now().UTC().AddDate(0, 0, -1*model.AssetGroupHistoryRecordRollingWindow)); err != nil {
		slog.WarnContext(
			ctx,
			"AGT: ClearAssetGroupHistoryRecords error",
			slog.String("count_deleted", strconv.FormatInt(recordsDeletedCount, 10)),
			attr.Error(err),
		)
	} else {
		slog.InfoContext(
			ctx,
			"AGT: ClearAssetGroupHistoryRecords",
			slog.String("count_deleted", strconv.FormatInt(recordsDeletedCount, 10)),
		)
	}
}

func migrateCustomObjectIdSelectorNames(ctx context.Context, db database.Database, graphDb graph.Database) error {
	if selectorsToMigrate, err := db.GetCustomAssetGroupTagSelectorsToMigrate(ctx); err != nil {
		return err
	} else {
		var countUpdated, countSkipped int

		for _, selector := range selectorsToMigrate {
			if len(selector.Seeds) > 1 {
				slog.WarnContext(ctx, "AGT: customSelectorMigration - Captured incorrect selector to migrate", slog.Any("selector", selector))
				continue
			} else if len(selector.Seeds) == 1 {
				if err = graphDb.ReadTransaction(ctx, func(tx graph.Transaction) error {
					if node, err := tx.Nodes().Filter(query.Equals(query.NodeProperty(common.ObjectID.String()), selector.Seeds[0].Value)).First(); err != nil {
						slog.DebugContext(
							ctx,
							"AGT: customSelectorMigration - Fetch objectid err",
							slog.String("objectid", selector.Seeds[0].Value),
							attr.Error(err),
						)
						countSkipped++
					} else {
						name, _ := node.Properties.GetWithFallback(common.Name.String(), "", common.DisplayName.String()).String()
						if name == "" {
							slog.DebugContext(ctx, "AGT: customSelectorMigration - No name found for node, skipping", slog.String("objectid", selector.Seeds[0].Value))
							countSkipped++
							return nil
						}
						selector.Name = name
						if _, err := db.UpdateAssetGroupTagSelector(ctx, model.AssetGroupActorBloodHound, "", selector); err != nil {
							slog.WarnContext(ctx, "AGT: customSelectorMigration - Failed to migrate custom selector name", slog.Any("selector", selector))
							countSkipped++
						}
						countUpdated++
					}
					return nil
				}); err != nil {
					return err
				}
			}
		}
		if len(selectorsToMigrate) > 0 {
			slog.InfoContext(
				ctx,
				"AGT: customSelectorMigration - Migrated custom selectors",
				slog.Int("count_found", len(selectorsToMigrate)),
				slog.Int("count_updated", countUpdated),
				slog.Int("count_skipped", countSkipped),
			)
		}
	}

	return nil
}

// TODO Cleanup tieringEnabled after Tiering GA
func TagAssetGroupsAndTierZero(ctx context.Context, db database.Database, graphDb graph.Database, additionalFilters ...graph.Criteria) []error {
	defer measure.ContextMeasure(
		ctx,
		slog.LevelInfo,
		"Tag Asset Groups and Tier Zero",
		attr.Namespace("analysis"),
		attr.Function("TagAssetGroupsAndTierZero"),
		attr.Scope("step"),
	)()

	var errs []error

	if appcfg.GetTieringEnabled(ctx, db) {
		// Tiering enabled, we don't want system tags present
		if err := clearSystemTags(ctx, graphDb, additionalFilters...); err != nil {
			slog.ErrorContext(ctx, "AGT: wiping old system tags", attr.Error(err))
			errs = append(errs, err)
		}

		if err := migrateCustomObjectIdSelectorNames(ctx, db, graphDb); err != nil {
			slog.ErrorContext(ctx, "AGT: migrating custom selector names failed", attr.Error(err))
			errs = append(errs, err)
		}

		if selectErrs := selectAssetGroupNodes(ctx, db, graphDb); len(selectErrs) > 0 {
			errs = append(errs, selectErrs...)
		}

		if tagErrs := tagAssetGroupNodes(ctx, db, graphDb, additionalFilters...); len(tagErrs) > 0 {
			errs = append(errs, tagErrs...)
		}
	} else {
		// Tiering disabled, we don't want nodes with tagged kinds
		if err := clearAssetGroupTags(ctx, db, graphDb); err != nil {
			slog.ErrorContext(ctx, "AGT: clearing tags failed", attr.Error(err))
			errs = append(errs, err)
		}

		if err := clearSystemTags(ctx, graphDb, additionalFilters...); err != nil {
			slog.ErrorContext(ctx, "Failed clearing system tags", attr.Error(err))
			errs = append(errs, err)
		} else if err := updateAssetGroupIsolationTags(ctx, db, graphDb); err != nil {
			slog.ErrorContext(ctx, "Failed updating asset group isolation tags", attr.Error(err))
			errs = append(errs, err)
		}

		if err := tagActiveDirectoryTierZero(ctx, db, graphDb); err != nil {
			slog.ErrorContext(ctx, "Failed tagging Active Directory attack path roots", attr.Error(err))
			errs = append(errs, err)
		}

		if err := parallelTagAzureTierZero(ctx, graphDb); err != nil {
			slog.ErrorContext(ctx, "Failed tagging Azure attack path roots", attr.Error(err))
			errs = append(errs, err)
		}
	}

	return errs
}
