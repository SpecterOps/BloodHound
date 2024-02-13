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

package v2_test

import (
	"testing"

	"github.com/specterops/bloodhound/lab"
	v2 "github.com/specterops/bloodhound/src/api/v2"
	"github.com/specterops/bloodhound/src/database"
	"github.com/specterops/bloodhound/src/test/lab/fixtures"
	"github.com/specterops/bloodhound/src/test/lab/harnesses"
	"github.com/stretchr/testify/require"
)

func Test_DatabaseManagement(t *testing.T) {
	var (
		harness            = harnesses.NewIntegrationTestHarness(fixtures.BHAdminApiClientFixture)
		assetGroup         = fixtures.NewAssetGroupFixture()
		assetGroupSelector = fixtures.NewAssetGroupSelectorFixture(assetGroup)
	)

	// packing `assetGroupSelector` packs all the things it depends on too
	lab.Pack(harness, assetGroupSelector)

	lab.NewSpec(t, harness).Run(
		lab.TestCase("dummy", func(assert *require.Assertions, harness *lab.Harness) {
			apiClient, ok := lab.Unpack(harness, fixtures.BHAdminApiClientFixture)
			assert.True(ok)

			selector, ok := lab.Unpack(harness, assetGroupSelector)
			assert.True(ok, "unable to unpack asset group selector")

			err := apiClient.HandleDatabaseManagement(
				v2.DatabaseManagement{
					HighValueSelectors: true,
					AssetGroupId:       int(selector.AssetGroupID),
				})
			assert.Nil(err, "error calling apiClient.HandleDatabaseManagement")

			db, ok := lab.Unpack(harness, fixtures.PostgresFixture)
			assert.True(ok)

			// verify that the selectors were deleted successfully
			_, err = db.GetAssetGroupSelector(selector.ID)
			assert.ErrorIs(err, database.ErrNotFound)

		}),
	)
}
