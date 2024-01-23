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
	"github.com/specterops/bloodhound/src/model"
)

type PermissionSet struct {
	GraphDBRead  model.Permission
	GraphDBWrite model.Permission

	AppReadApplicationConfiguration  model.Permission
	AppWriteApplicationConfiguration model.Permission

	CollectionManageJobs model.Permission

	ClientsManage  model.Permission
	ClientsTasking model.Permission

	AuthCreateToken                     model.Permission
	AuthManageSelf                      model.Permission
	AuthAcceptEULA                      model.Permission
	AuthManageUsers                     model.Permission
	AuthManageProviders                 model.Permission
	AuthManageApplicationConfigurations model.Permission

	APsGenerateReport model.Permission
	APsManageAPs      model.Permission

	SavedQueriesRead  model.Permission
	SavedQueriesWrite model.Permission

	ClientsRead model.Permission
}

func (s PermissionSet) All() model.Permissions {
	return model.Permissions{
		s.GraphDBWrite,
		s.GraphDBRead,
		s.AppReadApplicationConfiguration,
		s.AppWriteApplicationConfiguration,
		s.CollectionManageJobs,
		s.ClientsManage,
		s.ClientsTasking,
		s.AuthCreateToken,
		s.AuthManageUsers,
		s.AuthManageProviders,
		s.AuthManageSelf,
		s.AuthManageApplicationConfigurations,
		s.APsGenerateReport,
		s.APsManageAPs,
		s.SavedQueriesRead,
		s.SavedQueriesWrite,
		s.ClientsRead,
	}
}

func Permissions() PermissionSet {
	return PermissionSet{
		GraphDBRead:  model.NewPermission("graphdb", "Read"),
		GraphDBWrite: model.NewPermission("graphdb", "Write"),

		AppReadApplicationConfiguration:  model.NewPermission("app", "ReadAppConfig"),
		AppWriteApplicationConfiguration: model.NewPermission("app", "WriteAppConfig"),

		CollectionManageJobs: model.NewPermission("collection", "ManageJobs"),

		ClientsManage:  model.NewPermission("clients", "Manage"),
		ClientsTasking: model.NewPermission("clients", "Tasking"),

		AuthCreateToken:                     model.NewPermission("auth", "CreateToken"),
		AuthManageSelf:                      model.NewPermission("auth", "ManageSelf"),
		AuthAcceptEULA:                      model.NewPermission("auth", "AcceptEULA"),
		AuthManageProviders:                 model.NewPermission("auth", "ManageProviders"),
		AuthManageUsers:                     model.NewPermission("auth", "ManageUsers"),
		AuthManageApplicationConfigurations: model.NewPermission("auth", "ManageAppConfig"),

		APsGenerateReport: model.NewPermission("risks", "GenerateReport"),
		APsManageAPs:      model.NewPermission("risks", "ManageRisks"),

		SavedQueriesRead:  model.NewPermission("saved_queries", "Read"),
		SavedQueriesWrite: model.NewPermission("saved_queries", "Write"),

		ClientsRead: model.NewPermission("clients", "Read"),
	}
}
