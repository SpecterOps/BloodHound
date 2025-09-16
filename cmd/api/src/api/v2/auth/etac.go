// Copyright 2025 Specter Ops, Inc.
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

package auth

import (
	"context"
	"errors"
	"fmt"

	"github.com/specterops/bloodhound/cmd/api/src/api"
	v2 "github.com/specterops/bloodhound/cmd/api/src/api/v2"
	"github.com/specterops/bloodhound/cmd/api/src/auth"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/cmd/api/src/queries"
	"github.com/specterops/bloodhound/packages/go/graphschema/ad"
	"github.com/specterops/bloodhound/packages/go/graphschema/azure"
	"github.com/specterops/bloodhound/packages/go/graphschema/common"
	"github.com/specterops/dawgs/graph"
)

// handleETACRequest will modify the user passed in to assign an etac list or grant all environment access
// and will return an error on bad requests
// Administrators and Power Users may not have an ETAC list applied to them
// The user may not request all environments and have an ETAC list applied to them
func handleETACRequest(ctx context.Context, updateUserRequest v2.UpdateUserRequest, roles model.Roles, user *model.User, graphDB queries.Graph) error {
	var (
		allEnvironments = true
	)

	// Set allEnvironments to true by default, and only change it if the request contains an explicit set
	// This avoids Go from defaulting the boolean to true when decoding the payload
	if updateUserRequest.AllEnvironments.Valid {
		allEnvironments = updateUserRequest.AllEnvironments.Bool
	}

	if roles.Has(model.Role{Name: auth.RoleAdministrator}) || roles.Has(model.Role{Name: auth.RolePowerUser}) {
		if !allEnvironments || (updateUserRequest.EnvironmentAccessControl != nil && len(updateUserRequest.EnvironmentAccessControl.Environments) > 0) {
			return errors.New(api.ErrorResponseETACInvalidRoles)
		}
	} else if (allEnvironments && updateUserRequest.AllEnvironments.Valid) && (updateUserRequest.EnvironmentAccessControl != nil && len(updateUserRequest.EnvironmentAccessControl.Environments) != 0) {
		// Both all_environments and environment_access_control was set on the request
		// A user may only have all_environments true or an environment access control list
		return errors.New(api.ErrorResponseETACBadRequest)
	} else if updateUserRequest.AllEnvironments.Valid {
		user.EnvironmentAccessControl = make([]model.EnvironmentAccess, 0)
		user.AllEnvironments = allEnvironments
	}

	if updateUserRequest.EnvironmentAccessControl != nil {
		var (
			environments = make([]string, 0, len(updateUserRequest.EnvironmentAccessControl.Environments))
		)
		user.EnvironmentAccessControl = make([]model.EnvironmentAccess, 0)
		user.AllEnvironments = false

		for _, environment := range updateUserRequest.EnvironmentAccessControl.Environments {
			environments = append(environments, environment.EnvironmentID)
		}

		if nodes, err := graphDB.FetchNodesByObjectIDsAndKinds(ctx, graph.Kinds{
			ad.Domain, azure.Tenant,
		}, environments...); err != nil {
			return fmt.Errorf("error fetching environments: %w", err)
		} else {
			if nodesByObject, err := nodeSetToObjectIDMap(nodes); err != nil {
				return err
			} else {
				for _, environment := range environments {
					if _, ok := nodesByObject[environment]; !ok {
						return errors.New(fmt.Sprintf("domain or tenant not found: %s", environment))
					} else {
						user.EnvironmentAccessControl = append(user.EnvironmentAccessControl, model.EnvironmentAccess{
							UserID:        user.ID.String(),
							EnvironmentID: environment,
						})
					}
				}
			}
		}
	}

	return nil
}

func nodeSetToObjectIDMap(set graph.NodeSet) (map[string]bool, error) {
	var (
		objectIDs = make(map[string]bool)
	)

	for _, node := range set {
		if objectID, err := node.Properties.Get(common.ObjectID.String()).String(); err != nil {
			return objectIDs, err
		} else {
			objectIDs[objectID] = true
		}
	}

	return objectIDs, nil
}
