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
	"fmt"
	"log/slog"

	"github.com/specterops/dawgs/algo"

	"github.com/specterops/bloodhound/packages/go/analysis/ad/wellknown"
	"github.com/specterops/bloodhound/packages/go/analysis/post"
	"github.com/specterops/bloodhound/packages/go/bhlog/attr"
	"github.com/specterops/bloodhound/packages/go/bhlog/measure"
	"github.com/specterops/bloodhound/packages/go/graphschema/ad"
	"github.com/specterops/bloodhound/packages/go/graphschema/common"
	"github.com/specterops/dawgs/cardinality"
	"github.com/specterops/dawgs/graph"
	"github.com/specterops/dawgs/ops"
	"github.com/specterops/dawgs/query"
)

var ownsWriteOwnerPostProcessedEdges = graph.Kinds{
	ad.Owns,
	ad.WriteOwner,
}

func PostOwnsAndWriteOwner(ctx context.Context, db graph.Database, localGroupData *LocalGroupData) (*post.AtomicPostProcessingStats, error) {
	defer measure.ContextLogAndMeasure(
		ctx,
		slog.LevelInfo,
		"Post-processing Owns and WriteOwner",
		attr.Namespace("analysis"),
		attr.Function("PostOwnsAndWriteOwner"),
		attr.Scope("process"),
	)()

	// Clear old post-processed edges that will not have a `firstseen` property
	if err := post.MigrationForDCAPostProcessedEdges(ctx, db, ownsWriteOwnerPostProcessedEdges); err != nil {
		return &post.AtomicPostProcessingStats{}, err
	}

	// Pull a subgraph to compare against for tracking changes
	ownsWriteOwnerTracker, err := post.FetchTracker(ctx, db, ownsWriteOwnerPostProcessedEdges)
	if err != nil {
		return &post.AtomicPostProcessingStats{}, err
	}

	// Get the dSHeuristics values for all domains
	dsHeuristicsCache, anyEnforced, err := GetDsHeuristicsCache(ctx, db)
	if err != nil {
		slog.ErrorContext(ctx, "Failed fetching dsheuristics values for postownsandwriteowner", attr.Error(err))
		return &post.AtomicPostProcessingStats{}, err
	}

	adminGroupIds, err := FetchAdminGroupIds(ctx, db, localGroupData.GroupMembershipCache)
	if err != nil {
		// While it's ideal if this succeeds, the risk of this call failing will be a false negative case of missing an `Owns` edge sourced
		// from an Enterprise or Domain Admin, who will already have control in another manner anyways.
		slog.WarnContext(ctx, "Failed fetching admin group ids values for postownsandwriteowner", attr.Error(err))
	}

	sink := post.NewFilteredRelationshipSink(ctx, "PostOwnsAndWriteOwner", db, ownsWriteOwnerTracker)
	defer sink.Done()

	if err := postOwnsEdges(ctx, db, sink, dsHeuristicsCache, anyEnforced, adminGroupIds); err != nil {
		return sink.Stats(), err
	}

	if err := postWriteOwnerEdges(ctx, db, sink, dsHeuristicsCache, anyEnforced); err != nil {
		return sink.Stats(), err
	}

	return sink.Stats(), nil
}

func postOwnsEdges(ctx context.Context, db graph.Database, sink *post.FilteredRelationshipSink, dsHeuristicsCache map[string]bool, anyEnforced bool, adminGroupIds cardinality.Duplex[uint64]) error {
	return db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		if relationships, err := ops.FetchRelationships(tx.Relationships().Filterf(func() graph.Criteria {
			return query.And(
				query.Kind(query.Relationship(), ad.OwnsRaw),
				query.Kind(query.Start(), ad.Entity),
			)
		})); err != nil {
			slog.ErrorContext(ctx, "Failed to fetch OwnsRaw relationships for postownsandwriteowner", attr.Error(err))
			return err
		} else {
			// When anyEnforced, batch-fetch all target nodes upfront instead of one query per relationship.
			var targetNodes graph.NodeSet
			if anyEnforced {
				endIDs := collectUniqueEndIDs(relationships)
				if fetched, err := ops.FetchNodeSet(tx.Nodes().Filterf(func() graph.Criteria {
					return query.InIDs(query.NodeID(), endIDs...)
				})); err != nil {
					slog.ErrorContext(ctx, "Failed bulk-fetching OwnsRaw target nodes for postownsandwriteowner", attr.Error(err))
					return err
				} else {
					targetNodes = fetched
				}
			}

			for _, rel := range relationships {

				// Check if ANY domain enforces BlockOwnerImplicitRights (dSHeuristics[28] == 1)
				if anyEnforced {

					// Look up the target node from the pre-fetched set
					targetNode, ok := targetNodes[rel.EndID]
					if !ok {
						slog.ErrorContext(ctx, "Failed fetching OwnsRaw target node for postownsandwriteowner (not in bulk result)",
							slog.Uint64("end_id", uint64(rel.EndID)))
						continue
					}

					domainSid, err := targetNode.Properties.GetOrDefault(ad.DomainSID.String(), "").String()
					if err != nil {
						// Get the domain SID of the target node
						slog.ErrorContext(ctx, "Failed fetching domain SID for postownsandwriteowner", attr.Error(err))
						continue
					}

					enforced, ok := dsHeuristicsCache[domainSid]
					if !ok {
						enforced = false
					}
					// If THIS domain does NOT enforce BlockOwnerImplicitRights, add the Owns edge
					if !enforced {
						result := createPostRelFromRaw(rel, ad.Owns)
						if !sink.Submit(ctx, result) {
							return fmt.Errorf("unable to submit to channel in postOwnsEdges")
						}
					} else if isComputerDerived, err := isTargetNodeComputerDerived(targetNode); err != nil {
						// If no abusable permissions are granted to OWNER RIGHTS, check if the target node is a computer or derived object (MSA or GMSA)
						continue
					} else if (isComputerDerived && adminGroupIds != nil && adminGroupIds.Contains(rel.StartID.Uint64())) || !isComputerDerived {
						// If the target node is a computer or derived object, add the Owns edge if the owning principal is a member of DA/EA (or is either group's SID)
						// If the target node is NOT a computer or derived object, add the Owns edge
						result := createPostRelFromRaw(rel, ad.Owns)
						if !sink.Submit(ctx, result) {
							return fmt.Errorf("unable to submit to channel in postOwnsEdges")
						}
					}
				} else {
					// If no domain enforces BlockOwnerImplicitRights (dSHeuristics[28] == 1) or we can't fetch the attribute, we can skip this analysis and just add the Owns relationship
					result := createPostRelFromRaw(rel, ad.Owns)
					if !sink.Submit(ctx, result) {
						return fmt.Errorf("unable to submit to channel in postOwnsEdges")
					}
				}
			}
		}
		return nil
	})
}

func postWriteOwnerEdges(ctx context.Context, db graph.Database, sink *post.FilteredRelationshipSink, dsHeuristicsCache map[string]bool, anyEnforced bool) error {
	return db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		if relationships, err := ops.FetchRelationships(tx.Relationships().Filterf(func() graph.Criteria {
			return query.And(
				query.Kind(query.Relationship(), ad.WriteOwnerRaw),
				query.Kind(query.Start(), ad.Entity),
			)
		})); err != nil {
			slog.ErrorContext(ctx, "Failed to fetch WriteOwnerRaw relationships for postownsandwriteowner", attr.Error(err))
			return err
		} else {
			// When anyEnforced, batch-fetch all target nodes upfront instead of one query per relationship.
			var targetNodes graph.NodeSet
			if anyEnforced {
				endIDs := collectUniqueEndIDs(relationships)
				if fetched, err := ops.FetchNodeSet(tx.Nodes().Filterf(func() graph.Criteria {
					return query.InIDs(query.NodeID(), endIDs...)
				})); err != nil {
					slog.ErrorContext(ctx, "Failed bulk-fetching WriteOwnerRaw target nodes for postownsandwriteowner", attr.Error(err))
					return err
				} else {
					targetNodes = fetched
				}
			}

			for _, rel := range relationships {

				// Check if ANY domain enforces BlockOwnerImplicitRights (dSHeuristics[28] == 1)
				if anyEnforced {

					// Look up the target node from the pre-fetched set
					targetNode, ok := targetNodes[rel.EndID]
					if !ok {
						slog.ErrorContext(ctx, "Failed fetching WriteOwnerRaw target node for postownsandwriteowner (not in bulk result)",
							slog.Uint64("end_id", uint64(rel.EndID)))
						continue
					}

					domainSid, err := targetNode.Properties.GetOrDefault(ad.DomainSID.String(), "").String()
					if err != nil {
						// Get the domain SID of the target node
						slog.ErrorContext(ctx, "Failed fetching domain SID for postownsandwriteowner", attr.Error(err))
						continue
					}

					enforced, ok := dsHeuristicsCache[domainSid]
					if !ok {
						enforced = false
					}

					// If THIS domain does NOT enforce BlockOwnerImplicitRights, add the WriteOwner edge
					if !enforced {
						result := createPostRelFromRaw(rel, ad.WriteOwner)
						if !sink.Submit(ctx, result) {
							return fmt.Errorf("unable to submit to channel in postWriteOwnerEdges")
						}
					} else if isComputerDerived, err := isTargetNodeComputerDerived(targetNode); err == nil {
						// If no abusable permissions are granted to OWNER RIGHTS, check if the target node is a computer or derived object (MSA or GMSA)
						if !isComputerDerived {
							// If the target node is NOT a computer or derived object, add the WriteOwner edge
							result := createPostRelFromRaw(rel, ad.WriteOwner)
							if !sink.Submit(ctx, result) {
								return fmt.Errorf("unable to submit to channel in postWriteOwnerEdges")
							}
						}
					}
				} else {
					// If no domain enforces BlockOwnerImplicitRights (dSHeuristics[28] == 1) or we can't fetch the attribute, we can skip this analysis and just add the WriteOwner relationship
					result := createPostRelFromRaw(rel, ad.WriteOwner)
					if !sink.Submit(ctx, result) {
						return fmt.Errorf("unable to submit to channel in postWriteOwnerEdges")
					}
				}
			}
		}
		return nil
	})
}

// collectUniqueEndIDs extracts all unique EndIDs from a slice of relationships for bulk node fetching.
func collectUniqueEndIDs(relationships []*graph.Relationship) []graph.ID {
	seen := make(map[graph.ID]struct{}, len(relationships))
	ids := make([]graph.ID, 0, len(relationships))
	for _, rel := range relationships {
		if _, ok := seen[rel.EndID]; !ok {
			seen[rel.EndID] = struct{}{}
			ids = append(ids, rel.EndID)
		}
	}
	return ids
}

func createPostRelFromRaw(rel *graph.Relationship, kind graph.Kind) post.EnsureRelationshipJob {
	isInherited, _ := rel.Properties.GetOrDefault(common.IsInherited.String(), false).Bool()
	inheritanceHash, _ := rel.Properties.GetOrDefault(ad.InheritanceHash.String(), "").String()

	return post.EnsureRelationshipJob{
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
