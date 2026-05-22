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
	"fmt"
	"log/slog"

	"github.com/specterops/bloodhound/packages/go/analysis/post"
	"github.com/specterops/bloodhound/packages/go/bhlog/attr"
	"github.com/specterops/bloodhound/packages/go/bhlog/measure"
	"github.com/specterops/bloodhound/packages/go/graphschema/ad"
	"github.com/specterops/dawgs/graph"
	"github.com/specterops/dawgs/ops"
	"github.com/specterops/dawgs/query"
	"github.com/specterops/dawgs/traversal"
)

const gpoMaxDepth = 1024

var gpoPostProcessedEdges = graph.Kinds{
	ad.GPOAppliesTo,
	ad.CanApplyGPO,
}

func PostGPOs(ctx context.Context, db graph.Database) (*post.AtomicPostProcessingStats, error) {
	defer measure.ContextLogAndMeasure(
		ctx,
		slog.LevelInfo,
		"Post-processing GPOs",
		attr.Namespace("analysis"),
		attr.Function("PostGPOs"),
		attr.Scope("process"),
	)()

	var (
		gpoTracker            *post.Tracker
		domainNodes           []*graph.Node
		completedSuccessfully bool
		err                   error
	)

	if err = post.MigrationForDCAPostProcessedEdges(ctx, db, gpoPostProcessedEdges); err != nil {
		return &post.AtomicPostProcessingStats{}, err
	}

	if gpoTracker, err = post.FetchTracker(ctx, db, gpoPostProcessedEdges); err != nil {
		return &post.AtomicPostProcessingStats{}, err
	}

	if domainNodes, err = fetchCollectedDomainNodes(ctx, db); err != nil {
		return &post.AtomicPostProcessingStats{}, err
	}

	sink := post.NewFilteredRelationshipSink(ctx, "GPO Post Processing", db, gpoTracker)
	defer func() {
		if completedSuccessfully {
			sink.Done()
		} else {
			sink.Abort()
		}
	}()

	for _, domainNode := range domainNodes {
		if err = db.ReadTransaction(ctx, func(tx graph.Transaction) error {
			return processGPOs(ctx, tx, sink, domainNode, nil, nil, nil, 0)
		}); err != nil {
			return sink.Stats(), fmt.Errorf("failed processing GPOs for domain %d: %w", domainNode.ID, err)
		}
	}

	completedSuccessfully = true
	return sink.Stats(), nil
}

func processGPOs(ctx context.Context, tx graph.Transaction, sink *post.FilteredRelationshipSink, node *graph.Node, inheritedGPOs, inheritedEnforcedGPOs, inheritedGPOAppliers []graph.ID, depth int) error {
	var (
		applyingGPOs         []graph.ID
		currentGPOAppliers   = appendUniqueGraphIDs(nil, inheritedGPOAppliers...)
		nextEnforcedGPOs     = appendUniqueGraphIDs(nil, inheritedEnforcedGPOs...)
		blocksGPOInheritance bool
	)

	if err := ctx.Err(); err != nil {
		return err
	}

	if depth > gpoMaxDepth {
		return fmt.Errorf("max GPO recursion depth exceeded at node %d", node.ID)
	}

	if node.Kinds.ContainsOneOf(ad.Domain, ad.OU) {
		if additionalGPOAppliers, err := fetchAdditionalGPOAppliers(tx, node.ID, currentGPOAppliers); err != nil {
			return fmt.Errorf("failed fetching GPO appliers on node %d: %w", node.ID, err)
		} else {
			currentGPOAppliers = appendUniqueGraphIDs(currentGPOAppliers, additionalGPOAppliers...)
		}

		if directlyLinkedGPOs, directlyLinkedEnforcedGPOs, err := fetchGPOLinkedDirectly(tx, node.ID); err != nil {
			return fmt.Errorf("failed fetching directly linked GPOs on node %d: %w", node.ID, err)
		} else {
			applyingGPOs = appendUniqueGraphIDs(applyingGPOs, directlyLinkedGPOs...)
			nextEnforcedGPOs = appendUniqueGraphIDs(nextEnforcedGPOs, directlyLinkedEnforcedGPOs...)
		}

		if node.Kinds.ContainsOneOf(ad.OU) {
			var err error
			if blocksGPOInheritance, err = node.Properties.GetOrDefault(ad.BlocksInheritance.String(), false).Bool(); err != nil {
				return fmt.Errorf("failed reading %s on node %d: %w", ad.BlocksInheritance.String(), node.ID, err)
			}
		}
	}

	if blocksGPOInheritance {
		applyingGPOs = appendUniqueGraphIDs(applyingGPOs, inheritedEnforcedGPOs...)
	} else {
		applyingGPOs = appendUniqueGraphIDs(applyingGPOs, inheritedGPOs...)
	}

	if childIDs, childContainers, err := fetchDirectGPOChildren(tx, node.ID); err != nil {
		return fmt.Errorf("failed fetching direct children of node %d: %w", node.ID, err)
	} else {
		for _, childID := range childIDs {
			for _, gpoID := range applyingGPOs {
				if !sink.Submit(ctx, post.EnsureRelationshipJob{FromID: gpoID, ToID: childID, Kind: ad.GPOAppliesTo}) {
					return fmt.Errorf("unable to submit GPOAppliesTo relationship")
				}
			}

			for _, gpoApplierID := range currentGPOAppliers {
				if !sink.Submit(ctx, post.EnsureRelationshipJob{FromID: gpoApplierID, ToID: childID, Kind: ad.CanApplyGPO}) {
					return fmt.Errorf("unable to submit CanApplyGPO relationship")
				}
			}
		}

		for _, childContainer := range childContainers {
			if err = processGPOs(ctx, tx, sink, childContainer, applyingGPOs, nextEnforcedGPOs, currentGPOAppliers, depth+1); err != nil {
				return err
			}
		}
	}

	return nil
}

func appendUniqueGraphIDs(target []graph.ID, values ...graph.ID) []graph.ID {
	seenIDs := make(map[graph.ID]struct{}, len(target)+len(values))
	result := make([]graph.ID, 0, len(target)+len(values))

	for _, existingID := range target {
		if _, seen := seenIDs[existingID]; !seen {
			seenIDs[existingID] = struct{}{}
			result = append(result, existingID)
		}
	}

	for _, value := range values {
		if _, seen := seenIDs[value]; !seen {
			seenIDs[value] = struct{}{}
			result = append(result, value)
		}
	}

	return result
}

func fetchAdditionalGPOAppliers(tx graph.Transaction, containerID graph.ID, ignoredAppliers []graph.ID) ([]graph.ID, error) {
	criteria := []graph.Criteria{
		query.KindIn(query.Start(), ad.Group, ad.User, ad.Computer),
		query.KindIn(query.Relationship(), ad.WriteGPLink, ad.GenericWrite, ad.GenericAll, ad.WriteDACL, ad.Owns, ad.WriteOwner),
		query.Equals(query.EndID(), containerID),
	}

	if len(ignoredAppliers) > 0 {
		criteria = append(criteria, query.Not(query.InIDs(query.StartID(), ignoredAppliers...)))
	}

	return ops.FetchStartNodeIDs(tx.Relationships().Filterf(func() graph.Criteria {
		return query.And(criteria...)
	}))
}

func fetchGPOLinkedDirectly(tx graph.Transaction, containerID graph.ID) ([]graph.ID, []graph.ID, error) {
	linkedGPOs, err := ops.FetchRelationships(tx.Relationships().Filterf(func() graph.Criteria {
		return query.And(
			query.Kind(query.Start(), ad.GPO),
			query.Kind(query.Relationship(), ad.GPLink),
			query.Equals(query.EndID(), containerID),
		)
	}))
	if err != nil {
		return nil, nil, err
	}

	directlyLinkedGPOs := make([]graph.ID, 0, len(linkedGPOs))
	directlyLinkedEnforcedGPOs := make([]graph.ID, 0, len(linkedGPOs))

	for _, linkedGPO := range linkedGPOs {
		enforced, err := linkedGPO.Properties.GetOrDefault(ad.Enforced.String(), false).Bool()
		if err != nil {
			return nil, nil, fmt.Errorf("failed reading %s on GPLink relationship %d to container %d: %w", ad.Enforced.String(), linkedGPO.ID, containerID, err)
		}

		directlyLinkedGPOs = append(directlyLinkedGPOs, linkedGPO.StartID)
		if enforced {
			directlyLinkedEnforcedGPOs = append(directlyLinkedEnforcedGPOs, linkedGPO.StartID)
		}
	}

	return directlyLinkedGPOs, directlyLinkedEnforcedGPOs, nil
}

func fetchDirectGPOChildren(tx graph.Transaction, containerID graph.ID) ([]graph.ID, graph.NodeSet, error) {
	childNodes, err := ops.FetchEndNodes(tx.Relationships().Filterf(func() graph.Criteria {
		return query.And(
			query.Equals(query.StartID(), containerID),
			query.KindIn(query.Relationship(), ad.Contains),
			query.KindIn(query.End(), ad.User, ad.Computer, ad.OU, ad.Container),
		)
	}))
	if err != nil {
		return nil, nil, err
	}

	childUserAndComputerIDs := make([]graph.ID, 0, childNodes.Len())
	childContainers := graph.NewNodeSet()

	for _, childNode := range childNodes {
		if childNode.Kinds.ContainsOneOf(ad.User, ad.Computer) {
			childUserAndComputerIDs = append(childUserAndComputerIDs, childNode.ID)
		} else if childNode.Kinds.ContainsOneOf(ad.OU, ad.Container) {
			childContainers.Add(childNode)
		}
	}

	return childUserAndComputerIDs, childContainers, nil
}

func GetGPOAppliesToComposition(ctx context.Context, db graph.Database, edge *graph.Relationship) (graph.PathSet, error) {
	var paths graph.PathSet

	if err := db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		gpoNode, err := ops.FetchNode(tx, edge.StartID)
		if err != nil {
			return err
		}

		paths, err = GetGPOAffectedObjectsPath(tx, gpoNode, edge.EndID)
		return err
	}); err != nil {
		return nil, err
	}

	return paths, nil
}

func GetCanApplyGPOComposition(ctx context.Context, db graph.Database, edge *graph.Relationship) (graph.PathSet, error) {
	var (
		startNode     *graph.Node
		traversalInst = traversal.New(db, post.MaximumDatabaseParallelWorkers)
		paths         = graph.PathSet{}
	)

	if err := db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		var err error

		if startNode, err = ops.FetchNode(tx, edge.StartID); err != nil {
			return err
		}

		return nil
	}); err != nil {
		return nil, err
	}

	if err := traversalInst.BreadthFirst(ctx,
		traversal.Plan{
			Root: startNode,
			Driver: traversal.NewPattern().
				OutboundWithDepth(1, 1, query.And(
					query.KindIn(query.Relationship(), ad.WriteGPLink, ad.GenericWrite, ad.GenericAll, ad.WriteDACL, ad.Owns, ad.WriteOwner),
					query.KindIn(query.End(), ad.OU, ad.Domain),
				)).
				OutboundWithDepth(0, 0, query.KindIn(query.Relationship(), ad.Contains)).
				Outbound(query.And(
					query.KindIn(query.Relationship(), ad.Contains),
					query.Equals(query.EndID(), edge.EndID),
				)).Do(func(terminal *graph.PathSegment) error {
				paths.AddPath(terminal.Path())
				return nil
			}),
		}); err != nil {
		return nil, err
	}

	return paths, nil
}
