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
	"sort"
	"sync"

	"github.com/gofrs/uuid"
	"github.com/specterops/bloodhound/cmd/api/src/database"
	"github.com/specterops/bloodhound/cmd/api/src/database/types/nan"
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

type dataQualitySourceObjectCountKey struct {
	SourceKind string
	NodeKind   string
}

type dataQualityEnvironmentObjectCountKey struct {
	SourceKind      string
	EnvironmentKind string
	EnvironmentID   string
	NodeKind        string
}

type dataQualityEnvironmentContextKey struct {
	SourceKind      string
	EnvironmentKind string
}

type dataQualityEnvironmentContext struct {
	SourceKind            string
	EnvironmentKind       string
	EnvironmentGraphKind  graph.Kind
	EnvironmentIDProperty string
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

func graphObjectCounts(ctx context.Context, db database.DataQualityData, graphDB graph.Database, runID string) (model.DataQualitySourceObjectCounts, model.DataQualityEnvironmentObjectCounts, error) {
	var (
		environmentContexts         []dataQualityEnvironmentContext
		environmentContextsBySource = map[string][]dataQualityEnvironmentContext{}
		environmentIDsByContext     = map[dataQualityEnvironmentContextKey]map[string]struct{}{}
		environmentObjectCountByKey = map[dataQualityEnvironmentObjectCountKey]int64{}
		environmentObjectCounts     model.DataQualityEnvironmentObjectCounts
		environmentSourceKindIDs    []int32
		environments                []model.SchemaEnvironment
		primaryDisplayKinds         graphschema.PrimaryDisplayKinds
		sortedEnvironmentKeys       []dataQualityEnvironmentObjectCountKey
		sortedSourceKeys            []dataQualitySourceObjectCountKey
		sortedSourceKindNames       []string
		sourceKindKinds             []model.Kind
		sourceKindGraphKinds        = map[string]graph.Kind{}
		sourceKindNameByKindID      map[int32]string
		sourceKindNames             map[string]struct{}
		sourceKinds                 []model.SourceKind
		sourceObjectCountByKey      = map[dataQualitySourceObjectCountKey]int64{}
		sourceObjectCounts          model.DataQualitySourceObjectCounts
	)

	if fetchedSourceKinds, err := db.GetSourceKinds(ctx); err != nil {
		return sourceObjectCounts, environmentObjectCounts, fmt.Errorf("could not fetch source kinds: %w", err)
	} else {
		sourceKinds = fetchedSourceKinds
	}

	if len(sourceKinds) == 0 {
		return sourceObjectCounts, environmentObjectCounts, nil
	}

	if fetchedPrimaryDisplayKinds, err := db.GetPrimaryDisplayKinds(ctx); err != nil {
		return sourceObjectCounts, environmentObjectCounts, fmt.Errorf("could not fetch primary display kinds: %w", err)
	} else {
		primaryDisplayKinds = mergeDataQualityPrimaryDisplayKinds(fetchedPrimaryDisplayKinds)
	}

	if fetchedEnvironments, err := db.GetEnvironments(ctx); err != nil {
		return sourceObjectCounts, environmentObjectCounts, fmt.Errorf("could not fetch schema environments: %w", err)
	} else {
		environments = fetchedEnvironments
	}

	environmentSourceKindIDs = dataQualityEnvironmentSourceKindIDs(environments)
	if fetchedSourceKindKinds, err := db.GetKindsByIDs(ctx, environmentSourceKindIDs...); err != nil {
		return sourceObjectCounts, environmentObjectCounts, fmt.Errorf("could not fetch schema environment source kinds: %w", err)
	} else {
		sourceKindKinds = fetchedSourceKindKinds
	}

	sourceKindNames = dataQualitySourceKindNames(sourceKinds)
	sourceKindNameByKindID = dataQualityKindNameByID(sourceKindKinds)

	for _, sourceKind := range sourceKinds {
		if sourceGraphKind := sourceKind.ToKind(); sourceGraphKind != graph.EmptyKind {
			sourceKindGraphKinds[sourceKind.Name] = sourceGraphKind
			sortedSourceKindNames = append(sortedSourceKindNames, sourceKind.Name)
		}
	}

	sort.Strings(sortedSourceKindNames)

	environmentContexts = dataQualityEnvironmentContexts(environments, sourceKindNameByKindID)
	for _, environmentContext := range environmentContexts {
		contextKey := dataQualityEnvironmentContextKey{
			SourceKind:      environmentContext.SourceKind,
			EnvironmentKind: environmentContext.EnvironmentKind,
		}

		environmentContextsBySource[environmentContext.SourceKind] = append(environmentContextsBySource[environmentContext.SourceKind], environmentContext)
		environmentIDsByContext[contextKey] = map[string]struct{}{}
	}

	if err := graphDB.ReadTransaction(ctx, func(tx graph.Transaction) error {
		for _, environmentContext := range environmentContexts {
			contextKey := dataQualityEnvironmentContextKey{
				SourceKind:      environmentContext.SourceKind,
				EnvironmentKind: environmentContext.EnvironmentKind,
			}

			if err := tx.Nodes().Filterf(func() graph.Criteria {
				return query.And(
					graphschema.IgnoreMetaFilter,
					query.Kind(query.Node(), environmentContext.EnvironmentGraphKind),
				)
			}).Fetch(func(cursor graph.Cursor[*graph.Node]) error {
				for environmentNode := range cursor.Chan() {
					if environmentID, ok := dataQualityNodeStringProperty(environmentNode, environmentContext.EnvironmentIDProperty); ok {
						environmentIDsByContext[contextKey][environmentID] = struct{}{}
					}
				}

				return cursor.Error()
			}); err != nil {
				return fmt.Errorf("failed fetching environments for source kind %s and environment kind %s: %w", environmentContext.SourceKind, environmentContext.EnvironmentKind, err)
			}
		}

		for _, sourceKindName := range sortedSourceKindNames {
			sourceGraphKind := sourceKindGraphKinds[sourceKindName]
			sourceEnvironmentContexts := environmentContextsBySource[sourceKindName]

			if err := tx.Nodes().Filterf(func() graph.Criteria {
				return query.And(
					graphschema.IgnoreMetaFilter,
					query.Kind(query.Node(), sourceGraphKind),
				)
			}).Fetch(func(cursor graph.Cursor[*graph.Node]) error {
				for node := range cursor.Chan() {
					objectNodeKind := selectDataQualityObjectNodeKind(primaryDisplayKinds, sourceGraphKind, sourceKindNames, node.Kinds)
					sourceObjectCountByKey[dataQualitySourceObjectCountKey{
						SourceKind: sourceKindName,
						NodeKind:   objectNodeKind.String(),
					}]++

					for _, environmentContext := range sourceEnvironmentContexts {
						contextKey := dataQualityEnvironmentContextKey{
							SourceKind:      environmentContext.SourceKind,
							EnvironmentKind: environmentContext.EnvironmentKind,
						}

						if environmentID, ok := dataQualityNodeStringProperty(node, environmentContext.EnvironmentIDProperty); !ok {
							continue
						} else if _, ok := environmentIDsByContext[contextKey][environmentID]; !ok {
							continue
						} else {
							environmentObjectCountByKey[dataQualityEnvironmentObjectCountKey{
								SourceKind:      sourceKindName,
								EnvironmentKind: environmentContext.EnvironmentKind,
								EnvironmentID:   environmentID,
								NodeKind:        objectNodeKind.String(),
							}]++
						}
					}
				}

				return cursor.Error()
			}); err != nil {
				return fmt.Errorf("failed counting object kinds for source kind %s: %w", sourceKindName, err)
			}
		}

		return nil
	}); err != nil {
		return sourceObjectCounts, environmentObjectCounts, err
	}

	sortedSourceKeys = make([]dataQualitySourceObjectCountKey, 0, len(sourceObjectCountByKey))
	for sourceObjectCountKey := range sourceObjectCountByKey {
		sortedSourceKeys = append(sortedSourceKeys, sourceObjectCountKey)
	}

	sort.Slice(sortedSourceKeys, func(leftIndex, rightIndex int) bool {
		leftKey := sortedSourceKeys[leftIndex]
		rightKey := sortedSourceKeys[rightIndex]

		if leftKey.SourceKind == rightKey.SourceKind {
			return leftKey.NodeKind < rightKey.NodeKind
		}

		return leftKey.SourceKind < rightKey.SourceKind
	})

	sourceObjectCounts = make(model.DataQualitySourceObjectCounts, 0, len(sortedSourceKeys))
	for _, sourceObjectCountKey := range sortedSourceKeys {
		sourceObjectCounts = append(sourceObjectCounts, model.DataQualitySourceObjectCount{
			SourceKind: sourceObjectCountKey.SourceKind,
			NodeKind:   sourceObjectCountKey.NodeKind,
			Count:      sourceObjectCountByKey[sourceObjectCountKey],
			RunID:      runID,
		})
	}

	sortedEnvironmentKeys = make([]dataQualityEnvironmentObjectCountKey, 0, len(environmentObjectCountByKey))
	for environmentObjectCountKey := range environmentObjectCountByKey {
		sortedEnvironmentKeys = append(sortedEnvironmentKeys, environmentObjectCountKey)
	}

	sort.Slice(sortedEnvironmentKeys, func(leftIndex, rightIndex int) bool {
		leftKey := sortedEnvironmentKeys[leftIndex]
		rightKey := sortedEnvironmentKeys[rightIndex]

		if leftKey.SourceKind != rightKey.SourceKind {
			return leftKey.SourceKind < rightKey.SourceKind
		}

		if leftKey.EnvironmentKind != rightKey.EnvironmentKind {
			return leftKey.EnvironmentKind < rightKey.EnvironmentKind
		}

		if leftKey.EnvironmentID != rightKey.EnvironmentID {
			return leftKey.EnvironmentID < rightKey.EnvironmentID
		}

		return leftKey.NodeKind < rightKey.NodeKind
	})

	environmentObjectCounts = make(model.DataQualityEnvironmentObjectCounts, 0, len(sortedEnvironmentKeys))
	for _, environmentObjectCountKey := range sortedEnvironmentKeys {
		environmentObjectCounts = append(environmentObjectCounts, model.DataQualityEnvironmentObjectCount{
			SourceKind:      environmentObjectCountKey.SourceKind,
			EnvironmentKind: environmentObjectCountKey.EnvironmentKind,
			EnvironmentID:   environmentObjectCountKey.EnvironmentID,
			NodeKind:        environmentObjectCountKey.NodeKind,
			Count:           environmentObjectCountByKey[environmentObjectCountKey],
			RunID:           runID,
		})
	}

	return sourceObjectCounts, environmentObjectCounts, nil
}

func mergeDataQualityPrimaryDisplayKinds(primaryDisplayKinds graphschema.PrimaryDisplayKinds) graphschema.PrimaryDisplayKinds {
	var (
		mergedPrimaryDisplayKinds = make(graphschema.PrimaryDisplayKinds, len(graphschema.ValidKinds)+len(primaryDisplayKinds))
	)

	for kind, displayKind := range graphschema.ValidKinds {
		mergedPrimaryDisplayKinds[kind] = displayKind
	}

	for kind, displayKind := range primaryDisplayKinds {
		mergedPrimaryDisplayKinds[kind] = displayKind
	}

	return mergedPrimaryDisplayKinds
}

func dataQualitySourceKindNames(sourceKinds []model.SourceKind) map[string]struct{} {
	var (
		sourceKindNames = make(map[string]struct{}, len(sourceKinds))
	)

	for _, sourceKind := range sourceKinds {
		sourceKindNames[sourceKind.Name] = struct{}{}
	}

	return sourceKindNames
}

func dataQualityEnvironmentSourceKindIDs(environments []model.SchemaEnvironment) []int32 {
	var (
		sourceKindIDs []int32
		seenIDs       = map[int32]struct{}{}
	)

	for _, environment := range environments {
		if environment.SourceKindId == 0 {
			continue
		}

		if _, seen := seenIDs[environment.SourceKindId]; seen {
			continue
		}

		seenIDs[environment.SourceKindId] = struct{}{}
		sourceKindIDs = append(sourceKindIDs, environment.SourceKindId)
	}

	sort.Slice(sourceKindIDs, func(leftIndex, rightIndex int) bool {
		return sourceKindIDs[leftIndex] < sourceKindIDs[rightIndex]
	})

	return sourceKindIDs
}

func dataQualityKindNameByID(kinds []model.Kind) map[int32]string {
	var (
		kindNameByID = make(map[int32]string, len(kinds))
	)

	for _, kind := range kinds {
		kindNameByID[kind.ID] = kind.Name
	}

	return kindNameByID
}

func dataQualityEnvironmentContexts(environments []model.SchemaEnvironment, sourceKindNameByID map[int32]string) []dataQualityEnvironmentContext {
	var (
		environmentContexts []dataQualityEnvironmentContext
	)

	for _, environment := range environments {
		sourceKindName, hasSourceKind := sourceKindNameByID[environment.SourceKindId]
		if !hasSourceKind {
			continue
		}

		environmentGraphKind := graph.StringKind(environment.EnvironmentKindName)
		environmentContexts = append(environmentContexts, dataQualityEnvironmentContext{
			SourceKind:            sourceKindName,
			EnvironmentKind:       environment.EnvironmentKindName,
			EnvironmentGraphKind:  environmentGraphKind,
			EnvironmentIDProperty: dataQualityEnvironmentIDProperty(environmentGraphKind),
		})
	}

	sort.Slice(environmentContexts, func(leftIndex, rightIndex int) bool {
		leftContext := environmentContexts[leftIndex]
		rightContext := environmentContexts[rightIndex]

		if leftContext.SourceKind == rightContext.SourceKind {
			return leftContext.EnvironmentKind < rightContext.EnvironmentKind
		}

		return leftContext.SourceKind < rightContext.SourceKind
	})

	return environmentContexts
}

func dataQualityEnvironmentIDProperty(environmentKind graph.Kind) string {
	switch environmentKind {
	case ad.Domain:
		return ad.DomainSID.String()
	case azure.Tenant:
		return azure.TenantID.String()
	default:
		return graphschema.EnvironmentIDKey
	}
}

func dataQualityNodeStringProperty(node *graph.Node, property string) (string, bool) {
	if value, err := node.Properties.Get(property).String(); err != nil || value == "" {
		return "", false
	} else {
		return value, true
	}
}

func selectDataQualityObjectNodeKind(primaryDisplayKinds graphschema.PrimaryDisplayKinds, sourceKind graph.Kind, sourceKindNames map[string]struct{}, nodeKinds graph.Kinds) graph.Kind {
	var (
		fallbackKind graph.Kind
	)

	for _, nodeKind := range nodeKinds {
		if !isDataQualityObjectNodeKind(sourceKind, sourceKindNames, nodeKind) {
			continue
		}

		if _, ok := primaryDisplayKinds[nodeKind]; ok {
			return nodeKind
		} else if fallbackKind == nil {
			fallbackKind = nodeKind
		}
	}

	if fallbackKind != nil {
		return fallbackKind
	}

	return sourceKind
}

func isDataQualityObjectNodeKind(sourceKind graph.Kind, sourceKindNames map[string]struct{}, nodeKind graph.Kind) bool {
	if nodeKind == nil || nodeKind == graph.EmptyKind || nodeKind.Is(sourceKind) || model.IsExtendedNodeKind(nodeKind) {
		return false
	}

	_, isSourceKind := sourceKindNames[nodeKind.String()]
	return !isSourceKind
}

func SaveDataQuality(ctx context.Context, db database.DataQualityData, graphDB graph.Database) error {
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

	if newUUID, err := uuid.NewV4(); err != nil {
		return fmt.Errorf("could not generate new UUID for data quality object counts: %w", err)
	} else if sourceObjectCounts, environmentObjectCounts, err := graphObjectCounts(ctx, db, graphDB, newUUID.String()); err != nil {
		return fmt.Errorf("could not get data quality object counts: %w", err)
	} else {
		if len(sourceObjectCounts) > 0 {
			if _, err := db.CreateDataQualitySourceObjectCounts(ctx, sourceObjectCounts); err != nil {
				return fmt.Errorf("could not save data quality source object counts: %w", err)
			}
		}

		if len(environmentObjectCounts) > 0 {
			if _, err := db.CreateDataQualityEnvironmentObjectCounts(ctx, environmentObjectCounts); err != nil {
				return fmt.Errorf("could not save data quality environment object counts: %w", err)
			}
		}

		if _, err := db.CreateDataQualityObjectCountRun(ctx, model.DataQualityObjectCountRun{RunID: newUUID.String()}); err != nil {
			return fmt.Errorf("could not save data quality object count run: %w", err)
		}
	}

	return nil
}
