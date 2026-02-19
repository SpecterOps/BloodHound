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
	"errors"
	"log/slog"

	"github.com/specterops/dawgs/algo"
	"github.com/specterops/dawgs/util/channels"

	"github.com/specterops/bloodhound/packages/go/analysis"
	"github.com/specterops/bloodhound/packages/go/analysis/ad/wellknown"
	"github.com/specterops/bloodhound/packages/go/bhlog/attr"
	"github.com/specterops/bloodhound/packages/go/bhlog/measure"
	"github.com/specterops/bloodhound/packages/go/graphschema/ad"
	"github.com/specterops/bloodhound/packages/go/graphschema/common"
	"github.com/specterops/dawgs/cardinality"
	"github.com/specterops/dawgs/graph"
	"github.com/specterops/dawgs/ops"
	"github.com/specterops/dawgs/query"
)

func PostOwnsAndWriteOwner(ctx context.Context, db graph.Database, localGroupData *LocalGroupData) (*analysis.AtomicPostProcessingStats, error) {
	defer measure.ContextMeasure(
		ctx,
		slog.LevelInfo,
		"Post-processing Owns and WriteOwner",
		attr.Namespace("analysis"),
		attr.Function("PostOwnsAndWriteOwner"),
		attr.Scope("process"),
	)()

	operation := analysis.NewPostRelationshipOperation(ctx, db, "PostOwnsAndWriteOwner")

	// Get the dSHeuristics values for all domains
	if dsHeuristicsCache, anyEnforced, err := GetDsHeuristicsCache(ctx, db); err != nil {
		slog.ErrorContext(ctx, "Failed fetching dsheuristics values for postownsandwriteowner", attr.Error(err))
		return nil, err
	} else if adminGroupIds, err := FetchAdminGroupIds(ctx, db, localGroupData.GroupMembershipCache); err != nil {
		// Get the admin group IDs
		slog.ErrorContext(ctx, "Failed fetching admin group ids values for postownsandwriteowner", attr.Error(err))
	} else {

		// Get all source nodes of Owns ACEs (i.e., owning principals) where the target node has no ACEs granting abusable explicit permissions to OWNER RIGHTS
		if err := operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
			if relationships, err := ops.FetchRelationships(tx.Relationships().Filterf(func() graph.Criteria {
				return query.And(
					query.Kind(query.Relationship(), ad.OwnsRaw),
					query.Kind(query.Start(), ad.Entity),
				)
			})); err != nil {
				slog.ErrorContext(ctx, "Failed to fetch OwnsRaw relationships for postownsandwriteowner", attr.Error(err))
			} else {
				for _, rel := range relationships {

					// Check if ANY domain enforces BlockOwnerImplicitRights (dSHeuristics[28] == 1)
					if anyEnforced {

						// Get the target node of the OwnsRaw relationship
						if targetNode, err := ops.FetchNode(tx, rel.EndID); err != nil {
							slog.ErrorContext(ctx, "Failed fetching OwnsRaw target node for postownsandwriteowner", attr.Error(err))
							continue

						} else if domainSid, err := targetNode.Properties.GetOrDefault(ad.DomainSID.String(), "").String(); err != nil {
							// Get the domain SID of the target node
							slog.ErrorContext(ctx, "Failed fetching domain SID for postownsandwriteowner", attr.Error(err))
							continue

						} else {
							enforced, ok := dsHeuristicsCache[domainSid]
							if !ok {
								enforced = false
							}
							// If THIS domain does NOT enforce BlockOwnerImplicitRights, add the Owns edge
							if !enforced {
								result := createPostRelFromRaw(rel, ad.Owns)
								channels.Submit(ctx, outC, result)
							} else if isComputerDerived, err := isTargetNodeComputerDerived(targetNode); err != nil {
								// If no abusable permissions are granted to OWNER RIGHTS, check if the target node is a computer or derived object (MSA or GMSA)
								continue
							} else if (isComputerDerived && adminGroupIds != nil && adminGroupIds.Contains(rel.StartID.Uint64())) || !isComputerDerived {
								// If the target node is a computer or derived object, add the Owns edge if the owning principal is a member of DA/EA (or is either group's SID)
								// If the target node is NOT a computer or derived object, add the Owns edge
								result := createPostRelFromRaw(rel, ad.Owns)
								channels.Submit(ctx, outC, result)
							}
						}
					} else {
						// If no domain enforces BlockOwnerImplicitRights (dSHeuristics[28] == 1) or we can't fetch the attribute, we can skip this analysis and just add the Owns relationship
						result := createPostRelFromRaw(rel, ad.Owns)
						channels.Submit(ctx, outC, result)
					}
				}
			}
			return nil
		}); err != nil {
			slog.ErrorContext(ctx, "Failed to process Owns relationships for postownsandwriteowner", attr.Error(err))
		}

		// Get all source nodes of WriteOwner ACEs where the target node has no ACEs granting explicit abusable permissions to OWNER RIGHTS
		if err := operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {

			if relationships, err := ops.FetchRelationships(tx.Relationships().Filterf(func() graph.Criteria {
				return query.And(
					query.Kind(query.Relationship(), ad.WriteOwnerRaw),
					query.Kind(query.Start(), ad.Entity),
				)
			})); err != nil {
				slog.ErrorContext(ctx, "Failed to fetch WriteOwnerRaw relationships for postownsandwriteowner", attr.Error(err))
			} else {
				for _, rel := range relationships {

					// Check if ANY domain enforces BlockOwnerImplicitRights (dSHeuristics[28] == 1)
					if anyEnforced {

						// Get the target node of the WriteOwnerRaw relationship
						if targetNode, err := ops.FetchNode(tx, rel.EndID); err != nil {
							slog.ErrorContext(ctx, "Failed fetching WriteOwnerRaw target node for postownsandwriteowner", attr.Error(err))
							continue

						} else if domainSid, err := targetNode.Properties.GetOrDefault(ad.DomainSID.String(), "").String(); err != nil {
							// Get the domain SID of the target node
							slog.ErrorContext(ctx, "Failed fetching domain SID for postownsandwriteowner", attr.Error(err))
							continue

						} else {
							enforced, ok := dsHeuristicsCache[domainSid]
							if !ok {
								enforced = false
							}

							// If THIS domain does NOT enforce BlockOwnerImplicitRights, add the WriteOwner edge
							if !enforced {
								result := createPostRelFromRaw(rel, ad.WriteOwner)
								channels.Submit(ctx, outC, result)
							} else if isComputerDerived, err := isTargetNodeComputerDerived(targetNode); err == nil {
								// If no abusable permissions are granted to OWNER RIGHTS, check if the target node is a computer or derived object (MSA or GMSA)
								if !isComputerDerived {
									// If the target node is NOT a computer or derived object, add the WriteOwner edge
									result := createPostRelFromRaw(rel, ad.WriteOwner)
									channels.Submit(ctx, outC, result)
								}
							}
						}
					} else {
						// If no domain enforces BlockOwnerImplicitRights (dSHeuristics[28] == 1) or we can't fetch the attribute, we can skip this analysis and just add the WriteOwner relationship
						result := createPostRelFromRaw(rel, ad.WriteOwner)
						channels.Submit(ctx, outC, result)
					}
				}
			}
			return nil
		}); err != nil {
			slog.ErrorContext(ctx, "Failed to process WriteOwner relationships for postownsandwriteowner", attr.Error(err))
		}
	}
	return &operation.Stats, operation.Done()
}

func createPostRelFromRaw(rel *graph.Relationship, kind graph.Kind) analysis.CreatePostRelationshipJob {
	isInherited, _ := rel.Properties.GetOrDefault(common.IsInherited.String(), false).Bool()
	inheritanceHash, _ := rel.Properties.GetOrDefault(ad.InheritanceHash.String(), "").String()

	return analysis.CreatePostRelationshipJob{
		FromID: rel.StartID,
		ToID:   rel.EndID,
		Kind:   kind,
		RelProperties: map[string]any{
			ad.IsACL.String():           true,
			common.IsInherited.String(): isInherited,
			ad.InheritanceHash.String(): inheritanceHash,
		},
	}
}

func isTargetNodeComputerDerived(node *graph.Node) (bool, error) {
	if node.Kinds.ContainsOneOf(ad.Computer) {
		return true, nil
	} else if isGmsa, err := node.Properties.Get(ad.GMSA.String()).Bool(); err != nil && !errors.Is(err, graph.ErrPropertyNotFound) {
		return false, err
	} else if isGmsa {
		return true, nil
	} else if isMsa, err := node.Properties.Get(ad.MSA.String()).Bool(); err != nil && !errors.Is(err, graph.ErrPropertyNotFound) {
		return false, err
	} else {
		return isMsa, nil
	}
}

func FetchAdminGroupIds(ctx context.Context, db graph.Database, groupExpansions *algo.ReachabilityCache) (cardinality.Duplex[uint64], error) {
	adminIds := cardinality.NewBitmap64()

	return adminIds, db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		return tx.Nodes().Filter(
			query.Or(
				query.StringEndsWith(query.NodeProperty(common.ObjectID.String()), wellknown.DomainAdminsGroupSIDSuffix.String()),
				query.StringEndsWith(query.NodeProperty(common.ObjectID.String()), wellknown.EnterpriseAdminsGroupSIDSuffix.String()),
			),
		).FetchIDs(func(cursor graph.Cursor[graph.ID]) error {
			for id := range cursor.Chan() {
				adminIds.Add(id.Uint64())
				groupExpansions.OrReach(id.Uint64(), graph.DirectionInbound, adminIds)
			}

			return cursor.Error()
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
					switch enforcedChar := string(rawDsHeuristics[28]); enforcedChar {
					case "0", "2":
						dsHeuristicValues[domainSid] = false
					case "1":
						dsHeuristicValues[domainSid] = true
						anyEnforced = true
					}
				}
			}
		}

		return nil
	})
}
