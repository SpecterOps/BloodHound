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
	AppReadApplicationConfiguration  model.Permission
	AppWriteApplicationConfiguration model.Permission

	APsGenerateReport model.Permission
	APsManageAPs      model.Permission

	AuthAcceptEULA                      model.Permission
	AuthCreateToken                     model.Permission
	AuthManageApplicationConfigurations model.Permission
	AuthManageProviders                 model.Permission
	AuthManageSelf                      model.Permission
	AuthManageUsers                     model.Permission

	ClientsManage  model.Permission
	ClientsRead    model.Permission
	ClientsTasking model.Permission

	CollectionManageJobs model.Permission

	GraphDBRead  model.Permission
	GraphDBWrite model.Permission

	SavedQueriesRead  model.Permission
	SavedQueriesWrite model.Permission
}

func (s PermissionSet) All() model.Permissions {
	return model.Permissions{
		s.AppReadApplicationConfiguration,
		s.AppWriteApplicationConfiguration,
		s.APsGenerateReport,
		s.APsManageAPs,
		s.AuthCreateToken,
		s.AuthManageApplicationConfigurations,
		s.AuthManageProviders,
		s.AuthManageSelf,
		s.AuthManageUsers,
		s.ClientsManage,
		s.ClientsRead,
		s.ClientsTasking,
		s.CollectionManageJobs,
		s.GraphDBRead,
		s.GraphDBWrite,
		s.SavedQueriesRead,
		s.SavedQueriesWrite,
	}
}

func Permissions() PermissionSet {
	return PermissionSet{
		AppReadApplicationConfiguration:  model.NewPermission("app", "ReadAppConfig"),
		AppWriteApplicationConfiguration: model.NewPermission("app", "WriteAppConfig"),

		APsGenerateReport: model.NewPermission("risks", "GenerateReport"),
		APsManageAPs:      model.NewPermission("risks", "ManageRisks"),

		AuthAcceptEULA:                      model.NewPermission("auth", "AcceptEULA"),
		AuthCreateToken:                     model.NewPermission("auth", "CreateToken"),
		AuthManageApplicationConfigurations: model.NewPermission("auth", "ManageAppConfig"),
		AuthManageProviders:                 model.NewPermission("auth", "ManageProviders"),
		AuthManageSelf:                      model.NewPermission("auth", "ManageSelf"),
		AuthManageUsers:                     model.NewPermission("auth", "ManageUsers"),

		ClientsManage:  model.NewPermission("clients", "Manage"),
		ClientsRead:    model.NewPermission("clients", "Read"),
		ClientsTasking: model.NewPermission("clients", "Tasking"),

		CollectionManageJobs: model.NewPermission("collection", "ManageJobs"),

		GraphDBRead:  model.NewPermission("graphdb", "Read"),
		GraphDBWrite: model.NewPermission("graphdb", "Write"),

		SavedQueriesRead:  model.NewPermission("saved_queries", "Read"),
		SavedQueriesWrite: model.NewPermission("saved_queries", "Write"),
	}
}
