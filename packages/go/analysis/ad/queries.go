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
	"strings"
	"time"

	"github.com/RoaringBitmap/roaring/roaring64"
	"github.com/specterops/bloodhound/analysis"
	"github.com/specterops/bloodhound/dawgs/cardinality"
	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/dawgs/ops"
	"github.com/specterops/bloodhound/dawgs/query"
	"github.com/specterops/bloodhound/dawgs/traversal"
	"github.com/specterops/bloodhound/graphschema/ad"
	"github.com/specterops/bloodhound/graphschema/common"
)

func FetchGraphDBTierZeroTaggedAssets(tx graph.Transaction, domainSID string) (graph.NodeSet, error) {
	if nodeSet, err := ops.FetchNodeSet(tx.Nodes().Filterf(func() graph.Criteria {
		return query.And(
			query.Kind(query.Node(), ad.Entity),
			query.Equals(query.NodeProperty(ad.DomainSID.String()), domainSID),
			query.StringContains(query.NodeProperty(common.SystemTags.String()), ad.AdminTierZero),
		)
	})); err != nil {
		return nil, err
	} else {
		return nodeSet, nil
	}
}

func FetchAllEnforcedGPOs(tx graph.Transaction, targets graph.NodeSet) (graph.NodeSet, error) {
	enforcedGPOs := graph.NewNodeSet()

	for _, attackPathRoot := range targets {
		if enforced, err := FetchEnforcedGPOs(tx, attackPathRoot, 0, 0); err != nil {
			return nil, err
		} else {
			enforcedGPOs.AddSet(enforced)
		}
	}

	return enforcedGPOs, nil
}

func FetchAllDomains(ctx context.Context, db graph.Database) (graph.NodeSet, error) {
	var (
		nodes graph.NodeSet
		err   error
	)

	return nodes, db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		nodes, err = ops.FetchNodeSet(tx.Nodes().Filterf(func() graph.Criteria {
			return query.Kind(query.Node(), ad.Domain)
		}))

		return err
	})
}

func GetCollectedDomains(ctx context.Context, db graph.Database) (graph.NodeSet, error) {
	var domains graph.NodeSet
	if err := db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		if innerDomains, err := FetchCollectedDomains(tx); err != nil {
			return err
		} else {
			domains = innerDomains
			return nil
		}
	}); err != nil {
		return nil, err
	} else {
		return domains, nil
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
					//Its possible the property isn't here, so lets set enforced to false and let it roll
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
				//Its possible the property isn't here, so lets set enforced to false and let it roll
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
						if systemTags, err := segment.Node.Properties.Get(ad.AdminTierZero).String(); err != nil {
							return false
						} else {
							return strings.Contains(systemTags, ad.AdminTierZero)
						}
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
				//Its possible the property isn't here, so lets set enforced to false and let it roll
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
						//If the previous node blocks inheritance, and we've hit an OU, then the GPO is not enforced on this path, and we don't need to check any further
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
				return query.Kind(query.Relationship(), ad.TrustedBy)
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
				return query.Kind(query.Relationship(), ad.TrustedBy)
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
	if domainSID, err := node.Properties.Get(ad.DomainSID.String()).String(); err != nil {
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
	if domainSID, err := node.Properties.Get(ad.DomainSID.String()).String(); err != nil {
		return nil, err
	} else if gpoIDs, err := ops.FetchNodeIDs(tx.Nodes().Filterf(func() graph.Criteria {
		return query.And(
			query.Kind(query.Node(), ad.GPO),
			query.Equals(query.NodeProperty(ad.DomainSID.String()), domainSID),
		)
	})); err != nil {
		return nil, err
	} else {
		if directControllers, err := ops.FetchPathSet(tx, tx.Relationships().Filterf(func() graph.Criteria {
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
	if domainSid, err := node.Properties.Get(ad.DomainSID.String()).String(); err != nil {
		return nil, err
	} else {
		if directAdmins, err := ops.FetchStartNodes(tx.Relationships().Filterf(func() graph.Criteria {
			return query.And(
				query.Kind(query.End(), ad.Computer),
				query.Kind(query.Relationship(), ad.AdminTo),
				query.Equals(query.EndProperty(ad.DomainSID.String()), domainSid),
				query.Not(query.Equals(query.StartProperty(ad.DomainSID.String()), domainSid)),
			)
		})); err != nil {
			return nil, err
		} else if groups, err := ops.FetchStartNodes(tx.Relationships().Filterf(func() graph.Criteria {
			return query.And(
				query.Kind(query.Start(), ad.Group),
				query.Kind(query.Relationship(), ad.AdminTo),
				query.Equals(query.EndProperty(ad.DomainSID.String()), domainSid),
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
							query.Not(query.Equals(query.StartProperty(ad.DomainSID.String()), domainSid)),
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
	if domainSid, err := node.Properties.Get(ad.DomainSID.String()).String(); err != nil {
		return nil, err
	} else {
		if directAdmins, err := ops.FetchPathSet(tx, tx.Relationships().Filterf(func() graph.Criteria {
			return query.And(
				query.Kind(query.End(), ad.Computer),
				query.Kind(query.Relationship(), ad.AdminTo),
				query.Equals(query.EndProperty(ad.DomainSID.String()), domainSid),
				query.Not(query.Equals(query.StartProperty(ad.DomainSID.String()), domainSid)),
			)
		})); err != nil {
			return nil, err
		} else if groups, err := ops.FetchStartNodes(tx.Relationships().Filterf(func() graph.Criteria {
			return query.And(
				query.Kind(query.Start(), ad.Group),
				query.Kind(query.Relationship(), ad.AdminTo),
				query.Equals(query.EndProperty(ad.DomainSID.String()), domainSid),
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
							query.Not(query.Equals(query.StartProperty(ad.DomainSID.String()), domainSid)),
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
		if domainSid, err := node.Properties.Get(ad.DomainSID.String()).String(); err != nil {
			return nil, err
		} else if nodes, err := ops.FetchNodes(tx.Nodes().Filterf(func() graph.Criteria {
			return query.And(
				query.Kind(query.Node(), ad.Group),
				query.Equals(query.NodeProperty(ad.DomainSID.String()), domainSid),
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
					} else if s == domainSid {
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

		if domainSid, err := node.Properties.Get(ad.DomainSID.String()).String(); err != nil {
			return nil, err
		} else if nodes, err := ops.FetchNodes(tx.Nodes().Filterf(func() graph.Criteria {
			return query.And(
				query.Kind(query.Node(), ad.Group),
				query.Equals(query.NodeProperty(ad.DomainSID.String()), domainSid),
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
					} else if s == domainSid {
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
		DescentFilter: func(ctx *ops.TraversalContext, segment *graph.PathSegment) bool {
			return !(segment.Depth() > 1 && segment.Trunk.Edge.Kind.Is(ad.HasSession))
		},
		Skip:  skip,
		Limit: limit,
		PathFilter: func(ctx *ops.TraversalContext, segment *graph.PathSegment) bool {
			return segment.Node.Kinds.ContainsOneOf(ad.Computer)
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
		DescentFilter: func(ctx *ops.TraversalContext, segment *graph.PathSegment) bool {
			return !(segment.Depth() > 1 && segment.Trunk.Edge.Kind.Is(ad.HasSession))
		},
		PathFilter: func(ctx *ops.TraversalContext, segment *graph.PathSegment) bool {
			return segment.Node.Kinds.ContainsOneOf(ad.Computer)
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
			query.KindIn(query.Relationship(), append(ad.ACLRelationships(), ad.MemberOf)...),
			OutboundControlledNodes(collector, skip, limit),
		),
	}); err != nil {
		return nil, err
	}

	return collector.Nodes, collector.PopulateProperties(ctx, db, common.Name.String(), common.ObjectID.String(), common.SystemTags.String())
}

func FetchGroupMemberPaths(tx graph.Transaction, node *graph.Node) (graph.PathSet, error) {
	return ops.TraversePaths(tx, ops.TraversalPlan{
		Root:        node,
		Direction:   graph.DirectionInbound,
		BranchQuery: FilterGroupMembers,
	})
}

func FetchGroupMembers(tx graph.Transaction, root *graph.Node, skip, limit int) (graph.NodeSet, error) {
	return ops.AcyclicTraverseNodes(tx, ops.TraversalPlan{
		Root:        root,
		Direction:   graph.DirectionInbound,
		Skip:        skip,
		Limit:       limit,
		BranchQuery: FilterGroupMembership,
	}, func(node *graph.Node) bool {
		return node.ID != root.ID
	})
}

const (
	windows    = "WINDOWS"
	ninetyDays = time.Hour * 24 * 90
)

func FetchLocalGroupCompleteness(tx graph.Transaction, domainSIDs ...string) (float64, error) {
	completeness := float64(0)

	if computers, err := ops.FetchNodeSet(tx.Nodes().Filterf(func() graph.Criteria {
		filters := []graph.Criteria{
			query.Kind(query.Node(), ad.Computer),
			query.StringContains(query.NodeProperty(common.OperatingSystem.String()), windows),
			query.Exists(query.NodeProperty(common.PasswordLastSet.String())),
		}

		if len(domainSIDs) > 0 {
			filters = append(filters, query.In(query.NodeProperty(ad.DomainSID.String()), domainSIDs))
		}

		return query.And(filters...)
	})); err != nil {
		return completeness, err
	} else {
		mostRecentPasswordLastSetTime := time.Unix(0, 0)

		for _, computer := range computers {
			if passwordLastSet, err := computer.Properties.Get(common.PasswordLastSet.String()).Time(); err != nil {
				return completeness, err
			} else if passwordLastSet.After(mostRecentPasswordLastSetTime) {
				mostRecentPasswordLastSetTime = passwordLastSet
			}
		}

		activityThreshold := mostRecentPasswordLastSetTime.Add(-ninetyDays)

		for _, computer := range computers {
			if passwordLastSet, err := computer.Properties.Get(common.PasswordLastSet.String()).Time(); err != nil {
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
			computerBmp                   = cardinality.NewBitmap32()
		)

		if err := tx.Relationships().Filterf(func() graph.Criteria {
			return query.And(
				query.KindIn(query.Relationship(), ad.AdminTo, ad.CanRDP, ad.CanPSRemote, ad.ExecuteDCOM),
				query.InIDs(query.EndID(), computers.IDs()...),
			)
		}).Fetch(func(cursor graph.Cursor[*graph.Relationship]) error {
			for rel := range cursor.Chan() {
				computerBmp.Add(rel.EndID.Uint32())
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

		activityThreshold := mostRecentLogonTimestamp.Add(-ninetyDays)

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

func FetchAllGroupMembers(tx graph.Transaction, targets graph.NodeSet) (graph.NodeSet, error) {
	allGroupMembers := graph.NewNodeSet()

	for _, target := range targets {
		if target.Kinds.ContainsOneOf(ad.Group) {
			if groupMembers, err := FetchGroupMembers(tx, target, 0, 0); err != nil {
				return nil, err
			} else {
				allGroupMembers.AddSet(groupMembers)
			}
		}
	}

	return allGroupMembers, nil
}

func FetchDomainTierZeroAssets(tx graph.Transaction, domain *graph.Node) (graph.NodeSet, error) {
	domainSID, _ := domain.Properties.GetOrDefault(ad.DomainSID.String(), "").String()

	return ops.FetchNodeSet(tx.Nodes().Filterf(func() graph.Criteria {
		return query.And(
			query.Kind(query.Node(), ad.Entity),
			query.Equals(query.NodeProperty(ad.DomainSID.String()), domainSID),
			query.StringContains(query.NodeProperty(common.SystemTags.String()), ad.AdminTierZero),
		)
	}))
}
