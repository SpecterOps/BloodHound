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
	if updateUserRequest.AllEnvironments.Valid || updateUserRequest.EnvironmentTargetedAccessControl != nil {
		// Admin / Power Users can only have all_environments set to true
		if (roles.Has(model.Role{Name: auth.RoleAdministrator}) || roles.Has(model.Role{Name: auth.RolePowerUser})) &&
			(!updateUserRequest.AllEnvironments.Bool || (updateUserRequest.EnvironmentTargetedAccessControl != nil && len(updateUserRequest.EnvironmentTargetedAccessControl.Environments) > 0)) {
			return errors.New(api.ErrorResponseETACInvalidRoles)
		}
		user.AllEnvironments = updateUserRequest.AllEnvironments.Bool
	}

	if updateUserRequest.EnvironmentTargetedAccessControl == nil || len(updateUserRequest.EnvironmentTargetedAccessControl.Environments) == 0 {
		user.EnvironmentTargetedAccessControl = make([]model.EnvironmentTargetedAccessControl, 0)
		return nil
	}

	// Both all_environments and environment_access_control was set on the request
	// A user may only have all_environments true or an environment access control list
	if updateUserRequest.AllEnvironments.Bool {
		return errors.New(api.ErrorResponseETACBadRequest)
	}

	// Validate provided environment ids exist in the graph prior to adding ETAC control
	var envIds []string
	for _, environment := range updateUserRequest.EnvironmentTargetedAccessControl.Environments {
		envIds = append(envIds, environment.EnvironmentID)
	}

	if nodes, err := graphDB.FetchNodesByObjectIDsAndKinds(ctx, graph.Kinds{
		ad.Domain, azure.Tenant,
	}, envIds...); err != nil {
		return fmt.Errorf("error fetching environments: %w", err)
	} else if nodesByObject, err := nodeSetToObjectIDMap(nodes); err != nil {
		return err
	} else {
		user.EnvironmentTargetedAccessControl = make([]model.EnvironmentTargetedAccessControl, 0, len(envIds))
		for _, envId := range envIds {
			if _, ok := nodesByObject[envId]; !ok {
				return fmt.Errorf("domain or tenant not found: %s", envId)
			} else {
				user.EnvironmentTargetedAccessControl = append(user.EnvironmentTargetedAccessControl, model.EnvironmentTargetedAccessControl{
					UserID:        user.ID.String(),
					EnvironmentID: envId,
				})
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
