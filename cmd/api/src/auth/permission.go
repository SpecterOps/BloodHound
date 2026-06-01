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

type PermissionSet struct {
	AppReadApplicationConfiguration  model.Permission
	AppWriteApplicationConfiguration model.Permission

	APsGenerateReport model.Permission
	APsManageAPs      model.Permission

	AuditLogRead model.Permission

	AuthAcceptEULA       model.Permission
	AuthCreateToken      model.Permission
	AuthManageProviders  model.Permission
	AuthManageSelf       model.Permission
	AuthManageUsers      model.Permission
	AuthReadUsers        model.Permission
	AuthReadUsersMinimal model.Permission

	ClientsManage  model.Permission
	ClientsRead    model.Permission
	ClientsTasking model.Permission

	CollectionReadJobs   model.Permission
	CollectionManageJobs model.Permission

	GraphDBIngest model.Permission
	GraphDBMutate model.Permission
	GraphDBRead   model.Permission
	GraphDBWrite  model.Permission

	OpenGraphRead  model.Permission
	OpenGraphWrite model.Permission

	SavedQueriesRead  model.Permission
	SavedQueriesWrite model.Permission

	WipeDB model.Permission
}

func (s PermissionSet) All() model.Permissions {
	return model.Permissions{
		s.AppReadApplicationConfiguration,
		s.AppWriteApplicationConfiguration,
		s.APsGenerateReport,
		s.APsManageAPs,
		s.AuditLogRead,
		s.AuthCreateToken,
		s.AuthManageProviders,
		s.AuthManageSelf,
		s.AuthManageUsers,
		s.AuthReadUsers,
		s.AuthReadUsersMinimal,
		s.ClientsManage,
		s.ClientsRead,
		s.ClientsTasking,
		s.CollectionReadJobs,
		s.CollectionManageJobs,
		s.GraphDBIngest,
		s.GraphDBMutate,
		s.GraphDBRead,
		s.GraphDBWrite,
		s.OpenGraphRead,
		s.OpenGraphWrite,
		s.SavedQueriesRead,
		s.SavedQueriesWrite,
		s.WipeDB,
	}
}

func (s PermissionSet) ReadAll() model.Permissions {
	return model.Permissions{
		s.AppReadApplicationConfiguration,
		s.APsGenerateReport,
		s.AuditLogRead,
		s.AuthReadUsers,
		s.AuthReadUsersMinimal,
		s.ClientsRead,
		s.GraphDBRead,
		s.SavedQueriesRead,
		s.CollectionReadJobs,
		s.OpenGraphRead,
	}
}

// Permissions Note: Not the only source of truth, changes here must be added to a migration *.sql file to update the permissions table
func Permissions() PermissionSet {
	return PermissionSet{
		AppReadApplicationConfiguration:  model.NewPermission("app", "ReadAppConfig"),
		AppWriteApplicationConfiguration: model.NewPermission("app", "WriteAppConfig"),

		APsGenerateReport: model.NewPermission("risks", "GenerateReport"),
		APsManageAPs:      model.NewPermission("risks", "ManageRisks"),

		AuditLogRead: model.NewPermission("audit_log", "Read"),

		AuthAcceptEULA:       model.NewPermission("auth", "AcceptEULA"),
		AuthCreateToken:      model.NewPermission("auth", "CreateToken"),
		AuthManageProviders:  model.NewPermission("auth", "ManageProviders"),
		AuthManageSelf:       model.NewPermission("auth", "ManageSelf"),
		AuthManageUsers:      model.NewPermission("auth", "ManageUsers"),
		AuthReadUsers:        model.NewPermission("auth", "ReadUsers"),
		AuthReadUsersMinimal: model.NewPermission("auth", "ReadUsersMinimal"),

		ClientsManage:  model.NewPermission("clients", "Manage"),
		ClientsRead:    model.NewPermission("clients", "Read"),
		ClientsTasking: model.NewPermission("clients", "Tasking"),

		CollectionReadJobs:   model.NewPermission("collection", "ReadJobs"),
		CollectionManageJobs: model.NewPermission("collection", "ManageJobs"),

		GraphDBIngest: model.NewPermission("graphdb", "Ingest"),
		GraphDBMutate: model.NewPermission("graphdb", "Mutate"),
		GraphDBRead:   model.NewPermission("graphdb", "Read"),
		GraphDBWrite:  model.NewPermission("graphdb", "Write"),

		OpenGraphRead:  model.NewPermission("opengraph", "Read"),
		OpenGraphWrite: model.NewPermission("opengraph", "Write"),

		SavedQueriesRead:  model.NewPermission("saved_queries", "Read"),
		SavedQueriesWrite: model.NewPermission("saved_queries", "Write"),

		WipeDB: model.NewPermission("db", "Wipe"),
	}
}
