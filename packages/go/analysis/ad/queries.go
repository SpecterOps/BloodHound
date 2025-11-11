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

package ad

import (
	"context"
	"fmt"
	"log/slog"
	"slices"
	"time"

	"github.com/RoaringBitmap/roaring/v2/roaring64"
	"github.com/specterops/bloodhound/packages/go/analysis"
	"github.com/specterops/bloodhound/packages/go/analysis/ad/wellknown"
	"github.com/specterops/bloodhound/packages/go/analysis/tiering"
	"github.com/specterops/bloodhound/packages/go/bhlog/measure"
	"github.com/specterops/bloodhound/packages/go/graphschema/ad"
	"github.com/specterops/bloodhound/packages/go/graphschema/common"
	"github.com/specterops/dawgs/cardinality"
	"github.com/specterops/dawgs/graph"
	"github.com/specterops/dawgs/graphcache"
	"github.com/specterops/dawgs/ops"
	"github.com/specterops/dawgs/query"
	"github.com/specterops/dawgs/traversal"
)

func FetchGraphDBTierZeroTaggedAssets(ctx context.Context, db graph.Database, domainSID string) (graph.NodeSet, error) {
	defer measure.ContextMeasure(ctx, slog.LevelInfo, "FetchGraphDBTierZeroTaggedAssets")()

	var (
		nodes graph.NodeSet
		err   error
	)

	return nodes, db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		nodes, err = ops.FetchNodeSet(tx.Nodes().Filterf(func() graph.Criteria {
			return query.And(
				query.Kind(query.Node(), ad.Entity),
				query.Equals(query.NodeProperty(ad.DomainSID.String()), domainSID),
				query.StringContains(query.NodeProperty(common.SystemTags.String()), ad.AdminTierZero),
			)
		}))

		return err
	})

}

func FetchAllEnforcedGPOs(ctx context.Context, db graph.Database, targets graph.NodeSet) (graph.NodeSet, error) {
	defer measure.ContextMeasure(ctx, slog.LevelInfo, "FetchAllEnforcedGPOs")()

	enforcedGPOs := graph.NewNodeSet()

	return enforcedGPOs, db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		for _, attackPathRoot := range targets {
			if enforced, err := FetchEnforcedGPOs(tx, attackPathRoot, 0, 0); err != nil {
				return err
			} else {
				enforcedGPOs.AddSet(enforced)
			}
		}

		return nil
	})
}

func FetchOUContainers(ctx context.Context, db graph.Database, targets graph.NodeSet) (graph.NodeSet, error) {
	defer measure.ContextMeasure(ctx, slog.LevelInfo, "FetchOUContainers")()

	oUs := graph.NewNodeSet()

	return oUs, db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		for _, attackPathRoot := range targets {
			if ou, err := FetchOUContainersOfNode(tx, attackPathRoot); err != nil {
				return err
			} else if ou != nil {
				oUs.AddSet(ou)
			}
		}

		return nil
	})
}

func FetchAllDomains(ctx context.Context, db graph.Database) ([]*graph.Node, error) {
	var (
		nodes []*graph.Node
		err   error
	)

	return nodes, db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		nodes, err = ops.FetchNodes(tx.Nodes().Filterf(func() graph.Criteria {
			return query.Kind(query.Node(), ad.Domain)
		}).OrderBy(
			query.Order(query.NodeProperty(common.Name.String()), query.Descending()),
		))

		return err
	})
}

func FetchActiveDirectoryTierZeroRoots(ctx context.Context, db graph.Database, domain *graph.Node, autoTagT0ParentObjectsFlag bool) (graph.NodeSet, error) {
	defer measure.ContextLogAndMeasure(ctx, slog.LevelInfo, "FetchActiveDirectoryTierZeroRoots")()

	if domainSID, err := domain.Properties.Get(common.ObjectID.String()).String(); err != nil {
		return nil, err
	} else {
		attackPathRoots := graph.NewNodeSet()

		// Add the domain as one of the critical path roots
		attackPathRoots.Add(domain)

		// Pull in custom tier zero tagged assets
		if customTierZeroNodes, err := FetchGraphDBTierZeroTaggedAssets(ctx, db, domainSID); err != nil {
			return nil, err
		} else {
			attackPathRoots.AddSet(customTierZeroNodes)
		}

		// Pull in well known tier zero nodes by SID suffix
		if wellKnownTierZeroNodes, err := FetchWellKnownTierZeroEntities(ctx, db, domainSID); err != nil {
			return nil, err
		} else {
			attackPathRoots.AddSet(wellKnownTierZeroNodes)
		}

		// Pull in all group members of attack path roots
		if allGroupMembers, err := FetchAllGroupMembers(ctx, db, attackPathRoots); err != nil {
			return nil, err
		} else {
			attackPathRoots.AddSet(allGroupMembers)
		}

		// Add all enforced GPO nodes to the attack path roots
		if enforcedGPOs, err := FetchAllEnforcedGPOs(ctx, db, attackPathRoots); err != nil {
			return nil, err
		} else {
			attackPathRoots.AddSet(enforcedGPOs)
		}

		if autoTagT0ParentObjectsFlag {
			// Add the OUs to the attack path roots
			if ous, err := FetchOUContainers(ctx, db, attackPathRoots); err != nil {
				return nil, err
			} else {
				attackPathRoots.AddSet(ous)
			}

			// Add the containers to the attack path roots
			db.ReadTransaction(ctx, func(tx graph.Transaction) error {
				for _, attackPathRoot := range attackPathRoots {

					// Do not add container if ACL inheritance is disabled
					isACLProtected, err := attackPathRoot.Properties.Get(ad.IsACLProtected.String()).Bool()
					if err != nil || !isACLProtected {
						if containers, err := FetchContainersOfNode(tx, attackPathRoot); err != nil {
							return err
						} else if containers != nil {
							attackPathRoots.AddSet(containers)
						}
					}
				}
				return nil
			})
		}

		// Find all next-tier assets
		return attackPathRoots, nil
	}
}

func FetchCollectedDomains(tx graph.Transaction) (graph.NodeSet, error) {
	return ops.FetchNodeSet(tx.Nodes().Filterf(func() graph.Criteria {
		return query.And(
			query.Kind(query.Node(), ad.Domain),
			query.Equals(query.NodeProperty(common.Collected.String()), true),
		)
	}))
}

func GetCollectedDomains(ctx context.Context, db graph.Database) (graph.NodeSet, error) {
	var domains graph.NodeSet

	return domains, db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		if innerDomains, err := FetchCollectedDomains(tx); err != nil {
			return err
		} else {
			domains = innerDomains
			return nil
		}
	})
}

func getGPOLinks(tx graph.Transaction, node *graph.Node) ([]*graph.Relationship, error) {
	if gpLinks, err := ops.FetchRelationships(tx.Relationships().Filterf(func() graph.Criteria {
		return query.And(
			query.Equals(query.StartID(), node.ID),
			query.Kind(query.Relationship(), ad.GPLink),
			query.KindIn(query.End(), ad.Domain, ad.OU),
		)
	})); err != nil {
		return nil, err
	} else {
		return gpLinks, nil
	}
}

func CreateGPOAffectedIntermediariesListDelegate(candidateFilter ops.NodeFilter) analysis.ListDelegate {
	return func(tx graph.Transaction, node *graph.Node, skip, limit int) (graph.NodeSet, error) {
		nodeSet := graph.NewNodeSet()

		if gpLinks, err := getGPOLinks(tx, node); err != nil {
			return nil, err
		} else {
			for _, rel := range gpLinks {
				enforced, err := rel.Properties.Get(ad.Enforced.String()).Bool()
				if err != nil {
					// Its possible the property isn't here, so lets set enforced to false and let it roll
					enforced = false
				}

				if _, end, err := ops.FetchRelationshipNodes(tx, rel); err != nil {
					return nil, err
				} else {
					var descentFilter ops.SegmentFilter

					// Set our descent filter based on enforcement status
					if !enforced {
						descentFilter = BlocksInheritanceDescentFilter
					}

					if nodes, err := ops.AcyclicTraverseNodes(tx, ops.TraversalPlan{
						Root:          end,
						Direction:     graph.DirectionOutbound,
						BranchQuery:   FilterContainsRelationship,
						DescentFilter: descentFilter,
						Skip:          skip,
						Limit:         limit,
					}, candidateFilter); err != nil {
						return nil, err
					} else {
						nodeSet.AddSet(nodes)
					}
				}
			}

			return nodeSet, nil
		}
	}
}

func FetchGPOAffectedTierZeroPathDelegate(tx graph.Transaction, node *graph.Node) (graph.PathSet, error) {
	pathSet := graph.NewPathSet()

	if gpLinks, err := getGPOLinks(tx, node); err != nil {
		return nil, err
	} else {
		for _, rel := range gpLinks {
			enforced, err := rel.Properties.Get(ad.Enforced.String()).Bool()
			if err != nil {
				// Its possible the property isn't here, so lets set enforced to false and let it roll
				enforced = false
			}

			if _, end, err := ops.FetchRelationshipNodes(tx, rel); err != nil {
				return nil, err
			} else {
				var descentFilter ops.SegmentFilter

				// Set our descent filter based on enforcement status
				if !enforced {
					descentFilter = BlocksInheritanceDescentFilter
				}

				if paths, err := ops.TraverseIntermediaryPaths(tx, ops.TraversalPlan{
					Root:          end,
					Direction:     graph.DirectionOutbound,
					BranchQuery:   FilterContainsRelationship,
					DescentFilter: descentFilter,
					PathFilter: func(ctx *ops.TraversalContext, segment *graph.PathSegment) bool {
						return tiering.IsTierZero(node)
					},
				}, SelectGPOTierZeroCandidateFilter); err != nil {
					return nil, err
				} else {
					if paths.Len() > 0 {
						pathSet.AddPathSet(paths)
						pathSet.AddPath(graph.Path{
							Nodes: []*graph.Node{
								node,
								end,
							},

							Edges: []*graph.Relationship{
								rel,
							},
						})
					}
				}
			}
		}

		return pathSet, nil
	}
}

func FetchGPOAffectedContainerPaths(tx graph.Transaction, node *graph.Node) (graph.PathSet, error) {
	pathSet := graph.NewPathSet()

	if gpLinks, err := getGPOLinks(tx, node); err != nil {
		return nil, err
	} else {
		for _, rel := range gpLinks {
			enforced, err := rel.Properties.Get(ad.Enforced.String()).Bool()
			if err != nil {
				// Its possible the property isn't here, so lets set enforced to false and let it roll
				enforced = false
			}

			if _, end, err := ops.FetchRelationshipNodes(tx, rel); err != nil {
				return nil, err
			} else {
				var descentFilter ops.SegmentFilter

				// Set our descent filter based on enforcement status
				if !enforced {
					descentFilter = BlocksInheritanceDescentFilter
				}

				pathSet.AddPath(graph.Path{
					Nodes: []*graph.Node{node, end},
					Edges: []*graph.Relationship{rel},
				})

				if paths, err := ops.TraversePaths(tx, ops.TraversalPlan{
					Root:        end,
					Direction:   graph.DirectionOutbound,
					BranchQuery: FilterContainsRelationship,
					DescentFilter: func(ctx *ops.TraversalContext, segment *graph.PathSegment) bool {
						if !segment.Node.Kinds.ContainsOneOf(ad.OU, ad.Container, ad.Domain) {
							return false
						} else if descentFilter != nil {
							return descentFilter(ctx, segment)
						} else {
							return true
						}
					},
				}); err != nil {
					return nil, err
				} else {
					if paths.Len() > 0 {
						pathSet.AddPathSet(paths)
					}
				}
			}
		}

		return pathSet, nil
	}
}

func CreateGPOAffectedIntermediariesPathDelegate(targetKinds ...graph.Kind) analysis.PathDelegate {
	return func(tx graph.Transaction, node *graph.Node) (graph.PathSet, error) {
		pathSet := graph.NewPathSet()

		if gpLinks, err := getGPOLinks(tx, node); err != nil {
			return nil, err
		} else {
			for _, rel := range gpLinks {
				// It's possible the property isn't here, so lets set enforced to false and let it roll
				enforced, _ := rel.Properties.GetOrDefault(ad.Enforced.String(), false).Bool()

				if end, err := ops.FetchNode(tx, rel.EndID); err != nil {
					return nil, err
				} else if paths, err := ops.TraversePaths(tx, ops.TraversalPlan{
					Root:        end,
					Direction:   graph.DirectionOutbound,
					BranchQuery: FilterContainsRelationship,
					DescentFilter: func(ctx *ops.TraversalContext, segment *graph.PathSegment) bool {
						if !enforced {
							return BlocksInheritanceDescentFilter(ctx, segment)
						}

						return true
					},
					PathFilter: func(ctx *ops.TraversalContext, segment *graph.PathSegment) bool {
						return len(targetKinds) == 0 || segment.Node.Kinds.ContainsOneOf(targetKinds...)
					},
				}); err != nil {
					return nil, err
				} else if paths.Len() > 0 {
					pathSet.AddPathSet(paths)
					pathSet.AddPath(graph.Path{
						Nodes: []*graph.Node{node, end},
						Edges: []*graph.Relationship{rel},
					})
				}
			}

			return pathSet, nil
		}
	}
}

func FetchEnforcedGPOs(tx graph.Transaction, target *graph.Node, skip, limit int) (graph.NodeSet, error) {
	enforcedGPOs := graph.NewNodeSet()

	if err := ops.Traversal(tx, ops.TraversalPlan{
		Root:      target,
		Direction: graph.DirectionInbound,
		BranchQuery: func() graph.Criteria {
			return query.And(
				query.KindIn(query.Start(), ad.Domain, ad.OU, ad.GPO),
				query.KindIn(query.Relationship(), ad.Contains, ad.GPLink),
			)
		},
		Skip:  skip,
		Limit: limit,
	}, func(ctx *ops.TraversalContext, segment *graph.PathSegment) error {
		// Does this path terminate at a GPO that we haven't seen as a previously enforced GPO?
		if segment.Node.Kinds.ContainsOneOf(ad.GPO) && !enforcedGPOs.Contains(segment.Node) {
			gpLinkRelationship := segment.Edge

			// Check if the GPLink relationship is enforced
			if gpLinkEnforced, _ := gpLinkRelationship.Properties.GetOrDefault(ad.Enforced.String(), false).Bool(); gpLinkEnforced {
				if ctx.LimitSkipTracker.ShouldCollect() {
					// Add this GPO right away as enforced and exit
					enforcedGPOs.Add(segment.Node)
				}
			} else {
				// Assume that the GPO is enforced at the start
				isGPOEnforced := true
				lastNodeBlocks := false

				// Walk the GPO path to see if any of the nodes between the GPO and the enforcement target block GPO
				// inheritance. This walk starts at the GPO and moves down, with end being the GPO to start
				segment.Path().WalkReverse(func(start, end *graph.Node, relationship *graph.Relationship) bool {
					if !start.Kinds.ContainsOneOf(ad.OU, ad.Domain) {
						// If we run into anything that isn't an OU or a Domain node then we're done checking for
						// inheritance blocking
						return false
					} else if lastNodeBlocks && start.Kinds.ContainsOneOf(ad.OU) {
						// If the previous node blocks inheritance, and we've hit an OU, then the GPO is not enforced on this path, and we don't need to check any further
						isGPOEnforced = false
						return false
					}

					// Check to see if this node in the Domain and OU contains path blocks GPO inheritance
					if blocksInheritance, _ := start.Properties.GetOrDefault(ad.BlocksInheritance.String(), false).Bool(); blocksInheritance {
						// If this Domain or OU node blocks inheritance then we're done walking this GPO enforcement
						// path
						lastNodeBlocks = true
						return true
					}

					// Continue walking the path otherwise
					return true
				})

				// If the GPO is still marked as enforced, meaning that the Domain node nor any of the OU nodes blocked
				// inheritance of it
				if isGPOEnforced {
					if ctx.LimitSkipTracker.ShouldCollect() {
						enforcedGPOs.Add(segment.Node)
					}
				}
			}
		}

		return nil
	}); err != nil {
		return nil, err
	} else {
		return enforcedGPOs, nil
	}
}

func FetchEnforcedGPOsPaths(ctx context.Context, db graph.Database, target *graph.Node) (graph.PathSet, error) {
	var (
		pathSet      = graph.NewPathSet()
		enforcedGPOs = graph.NewNodeSet()
	)

	return pathSet, db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		return ops.Traversal(tx, ops.TraversalPlan{
			Root:      target,
			Direction: graph.DirectionInbound,
			BranchQuery: func() graph.Criteria {
				return query.And(
					query.KindIn(query.Start(), ad.Domain, ad.OU, ad.GPO),
					query.KindIn(query.Relationship(), ad.Contains, ad.GPLink),
				)
			},
		}, func(ctx *ops.TraversalContext, segment *graph.PathSegment) error {
			// Does this path terminate at a GPO that we haven't seen as a previously enforced GPO?
			if segment.Node.Kinds.ContainsOneOf(ad.GPO) && !enforcedGPOs.Contains(segment.Node) {
				gpLinkRelationship := segment.Edge

				// Check if the GPLink relationship is enforced
				if gpLinkEnforced, _ := gpLinkRelationship.Properties.GetOrDefault(ad.Enforced.String(), false).Bool(); gpLinkEnforced {
					if ctx.LimitSkipTracker.ShouldCollect() {
						// Add this GPO right away as enforced and exit
						enforcedGPOs.Add(segment.Node)
						pathSet.AddPath(segment.Path())
					}
				} else {
					// Assume that the GPO is enforced at the start
					isGPOEnforced := true
					lastNodeBlocks := false

					// Walk the GPO path to see if any of the nodes between the GPO and the enforcement target block GPO
					// inheritance. This walk starts at the GPO and moves down, with end being the GPO to start
					segment.Path().WalkReverse(func(start, end *graph.Node, relationship *graph.Relationship) bool {
						if !start.Kinds.ContainsOneOf(ad.OU, ad.Domain) {
							// If we run into anything that isn't an OU or a Domain node then we're done checking for
							// inheritance blocking
							return false
						} else if lastNodeBlocks && start.Kinds.ContainsOneOf(ad.OU) {
							// If the previous node blocks inheritance, and we've hit an OU, then the GPO is not enforced on this path, and we don't need to check any further
							isGPOEnforced = false
							return false
						}

						// Check to see if this node in the Domain and OU contains path blocks GPO inheritance
						if blocksInheritance, _ := start.Properties.GetOrDefault(ad.BlocksInheritance.String(), false).Bool(); blocksInheritance {
							// If this Domain or OU node blocks inheritance then we're done walking this GPO enforcement
							// path
							lastNodeBlocks = true
							return true
						}

						// Continue walking the path otherwise
						return true
					})

					// If the GPO is still marked as enforced, meaning that the Domain node nor any of the OU nodes blocked
					// inheritance of it
					if isGPOEnforced {
						// Add this GPO as enforced
						enforcedGPOs.Add(segment.Node)
						pathSet.AddPath(segment.Path())
					}
				}
			}

			return nil
		})
	})
}

func FetchACLInheritancePath(ctx context.Context, db graph.Database, edge *graph.Relationship) (graph.PathSet, error) {
	pathSet := graph.NewPathSet()

	return pathSet, db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		var (
			hash, _        = edge.Properties.GetOrDefault(ad.InheritanceHash.String(), "").String()
			isAcl, _       = edge.Properties.GetOrDefault(ad.IsACL.String(), false).Bool()
			isInherited, _ = edge.Properties.GetOrDefault(common.IsInherited.String(), false).Bool()
		)

		// If the target edge is not ACL-related or does not have an inheritance hash to match against, return an empty result set
		if !isAcl || !isInherited || len(hash) == 0 {
			return nil
		} else if startNode, endNode, err := ops.FetchRelationshipNodes(tx, edge); err != nil {
			return err
		} else {
			err = ops.Traversal(tx, ops.TraversalPlan{
				Root:      endNode,
				Direction: graph.DirectionInbound,
				BranchQuery: func() graph.Criteria {
					return query.And(
						query.KindIn(query.Start(), ad.Domain, ad.OU, ad.Container),
						query.KindIn(query.Relationship(), ad.Contains),
					)
				},
				ExpansionFilter: func(segment *graph.PathSegment) bool {
					// Check that our hash is included in the current node
					hashes, _ := segment.Node.Properties.GetOrDefault(ad.InheritanceHashes.String(), []string{}).StringSlice()

					if slices.Contains(hashes, hash) {
						isInheritable := true
						// Walk back up the inheritance chain until we reach our start node, checking that inheritance is not blocked
						segment.Path().WalkReverse(func(start, end *graph.Node, relationship *graph.Relationship) bool {
							// If we run into an intermediary node that is protected, we can stop walking this path. Checking just
							// the start node of each segment purposefully excludes the inheritance source from this check.
							if isACLProtected, _ := start.Properties.GetOrDefault(ad.IsACLProtected.String(), false).Bool(); isACLProtected {
								isInheritable = false
								return false
							}
							return true
						})

						if isInheritable {
							pathSet.AddPath(segment.Path())
						}
					}
					return true
				},
			}, nil)

			// If an inheritance path was found, append the starting path to our result
			if pathSet.AllNodes().Len() > 0 {
				pathSet.AddPath(graph.Path{
					Nodes: []*graph.Node{startNode, endNode},
					Edges: []*graph.Relationship{edge},
				})
			}

			return err
		}
	})
}

func FetchOUContainersOfNode(tx graph.Transaction, target *graph.Node) (graph.NodeSet, error) {
	oUContainers := graph.NewNodeSet()
	if paths, err := ops.TraversePaths(tx, ops.TraversalPlan{
		Root:      target,
		Direction: graph.DirectionInbound,
		BranchQuery: func() graph.Criteria {
			return query.And(
				query.Kind(query.Start(), ad.OU),
				query.Kind(query.Relationship(), ad.Contains),
			)
		},
	}); err != nil {
		return nil, err
	} else {
		oUContainers.AddSet(paths.AllNodes())
	}
	return oUContainers, nil
}

func FetchContainersOfNode(tx graph.Transaction, target *graph.Node) (graph.NodeSet, error) {
	containers := graph.NewNodeSet()
	if paths, err := ops.TraversePaths(tx, ops.TraversalPlan{
		Root:      target,
		Direction: graph.DirectionInbound,
		BranchQuery: func() graph.Criteria {
			return query.And(
				query.Kind(query.Start(), ad.Container),
				query.Kind(query.Relationship(), ad.Contains),
			)
		},
	}); err != nil {
		return nil, err
	} else {
		containers.AddSet(paths.AllNodes())
	}
	return containers, nil
}

func CreateOUContainedListDelegate(kind graph.Kind) analysis.ListDelegate {
	return func(tx graph.Transaction, node *graph.Node, skip, limit int) (graph.NodeSet, error) {
		return ops.AcyclicTraverseTerminals(tx, ops.TraversalPlan{
			Root:      node,
			Direction: graph.DirectionOutbound,
			BranchQuery: func() graph.Criteria {
				return query.And(
					query.Kind(query.Relationship(), ad.Contains),
				)
			},
			PathFilter: func(ctx *ops.TraversalContext, segment *graph.PathSegment) bool {
				return segment.Node.Kinds.ContainsOneOf(kind)
			},
			Skip:  skip,
			Limit: limit,
		})
	}
}

func CreateOUContainedPathDelegate(kind graph.Kind) analysis.PathDelegate {
	return func(tx graph.Transaction, node *graph.Node) (graph.PathSet, error) {
		return ops.TraversePaths(tx, ops.TraversalPlan{
			Root:      node,
			Direction: graph.DirectionOutbound,
			BranchQuery: func() graph.Criteria {
				return query.And(
					query.Kind(query.Relationship(), ad.Contains),
				)
			},
			PathFilter: func(ctx *ops.TraversalContext, segment *graph.PathSegment) bool {
				return segment.Node.Kinds.ContainsOneOf(kind)
			},
		})
	}
}

func CreateDomainTrustListDelegate(direction graph.Direction) analysis.ListDelegate {
	return func(tx graph.Transaction, node *graph.Node, skip, limit int) (graph.NodeSet, error) {
		return ops.AcyclicTraverseNodes(tx, ops.TraversalPlan{
			Root:      node,
			Direction: direction,
			BranchQuery: func() graph.Criteria {
				return query.KindIn(query.Relationship(), ad.SameForestTrust, ad.CrossForestTrust)
			},
			Skip:  skip,
			Limit: limit,
		}, func(candidate *graph.Node) bool {
			return candidate.Kinds.ContainsOneOf(ad.Domain) && candidate.ID != node.ID
		})
	}
}

func CreateDomainTrustPathDelegate(direction graph.Direction) analysis.PathDelegate {
	return func(tx graph.Transaction, node *graph.Node) (graph.PathSet, error) {
		return ops.TraversePaths(tx, ops.TraversalPlan{
			Root:      node,
			Direction: direction,
			BranchQuery: func() graph.Criteria {
				return query.KindIn(query.Relationship(), ad.SameForestTrust, ad.CrossForestTrust)
			},
		})
	}
}

func CreateDomainContainedEntityListDelegate(kind graph.Kind) analysis.ListDelegate {
	return func(tx graph.Transaction, node *graph.Node, skip, limit int) (graph.NodeSet, error) {
		if domainSid, err := node.Properties.Get(ad.DomainSID.String()).String(); err != nil {
			return nil, err
		} else {
			return ops.FetchNodeSet(tx.Nodes().Filterf(func() graph.Criteria {
				return query.And(
					query.Kind(query.Node(), kind),
					query.Equals(query.NodeProperty(ad.DomainSID.String()), domainSid),
				)
			}))
		}
	}
}

func FetchDCSyncers(tx graph.Transaction, node *graph.Node, skip, limit int) (graph.NodeSet, error) {
	if nodes, err := ops.FetchStartNodes(tx.Relationships().Filterf(func() graph.Criteria {
		return query.And(
			query.Equals(query.EndID(), node.ID),
			query.Kind(query.Relationship(), ad.DCSync),
		)
	})); err != nil {
		return nil, err
	} else {
		return graph.SortAndSliceNodeSet(nodes, skip, limit), nil
	}

}

func FetchDCSyncerPaths(tx graph.Transaction, node *graph.Node) (graph.PathSet, error) {
	if dcSyncerNodes, err := ops.FetchStartNodes(tx.Relationships().Filterf(func() graph.Criteria {
		return query.And(
			query.Equals(query.EndID(), node.ID),
			query.Kind(query.Relationship(), ad.DCSync),
		)
	})); err != nil {
		return nil, err
	} else {
		pathSet := graph.NewPathSet()

		for _, dcsyncer := range dcSyncerNodes {
			if paths, err := ops.TraversePaths(tx, ops.TraversalPlan{
				Root:      dcsyncer,
				Direction: graph.DirectionOutbound,
				BranchQuery: func() graph.Criteria {
					return query.KindIn(query.Relationship(), ad.MemberOf, ad.GetChanges, ad.GetChangesAll)
				},
				PathFilter: func(ctx *ops.TraversalContext, segment *graph.PathSegment) bool {
					return segment.Node.Kinds.ContainsOneOf(ad.Domain)
				},
			}); err != nil {
				return nil, err
			} else {
				pathSet.AddPathSet(paths)
			}
		}

		return pathSet, nil
	}
}

func FetchForeignGPOControllers(tx graph.Transaction, node *graph.Node, skip, limit int) (graph.NodeSet, error) {
	if domainSID, err := getNodeDomainSIDOrObjectID(node); err != nil {
		return nil, err
	} else if gpoIDs, err := ops.FetchNodeIDs(tx.Nodes().Filterf(func() graph.Criteria {
		return query.And(
			query.Kind(query.Node(), ad.GPO),
			query.Equals(query.NodeProperty(ad.DomainSID.String()), domainSID),
		)
	})); err != nil {
		return nil, err
	} else {
		if directControllers, err := ops.FetchStartNodes(tx.Relationships().Filterf(func() graph.Criteria {
			return query.And(
				query.InIDs(query.EndID(), gpoIDs...),
				query.KindIn(query.Relationship(), ad.ACLRelationships()...),
				query.Not(query.Equals(query.StartProperty(ad.DomainSID.String()), domainSID)),
			)
		})); err != nil {
			return nil, err
		} else if groups, err := ops.FetchStartNodes(tx.Relationships().Filterf(func() graph.Criteria {
			return query.And(
				query.Kind(query.Start(), ad.Group),
				query.KindIn(query.Relationship(), ad.ACLRelationships()...),
				query.Equals(query.EndProperty(ad.DomainSID.String()), domainSID),
				query.Kind(query.End(), ad.GPO),
			)
		})); err != nil {
			return nil, err
		} else {
			nodeSet := graph.NewNodeSet()

			for _, group := range groups {
				if nodes, err := ops.AcyclicTraverseTerminals(tx, ops.TraversalPlan{
					Root:      group,
					Direction: graph.DirectionInbound,
					BranchQuery: func() graph.Criteria {
						return query.Or(
							query.Kind(query.Start(), ad.Group),
							query.Not(query.Equals(query.StartProperty(ad.DomainSID.String()), domainSID)),
						)
					},
				}); err != nil {
					return nil, err
				} else {
					nodeSet.AddSet(nodes)
				}
			}

			nodeSet.AddSet(directControllers)
			return graph.SortAndSliceNodeSet(nodeSet, skip, limit), nil
		}
	}
}

func FetchForeignGPOControllerPaths(tx graph.Transaction, node *graph.Node) (graph.PathSet, error) {
	if domainSID, err := getNodeDomainSIDOrObjectID(node); err != nil {
		return nil, err
	} else if gpoIDs, err := ops.FetchNodeIDs(tx.Nodes().Filterf(func() graph.Criteria {
		return query.And(
			query.Kind(query.Node(), ad.GPO),
			query.Equals(query.NodeProperty(ad.DomainSID.String()), domainSID),
		)
	})); err != nil {
		return nil, err
	} else {
		if directControllers, err := ops.FetchPathSet(tx.Relationships().Filterf(func() graph.Criteria {
			return query.And(
				query.InIDs(query.EndID(), gpoIDs...),
				query.KindIn(query.Relationship(), ad.ACLRelationships()...),
				query.Not(query.Equals(query.StartProperty(ad.DomainSID.String()), domainSID)),
			)
		})); err != nil {
			return nil, err
		} else if groups, err := ops.FetchStartNodes(tx.Relationships().Filterf(func() graph.Criteria {
			return query.And(
				query.Kind(query.Start(), ad.Group),
				query.KindIn(query.Relationship(), ad.ACLRelationships()...),
				query.Equals(query.EndProperty(ad.DomainSID.String()), domainSID),
				query.Kind(query.End(), ad.GPO),
			)
		})); err != nil {
			return nil, err
		} else {
			pathSet := graph.NewPathSet()
			for _, group := range groups {
				if paths, err := ops.TraversePaths(tx, ops.TraversalPlan{
					Root:      group,
					Direction: graph.DirectionInbound,
					BranchQuery: func() graph.Criteria {
						return query.Or(
							query.Kind(query.Start(), ad.Group),
							query.Not(query.Equals(query.StartProperty(ad.DomainSID.String()), domainSID)),
						)
					},
				}); err != nil {
					return nil, err
				} else {
					pathSet.AddPathSet(paths)
				}
			}

			pathSet.AddPathSet(directControllers)
			return pathSet, nil
		}
	}
}

func FetchForeignAdmins(tx graph.Transaction, node *graph.Node, skip, limit int) (graph.NodeSet, error) {
	if domainSID, err := getNodeDomainSIDOrObjectID(node); err != nil {
		return nil, err
	} else {
		if directAdmins, err := ops.FetchStartNodes(tx.Relationships().Filterf(func() graph.Criteria {
			return query.And(
				query.Kind(query.End(), ad.Computer),
				query.Kind(query.Relationship(), ad.AdminTo),
				query.Equals(query.EndProperty(ad.DomainSID.String()), domainSID),
				query.Not(query.Equals(query.StartProperty(ad.DomainSID.String()), domainSID)),
			)
		})); err != nil {
			return nil, err
		} else if groups, err := ops.FetchStartNodes(tx.Relationships().Filterf(func() graph.Criteria {
			return query.And(
				query.Kind(query.Start(), ad.Group),
				query.Kind(query.Relationship(), ad.AdminTo),
				query.Equals(query.EndProperty(ad.DomainSID.String()), domainSID),
				query.Kind(query.End(), ad.Computer),
			)
		})); err != nil {
			return nil, err
		} else {
			nodeSet := graph.NewNodeSet()

			for _, group := range groups {
				if nodes, err := ops.AcyclicTraverseTerminals(tx, ops.TraversalPlan{
					Root:      group,
					Direction: graph.DirectionInbound,
					BranchQuery: func() graph.Criteria {
						return query.Or(
							query.Kind(query.Start(), ad.Group),
							query.Not(query.Equals(query.StartProperty(ad.DomainSID.String()), domainSID)),
						)
					},
				}); err != nil {
					return nil, err
				} else {
					nodeSet.AddSet(nodes)
				}
			}

			nodeSet.AddSet(directAdmins)
			return graph.SortAndSliceNodeSet(nodeSet, skip, limit), nil
		}
	}
}

func FetchForeignAdminPaths(tx graph.Transaction, node *graph.Node) (graph.PathSet, error) {
	if domainSID, err := getNodeDomainSIDOrObjectID(node); err != nil {
		return nil, err
	} else {
		if directAdmins, err := ops.FetchPathSet(tx.Relationships().Filterf(func() graph.Criteria {
			return query.And(
				query.Kind(query.End(), ad.Computer),
				query.Kind(query.Relationship(), ad.AdminTo),
				query.Equals(query.EndProperty(ad.DomainSID.String()), domainSID),
				query.Not(query.Equals(query.StartProperty(ad.DomainSID.String()), domainSID)),
			)
		})); err != nil {
			return nil, err
		} else if groups, err := ops.FetchStartNodes(tx.Relationships().Filterf(func() graph.Criteria {
			return query.And(
				query.Kind(query.Start(), ad.Group),
				query.Kind(query.Relationship(), ad.AdminTo),
				query.Equals(query.EndProperty(ad.DomainSID.String()), domainSID),
				query.Kind(query.End(), ad.Computer),
			)
		})); err != nil {
			return nil, err
		} else {
			pathSet := graph.NewPathSet()

			for _, group := range groups {
				if paths, err := ops.TraversePaths(tx, ops.TraversalPlan{
					Root:      group,
					Direction: graph.DirectionInbound,
					BranchQuery: func() graph.Criteria {
						return query.Or(
							query.Kind(query.Start(), ad.Group),
							query.Not(query.Equals(query.StartProperty(ad.DomainSID.String()), domainSID)),
						)
					},
				}); err != nil {
					return nil, err
				} else {
					pathSet.AddPathSet(paths)
				}
			}

			pathSet.AddPathSet(directAdmins)
			return pathSet, nil
		}
	}
}

// TODO: This query appears to be slow
func CreateForeignEntityMembershipListDelegate(kind graph.Kind) analysis.ListDelegate {
	return func(tx graph.Transaction, node *graph.Node, skip, limit int) (graph.NodeSet, error) {
		foreignNodes := graph.NewNodeSet()
		if domainSID, err := getNodeDomainSIDOrObjectID(node); err != nil {
			return nil, err
		} else if nodes, err := ops.FetchNodes(tx.Nodes().Filterf(func() graph.Criteria {
			return query.And(
				query.Kind(query.Node(), ad.Group),
				query.Equals(query.NodeProperty(ad.DomainSID.String()), domainSID),
			)
		})); err != nil {
			return nil, err
		} else {
			for _, node := range nodes {
				if n, err := ops.AcyclicTraverseNodes(tx, ops.TraversalPlan{
					Root:      node,
					Direction: graph.DirectionInbound,
					BranchQuery: func() graph.Criteria {
						return query.Kind(query.Relationship(), ad.MemberOf)
					},
				}, func(node *graph.Node) bool {
					if !node.Kinds.ContainsOneOf(kind) {
						return false
					} else if s, err := node.Properties.Get(ad.DomainSID.String()).String(); err != nil {
						return false
					} else if s == domainSID {
						return false
					} else {
						return true
					}
				}); err != nil {
					return nil, err
				} else {
					foreignNodes.AddSet(n)
				}
			}

			return graph.SortAndSliceNodeSet(foreignNodes, skip, limit), nil
		}
	}
}

func CreateForeignEntityMembershipPathDelegate(kind graph.Kind) analysis.PathDelegate {
	return func(tx graph.Transaction, node *graph.Node) (graph.PathSet, error) {
		foreignPaths := graph.NewPathSet()

		if domainSID, err := getNodeDomainSIDOrObjectID(node); err != nil {
			return nil, err
		} else if nodes, err := ops.FetchNodes(tx.Nodes().Filterf(func() graph.Criteria {
			return query.And(
				query.Kind(query.Node(), ad.Group),
				query.Equals(query.NodeProperty(ad.DomainSID.String()), domainSID),
			)
		})); err != nil {
			return nil, err
		} else {
			for _, node := range nodes {
				if n, err := ops.TraverseIntermediaryPaths(tx, ops.TraversalPlan{
					Root:      node,
					Direction: graph.DirectionInbound,
					BranchQuery: func() graph.Criteria {
						return query.Kind(query.Relationship(), ad.MemberOf)
					},
				}, func(node *graph.Node) bool {
					if !node.Kinds.ContainsOneOf(kind) {
						return false
					} else if s, err := node.Properties.Get(ad.DomainSID.String()).String(); err != nil {
						return false
					} else if s == domainSID {
						return false
					} else {
						return true
					}
				}); err != nil {
					return nil, err
				} else {
					foreignPaths.AddPathSet(n)
				}
			}

			return foreignPaths, nil
		}
	}
}

func FetchEntityLinkedGPOList(tx graph.Transaction, node *graph.Node, skip, limit int) (graph.NodeSet, error) {
	return ops.AcyclicTraverseTerminals(tx, ops.TraversalPlan{
		Root:      node,
		Direction: graph.DirectionInbound,
		BranchQuery: func() graph.Criteria {
			return query.And(
				query.Kind(query.Relationship(), ad.GPLink),
				query.Kind(query.Start(), ad.GPO),
			)
		},
		Skip:  skip,
		Limit: limit,
	})
}

func FetchEntityLinkedGPOPaths(tx graph.Transaction, node *graph.Node) (graph.PathSet, error) {
	return ops.TraversePaths(tx, ops.TraversalPlan{
		Root:      node,
		Direction: graph.DirectionInbound,
		BranchQuery: func() graph.Criteria {
			return query.And(
				query.Kind(query.Relationship(), ad.GPLink),
				query.Kind(query.Start(), ad.GPO),
			)
		},
	})
}

func CreateInboundLocalGroupListDelegate(edge graph.Kind) analysis.ListDelegate {
	return func(tx graph.Transaction, node *graph.Node, skip, limit int) (graph.NodeSet, error) {
		return ops.AcyclicTraverseTerminals(tx, ops.TraversalPlan{
			Root:      node,
			Direction: graph.DirectionInbound,
			BranchQuery: func() graph.Criteria {
				return query.KindIn(query.Relationship(), ad.MemberOf, edge)
			},
			DescentFilter: func(ctx *ops.TraversalContext, segment *graph.PathSegment) bool {
				if segment.Depth() > 1 && segment.Trunk.Node.Kinds.ContainsOneOf(ad.Computer, ad.User) {
					return false
				}

				return true
			},
			Skip:  skip,
			Limit: limit,
		})
	}
}

func CreateInboundLocalGroupPathDelegate(edge graph.Kind) analysis.PathDelegate {
	return func(tx graph.Transaction, node *graph.Node) (graph.PathSet, error) {
		return ops.TraversePaths(tx, ops.TraversalPlan{
			Root:      node,
			Direction: graph.DirectionInbound,
			BranchQuery: func() graph.Criteria {
				return query.KindIn(query.Relationship(), ad.MemberOf, edge)
			},
			DescentFilter: func(ctx *ops.TraversalContext, segment *graph.PathSegment) bool {
				if segment.Depth() > 1 && segment.Trunk.Node.Kinds.ContainsOneOf(ad.Computer, ad.User) {
					return false
				}

				return true
			},
		})
	}
}

func CreateOutboundLocalGroupListDelegate(edge graph.Kind) analysis.ListDelegate {
	return func(tx graph.Transaction, node *graph.Node, skip, limit int) (graph.NodeSet, error) {
		return ops.AcyclicTraverseTerminals(tx, ops.TraversalPlan{
			Root:      node,
			Direction: graph.DirectionOutbound,
			BranchQuery: func() graph.Criteria {
				return query.KindIn(query.Relationship(), ad.MemberOf, edge)
			},
			DescentFilter: func(ctx *ops.TraversalContext, segment *graph.PathSegment) bool {
				if segment.Depth() > 1 && segment.Trunk.Node.Kinds.ContainsOneOf(ad.Computer) {
					return false
				}

				return true
			},
			PathFilter: func(ctx *ops.TraversalContext, segment *graph.PathSegment) bool {
				return segment.Node.Kinds.ContainsOneOf(ad.Computer)
			},
		})
	}
}

func CreateOutboundLocalGroupPathDelegate(edge graph.Kind) analysis.PathDelegate {
	return func(tx graph.Transaction, node *graph.Node) (graph.PathSet, error) {
		return ops.TraversePaths(tx, ops.TraversalPlan{
			Root:      node,
			Direction: graph.DirectionOutbound,
			BranchQuery: func() graph.Criteria {
				return query.KindIn(query.Relationship(), ad.MemberOf, edge)
			},
			DescentFilter: func(ctx *ops.TraversalContext, segment *graph.PathSegment) bool {
				if segment.Depth() > 1 && segment.Trunk.Node.Kinds.ContainsOneOf(ad.Computer) {
					return false
				}

				return true
			},
			PathFilter: func(ctx *ops.TraversalContext, segment *graph.PathSegment) bool {
				return segment.Node.Kinds.ContainsOneOf(ad.Computer)
			},
		})
	}
}

func CreateSQLAdminPathDelegate(direction graph.Direction) analysis.PathDelegate {
	return func(tx graph.Transaction, node *graph.Node) (graph.PathSet, error) {
		return ops.TraversePaths(tx, ops.TraversalPlan{
			Root:      node,
			Direction: direction,
			BranchQuery: func() graph.Criteria {
				return query.Kind(query.Relationship(), ad.SQLAdmin)
			},
		})
	}
}

func CreateSQLAdminListDelegate(direction graph.Direction) analysis.ListDelegate {
	return func(tx graph.Transaction, node *graph.Node, skip, limit int) (graph.NodeSet, error) {
		return ops.AcyclicTraverseTerminals(tx, ops.TraversalPlan{
			Root:      node,
			Direction: direction,
			BranchQuery: func() graph.Criteria {
				return query.Kind(query.Relationship(), ad.SQLAdmin)
			},
			Skip:  skip,
			Limit: limit,
		})
	}
}

func CreateConstrainedDelegationPathDelegate(direction graph.Direction) analysis.PathDelegate {
	return func(tx graph.Transaction, node *graph.Node) (graph.PathSet, error) {
		return ops.TraversePaths(tx, ops.TraversalPlan{
			Root:      node,
			Direction: direction,
			BranchQuery: func() graph.Criteria {
				return query.Kind(query.Relationship(), ad.AllowedToDelegate)
			},
		})
	}
}

func CreateConstrainedDelegationListDelegate(direction graph.Direction) analysis.ListDelegate {
	return func(tx graph.Transaction, node *graph.Node, skip, limit int) (graph.NodeSet, error) {
		return ops.AcyclicTraverseTerminals(tx, ops.TraversalPlan{
			Root:      node,
			Direction: direction,
			BranchQuery: func() graph.Criteria {
				return query.Kind(query.Relationship(), ad.AllowedToDelegate)
			},
			Skip:  skip,
			Limit: limit,
		})
	}
}

func FetchGroupSessions(tx graph.Transaction, node *graph.Node, skip, limit int) (graph.NodeSet, error) {
	return ops.AcyclicTraverseTerminals(tx, ops.TraversalPlan{
		Root:      node,
		Direction: graph.DirectionInbound,
		BranchQuery: func() graph.Criteria {
			return query.KindIn(query.Relationship(), ad.HasSession, ad.MemberOf)
		},
		Skip:  skip,
		Limit: limit,
		PathFilter: func(ctx *ops.TraversalContext, segment *graph.PathSegment) bool {
			return segment.Edge.Kind.Is(ad.HasSession) && segment.Node.Kinds.ContainsOneOf(ad.Computer)
		},
	})
}

func FetchGroupSessionPaths(tx graph.Transaction, node *graph.Node) (graph.PathSet, error) {
	return ops.TraversePaths(tx, ops.TraversalPlan{
		Root:      node,
		Direction: graph.DirectionInbound,
		BranchQuery: func() graph.Criteria {
			return query.KindIn(query.Relationship(), ad.HasSession, ad.MemberOf)
		},
		PathFilter: func(ctx *ops.TraversalContext, segment *graph.PathSegment) bool {
			return segment.Edge.Kind.Is(ad.HasSession) && segment.Node.Kinds.ContainsOneOf(ad.Computer)
		},
	})
}

func FetchUserSessionPaths(tx graph.Transaction, node *graph.Node) (graph.PathSet, error) {
	return ops.TraversePaths(tx, ops.TraversalPlan{
		Root:        node,
		Direction:   graph.DirectionInbound,
		BranchQuery: FilterSessions,
	})
}

func FetchUserSessions(tx graph.Transaction, node *graph.Node, skip, limit int) (graph.NodeSet, error) {
	return ops.AcyclicTraverseTerminals(tx, ops.TraversalPlan{
		Root:        node,
		Direction:   graph.DirectionInbound,
		BranchQuery: FilterSessions,
		Skip:        skip,
		Limit:       limit,
	})
}

func FetchComputerSessionPaths(tx graph.Transaction, node *graph.Node) (graph.PathSet, error) {
	return ops.TraversePaths(tx, ops.TraversalPlan{
		Root:        node,
		Direction:   graph.DirectionOutbound,
		BranchQuery: FilterSessions,
	})
}

func FetchComputerSessions(tx graph.Transaction, node *graph.Node, skip, limit int) (graph.NodeSet, error) {
	return ops.AcyclicTraverseTerminals(tx, ops.TraversalPlan{
		Root:        node,
		Direction:   graph.DirectionOutbound,
		BranchQuery: FilterSessions,
		Skip:        skip,
		Limit:       limit,
	})
}

func FetchEntityGroupMembershipPaths(tx graph.Transaction, node *graph.Node) (graph.PathSet, error) {
	return ops.TraversePaths(tx, ops.TraversalPlan{
		Root:        node,
		Direction:   graph.DirectionOutbound,
		BranchQuery: FilterGroupMembership,
	})
}

func FetchEntityGroupMembership(tx graph.Transaction, root *graph.Node, skip, limit int) (graph.NodeSet, error) {
	return ops.AcyclicTraverseNodes(tx, ops.TraversalPlan{
		Root:        root,
		Direction:   graph.DirectionOutbound,
		Skip:        skip,
		Limit:       limit,
		BranchQuery: FilterGroupMembership,
	}, func(node *graph.Node) bool {
		return node.ID != root.ID
	})
}

func FetchInboundADEntityControllerPaths(ctx context.Context, db graph.Database, node *graph.Node) (graph.PathSet, error) {
	var (
		traversalInstance = traversal.New(db, analysis.MaximumDatabaseParallelWorkers)
		collector         = traversal.NewPathCollector()
	)

	if err := traversalInstance.BreadthFirst(ctx, traversal.Plan{
		Root: node,
		Driver: traversal.LightweightDriver(
			graph.DirectionInbound,
			graphcache.New(),
			query.KindIn(query.Relationship(), append(ad.ACLRelationships(), ad.MemberOf)...),
			InboundControllerPaths(),
			func(next *graph.PathSegment) {
				if IsValidInboundControllerPath(next) {
					collector.Add(next.Path())
				}
			},
		),
	}); err != nil {
		return nil, err
	}

	return collector.Paths, collector.PopulateNodeProperties(ctx, db, common.Name.String(), common.ObjectID.String(), common.SystemTags.String())
}

func FetchInboundADEntityControllers(ctx context.Context, db graph.Database, node *graph.Node, skip, limit int) (graph.NodeSet, error) {
	var (
		traversalInstance = traversal.New(db, analysis.MaximumDatabaseParallelWorkers)
		collector         = traversal.NewNodeCollector()
	)

	if err := traversalInstance.BreadthFirst(ctx, traversal.Plan{
		Root: node,
		Driver: traversal.LightweightDriver(
			graph.DirectionInbound,
			graphcache.New(),
			query.KindIn(query.Relationship(), append(ad.ACLRelationships(), ad.MemberOf)...),
			InboundControllerNodes(collector, skip, limit),
		),
	}); err != nil {
		return nil, err
	}

	return collector.Nodes, collector.PopulateProperties(ctx, db, common.Name.String(), common.ObjectID.String(), common.SystemTags.String())
}

func FetchOutboundADEntityControlPaths(ctx context.Context, db graph.Database, node *graph.Node) (graph.PathSet, error) {
	var (
		traversalInstance = traversal.New(db, analysis.MaximumDatabaseParallelWorkers)
		collector         = traversal.NewPathCollector()
	)

	if err := traversalInstance.BreadthFirst(ctx, traversal.Plan{
		Root: node,
		Driver: traversal.LightweightDriver(
			graph.DirectionOutbound,
			graphcache.New(),
			query.KindIn(query.Relationship(), append(ad.ACLRelationships(), ad.MemberOf)...),
			OutboundControlledPaths(collector),
		),
	}); err != nil {
		return nil, err
	}

	return collector.Paths, collector.PopulateNodeProperties(ctx, db, common.Name.String(), common.ObjectID.String(), common.SystemTags.String())
}

func FetchOutboundADEntityControl(ctx context.Context, db graph.Database, node *graph.Node, skip, limit int) (graph.NodeSet, error) {
	var (
		traversalInstance = traversal.New(db, analysis.MaximumDatabaseParallelWorkers)
		collector         = traversal.NewNodeCollector()
	)

	if err := traversalInstance.BreadthFirst(ctx, traversal.Plan{
		Root: node,
		Driver: traversal.LightweightDriver(
			graph.DirectionOutbound,
			graphcache.New(),
			query.KindIn(query.Relationship(), append(ad.ACLRelationships(), ad.MemberOf)...),
			OutboundControlledNodes(collector, skip, limit),
		),
	}); err != nil {
		return nil, err
	}

	return collector.Nodes, collector.PopulateProperties(ctx, db, common.Name.String(), common.ObjectID.String(), common.SystemTags.String())
}

func FetchPolicyLinkedCertTemplatePaths(tx graph.Transaction, root *graph.Node) (graph.PathSet, error) {
	return ops.TraversePaths(tx, ops.TraversalPlan{
		Root:      root,
		Direction: graph.DirectionInbound,
		BranchQuery: func() graph.Criteria {
			return query.And(
				query.Kind(query.Start(), ad.CertTemplate),
				query.Kind(query.Relationship(), ad.ExtendedByPolicy),
			)
		},
	})
}

func FetchPolicyLinkedCertTemplates(tx graph.Transaction, root *graph.Node, skip, limit int) (graph.NodeSet, error) {
	return ops.AcyclicTraverseNodes(tx, ops.TraversalPlan{
		Root:      root,
		Direction: graph.DirectionInbound,
		BranchQuery: func() graph.Criteria {
			return query.And(
				query.Kind(query.Start(), ad.CertTemplate),
				query.Kind(query.Relationship(), ad.ExtendedByPolicy),
			)
		},
		Skip:  skip,
		Limit: limit,
	}, func(node *graph.Node) bool {
		return node.ID != root.ID
	})
}

func FetchLinkedGroup(ctx context.Context, db graph.Database, node *graph.Node) (*graph.Node, error) {
	var linkedNode *graph.Node
	return linkedNode, db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		if endNodes, err := ops.FetchEndNodes(tx.Relationships().Filterf(func() graph.Criteria {
			return query.And(
				query.InIDs(query.StartID(), node.ID),
				query.Kind(query.Relationship(), ad.OIDGroupLink),
				query.Kind(query.End(), ad.Group),
			)
		})); err != nil {
			if graph.IsErrNotFound(err) {
				return nil
			}
			return err
		} else {
			// Pick is safe because there should only ever be one node here
			linkedNode = endNodes.Pick()
			return nil
		}
	})
}

func FetchGroupMemberPaths(tx graph.Transaction, node *graph.Node) (graph.PathSet, error) {
	return ops.TraversePaths(tx, ops.TraversalPlan{
		Root:        node,
		Direction:   graph.DirectionInbound,
		BranchQuery: FilterGroupMembers,
	})
}

func FetchGroupMembers(ctx context.Context, db graph.Database, root *graph.Node, skip, limit int) (graph.NodeSet, error) {
	collector := traversal.NewNodeCollector()

	if err := traversal.New(db, analysis.MaximumDatabaseParallelWorkers).BreadthFirst(ctx, traversal.Plan{
		Root: root,
		Driver: traversal.LightweightDriver(
			graph.DirectionInbound,
			graphcache.New(),
			query.Kind(query.Relationship(), ad.MemberOf),
			traversal.AcyclicNodeFilter(
				traversal.FilteredSkipLimit(
					func(next *graph.PathSegment) (bool, bool) {
						return true, next.Node.Kinds.ContainsOneOf(ad.Group)
					},
					collector.Collect,
					skip,
					limit,
				),
			),
		),
	}); err != nil {
		return nil, err
	}

	return collector.Nodes, db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		return ops.FetchAllNodeProperties(tx, collector.Nodes)
	})
}

const (
	windows      = "WINDOWS"
	fourteenDays = time.Hour * 24 * 14
)

func FetchLocalGroupCompleteness(tx graph.Transaction, domainSIDs ...string) (float64, error) {
	completeness := float64(0)

	if computers, err := ops.FetchNodeSet(tx.Nodes().Filterf(func() graph.Criteria {
		filters := []graph.Criteria{
			query.Kind(query.Node(), ad.Computer),
			query.StringContains(query.NodeProperty(common.OperatingSystem.String()), windows),
			query.Equals(query.NodeProperty(common.Enabled.String()), true),
			query.Exists(query.NodeProperty(ad.LastLogonTimestamp.String())),
		}

		if len(domainSIDs) > 0 {
			filters = append(filters, query.In(query.NodeProperty(ad.DomainSID.String()), domainSIDs))
		}

		return query.And(filters...)
	})); err != nil {
		return completeness, err
	} else {
		mostRecentLogonTimestamp := time.Unix(0, 0)

		for _, computer := range computers {
			if lastLogonTimestamp, err := computer.Properties.Get(ad.LastLogonTimestamp.String()).Time(); err != nil {
				return completeness, err
			} else if lastLogonTimestamp.After(mostRecentLogonTimestamp) {
				mostRecentLogonTimestamp = lastLogonTimestamp
			}
		}

		activityThreshold := mostRecentLogonTimestamp.Add(-fourteenDays)

		for _, computer := range computers {
			if passwordLastSet, err := computer.Properties.Get(ad.LastLogonTimestamp.String()).Time(); err != nil {
				return completeness, err
			} else if passwordLastSet.Before(activityThreshold) {
				computers.Remove(computer.ID)
			}
		}

		if computers.Len() == 0 {
			return 0, nil
		}

		var (
			activeComputerCount           = float64(computers.Len())
			activeComputerCountWithAdmins = float64(0)
			computerBmp                   = cardinality.NewBitmap64()
		)

		if err := tx.Relationships().Filterf(func() graph.Criteria {
			return query.And(
				query.KindIn(query.Relationship(), ad.AdminTo, ad.CanRDP, ad.CanPSRemote, ad.ExecuteDCOM, ad.CanBackup),
				query.InIDs(query.EndID(), computers.IDs()...),
			)
		}).Fetch(func(cursor graph.Cursor[*graph.Relationship]) error {
			for rel := range cursor.Chan() {
				computerBmp.Add(rel.EndID.Uint64())
			}

			return nil
		}); err != nil {
			return completeness, err
		} else {
			activeComputerCountWithAdmins += float64(computerBmp.Cardinality())
		}

		return activeComputerCountWithAdmins / activeComputerCount, nil
	}
}

func FetchUserSessionCompleteness(tx graph.Transaction, domainSIDs ...string) (float64, error) {
	completeness := float64(0)

	if users, err := ops.FetchNodeSet(tx.Nodes().Filterf(func() graph.Criteria {
		filters := []graph.Criteria{
			query.Kind(query.Node(), ad.User),
			query.Equals(query.NodeProperty(common.Enabled.String()), true),
			query.Exists(query.NodeProperty(ad.LastLogonTimestamp.String())),
		}

		if len(domainSIDs) > 0 {
			filters = append(filters, query.Equals(query.NodeProperty(ad.DomainSID.String()), domainSIDs[0]))
		}

		return query.And(filters...)
	})); err != nil {
		return completeness, err
	} else {
		mostRecentLogonTimestamp := time.Unix(0, 0)

		for _, user := range users {
			if userLastLogonTimestamp, err := user.Properties.Get(ad.LastLogonTimestamp.String()).Time(); err != nil {
				return completeness, err
			} else if userLastLogonTimestamp.After(mostRecentLogonTimestamp) {
				mostRecentLogonTimestamp = userLastLogonTimestamp
			}
		}

		activityThreshold := mostRecentLogonTimestamp.Add(-fourteenDays)

		for _, user := range users {
			if userLastLogonTimestamp, err := user.Properties.Get(ad.LastLogonTimestamp.String()).Time(); err != nil {
				return completeness, err
			} else if userLastLogonTimestamp.Before(activityThreshold) {
				users.Remove(user.ID)
			}
		}

		if users.Len() == 0 {
			return 0, nil
		}

		var (
			activeUserCount             = float64(users.Len())
			activeUserCountWithSessions = float64(0)
			userBmp                     = roaring64.NewBitmap()
		)

		if err := tx.Relationships().Filterf(func() graph.Criteria {
			return query.And(
				// Only computers may have a HasSession relationship to the user so no need to filter on start kind
				query.Kind(query.Relationship(), ad.HasSession),
				query.InIDs(query.EndID(), users.IDs()...),
			)
		}).Fetch(func(cursor graph.Cursor[*graph.Relationship]) error {
			for rel := range cursor.Chan() {
				userBmp.Add(rel.EndID.Uint64())
			}

			return nil
		}); err != nil {
			return completeness, err
		} else {
			activeUserCountWithSessions += float64(userBmp.GetCardinality())
		}

		return activeUserCountWithSessions / activeUserCount, nil
	}
}

func FetchAllGroupMembers(ctx context.Context, db graph.Database, targets graph.NodeSet) (graph.NodeSet, error) {
	defer measure.ContextMeasure(ctx, slog.LevelInfo, "FetchAllGroupMembers")()

	slog.InfoContext(ctx, fmt.Sprintf("Fetching group members for %d AD nodes", len(targets)))

	allGroupMembers := graph.NewNodeSet()

	for _, target := range targets {
		if target.Kinds.ContainsOneOf(ad.Group) {
			if groupMembers, err := FetchGroupMembers(ctx, db, target, 0, 0); err != nil {
				return nil, err
			} else {
				allGroupMembers.AddSet(groupMembers)
			}
		}
	}

	slog.InfoContext(ctx, fmt.Sprintf("Collected %d group members", len(allGroupMembers)))
	return allGroupMembers, nil
}

func FetchCertTemplatesPublishedToCA(tx graph.Transaction, ca *graph.Node) (graph.NodeSet, error) {
	return ops.FetchStartNodes(tx.Relationships().Filterf(func() graph.Criteria {
		return query.And(
			query.Equals(query.EndID(), ca.ID),
			query.Kind(query.Relationship(), ad.PublishedTo),
			query.Kind(query.Start(), ad.CertTemplate),
		)
	}))
}

func FetchNodesWithSameForestTrustRelationship(tx graph.Transaction, root *graph.Node) (graph.NodeSet, error) {
	if pathSet, err := ops.TraversePaths(tx, ops.TraversalPlan{
		Root:      root,
		Direction: graph.DirectionInbound,
		BranchQuery: func() graph.Criteria {
			return query.And(
				query.KindIn(query.Start(), ad.Domain),
				query.KindIn(query.Relationship(), ad.SameForestTrust),
			)
		},
	}); err != nil {
		return nil, err
	} else {
		alldomains := pathSet.AllNodes()
		if alldomains.Len() == 0 {
			alldomains.Add(root)
		}
		return alldomains, nil
	}
}

func FetchNodesWithDCForEdge(tx graph.Transaction, rootNode *graph.Node) (graph.NodeSet, error) {
	return ops.AcyclicTraverseTerminals(tx, ops.TraversalPlan{
		Root:      rootNode,
		Direction: graph.DirectionInbound,
		BranchQuery: func() graph.Criteria {
			return query.And(
				query.KindIn(query.Start(), ad.Computer),
				query.KindIn(query.Relationship(), ad.DCFor),
			)
		},
	})
}

func FetchEnterpriseCAsCertChainPathToDomain(tx graph.Transaction, enterpriseCA, domain *graph.Node) (graph.PathSet, error) {
	return ops.TraversePaths(tx, ops.TraversalPlan{
		Root:      enterpriseCA,
		Direction: graph.DirectionOutbound,
		BranchQuery: func() graph.Criteria {
			return query.KindIn(query.Relationship(), ad.IssuedSignedBy, ad.EnterpriseCAFor, ad.RootCAFor)
		},
		DescentFilter: func(ctx *ops.TraversalContext, segment *graph.PathSegment) bool {
			return !segment.Trunk.Node.Kinds.ContainsOneOf(ad.Domain)
		},
		PathFilter: func(ctx *ops.TraversalContext, segment *graph.PathSegment) bool {
			return segment.Node.ID == domain.ID
		},
	})
}

func FetchEnterpriseCAsTrustedForAuthPathToDomain(tx graph.Transaction, enterpriseCA, domain *graph.Node) (graph.PathSet, error) {
	return ops.TraversePaths(tx, ops.TraversalPlan{
		Root:      enterpriseCA,
		Direction: graph.DirectionOutbound,
		BranchQuery: func() graph.Criteria {
			return query.KindIn(query.Relationship(), ad.TrustedForNTAuth, ad.NTAuthStoreFor)
		},
		PathFilter: func(ctx *ops.TraversalContext, segment *graph.PathSegment) bool {
			return segment.Node.ID == domain.ID
		},
	})
}

func FetchHostsCAServiceComputers(tx graph.Transaction, enterpriseCA *graph.Node) (graph.NodeSet, error) {
	return ops.FetchStartNodes(tx.Relationships().Filter(
		query.And(
			query.Kind(query.Start(), ad.Computer),
			query.Kind(query.Relationship(), ad.HostsCAService),
			query.Equals(query.EndID(), enterpriseCA.ID),
		)))
}

func FetchEnterpriseCAsTrustedForNTAuthToDomain(tx graph.Transaction, domain *graph.Node) (graph.NodeSet, error) {
	return ops.AcyclicTraverseTerminals(tx, ops.TraversalPlan{
		Root:      domain,
		Direction: graph.DirectionInbound,
		BranchQuery: func() graph.Criteria {
			return query.KindIn(query.Relationship(), ad.TrustedForNTAuth, ad.NTAuthStoreFor)
		},
		DescentFilter: func(ctx *ops.TraversalContext, segment *graph.PathSegment) bool {
			depth := segment.Depth()
			if depth == 1 && !segment.Edge.Kind.Is(ad.NTAuthStoreFor) {
				return false
			} else if depth == 2 && !segment.Edge.Kind.Is(ad.TrustedForNTAuth) {
				return false
			} else {
				return true
			}
		},
		PathFilter: func(ctx *ops.TraversalContext, segment *graph.PathSegment) bool {
			return segment.Node.Kinds.ContainsOneOf(ad.EnterpriseCA)
		},
	})
}

// This was created to aid in the new post composition methodology, but was ultimately scrapped.
// Leaving the code here for future use as full paths are necessary for post based composition
func FetchEnterpriseCAsTrustedForNTAuthToDomainFull(tx graph.Transaction, domain *graph.Node) (graph.PathSet, error) {
	return ops.TraversePaths(tx, ops.TraversalPlan{
		Root:      domain,
		Direction: graph.DirectionInbound,
		BranchQuery: func() graph.Criteria {
			return query.KindIn(query.Relationship(), ad.TrustedForNTAuth, ad.NTAuthStoreFor)
		},
		DescentFilter: func(ctx *ops.TraversalContext, segment *graph.PathSegment) bool {
			depth := segment.Depth()
			if depth == 1 && !segment.Edge.Kind.Is(ad.NTAuthStoreFor) {
				return false
			} else if depth == 2 && !segment.Edge.Kind.Is(ad.TrustedForNTAuth) {
				return false
			} else {
				return true
			}
		},
		PathFilter: func(ctx *ops.TraversalContext, segment *graph.PathSegment) bool {
			return segment.Node.Kinds.ContainsOneOf(ad.EnterpriseCA)
		},
	})
}

func FetchEnterpriseCAsRootCAForPathToDomain(tx graph.Transaction, domain *graph.Node) (graph.NodeSet, error) {
	return ops.AcyclicTraverseNodes(tx, ops.TraversalPlan{
		Root:      domain,
		Direction: graph.DirectionInbound,
		BranchQuery: func() graph.Criteria {
			return query.KindIn(query.Relationship(), ad.IssuedSignedBy, ad.EnterpriseCAFor, ad.RootCAFor)
		},
	}, func(node *graph.Node) bool {
		return node.Kinds.ContainsOneOf(ad.EnterpriseCA)
	})
}

// This was created to aid in the new post composition methodology, but was ultimately scrapped.
// Leaving the code here for future use as full paths are necessary for post based composition
func FetchEnterpriseCAsRootCAForPathToDomainFull(tx graph.Transaction, domain *graph.Node) (graph.PathSet, error) {
	return ops.TraversePaths(tx, ops.TraversalPlan{
		Root:      domain,
		Direction: graph.DirectionInbound,
		BranchQuery: func() graph.Criteria {
			return query.KindIn(query.Relationship(), ad.IssuedSignedBy, ad.EnterpriseCAFor, ad.RootCAFor)
		},
		PathFilter: func(ctx *ops.TraversalContext, segment *graph.PathSegment) bool {
			return segment.Node.Kinds.ContainsOneOf(ad.EnterpriseCA)
		},
	})
}

func DoesCertTemplateLinkToDomain(tx graph.Transaction, certTemplate, domainNode *graph.Node) (bool, error) {
	if pathSet, err := FetchCertTemplatePathToDomain(tx, certTemplate, domainNode); err != nil {
		return false, err
	} else {
		return pathSet.Len() > 0, nil
	}
}

func FetchCertTemplatePathToDomain(tx graph.Transaction, certTemplate, domain *graph.Node) (graph.PathSet, error) {
	var (
		paths = graph.NewPathSet()
	)

	return paths, tx.Relationships().Filter(
		query.And(
			query.Equals(query.StartID(), certTemplate.ID),
			query.KindIn(query.Relationship(), ad.PublishedTo, ad.IssuedSignedBy, ad.EnterpriseCAFor, ad.RootCAFor),
			query.Equals(query.EndID(), domain.ID),
		),
	).FetchAllShortestPaths(func(cursor graph.Cursor[graph.Path]) error {
		for path := range cursor.Chan() {
			paths.AddPath(path)
		}

		return cursor.Error()
	})
}

// fetchFirstDegreeNodes fetches all entities that are connected to the provided targetNode with a relationship kind that matches any of the provided relKinds
func fetchFirstDegreeNodes(tx graph.Transaction, targetNode *graph.Node, relKinds ...graph.Kind) (graph.NodeSet, error) {
	return ops.FetchStartNodes(tx.Relationships().Filter(
		query.And(
			query.Kind(query.Start(), ad.Entity),
			query.KindIn(query.Relationship(), relKinds...),
			query.Equals(query.EndID(), targetNode.ID),
		),
	))
}

func FetchAttackersForEscalations9and10(tx graph.Transaction, victimBitmap cardinality.Duplex[uint64], scenarioB bool) ([]graph.ID, error) {
	if attackers, err := ops.FetchStartNodeIDs(tx.Relationships().Filterf(func() graph.Criteria {
		criteria := query.And(
			query.KindIn(query.Start(), ad.Group, ad.User, ad.Computer),
			query.KindIn(query.Relationship(), ad.GenericAll, ad.GenericWrite, ad.Owns, ad.WriteOwner, ad.WriteDACL),
			query.InIDs(query.EndID(), graph.DuplexToGraphIDs(victimBitmap)...),
		)
		if scenarioB {
			return query.And(criteria, query.KindIn(query.End(), ad.Computer))
		}
		return criteria
	})); err != nil {
		return nil, err
	} else {
		return attackers, nil
	}
}

func FetchCertTemplateCAs(tx graph.Transaction, certTemplate *graph.Node) (graph.NodeSet, error) {
	return ops.FetchEndNodes(tx.Relationships().Filter(
		FilterPublishedCAs(certTemplate),
	))
}

func FetchAuthUsersAndEveryoneGroups(tx graph.Transaction) (graph.NodeSet, error) {
	return ops.FetchNodeSet(tx.Nodes().Filterf(func() graph.Criteria {
		return query.And(
			query.Kind(query.Node(), ad.Group),
			query.Or(
				query.StringEndsWith(query.NodeProperty(common.ObjectID.String()), wellknown.AuthenticatedUsersSIDSuffix.String()),
				query.StringEndsWith(query.NodeProperty(common.ObjectID.String()), wellknown.EveryoneSIDSuffix.String()),
			),
		)
	}))
}

func getNodeDomainSIDOrObjectID(node *graph.Node) (string, error) {
	if sid, err := node.Properties.Get(ad.DomainSID.String()).String(); err == nil {
		return sid, nil
	} else if sid, err := node.Properties.Get(common.ObjectID.String()).String(); err == nil {
		return sid, nil
	} else {
		return "", err
	}
}

func CreateRootCAPKIHierarchyPathDelegate(tx graph.Transaction, node *graph.Node) (graph.PathSet, error) {
	return ops.TraversePaths(tx, ops.TraversalPlan{
		Root:      node,
		Direction: graph.DirectionInbound,
		BranchQuery: func() graph.Criteria {
			return query.KindIn(query.Relationship(), ad.EnterpriseCAFor, ad.HostsCAService, ad.IssuedSignedBy)
		},
	})
}

func CreateRootCAPKIHierarchyListDelegate(tx graph.Transaction, node *graph.Node, skip, limit int) (graph.NodeSet, error) {
	return ops.AcyclicTraverseNodes(tx, ops.TraversalPlan{
		Root:      node,
		Direction: graph.DirectionInbound,
		BranchQuery: func() graph.Criteria {
			return query.KindIn(query.Relationship(), ad.HostsCAService, ad.IssuedSignedBy, ad.EnterpriseCAFor)
		},
		Skip:  skip,
		Limit: limit,
	}, func(candidate *graph.Node) bool {
		return candidate.ID != node.ID
	})
}

func CreateCAPKIHierarchyPathDelegate(tx graph.Transaction, node *graph.Node) (graph.PathSet, error) {
	allPaths := graph.PathSet{}
	if inboundPaths, err := ops.TraversePaths(tx, ops.TraversalPlan{
		Root:      node,
		Direction: graph.DirectionInbound,
		BranchQuery: func() graph.Criteria {
			return query.KindIn(query.Relationship(), ad.EnterpriseCAFor, ad.HostsCAService, ad.IssuedSignedBy)
		},
	}); err != nil {
		return nil, err
	} else if outboundPaths, err := ops.TraversePaths(tx, ops.TraversalPlan{
		Root:      node,
		Direction: graph.DirectionOutbound,
		BranchQuery: func() graph.Criteria {
			return query.KindIn(query.Relationship(), ad.EnterpriseCAFor, ad.TrustedForNTAuth, ad.IssuedSignedBy, ad.RootCAFor, ad.NTAuthStoreFor)
		},
	}); err != nil {
		return nil, err
	} else {
		allPaths.AddPathSet(inboundPaths)
		allPaths.AddPathSet(outboundPaths)
		return allPaths, nil
	}
}

func CreateCAPKIHierarchyListDelegate(tx graph.Transaction, node *graph.Node, skip, limit int) (graph.NodeSet, error) {
	allNodes := graph.NodeSet{}

	if inboundNodes, err := ops.AcyclicTraverseNodes(tx, ops.TraversalPlan{
		Root:      node,
		Direction: graph.DirectionInbound,
		BranchQuery: func() graph.Criteria {
			return query.KindIn(query.Relationship(), ad.HostsCAService, ad.IssuedSignedBy, ad.EnterpriseCAFor)
		},
		Skip:  skip,
		Limit: limit,
	}, func(candidate *graph.Node) bool {
		return candidate.ID != node.ID
	}); err != nil {
		return nil, err
	} else if outboundNodes, err := ops.AcyclicTraverseNodes(tx, ops.TraversalPlan{
		Root:      node,
		Direction: graph.DirectionOutbound,
		BranchQuery: func() graph.Criteria {
			return query.KindIn(query.Relationship(), ad.EnterpriseCAFor, ad.TrustedForNTAuth, ad.IssuedSignedBy, ad.RootCAFor, ad.NTAuthStoreFor)
		},
		Skip:  skip,
		Limit: limit,
	}, func(candidate *graph.Node) bool {
		return candidate.ID != node.ID
	}); err != nil {
		return nil, err
	} else {
		allNodes.AddSet(inboundNodes)
		allNodes.AddSet(outboundNodes)
		return allNodes, nil
	}
}

func CreatePublishedTemplatesPathDelegate(tx graph.Transaction, node *graph.Node) (graph.PathSet, error) {
	return ops.TraversePaths(tx, ops.TraversalPlan{
		Root:      node,
		Direction: graph.DirectionInbound,
		BranchQuery: func() graph.Criteria {
			return query.Kind(query.Relationship(), ad.PublishedTo)
		},
	})
}

func CreatePublishedTemplatesListDelegate(tx graph.Transaction, node *graph.Node, skip, limit int) (graph.NodeSet, error) {
	return ops.AcyclicTraverseNodes(tx, ops.TraversalPlan{
		Root:      node,
		Direction: graph.DirectionInbound,
		BranchQuery: func() graph.Criteria {
			return query.Kind(query.Relationship(), ad.PublishedTo)
		},
		Skip:  skip,
		Limit: limit,
	}, func(candidate *graph.Node) bool {
		return candidate.ID != node.ID
	})
}

func CreatePublishedToCAsPathDelegate(tx graph.Transaction, node *graph.Node) (graph.PathSet, error) {
	return ops.TraversePaths(tx, ops.TraversalPlan{
		Root:      node,
		Direction: graph.DirectionOutbound,
		BranchQuery: func() graph.Criteria {
			return query.Kind(query.Relationship(), ad.PublishedTo)
		},
	})
}

func CreatePublishedToCAsListDelegate(tx graph.Transaction, node *graph.Node, skip, limit int) (graph.NodeSet, error) {
	return ops.AcyclicTraverseNodes(tx, ops.TraversalPlan{
		Root:      node,
		Direction: graph.DirectionOutbound,
		BranchQuery: func() graph.Criteria {
			return query.Kind(query.Relationship(), ad.PublishedTo)
		},
		Skip:  skip,
		Limit: limit,
	}, func(candidate *graph.Node) bool {
		return candidate.ID != node.ID
	})
}

func CreateTrustedCAsPathDelegate(tx graph.Transaction, node *graph.Node) (graph.PathSet, error) {
	return ops.TraversePaths(tx, ops.TraversalPlan{
		Root:      node,
		Direction: graph.DirectionInbound,
		BranchQuery: func() graph.Criteria {
			return query.Kind(query.Relationship(), ad.TrustedForNTAuth)
		},
	})
}

func CreateTrustedCAsListDelegate(tx graph.Transaction, node *graph.Node, skip, limit int) (graph.NodeSet, error) {
	return ops.AcyclicTraverseNodes(tx, ops.TraversalPlan{
		Root:      node,
		Direction: graph.DirectionInbound,
		BranchQuery: func() graph.Criteria {
			return query.Kind(query.Relationship(), ad.TrustedForNTAuth)
		},
		Skip:  skip,
		Limit: limit,
	}, func(candidate *graph.Node) bool {
		return candidate.ID != node.ID
	})
}

func CreateADCSEscalationsPathDelegate(tx graph.Transaction, node *graph.Node) (graph.PathSet, error) {
	return ops.TraversePaths(tx, ops.TraversalPlan{
		Root:      node,
		Direction: graph.DirectionInbound,
		BranchQuery: func() graph.Criteria {
			return query.Kind(query.Relationship(), ad.GoldenCert, ad.ADCSESC1, ad.ADCSESC3, ad.ADCSESC4, ad.ADCSESC6a, ad.ADCSESC6b, ad.ADCSESC9a, ad.ADCSESC9b, ad.ADCSESC10a, ad.ADCSESC10b)
		},
	})
}

func CreateADCSEscalationsListDelegate(tx graph.Transaction, node *graph.Node, skip, limit int) (graph.NodeSet, error) {
	return ops.AcyclicTraverseNodes(tx, ops.TraversalPlan{
		Root:      node,
		Direction: graph.DirectionInbound,
		BranchQuery: func() graph.Criteria {
			return query.Kind(query.Relationship(), ad.GoldenCert, ad.ADCSESC1, ad.ADCSESC3, ad.ADCSESC4, ad.ADCSESC6a, ad.ADCSESC6b, ad.ADCSESC9a, ad.ADCSESC9b, ad.ADCSESC10a, ad.ADCSESC10b)
		},
		Skip:  skip,
		Limit: limit,
	}, func(candidate *graph.Node) bool {
		return candidate.ID != node.ID
	})
}
