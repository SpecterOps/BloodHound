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
func handleETACRequest(ctx context.Context, etacRequest v2.UpdateUserEnvironmentRequest, roles model.Roles, user *model.User, graphDB queries.Graph) error {
	user.AllEnvironments = etacRequest.AllEnvironments

	if roles.Has(model.Role{Name: auth.RoleAdministrator}) || roles.Has(model.Role{Name: auth.RolePowerUser}) {
		return errors.New(api.ErrorResponseETACInvalidRoles)
	} else if len(etacRequest.Environments) != 0 && etacRequest.AllEnvironments {
		return errors.New(api.ErrorResponseETACBadRequest)
	} else if etacRequest.AllEnvironments {
		user.EnvironmentAccessControl = make([]model.EnvironmentAccess, 0)
	} else {
		if nodes, err := graphDB.FetchNodesByObjectIDsAndKinds(ctx, graph.Kinds{
			ad.Domain, azure.Tenant,
		}, etacRequest.Environments...); err != nil {
			return fmt.Errorf("error fetching environments: %w", err)
		} else {
			if nodesByObject, err := nodeSetToObjectIDMap(nodes); err != nil {
				return err
			} else {
				environments := make([]model.EnvironmentAccess, 0, len(etacRequest.Environments))
				for _, environment := range etacRequest.Environments {
					if _, ok := nodesByObject[environment]; !ok {
						return errors.New(fmt.Sprintf("domain or tenant not found: %s", environment))
					} else {
						environments = append(environments, model.EnvironmentAccess{
							UserID:      user.ID.String(),
							Environment: environment,
						})
					}
				}
				user.EnvironmentAccessControl = environments
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
