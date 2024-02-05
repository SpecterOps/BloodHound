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

	"github.com/RoaringBitmap/roaring/roaring64"
	"github.com/specterops/bloodhound/analysis"
	"github.com/specterops/bloodhound/analysis/impact"
	"github.com/specterops/bloodhound/dawgs/cardinality"
	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/dawgs/ops"
	"github.com/specterops/bloodhound/dawgs/query"
	"github.com/specterops/bloodhound/dawgs/util/channels"
	"github.com/specterops/bloodhound/graphschema/ad"
	"github.com/specterops/bloodhound/graphschema/common"
	"github.com/specterops/bloodhound/log"
)

func PostProcessedRelationships() []graph.Kind {
	return []graph.Kind{
		ad.DCSync,
		ad.SyncLAPSPassword,
		ad.CanRDP,
		ad.AdminTo,
		ad.CanPSRemote,
		ad.ExecuteDCOM,
		ad.TrustedForNTAuth,
		ad.IssuedSignedBy,
		ad.EnterpriseCAFor,
		ad.CanAbuseWeakCertBinding,
		ad.CanAbuseUPNCertMapping,
		ad.GoldenCert,
		ad.CanAbuseUPNCertMapping,
		ad.CanAbuseWeakCertBinding,
		ad.ADCSESC1,
		ad.ADCSESC3,
		ad.ADCSESC4,
		ad.ADCSESC5,
		ad.ADCSESC6a,
		ad.ADCSESC6b,
		ad.ADCSESC7,
		ad.ADCSESC10a,
		ad.ADCSESC10b,
		ad.ADCSESC9a,
		ad.EnrollOnBehalfOf,
	}
}

func PostSyncLAPSPassword(ctx context.Context, db graph.Database) (*analysis.AtomicPostProcessingStats, error) {
	if domainNodes, err := fetchCollectedDomainNodes(ctx, db); err != nil {
		return &analysis.AtomicPostProcessingStats{}, err
	} else {
		operation := analysis.NewPostRelationshipOperation(ctx, db, "SyncLAPSPassword Post Processing")
		for _, domain := range domainNodes {
			innerDomain := domain
			operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
				if lapsSyncers, err := analysis.GetLAPSSyncers(tx, innerDomain); err != nil {
					return err
				} else if len(lapsSyncers) == 0 {
					return nil
				} else if computers, err := getLAPSComputersForDomain(tx, innerDomain); err != nil {
					return err
				} else {
					for _, computer := range computers {
						for _, lapsSyncer := range lapsSyncers {
							nextJob := analysis.CreatePostRelationshipJob{
								FromID: lapsSyncer.ID,
								ToID:   computer,
								Kind:   ad.SyncLAPSPassword,
							}

							if !channels.Submit(ctx, outC, nextJob) {
								return nil
							}
						}
					}

					return nil
				}
			})
		}

		return &operation.Stats, operation.Done()
	}
}

func PostDCSync(ctx context.Context, db graph.Database) (*analysis.AtomicPostProcessingStats, error) {
	if domainNodes, err := fetchCollectedDomainNodes(ctx, db); err != nil {
		return &analysis.AtomicPostProcessingStats{}, err
	} else {
		operation := analysis.NewPostRelationshipOperation(ctx, db, "DCSync Post Processing")

		for _, domain := range domainNodes {
			innerDomain := domain
			operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
				if dcSyncers, err := analysis.GetDCSyncers(tx, innerDomain, true); err != nil {
					return err
				} else if len(dcSyncers) == 0 {
					return nil
				} else {
					for _, dcSyncer := range dcSyncers {
						nextJob := analysis.CreatePostRelationshipJob{
							FromID: dcSyncer.ID,
							ToID:   innerDomain.ID,
							Kind:   ad.DCSync,
						}

						if !channels.Submit(ctx, outC, nextJob) {
							return nil
						}
					}

					return nil
				}
			})
		}

		return &operation.Stats, operation.Done()
	}
}

func FetchComputers(ctx context.Context, db graph.Database) (*roaring64.Bitmap, error) {
	computerNodeIds := roaring64.NewBitmap()

	return computerNodeIds, db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		return tx.Nodes().Filterf(func() graph.Criteria {
			return query.Kind(query.Node(), ad.Computer)
		}).FetchIDs(func(cursor graph.Cursor[graph.ID]) error {
			for id := range cursor.Chan() {
				computerNodeIds.Add(id.Uint64())
			}

			return nil
		})
	})
}

func FetchNodesByKind(ctx context.Context, db graph.Database, kinds ...graph.Kind) ([]*graph.Node, error) {
	var nodes []*graph.Node
	return nodes, db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		var err error
		if nodes, err = ops.FetchNodes(tx.Nodes().Filterf(func() graph.Criteria {
			return query.And(
				query.KindIn(query.Node(), kinds...),
			)
		})); err != nil {
			return err
		} else {
			return nil
		}
	})
}

func fetchCollectedDomainNodes(ctx context.Context, db graph.Database) ([]*graph.Node, error) {
	var nodes []*graph.Node
	return nodes, db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		var err error
		if nodes, err = ops.FetchNodes(tx.Nodes().Filterf(func() graph.Criteria {
			return query.And(
				query.Kind(query.Node(), ad.Domain),
				query.Equals(query.NodeProperty(common.Collected.String()), true),
			)
		})); err != nil {
			return err
		} else {
			return nil
		}
	})
}

func getLAPSComputersForDomain(tx graph.Transaction, domain *graph.Node) ([]graph.ID, error) {
	if domainSid, err := domain.Properties.Get(ad.DomainSID.String()).String(); err != nil {
		return nil, err
	} else {
		return ops.FetchNodeIDs(tx.Nodes().Filterf(func() graph.Criteria {
			return query.And(
				query.Kind(query.Node(), ad.Computer),
				query.Equals(
					query.Property(query.Node(), ad.HasLAPS.String()), true),
				query.Equals(query.Property(query.Node(), ad.DomainSID.String()), domainSid),
			)
		}))
	}
}

func PostLocalGroups(ctx context.Context, db graph.Database, localGroupExpansions impact.PathAggregator, enforceURA bool) (*analysis.AtomicPostProcessingStats, error) {
	var (
		adminGroupSuffix    = "-544"
		psRemoteGroupSuffix = "-580"
		dcomGroupSuffix     = "-562"
	)

	if computers, err := FetchComputers(ctx, db); err != nil {
		return &analysis.AtomicPostProcessingStats{}, err
	} else {
		var (
			threadSafeLocalGroupExpansions = impact.NewThreadSafeAggregator(localGroupExpansions)
			operation                      = analysis.NewPostRelationshipOperation(ctx, db, "LocalGroup Post Processing")
		)

		for idx, computer := range computers.ToArray() {
			computerID := graph.ID(computer)

			if idx > 0 && idx%10000 == 0 {
				log.Infof("Post processed %d active directory computers", idx)
			}

			if err := operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
				if entities, err := FetchLocalGroupBitmapForComputer(tx, computerID, dcomGroupSuffix); err != nil {
					return err
				} else {
					for _, admin := range entities.Slice() {
						nextJob := analysis.CreatePostRelationshipJob{
							FromID: graph.ID(admin),
							ToID:   computerID,
							Kind:   ad.ExecuteDCOM,
						}

						if !channels.Submit(ctx, outC, nextJob) {
							return nil
						}
					}

					return nil
				}
			}); err != nil {
				return &analysis.AtomicPostProcessingStats{}, fmt.Errorf("failed submitting reader for operation involving computer %d: %w", computerID, err)
			}

			if err := operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
				if entities, err := FetchLocalGroupBitmapForComputer(tx, computerID, psRemoteGroupSuffix); err != nil {
					return err
				} else {
					for _, admin := range entities.Slice() {
						nextJob := analysis.CreatePostRelationshipJob{
							FromID: graph.ID(admin),
							ToID:   computerID,
							Kind:   ad.CanPSRemote,
						}

						if !channels.Submit(ctx, outC, nextJob) {
							return nil
						}
					}

					return nil
				}
			}); err != nil {
				return &analysis.AtomicPostProcessingStats{}, fmt.Errorf("failed submitting reader for operation involving computer %d: %w", computerID, err)
			}

			if err := operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
				if entities, err := FetchLocalGroupBitmapForComputer(tx, computerID, adminGroupSuffix); err != nil {
					return err
				} else {
					for _, admin := range entities.Slice() {
						nextJob := analysis.CreatePostRelationshipJob{
							FromID: graph.ID(admin),
							ToID:   computerID,
							Kind:   ad.AdminTo,
						}

						if !channels.Submit(ctx, outC, nextJob) {
							return nil
						}
					}

					return nil
				}
			}); err != nil {
				return &analysis.AtomicPostProcessingStats{}, fmt.Errorf("failed submitting reader for operation involving computer %d: %w", computerID, err)
			}

			if err := operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
				if entities, err := FetchRDPEntityBitmapForComputer(tx, computerID, threadSafeLocalGroupExpansions, enforceURA); err != nil {
					return err
				} else {
					for _, rdp := range entities.Slice() {
						nextJob := analysis.CreatePostRelationshipJob{
							FromID: graph.ID(rdp),
							ToID:   computerID,
							Kind:   ad.CanRDP,
						}

						if !channels.Submit(ctx, outC, nextJob) {
							return nil
						}
					}
				}

				return nil
			}); err != nil {
				return &analysis.AtomicPostProcessingStats{}, fmt.Errorf("failed submitting reader for operation involving computer %d: %w", computerID, err)
			}
		}

		log.Infof("Finished post-processing %d active directory computers", computers.GetCardinality())
		return &operation.Stats, operation.Done()
	}
}

func ExpandLocalGroupMembership(tx graph.Transaction, candidates graph.NodeSet) (graph.NodeSet, error) {
	if paths, err := ExpandLocalGroupMembershipPaths(tx, candidates); err != nil {
		return nil, err
	} else {
		return paths.AllNodes(), nil
	}
}

func ExpandLocalGroupMembershipPaths(tx graph.Transaction, candidates graph.NodeSet) (graph.PathSet, error) {
	groupMemberPaths := graph.NewPathSet()

	for _, candidate := range candidates {
		if candidate.Kinds.ContainsOneOf(ad.Group) {
			if membershipPaths, err := ops.TraversePaths(tx, ops.TraversalPlan{
				Root:      candidate,
				Direction: graph.DirectionInbound,
				BranchQuery: func() graph.Criteria {
					return query.KindIn(query.Relationship(), ad.MemberOf, ad.MemberOfLocalGroup)
				},
			}); err != nil {
				return nil, err
			} else {
				groupMemberPaths.AddPathSet(membershipPaths)
			}
		}
	}

	return groupMemberPaths, nil
}

func Uint64ToIDSlice(uint64IDs []uint64) []graph.ID {
	ids := make([]graph.ID, len(uint64IDs))
	for idx := 0; idx < len(uint64IDs); idx++ {
		ids[idx] = graph.ID(uint64IDs[idx])
	}

	return ids
}

func ExpandGroupMembershipIDBitmap(tx graph.Transaction, group *graph.Node) (*roaring64.Bitmap, error) {
	groupMembers := roaring64.NewBitmap()

	if membershipPaths, err := ops.TraversePaths(tx, ops.TraversalPlan{
		Root:      group,
		Direction: graph.DirectionInbound,
		BranchQuery: func() graph.Criteria {
			return query.Kind(query.Relationship(), ad.MemberOf)
		},
	}); err != nil {
		return nil, err
	} else {
		for _, node := range membershipPaths.AllNodes() {
			groupMembers.Add(node.ID.Uint64())
		}
	}

	return groupMembers, nil
}

func FetchComputerLocalGroupBySIDSuffix(tx graph.Transaction, computer graph.ID, groupSuffix string) (*graph.Node, error) {
	if rel, err := tx.Relationships().Filter(query.And(
		query.StringEndsWith(query.StartProperty(common.ObjectID.String()), groupSuffix),
		query.Kind(query.Relationship(), ad.LocalToComputer),
		query.InIDs(query.EndID(), computer),
	)).First(); err != nil {
		return nil, err
	} else {
		return ops.FetchNode(tx, rel.StartID)
	}
}

func FetchLocalGroupMembership(tx graph.Transaction, computer graph.ID, groupSuffix string) (graph.NodeSet, error) {
	if localGroup, err := FetchComputerLocalGroupBySIDSuffix(tx, computer, groupSuffix); err != nil {
		return nil, err
	} else {
		return ops.FetchStartNodes(tx.Relationships().Filter(query.And(
			query.KindIn(query.Start(), ad.User, ad.Group, ad.Computer),
			query.Kind(query.Relationship(), ad.MemberOfLocalGroup),
			query.InIDs(query.EndID(), localGroup.ID),
		)))
	}
}

func FetchRemoteInteractiveLogonPrivilegedEntities(tx graph.Transaction, computerId graph.ID) (graph.NodeSet, error) {
	return ops.FetchStartNodes(tx.Relationships().Filterf(func() graph.Criteria {
		return query.And(
			query.Kind(query.Relationship(), ad.RemoteInteractiveLogonPrivilege),
			query.Equals(query.EndID(), computerId),
		)
	}))
}

func HasRemoteInteractiveLogonPrivilege(tx graph.Transaction, groupId, computerId graph.ID) bool {
	if _, err := tx.Relationships().Filterf(func() graph.Criteria {
		return query.And(
			query.Equals(query.StartID(), groupId),
			query.Equals(query.EndID(), computerId),
			query.Kind(query.Relationship(), ad.RemoteInteractiveLogonPrivilege),
		)
	}).First(); err != nil {
		return false
	}

	return true
}

func FetchLocalGroupBitmapForComputer(tx graph.Transaction, computer graph.ID, suffix string) (cardinality.Duplex[uint32], error) {
	if members, err := FetchLocalGroupMembership(tx, computer, suffix); err != nil {
		if graph.IsErrNotFound(err) {
			return cardinality.NewBitmap32(), nil
		}

		return nil, err
	} else {
		return cardinality.NodeSetToDuplex(members), nil
	}
}

func ExpandAllRDPLocalGroups(ctx context.Context, db graph.Database) (impact.PathAggregator, error) {
	log.Infof("Expanding all AD group and local group memberships")

	return ResolveAllGroupMemberships(ctx, db, query.Not(
		query.Or(
			query.StringEndsWith(query.StartProperty(common.ObjectID.String()), AdminGroupSuffix),
			query.StringEndsWith(query.EndProperty(common.ObjectID.String()), AdminGroupSuffix),
		),
	))
}

func FetchRDPEntityBitmapForComputer(tx graph.Transaction, computer graph.ID, localGroupExpansions impact.PathAggregator, enforceURA bool) (cardinality.Duplex[uint32], error) {
	if rdpLocalGroup, err := FetchComputerLocalGroupBySIDSuffix(tx, computer, RDPGroupSuffix); err != nil {
		if graph.IsErrNotFound(err) {
			return cardinality.NewBitmap32(), nil
		}

		return nil, err
	} else if enforceURA || ComputerHasURACollection(tx, computer) {
		return ProcessRDPWithUra(tx, rdpLocalGroup, computer, localGroupExpansions)
	} else if bitmap, err := FetchLocalGroupBitmapForComputer(tx, computer, RDPGroupSuffix); err != nil {
		return nil, err
	} else {
		return bitmap, nil
	}
}

func ComputerHasURACollection(tx graph.Transaction, computerID graph.ID) bool {
	if computer, err := tx.Nodes().Filterf(func() graph.Criteria {
		return query.Equals(query.NodeID(), computerID)
	}).First(); err != nil {
		return false
	} else {
		if ura, err := computer.Properties.Get(ad.HasURA.String()).Bool(); err != nil {
			return false
		} else {
			return ura
		}
	}
}

func ProcessRDPWithUra(tx graph.Transaction, rdpLocalGroup *graph.Node, computer graph.ID, localGroupExpansions impact.PathAggregator) (cardinality.Duplex[uint32], error) {
	rdpLocalGroupMembers := localGroupExpansions.Cardinality(rdpLocalGroup.ID.Uint32()).(cardinality.Duplex[uint32])
	// Shortcut opportunity: see if the RDP group has RIL privilege. If it does, get the first degree members and return those ids, since everything in RDP group has CanRDP privs. No reason to look any further
	if HasRemoteInteractiveLogonPrivilege(tx, rdpLocalGroup.ID, computer) {
		firstDegreeMembers := cardinality.NewBitmap32()

		return firstDegreeMembers, tx.Relationships().Filter(
			query.And(
				query.Kind(query.Relationship(), ad.MemberOfLocalGroup),
				query.KindIn(query.Start(), ad.Group, ad.User),
				query.Equals(query.EndID(), rdpLocalGroup.ID),
			),
		).FetchTriples(func(cursor graph.Cursor[graph.RelationshipTripleResult]) error {
			for result := range cursor.Chan() {
				firstDegreeMembers.Add(result.StartID.Uint32())
			}
			return cursor.Error()
		})
	} else if baseRilEntities, err := FetchRemoteInteractiveLogonPrivilegedEntities(tx, computer); err != nil {
		return nil, err
	} else {
		var (
			rdpEntities      = cardinality.NewBitmap32()
			secondaryTargets = cardinality.NewBitmap32()
		)

		// Attempt 2: look at each RIL entity directly and see if it has membership to the RDP group. If not, and it's a group, expand its membership for further processing
		for _, entity := range baseRilEntities {
			if rdpLocalGroupMembers.Contains(entity.ID.Uint32()) {
				// If we have membership to the RDP group, then this is a valid CanRDP entity
				rdpEntities.Add(entity.ID.Uint32())
			} else if entity.Kinds.ContainsOneOf(ad.Group, ad.LocalGroup) {
				secondaryTargets.Or(localGroupExpansions.Cardinality(entity.ID.Uint32()).(cardinality.Duplex[uint32]))
			}
		}

		// Attempt 3: Look at each member of expanded groups and see if they have the correct permissions
		for _, entity := range secondaryTargets.Slice() {
			// If we have membership to the RDP group then this is a valid CanRDP entity
			if rdpLocalGroupMembers.Contains(entity) {
				rdpEntities.Add(entity)
			}
		}

		return rdpEntities, nil
	}
}

func ExpandAllEnrollers(ctx context.Context, db graph.Database) (impact.PathAggregator, error) {
	log.Infof("Expanding all cert template and enterprise ca enrollers")

	return ResolveAllGroupMemberships(ctx, db, query.Not(
		query.Or(
			query.StringEndsWith(query.StartProperty(common.ObjectID.String()), AdminGroupSuffix),
			query.StringEndsWith(query.EndProperty(common.ObjectID.String()), AdminGroupSuffix),
		),
	))
}
