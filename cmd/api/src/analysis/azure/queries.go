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

package azure

import (
	"context"
	"fmt"
	"sync"

	"github.com/gofrs/uuid"
	"github.com/specterops/bloodhound/analysis"
	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/dawgs/ops"
	"github.com/specterops/bloodhound/dawgs/query"
	"github.com/specterops/bloodhound/graphschema/azure"
	"github.com/specterops/bloodhound/graphschema/common"
	"github.com/specterops/bloodhound/log"
	"github.com/specterops/bloodhound/src/model"
)

func GraphStats(ctx context.Context, db graph.Database) (model.AzureDataQualityStats, model.AzureDataQualityAggregation, error) {
	var (
		aggregation model.AzureDataQualityAggregation
		stats       = model.AzureDataQualityStats{}
		runID       string

		kinds = graph.Kinds{azure.User, azure.Group, azure.App, azure.ServicePrincipal, azure.Device,
			azure.ManagementGroup, azure.Subscription, azure.ResourceGroup, azure.VM, azure.KeyVault}
	)

	if newUUID, err := uuid.NewV4(); err != nil {
		return stats, aggregation, fmt.Errorf("could not generate new UUID: %w", err)
	} else {
		runID = newUUID.String()
	}

	err := db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		if tenants, err := ops.FetchNodes(tx.Nodes().Filterf(func() graph.Criteria {
			return query.Kind(query.Node(), azure.Tenant)
		})); err != nil {
			return err
		} else {
			for _, tenant := range tenants {
				if tenantObjectID, err := tenant.Properties.Get(common.ObjectID.String()).String(); err != nil {
					log.Errorf("Tenant node %d does not have a valid %s property: %v", tenant.ID, common.ObjectID, err)
				} else {
					aggregation.Tenants++

					var (
						stat = model.AzureDataQualityStat{
							RunID:    runID,
							TenantID: tenantObjectID,
						}
						operation = ops.StartNewOperation[any](ops.OperationContext{
							Parent:     ctx,
							DB:         db,
							NumReaders: analysis.MaximumDatabaseParallelWorkers,
							NumWriters: 0,
						})
						mutex = &sync.Mutex{}
					)

					for _, kind := range kinds {
						innerKind := kind
						if err := operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, _ chan<- any) error {
							if count, err := tx.Nodes().Filterf(func() graph.Criteria {
								return query.And(
									query.Kind(query.Node(), innerKind),
									query.Equals(query.NodeProperty(azure.TenantID.String()), tenantObjectID),
								)
							}).Count(); err != nil {
								return err
							} else {
								mutex.Lock()
								switch innerKind {
								case azure.User:
									stat.Users = int(count)
									aggregation.Users += int(count)

								case azure.Group:
									stat.Groups = int(count)
									aggregation.Groups += int(count)

								case azure.App:
									stat.Apps = int(count)
									aggregation.Apps += int(count)

								case azure.ServicePrincipal:
									stat.ServicePrincipals = int(count)
									aggregation.ServicePrincipals += int(count)

								case azure.Device:
									stat.Devices = int(count)
									aggregation.Devices += int(count)

								case azure.ManagementGroup:
									stat.ManagementGroups = int(count)
									aggregation.ManagementGroups += int(count)

								case azure.Subscription:
									stat.Subscriptions = int(count)
									aggregation.Subscriptions += int(count)

								case azure.ResourceGroup:
									stat.ResourceGroups = int(count)
									aggregation.ResourceGroups += int(count)

								case azure.VM:
									stat.VMs = int(count)
									aggregation.VMs += int(count)

								case azure.KeyVault:
									stat.KeyVaults = int(count)
									aggregation.KeyVaults += int(count)
								}
								mutex.Unlock()
								return nil
							}
						}); err != nil {
							return fmt.Errorf("failed while submitting reader for kind counts of type %s in tenant %s: %w", innerKind, tenantObjectID, err)
						}
					}

					if err := operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, _ chan<- any) error {
						if count, err := tx.Relationships().Filterf(func() graph.Criteria {
							return query.And(
								query.Kind(query.Start(), azure.Entity),
								query.Equals(query.StartProperty(azure.TenantID.String()), tenantObjectID),
							)
						}).Count(); err != nil {
							return err
						} else {
							mutex.Lock()
							stat.Relationships = int(count)
							aggregation.Relationships += int(count)
							mutex.Unlock()
							return nil
						}
					}); err != nil {
						return fmt.Errorf("failed while submitting reader for relationship counts in tenant %s: %w", tenantObjectID, err)
					}

					if err := operation.Done(); err != nil {
						return err
					}

					stats = append(stats, stat)
				}
			}
		}

		return nil
	})

	return stats, aggregation, err
}
