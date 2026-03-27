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

	"github.com/specterops/bloodhound/packages/go/analysis"
	"github.com/specterops/bloodhound/packages/go/analysis/azure"
	"github.com/specterops/bloodhound/packages/go/bhlog/attr"
	"github.com/specterops/bloodhound/packages/go/bhlog/measure"
	adSchema "github.com/specterops/bloodhound/packages/go/graphschema/ad"
	azureSchema "github.com/specterops/bloodhound/packages/go/graphschema/azure"
	"github.com/specterops/bloodhound/packages/go/graphschema/common"
	"github.com/specterops/dawgs/graph"
	"github.com/specterops/dawgs/ops"
	"github.com/specterops/dawgs/query"
	"github.com/specterops/dawgs/util/channels"
)

func PostHybrid(ctx context.Context, db graph.Database) (*analysis.AtomicPostProcessingStats, error) {
	defer measure.ContextLogAndMeasure(
		ctx,
		slog.LevelInfo,
		"Post-processing AD-Azure Hybrid Edges",
		attr.Namespace("analysis"),
		attr.Function("PostHybrid"),
		attr.Scope("process"),
	)()

	// Fetch all Azure tenants first
	tenants, err := azure.FetchTenants(ctx, db)
	if err != nil {
		return &analysis.AtomicPostProcessingStats{}, fmt.Errorf("fetching Entra tenants: %w", err)
	}

	// Spin up a new parallel operation to speed up processing
	operation := analysis.NewPostRelationshipOperation(ctx, db, "Hybrid Attack Paths Post Processing")

	err = db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		var (
			// adObjIDMap is used as a reverse mapping of a list of Entra node ids indexed by the AD user objectids
			adObjIDMap = make(map[string][]graph.ID, 1024)
			// entraToADMap is the final mapping between an Entra user node id to an AD user node id
			entraToADMap = make(map[graph.ID]graph.ID, 1024)
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
						// We know this user has an onPrem counterpart, so add the node id and onPremID to our mapping inputs.
						adObjIDMap[onPremID] = append(adObjIDMap[onPremID], tenantUser.ID)
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
					// Skip AD users that do not correspond to any synced Entra users.
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
			for azUser, adUser := range entraToADMap {
				SyncedToEntraUserRelationship := analysis.CreatePostRelationshipJob{
					FromID: adUser,
					ToID:   azUser,
					Kind:   azureSchema.SyncedToEntraUser,
				}

				if !channels.Submit(ctx, outC, SyncedToEntraUserRelationship) {
					return nil
				}

				SyncedToADUserRelationship := analysis.CreatePostRelationshipJob{
					FromID: azUser,
					ToID:   adUser,
					Kind:   adSchema.SyncedToADUser,
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
