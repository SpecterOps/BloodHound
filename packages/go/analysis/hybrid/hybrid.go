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
	tenants, err := azure.FetchTenants(ctx, db)
	if err != nil {
		return &analysis.AtomicPostProcessingStats{}, fmt.Errorf("fetching Entra tenants: %w", err)
	}

	operation := analysis.NewPostRelationshipOperation(ctx, db, "Hybrid Attack Paths Post Processing")

	err = db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		var (
			// entraObjIDMap is used to index AD user objectids by Entra node ids
			entraObjIDMap = make(map[graph.ID]string, 1024)
			// adObjIDMap is used as a reverse mapping of a list of Entra node ids indexed by the AD user objectids
			adObjIDMap = make(map[string][]graph.ID, 1024)
			// entraToADMap is the final mapping between an Entra user node id to an AD user node id
			entraToADMap = make(map[graph.ID]graph.ID, 1024)
		)

		// Work on Entra users by their tenant association
		for _, tenant := range tenants {
			if tenantUsers, err := fetchEntraUsers(tx, tenant); err != nil {
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
						// We know this user has an onPrem counterpart, so add the node id and onPremID to our three maps
						adObjIDMap[onPremID] = append(adObjIDMap[onPremID], tenantUser.ID)
						entraObjIDMap[tenantUser.ID] = onPremID
						// Initialize the current user id as an index in the entraToADMap, but use 0 as the nodeid for AD since we
						// currently don't know it and 0 is never going to be a valid user node id
						entraToADMap[tenantUser.ID] = 0
					}
				}
			}
		}

		// Because there's a chance for AD users to exist in the graph without having a valid domain node linked to them,
		// we need to grab all of them directly, unlike Entra
		if adUsers, err := fetchADUsers(tx); err != nil {
			return err
		} else {
			for _, adUser := range adUsers {
				if objectID, err := adUser.Properties.Get(common.ObjectID.String()).String(); err != nil {
					return err
				} else if azUsers, ok := adObjIDMap[objectID]; !ok {
					continue
				} else {
					// Because there could theoretically be more than one Entra user mapped to this objectid, we want to loop through all when adding our current id to the final map
					for _, azUser := range azUsers {
						entraToADMap[azUser] = adUser.ID
					}
				}
			}
		}

		if err := operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
			for azUser, potentialADUser := range entraToADMap {
				var adUser = potentialADUser

				// The 0 value should never be a valid id for an AD user node, just by the nature of the graph, so we're cheating
				// by checking if we set it to 0 as a flag that this node was never actually found, meaning it needs to be created first
				if potentialADUser == 0 {
					if adUserNode, err := createMissingADUser(ctx, db, entraObjIDMap[azUser]); err != nil {
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

	// Because we need to close the operation either way at this stage, we attempt to close it and then report either or
	// both errors in one line
	if opErr := operation.Done(); opErr != nil || err != nil {
		return &operation.Stats, fmt.Errorf("marking operation as done: %w; transaction error (if any): %v", opErr, err)
	}

	return &operation.Stats, nil
}

// hasOnPremUser takes a node and returns the OnPremID as a string, whether the node has an onPrem user defined as a bool
// and any errors in negotiation of the required properties
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

// createMissingADUser will create a new standalone AD User node with the required objectID for displaying in hybrid graphs
func createMissingADUser(ctx context.Context, db graph.Database, objectID string) (*graph.Node, error) {
	var (
		err     error
		newNode *graph.Node
	)

	log.Debugf("Matching AD User node with objectID %s not found, creating a new one", objectID)
	properties := graph.AsProperties(map[string]any{
		common.ObjectID.String(): objectID,
	})

	err = db.WriteTransaction(ctx, func(tx graph.Transaction) error {
		if newNode, err = analysis.FetchNodeByObjectID(tx, objectID); errors.Is(err, graph.ErrNoResultsFound) {
			if newNode, err = tx.CreateNode(properties, adSchema.Entity, adSchema.User); err != nil {
				return fmt.Errorf("create missing ad user: %w", err)
			} else {
				return nil
			}
		} else if err != nil {
			return fmt.Errorf("create missing ad user precheck: %w", err)
		} else {
			return nil
		}
	})

	return newNode, err
}

// fetchEntraUsers fetches all the Entra users for a given root node (generally the tenant node)
func fetchEntraUsers(tx graph.Transaction, root *graph.Node) (graph.NodeSet, error) {
	return ops.FetchEndNodes(tx.Relationships().Filterf(func() graph.Criteria {
		return query.And(
			query.InIDs(query.StartID(), root.ID),
			query.Kind(query.Relationship(), azureSchema.Contains),
			query.KindIn(query.End(), azureSchema.User),
		)
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
