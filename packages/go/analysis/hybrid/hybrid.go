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

	"github.com/specterops/bloodhound/analysis"
	"github.com/specterops/bloodhound/analysis/ad"
	"github.com/specterops/bloodhound/analysis/azure"
	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/dawgs/ops"
	"github.com/specterops/bloodhound/dawgs/query"
	"github.com/specterops/bloodhound/dawgs/util/channels"
	adSchema "github.com/specterops/bloodhound/graphschema/ad"
	azureSchema "github.com/specterops/bloodhound/graphschema/azure"
	"github.com/specterops/bloodhound/graphschema/common"
	"github.com/specterops/bloodhound/log"
)

func PostHybrid(ctx context.Context, db graph.Database) (*analysis.AtomicPostProcessingStats, error) {
	defer log.Measure(log.LevelInfo, "Hybrid Post Processing")()

	log.Infof("Running Hybrid Post Processing")

	tenants, err := azure.FetchTenants(ctx, db)
	if err != nil {
		return &analysis.AtomicPostProcessingStats{}, fmt.Errorf("fetching Entra tenants: %w", err)
	}

	domains, err := ad.FetchAllDomains(ctx, db)
	if err != nil {
		return &analysis.AtomicPostProcessingStats{}, fmt.Errorf("fetching AD domains: %w", err)
	}

	operation := analysis.NewPostRelationshipOperation(ctx, db, "Hybrid Attack Paths Post Processing")

	err = db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		var (
			azObjIDMap = make(map[graph.ID]string, 1024)
			adObjIDMap = make(map[string][]graph.ID, 1024)
			azToADMap  = make(map[graph.ID]graph.ID, 1024)
		)

		for _, tenant := range tenants {
			if tenantUsers, err := fetchAZUsers(tx, tenant); err != nil {
				return err
			} else if len(tenantUsers) == 0 {
				continue
			} else {
				for _, tenantUser := range tenantUsers {
					if onPremID, hasOnPrem, err := hasOnPremUser(tenantUser); !hasOnPrem {
						continue
					} else if err != nil {
						return err
					} else {
						adObjIDMap[onPremID] = append(adObjIDMap[onPremID], tenantUser.ID)
						azObjIDMap[tenantUser.ID] = onPremID
						azToADMap[tenantUser.ID] = 0
					}
				}
			}
		}

		for _, domain := range domains {
			if domainUsers, err := fetchADUsers(tx, domain); err != nil {
				return err
			} else if len(domainUsers) == 0 {
				continue
			} else {
				for _, adUser := range domainUsers {
					if objectID, err := adUser.Properties.Get(common.ObjectID.String()).String(); err != nil {
						return err
					} else if azUsers, ok := adObjIDMap[objectID]; !ok {
						continue
					} else {
						for _, azUser := range azUsers {
							azToADMap[azUser] = adUser.ID
						}
					}
				}
			}
		}

		if err := operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
			for azUser, potentialADUser := range azToADMap {
				var adUser = potentialADUser

				if potentialADUser == 0 {
					if adUserNode, err := createMissingAdUser(ctx, db, azObjIDMap[azUser]); err != nil {
						return err
					} else {
						adUser = adUserNode.ID
					}
				}

				SyncedToEntraUserRelationship := analysis.CreatePostRelationshipJob{
					FromID: adUser,
					ToID:   azUser,
					Kind:   adSchema.SyncedToEntraUser,
				}

				if !channels.Submit(ctx, outC, SyncedToEntraUserRelationship) {
					return nil
				}

				SyncedToADUserRelationship := analysis.CreatePostRelationshipJob{
					FromID: azUser,
					ToID:   adUser,
					Kind:   azureSchema.SyncedToADUser,
				}

				if !channels.Submit(ctx, outC, SyncedToADUserRelationship) {
					return nil
				}
			}

			return nil
		}); err != nil {
			return err
		}

		return tx.Commit()
	})

	if opErr := operation.Done(); opErr != nil || err != nil {
		return &operation.Stats, fmt.Errorf("marking operation as done: %w; transaction error (if any): %v", opErr, err)
	}

	return &operation.Stats, nil
}

func hasOnPremUser(node *graph.Node) (string, bool, error) {
	if onPremSyncEnabled, err := node.Properties.Get(azureSchema.OnPremSyncEnabled.String()).Bool(); errors.Is(err, graph.ErrPropertyNotFound) {
		return "", false, nil
	} else if err != nil {
		return "", false, err
	} else if onPremID, err := node.Properties.Get(azureSchema.OnPremID.String()).String(); errors.Is(err, graph.ErrPropertyNotFound) {
		return onPremID, false, nil
	} else if err != nil {
		return onPremID, false, err
	} else {
		return onPremID, (onPremSyncEnabled && len(onPremID) != 0), nil
	}
}

func createMissingAdUser(ctx context.Context, db graph.Database, objectID string) (*graph.Node, error) {
	var (
		err     error
		newNode *graph.Node
	)

	log.Debugf("Matching AD User node with objectID %s not found, creating a new one", objectID)
	properties := graph.AsProperties(map[string]any{
		common.ObjectID.String(): objectID,
	})

	err = db.WriteTransaction(ctx, func(tx graph.Transaction) error {
		newNode, err = tx.CreateNode(properties, adSchema.Entity, adSchema.User)
		return err
	})

	return newNode, err
}

// TODO: decide which direction we're going, whether to grab all users with onPrem or iterate over each tenant
func fetchAZUsers(tx graph.Transaction, root *graph.Node) (graph.NodeSet, error) {
	return ops.FetchEndNodes(tx.Relationships().Filterf(func() graph.Criteria {
		return query.And(
			query.InIDs(query.StartID(), root.ID),
			query.Kind(query.Relationship(), azureSchema.Contains),
			query.KindIn(query.End(), azureSchema.User),
		)
	}))
}

func fetchADUsers(tx graph.Transaction, root *graph.Node) (graph.NodeSet, error) {
	return ops.FetchEndNodes(tx.Relationships().Filterf(func() graph.Criteria {
		return query.And(
			query.InIDs(query.StartID(), root.ID),
			query.Kind(query.Relationship(), adSchema.Contains),
			query.KindIn(query.End(), adSchema.User),
		)
	}))
}
