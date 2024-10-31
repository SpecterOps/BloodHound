// Copyright 2024 Specter Ops, Inc.
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
	"github.com/specterops/bloodhound/analysis"
	"github.com/specterops/bloodhound/analysis/impact"
	"github.com/specterops/bloodhound/dawgs/cardinality"
	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/dawgs/ops"
	"github.com/specterops/bloodhound/dawgs/query"
	"github.com/specterops/bloodhound/graphschema/ad"
	"github.com/specterops/bloodhound/graphschema/common"
)

func PostOwner(ctx context.Context, db graph.Database, groupExpansions impact.PathAggregator) (*analysis.AtomicPostProcessingStats, error) {
	if dsHeuristicsCache, anyEnforced, err := GetDsHeuristicsCache(ctx, db); err != nil {
		return &analysis.AtomicPostProcessingStats{}, fmt.Errorf("failed fetching dsheuristics values for postowner: %w", err)
	} else if adminGroupIds, err := FetchAdminGroupIds(ctx, db, groupExpansions); err != nil {
		return &analysis.AtomicPostProcessingStats{}, fmt.Errorf("failed fetching admin group ids values for postowner: %w", err)
	} else {
		operation := analysis.NewPostRelationshipOperation(ctx, db, "PostOwnerWriteOwner")
		operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
			return tx.Relationships().Filter(
				query.And(
					query.Kind(query.Relationship(), ad.OwnsRaw),
					query.Kind(query.Start(), ad.Entity),
					query.Equals(query.RelationshipProperty(ad.LimitedRightsCreated.String()), false),
				),
			).Fetch(func(cursor graph.Cursor[*graph.Relationship]) error {
				for rel := range cursor.Chan() {
					if anyEnforced {
						if targetNode, err := ops.FetchNode(tx, rel.EndID); err != nil {
							continue
						} else if domainSid, err := targetNode.Properties.GetOrDefault(ad.DomainSID.String(), "").String(); err != nil {
							continue
						} else {
							enforced, ok := dsHeuristicsCache[domainSid]
							if !ok {
								enforced = false
							}

							if !enforced {
								outC <- analysis.CreatePostRelationshipJob{
									FromID:        rel.StartID,
									ToID:          rel.EndID,
									Kind:          ad.Owns,
									RelProperties: map[string]any{ad.IsACL.String(): true, ad.IsInherited.String(): rel.Properties.GetOrDefault(ad.IsInherited.String(), false)},
								}
							} else if isComputerDerived, err := isTargetNodeComputerDerived(targetNode); err != nil {
								continue
							} else if (isComputerDerived && adminGroupIds.Contains(rel.StartID.Uint64())) || !isComputerDerived {
								outC <- analysis.CreatePostRelationshipJob{
									FromID:        rel.StartID,
									ToID:          rel.EndID,
									Kind:          ad.Owns,
									RelProperties: map[string]any{ad.IsACL.String(): true, ad.IsInherited.String(): rel.Properties.GetOrDefault(ad.IsInherited.String(), false)},
								}
							}
						}
					} else {
						outC <- analysis.CreatePostRelationshipJob{
							FromID:        rel.StartID,
							ToID:          rel.EndID,
							Kind:          ad.Owns,
							RelProperties: map[string]any{ad.IsACL.String(): true, ad.IsInherited.String(): rel.Properties.GetOrDefault(ad.IsInherited.String(), false)},
						}
					}
				}

				return nil
			})
		})

		operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
			return tx.Relationships().Filter(
				query.And(
					query.Kind(query.Relationship(), ad.WriteOwnerRaw),
					query.Kind(query.Start(), ad.Entity),
					query.Equals(query.RelationshipProperty(ad.LimitedRightsCreated.String()), false),
				),
			).Fetch(func(cursor graph.Cursor[*graph.Relationship]) error {
				for rel := range cursor.Chan() {
					if anyEnforced {
						if targetNode, err := ops.FetchNode(tx, rel.EndID); err != nil {
							continue
						} else if domainSid, err := targetNode.Properties.GetOrDefault(ad.DomainSID.String(), "").String(); err != nil {
							continue
						} else {
							enforced, ok := dsHeuristicsCache[domainSid]
							if !ok {
								enforced = false
							}

							if !enforced {
								outC <- analysis.CreatePostRelationshipJob{
									FromID:        rel.StartID,
									ToID:          rel.EndID,
									Kind:          ad.WriteOwner,
									RelProperties: map[string]any{ad.IsACL.String(): true, ad.IsInherited.String(): rel.Properties.GetOrDefault(ad.IsInherited.String(), false)},
								}
							} else if isComputerDerived, err := isTargetNodeComputerDerived(targetNode); err != nil {
								continue
							} else if !isComputerDerived {
								outC <- analysis.CreatePostRelationshipJob{
									FromID:        rel.StartID,
									ToID:          rel.EndID,
									Kind:          ad.WriteOwner,
									RelProperties: map[string]any{ad.IsACL.String(): true, ad.IsInherited.String(): rel.Properties.GetOrDefault(ad.IsInherited.String(), false)},
								}
							}
						}
					} else {
						outC <- analysis.CreatePostRelationshipJob{
							FromID:        rel.StartID,
							ToID:          rel.EndID,
							Kind:          ad.WriteOwner,
							RelProperties: map[string]any{ad.IsACL.String(): true, ad.IsInherited.String(): rel.Properties.GetOrDefault(ad.IsInherited.String(), false)},
						}
					}
				}

				return nil
			})
		})

	}
	return nil, nil
}

func isTargetNodeComputerDerived(node *graph.Node) (bool, error) {
	if node.Kinds.ContainsOneOf(ad.Computer) {
		return true, nil
	} else if isGmsa, err := node.Properties.Get(ad.GMSA.String()).Bool(); err != nil {
		return false, err
	} else if isGmsa {
		return true, nil
	} else {
		return node.Properties.Get(ad.MSA.String()).Bool()
	}
}

func FetchAdminGroupIds(ctx context.Context, db graph.Database, groupExpansions impact.PathAggregator) (cardinality.Duplex[uint64], error) {
	adminIds := cardinality.NewBitmap64()

	return adminIds, db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		return tx.Nodes().Filter(
			query.Or(
				query.StringEndsWith(query.NodeProperty(common.ObjectID.String()), DomainAdminsGroupSIDSuffix),
				query.StringEndsWith(query.NodeProperty(common.ObjectID.String()), EnterpriseAdminsGroupSIDSuffix),
			),
		).FetchIDs(func(cursor graph.Cursor[graph.ID]) error {
			for id := range cursor.Chan() {
				adminIds.Add(id.Uint64())
				adminIds.Or(groupExpansions.Cardinality(id.Uint64()))
			}

			return nil
		})
	})
}

func GetDsHeuristicsCache(ctx context.Context, db graph.Database) (map[string]bool, bool, error) {
	var (
		dsHeuristicValues = make(map[string]bool)
		anyEnforced       bool
	)
	return dsHeuristicValues, anyEnforced, db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		if domains, err := ops.FetchNodes(tx.Nodes().Filter(query.Kind(query.Node(), ad.Domain))); err != nil {
			return err
		} else {
			for _, domain := range domains {
				if domainSid, err := domain.Properties.Get(ad.DomainSID.String()).String(); err != nil {
					continue
				} else if rawDsHeuristics, err := domain.Properties.Get(ad.DSHeuristics.String()).String(); err != nil {
					continue
				} else if len(rawDsHeuristics) < 29 {
					continue
				} else {
					enforcedChar := string(rawDsHeuristics[28])
					switch enforcedChar {
					case "0":
					case "2":
						dsHeuristicValues[domainSid] = false
					case "1":
						dsHeuristicValues[domainSid] = true
						anyEnforced = true
					default:
						continue
					}
				}
			}
		}

		return nil
	})
}
