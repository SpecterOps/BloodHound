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
	"github.com/specterops/bloodhound/cmd/api/src/model"
)

const (
	RoleUploadOnly    = "Upload-Only"
	RoleReadOnly      = "Read-Only"
	RoleUser          = "User"
	RolePowerUser     = "Power User"
	RoleAdministrator = "Administrator"
)

type RoleTemplate struct {
	Name        string
	Description string
	Permissions model.Permissions
}

// Roles Note: Not the source of truth, changes here must be added to a migration *.sql file to update the roles & roles_permissions table
func Roles() map[string]RoleTemplate {
	permissions := Permissions()

	return map[string]RoleTemplate{
		RoleReadOnly: {
			Name:        RoleReadOnly,
			Description: "Used for integrations",
			Permissions: model.Permissions{
				permissions.AppReadApplicationConfiguration,
				permissions.APsGenerateReport,
				permissions.AuthCreateToken,
				permissions.AuthManageSelf,
				permissions.GraphDBRead,
				permissions.SavedQueriesRead,
			},
		},
		RoleUploadOnly: {
			Name:        RoleUploadOnly,
			Description: "Used for data collection clients, can post data but cannot read data",
			Permissions: model.Permissions{
				permissions.ClientsTasking,
				permissions.GraphDBIngest,
			},
		},
		RoleUser: {
			Name:        RoleUser,
			Description: "Can read data, modify asset group memberships",
			Permissions: model.Permissions{
				permissions.AppReadApplicationConfiguration,
				permissions.APsGenerateReport,
				permissions.AuthCreateToken,
				permissions.AuthManageSelf,
				permissions.ClientsRead,
				permissions.GraphDBRead,
				permissions.SavedQueriesRead,
				permissions.SavedQueriesWrite,
			},
		},
		RolePowerUser: {
			Name:        RolePowerUser,
			Description: "Can upload data, manage clients, and perform any action a User can",
			Permissions: model.Permissions{
				permissions.APsGenerateReport,
				permissions.APsManageAPs,
				permissions.AuthCreateToken,
				permissions.AuthManageSelf,
				permissions.ClientsManage,
				permissions.ClientsRead,
				permissions.ClientsTasking,
				permissions.CollectionManageJobs,
				permissions.GraphDBIngest,
				permissions.GraphDBWrite,
				permissions.GraphDBRead,
				permissions.SavedQueriesRead,
				permissions.SavedQueriesWrite,
				permissions.GraphDBMutate,
			},
		},
		RoleAdministrator: {
			Name:        RoleAdministrator,
			Description: "Can manage users, clients, and application configuration",
			Permissions: permissions.All(),
		},
	}
}
