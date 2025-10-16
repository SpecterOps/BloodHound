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
	"log/slog"
	"strings"

	"github.com/specterops/bloodhound/packages/go/analysis"
	"github.com/specterops/bloodhound/packages/go/analysis/azure"
	adSchema "github.com/specterops/bloodhound/packages/go/graphschema/ad"
	azureSchema "github.com/specterops/bloodhound/packages/go/graphschema/azure"
	"github.com/specterops/bloodhound/packages/go/graphschema/common"
	"github.com/specterops/dawgs/graph"
	"github.com/specterops/dawgs/ops"
	"github.com/specterops/dawgs/query"
	"github.com/specterops/dawgs/util/channels"
)

func PostHybrid(ctx context.Context, db graph.Database) (*analysis.AtomicPostProcessingStats, error) {
	// Fetch all Azure tenants first
	tenants, err := azure.FetchTenants(ctx, db)
	if err != nil {
		return &analysis.AtomicPostProcessingStats{}, fmt.Errorf("fetching Entra tenants: %w", err)
	}

	// Spin up a new parallel operation to speed up processing
	operation := analysis.NewPostRelationshipOperation(ctx, db, "Hybrid Attack Paths Post Processing")

	err = db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		var (
			// entraObjIDMap is used to index AD user objectids by Entra node ids
			entraObjIDMap = make(map[graph.ID]string, 1024)
			// adObjIDMap is used as a reverse mapping of a list of Entra node ids indexed by the AD user objectids
			adObjIDMap = make(map[string][]graph.ID, 1024)
			// entraToADUserMap is the final mapping between an Entra user node id to an AD user node id
			entraToADUserMap = make(map[graph.ID]graph.ID, 1024)
			// entraToADComputerMap is the mapping between an Entra computer node id to an AD computer node id
			entraToADComputerMap = make(map[graph.ID]graph.ID, 1024)
			// A map of device ids to the entra node Id of the device
			adObjIdToEntraDeviceMap = make(map[string][]graph.ID, 1024)
		)

		// Work on Entra users by their tenant association. Loop therefore through each Entra tenant
		for _, tenant := range tenants {
			// Fetch all users in this Entra tenant
			if tenantUsers, err := fetchEntraUsers(tx, tenant); err != nil {
				return err
			} else if len(tenantUsers) == 0 {
				// If there are no users present, exit this loop
				continue
			} else {
				// Loop through each Entra user in this tenant
				for _, tenantUser := range tenantUsers {
					// Check to see if the Entra user has an on prem sync property set
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
						entraToADUserMap[tenantUser.ID] = 0
					}
				}
			}

			// We just get the tenantDevices for now since we don't haveh the AD devices
			if tenantDevices, err := fetchEntraDevices(tx, tenant); err != nil {
				return err
			} else if len(tenantDevices) == 0 {
				continue
			} else {
				for _, tenantDevice := range tenantDevices {
					// The deviceId is the object Id of the on-premises computer
					if deviceId, err := tenantDevice.Properties.Get(azureSchema.DeviceID.String()).String(); err != nil {
						return err
					} else {
						// Canonicalize to all uppercase
						deviceId = strings.ToUpper(deviceId)
						// [device Id uuid] -> node id of entra node id
						adObjIdToEntraDeviceMap[deviceId] = append(adObjIdToEntraDeviceMap[deviceId], tenantDevice.ID)
					}
				}
			}
		}

		// Because there's a chance for AD users to exist in the graph without having a valid domain node linked to them,
		// we need to grab all of them directly, unlike Entra
		if adUsers, err := fetchADUsers(tx); err != nil {
			return err
		} else {
			// Loop through each Active Directory user
			for _, adUser := range adUsers {
				// Get the user's Object ID
				if objectID, err := adUser.Properties.Get(common.ObjectID.String()).String(); err != nil {
					return err
				} else if azUsers, ok := adObjIDMap[objectID]; !ok {
					// Skip adding this relationship if we've already seen it before as that implies it will be created
					continue
				} else {
					// Because there could theoretically be more than one Entra user mapped to this objectid, we want to loop through all when adding our current id to the final map
					for _, azUser := range azUsers {
						entraToADUserMap[azUser] = adUser.ID
					}
				}
			}
		}

		if adComputers, err := fetchADComputers(tx); err != nil {
			return err
		} else {
			for _, adComputer := range adComputers {
				if objectID, err := adComputer.Properties.Get(string(adSchema.ObjectGUID)).String(); err != nil {
					// This node doesn't have an objectguid, continue
					continue
				} else if azComputers, ok := adObjIdToEntraDeviceMap[strings.ToUpper(objectID)]; !ok {
					continue
				} else {
					// There should only be a one to one mapping but just in case
					for _, azComputer := range azComputers {
						entraToADComputerMap[azComputer] = adComputer.ID
					}
				}
			}
		}

		if err := operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
			for azUser, potentialADUser := range entraToADUserMap {
				var adUser = potentialADUser

				// The 0 value should never be a valid id for an AD user node, just by the nature of the graph, so we're cheating
				// by checking if we set it to 0 as a flag that this node was never actually found, meaning it needs to be created first
				if potentialADUser == 0 {
					if adUserNode, err := createMissingADKind(ctx, db, entraObjIDMap[azUser], adSchema.User); err != nil {
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

			for azComputer, potentialADComputer := range entraToADComputerMap {
				var adComputer = potentialADComputer

				if potentialADComputer == 0 {
					if adComputerNode, err := createMissingADKind(ctx, db, entraObjIDMap[azComputer], adSchema.Computer); err != nil {
						return err
					} else {
						adComputer = adComputerNode.ID
					}
				}

				SyncedToEntraComputerRelationship := analysis.CreatePostRelationshipJob{
					FromID: adComputer,
					ToID:   azComputer,
					Kind:   adSchema.SyncedToEntraComputer,
				}

				if !channels.Submit(ctx, outC, SyncedToEntraComputerRelationship) {
					return nil
				}

				SyncedToADComputerRelationship := analysis.CreatePostRelationshipJob{
					FromID: azComputer,
					ToID:   adComputer,
					Kind:   azureSchema.SyncedToADComputer,
				}

				if !channels.Submit(ctx, outC, SyncedToADComputerRelationship) {
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
func createMissingADKind(ctx context.Context, db graph.Database, objectID string, kind graph.Kind) (*graph.Node, error) {
	var (
		err     error
		newNode *graph.Node
	)

	slog.DebugContext(ctx, fmt.Sprintf("Matching AD User node with objectID %s not found, creating a new one", objectID))
	properties := graph.AsProperties(map[string]any{
		common.ObjectID.String(): objectID,
	})

	err = db.WriteTransaction(ctx, func(tx graph.Transaction) error {
		if newNode, err = analysis.FetchNodeByObjectID(tx, objectID); errors.Is(err, graph.ErrNoResultsFound) {
			if newNode, err = tx.CreateNode(properties, adSchema.Entity, kind); err != nil {
				return fmt.Errorf("create missing %s: %w", kind, err)
			} else {
				return nil
			}
		} else if err != nil {
			return fmt.Errorf("create missing %s precheck: %w", kind, err)
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

func fetchEntraDevices(tx graph.Transaction, root *graph.Node) (graph.NodeSet, error) {
	return ops.FetchEndNodes(tx.Relationships().Filterf(func() graph.Criteria {
		return query.And(
			query.InIDs(query.StartID(), root.ID),
			query.Kind(query.Relationship(), azureSchema.Contains),
			query.KindIn(query.End(), azureSchema.Device),
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

func fetchADComputers(tx graph.Transaction) ([]*graph.Node, error) {
	return ops.FetchNodes(tx.Nodes().Filterf(func() graph.Criteria {
		return query.And(
			query.Kind(query.Node(), adSchema.Computer),
		)
	}))
}
