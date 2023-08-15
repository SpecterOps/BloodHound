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
	"sync"

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

		// log missing tenant IDs for easier debugging
		for _, tenant := range tenants {
			if _, err = tenant.Properties.Get(azure.TenantID.String()).String(); err != nil {
				log.Errorf("Error getting tenant id for tenant %d: %v", tenant.ID, err)
			}
		}

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
						if roots, err := azureAnalysis.FetchAzureAttackPathRoots(tx, tenant); err != nil {
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

func TagActiveDirectoryTierZero(ctx context.Context, db graph.Database) error {
	defer log.Measure(log.LevelInfo, "Finished tagging Active Directory Tier Zero")()

	if domains, err := adAnalysis.FetchAllDomains(ctx, db); err != nil {
		return err
	} else {
		for _, domain := range domains {
			if roots, err := adAnalysis.FetchActiveDirectoryTierZeroRoots(ctx, db, domain); err != nil {
				return err
			} else {
				properties := graph.NewProperties()
				properties.Set(common.SystemTags.String(), ad.AdminTierZero)

				if err := db.WriteTransaction(ctx, func(tx graph.Transaction) error {
					return tx.Nodes().Filter(query.InIDs(query.Node(), roots.IDs()...)).Update(properties)
				}); err != nil {
					return err
				}
			}
		}
	}

	return nil
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
