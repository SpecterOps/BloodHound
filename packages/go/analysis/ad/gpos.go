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

package ad

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/specterops/bloodhound/analysis"
	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/dawgs/ops"
	"github.com/specterops/bloodhound/dawgs/query"
	"github.com/specterops/bloodhound/dawgs/traversal"
	"github.com/specterops/bloodhound/dawgs/util/channels"
	"github.com/specterops/bloodhound/graphschema/ad"
)

func PostGPOs(ctx context.Context, db graph.Database) (*analysis.AtomicPostProcessingStats, error) {
	if domainNodes, err := fetchCollectedDomainNodes(ctx, db); err != nil {
		return &analysis.AtomicPostProcessingStats{}, err
	} else {
		operation := analysis.NewPostRelationshipOperation(ctx, db, "GPOs Post Processing")

		for _, domain := range domainNodes {
			if err := operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
				processGPOs(ctx, tx, outC, domain, []graph.ID{}, []graph.ID{}, []graph.ID{})
				return nil
			}); err != nil {
				slog.ErrorContext(ctx, fmt.Sprintf("Failded processing GPOs of domain %d: %v", domain.ID, err))
			}
		}

		return &operation.Stats, operation.Done()
	}
}

func processGPOs(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob, node *graph.Node, gposInherited, gposInheritedEnforced, gpoAppliers []graph.ID) {
	applyingGPOs := []graph.ID{}
	gposInheritedEnforcedNew := gposInheritedEnforced
	blocksGPOInheritance := false

	if node.Kinds.ContainsOneOf(ad.Domain, ad.OU) {
		// Find principals with permission to link GPOs to this node
		if gpoAppliersOnThisContainer, err := fetchGPOAppliers(tx, node.ID); err != nil {
			slog.ErrorContext(ctx, fmt.Sprintf("Failded fetching GPO appliers on node %d: %v", node.ID, err))
		} else {
			gpoAppliers = append(gpoAppliers, gpoAppliersOnThisContainer...)
		}

		// Find GPOs linked to this node
		if gposLinkedDirectly, err := fetchGPOLinkedDirectly(tx, node.ID, false); err != nil {
			slog.ErrorContext(ctx, fmt.Sprintf("Failded fetching GPO linked directly on node %d: %v", node.ID, err))
		} else if gposLinkedDirectlyEnforced, err := fetchGPOLinkedDirectly(tx, node.ID, true); err != nil {
			slog.ErrorContext(ctx, fmt.Sprintf("Failded fetching GPO linked directly on node %d: %v", node.ID, err))
		} else {
			applyingGPOs = gposLinkedDirectly
			gposInheritedEnforcedNew = append(gposInheritedEnforced, gposLinkedDirectlyEnforced...)

			if node.Kinds.ContainsOneOf(ad.OU) {
				if blocksInheritance, err := node.Properties.Get(ad.BlocksInheritance.String()).Bool(); err != nil {
					slog.ErrorContext(ctx, fmt.Sprintf("Failded fetching %s on node %d: %v", ad.BlocksInheritance.String(), node.ID, err))
				} else {
					blocksGPOInheritance = blocksInheritance
				}
			}
		}
	}

	if blocksGPOInheritance {
		applyingGPOs = append(applyingGPOs, gposInheritedEnforced...)
	} else {
		applyingGPOs = append(applyingGPOs, gposInherited...)
	}

	// Create GPO edges to direct child users and computers
	if children, err := fetchDirectChildUsersAndComputers(tx, node.ID); err != nil {
		slog.ErrorContext(ctx, fmt.Sprintf("Failded fetching direct child user and computer nodes of node %d: %v", node.ID, err))
	} else {
		for _, childId := range children {
			for _, gpoId := range applyingGPOs {
				channels.Submit(ctx, outC, analysis.CreatePostRelationshipJob{
					FromID: gpoId,
					ToID:   childId,
					Kind:   ad.GPOAppliesTo,
				})
			}

			for _, gpoApplierId := range gpoAppliers {
				channels.Submit(ctx, outC, analysis.CreatePostRelationshipJob{
					FromID: gpoApplierId,
					ToID:   childId,
					Kind:   ad.CanApplyGPO,
				})
			}
		}
	}

	// Continue recursively with child container nodes
	if childContainers, err := fetchDirectChildContainers(tx, node.ID); err != nil {
		slog.ErrorContext(ctx, fmt.Sprintf("Failded fetching direct child container nodes of node %d: %v", node.ID, err))
	} else {
		for _, childContainer := range childContainers {
			processGPOs(ctx, tx, outC, childContainer, applyingGPOs, gposInheritedEnforcedNew, gpoAppliers)
		}
	}
}

func fetchGPOAppliers(tx graph.Transaction, containerID graph.ID) ([]graph.ID, error) {
	if gpoAppliers, err := ops.FetchStartNodeIDs(tx.Relationships().Filterf(func() graph.Criteria {
		return query.And(
			query.KindIn(query.Start(), ad.Group, ad.User, ad.Computer),
			query.KindIn(query.Relationship(), ad.WriteGPLink, ad.GenericWrite, ad.GenericAll, ad.WriteDACL, ad.Owns, ad.WriteOwner),
			query.Equals(query.EndID(), containerID),
		)
	})); err != nil {
		return nil, err
	} else {
		return gpoAppliers, nil
	}
}

func fetchGPOLinkedDirectly(tx graph.Transaction, containerID graph.ID, enforcedOnly bool) ([]graph.ID, error) {
	if gposLinked, err := ops.FetchStartNodeIDs(tx.Relationships().Filterf(func() graph.Criteria {
		linkedGPOsQuery := query.And(
			query.Kind(query.Start(), ad.GPO),
			query.Kind(query.Relationship(), ad.GPLink),
			query.Equals(query.EndID(), containerID),
		)
		if enforcedOnly {
			return query.And(linkedGPOsQuery,
				query.Equals(query.RelationshipProperty(ad.Enforced.String()), true),
			)
		} else {
			return linkedGPOsQuery
		}

	})); err != nil {
		return nil, err
	} else {
		return gposLinked, nil
	}
}

func fetchDirectChildUsersAndComputers(tx graph.Transaction, containerID graph.ID) ([]graph.ID, error) {
	if childUsersAndComputers, err := ops.FetchEndNodes(tx.Relationships().Filterf(func() graph.Criteria {
		return query.And(
			query.Equals(query.StartID(), containerID),
			query.KindIn(query.Relationship(), ad.Contains),
			query.KindIn(query.End(), ad.User, ad.Computer),
		)
	})); err != nil {
		return nil, err
	} else {
		return childUsersAndComputers.IDs(), nil
	}
}

func fetchDirectChildContainers(tx graph.Transaction, containerID graph.ID) (graph.NodeSet, error) {
	if childContainers, err := ops.FetchEndNodes(tx.Relationships().Filterf(func() graph.Criteria {
		return query.And(
			query.Equals(query.StartID(), containerID),
			query.KindIn(query.Relationship(), ad.Contains),
			query.KindIn(query.End(), ad.OU, ad.Container),
		)
	})); err != nil {
		return nil, err
	} else {
		return childContainers, nil
	}
}

func GetGPOAppliesToComposition(ctx context.Context, db graph.Database, edge *graph.Relationship) (graph.PathSet, error) {
	var paths graph.PathSet

	err := db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		gpoNode, err := ops.FetchNode(tx, edge.StartID)
		if err != nil {
			return err
		}

		paths, err = GetGPOAffectedObjectsPath(tx, gpoNode, edge.EndID)
		return err
	})
	if err != nil {
		return nil, err
	}

	return paths, nil
}

func GetCanApplyGPOComposition(ctx context.Context, db graph.Database, edge *graph.Relationship) (graph.PathSet, error) {
	// Composition query
	// MATCH (x)-[:CanApplyGPO]->(y)
	// MATCH p = (x)-[:WriteGPLink|GenericWrite|GenericAll|WriteDacl|Owns|WriteOwner]->(z:Base)-[:Contains*..]->(y)
	// WHERE z:OU OR z:Domain
	// RETURN p LIMIT 100

	var (
		startNode *graph.Node

		traversalInst = traversal.New(db, analysis.MaximumDatabaseParallelWorkers)
		paths         = graph.PathSet{}
	)

	if err := db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		var err error
		if startNode, err = ops.FetchNode(tx, edge.StartID); err != nil {
			return err
		} else {
			return nil
		}
	}); err != nil {
		return nil, err
	}

	if err := traversalInst.BreadthFirst(ctx,
		traversal.Plan{
			Root: startNode,
			Driver: traversal.NewPattern().
				OutboundWithDepth(1, 1,
					query.And(
						query.Kind(query.Relationship(), ad.WriteGPLink, ad.GenericWrite, ad.GenericAll, ad.WriteDACL, ad.Owns, ad.WriteOwner),
						query.KindIn(query.End(), ad.OU, ad.Domain),
					)).
				Outbound(query.And(
					query.KindIn(query.Relationship(), ad.Contains),
					query.Equals(query.EndID(), edge.EndID),
				)).Do(
				func(terminal *graph.PathSegment) error {
					paths.AddPath(terminal.Path())
					return nil
				}),
		}); err != nil {
		return nil, err
	}

	return paths, nil
}
