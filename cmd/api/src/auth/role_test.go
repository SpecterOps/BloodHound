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

//go:build integration
// +build integration

package auth_test

import (
	"encoding/json"
	"fmt"
	"github.com/gofrs/uuid"
	"github.com/specterops/bloodhound/lab"
	"github.com/specterops/bloodhound/src/api"
	v2 "github.com/specterops/bloodhound/src/api/v2"
	"github.com/specterops/bloodhound/src/auth"
	"github.com/specterops/bloodhound/src/model"
	"github.com/specterops/bloodhound/src/model/appcfg"
	"github.com/specterops/bloodhound/src/test/lab/fixtures"
	"github.com/stretchr/testify/require"
	"net/http"
	"testing"
	"time"
)

func testCondition(role auth.RoleTemplate, permission model.Permission) string {
	if role.Permissions.Has(permission) {
		return "SHOULD"
	}
	return "SHOULD NOT"
}

func requireForbidden(t *testing.T, err error) {
	var errByte []byte
	errByte, err = json.Marshal(err)
	require.Nil(t, err)

	errWrapper := api.ErrorWrapper{}
	err = json.Unmarshal(errByte, &errWrapper)
	require.Nilf(t, err, "Failed to unmarshal error %v", string(errByte))
	require.Equal(t, errWrapper.HTTPStatus, http.StatusForbidden)
}

func testRoleAccess(t *testing.T, roleName string) {
	role, ok := auth.Roles()[roleName]
	require.Truef(t, ok, "invalid role name")

	harness := lab.NewHarness()
	lab.Pack(harness, fixtures.BHAdminApiClientFixture)
	userClientFixture := fixtures.NewUserApiClientFixture(fixtures.BHAdminApiClientFixture, role.Name)
	lab.Pack(harness, userClientFixture)

	lab.NewSpec(t, harness).Run(
		lab.TestCase(fmt.Sprintf("%s be able to access AppReadApplicationConfiguration endpoints", testCondition(role, auth.Permissions().AppReadApplicationConfiguration)), func(assert *require.Assertions, harness *lab.Harness) {
			userClient, ok := lab.Unpack(harness, userClientFixture)
			assert.True(ok)

			_, err := userClient.GetAppConfigs()
			if role.Permissions.Has(auth.Permissions().AppReadApplicationConfiguration) {
				require.Nil(t, err)
			} else {
				requireForbidden(t, err)
			}
		}),

		lab.TestCase(fmt.Sprintf("%s be able to access AppWriteApplicationConfiguration endpoints", testCondition(role, auth.Permissions().AppWriteApplicationConfiguration)), func(assert *require.Assertions, harness *lab.Harness) {
			userClient, ok := lab.Unpack(harness, userClientFixture)
			assert.True(ok)

			updatedPasswordExpirationWindowParameter := v2.AppConfigUpdateRequest{
				Key: appcfg.PasswordExpirationWindow,
				Value: map[string]any{
					"duration": "P30D",
				},
			}
			_, err := userClient.PutAppConfig(updatedPasswordExpirationWindowParameter)
			if role.Permissions.Has(auth.Permissions().AppWriteApplicationConfiguration) {
				require.Nil(t, err)
			} else {
				requireForbidden(t, err)
			}
		}),

		lab.TestCase(fmt.Sprintf("%s be able to access own AuthCreateToken endpoints", testCondition(role, auth.Permissions().AuthCreateToken)), func(assert *require.Assertions, harness *lab.Harness) {
			userClient, ok := lab.Unpack(harness, userClientFixture)
			assert.True(ok)

			user, err := userClient.GetSelf()
			require.Nilf(t, err, "failed looking up user details")

			_, err = userClient.ListUserTokens(user.ID)
			if role.Permissions.Has(auth.Permissions().AuthCreateToken) {
				require.Nil(t, err)
			} else {
				requireForbidden(t, err)
			}
		}),

		lab.TestCase(fmt.Sprintf("%s be able to access AuthManageProviders endpoints", testCondition(role, auth.Permissions().AuthManageProviders)), func(assert *require.Assertions, harness *lab.Harness) {
			userClient, ok := lab.Unpack(harness, userClientFixture)
			assert.True(ok)

			_, err := userClient.ListSAMLIdentityProviders()
			if role.Permissions.Has(auth.Permissions().AuthManageProviders) {
				require.Nil(t, err)
			} else {
				requireForbidden(t, err)
			}
		}),

		lab.TestCase(fmt.Sprintf("%s be able to access AuthManageSelf endpoints", testCondition(role, auth.Permissions().AuthManageSelf)), func(assert *require.Assertions, harness *lab.Harness) {
			userClient, ok := lab.Unpack(harness, userClientFixture)
			assert.True(ok)

			_, err := userClient.ListPermissions()
			if role.Permissions.Has(auth.Permissions().AuthManageSelf) {
				require.Nil(t, err)
			} else {
				requireForbidden(t, err)
			}
		}),

		lab.TestCase(fmt.Sprintf("%s be able to access AuthManageUsers endpoints", testCondition(role, auth.Permissions().AuthManageUsers)), func(assert *require.Assertions, harness *lab.Harness) {
			userClient, ok := lab.Unpack(harness, userClientFixture)
			assert.True(ok)

			_, err := userClient.ListAuditLogs(time.Now(), time.Now(), 0, 0)
			if role.Permissions.Has(auth.Permissions().AuthManageUsers) {
				require.Nil(t, err)
			} else {
				requireForbidden(t, err)
			}
		}),

		lab.TestCase(fmt.Sprintf("%s be able to access GraphDBWrite endpoints", testCondition(role, auth.Permissions().GraphDBWrite)), func(assert *require.Assertions, harness *lab.Harness) {
			userClient, ok := lab.Unpack(harness, userClientFixture)
			assert.True(ok)

			_, err := userClient.CreateFileUploadTask()
			if role.Permissions.Has(auth.Permissions().GraphDBWrite) {
				require.Nil(t, err)
			} else {
				requireForbidden(t, err)
			}
		}),

		lab.TestCase(fmt.Sprintf("%s be able to access GraphDBRead endpoints", testCondition(role, auth.Permissions().GraphDBRead)), func(assert *require.Assertions, harness *lab.Harness) {
			userClient, ok := lab.Unpack(harness, userClientFixture)
			assert.True(ok)

			_, err := userClient.ListAssetGroups()
			if role.Permissions.Has(auth.Permissions().GraphDBRead) {
				require.Nil(t, err)
			} else {
				requireForbidden(t, err)
			}
		}),

		lab.TestCase(fmt.Sprintf("%s be able to access SavedQueriesRead endpoints", testCondition(role, auth.Permissions().SavedQueriesRead)), func(assert *require.Assertions, harness *lab.Harness) {
			userClient, ok := lab.Unpack(harness, userClientFixture)
			assert.True(ok)

			_, err := userClient.ListSavedQueries()
			if role.Permissions.Has(auth.Permissions().SavedQueriesRead) {
				require.Nil(t, err)
			} else {
				requireForbidden(t, err)
			}
		}),

		lab.TestCase(fmt.Sprintf("%s be able to access SavedQueriesWrite endpoints", testCondition(role, auth.Permissions().SavedQueriesWrite)), func(assert *require.Assertions, harness *lab.Harness) {
			userClient, ok := lab.Unpack(harness, userClientFixture)
			assert.True(ok)

			_, err := userClient.CreateSavedQuery()
			if role.Permissions.Has(auth.Permissions().SavedQueriesWrite) {
				require.Nil(t, err)
			} else {
				requireForbidden(t, err)
			}
		}),
	)
}

func TestRole_ReadOnly(t *testing.T) {
	testRoleAccess(t, auth.RoleReadOnly)
}

func TestRole_UploadOnly(t *testing.T) {
	testRoleAccess(t, auth.RoleUploadOnly)
}

func TestRole_User(t *testing.T) {
	testRoleAccess(t, auth.RoleUser)
}

func TestRole_PowerUser(t *testing.T) {
	testRoleAccess(t, auth.RolePowerUser)
}

func TestRole_Administrator(t *testing.T) {
	testRoleAccess(t, auth.RoleAdministrator)
}

func TestRole_Administrator_ListOtherUserTokens(t *testing.T) {
	harness := lab.NewHarness()
	lab.Pack(harness, fixtures.BHAdminApiClientFixture)
	lab.NewSpec(t, harness).Run(
		lab.TestCase("Should be able to access AuthCreateToken endpoints for other users", func(assert *require.Assertions, harness *lab.Harness) {
			adminClient, ok := lab.Unpack(harness, fixtures.BHAdminApiClientFixture)
			assert.True(ok)

			randoUser, err := uuid.NewV4()
			require.Nilf(t, err, "failed to create rando user")

			_, err = adminClient.ListUserTokens(randoUser)
			require.Nil(t, err)
		}),
	)
}
