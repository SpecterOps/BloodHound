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

package dataquality

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/gofrs/uuid"
	"github.com/specterops/bloodhound/cmd/api/src/database"
	"github.com/specterops/bloodhound/cmd/api/src/database/types/nan"
	"github.com/specterops/bloodhound/cmd/api/src/database/types/null"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	adAnalysis "github.com/specterops/bloodhound/packages/go/analysis/ad"
	"github.com/specterops/bloodhound/packages/go/analysis/post"
	"github.com/specterops/bloodhound/packages/go/bhlog/attr"
	"github.com/specterops/bloodhound/packages/go/bhlog/measure"
	"github.com/specterops/bloodhound/packages/go/graphschema"
	"github.com/specterops/bloodhound/packages/go/graphschema/ad"
	"github.com/specterops/bloodhound/packages/go/graphschema/azure"
	"github.com/specterops/bloodhound/packages/go/graphschema/common"
	"github.com/specterops/dawgs/graph"
	"github.com/specterops/dawgs/ops"
	"github.com/specterops/dawgs/query"
)

type DataQualityData interface {
	database.DataQualityData
	GetEnvironments(ctx context.Context) ([]model.SchemaEnvironment, error)
	GetGraphSchemaNodeKindsByEnvironmentIds(ctx context.Context, environmentIDs ...int32) (map[int32]model.GraphSchemaNodeKinds, error)
	GetKindsByIDs(ctx context.Context, ids ...int32) ([]model.Kind, error)
	GetKindsByNames(ctx context.Context, names ...string) ([]model.Kind, error)
}

func adGraphStats(ctx context.Context, db graph.Database) (model.ADDataQualityStats, model.ADDataQualityAggregation, error) {
	var (
		aggregation model.ADDataQualityAggregation
		adStats     = model.ADDataQualityStats{}
		runID       string

		kinds = ad.NodeKinds()
	)

	if newUUID, err := uuid.NewV4(); err != nil {
		return adStats, aggregation, fmt.Errorf("could not generate new UUID: %w", err)
	} else {
		runID = newUUID.String()
		aggregation.RunID = runID
	}

	return adStats, aggregation, db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		if domains, err := ops.FetchNodes(tx.Nodes().Filterf(func() graph.Criteria {
			return query.Kind(query.Node(), ad.Domain)
		})); err != nil {
			return err
		} else {
			for _, domain := range domains {
				if domainSID, err := domain.Properties.Get(common.ObjectID.String()).String(); err != nil {
					slog.ErrorContext(
						ctx,
						"Domain node does not have a valid objectid property",
						slog.Uint64("domain_id", uint64(domain.ID)),
						attr.Error(err),
					)
				} else {
					aggregation.Domains++

					var (
						stat = model.ADDataQualityStat{
							RunID:     runID,
							DomainSID: domainSID,
						}
						operation = ops.StartNewOperation[any](ops.OperationContext{
							Parent:     ctx,
							DB:         db,
							NumReaders: post.MaximumDatabaseParallelWorkers,
						})
						mutex = &sync.Mutex{}
					)

					for _, kind := range kinds {
						innerKind := kind

						if innerKind == ad.Entity {
							continue
						}

						if err := operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, _ chan<- any) error {
							if count, err := tx.Nodes().Filterf(func() graph.Criteria {
								return query.And(
									query.Kind(query.Node(), innerKind),
									query.Equals(query.NodeProperty(ad.DomainSID.String()), domainSID),
								)
							}).Count(); err != nil {
								return err
							} else {
								mutex.Lock()
								switch innerKind {
								case ad.User:
									stat.Users = int(count)
									aggregation.Users += int(count)

								case ad.Group:
									stat.Groups = int(count)
									aggregation.Groups += int(count)

								case ad.Computer:
									stat.Computers = int(count)
									aggregation.Computers += int(count)

								case ad.Container:
									stat.Containers = int(count)
									aggregation.Containers += int(count)

								case ad.OU:
									stat.OUs = int(count)
									aggregation.OUs += int(count)

								case ad.GPO:
									stat.GPOs = int(count)
									aggregation.GPOs += int(count)

								case ad.AIACA:
									stat.AIACAs = int(count)
									aggregation.AIACAs += int(count)

								case ad.RootCA:
									stat.RootCAs = int(count)
									aggregation.RootCAs += int(count)

								case ad.EnterpriseCA:
									stat.EnterpriseCAs = int(count)
									aggregation.EnterpriseCAs += int(count)

								case ad.NTAuthStore:
									stat.NTAuthStores = int(count)
									aggregation.NTAuthStores += int(count)

								case ad.CertTemplate:
									stat.CertTemplates = int(count)
									aggregation.CertTemplates += int(count)

								case ad.IssuancePolicy:
									stat.IssuancePolicies = int(count)
									aggregation.IssuancePolicies += int(count)

								case ad.Domain:
									// Do nothing. Only ADDataQualityAggregation stats have domain stats and the domain stats are handled in the outer domain loop
								}

								mutex.Unlock()
								return nil
							}
						}); err != nil {
							return fmt.Errorf("failed while submitting reader for kind counts of type %s in domain %s: %w", innerKind, domainSID, err)
						}
					}

					if err := operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, _ chan<- any) error {
						// Rel counts
						if count, err := tx.Relationships().Filterf(func() graph.Criteria {
							return query.And(
								query.Kind(query.Start(), ad.Entity),
								query.Equals(query.StartProperty(ad.DomainSID.String()), domainSID),
							)
						}).Count(); err != nil {
							return err
						} else {
							mutex.Lock()
							stat.Relationships += int(count)
							aggregation.Relationships += int(count)
							mutex.Unlock()
							return nil
						}
					}); err != nil {
						return fmt.Errorf("failed while submitting reader for rel counts in domain %s: %w", domainSID, err)
					}

					if err := operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, _ chan<- any) error {
						if count, err := tx.Relationships().Filterf(func() graph.Criteria {
							return query.And(
								query.KindIn(query.Relationship(), ad.ACLRelationships()...),
								query.Equals(query.StartProperty(ad.DomainSID.String()), domainSID),
								query.Kind(query.Start(), ad.Entity),
							)
						}).Count(); err != nil {
							return err
						} else {
							mutex.Lock()
							stat.ACLs += int(count)
							aggregation.Acls += int(count)
							mutex.Unlock()
							return nil
						}
					}); err != nil {
						return fmt.Errorf("failed while submitting reader for ACL counts in domain %s: %w", domainSID, err)
					}

					if err := operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, _ chan<- any) error {
						if count, err := tx.Relationships().Filterf(func() graph.Criteria {
							return query.And(
								query.Kind(query.Relationship(), ad.HasSession),
								query.Equals(query.StartProperty(ad.DomainSID.String()), domainSID),
								query.Kind(query.Start(), ad.Computer),
							)
						}).Count(); err != nil {
							return err
						} else {
							mutex.Lock()
							stat.Sessions += int(count)
							aggregation.Sessions += int(count)
							mutex.Unlock()
							return nil
						}
					}); err != nil {
						return fmt.Errorf("failed while submitting reader for session counts in domain %s: %w", domainSID, err)
					}

					if err := operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, _ chan<- any) error {
						// Get completeness lastly
						if userSessionCompleteness, err := adAnalysis.FetchUserSessionCompleteness(tx, domainSID); err != nil {
							return err
						} else {
							mutex.Lock()
							stat.SessionCompleteness = nan.Float64(userSessionCompleteness)
							mutex.Unlock()
							return nil
						}
					}); err != nil {
						return fmt.Errorf("failed while submitting reader for session completeness in domain %s: %w", domainSID, err)
					}

					if err := operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, _ chan<- any) error {
						if localGroupCompleteness, err := adAnalysis.FetchLocalGroupCompleteness(tx, domainSID); err != nil {
							return err
						} else {
							mutex.Lock()
							stat.LocalGroupCompleteness = nan.Float64(localGroupCompleteness)
							mutex.Unlock()
							return nil
						}
					}); err != nil {
						return fmt.Errorf("failed while submitting reader for local group completeness in domain %s: %w", domainSID, err)
					}

					if err := operation.Done(); err != nil {
						return err
					}

					adStats = append(adStats, stat)
				}
			}

			// Get completeness lastly
			if userSessionCompleteness, err := adAnalysis.FetchUserSessionCompleteness(tx); err != nil {
				return err
			} else {
				aggregation.SessionCompleteness = float32(userSessionCompleteness)
			}

			if localGroupCompleteness, err := adAnalysis.FetchLocalGroupCompleteness(tx); err != nil {
				return err
			} else {
				aggregation.LocalGroupCompleteness = float32(localGroupCompleteness)
			}
		}

		return nil
	})
}

func azureGraphStats(ctx context.Context, db graph.Database) (model.AzureDataQualityStats, model.AzureDataQualityAggregation, error) {
	var (
		aggregation model.AzureDataQualityAggregation
		stats       = model.AzureDataQualityStats{}
		runID       string

		kinds = azure.NodeKinds()
	)

	if newUUID, err := uuid.NewV4(); err != nil {
		return stats, aggregation, fmt.Errorf("could not generate new UUID: %w", err)
	} else {
		runID = newUUID.String()
		aggregation.RunID = runID
	}

	err := db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		if tenants, err := ops.FetchNodes(tx.Nodes().Filterf(func() graph.Criteria {
			return query.Kind(query.Node(), azure.Tenant)
		})); err != nil {
			return err
		} else {
			for _, tenant := range tenants {
				if tenantObjectID, err := tenant.Properties.Get(common.ObjectID.String()).String(); err != nil {
					slog.ErrorContext(
						ctx,
						"Tenant node does not have a valid objectid property",
						slog.Uint64("tenant_id", uint64(tenant.ID)),
						attr.Error(err),
					)
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
							NumReaders: post.MaximumDatabaseParallelWorkers,
							NumWriters: 0,
						})
						mutex = &sync.Mutex{}
					)

					for _, kind := range kinds {
						innerKind := kind

						if innerKind == azure.Entity {
							continue
						}

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

								case azure.AutomationAccount:
									stat.AutomationAccounts = int(count)
									aggregation.AutomationAccounts += int(count)

								case azure.ContainerRegistry:
									stat.ContainerRegistries = int(count)
									aggregation.ContainerRegistries += int(count)

								case azure.FunctionApp:
									stat.FunctionApps = int(count)
									aggregation.FunctionApps += int(count)

								case azure.LogicApp:
									stat.LogicApps = int(count)
									aggregation.LogicApps += int(count)

								case azure.ManagedCluster:
									stat.ManagedClusters = int(count)
									aggregation.ManagedClusters += int(count)

								case azure.VMScaleSet:
									stat.VMScaleSets = int(count)
									aggregation.VMScaleSets += int(count)

								case azure.WebApp:
									stat.WebApps = int(count)
									aggregation.WebApps += int(count)

								case azure.Tenant:
									// Do nothing. Only AzureDataQualityAggregation stats have tenant stats and the tenants stats are handled in the outer tenant loop
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

func schemaEnvironmentCollected(node *graph.Node) bool {
	if !node.Properties.Exists(common.Collected.String()) {
		return false
	}

	collected, _ := node.Properties.Get(common.Collected.String()).Bool()
	if node.Kinds.ContainsOneOf(azure.Tenant, ad.Domain) {
		return collected
	}

	return true
}

func schemaEnvironmentIDCriteria(environmentKind graph.Kind, sourceKind graph.Kind, environmentID string) graph.Criteria {
	var criteria = []graph.Criteria{
		query.Equals(query.NodeProperty(graphschema.EnvironmentIDKey), environmentID),
	}

	if environmentKind.Is(ad.Domain) || sourceKind.Is(ad.Entity) {
		criteria = append(criteria, query.Equals(query.NodeProperty(ad.DomainSID.String()), environmentID))
	}

	if environmentKind.Is(azure.Tenant) || sourceKind.Is(azure.Entity) {
		criteria = append(criteria, query.Equals(query.NodeProperty(azure.TenantID.String()), environmentID))
	}

	return query.Or(criteria...)
}

func schemaSourceKinds(ctx context.Context, db DataQualityData, schemaEnvironments []model.SchemaEnvironment) (map[int32]graph.Kind, error) {
	var sourceKindIDs []int32

	for _, schemaEnvironment := range schemaEnvironments {
		sourceKindIDs = append(sourceKindIDs, schemaEnvironment.SourceKindId)
	}

	kinds, err := db.GetKindsByIDs(ctx, sourceKindIDs...)
	if err != nil {
		return nil, err
	}

	sourceKinds := make(map[int32]graph.Kind, len(kinds))
	for _, kind := range kinds {
		sourceKinds[kind.ID] = kind.ToKind()
	}

	return sourceKinds, nil
}

func schemaNodeKindMaps(ctx context.Context, db DataQualityData, schemaEnvironments []model.SchemaEnvironment) (map[int32]model.GraphSchemaNodeKinds, map[string]int32, error) {
	var (
		schemaEnvironmentIDs = make([]int32, 0, len(schemaEnvironments))
		nodeKindNames        []string
		seenNodeKindNames    = make(map[string]struct{})
	)

	for _, schemaEnvironment := range schemaEnvironments {
		schemaEnvironmentIDs = append(schemaEnvironmentIDs, schemaEnvironment.ID)
	}

	nodeKindsBySchemaEnvironmentID, err := db.GetGraphSchemaNodeKindsByEnvironmentIds(ctx, schemaEnvironmentIDs...)
	if err != nil {
		return nil, nil, err
	}

	for _, nodeKinds := range nodeKindsBySchemaEnvironmentID {
		for _, nodeKind := range nodeKinds {
			if _, seen := seenNodeKindNames[nodeKind.Name]; seen {
				continue
			}

			seenNodeKindNames[nodeKind.Name] = struct{}{}
			nodeKindNames = append(nodeKindNames, nodeKind.Name)
		}
	}

	if len(nodeKindNames) == 0 {
		return nodeKindsBySchemaEnvironmentID, map[string]int32{}, nil
	}

	kinds, err := db.GetKindsByNames(ctx, nodeKindNames...)
	if err != nil {
		return nil, nil, err
	}

	kindIDByName := make(map[string]int32, len(kinds))
	for _, kind := range kinds {
		kindIDByName[kind.Name] = kind.ID
	}

	return nodeKindsBySchemaEnvironmentID, kindIDByName, nil
}

func schemaGraphStats(ctx context.Context, db DataQualityData, graphDB graph.Database) (model.DataQualityStats, error) {
	var (
		runID string
		stats = model.DataQualityStats{}
	)

	if newUUID, err := uuid.NewV4(); err != nil {
		return stats, fmt.Errorf("could not generate new UUID: %w", err)
	} else {
		runID = newUUID.String()
	}

	schemaEnvironments, err := db.GetEnvironments(ctx)
	if err != nil {
		return stats, err
	} else if len(schemaEnvironments) == 0 {
		return stats, nil
	}

	sourceKinds, err := schemaSourceKinds(ctx, db, schemaEnvironments)
	if err != nil {
		return stats, err
	}

	nodeKindsBySchemaEnvironmentID, kindIDByName, err := schemaNodeKindMaps(ctx, db, schemaEnvironments)
	if err != nil {
		return stats, err
	}

	err = graphDB.ReadTransaction(ctx, func(tx graph.Transaction) error {
		for _, schemaEnvironment := range schemaEnvironments {
			var (
				environmentKind = graph.StringKind(schemaEnvironment.EnvironmentKindName)
				sourceKind      graph.Kind
				sourceKindFound bool
			)

			if sourceKind, sourceKindFound = sourceKinds[schemaEnvironment.SourceKindId]; !sourceKindFound {
				return fmt.Errorf("source kind %d not found for schema environment %d", schemaEnvironment.SourceKindId, schemaEnvironment.ID)
			}

			environmentNodes, err := ops.FetchNodes(tx.Nodes().Filterf(func() graph.Criteria {
				return query.Kind(query.Node(), environmentKind)
			}))
			if err != nil {
				return err
			}

			for _, environmentNode := range environmentNodes {
				if !schemaEnvironmentCollected(environmentNode) {
					continue
				}

				environmentID, err := environmentNode.Properties.Get(common.ObjectID.String()).String()
				if err != nil {
					slog.ErrorContext(
						ctx,
						"Schema environment node does not have a valid objectid property",
						slog.Uint64("environment_node_id", uint64(environmentNode.ID)),
						slog.String("environment_kind", schemaEnvironment.EnvironmentKindName),
						attr.Error(err),
					)
					continue
				}

				for _, nodeKind := range nodeKindsBySchemaEnvironmentID[schemaEnvironment.ID] {
					graphNodeKind := nodeKind.ToKind()
					if graphNodeKind.Is(sourceKind) {
						continue
					}

					kindID, found := kindIDByName[nodeKind.Name]
					if !found {
						return fmt.Errorf("kind ID not found for schema node kind %s", nodeKind.Name)
					}

					count, err := tx.Nodes().Filterf(func() graph.Criteria {
						return query.And(
							query.Kind(query.Node(), sourceKind),
							query.Kind(query.Node(), graphNodeKind),
							schemaEnvironmentIDCriteria(environmentKind, sourceKind, environmentID),
						)
					}).Count()
					if err != nil {
						return err
					}

					stats = append(stats, model.DataQualityStat{
						RunID:                   runID,
						SchemaExtensionID:       schemaEnvironment.SchemaExtensionId,
						SchemaEnvironmentKindID: schemaEnvironment.EnvironmentKindId,
						EnvironmentID:           environmentID,
						MetricType:              model.DataQualityMetricTypeNode,
						MetricName:              nodeKind.Name,
						MetricValue:             float64(count),
						KindID:                  null.Int32From(kindID),
					})
				}
			}
		}

		return nil
	})

	return stats, err
}

func SaveDataQuality(ctx context.Context, db DataQualityData, graphDB graph.Database) error {
	slog.InfoContext(
		ctx,
		"Started Data Quality Stats Collection",
		attr.Namespace("analysis"),
		attr.Function("SaveDataQuality"),
		attr.Scope("process"),
	)
	defer measure.ContextMeasure(
		ctx,
		slog.LevelInfo,
		"Completed Data Quality Stats Collection",
		attr.Namespace("analysis"),
		attr.Function("SaveDataQuality"),
		attr.Scope("process"),
	)()

	if stats, aggregation, err := adGraphStats(ctx, graphDB); err != nil {
		return fmt.Errorf("could not get active directory data quality stats: %w", err)
	} else if len(stats) > 0 {
		// We only want to save stats if there are stats to save
		if _, err := db.CreateADDataQualityStats(ctx, stats); err != nil {
			return fmt.Errorf("could not save active directory data quality stats: %w", err)
		} else if _, err := db.CreateADDataQualityAggregation(ctx, aggregation); err != nil {
			return fmt.Errorf("could not save active directory data quality aggregation: %w", err)
		}
	}

	if stats, aggregation, err := azureGraphStats(ctx, graphDB); err != nil {
		return fmt.Errorf("could not get azure data quality stats: %w", err)
	} else if len(stats) > 0 {
		// We only want to save stats if there are stats to save
		if _, err := db.CreateAzureDataQualityStats(ctx, stats); err != nil {
			return fmt.Errorf("could not save azure data quality stats: %w", err)
		} else if _, err := db.CreateAzureDataQualityAggregation(ctx, aggregation); err != nil {
			return fmt.Errorf("could not save azure data quality stats: %w", err)
		}
	}

	if stats, err := schemaGraphStats(ctx, db, graphDB); err != nil {
		return fmt.Errorf("could not get schema data quality stats: %w", err)
	} else if len(stats) > 0 {
		if _, err := db.CreateDataQualityStats(ctx, stats); err != nil {
			return fmt.Errorf("could not save schema data quality stats: %w", err)
		}
	}

	return nil
}
