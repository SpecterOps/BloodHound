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
	GetGraphSchemaExtensions(ctx context.Context, extensionFilters model.Filters, sortOrder model.Sort, skip, limit int) (model.GraphSchemaExtensions, int, error)
	GetGraphSchemaNodeKindsByExtensionId(ctx context.Context, extensionId int32) (model.GraphSchemaNodeKinds, error)
	GetGraphSchemaRelationshipKinds(ctx context.Context, filters model.Filters, sortOrder model.Sort, skip, limit int) (model.GraphSchemaRelationshipKinds, int, error)
	GetEnvironmentsByExtensionId(ctx context.Context, extensionId int32) ([]model.SchemaEnvironment, error)
	GetKindsByIDs(ctx context.Context, ids ...int32) ([]model.Kind, error)
}

type openGraphDataQualityAggregationKey struct {
	schemaExtensionID       int32
	schemaEnvironmentKindID int32
	metricType              model.DataQualityMetricType
	metricName              string
	kindID                  null.Int32
}

const openGraphDataQualityRelationshipsMetricName = "relationships"

// Source kinds are counted by their owning built-in extension, so OG stats skip them here.
func excludeSourceKindsFromOpenGraphNodeKinds(nodeKinds model.GraphSchemaNodeKinds, sourceKinds []model.Kind) model.GraphSchemaNodeKinds {
	var (
		result          = make(model.GraphSchemaNodeKinds, 0, len(nodeKinds))
		sourceKindNames = map[string]struct{}{}
	)

	for _, sourceKind := range sourceKinds {
		sourceKindNames[sourceKind.Name] = struct{}{}
	}

	for _, nodeKind := range nodeKinds {
		if _, isSourceKind := sourceKindNames[nodeKind.Name]; !isSourceKind {
			result = append(result, nodeKind)
		}
	}

	return result
}

// Schema environments define the source kind used to scope each OpenGraph environment kind.
func openGraphDataQualitySourceKinds(ctx context.Context, db DataQualityData, environments []model.SchemaEnvironment) (map[int32]model.Kind, []model.Kind, error) {
	var (
		sourceKindIDs              = make([]int32, 0, len(environments))
		sourceKindsByEnvironmentID = make(map[int32]model.Kind, len(environments))
	)

	for _, environment := range environments {
		sourceKindIDs = append(sourceKindIDs, environment.SourceKindId)
	}

	sourceKinds, err := db.GetKindsByIDs(ctx, sourceKindIDs...)
	if err != nil {
		return nil, nil, err
	}

	sourceKindsByID := make(map[int32]model.Kind, len(sourceKinds))
	for _, sourceKind := range sourceKinds {
		sourceKindsByID[sourceKind.ID] = sourceKind
	}

	for _, environment := range environments {
		sourceKind, ok := sourceKindsByID[environment.SourceKindId]
		if !ok {
			return nil, nil, fmt.Errorf("source kind %d not found for schema environment %d", environment.SourceKindId, environment.ID)
		}

		sourceKindsByEnvironmentID[environment.ID] = sourceKind
	}

	return sourceKindsByEnvironmentID, sourceKinds, nil
}

// Node kind counts exclude source kinds so shared AD/Azure kinds are not double-counted for OG.
func openGraphDataQualityNodeKinds(ctx context.Context, db DataQualityData, extensionID int32, sourceKinds []model.Kind) (model.GraphSchemaNodeKinds, error) {
	nodeKinds, err := db.GetGraphSchemaNodeKindsByExtensionId(ctx, extensionID)
	if err != nil {
		return nil, err
	}

	return excludeSourceKindsFromOpenGraphNodeKinds(nodeKinds, sourceKinds), nil
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

// OG environment nodes identify the environment IDs to count; stats remain scoped to those IDs.
func fetchOpenGraphEnvironmentIDs(ctx context.Context, tx graph.Transaction, environmentKind graph.Kind) ([]string, error) {
	var environmentIDs []string

	if err := tx.Nodes().Filterf(func() graph.Criteria {
		return query.And(
			query.Kind(query.Node(), environmentKind),
			query.Equals(query.NodeProperty(common.Collected.String()), true),
		)
	}).Fetch(func(cursor graph.Cursor[*graph.Node]) error {
		for node := range cursor.Chan() {
			if environmentID, err := node.Properties.Get(graphschema.EnvironmentIDKey).String(); err != nil {
				if graph.IsErrPropertyNotFound(err) {
					slog.WarnContext(ctx, "Skipping OpenGraph environment node missing environment id property",
						slog.Uint64("node_id", node.ID.Uint64()),
						slog.String("environment_kind", environmentKind.String()),
						slog.String("property", graphschema.EnvironmentIDKey),
					)
					continue
				}

				return fmt.Errorf("failed to get %s from node %d: %w", graphschema.EnvironmentIDKey, node.ID, err)
			} else {
				environmentIDs = append(environmentIDs, environmentID)
			}
		}

		return cursor.Error()
	}); err != nil {
		return nil, err
	}

	return environmentIDs, nil
}

func openGraphNodeEnvironmentIDFilter(environmentID string) graph.Criteria {
	return query.Equals(query.NodeProperty(graphschema.EnvironmentIDKey), environmentID)
}

func openGraphRelationshipEnvironmentIDFilter(environmentID string) graph.Criteria {
	return query.Or(
		query.Equals(query.StartProperty(graphschema.EnvironmentIDKey), environmentID),
		query.Equals(query.EndProperty(graphschema.EnvironmentIDKey), environmentID),
	)
}

func openGraphExtensionGraphStats(
	ctx context.Context,
	graphDB graph.Database,
	runID string,
	extension model.GraphSchemaExtension,
	environments []model.SchemaEnvironment,
	nodeKinds model.GraphSchemaNodeKinds,
	relationshipKinds model.GraphSchemaRelationshipKinds,
) (model.DataQualityStats, model.DataQualityAggregations, error) {
	var (
		relationshipGraphKinds = make(graph.Kinds, 0, len(relationshipKinds))
		stats                  = model.DataQualityStats{}
		// Aggregations roll up all environment IDs for the same schema environment kind and metric.
		aggregationCounts = map[openGraphDataQualityAggregationKey]float64{}
	)

	for _, relationshipKind := range relationshipKinds {
		relationshipGraphKinds = append(relationshipGraphKinds, relationshipKind.ToKind())
	}

	if err := graphDB.ReadTransaction(ctx, func(tx graph.Transaction) error {
		for _, environment := range environments {
			environmentKind := graph.StringKind(environment.EnvironmentKindName)
			environmentIDs, err := fetchOpenGraphEnvironmentIDs(ctx, tx, environmentKind)
			if err != nil {
				return fmt.Errorf("failed to fetch OpenGraph environment IDs for environment kind %s: %w", environmentKind.String(), err)
			}

			for _, environmentID := range environmentIDs {
				for _, nodeKind := range nodeKinds {
					count, err := tx.Nodes().Filter(query.And(
						graphschema.IgnoreMetaFilter,
						query.Kind(query.Node(), nodeKind.ToKind()),
						openGraphNodeEnvironmentIDFilter(environmentID),
					)).Count()
					if err != nil {
						return fmt.Errorf("failed to count OpenGraph nodes of type %s in environment %s: %w", nodeKind.Name, environmentID, err)
					}

					stat := model.DataQualityStat{
						RunID:                   runID,
						SchemaExtensionID:       extension.ID,
						SchemaEnvironmentKindID: environment.EnvironmentKindId,
						EnvironmentID:           environmentID,
						MetricType:              model.DataQualityMetricTypeNode,
						MetricName:              nodeKind.Name,
						MetricValue:             float64(count),
						KindID:                  null.Int32From(nodeKind.KindId),
					}
					aggregationKey := openGraphDataQualityAggregationKey{
						schemaExtensionID:       extension.ID,
						schemaEnvironmentKindID: environment.EnvironmentKindId,
						metricType:              model.DataQualityMetricTypeNode,
						metricName:              nodeKind.Name,
						kindID:                  null.Int32From(nodeKind.KindId),
					}

					stats = append(stats, stat)
					aggregationCounts[aggregationKey] += float64(count)
				}

				if len(relationshipGraphKinds) > 0 {
					count, err := tx.Relationships().Filter(query.And(
						query.KindIn(query.Relationship(), relationshipGraphKinds...),
						openGraphRelationshipEnvironmentIDFilter(environmentID),
					)).Count()
					if err != nil {
						return fmt.Errorf("failed to count OpenGraph relationships in environment %s: %w", environmentID, err)
					}

					stat := model.DataQualityStat{
						RunID:                   runID,
						SchemaExtensionID:       extension.ID,
						SchemaEnvironmentKindID: environment.EnvironmentKindId,
						EnvironmentID:           environmentID,
						MetricType:              model.DataQualityMetricTypeRelationship,
						MetricName:              openGraphDataQualityRelationshipsMetricName,
						MetricValue:             float64(count),
					}
					aggregationKey := openGraphDataQualityAggregationKey{
						schemaExtensionID:       extension.ID,
						schemaEnvironmentKindID: environment.EnvironmentKindId,
						metricType:              model.DataQualityMetricTypeRelationship,
						metricName:              openGraphDataQualityRelationshipsMetricName,
					}

					stats = append(stats, stat)
					aggregationCounts[aggregationKey] += float64(count)
				}
			}
		}

		return nil
	}); err != nil {
		return stats, nil, err
	}

	aggregations := make(model.DataQualityAggregations, 0, len(aggregationCounts))
	for aggregationKey, metricValue := range aggregationCounts {
		aggregations = append(aggregations, model.DataQualityAggregation{
			RunID:                   runID,
			SchemaExtensionID:       aggregationKey.schemaExtensionID,
			SchemaEnvironmentKindID: aggregationKey.schemaEnvironmentKindID,
			MetricType:              aggregationKey.metricType,
			MetricName:              aggregationKey.metricName,
			MetricValue:             metricValue,
			KindID:                  aggregationKey.kindID,
		})
	}

	return stats, aggregations, nil
}

// Collect OpenGraph stats using schema environments to scope environment IDs.
func openGraphGraphStats(ctx context.Context, db DataQualityData, graphDB graph.Database) (model.DataQualityStats, model.DataQualityAggregations, error) {
	var (
		aggregations = model.DataQualityAggregations{}
		stats        = model.DataQualityStats{}
		runID        string
	)

	if newUUID, err := uuid.NewV4(); err != nil {
		return stats, aggregations, fmt.Errorf("could not generate new UUID: %w", err)
	} else {
		runID = newUUID.String()
	}

	extensions, _, err := db.GetGraphSchemaExtensions(ctx, model.Filters{"is_builtin": []model.Filter{{Operator: model.Equals, Value: "false"}}}, model.Sort{}, 0, 0)
	if err != nil {
		return stats, aggregations, fmt.Errorf("failed to get OpenGraph extensions: %w", err)
	}

	for _, extension := range extensions {
		environments, err := db.GetEnvironmentsByExtensionId(ctx, extension.ID)
		if err != nil {
			return stats, aggregations, fmt.Errorf("failed to get environments for OpenGraph extension %d: %w", extension.ID, err)
		} else if len(environments) == 0 {
			continue
		}

		_, sourceKinds, err := openGraphDataQualitySourceKinds(ctx, db, environments)
		if err != nil {
			return stats, aggregations, fmt.Errorf("failed to get source kinds for OpenGraph extension %d: %w", extension.ID, err)
		}

		nodeKinds, err := openGraphDataQualityNodeKinds(ctx, db, extension.ID, sourceKinds)
		if err != nil {
			return stats, aggregations, fmt.Errorf("failed to get node kinds for OpenGraph extension %d: %w", extension.ID, err)
		}

		relationshipKinds, _, err := db.GetGraphSchemaRelationshipKinds(ctx, model.Filters{"schema_extension_id": []model.Filter{{
			Operator:    model.Equals,
			Value:       fmt.Sprintf("%d", extension.ID),
			SetOperator: model.FilterAnd,
		}}}, model.Sort{}, 0, 0)
		if err != nil {
			return stats, aggregations, fmt.Errorf("failed to get relationship kinds for OpenGraph extension %d: %w", extension.ID, err)
		}

		if len(nodeKinds) == 0 && len(relationshipKinds) == 0 {
			continue
		}

		extensionStats, extensionAggregations, err := openGraphExtensionGraphStats(ctx, graphDB, runID, extension, environments, nodeKinds, relationshipKinds)
		if err != nil {
			return stats, aggregations, fmt.Errorf("failed to get OpenGraph data quality stats for extension %d: %w", extension.ID, err)
		}

		stats = append(stats, extensionStats...)
		aggregations = append(aggregations, extensionAggregations...)
	}

	return stats, aggregations, nil
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

	// OpenGraph runs after built-in DQ collection so shared source-kind nodes are already attributed.
	if stats, aggregations, err := openGraphGraphStats(ctx, db, graphDB); err != nil {
		return fmt.Errorf("could not get OpenGraph data quality stats: %w", err)
	} else if len(stats) > 0 {
		// We only want to save stats if there are stats to save
		if _, err := db.CreateOpenGraphDataQualityStats(ctx, stats); err != nil {
			return fmt.Errorf("could not save OpenGraph data quality stats: %w", err)
		} else if _, err := db.CreateOpenGraphDataQualityAggregations(ctx, aggregations); err != nil {
			return fmt.Errorf("could not save OpenGraph data quality aggregations: %w", err)
		}
	}

	return nil
}
