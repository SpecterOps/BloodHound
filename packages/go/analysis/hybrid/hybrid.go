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

package hybrid

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/specterops/bloodhound/packages/go/analysis/post"
	"github.com/specterops/bloodhound/packages/go/bhlog/attr"
	"github.com/specterops/bloodhound/packages/go/bhlog/measure"
	adSchema "github.com/specterops/bloodhound/packages/go/graphschema/ad"
	"github.com/specterops/bloodhound/packages/go/graphschema/azure"
	"github.com/specterops/bloodhound/packages/go/graphschema/common"
	"github.com/specterops/dawgs/graph"
	"github.com/specterops/dawgs/ops"
	"github.com/specterops/dawgs/query"
	"github.com/specterops/dawgs/util/channels"
)

const (
	entraDSAdminGroupNamePrefix    = "AAD DC ADMINISTRATORS@"
	entraDSScopedSyncApplicationID = "2565BD9D-DA50-47D4-8B85-4C97F669DC36"
	domainUsersObjectIDSuffix      = "-513"
	entraDSFilteredSyncEnabled     = "ENABLED"
	entraDSSyncScopeAll            = "ALL"
)

func fetchTenants(ctx context.Context, db graph.Database) (graph.NodeSet, error) {
	var nodeSet graph.NodeSet
	if err := db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		var err error
		if nodeSet, err = ops.FetchNodeSet(tx.Nodes().Filterf(func() graph.Criteria {
			return query.Kind(query.Node(), azure.Tenant)
		})); err != nil {
			return err
		} else {
			return nil
		}
	}); err != nil {
		return nil, err
	} else {
		return nodeSet, nil
	}
}

func PostHybrid(ctx context.Context, db graph.Database) (*post.AtomicPostProcessingStats, error) {
	defer measure.ContextLogAndMeasure(
		ctx,
		slog.LevelInfo,
		"Post-processing AD-Azure Hybrid Edges",
		attr.Namespace("analysis"),
		attr.Function("PostHybrid"),
		attr.Scope("process"),
	)()

	// Fetch all Azure tenants first
	tenants, err := fetchTenants(ctx, db)
	if err != nil {
		emptyStats := post.NewAtomicPostProcessingStats()
		return &emptyStats, fmt.Errorf("fetching Entra tenants: %w", err)
	}

	// Spin up a new parallel operation to speed up processing
	operation := post.NewPostRelationshipOperation(ctx, db, "Hybrid Attack Paths Post Processing")

	err = db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		var (
			// adObjIDMap is used as a reverse mapping of a list of Entra node ids indexed by the AD user objectids
			adObjIDMap = make(map[string][]graph.ID, 1024)
			// entraToADMap is the final mapping between an Entra user node id to an AD user node id
			entraToADMap                 = make(map[graph.ID]graph.ID, 1024)
			entraDSUserAADObjectIDMap    = make(map[string][]graph.ID, 1024)
			entraDSGroupAADObjectIDMap   = make(map[string][]graph.ID, 1024)
			entraDSAdminGroupTenantMap   = make(map[graph.ID]string, 16)
			syncedToEntraDSUserEdgeMap   = make(map[graph.ID][]graph.ID, 1024)
			syncedToEntraDSGroupEdgeMap  = make(map[graph.ID][]graph.ID, 1024)
			addEntraDSGroupMemberEdgeMap = make(map[graph.ID][]graph.ID, 1024)
			syncEntraDSUsersEdgeMap      = make(map[graph.ID][]graph.ID, 16)
		)

		// Work on Entra users by their tenant association. Loop therefore through each Entra tenant
		for _, tenant := range tenants {
			// Fetch all users in this Entra tenant
			if tenantUsers, err := fetchEntraUsers(tx, tenant); err != nil {
				return err
			} else {
				// Loop through each Entra user in this tenant
				for _, tenantUser := range tenantUsers {
					if err := addNodeToObjectIDMap(entraDSUserAADObjectIDMap, tenantUser); err != nil {
						return err
					}

					// Check to see if the Entra user has an on prem sync property set
					if onPremID, hasOnPrem, err := hasOnPremUser(tenantUser); !hasOnPrem {
						continue
					} else if err != nil {
						return err
					} else {
						// We know this user has an onPrem counterpart, so add the node id and onPremID to our mapping inputs.
						adObjIDMap[onPremID] = append(adObjIDMap[onPremID], tenantUser.ID)
					}
				}
			}

			if tenantGroups, err := fetchEntraGroups(tx, tenant); err != nil {
				return err
			} else {
				for _, tenantGroup := range tenantGroups {
					if err := addNodeToObjectIDMap(entraDSGroupAADObjectIDMap, tenantGroup); err != nil {
						return err
					}

					if err := addEntraDSAdminGroupTenant(entraDSAdminGroupTenantMap, tenantGroup); err != nil {
						return err
					}
				}
			}
		}

		// Because there's a chance for AD users to exist in the graph without having a valid domain node linked to them,
		// we need to grab all of them directly, unlike Entra
		if adUsers, err := fetchADUsers(tx); err != nil {
			return err
		} else {
			// Loop through each Active Directory user
			for _, adUser := range adUsers {
				// Get the user's Object ID
				if objectID, err := adUser.Properties.Get(common.ObjectID.String()).String(); err != nil {
					return err
				} else if azUsers, ok := adObjIDMap[objectID]; ok {
					// Because there could theoretically be more than one Entra user mapped to this objectid, we want to loop through all when adding our current id to the final map
					for _, azUser := range azUsers {
						entraToADMap[azUser] = adUser.ID
					}
				}

				if err := addSyncedToEntraDSEdges(syncedToEntraDSUserEdgeMap, adUser, entraDSUserAADObjectIDMap); err != nil {
					return err
				}
			}
		}

		adGroups, err := fetchADGroups(tx)
		if err != nil {
			return err
		}

		for _, adGroup := range adGroups {
			if err := addSyncedToEntraDSEdges(syncedToEntraDSGroupEdgeMap, adGroup, entraDSGroupAADObjectIDMap); err != nil {
				return err
			}
		}

		// Now that we know which AZ users and AZ groups are synced to Entra Domain Services, compute the
		// AddEntraDSGroupMember edges (an Entra DS-synced AZUser that can add or remove members from an Entra DS-synced AZGroup)
		if err := addAddEntraDSGroupMemberEdges(tx, syncedToEntraDSUserEdgeMap, syncedToEntraDSGroupEdgeMap, addEntraDSGroupMemberEdgeMap); err != nil {
			return err
		}

		// The managed domain controls the broad synchronization boundary. The Domain Controller Services service
		// principal can only add users through filtered group scope when the related managed domain is currently
		// configured for filtered synchronization across all users.
		if err := addSyncEntraDSUsersEdges(tx, adGroups, entraDSAdminGroupTenantMap, syncedToEntraDSGroupEdgeMap, syncEntraDSUsersEdgeMap); err != nil {
			return err
		}

		if err := operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- post.EnsureRelationshipJob) error {
			for azUser, adUser := range entraToADMap {
				SyncedToEntraUserRelationship := post.EnsureRelationshipJob{
					FromID: adUser,
					ToID:   azUser,
					Kind:   azure.SyncedToEntraUser,
				}

				if !channels.Submit(ctx, outC, SyncedToEntraUserRelationship) {
					return nil
				}

				SyncedToADUserRelationship := post.EnsureRelationshipJob{
					FromID: azUser,
					ToID:   adUser,
					Kind:   adSchema.SyncedToADUser,
				}

				if !channels.Submit(ctx, outC, SyncedToADUserRelationship) {
					return nil
				}
			}

			for adNode, azNodes := range syncedToEntraDSUserEdgeMap {
				for _, azNode := range azNodes {
					syncedToEntraDSUserRelationship := post.EnsureRelationshipJob{
						FromID: azNode,
						ToID:   adNode,
						Kind:   azure.SyncedToEntraDSUser,
					}

					if !channels.Submit(ctx, outC, syncedToEntraDSUserRelationship) {
						return nil
					}
				}
			}

			for adNode, azNodes := range syncedToEntraDSGroupEdgeMap {
				for _, azNode := range azNodes {
					syncedToEntraDSGroupRelationship := post.EnsureRelationshipJob{
						FromID: azNode,
						ToID:   adNode,
						Kind:   azure.SyncedToEntraDSGroup,
					}

					if !channels.Submit(ctx, outC, syncedToEntraDSGroupRelationship) {
						return nil
					}
				}
			}

			for azUser, adGroups := range addEntraDSGroupMemberEdgeMap {
				for _, adGroup := range adGroups {
					addEntraDSGroupMemberRelationship := post.EnsureRelationshipJob{
						FromID: azUser,
						ToID:   adGroup,
						Kind:   azure.AddEntraDSGroupMember,
					}

					if !channels.Submit(ctx, outC, addEntraDSGroupMemberRelationship) {
						return nil
					}
				}
			}

			for sourceNode, domainUserGroups := range syncEntraDSUsersEdgeMap {
				for _, domainUserGroup := range domainUserGroups {
					syncEntraDSUsersRelationship := post.EnsureRelationshipJob{
						FromID: sourceNode,
						ToID:   domainUserGroup,
						Kind:   azure.SyncEntraDSUsers,
					}

					if !channels.Submit(ctx, outC, syncEntraDSUsersRelationship) {
						return nil
					}
				}
			}

			return nil
		}); err != nil {
			return err
		}

		return tx.Commit()
	})

	// Because we need to close the operation either way at this stage, we attempt to close it and then report either or
	// both errors in one line
	if opErr := operation.Done(); opErr != nil || err != nil {
		return &operation.Stats, fmt.Errorf("marking operation as done: %w; transaction error (if any): %v", opErr, err)
	}

	return &operation.Stats, nil
}

func addNodeToObjectIDMap(nodeObjectIDMap map[string][]graph.ID, node *graph.Node) error {
	if objectID, err := node.Properties.Get(common.ObjectID.String()).String(); err != nil {
		return err
	} else if normalizedObjectID := normalizeObjectID(objectID); len(normalizedObjectID) != 0 {
		nodeObjectIDMap[normalizedObjectID] = append(nodeObjectIDMap[normalizedObjectID], node.ID)
	}

	return nil
}

func addSyncedToEntraDSEdges(edgeMap map[graph.ID][]graph.ID, adNode *graph.Node, azNodeMap map[string][]graph.ID) error {
	if aadObjectID, hasAADObjectID, err := getEntraDSAADObjectID(adNode); err != nil {
		return err
	} else if !hasAADObjectID {
		return nil
	} else if azNodeIDs, ok := azNodeMap[aadObjectID]; ok {
		edgeMap[adNode.ID] = append(edgeMap[adNode.ID], azNodeIDs...)
	}

	return nil
}

func addEntraDSAdminGroupTenant(entraDSAdminGroupTenantMap map[graph.ID]string, group *graph.Node) error {
	if name, err := group.Properties.Get(common.Name.String()).String(); err != nil {
		return err
	} else if !strings.HasPrefix(strings.ToUpper(strings.TrimSpace(name)), entraDSAdminGroupNamePrefix) {
		return nil
	} else if tenantID, err := group.Properties.Get(azure.TenantID.String()).String(); err != nil {
		return err
	} else if normalizedTenantID := normalizeObjectID(tenantID); len(normalizedTenantID) != 0 {
		entraDSAdminGroupTenantMap[group.ID] = normalizedTenantID
	}

	return nil
}

// addSyncEntraDSUsersEdges computes the SyncEntraDSUsers edges. A synchronized AAD DC Administrators group identifies
// each tenant's Entra Domain Services domain, and the domain SID identifies its Domain Users group. The managed domain
// always receives an edge because control of the ARM resource can change the synchronization boundary. The known
// Domain Controller Services service principal receives a narrower edge only when filtered synchronization is enabled
// with sync scope All.
func addSyncEntraDSUsersEdges(tx graph.Transaction, adGroups []*graph.Node, entraDSAdminGroupTenantMap map[graph.ID]string, syncedToEntraDSGroupEdgeMap, syncEntraDSUsersEdgeMap map[graph.ID][]graph.ID) error {
	var (
		adGroupsByID              = make(map[graph.ID]*graph.Node, len(adGroups))
		domainUsersByDomainSID    = make(map[string][]graph.ID)
		domainUserGroupsByTenant  = make(map[string][]graph.ID)
		scopedSyncAllowedByTenant = make(map[string]bool)
		seen                      = make(map[string]struct{})
	)

	for _, adGroup := range adGroups {
		adGroupsByID[adGroup.ID] = adGroup

		if objectID, err := adGroup.Properties.Get(common.ObjectID.String()).String(); errors.Is(err, graph.ErrPropertyNotFound) {
			continue
		} else if err != nil {
			return err
		} else if !strings.HasSuffix(strings.ToUpper(strings.TrimSpace(objectID)), domainUsersObjectIDSuffix) {
			continue
		} else if domainSID, err := adGroup.Properties.Get(adSchema.DomainSID.String()).String(); errors.Is(err, graph.ErrPropertyNotFound) {
			continue
		} else if err != nil {
			return err
		} else if normalizedDomainSID := normalizeObjectID(domainSID); len(normalizedDomainSID) != 0 {
			domainUsersByDomainSID[normalizedDomainSID] = append(domainUsersByDomainSID[normalizedDomainSID], adGroup.ID)
		}
	}

	if len(domainUsersByDomainSID) == 0 || len(entraDSAdminGroupTenantMap) == 0 {
		return nil
	}

	for adAdminGroupID, azGroupIDs := range syncedToEntraDSGroupEdgeMap {
		adAdminGroup, hasADAdminGroup := adGroupsByID[adAdminGroupID]
		if !hasADAdminGroup {
			continue
		}

		domainSID, err := adAdminGroup.Properties.Get(adSchema.DomainSID.String()).String()
		if errors.Is(err, graph.ErrPropertyNotFound) {
			continue
		} else if err != nil {
			return err
		}

		domainUserGroups := domainUsersByDomainSID[normalizeObjectID(domainSID)]
		if len(domainUserGroups) == 0 {
			continue
		}

		for _, azGroupID := range azGroupIDs {
			if tenantID, isEntraDSAdminGroup := entraDSAdminGroupTenantMap[azGroupID]; isEntraDSAdminGroup {
				domainUserGroupsByTenant[tenantID] = append(domainUserGroupsByTenant[tenantID], domainUserGroups...)
			}
		}
	}

	if len(domainUserGroupsByTenant) == 0 {
		return nil
	}

	domainServices, err := fetchEntraDomainServices(tx)
	if err != nil {
		return err
	}

	for _, domainService := range domainServices {
		domainServiceTenantID, err := domainService.Properties.Get(azure.TenantID.String()).String()
		if errors.Is(err, graph.ErrPropertyNotFound) {
			continue
		} else if err != nil {
			return err
		}

		normalizedTenantID := normalizeObjectID(domainServiceTenantID)
		for _, domainUserGroupID := range domainUserGroupsByTenant[normalizedTenantID] {
			addSyncEntraDSUsersEdge(syncEntraDSUsersEdgeMap, seen, domainService.ID, domainUserGroupID)
		}

		if allowed, err := allowsScopedSyncServicePrincipalEdge(domainService); err != nil {
			return err
		} else if allowed {
			scopedSyncAllowedByTenant[normalizedTenantID] = true
		}
	}

	if len(scopedSyncAllowedByTenant) == 0 {
		return nil
	}

	runsAsRelationships, err := ops.FetchRelationships(tx.Relationships().Filterf(func() graph.Criteria {
		return query.And(
			query.Kind(query.Relationship(), azure.RunsAs),
			query.Kind(query.Start(), azure.App),
			query.Kind(query.End(), azure.ServicePrincipal),
		)
	}))
	if err != nil {
		return err
	}

	for _, runsAsRelationship := range runsAsRelationships {
		application, servicePrincipal, err := ops.FetchRelationshipNodes(tx, runsAsRelationship)
		if err != nil {
			return err
		}

		applicationID, err := application.Properties.Get(common.ObjectID.String()).String()
		if err != nil {
			return err
		} else if normalizeObjectID(applicationID) != entraDSScopedSyncApplicationID {
			continue
		}

		servicePrincipalTenantID, err := servicePrincipal.Properties.Get(azure.TenantID.String()).String()
		if err != nil {
			return err
		}

		normalizedTenantID := normalizeObjectID(servicePrincipalTenantID)
		if !scopedSyncAllowedByTenant[normalizedTenantID] {
			continue
		}

		for _, domainUserGroupID := range domainUserGroupsByTenant[normalizedTenantID] {
			addSyncEntraDSUsersEdge(syncEntraDSUsersEdgeMap, seen, servicePrincipal.ID, domainUserGroupID)
		}
	}

	return nil
}

func allowsScopedSyncServicePrincipalEdge(domainService *graph.Node) (bool, error) {
	filteredSync, err := domainService.Properties.Get(azure.FilteredSync.String()).String()
	if errors.Is(err, graph.ErrPropertyNotFound) {
		return false, nil
	} else if err != nil {
		return false, err
	}

	syncScope, err := domainService.Properties.Get(azure.SyncScope.String()).String()
	if errors.Is(err, graph.ErrPropertyNotFound) {
		return false, nil
	} else if err != nil {
		return false, err
	}

	return normalizeObjectID(filteredSync) == entraDSFilteredSyncEnabled && normalizeObjectID(syncScope) == entraDSSyncScopeAll, nil
}

func addSyncEntraDSUsersEdge(syncEntraDSUsersEdgeMap map[graph.ID][]graph.ID, seen map[string]struct{}, sourceNodeID, domainUserGroupID graph.ID) {
	key := sourceNodeID.String() + "|" + domainUserGroupID.String()
	if _, duplicate := seen[key]; duplicate {
		return
	}

	seen[key] = struct{}{}
	syncEntraDSUsersEdgeMap[sourceNodeID] = append(syncEntraDSUsersEdgeMap[sourceNodeID], domainUserGroupID)
}

// addAddEntraDSGroupMemberEdges computes the AddEntraDSGroupMember edges. An edge is created from an AZUser to an
// on-prem Group when the AZUser is synced to Entra Domain Services, the AZUser owns or can add and remove members from an AZGroup
// (AZOwns / AZAddMembers), and that AZGroup is itself synced to Entra Domain Services. The resulting edge is drawn
// from the AZUser to the on-prem Group that the AZGroup is synced to.
func addAddEntraDSGroupMemberEdges(tx graph.Transaction, syncedToEntraDSUserEdgeMap, syncedToEntraDSGroupEdgeMap, addEntraDSGroupMemberEdgeMap map[graph.ID][]graph.ID) error {
	// Build the set of AZUser node ids that are synced to Entra Domain Services
	entraDSSyncedAZUsers := make(map[graph.ID]struct{}, len(syncedToEntraDSUserEdgeMap))
	for _, azUserIDs := range syncedToEntraDSUserEdgeMap {
		for _, azUserID := range azUserIDs {
			entraDSSyncedAZUsers[azUserID] = struct{}{}
		}
	}

	// Build a reverse mapping of Entra DS-synced AZGroup node ids to the on-prem Group node ids they are synced to
	azGroupToADGroups := make(map[graph.ID][]graph.ID, len(syncedToEntraDSGroupEdgeMap))
	for adGroupID, azGroupIDs := range syncedToEntraDSGroupEdgeMap {
		for _, azGroupID := range azGroupIDs {
			azGroupToADGroups[azGroupID] = append(azGroupToADGroups[azGroupID], adGroupID)
		}
	}

	// No AddEntraDSGroupMember edges are possible unless there is at least one synced AZUser and one synced AZGroup
	if len(entraDSSyncedAZUsers) == 0 || len(azGroupToADGroups) == 0 {
		return nil
	}

	// Fetch all AZAddMembers / AZOwns relationships. Filtering the end node against azGroupToADGroups below naturally
	// restricts these to relationships that target an Entra DS-synced AZGroup.
	memberAddEdges, err := ops.FetchRelationships(tx.Relationships().Filterf(func() graph.Criteria {
		return query.KindIn(query.Relationship(), azure.AddMembers, azure.Owns)
	}))
	if err != nil {
		return err
	}

	// Track emitted (AZUser, Group) pairs so an AZUser holding both AZOwns and AZAddMembers over the same group only
	// yields a single edge
	seen := make(map[string]struct{})
	for _, edge := range memberAddEdges {
		if _, ok := entraDSSyncedAZUsers[edge.StartID]; !ok {
			continue
		} else if adGroupIDs, ok := azGroupToADGroups[edge.EndID]; ok {
			for _, adGroupID := range adGroupIDs {
				key := edge.StartID.String() + "|" + adGroupID.String()
				if _, dup := seen[key]; dup {
					continue
				}
				seen[key] = struct{}{}
				addEntraDSGroupMemberEdgeMap[edge.StartID] = append(addEntraDSGroupMemberEdgeMap[edge.StartID], adGroupID)
			}
		}
	}

	return nil
}

func getEntraDSAADObjectID(node *graph.Node) (string, bool, error) {
	if aadObjectID, err := node.Properties.Get(adSchema.AADObjectID.String()).String(); errors.Is(err, graph.ErrPropertyNotFound) {
		return "", false, nil
	} else if err != nil {
		return "", false, err
	} else if normalizedAADObjectID := normalizeObjectID(aadObjectID); len(normalizedAADObjectID) == 0 {
		return "", false, nil
	} else {
		return normalizedAADObjectID, true, nil
	}
}

func normalizeObjectID(objectID string) string {
	return strings.ToUpper(strings.TrimSpace(objectID))
}

// hasOnPremUser takes a node and returns the OnPremID as a string, whether the node has an onPrem user defined as a bool
// and any errors in negotiation of the required properties
func hasOnPremUser(node *graph.Node) (string, bool, error) {
	if onPremSyncEnabled, err := node.Properties.Get(azure.OnPremSyncEnabled.String()).Bool(); errors.Is(err, graph.ErrPropertyNotFound) {
		return "", false, nil
	} else if err != nil {
		return "", false, err
	} else if onPremID, err := node.Properties.Get(azure.OnPremID.String()).String(); errors.Is(err, graph.ErrPropertyNotFound) {
		return onPremID, false, nil
	} else if err != nil {
		return onPremID, false, err
	} else {
		return onPremID, (onPremSyncEnabled && len(onPremID) != 0), nil
	}
}

// fetchEntraUsers fetches all the Entra users for a given root node (generally the tenant node)
func fetchEntraUsers(tx graph.Transaction, root *graph.Node) (graph.NodeSet, error) {
	return ops.FetchEndNodes(tx.Relationships().Filterf(func() graph.Criteria {
		return query.And(
			query.InIDs(query.StartID(), root.ID),
			query.Kind(query.Relationship(), azure.Contains),
			query.KindIn(query.End(), azure.User),
		)
	}))
}

// fetchEntraGroups fetches all the Entra groups for a given root node (generally the tenant node)
func fetchEntraGroups(tx graph.Transaction, root *graph.Node) (graph.NodeSet, error) {
	return ops.FetchEndNodes(tx.Relationships().Filterf(func() graph.Criteria {
		return query.And(
			query.InIDs(query.StartID(), root.ID),
			query.Kind(query.Relationship(), azure.Contains),
			query.KindIn(query.End(), azure.Group),
		)
	}))
}

func fetchEntraDomainServices(tx graph.Transaction) ([]*graph.Node, error) {
	return ops.FetchNodes(tx.Nodes().Filterf(func() graph.Criteria {
		return query.Kind(query.Node(), azure.EntraDS)
	}))
}

// fetchADUsers gets all AD Users in the graph
func fetchADUsers(tx graph.Transaction) ([]*graph.Node, error) {
	return ops.FetchNodes(tx.Nodes().Filterf(func() graph.Criteria {
		return query.And(
			query.Kind(query.Node(), adSchema.User),
		)
	}))
}

// fetchADGroups gets all AD Groups in the graph
func fetchADGroups(tx graph.Transaction) ([]*graph.Node, error) {
	return ops.FetchNodes(tx.Nodes().Filterf(func() graph.Criteria {
		return query.And(
			query.Kind(query.Node(), adSchema.Group),
		)
	}))
}
