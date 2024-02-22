// Copyright 2024 Specter Ops, Inc.
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
	"testing"

	"github.com/gofrs/uuid"
	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/lab"
	v2 "github.com/specterops/bloodhound/src/api/v2"
	"github.com/specterops/bloodhound/src/test/lab/fixtures"
	"github.com/specterops/bloodhound/src/test/lab/harnesses"
	"github.com/stretchr/testify/require"
)

func TestDatabaseWipe_CollectedGraphData(t *testing.T) {
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

			graphdb, ok := lab.Unpack(harness, fixtures.GraphDBFixture)
			assert.True(ok)

			// verify that nodes exist
			graphdb.ReadTransaction(context.Background(),
				func(tx graph.Transaction) error {
					nodeCount, err := tx.Nodes().Count()

					require.Nil(t, err)
					require.NotZero(t, nodeCount)
					return nil
				})

			err := apiClient.HandleDatabaseWipe(
				v2.DatabaseManagement{
					DeleteCollectedGraphData: true,
				})
			assert.Nil(err, "error calling apiClient.HandleDatabaseWipe")

			// verify that zero nodes exist
			graphdb.ReadTransaction(context.Background(),
				func(tx graph.Transaction) error {
					nodeCount, err := tx.Nodes().Count()

					require.Nil(t, err)
					require.Zero(t, nodeCount)
					return nil
				})

		}),
	)

}
