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

	"github.com/specterops/bloodhound/packages/go/analysis/ad/wellknown"
	"github.com/specterops/bloodhound/packages/go/analysis/post"
	"github.com/specterops/bloodhound/packages/go/bhlog/attr"
	"github.com/specterops/bloodhound/packages/go/bhlog/measure"
	"github.com/specterops/bloodhound/packages/go/graphschema/ad"
	"github.com/specterops/bloodhound/packages/go/graphschema/azure"
	"github.com/specterops/bloodhound/packages/go/graphschema/common"
	"github.com/specterops/dawgs/algo"
	"github.com/specterops/dawgs/cardinality"
	"github.com/specterops/dawgs/container"
	"github.com/specterops/dawgs/graph"
	"github.com/specterops/dawgs/ops"
	"github.com/specterops/dawgs/query"
	"github.com/specterops/dawgs/util/channels"
)

func PostSyncLAPSPassword(ctx context.Context, db graph.Database, localGroupData *LocalGroupData) (*post.AtomicPostProcessingStats, error) {
	defer measure.ContextLogAndMeasure(
		ctx,
		slog.LevelInfo,
		"Post-processing SyncLAPSPassword",
		attr.Namespace("analysis"),
		attr.Function("PostSyncLAPSPassword"),
		attr.Scope("process"),
	)()

	if domainNodes, err := fetchCollectedDomainNodes(ctx, db); err != nil {
		return &post.AtomicPostProcessingStats{}, err
	} else {
		operation := post.NewPostRelationshipOperation(ctx, db, "SyncLAPSPassword Post Processing")
		for _, domain := range domainNodes {
			innerDomain := domain
			operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- post.EnsureRelationshipJob) error {
				if lapsSyncers, err := getLAPSSyncers(tx, innerDomain, localGroupData); err != nil {
					return err
				} else if lapsSyncers.Cardinality() == 0 {
					return nil
				} else if computers, err := getLAPSComputersForDomain(tx, innerDomain); err != nil {
					return err
				} else {
					for _, computer := range computers {
						lapsSyncers.Each(func(value uint64) bool {
							channels.Submit(ctx, outC, post.EnsureRelationshipJob{
								FromID: graph.ID(value),
								ToID:   computer,
								Kind:   ad.SyncLAPSPassword,
							})
							return true
						})
					}

					return nil
				}
			})
		}

		return &operation.Stats, operation.Done()
	}
}

func PostDCSync(ctx context.Context, db graph.Database, localGroupData *LocalGroupData) (*post.AtomicPostProcessingStats, error) {
	defer measure.ContextLogAndMeasure(
		ctx,
		slog.LevelInfo,
		"Post-processing DCSync",
		attr.Namespace("analysis"),
		attr.Function("PostDCSync"),
		attr.Scope("process"),
	)()

	if domainNodes, err := fetchCollectedDomainNodes(ctx, db); err != nil {
		return &post.AtomicPostProcessingStats{}, err
	} else {
		operation := post.NewPostRelationshipOperation(ctx, db, "DCSync Post Processing")

		for _, domain := range domainNodes {
			innerDomain := domain
			operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- post.EnsureRelationshipJob) error {
				if dcSyncers, err := getDCSyncers(tx, innerDomain, localGroupData); err != nil {
					return err
				} else if dcSyncers.Cardinality() == 0 {
					return nil
				} else {
					dcSyncers.Each(func(value uint64) bool {
						channels.Submit(ctx, outC, post.EnsureRelationshipJob{
							FromID: graph.ID(value),
							ToID:   innerDomain.ID,
							Kind:   ad.DCSync,
						})
						return true
					})

					return nil
				}
			})
		}

		return &operation.Stats, operation.Done()
	}
}

func PostProtectAdminGroups(ctx context.Context, db graph.Database) (*post.AtomicPostProcessingStats, error) {
	defer measure.ContextLogAndMeasure(
		ctx,
		slog.LevelInfo,
		"Post-processing protected admin groups",
		attr.Namespace("analysis"),
		attr.Function("PostProtectAdminGroups"),
		attr.Scope("process"),
	)()

	domainNodes, err := fetchCollectedDomainNodes(ctx, db)
	if err != nil {
		return &post.AtomicPostProcessingStats{}, err
	}

	operation := post.NewPostRelationshipOperation(ctx, db, "ProtectAdminGroups Post Processing")

	for _, domain := range domainNodes {

		operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- post.EnsureRelationshipJob) error {
			if adminSDHolderIDs, err := getAdminSDHolder(tx, domain); graph.IsErrNotFound(err) {
				// No AdminSDHolder IDs found for this domain
				return nil
			} else if err != nil {
				return err
			} else if len(adminSDHolderIDs) == 0 {
				// No AdminSDHolder IDs found for this domain
				return nil
			} else if protectedObjectIDs, err := getAdminSDHolderProtected(tx, domain); err != nil {
				return err
			} else {
				fromID := adminSDHolderIDs[0] // AdminSDHolder should be unique per domain
				for _, toID := range protectedObjectIDs {
					channels.Submit(ctx, outC, post.EnsureRelationshipJob{
						FromID: fromID,
						ToID:   toID,
						Kind:   ad.ProtectAdminGroups,
					})
				}
				return nil
			}
		})
	}

	return &operation.Stats, operation.Done()
}

func PostHasTrustKeys(ctx context.Context, db graph.Database) (*post.AtomicPostProcessingStats, error) {
	defer measure.ContextLogAndMeasure(
		ctx,
		slog.LevelInfo,
		"Post-processing HasTrustKeys",
		attr.Namespace("analysis"),
		attr.Function("PostHasTrustKeys"),
		attr.Scope("process"),
	)()

	if domainNodes, err := fetchCollectedDomainNodes(ctx, db); err != nil {
		return &post.AtomicPostProcessingStats{}, err
	} else {
		operation := post.NewPostRelationshipOperation(ctx, db, "HasTrustKeys Post Processing")
		if err := operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- post.EnsureRelationshipJob) error {
			for _, domain := range domainNodes {
				if netbios, err := domain.Properties.Get(ad.NetBIOS.String()).String(); err != nil {
					// The property is new and may therefore not exist
					slog.DebugContext(
						ctx,
						"Skipping domain. Missing NetBIOS property",
						slog.Uint64("domain_id", uint64(domain.ID)),
					)
					continue
				} else if trustingDomains, err := getDirectOutboundTrustDomains(tx, domain); err != nil {
					slog.ErrorContext(
						ctx,
						"Error getting outbound trust edges from domain",
						slog.Uint64("domain_id", uint64(domain.ID)),
						attr.Error(err),
					)
					continue
				} else {
					for _, trustingDomain := range trustingDomains {
						if trustingDomainSid, err := trustingDomain.Properties.Get(ad.DomainSID.String()).String(); err != nil {
							// DomainSID is only created after we have performed collection of the domain
							slog.DebugContext(
								ctx,
								"Skipping trusting domain. Missing DomainSID property",
								slog.Uint64("trusting_domain_id", uint64(trustingDomain.ID)),
							)
							continue
						} else if trustAccount, err := getTrustAccount(tx, trustingDomainSid, netbios); err != nil {
							// The account may not exist if we have not collected it
							slog.DebugContext(
								ctx,
								"Trust account not found for domain SID and NetBIOS",
								slog.String("trusting_domain_sid", trustingDomainSid),
								slog.String("netbios", netbios),
							)
							continue
						} else {
							channels.Submit(ctx, outC, post.EnsureRelationshipJob{
								FromID: domain.ID,
								ToID:   trustAccount.ID,
								Kind:   ad.HasTrustKeys,
							})
						}
					}
				}
			}
			return nil
		}); err != nil {
			return &post.AtomicPostProcessingStats{}, fmt.Errorf("error creating HasTrustKeys edges: %w", err)
		}

		return &operation.Stats, operation.Done()
	}
}

// FetchNodeIDsByKind fetches a bitmap of node IDs where each node has at least one kind assignment
// that matches the given kind.
func FetchNodeIDsByKind(tx graph.Transaction, targetKind graph.Kind) (cardinality.Duplex[uint64], error) {
	defer measure.LogAndMeasure(
		slog.LevelInfo,
		"FetchNodeIDsByKind",
		slog.String("kind", targetKind.String()),
		attr.Namespace("analysis"),
		attr.Function("FetchNodeIDsByKind"),
		attr.Scope("routine"),
	)()

	nodes := cardinality.NewBitmap64()

	if err := tx.Nodes().Filterf(func() graph.Criteria {
		return query.Kind(query.Node(), targetKind)
	}).FetchIDs(func(cursor graph.Cursor[graph.ID]) error {
		for id := range cursor.Chan() {
			nodes.Add(id.Uint64())
		}

		return cursor.Error()
	}); err != nil {
		return nil, err
	}

	return nodes, nil
}

// FetchComputerIDsWithLocalToComputer fetches a bitmap of Computer node IDs that have at least one
// inbound LocalToComputer edge.
func FetchComputerIDsWithLocalToComputer(tx graph.Transaction) (cardinality.Duplex[uint64], error) {
	defer measure.LogAndMeasure(
		slog.LevelInfo,
		"FetchComputerIDsWithLocalToComputer",
		attr.Namespace("analysis"),
		attr.Function("FetchComputerIDsWithLocalToComputer"),
		attr.Scope("routine"),
	)()

	var (
		computers = cardinality.NewBitmap64()
	)

	if err := tx.Relationships().Filterf(func() graph.Criteria {
		return query.And(
			query.Kind(query.Relationship(), ad.LocalToComputer),
			query.Kind(query.End(), ad.Computer),
		)
	}).FetchKinds(func(cursor graph.Cursor[graph.RelationshipKindsResult]) error {
		for result := range cursor.Chan() {
			computers.Add(result.EndID.Uint64())
		}

		return cursor.Error()
	}); err != nil {
		return nil, err
	}

	return computers, nil
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

func getDirectOutboundTrustDomains(tx graph.Transaction, domain *graph.Node) (graph.NodeSet, error) {
	return ops.FetchEndNodes(tx.Relationships().Filterf(func() graph.Criteria {
		return query.And(
			query.Equals(query.StartID(), domain.ID),
			query.KindIn(query.Relationship(), ad.SameForestTrust, ad.CrossForestTrust),
			query.Kind(query.End(), ad.Domain),
		)
	}))
}

func getTrustAccount(tx graph.Transaction, domainSid, netbios string) (*graph.Node, error) {
	nodes, err := ops.FetchNodes(tx.Nodes().Filterf(func() graph.Criteria {
		return query.And(
			query.Kind(query.Node(), ad.User),
			query.Equals(query.NodeProperty(ad.DomainSID.String()), domainSid),
			query.Equals(query.NodeProperty(ad.SamAccountName.String()), netbios+"$"),
		)
	}).Limit(1))
	if err != nil {
		return nil, err
	}
	if len(nodes) == 0 {
		return nil, graph.ErrNoResultsFound
	}
	return nodes[0], err
}

func fromEntityToEntityWithRelationshipKind(tx graph.Transaction, target *graph.Node, relKind graph.Kind) graph.RelationshipQuery {
	return tx.Relationships().Filterf(func() graph.Criteria {
		filters := []graph.Criteria{
			query.Kind(query.Start(), ad.Entity),
			query.Kind(query.Relationship(), relKind),
			query.Equals(query.EndID(), target.ID),
		}

		return query.And(filters...)
	})
}

func getLAPSSyncers(tx graph.Transaction, domain *graph.Node, localGroupData *LocalGroupData) (cardinality.Duplex[uint64], error) {
	var (
		getChangesQuery         = fromEntityToEntityWithRelationshipKind(tx, domain, ad.GetChanges)
		getChangesFilteredQuery = fromEntityToEntityWithRelationshipKind(tx, domain, ad.GetChangesInFilteredSet)
	)

	if getChangesNodes, err := ops.FetchStartNodes(getChangesQuery); err != nil {
		return nil, err
	} else if getChangesFilteredNodes, err := ops.FetchStartNodes(getChangesFilteredQuery); err != nil {
		return nil, err
	} else {
		results := CalculateCrossProductNodeSets(localGroupData, NewCachedPrincipalSet(getChangesNodes.Slice()), NewCachedPrincipalSet(getChangesFilteredNodes.Slice()))

		return results, nil
	}
}

func getDCSyncers(tx graph.Transaction, domain *graph.Node, localGroupData *LocalGroupData) (cardinality.Duplex[uint64], error) {
	var (
		getChangesQuery    = fromEntityToEntityWithRelationshipKind(tx, domain, ad.GetChanges)
		getChangesAllQuery = fromEntityToEntityWithRelationshipKind(tx, domain, ad.GetChangesAll)
	)

	if getChangesNodes, err := ops.FetchStartNodes(getChangesQuery); err != nil {
		return nil, err
	} else if getChangesAllNodes, err := ops.FetchStartNodes(getChangesAllQuery); err != nil {
		return nil, err
	} else {
		results := CalculateCrossProductNodeSets(localGroupData, NewCachedPrincipalSet(getChangesNodes.Slice()), NewCachedPrincipalSet(getChangesAllNodes.Slice()))

		return results, nil
	}
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

func getAdminSDHolder(tx graph.Transaction, domain *graph.Node) ([]graph.ID, error) {
	if domainSid, err := domain.Properties.Get(ad.DomainSID.String()).String(); err != nil {
		return nil, err
	} else {
		return ops.FetchNodeIDs(tx.Nodes().Filterf(func() graph.Criteria {
			return query.And(
				query.KindIn(query.Node(), ad.Container),
				query.StringStartsWith(query.NodeProperty(ad.DistinguishedName.String()), AdminSDHolderDNPrefix),
				query.Equals(query.NodeProperty(ad.DomainSID.String()), domainSid),
			)
		}))
	}
}

func getAdminSDHolderProtected(tx graph.Transaction, domain *graph.Node) ([]graph.ID, error) {
	if domainSid, err := domain.Properties.Get(ad.DomainSID.String()).String(); err != nil {
		return nil, err
	} else {
		return ops.FetchNodeIDs(tx.Nodes().Filterf(func() graph.Criteria {
			return query.And(
				query.KindIn(query.Node(), ad.Computer, ad.User, ad.Group),
				query.Equals(
					query.Property(query.Node(), ad.AdminSDHolderProtected.String()), true),
				query.Equals(query.Property(query.Node(), ad.DomainSID.String()), domainSid),
			)
		}))
	}
}

// Fetches a LocalGroup belonging to the given computer by the given LocalGroup SID suffix.
func FetchComputerLocalGroupBySIDSuffix(tx graph.Transaction, computer graph.ID, groupSuffix string) (*graph.Node, error) {
	var (
		groupNode graph.Node
		err       = tx.Relationships().Filter(query.And(
			query.StringEndsWith(query.StartProperty(common.ObjectID.String()), groupSuffix),
			query.Kind(query.Relationship(), ad.LocalToComputer),
			query.InIDs(query.EndID(), computer),
		)).Query(
			func(results graph.Result) error {
				defer results.Close()

				if results.Next() {
					if err := results.Scan(&groupNode); err != nil {
						return err
					}
				} else {
					return graph.ErrNoResultsFound
				}

				return results.Error()
			},
			query.Returning(
				query.Start(),
			),
		)
	)

	if err != nil {
		return nil, err
	}

	return &groupNode, nil
}

// FetchComputerLocalGroupIDBySIDSuffix fetches a local group attached to the given computer with a SID suffix that matches
// the given suffix.
func FetchComputerLocalGroupIDBySIDSuffix(tx graph.Transaction, computer graph.ID, groupSuffix string) (graph.ID, error) {
	var (
		startID graph.ID
		err     = tx.Relationships().Filter(query.And(
			query.StringEndsWith(query.StartProperty(common.ObjectID.String()), groupSuffix),
			query.Kind(query.Relationship(), ad.LocalToComputer),
			query.InIDs(query.EndID(), computer),
		)).Query(
			func(results graph.Result) error {
				defer results.Close()

				if results.Next() {
					if err := results.Scan(&startID); err != nil {
						return err
					}
				} else {
					return graph.ErrNoResultsFound
				}

				return results.Error()
			},
			query.Returning(
				query.StartID(),
			),
		)
	)

	if err != nil {
		return 0, err
	}

	return startID, nil
}

func ExpandGroupMembershipIDBitmap(tx graph.Transaction, group *graph.Node) (cardinality.Duplex[uint64], error) {
	groupMembers := cardinality.NewBitmap64()

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

// FetchComputerLocalGroupByName looks up a local group attacked to a given computer by its name.
func FetchComputerLocalGroupByName(tx graph.Transaction, computer graph.ID, groupName string) (*graph.Node, error) {
	if rel, err := tx.Relationships().Filter(
		query.And(
			query.Kind(query.Start(), ad.LocalGroup),
			query.CaseInsensitiveStringStartsWith(query.StartProperty(common.Name.String()), groupName),
			query.Kind(query.Relationship(), ad.LocalToComputer),
			query.InIDs(query.EndID(), computer),
		),
	).First(); err != nil {
		return nil, err
	} else {
		return ops.FetchNode(tx, rel.StartID)
	}
}

// FetchRemoteDesktopUsersBitmapForComputerWithoutURA uses the cached local group information in the passed CanRDPComputerData
// struct to compute the membership of the computer's "Remote Desktop Users" local group. This membership is returned
// as a bitmap.
func FetchRemoteDesktopUsersBitmapForComputerWithoutURA(canRDPData *CanRDPComputerData) cardinality.Duplex[uint64] {
	adjacentNodes := container.AdjacentNodes(
		canRDPData.LocalGroupMembershipDigraph,
		canRDPData.RemoteDesktopUsersLocalGroup.ID.Uint64(),
		graph.DirectionInbound,
	)

	return cardinality.NewBitmap64With(adjacentNodes...)
}

// FetchRemoteInteractiveLogonRightEntities expands all RemoteInteractiveLogonRight to a given computer and returns the
// nodes as a set.
func FetchRemoteInteractiveLogonRightEntities(tx graph.Transaction, computerId graph.ID) (graph.NodeSet, error) {
	return ops.FetchStartNodes(tx.Relationships().Filterf(func() graph.Criteria {
		return query.And(
			query.Kind(query.Relationship(), ad.RemoteInteractiveLogonRight),
			query.Equals(query.EndID(), computerId),
		)
	}))
}

// FetchCanRDPEntityBitmapForComputer computes eligible nodes (aggregated into a bitmap) for the given computer
// in the passed CanRDPComputerData struct.
func FetchCanRDPEntityBitmapForComputer(computerData *CanRDPComputerData, enforceURA bool, citrixEnabled bool) (cardinality.Duplex[uint64], error) {
	var (
		uraEnabled = enforceURA || computerData.ComputersWithURA.Contains(computerData.Computer.Uint64())

		// Shortcut opportunity when citrix is disabled: see if the RDP group has RIL privilege. If
		// it does, get the first degree members and return those ids, since everything in RDP group
		// has CanRDP privs. No reason to look any further.
		canSkipURAProcessing = !uraEnabled || computerData.HasRemoteInteractiveLogonRight()
	)

	if citrixEnabled {
		if computerData.DAUGroup == nil {
			// "Direct Access Users" is a group that Citrix creates.  If the group does not exist, then the computer does not have Citrix installed and post-processing logic can continue by enumerating the "Remote Desktop Users" AD group.
			if canSkipURAProcessing {
				return FetchRemoteDesktopUsersBitmapForComputerWithoutURA(computerData), nil
			} else {
				return FetchRemoteDesktopUsersBitmapForComputerWithURA(computerData)
			}
		}

		if !uraEnabled {
			// In cases where we do not need to check for the existence of the RIL privilege, return the cross product of both groups
			return CalculateCrossProductNodeSets(&computerData.LocalGroupData, NewCachedPrincipalSet([]*graph.Node{computerData.RemoteDesktopUsersLocalGroup}), NewCachedPrincipalSet([]*graph.Node{computerData.DAUGroup})), nil
		} else {
			// Otherwise, return the cross product of all three criteria
			return CalculateCrossProductNodeSets(&computerData.LocalGroupData, NewCachedPrincipalSet([]*graph.Node{computerData.RemoteDesktopUsersLocalGroup}), NewCachedPrincipalSet([]*graph.Node{computerData.DAUGroup}), NewCachedPrincipalSet(computerData.RemoteInteractiveLogonRightEntities.Slice())), nil
		}
	} else if canSkipURAProcessing {
		return FetchRemoteDesktopUsersBitmapForComputerWithoutURA(computerData), nil
	} else {
		return FetchRemoteDesktopUsersBitmapForComputerWithURA(computerData)
	}
}

// FetchComputersWithURA fetches all computers with the "hasura" property set to true and
// aggregates the computer IDs into a bitmap.
func FetchComputersWithURA(tx graph.Transaction) (cardinality.Duplex[uint64], error) {
	defer measure.LogAndMeasure(
		slog.LevelInfo,
		"FetchComputersWithURA",
		attr.Namespace("analysis"),
		attr.Function("FetchComputersWithURA"),
		attr.Scope("routine"),
	)()

	nodesWithURA := cardinality.NewBitmap64()

	if err := tx.Nodes().Filter(
		query.And(
			query.Kind(query.Node(), ad.Computer),
			query.Equals(query.NodeProperty(ad.HasURA.String()), true),
		),
	).Query(func(results graph.Result) error {
		for results.Next() {
			var (
				nodeID        graph.ID
				propertyValue bool
			)

			if err := results.Scan(&nodeID, &propertyValue); err != nil {
				return err
			} else if propertyValue {
				nodesWithURA.Add(nodeID.Uint64())
			}
		}

		return results.Error()
	}, query.Returning(
		query.NodeID(),
		query.NodeProperty(ad.HasURA.String()),
	)); err != nil {
		return nil, err
	}

	return nodesWithURA, nil
}

// LocalGroupData contains data common to AD local group and domain group post-processing business logic. This allows for
// business logic to avoid database interactions.
type LocalGroupData struct {
	// All computer IDs in all domains
	Computers cardinality.Duplex[uint64]

	// All computer IDs in all domains that have at least one inbound LocalToComputer edge
	ComputersWithLocalGroups cardinality.Duplex[uint64]

	// All group IDs in all domains
	Groups cardinality.Duplex[uint64]

	// All edges where: (:Base)-[:MemberOf|MemberOfLocalGroup*..]->(:Group|LocalGroup)
	GroupMembershipCache *algo.ReachabilityCache

	// All edges where: (:Base)-[:MemberOfLocalGroup]->(:LocalGroup)
	LocalGroupMembershipDigraph container.DirectedGraph

	// Contains groups that we want to stop post-processed edge propagation at, for example: EVERYONE@DOMAIN.COM
	ExcludedShortcutGroups cardinality.Duplex[uint64]

	// Unrolled variant to avoid allocations during iteration for excluded shortcut groups
	ExcludedShortcutGroupsSlice []uint64
}

// FetchLocalGroupData access the given graph database and fetches all of the required data for LocalGroup post processing.
func FetchLocalGroupData(ctx context.Context, graphDB graph.Database) (*LocalGroupData, error) {
	defer measure.ContextLogAndMeasure(
		ctx,
		slog.LevelInfo,
		"Fetching local group data",
		attr.Namespace("analysis"),
		attr.Function("FetchLocalGroupData"),
		attr.Scope("process"),
	)()

	localGroupData := &LocalGroupData{}

	if err := graphDB.ReadTransaction(ctx, func(tx graph.Transaction) error {
		if excludedGroups, err := FetchAuthUsersAndEveryoneGroups(tx); err != nil {
			return err
		} else {
			localGroupData.ExcludedShortcutGroups = excludedGroups.IDBitmap()
			localGroupData.ExcludedShortcutGroupsSlice = localGroupData.ExcludedShortcutGroups.Slice()
		}

		if computerIDs, err := FetchNodeIDsByKind(tx, ad.Computer); err != nil {
			return err
		} else {
			localGroupData.Computers = computerIDs
		}

		if computersWithLocalGroups, err := FetchComputerIDsWithLocalToComputer(tx); err != nil {
			return err
		} else {
			localGroupData.ComputersWithLocalGroups = computersWithLocalGroups
		}

		if allGroupIDs, err := FetchNodeIDsByKind(tx, ad.Group); err != nil {
			return err
		} else {
			localGroupData.Groups = allGroupIDs
		}

		return nil
	}); err != nil {
		return nil, err
	}

	if groupMembershipCache, err := algo.FetchFilteredReachabilityCache(ctx, graphDB, ad.MemberOf, ad.MemberOfLocalGroup); err != nil {
		return nil, err
	} else {
		localGroupData.GroupMembershipCache = groupMembershipCache
	}

	if localGroupMembershipDigraph, err := container.FetchFilteredDirectedGraph(ctx, graphDB, ad.MemberOfLocalGroup); err != nil {
		return nil, err
	} else {
		localGroupData.LocalGroupMembershipDigraph = localGroupMembershipDigraph
	}

	return localGroupData, nil
}

// FetchCanRDPData access the given graph database and fetches all of the required data for
// CanRDP post processing that is not unqiue to a single computer. This allows for these data
// elements to be shared between post processing runs for each computer.
func (s *LocalGroupData) FetchCanRDPData(ctx context.Context, graphDB graph.Database) (*CanRDPData, error) {
	components := &CanRDPData{
		LocalGroupData: *s,
	}

	if err := graphDB.ReadTransaction(ctx, func(tx graph.Transaction) error {
		if computersWithURA, err := FetchComputersWithURA(tx); err != nil {
			return err
		} else {
			components.ComputersWithURA = computersWithURA
		}

		return nil
	}); err != nil {
		return components, err
	}

	if remoteInteractiveLogonRightDigraph, err := container.FetchFilteredDirectedGraph(ctx, graphDB, ad.RemoteInteractiveLogonRight); err != nil {
		return components, err
	} else {
		components.RemoteInteractiveLogonRightDigraph = remoteInteractiveLogonRightDigraph
	}

	return components, nil
}

// CanRDPData contains data common to CanRDP post-processing business logic. This allows for
// business logic to avoid database interactions.
type CanRDPData struct {
	LocalGroupData

	// Duplex of computer IDs that have the "hasura" property set to true
	ComputersWithURA cardinality.Duplex[uint64]

	// All edges where: (:Base)-[:RemoteInteractiveLogonRight]->(:Computer)
	RemoteInteractiveLogonRightDigraph container.DirectedGraph
}

// CanRDPData contains data common to CanRDP post-processing business logic for a single computer. This allows for
// business logic to avoid database interactions.
type CanRDPComputerData struct {
	CanRDPData

	// Computer ID being analyzed
	Computer graph.ID

	// The Citrix "Direct Access Users" group
	DAUGroup *graph.Node

	// The "Remote Desktop Users" LocalGroup attached to this computer
	RemoteDesktopUsersLocalGroup *graph.Node

	// Nodes that have a RemoteInteractiveLogonRight edge inbound to this computer
	RemoteInteractiveLogonRightEntities graph.NodeSet
}

// FetchCanRDPComputerData uses the given transaction to fetch all computer-specific data related to the given computer
// that is required to compute the computer's inbound CanRDP edges.
func (s *CanRDPData) FetchCanRDPComputerData(tx graph.Transaction, computer graph.ID) (*CanRDPComputerData, error) {
	computerData := &CanRDPComputerData{
		CanRDPData: *s,
		Computer:   computer,
	}

	if dauGroup, err := FetchComputerLocalGroupByName(tx, computer, "Direct Access Users"); err != nil {
		if !graph.IsErrNotFound(err) {
			return nil, err
		}
	} else {
		computerData.DAUGroup = dauGroup
	}

	if remoteDesktopUsersLocalGroup, err := FetchComputerLocalGroupBySIDSuffix(tx, computer, wellknown.RemoteDesktopUsersSIDSuffix.String()); err != nil {
		return nil, err
	} else {
		computerData.RemoteDesktopUsersLocalGroup = remoteDesktopUsersLocalGroup
	}

	if rilEntities, err := FetchRemoteInteractiveLogonRightEntities(tx, computer); err != nil {
		return nil, err
	} else {
		computerData.RemoteInteractiveLogonRightEntities = rilEntities
	}

	return computerData, nil
}

// HasRemoteInteractiveLogonRight looks up if the associated Remote Desktop Users Local Group has
// a valid RemoteInteractiveLogonRight edge to the computer.
func (s *CanRDPComputerData) HasRemoteInteractiveLogonRight() bool {
	found := false

	s.RemoteInteractiveLogonRightDigraph.EachAdjacentNode(s.RemoteDesktopUsersLocalGroup.ID.Uint64(), graph.DirectionOutbound, func(adjacent uint64) bool {
		found = adjacent == s.Computer.Uint64()
		return !found
	})

	return found
}

func FetchRemoteDesktopUsersBitmapForComputerWithURA(canRDPData *CanRDPComputerData) (cardinality.Duplex[uint64], error) {
	var (
		rdpLocalGroupMembers = canRDPData.GroupMembershipCache.ReachOfComponentContainingMember(canRDPData.RemoteDesktopUsersLocalGroup.ID.Uint64(), graph.DirectionInbound)
		baseRILEntities      = container.AdjacentNodes(canRDPData.RemoteInteractiveLogonRightDigraph, canRDPData.Computer.Uint64(), graph.DirectionInbound)
		rdpEntities          = cardinality.NewBitmap64()
		secondaryTargetMaps  []cardinality.Duplex[uint64]
	)

	// Attempt 1: look at each RIL entity directly and see if it has membership to the RDP group. If not, and it's a group, expand its membership for further processing
	for _, entityID := range baseRILEntities {
		if rdpLocalGroupMembers.Contains(entityID) {
			// If we have membership to the RDP group, then this is a valid CanRDP entity
			rdpEntities.Add(entityID)
		} else {
			secondaryTargetMaps = append(secondaryTargetMaps, canRDPData.GroupMembershipCache.ReachOfComponentContainingMember(entityID, graph.DirectionInbound))
		}
	}

	// Attempt 2: Look at each member of expanded groups and see if they have the correct permissions
	for _, secondaryTargetMap := range secondaryTargetMaps {
		// If we have membership to the RDP group then this is a valid CanRDP entity
		secondaryTargetMap.Each(func(entity uint64) bool {
			if rdpLocalGroupMembers.Contains(entity) {
				rdpEntities.Add(entity)
			}

			return true
		})
	}

	return rdpEntities, nil
}

func Post(ctx context.Context, db graph.Database, citrixEnabled, ntlmEnabled bool) (*post.AtomicPostProcessingStats, error) {
	defer measure.ContextLogAndMeasure(
		ctx,
		slog.LevelInfo,
		"Active Directory Post Processing",
		attr.Namespace("analysis"),
		attr.Function("Post"),
		attr.Scope("step"),
	)()

	aggregateStats := post.NewAtomicPostProcessingStats()

	if err := FixWellKnownNodeTypes(ctx, db); err != nil {
		return &aggregateStats, err
	} else if err := RunDomainAssociations(ctx, db); err != nil {
		return &aggregateStats, err
	} else if err := LinkWellKnownNodes(ctx, db); err != nil {
		return &aggregateStats, err
	} else if deleteTransitEdgesStats, err := post.DeleteTransitEdges(ctx, db, graph.Kinds{ad.Entity, azure.Entity}, ad.PostProcessedRelationships()); err != nil {
		return &aggregateStats, err
	} else if localGroupData, err := FetchLocalGroupData(ctx, db); err != nil {
		return &aggregateStats, err
	} else if dcSyncStats, err := PostDCSync(ctx, db, localGroupData); err != nil {
		return &aggregateStats, err
	} else if protectAdminGroupsStats, err := PostProtectAdminGroups(ctx, db); err != nil {
		return &aggregateStats, err
	} else if syncLAPSStats, err := PostSyncLAPSPassword(ctx, db, localGroupData); err != nil {
		return &aggregateStats, err
	} else if hasTrustKeyStats, err := PostHasTrustKeys(ctx, db); err != nil {
		return &aggregateStats, err
	} else if localGroupStats, err := PostLocalGroups(ctx, db, localGroupData); err != nil {
		return &aggregateStats, err
	} else if canRDPStats, err := PostCanRDP(ctx, db, localGroupData, true, citrixEnabled); err != nil {
		return &aggregateStats, err
	} else if adcsStats, adcsCache, err := PostADCS(ctx, db, localGroupData); err != nil {
		return &aggregateStats, err
	} else if ownsStats, err := PostOwnsAndWriteOwner(ctx, db, localGroupData); err != nil {
		return &aggregateStats, err
	} else if canReanimateTombstoneStats, err := PostCanReanimateTombstone(ctx, db, localGroupData); err != nil {
		return &aggregateStats, err
	} else if canUseBadSuccessorStats, err := PostCanUseBadSuccessor(ctx, db, localGroupData); err != nil {
		return &aggregateStats, err
	} else if ntlmStats, err := PostNTLM(ctx, db, localGroupData, adcsCache, ntlmEnabled); err != nil {
		return &aggregateStats, err
	} else {
		aggregateStats.Merge(deleteTransitEdgesStats)
		aggregateStats.Merge(syncLAPSStats)
		aggregateStats.Merge(hasTrustKeyStats)
		aggregateStats.Merge(dcSyncStats)
		aggregateStats.Merge(protectAdminGroupsStats)
		aggregateStats.Merge(localGroupStats)
		aggregateStats.Merge(canRDPStats)
		aggregateStats.Merge(adcsStats)
		aggregateStats.Merge(ownsStats)
		aggregateStats.Merge(canReanimateTombstoneStats)
		aggregateStats.Merge(canUseBadSuccessorStats)
		aggregateStats.Merge(ntlmStats)

		return &aggregateStats, nil
	}
}
