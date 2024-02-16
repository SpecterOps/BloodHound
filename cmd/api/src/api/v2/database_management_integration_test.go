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
	"context"
	"fmt"
	"testing"

	"github.com/gofrs/uuid"
	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/lab"
	v2 "github.com/specterops/bloodhound/src/api/v2"
	"github.com/specterops/bloodhound/src/database"
	"github.com/specterops/bloodhound/src/model"
	"github.com/specterops/bloodhound/src/test/lab/fixtures"
	"github.com/specterops/bloodhound/src/test/lab/harnesses"
	"github.com/stretchr/testify/require"
)

// TODO: the following test seems to infinitely hang when using the new transactional fixture stuff.  check this out with big D
func Test_DatabaseManagement_FileUploadHistory(t *testing.T) {
	var (
		harness           = harnesses.NewIntegrationTestHarness(fixtures.BHAdminApiClientFixture)
		userFixture       *lab.Fixture[*model.User]
		fileUploadFixture *lab.Fixture[*model.FileUploadJobs]
	)

	fixtures.TransactionalFixtures(
		func(db *lab.Fixture[*database.BloodhoundDB]) lab.Depender {
			// create a user
			userFixture = fixtures.NewUserFixture(db)
			return userFixture
		},
		func(db *lab.Fixture[*database.BloodhoundDB]) lab.Depender {
			// create some file upload jobs. file upload jobs have a fk constraint on a user
			fileUploadFixture = fixtures.NewFileUploadFixture(db, userFixture)
			return fileUploadFixture
		},
	)

	lab.Pack(harness, fileUploadFixture)

	lab.NewSpec(t, harness).Run(
		lab.TestCase("the endpoint can delete file ingest history", func(assert *require.Assertions, harness *lab.Harness) {
			apiClient, ok := lab.Unpack(harness, fixtures.BHAdminApiClientFixture)
			assert.True(ok)

			err := apiClient.HandleDatabaseManagement(
				v2.DatabaseManagement{
					FileIngestHistory: true,
				})
			assert.Nil(err, "error calling apiClient.HandleDatabaseManagement")

			db, ok := lab.Unpack(harness, fixtures.PostgresFixture)
			assert.True(ok)

			_, numJobs, err := db.GetAllFileUploadJobs(0, 0, "", model.SQLFilter{})
			assert.Nil(err)
			assert.Zero(numJobs)

			// actual, _ := db.DeleteAllFileUploads()
		}),
	)
}

// TODO: same here.
func Test_DatabaseManagement_AssetGroupSelectors(t *testing.T) {
	var (
		harness    = harnesses.NewIntegrationTestHarness(fixtures.BHAdminApiClientFixture)
		assetGroup *lab.Fixture[*model.AssetGroup]
		selector1  *lab.Fixture[*model.AssetGroupSelector]
		selector2  *lab.Fixture[*model.AssetGroupSelector]
	)

	fixtures.TransactionalFixtures(
		func(db *lab.Fixture[*database.BloodhoundDB]) lab.Depender {
			assetGroup = fixtures.NewAssetGroupFixture(db, "mycoolassetgroup", "customtag", false)
			return assetGroup
		}, func(db *lab.Fixture[*database.BloodhoundDB]) lab.Depender {
			selector1 = fixtures.NewAssetGroupSelectorFixture(db, assetGroup, "mycoolassetgroupselector1", "someobjectid")
			return selector1
		},
		func(db *lab.Fixture[*database.BloodhoundDB]) lab.Depender {
			selector2 = fixtures.NewAssetGroupSelectorFixture(db, assetGroup, "mycoolassetgroupselector2", "someobjectid")
			return selector2
		},
	)

	// packing `assetGroupSelector` packs all the things it depends on too
	lab.Pack(harness, selector1)
	lab.Pack(harness, selector2)

	lab.NewSpec(t, harness).Run(
		lab.TestCase("the endpoint can delete asset group selectors", func(assert *require.Assertions, harness *lab.Harness) {
			apiClient, ok := lab.Unpack(harness, fixtures.BHAdminApiClientFixture)
			assert.True(ok)

			// selector, ok := lab.Unpack(harness, selector1)
			// assert.True(ok, "unable to unpack asset group selector")
			assetGroup, ok := lab.Unpack(harness, assetGroup)
			assert.True(ok)

			err := apiClient.HandleDatabaseManagement(
				v2.DatabaseManagement{
					HighValueSelectors: true,
					// AssetGroupId:       int(selector.AssetGroupID),
					AssetGroupId: int(assetGroup.ID),
				})
			assert.Nil(err, "error calling apiClient.HandleDatabaseManagement")

			db, ok := lab.Unpack(harness, fixtures.PostgresFixture)
			assert.True(ok)

			ag, err := db.GetAssetGroup(assetGroup.ID)
			assert.Nil(err)
			fmt.Println(ag)

			// actual, _ := db.GetAssetGroupSelector(selector.ID)
			// expected := model.AssetGroupSelector{}
			// when selector is not found in the db, `db.GetAssetGroupSelector` returns an empty struct.
			// assert.Equal(expected, actual)
		}),
	)
}

func TestDatabaseManagement_CollectedGraphData(t *testing.T) {
	var (
		harness       = harnesses.NewIntegrationTestHarness(fixtures.BHAdminApiClientFixture)
		graphdb       = fixtures.NewGraphDBFixture()
		computer1UUID = uuid.Must(uuid.NewV4())
		computer1     = fixtures.NewComputerFixture(computer1UUID, "computer 1", fixtures.BasicDomainFixture)
		computer2UUID = uuid.Must(uuid.NewV4())
		computer2     = fixtures.NewComputerFixture(computer2UUID, "computer 2", fixtures.BasicDomainFixture)
		computer3UUID = uuid.Must(uuid.NewV4())
		computer3     = fixtures.NewComputerFixture(computer3UUID, "computer 3", fixtures.BasicDomainFixture)
	)

	lab.Pack(harness, graphdb)
	lab.Pack(harness, computer1)
	lab.Pack(harness, computer2)
	lab.Pack(harness, computer3)

	lab.NewSpec(t, harness).Run(
		lab.TestCase("the endpoint deletes graph data", func(assert *require.Assertions, harness *lab.Harness) {
			apiClient, ok := lab.Unpack(harness, fixtures.BHAdminApiClientFixture)
			assert.True(ok)

			err := apiClient.HandleDatabaseManagement(
				v2.DatabaseManagement{
					CollectedGraphData: true,
				})
			assert.Nil(err, "error calling apiClient.HandleDatabaseManagement")

			graphdb, ok := lab.Unpack(harness, fixtures.GraphDBFixture)
			assert.True(ok)

			graphdb.ReadTransaction(context.Background(),
				func(tx graph.Transaction) error {

					return nil
				})

		}),
	)

}
