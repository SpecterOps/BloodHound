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
	"sync"

	"github.com/gofrs/uuid"
	"github.com/specterops/bloodhound/analysis"
	adAnalysis "github.com/specterops/bloodhound/analysis/ad"
	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/dawgs/ops"
	"github.com/specterops/bloodhound/dawgs/query"
	"github.com/specterops/bloodhound/graphschema/ad"
	"github.com/specterops/bloodhound/graphschema/common"
	"github.com/specterops/bloodhound/log"
	"github.com/specterops/bloodhound/src/database/types/nan"
	"github.com/specterops/bloodhound/src/model"
	"github.com/specterops/bloodhound/src/queries"
)

func ValidateDomains(ctx context.Context, queries queries.Graph, objectIDs ...string) ([]string, error) {
	var validated = make([]string, 0, len(objectIDs))

	if domains, err := queries.FetchNodesByObjectIDs(ctx, objectIDs...); err != nil {
		return nil, err
	} else {
		for _, objectID := range objectIDs {
			var found bool

			for _, domain := range domains.Slice() {
				if domainID, err := domain.Properties.Get(common.ObjectID.String()).String(); err != nil {
					return nil, err
				} else if domainID == objectID {
					found = true
					validated = append(validated, domainID)
					break
				}
			}

			if !found {
				return nil, fmt.Errorf("failed to find domain: %s", objectID)
			}
		}

		return validated, nil
	}
}

func GraphStats(ctx context.Context, db graph.Database) (model.ADDataQualityStats, model.ADDataQualityAggregation, error) {
	var (
		aggregation model.ADDataQualityAggregation
		adStats     = model.ADDataQualityStats{}
		runID       string

		kinds = graph.Kinds{ad.User, ad.Group, ad.Computer, ad.Container, ad.OU, ad.GPO, ad.AIACA, ad.RootCA, ad.EnterpriseCA, ad.NTAuthStore, ad.CertTemplate}
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
					log.Errorf("Domain node %d does not have a valid %s property: %v", domain.ID, common.ObjectID, err)
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
							NumReaders: analysis.MaximumDatabaseParallelWorkers,
						})
						mutex = &sync.Mutex{}
					)

					for _, kind := range kinds {
						innerKind := kind
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
