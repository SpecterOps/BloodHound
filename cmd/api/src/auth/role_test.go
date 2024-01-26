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
	"github.com/gofrs/uuid"
	"github.com/specterops/bloodhound/lab"
	"github.com/specterops/bloodhound/src/api"
	v2 "github.com/specterops/bloodhound/src/api/v2"
	"github.com/specterops/bloodhound/src/auth"
	"github.com/specterops/bloodhound/src/model/appcfg"
	"github.com/specterops/bloodhound/src/test/lab/fixtures"
	"github.com/stretchr/testify/require"
	"net/http"
	"testing"
	"time"
)

func TestRole_PowerUser(t *testing.T) {

	harness := lab.NewHarness()
	lab.Pack(harness, fixtures.BHAdminApiClientFixture)
	powerUserClientFixture := fixtures.NewUserApiClientFixture(fixtures.BHAdminApiClientFixture, auth.RolePowerUser)
	lab.Pack(harness, powerUserClientFixture)

	lab.NewSpec(t, harness).Run(
		lab.TestCase("SHOULD be able to access AppReadApplicationConfiguration endpoints", func(assert *require.Assertions, harness *lab.Harness) {
			powerUserClient, ok := lab.Unpack(harness, powerUserClientFixture)
			assert.True(ok)

			_, err := powerUserClient.GetAppConfigs()
			require.Nil(t, err)
		}),

		lab.TestCase("SHOULD be able to access AppWriteApplicationConfiguration endpoints", func(assert *require.Assertions, harness *lab.Harness) {
			powerUserClient, ok := lab.Unpack(harness, powerUserClientFixture)
			assert.True(ok)

			updatedPasswordExpirationWindowParameter := v2.AppConfigUpdateRequest{
				Key: appcfg.PasswordExpirationWindow,
				Value: map[string]any{
					"duration": "P30D",
				},
			}
			_, err := powerUserClient.PutAppConfig(updatedPasswordExpirationWindowParameter)
			require.Nil(t, err)
		}),

		lab.TestCase("SHOULD be able to access AuthCreateToken endpoints", func(assert *require.Assertions, harness *lab.Harness) {
			powerUserClient, ok := lab.Unpack(harness, powerUserClientFixture)
			assert.True(ok)

			randomUserUuid, err := uuid.NewV4()
			require.Nil(t, err)

			_, err = powerUserClient.ListUserTokens(randomUserUuid)
			require.Nil(t, err)
		}),

		lab.TestCase("SHOULD NOT be able to access AuthManageProviders endpoints", func(assert *require.Assertions, harness *lab.Harness) {
			powerUserClient, ok := lab.Unpack(harness, powerUserClientFixture)
			assert.True(ok)

			_, err := powerUserClient.ListSAMLIdentityProviders()
			var errByte []byte
			errByte, err = json.Marshal(err)
			require.Nil(t, err)

			errWrapper := api.ErrorWrapper{}
			err = json.Unmarshal(errByte, &errWrapper)
			require.Nil(t, err)
			require.Equal(t, errWrapper.HTTPStatus, http.StatusForbidden)
		}),

		lab.TestCase("SHOULD be able to access AuthManageSelf endpoints", func(assert *require.Assertions, harness *lab.Harness) {
			powerUserClient, ok := lab.Unpack(harness, powerUserClientFixture)
			assert.True(ok)

			_, err := powerUserClient.ListPermissions()
			require.Nil(t, err)
		}),

		lab.TestCase("SHOULD NOT be able to access AuthManageUsers endpoints", func(assert *require.Assertions, harness *lab.Harness) {
			powerUserClient, ok := lab.Unpack(harness, powerUserClientFixture)
			assert.True(ok)

			_, err := powerUserClient.ListAuditLogs(time.Now(), time.Now(), 0, 0)
			require.NotNil(t, err)

			var errByte []byte
			errByte, err = json.Marshal(err)
			require.Nil(t, err)

			errWrapper := api.ErrorWrapper{}
			err = json.Unmarshal(errByte, &errWrapper)
			require.Nil(t, err)
			require.Equal(t, errWrapper.HTTPStatus, http.StatusForbidden)
		}),

		lab.TestCase("SHOULD be able to access GraphDbWrite endpoints", func(assert *require.Assertions, harness *lab.Harness) {
			powerUserClient, ok := lab.Unpack(harness, powerUserClientFixture)
			assert.True(ok)

			_, err := powerUserClient.CreateFileUploadTask()
			require.Nil(t, err)
		}),

		lab.TestCase("SHOULD be able to access GraphDbRead endpoints", func(assert *require.Assertions, harness *lab.Harness) {
			powerUserClient, ok := lab.Unpack(harness, powerUserClientFixture)
			assert.True(ok)

			_, err := powerUserClient.ListAssetGroups()
			require.Nil(t, err)
		}),

		lab.TestCase("SHOULD be able to access SavedQueriesRead endpoints", func(assert *require.Assertions, harness *lab.Harness) {
			powerUserClient, ok := lab.Unpack(harness, powerUserClientFixture)
			assert.True(ok)

			_, err := powerUserClient.ListPermissions()
			require.Nil(t, err)
		}),

		lab.TestCase("SHOULD be able to access SavedQueriesWrite endpoints", func(assert *require.Assertions, harness *lab.Harness) {
			powerUserClient, ok := lab.Unpack(harness, powerUserClientFixture)
			assert.True(ok)

			_, err := powerUserClient.ListPermissions()
			require.Nil(t, err)
		}),
	)
}
