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
	entraDSSyncScopeAll            = "ALL"
)

var entraDSPostProcessedEdges = graph.Kinds{
	azure.SyncedToEntraDSUser,
	azure.SyncedToEntraDSGroup,
	azure.AddEntraDSGroupMember,
	azure.EntraDSFor,
	azure.ManageEntraDSSync,
	azure.ManageEntraDSSyncFilter,
}

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

func PostHybrid(ctx context.Context, db graph.Database, entraDomainServicesEnabled bool) (*post.AtomicPostProcessingStats, error) {
	defer measure.ContextLogAndMeasure(
		ctx,
		slog.LevelInfo,
		"Post-processing AD-Azure Hybrid Edges",
		attr.Namespace("analysis"),
		attr.Function("PostHybrid"),
		attr.Scope("process"),
	)()

	emptyStats := post.NewAtomicPostProcessingStats()
	tracker, err := post.FetchTracker(ctx, db, entraDSPostProcessedEdges)
	if err != nil {
		return &emptyStats, err
	}

	sink := post.NewFilteredRelationshipSink(ctx, "Entra DS Hybrid Attack Paths Post Processing", db, tracker)

	// Fetch all Azure tenants first
	tenants, err := fetchTenants(ctx, db)
	if err != nil {
		sink.Done()
		return &emptyStats, fmt.Errorf("fetching Entra tenants: %w", err)
	}

	// Spin up a new parallel operation to speed up processing
	operation := post.NewPostRelationshipOperation(ctx, db, "Hybrid Attack Paths Post Processing")

	err = db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		var (
			// adObjIDMap is used as a reverse mapping of a list of Entra node ids indexed by the AD user objectids
			adObjIDMap = make(map[string][]graph.ID, 1024)
			// entraToADMap is the final mapping between an Entra user node id to an AD user node id
			entraToADMap                   = make(map[graph.ID]graph.ID, 1024)
			entraDSUserAADObjectIDMap      = make(map[string][]graph.ID, 1024)
			entraDSGroupAADObjectIDMap     = make(map[string][]graph.ID, 1024)
			entraDSAdminGroupTenantMap     = make(map[graph.ID]string, 16)
			syncedToEntraDSUserEdgeMap     = make(map[graph.ID][]graph.ID, 1024)
			syncedToEntraDSGroupEdgeMap    = make(map[graph.ID][]graph.ID, 1024)
			addEntraDSGroupMemberEdgeMap   = make(map[graph.ID][]graph.ID, 1024)
			entraDSForEdgeMap              = make(map[graph.ID][]graph.ID, 16)
			manageEntraDSSyncEdgeMap       = make(map[graph.ID][]graph.ID, 16)
			manageEntraDSSyncFilterEdgeMap = make(map[graph.ID][]graph.ID, 16)
		)

		// Work on Entra users by their tenant association. Loop therefore through each Entra tenant
		for _, tenant := range tenants {
			// Fetch all users in this Entra tenant
			if tenantUsers, err := fetchEntraUsers(tx, tenant); err != nil {
				return err
			} else {
				// Loop through each Entra user in this tenant
				for _, tenantUser := range tenantUsers {
					if entraDomainServicesEnabled {
						if err := addNodeToObjectIDMap(entraDSUserAADObjectIDMap, tenantUser); err != nil {
							return err
						}
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

			if entraDomainServicesEnabled {
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

				if entraDomainServicesEnabled {
					if err := addSyncedToEntraDSEdges(syncedToEntraDSUserEdgeMap, adUser, entraDSUserAADObjectIDMap); err != nil {
						return err
					}
				}
			}
		}

		if entraDomainServicesEnabled {
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
				return fmt.Errorf("adding Entra DS group membership relationships: %w", err)
			}

			// A qualified AZManageEntraDS principal controls the broad synchronization boundary. The Domain Controller
			// Services service principal can only add users through filtered group scope when the related managed domain is
			// currently configured for filtered synchronization across all users.
			if err := addManageEntraDSSyncEdges(tx, adGroups, entraDSAdminGroupTenantMap, syncedToEntraDSGroupEdgeMap, entraDSForEdgeMap, manageEntraDSSyncEdgeMap, manageEntraDSSyncFilterEdgeMap); err != nil {
				if !errors.Is(err, graph.ErrNoResultsFound) {
					return fmt.Errorf("adding Entra DS synchronization relationships: %w", err)
				}

				// The synchronization-control relationships depend on optional evidence from both graph platforms. Missing
				// evidence must fail closed for those relationships without suppressing the independently supported identity
				// correlation and group-control relationships assembled above.
				slog.WarnContext(ctx, "Skipping incomplete Entra DS synchronization correlation", attr.Error(err))
			}

			for kind, edgeMap := range map[graph.Kind]map[graph.ID][]graph.ID{
				azure.SyncedToEntraDSUser:     reverseRelationshipMap(syncedToEntraDSUserEdgeMap),
				azure.SyncedToEntraDSGroup:    reverseRelationshipMap(syncedToEntraDSGroupEdgeMap),
				azure.AddEntraDSGroupMember:   addEntraDSGroupMemberEdgeMap,
				azure.EntraDSFor:              entraDSForEdgeMap,
				azure.ManageEntraDSSync:       manageEntraDSSyncEdgeMap,
				azure.ManageEntraDSSyncFilter: manageEntraDSSyncFilterEdgeMap,
			} {
				if err := submitMappedRelationships(ctx, sink, edgeMap, kind); err != nil {
					return err
				}
			}
		}

		if err := operation.Operation.SubmitReader(func(ctx context.Context, _ graph.Transaction, outC chan<- post.EnsureRelationshipJob) error {
			for azUser, adUser := range entraToADMap {
				syncedToEntraUserRelationship := post.EnsureRelationshipJob{
					FromID: adUser,
					ToID:   azUser,
					Kind:   azure.SyncedToEntraUser,
				}

				if !channels.Submit(ctx, outC, syncedToEntraUserRelationship) {
					return nil
				}

				syncedToADUserRelationship := post.EnsureRelationshipJob{
					FromID: azUser,
					ToID:   adUser,
					Kind:   adSchema.SyncedToADUser,
				}

				if !channels.Submit(ctx, outC, syncedToADUserRelationship) {
					return nil
				}
			}

			return nil
		}); err != nil {
			return err
		}

		return nil
	})

	// Close both writers even when the read phase failed so in-flight workers cannot leak. Keep the read and write
	// failures distinct; formatting a nil operation error with %w obscures the actual cause.
	opErr := operation.Done()
	sink.Done()

	aggregateStats := post.NewAtomicPostProcessingStats()
	aggregateStats.Merge(&operation.Stats)
	aggregateStats.Merge(sink.Stats())

	if opErr != nil {
		if err != nil {
			return &aggregateStats, fmt.Errorf("marking hybrid operation as done: %w; read error: %v", opErr, err)
		}

		return &aggregateStats, fmt.Errorf("marking hybrid operation as done: %w", opErr)
	} else if err != nil {
		return &aggregateStats, fmt.Errorf("reading hybrid relationship inputs: %w", err)
	}

	return &aggregateStats, nil
}

func reverseRelationshipMap(edgeMap map[graph.ID][]graph.ID) map[graph.ID][]graph.ID {
	reversedEdgeMap := make(map[graph.ID][]graph.ID)
	for sourceNodeID, targetNodeIDs := range edgeMap {
		for _, targetNodeID := range targetNodeIDs {
			reversedEdgeMap[targetNodeID] = append(reversedEdgeMap[targetNodeID], sourceNodeID)
		}
	}

	return reversedEdgeMap
}

func submitMappedRelationships(ctx context.Context, sink *post.FilteredRelationshipSink, edgeMap map[graph.ID][]graph.ID, kind graph.Kind) error {
	for sourceNodeID, targetNodeIDs := range edgeMap {
		for _, targetNodeID := range targetNodeIDs {
			if !sink.Submit(ctx, post.EnsureRelationshipJob{
				FromID: sourceNodeID,
				ToID:   targetNodeID,
				Kind:   kind,
			}) {
				return fmt.Errorf("submitting %s relationship", kind)
			}
		}
	}

	return nil
}

func addNodeToObjectIDMap(nodeObjectIDMap map[string][]graph.ID, node *graph.Node) error {
	if objectID, hasObjectID, err := normalizedNodeProperty(node, common.ObjectID.String()); err != nil {
		return err
	} else if hasObjectID {
		nodeObjectIDMap[objectID] = append(nodeObjectIDMap[objectID], node.ID)
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
	if name, hasName, err := normalizedNodeProperty(group, common.Name.String()); err != nil {
		return err
	} else if !hasName || !strings.HasPrefix(name, entraDSAdminGroupNamePrefix) {
		return nil
	} else if tenantID, hasTenantID, err := normalizedNodeProperty(group, azure.TenantID.String()); err != nil {
		return err
	} else if hasTenantID {
		entraDSAdminGroupTenantMap[group.ID] = tenantID
	}

	return nil
}

// addManageEntraDSSyncEdges correlates each AZEntraDS resource to an AD Domain by normalized domain name and the
// synchronized AAD DC Administrators group. The correlated Domain SID identifies Domain Users by RID 513. Principals
// with AZManageEntraDS receive the broad synchronization edge only when Domain Users is reachable through AD
// containment. The known Domain Controller Services service principal receives the filter-specific edge only when
// filtered synchronization is enabled with sync scope All.
func addManageEntraDSSyncEdges(tx graph.Transaction, adGroups []*graph.Node, entraDSAdminGroupTenantMap map[graph.ID]string, syncedToEntraDSGroupEdgeMap, entraDSForEdgeMap, manageEntraDSSyncEdgeMap, manageEntraDSSyncFilterEdgeMap map[graph.ID][]graph.ID) error {
	var (
		adGroupsByID                 = make(map[graph.ID]*graph.Node, len(adGroups))
		domainUsersByDomainSID       = make(map[string][]graph.ID)
		adminGroupDomainSIDsByTenant = make(map[string]map[string]struct{})
		domainsByName                = make(map[string][]*graph.Node)
		scopedSyncTargetsByTenant    = make(map[string][]graph.ID)
		manageSyncSeen               = make(map[string]struct{})
		manageFilterSeen             = make(map[string]struct{})
	)

	for _, adGroup := range adGroups {
		adGroupsByID[adGroup.ID] = adGroup

		objectID, hasObjectID, err := normalizedNodeProperty(adGroup, common.ObjectID.String())
		if err != nil {
			return err
		} else if !hasObjectID {
			continue
		}

		domainSID, hasDomainSID, err := normalizedNodeProperty(adGroup, adSchema.DomainSID.String())
		if err != nil {
			return err
		} else if hasDomainSID && objectID == domainSID+domainUsersObjectIDSuffix {
			domainUsersByDomainSID[domainSID] = append(domainUsersByDomainSID[domainSID], adGroup.ID)
		}
	}

	for adAdminGroupID, azGroupIDs := range syncedToEntraDSGroupEdgeMap {
		adAdminGroup, hasADAdminGroup := adGroupsByID[adAdminGroupID]
		if !hasADAdminGroup {
			continue
		}

		domainSID, hasDomainSID, err := normalizedNodeProperty(adAdminGroup, adSchema.DomainSID.String())
		if err != nil {
			return err
		} else if !hasDomainSID {
			continue
		}

		for _, azGroupID := range azGroupIDs {
			if tenantID, isEntraDSAdminGroup := entraDSAdminGroupTenantMap[azGroupID]; isEntraDSAdminGroup {
				if _, ok := adminGroupDomainSIDsByTenant[tenantID]; !ok {
					adminGroupDomainSIDsByTenant[tenantID] = make(map[string]struct{})
				}
				adminGroupDomainSIDsByTenant[tenantID][domainSID] = struct{}{}
			}
		}
	}

	domains, err := fetchADDomains(tx)
	if err != nil {
		return fmt.Errorf("fetching AD domains: %w", err)
	}

	for _, domain := range domains {
		domainName, hasDomainName, err := normalizedNodeProperty(domain, common.Name.String())
		if err != nil {
			return err
		} else if hasDomainName {
			domainsByName[domainName] = append(domainsByName[domainName], domain)
		}
	}

	domainServices, err := fetchEntraDomainServices(tx)
	if err != nil {
		return fmt.Errorf("fetching Entra DS resources: %w", err)
	}

	for _, domainService := range domainServices {
		tenantID, hasTenantID, err := normalizedNodeProperty(domainService, azure.TenantID.String())
		if err != nil {
			return err
		} else if !hasTenantID {
			continue
		}

		domainName, hasDomainName, err := normalizedNodeProperty(domainService, azure.DomainName.String())
		if err != nil {
			return err
		} else if !hasDomainName {
			continue
		}

		candidateDomains := domainsByName[domainName]
		if len(candidateDomains) != 1 {
			continue
		}

		domain := candidateDomains[0]
		domainSID, hasDomainSID, err := normalizedNodeProperty(domain, adSchema.DomainSID.String())
		if err != nil {
			return err
		} else if !hasDomainSID {
			continue
		} else if tenantDomainSIDs := adminGroupDomainSIDsByTenant[tenantID]; tenantDomainSIDs == nil {
			continue
		} else if _, corroborated := tenantDomainSIDs[domainSID]; !corroborated {
			continue
		}

		addMappedRelationship(entraDSForEdgeMap, nil, domainService.ID, domain.ID)

		domainUserGroups := domainUsersByDomainSID[domainSID]
		if len(domainUserGroups) == 0 {
			continue
		}

		containedDomainUserGroups, err := filterContainedDomainUsers(tx, domain, domainUserGroups)
		if err != nil {
			return fmt.Errorf("finding Domain Users containment for domain %d: %w", domain.ID, err)
		}

		if len(containedDomainUserGroups) > 0 {
			managers, err := ops.FetchStartNodes(tx.Relationships().Filterf(func() graph.Criteria {
				return query.And(
					query.InIDs(query.EndID(), domainService.ID),
					query.Kind(query.Relationship(), azure.ManageEntraDS),
					query.KindIn(query.Start(), azure.User, azure.Group, azure.ServicePrincipal),
				)
			}))
			if err != nil {
				return fmt.Errorf("fetching Entra DS managers for resource %d: %w", domainService.ID, err)
			}

			for _, manager := range managers {
				for _, domainUserGroupID := range containedDomainUserGroups {
					addMappedRelationship(manageEntraDSSyncEdgeMap, manageSyncSeen, manager.ID, domainUserGroupID)
				}
			}
		}

		if allowed, err := allowsScopedSyncServicePrincipalEdge(domainService); err != nil {
			return err
		} else if allowed {
			scopedSyncTargetsByTenant[tenantID] = append(scopedSyncTargetsByTenant[tenantID], domainUserGroups...)
		}
	}

	if len(scopedSyncTargetsByTenant) == 0 {
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
		return fmt.Errorf("fetching Domain Controller Services application relationships: %w", err)
	}

	for _, runsAsRelationship := range runsAsRelationships {
		// Merged Azure application and service-principal records can produce a
		// self-referential AZRunsAs edge. It cannot identify two distinct sides
		// of the scoped-sync application relationship and FetchRelationshipNodes
		// intentionally requires two nodes, so ignore it as non-evidence.
		if runsAsRelationship.StartID == runsAsRelationship.EndID {
			continue
		}

		application, servicePrincipal, err := ops.FetchRelationshipNodes(tx, runsAsRelationship)
		if err != nil {
			return fmt.Errorf("fetching endpoints for AZRunsAs relationship %d: %w", runsAsRelationship.ID, err)
		}

		applicationID, hasApplicationID, err := normalizedNodeProperty(application, common.ObjectID.String())
		if err != nil {
			return err
		} else if !hasApplicationID || applicationID != entraDSScopedSyncApplicationID {
			continue
		}

		servicePrincipalTenantID, hasTenantID, err := normalizedNodeProperty(servicePrincipal, azure.TenantID.String())
		if err != nil {
			return err
		} else if !hasTenantID {
			continue
		}

		for _, domainUserGroupID := range scopedSyncTargetsByTenant[servicePrincipalTenantID] {
			addMappedRelationship(manageEntraDSSyncFilterEdgeMap, manageFilterSeen, servicePrincipal.ID, domainUserGroupID)
		}
	}

	return nil
}

func filterContainedDomainUsers(tx graph.Transaction, domain *graph.Node, domainUserGroupIDs []graph.ID) ([]graph.ID, error) {
	targets := make(map[graph.ID]struct{}, len(domainUserGroupIDs))
	for _, domainUserGroupID := range domainUserGroupIDs {
		targets[domainUserGroupID] = struct{}{}
	}

	paths, err := ops.TraversePaths(tx, ops.TraversalPlan{
		Root:      domain,
		Direction: graph.DirectionOutbound,
		BranchQuery: func() graph.Criteria {
			return query.Kind(query.Relationship(), adSchema.Contains)
		},
		PathFilter: func(_ *ops.TraversalContext, segment *graph.PathSegment) bool {
			_, isDomainUserGroup := targets[segment.Node.ID]
			return isDomainUserGroup
		},
	})
	// PostgreSQL reports an empty traversal as ErrNoResultsFound while other
	// graph drivers return an empty PathSet. Missing containment means that the
	// broad synchronization edge is not supported; it must not fail the entire
	// hybrid post-processing operation or suppress unrelated derived edges.
	if errors.Is(err, graph.ErrNoResultsFound) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	reachable := paths.AllNodes()
	containedDomainUserGroupIDs := make([]graph.ID, 0, len(domainUserGroupIDs))
	for _, domainUserGroupID := range domainUserGroupIDs {
		if _, ok := reachable[domainUserGroupID]; ok {
			containedDomainUserGroupIDs = append(containedDomainUserGroupIDs, domainUserGroupID)
		}
	}

	return containedDomainUserGroupIDs, nil
}

func allowsScopedSyncServicePrincipalEdge(domainService *graph.Node) (bool, error) {
	filteredSyncEnabled, err := domainService.Properties.Get(azure.FilteredSyncEnabled.String()).Bool()
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

	return filteredSyncEnabled && normalizeObjectID(syncScope) == entraDSSyncScopeAll, nil
}

func addMappedRelationship(edgeMap map[graph.ID][]graph.ID, seen map[string]struct{}, sourceNodeID, targetNodeID graph.ID) {
	if seen != nil {
		key := sourceNodeID.String() + "|" + targetNodeID.String()
		if _, duplicate := seen[key]; duplicate {
			return
		}
		seen[key] = struct{}{}
	}

	edgeMap[sourceNodeID] = append(edgeMap[sourceNodeID], targetNodeID)
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

	azGroupIDs := make([]graph.ID, 0, len(azGroupToADGroups))
	for azGroupID := range azGroupToADGroups {
		azGroupIDs = append(azGroupIDs, azGroupID)
	}

	// Fetch only AZAddMembers / AZOwns relationships that target an Entra DS-synced AZGroup.
	memberAddEdges, err := ops.FetchRelationships(tx.Relationships().Filterf(func() graph.Criteria {
		return query.And(
			query.KindIn(query.Relationship(), azure.AddMembers, azure.Owns),
			query.InIDs(query.EndID(), azGroupIDs...),
		)
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

func normalizedNodeProperty(node *graph.Node, property string) (string, bool, error) {
	value, err := node.Properties.Get(property).String()
	if errors.Is(err, graph.ErrPropertyNotFound) {
		return "", false, nil
	} else if err != nil {
		return "", false, err
	}

	normalizedValue := normalizeObjectID(value)
	return normalizedValue, normalizedValue != "", nil
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

func fetchADDomains(tx graph.Transaction) ([]*graph.Node, error) {
	return ops.FetchNodes(tx.Nodes().Filterf(func() graph.Criteria {
		return query.Kind(query.Node(), adSchema.Domain)
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
