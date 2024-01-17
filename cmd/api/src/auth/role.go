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

package auth

import (
	"fmt"

	"github.com/specterops/bloodhound/src/model"
)

const (
	RoleUploadOnly    = "Upload-Only"
	RoleReadOnly      = "Read-Only"
	RoleUser          = "User"
	RoleAdministrator = "Administrator"
)

type RoleTemplate struct {
	Name        string
	Description string
	Permissions model.Permissions
}

func (s RoleTemplate) Build(allPermissions model.Permissions) (model.Role, error) {
	role := model.Role{
		Name:        s.Name,
		Description: s.Description,
		Permissions: make(model.Permissions, len(s.Permissions)),
	}

	for idx, requiredPermission := range s.Permissions {
		found := false

		for _, permission := range allPermissions {
			if permission.Equals(requiredPermission) {
				role.Permissions[idx] = permission
				found = true

				break
			}
		}

		if !found {
			return role, fmt.Errorf("unable to locate required permission %s for role template %s", requiredPermission, s.Name)
		}
	}

	return role, nil
}

func Roles() map[string]RoleTemplate {
	permissions := Permissions()

	return map[string]RoleTemplate{
		RoleReadOnly: {
			Name:        RoleReadOnly,
			Description: "Used for integrations",
			Permissions: model.Permissions{
				permissions.GraphDBRead,
				permissions.AuthManageSelf,
				permissions.APsGenerateReport,
				permissions.AppReadApplicationConfiguration,
			},
		},
		RoleUploadOnly: {
			Name:        RoleUploadOnly,
			Description: "Used for data collection clients, can post data but cannot read data",
			Permissions: model.Permissions{
				permissions.GraphDBWrite,
				permissions.ClientsTasking,
			},
		},
		RoleUser: {
			Name:        RoleUser,
			Description: "Can read data, modify asset group memberships",
			Permissions: model.Permissions{
				permissions.GraphDBRead,
				permissions.ClientsManage,
				permissions.AuthCreateToken,
				permissions.AuthManageSelf,
				permissions.APsGenerateReport,
				permissions.AppReadApplicationConfiguration,
				permissions.SavedQueriesRead,
				permissions.SavedQueriesWrite,
			},
		},
		RoleAdministrator: {
			Name:        RoleAdministrator,
			Description: "Can manage users, clients, and application configuration",
			Permissions: permissions.All(),
		},
	}
}
