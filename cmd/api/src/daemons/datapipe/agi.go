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

package datapipe

import (
	"context"
	"strings"
	"sync"

	"github.com/RoaringBitmap/roaring/roaring64"
	commonanalysis "github.com/specterops/bloodhound/analysis"
	adAnalysis "github.com/specterops/bloodhound/analysis/ad"
	azureAnalysis "github.com/specterops/bloodhound/analysis/azure"
	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/dawgs/ops"
	"github.com/specterops/bloodhound/dawgs/query"
	"github.com/specterops/bloodhound/graphschema/ad"
	"github.com/specterops/bloodhound/graphschema/azure"
	"github.com/specterops/bloodhound/graphschema/common"
	"github.com/specterops/bloodhound/log"
	"github.com/specterops/bloodhound/src/database"
	"github.com/specterops/bloodhound/src/model"
	"github.com/specterops/bloodhound/src/services/agi"
)

func updateAssetGroupIsolationTags(ctx context.Context, db agi.AgiData, graphDB graph.Database) error {
	defer log.Measure(log.LevelInfo, "Updated asset group isolation tags")()

	if err := commonanalysis.ClearSystemTags(ctx, graphDB); err != nil {
		return err
	}

	return agi.UpdateAssetGroupIsolationTags(ctx, db, graphDB)
}

func ParallelTagAzureTierZero(ctx context.Context, db graph.Database) error {
	defer log.Measure(log.LevelInfo, "Finished tagging Azure Tier Zero")()

	var tenants graph.NodeSet

	if err := db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		if innerTenants, err := ops.FetchNodeSet(tx.Nodes().Filterf(func() graph.Criteria {
			return query.Kind(query.Node(), azure.Tenant)
		})); err != nil {
			return err
		} else {
			tenants = innerTenants
			return nil
		}
	}); err != nil {
		return err
	} else {
		var (
			tenantC  = make(chan *graph.Node)
			rootsC   = make(chan graph.ID)
			writerWG = &sync.WaitGroup{}
			readerWG = &sync.WaitGroup{}
		)

		readerWG.Add(1)

		go func() {
			defer readerWG.Done()

			var (
				tierZeroProperties = graph.NewProperties()
				rootIDs            []graph.ID
			)

			tierZeroProperties.Set(common.SystemTags.String(), ad.AdminTierZero)

			for rootID := range rootsC {
				seen := false

				for _, seenRootID := range rootIDs {
					if seenRootID == rootID {
						seen = true
						break
					}
				}

				if !seen {
					rootIDs = append(rootIDs, rootID)
				}
			}

			if err := db.WriteTransaction(ctx, func(tx graph.Transaction) error {
				if err := tx.Nodes().Filterf(func() graph.Criteria {
					return query.InIDs(query.NodeID(), rootIDs...)
				}).Update(tierZeroProperties); err != nil {
					return err
				}

				return nil
			}); err != nil {
				log.Errorf("Failed tagging update: %v", err)
			}
		}()

		for workerID := 0; workerID < commonanalysis.MaximumDatabaseParallelWorkers; workerID++ {
			writerWG.Add(1)

			go func(workerID int) {
				defer writerWG.Done()

				if err := db.ReadTransaction(ctx, func(tx graph.Transaction) error {
					for tenant := range tenantC {
						if roots, err := FetchAzureAttackPathRoots(tx, tenant); err != nil {
							log.Errorf("Failed fetching roots for tenant %d: %v", tenant.ID, err)
						} else {
							for _, root := range roots {
								rootsC <- root.ID
							}
						}
					}

					return nil
				}); err != nil {
					log.Errorf("Error reading attack path roots for tenants: %v", err)
				}
			}(workerID)
		}

		for _, tenant := range tenants {
			tenantC <- tenant
		}

		close(tenantC)
		writerWG.Wait()

		close(rootsC)
		readerWG.Wait()
	}

	return nil
}

func FetchAzureAttackPathRoots(tx graph.Transaction, tenant *graph.Node) (graph.NodeSet, error) {
	attackPathRoots := graph.NewNodeKindSet()

	// Add the tenant as one of the critical path roots
	attackPathRoots.Add(tenant)

	// Pull in custom tier zero tagged assets
	if customTierZeroNodes, err := azureAnalysis.FetchGraphDBTierZeroTaggedAssets(tx, tenant); err != nil {
		return nil, err
	} else {
		attackPathRoots.AddSets(customTierZeroNodes)
	}

	// The CompanyAdministratorRole, PrivilegedRoleAdministratorRole tenant roles are critical attack path roots
	if adminRoles, err := azureAnalysis.TenantRoles(tx, tenant, azure.CompanyAdministratorRole, azure.PrivilegedRoleAdministratorRole, azure.PrivilegedAuthenticationAdministratorRole); err != nil {
		return nil, err
	} else {
		attackPathRoots.AddSets(adminRoles)
	}

	// Find users that have CompanyAdministratorRole, PrivilegedRoleAdministratorRole
	if adminRoleMembers, err := azureAnalysis.RoleMembersWithGrants(tx, tenant, azure.CompanyAdministratorRole, azure.PrivilegedRoleAdministratorRole, azure.PrivilegedAuthenticationAdministratorRole); err != nil {
		return nil, err
	} else {
		for _, adminRoleMember := range adminRoleMembers {
			// Add this role member as one of the critical path roots
			attackPathRoots.Add(adminRoleMember)
		}

		// Look for any apps that may run as a critical service principal
		if criticalServicePrincipals := adminRoleMembers.ContainingNodeKinds(azure.ServicePrincipal); criticalServicePrincipals.Len() > 0 {
			if criticalApps, err := ops.FetchEndNodes(tx.Relationships().Filterf(func() graph.Criteria {
				return query.And(
					query.Kind(query.Start(), azure.App),
					query.Kind(query.Relationship(), azure.RunsAs),
					query.InIDs(query.EndID(), criticalServicePrincipals.IDs()...),
				)
			})); err != nil {
				return nil, err
			} else {
				for _, criticalApp := range criticalApps {
					// Add this app as one of the critical path roots
					attackPathRoots.Add(criticalApp)
				}
			}
		}
	}

	// Find any tenant virtual machines that are tied to an AD Admin Tier 0 security group
	if err := ops.ForEachEndNode(tx.Relationships().Filterf(func() graph.Criteria {
		return query.And(
			query.Equals(query.StartID(), tenant.ID),
			query.Kind(query.Relationship(), azure.Contains),
			query.Kind(query.End(), azure.VM),
		)
	}), func(_ *graph.Relationship, tenantVM *graph.Node) error {
		if activeDirectoryTierZeroNodes, err := ops.AcyclicTraverseTerminals(tx, ops.TraversalPlan{
			Root:      tenantVM,
			Direction: graph.DirectionOutbound,
			BranchQuery: func() graph.Criteria {
				return query.Kind(query.Relationship(), ad.MemberOf)
			},
			PathFilter: func(ctx *ops.TraversalContext, segment *graph.PathSegment) bool {
				terminalSystemTags, _ := segment.Node.Properties.GetOrDefault(common.SystemTags.String(), "").String()
				return strings.Contains(terminalSystemTags, ad.AdminTierZero)
			},
		}); err != nil {
			return err
		} else if activeDirectoryTierZeroNodes.Len() > 0 {
			// This VM is an AD computer with membership to an AD admin tier zero group. Track it as a critical
			// path root
			attackPathRoots.Add(tenantVM)
		}

		return nil
	}); err != nil {
		return nil, err
	}

	// Any ResourceGroup that contains a critical attack path root is also a critical attack path root
	if err := ops.ForEachStartNode(tx.Relationships().Filterf(func() graph.Criteria {
		return query.And(
			query.Kind(query.Start(), azure.ResourceGroup),
			query.Kind(query.Relationship(), azure.Contains),
			query.InIDs(query.EndID(), attackPathRoots.AllNodeIDs()...),
		)
	}), func(_ *graph.Relationship, node *graph.Node) error {
		// This resource group contains a critical attack path root. Track it as a critical attack path root
		attackPathRoots.Add(node)
		return nil
	}); err != nil {
		return nil, err
	}

	// Any Subscription that contains a critical ResourceGroup is also a critical attack path root
	if err := ops.ForEachStartNode(tx.Relationships().Filterf(func() graph.Criteria {
		return query.And(
			query.Kind(query.Start(), azure.Subscription),
			query.Kind(query.Relationship(), azure.Contains),
			query.InIDs(query.EndID(), attackPathRoots.Get(azure.ResourceGroup).IDs()...),
		)
	}), func(_ *graph.Relationship, node *graph.Node) error {
		// This subscription contains a critical attack path root. Track it as a critical attack path root
		attackPathRoots.Add(node)
		return nil
	}); err != nil {
		return nil, err
	}

	// Any ManagementGroup that contains a critical Subscription is also a critical attack path root
	for _, criticalSubscription := range attackPathRoots.Get(azure.Subscription) {
		walkBitmap := roaring64.New()

		if criticalManagementGroups, err := ops.AcyclicTraverseTerminals(tx, ops.TraversalPlan{
			Root:      criticalSubscription,
			Direction: graph.DirectionInbound,
			BranchQuery: func() graph.Criteria {
				return query.Kind(query.Relationship(), azure.Contains)
			},
			DescentFilter: func(ctx *ops.TraversalContext, segment *graph.PathSegment) bool {
				if nodeID := segment.Node.ID.Uint64(); !walkBitmap.Contains(nodeID) {
					walkBitmap.Add(nodeID)
					return true
				}

				return false
			},
			PathFilter: func(ctx *ops.TraversalContext, segment *graph.PathSegment) bool {
				return segment.Node.Kinds.ContainsOneOf(azure.ManagementGroup)
			},
		}); err != nil {
			return nil, err
		} else {
			attackPathRoots.AddSets(criticalManagementGroups)
		}
	}

	var (
		inboundNodes  = graph.NewNodeSet()
		tierZeroNodes = attackPathRoots.AllNodes()
	)

	// For each root look up collapsable inbound relationships to complete tier zero
	for _, attackPathRoot := range attackPathRoots.AllNodes() {
		if inboundCollapsablePaths, err := ops.TraversePaths(tx, ops.TraversalPlan{
			Root:      attackPathRoot,
			Direction: graph.DirectionInbound,
			BranchQuery: func() graph.Criteria {
				return query.KindIn(query.Relationship(), azureAnalysis.AzureNonDescentKinds()...)
			},
		}); err != nil {
			return nil, err
		} else {
			inboundNodes.AddSet(inboundCollapsablePaths.AllNodes())
		}
	}

	log.Infof("Collapsed an additional %d nodes into tier zero for non-descent relationships", inboundNodes.Len())

	tierZeroNodes.AddSet(inboundNodes)
	return tierZeroNodes, nil
}

func ParallelTagActiveDirectoryTierZero(ctx context.Context, db graph.Database) error {
	if domains, err := adAnalysis.FetchAllDomains(ctx, db); err != nil {
		return err
	} else {
		var (
			domainC  = make(chan *graph.Node)
			rootsC   = make(chan graph.ID)
			writerWG = &sync.WaitGroup{}
			readerWG = &sync.WaitGroup{}
		)

		readerWG.Add(1)

		go func() {
			defer readerWG.Done()

			var (
				tierZeroProperties = graph.NewProperties()
				rootIDs            []graph.ID
			)

			tierZeroProperties.Set(common.SystemTags.String(), ad.AdminTierZero)

			for rootID := range rootsC {
				seen := false

				for _, seenRootID := range rootIDs {
					if seenRootID == rootID {
						seen = true
						break
					}
				}

				if !seen {
					rootIDs = append(rootIDs, rootID)
				}
			}

			if err := db.WriteTransaction(ctx, func(tx graph.Transaction) error {
				if err := tx.Nodes().Filterf(func() graph.Criteria {
					return query.InIDs(query.NodeID(), rootIDs...)
				}).Update(tierZeroProperties); err != nil {
					return err
				}

				return nil
			}); err != nil {
				log.Errorf("Failed tagging update: %v", err)
			}
		}()

		for workerID := 0; workerID < commonanalysis.MaximumDatabaseParallelWorkers; workerID++ {
			writerWG.Add(1)

			go func(workerID int) {
				defer writerWG.Done()

				if err := db.ReadTransaction(ctx, func(tx graph.Transaction) error {
					for domain := range domainC {
						if roots, err := fetchActiveDirectoryTierZeroRoots(tx, domain); err != nil {
							log.Errorf("Failed fetching tier zero for domain %d: %v", domain.ID, err)
						} else {
							for _, root := range roots {
								rootsC <- root.ID
							}
						}
					}

					return nil
				}); err != nil {
					log.Errorf("Error reading tier zero for domains: %v", err)
				}
			}(workerID)
		}

		for _, domain := range domains {
			domainC <- domain
		}

		close(domainC)
		writerWG.Wait()

		close(rootsC)
		readerWG.Wait()
	}

	return nil
}

func fetchActiveDirectoryTierZeroRoots(tx graph.Transaction, domain *graph.Node) (graph.NodeSet, error) {
	log.Infof("Fetching tier zero nodes for domain %d", domain.ID)
	defer log.Measure(log.LevelInfo, "Finished fetching tier zero nodes for domain %d", domain.ID)()

	if domainSID, err := domain.Properties.Get(common.ObjectID.String()).String(); err != nil {
		return nil, err
	} else {
		attackPathRoots := graph.NewNodeSet()

		// Add the domain as one of the critical path roots
		attackPathRoots.Add(domain)

		// Pull in custom tier zero tagged assets
		if customTierZeroNodes, err := adAnalysis.FetchGraphDBTierZeroTaggedAssets(tx, domainSID); err != nil {
			return nil, err
		} else {
			attackPathRoots.AddSet(customTierZeroNodes)
		}

		// Pull in well known tier zero nodes by SID suffix
		if wellKnownTierZeroNodes, err := adAnalysis.FetchWellKnownTierZeroEntities(tx, domainSID); err != nil {
			return nil, err
		} else {
			attackPathRoots.AddSet(wellKnownTierZeroNodes)
		}

		// Pull in all group members of attack path roots
		if allGroupMembers, err := adAnalysis.FetchAllGroupMembers(tx, attackPathRoots); err != nil {
			return nil, err
		} else {
			attackPathRoots.AddSet(allGroupMembers)
		}

		// Add all enforced GPO nodes to the attack path roots
		if enforcedGPOs, err := adAnalysis.FetchAllEnforcedGPOs(tx, attackPathRoots); err != nil {
			return nil, err
		} else {
			attackPathRoots.AddSet(enforcedGPOs)
		}

		// Find all next-tier assets
		return attackPathRoots, nil
	}
}

func RunAssetGroupIsolationCollections(ctx context.Context, db database.Database, graphDB graph.Database, kindGetter func(*graph.Node) string) error {
	defer log.Measure(log.LevelInfo, "Asset Group Isolation Collections")()

	if assetGroups, err := db.GetAllAssetGroups("", model.SQLFilter{}); err != nil {
		return err
	} else {
		return graphDB.WriteTransaction(ctx, func(tx graph.Transaction) error {
			for _, assetGroup := range assetGroups {
				if assetGroupNodes, err := ops.FetchNodes(tx.Nodes().Filterf(func() graph.Criteria {
					tagPropertyStr := common.SystemTags.String()

					if !assetGroup.SystemGroup {
						tagPropertyStr = common.UserTags.String()
					}

					return query.And(
						query.KindIn(query.Node(), ad.Entity, azure.Entity),
						query.StringContains(query.NodeProperty(tagPropertyStr), assetGroup.Tag),
					)
				})); err != nil {
					return err
				} else {
					var (
						entries    = make(model.AssetGroupCollectionEntries, len(assetGroupNodes))
						collection = model.AssetGroupCollection{
							AssetGroupID: assetGroup.ID,
						}
					)

					for idx, node := range assetGroupNodes {
						if objectID, err := node.Properties.Get(common.ObjectID.String()).String(); err != nil {
							log.Errorf("Node %d that does not have valid %s property", node.ID, common.ObjectID)
						} else {
							entries[idx] = model.AssetGroupCollectionEntry{
								ObjectID:   objectID,
								NodeLabel:  kindGetter(node),
								Properties: node.Properties.Map,
							}
						}
					}

					// Enter a collection, even if it's empty to signal that we did do a tagging/collection run
					if err := db.CreateAssetGroupCollection(collection, entries); err != nil {
						return err
					}
				}
			}

			return nil
		})
	}
}
