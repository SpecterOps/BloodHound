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
	"errors"
	"fmt"
	"log/slog"
	"sync"

	"github.com/gofrs/uuid"
	"github.com/specterops/bloodhound/cmd/api/src/database"
	"github.com/specterops/bloodhound/cmd/api/src/database/types/nan"
	"github.com/specterops/bloodhound/cmd/api/src/database/types/null"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/cmd/api/src/model/appcfg"
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

func SaveDataQuality(ctx context.Context, db database.Database, graphDB graph.Database) error {
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

	// OpenGraph node counts
	if openGraphExtensionManagementFlag, err := db.GetFlagByKey(ctx, appcfg.FeatureOpenGraphExtensionManagement); err != nil {
		return fmt.Errorf("could not get open graph extension management feature flag: %w", err)
	} else if openGraphExtensionManagementFlag.Enabled {
		if stats, aggregations, err := openGraphStats(ctx, db, graphDB); err != nil {
			return fmt.Errorf("could not get open graph stats: %w", err)
		} else {
			if len(stats) == 0 {
				return nil
			}

			if _, err := db.CreateDataQualityStats(ctx, stats); err != nil {
				return fmt.Errorf("could not save open graph stats: %w", err)
			} else if _, err := db.CreateDataQualityAggregations(ctx, aggregations); err != nil {
				return fmt.Errorf("could not save open graph aggregations: %w", err)
			}
		}
	}

	return nil
}

func openGraphStats(ctx context.Context, db database.Database, graphDB graph.Database) (model.DataQualityStats, model.DataQualityAggregations, error) {
	var (
		aggregations = make(model.DataQualityAggregations, 0)
		stats        = make(model.DataQualityStats, 0)

		// We only are concerned with non-builtin types since AD and AZ have their own stat collection
		// When we move AD and AZ stat collection to this common table, this will need to be changed
		builtInFilter = model.Filters{"is_builtin": []model.Filter{{Operator: model.Equals, Value: "false"}}}
	)

	runID, err := uuid.NewV4()
	if err != nil {
		return stats, aggregations, fmt.Errorf("could not generate new UUID: %w", err)
	}

	extensions, _, err := db.GetGraphSchemaExtensions(ctx, builtInFilter, model.Sort{}, 0, 0)
	if err != nil {
		return stats, aggregations, fmt.Errorf("could not get graph schema extensions: %w", err)
	}

	for _, extension := range extensions {
		environments, err := db.GetEnvironmentsByExtensionId(ctx, extension.ID)
		if err != nil {
			return stats, aggregations, fmt.Errorf("could not get environments for extension %d: %w", extension.ID, err)
		}

		// Each extension may define multiple environments with their own environment kind, principal kinds, and source kind.
		for _, environment := range environments {
			environmentSourceKind, err := getOpenGraphSourceKind(ctx, db, environment)
			if err != nil {
				return stats, aggregations, err
			}

			nodeKindCounts, err := countOpenGraphNodesByEnvironment(ctx, graphDB, graph.StringKind(environment.EnvironmentKindName), environmentSourceKind.ToKind())
			if err != nil {
				return stats, aggregations, fmt.Errorf("could not count open graph nodes for environment kind %s: %w", environment.EnvironmentKindName, err)
			}

			kindIDs, err := getOpenGraphCountKindIDs(ctx, db, nodeKindCounts)
			if err != nil {
				return stats, aggregations, err
			}

			environmentStats, environmentAggregations := buildOpenGraphNodeCountStats(runID.String(), extension, environment, nodeKindCounts, kindIDs)
			stats = append(stats, environmentStats...)
			aggregations = append(aggregations, environmentAggregations...)
		}
	}

	return stats, aggregations, nil
}

func getOpenGraphSourceKind(ctx context.Context, db database.Database, environment model.SchemaEnvironment) (model.Kind, error) {
	kinds, err := db.GetKindsByIDs(ctx, environment.SourceKindId)
	if err != nil {
		return model.Kind{}, fmt.Errorf("could not get source kind for schema environment %d: %w", environment.ID, err)
	}

	if len(kinds) == 0 {
		return model.Kind{}, fmt.Errorf("could not get source kind for schema environment %d: %w", environment.ID, database.ErrNotFound)
	}

	return kinds[0], nil
}

func countOpenGraphNodesByEnvironment(ctx context.Context, graphDB graph.Database, environmentKind graph.Kind, sourceKind graph.Kind) (map[string]map[string]int, error) {
	nodeKindCountsByEnvironmentID := map[string]map[string]int{}

	err := graphDB.ReadTransaction(ctx, func(tx graph.Transaction) error {
		environmentIDs, err := getOpenGraphEnvironmentIDs(ctx, tx, environmentKind)
		if err != nil {
			return err
		}

		for _, environmentID := range environmentIDs {
			nodeKindCounts := map[string]int{}

			if err := tx.Nodes().Filterf(func() graph.Criteria {
				return query.And(
					query.Kind(query.Node(), sourceKind),
					query.Equals(query.NodeProperty(graphschema.EnvironmentIDKey), environmentID),
				)
			}).FetchKinds(func(cursor graph.Cursor[graph.KindsResult]) error {
				for result := range cursor.Chan() {
					countOpenGraphNodeKinds(result.Kinds, sourceKind, nodeKindCounts)
				}

				return cursor.Error()
			}); err != nil {
				return err
			}

			if len(nodeKindCounts) > 0 {
				nodeKindCountsByEnvironmentID[environmentID] = nodeKindCounts
			}
		}

		return nil
	})

	return nodeKindCountsByEnvironmentID, err
}

func countOpenGraphNodeKinds(kinds graph.Kinds, sourceKind graph.Kind, nodeKindCounts map[string]int) {
	for _, kind := range kinds {
		if kind.Is(sourceKind) || model.IsExtendedNodeKind(kind) {
			continue
		}

		nodeKindCounts[kind.String()]++
	}
}

func getOpenGraphEnvironmentIDs(ctx context.Context, tx graph.Transaction, environmentKind graph.Kind) ([]string, error) {
	var environmentIDs []string

	environments, err := ops.FetchNodes(tx.Nodes().Filterf(func() graph.Criteria {
		return query.Kind(query.Node(), environmentKind)
	}))
	if err != nil {
		return nil, err
	}

	for _, environment := range environments {
		environmentID, err := environment.Properties.Get(graphschema.EnvironmentIDKey).String()
		if err != nil {
			slog.WarnContext(
				ctx,
				"OpenGraph environment node does not have a valid environment ID property",
				slog.Uint64("node_id", uint64(environment.ID)),
				slog.String("environment_kind", environmentKind.String()),
				attr.Error(err),
			)
			continue
		}

		environmentIDs = append(environmentIDs, environmentID)
	}

	return environmentIDs, nil
}

func getOpenGraphCountKindIDs(ctx context.Context, db database.Database, countsByEnvironmentID map[string]map[string]int) (map[string]null.Int32, error) {
	kindNames := make([]string, 0)
	seenKindNames := map[string]struct{}{}

	for _, nodeKindCounts := range countsByEnvironmentID {
		for kindName := range nodeKindCounts {
			if _, seen := seenKindNames[kindName]; !seen {
				kindNames = append(kindNames, kindName)
				seenKindNames[kindName] = struct{}{}
			}
		}
	}

	if len(kindNames) == 0 {
		return map[string]null.Int32{}, nil
	}

	kinds, err := db.GetKindsByNames(ctx, kindNames...)
	if err != nil && !errors.Is(err, database.ErrNotFound) {
		return nil, fmt.Errorf("could not get data quality metric kinds: %w", err)
	}

	kindIDsByName := make(map[string]null.Int32, len(kinds))
	for _, kind := range kinds {
		kindIDsByName[kind.Name] = null.Int32From(kind.ID)
	}

	return kindIDsByName, nil
}

func buildOpenGraphNodeCountStats(runID string, extension model.GraphSchemaExtension, environment model.SchemaEnvironment, countsByEnvironmentID map[string]map[string]int, kindIDsByName map[string]null.Int32) (model.DataQualityStats, model.DataQualityAggregations) {
	var (
		aggregatedCounts = map[string]int{}
		stats            = make(model.DataQualityStats, 0)
		aggregations     = make(model.DataQualityAggregations, 0)
	)

	for environmentID, kindCounts := range countsByEnvironmentID {
		for kindName, count := range kindCounts {
			stats = append(stats, model.DataQualityStat{
				RunID:                   runID,
				SchemaExtensionID:       extension.ID,
				SchemaEnvironmentKindID: environment.EnvironmentKindId,
				EnvironmentID:           environmentID,
				MetricType:              model.DataQualityMetricTypeNode,
				MetricName:              kindName,
				MetricValue:             float64(count),
				KindID:                  kindIDsByName[kindName],
			})

			aggregatedCounts[kindName] += count
		}
	}

	for kindName, count := range aggregatedCounts {
		aggregations = append(aggregations, model.DataQualityAggregation{
			RunID:                   runID,
			SchemaExtensionID:       extension.ID,
			SchemaEnvironmentKindID: environment.EnvironmentKindId,
			MetricType:              model.DataQualityMetricTypeNode,
			MetricName:              kindName,
			MetricValue:             float64(count),
			KindID:                  kindIDsByName[kindName],
		})
	}

	return stats, aggregations
}
