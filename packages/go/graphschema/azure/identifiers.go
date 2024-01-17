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

package azure

import "github.com/specterops/bloodhound/dawgs/graph"

const (
	MSGraphAppUniversalID                  = "00000003-0000-0000-c000-000000000000"
	ApplicationReadWriteAllID              = "1bfefb4e-e0b5-418b-a88f-73c46d2cc8e9"
	AppRoleAssignmentReadWriteAllID        = "06b708a9-e830-4db3-a914-8e69da51d44f"
	DirectoryReadWriteAllID                = "19dbc75e-c2e2-444c-a770-ec69d8559fc7"
	GroupReadWriteAllID                    = "62a82d76-70ea-41e2-9197-370581804d09"
	GroupMemberReadWriteAllID              = "dbaae8cf-10b5-4b86-a4a1-f871c94c6695"
	RoleManagementReadWriteDirectoryID     = "9e3f62cf-ca93-4989-b6ce-bf83c28f9fe8"
	ServicePrincipalEndpointReadWriteAllID = "89c8469c-83ad-45f7-8ff2-6e3d4285709e"
)

var (
	AllAppRoleIDs = []string{
		ApplicationReadWriteAllID,
		AppRoleAssignmentReadWriteAllID,
		DirectoryReadWriteAllID,
		GroupReadWriteAllID,
		GroupMemberReadWriteAllID,
		RoleManagementReadWriteDirectoryID,
		ServicePrincipalEndpointReadWriteAllID,
	}

	RelationshipKindByAppRoleID = map[string]graph.Kind{
		ApplicationReadWriteAllID:              ApplicationReadWriteAll,
		AppRoleAssignmentReadWriteAllID:        AppRoleAssignmentReadWriteAll,
		DirectoryReadWriteAllID:                DirectoryReadWriteAll,
		GroupReadWriteAllID:                    GroupReadWriteAll,
		GroupMemberReadWriteAllID:              GroupMemberReadWriteAll,
		RoleManagementReadWriteDirectoryID:     RoleManagementReadWriteDirectory,
		ServicePrincipalEndpointReadWriteAllID: ServicePrincipalEndpointReadWriteAll,
	}
)
