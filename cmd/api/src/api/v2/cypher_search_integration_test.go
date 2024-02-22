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

	"github.com/specterops/bloodhound/cypher/frontend"
	"github.com/specterops/bloodhound/graphschema/common"
	"github.com/specterops/bloodhound/lab"
	v2 "github.com/specterops/bloodhound/src/api/v2"
	"github.com/specterops/bloodhound/src/test/lab/fixtures"
	"github.com/specterops/bloodhound/src/test/lab/harnesses"
	"github.com/stretchr/testify/require"
)

func Test_CypherSearch(t *testing.T) {
	var (
		harness = harnesses.NewIntegrationTestHarness(fixtures.BHAdminApiClientFixture)
	)

	lab.Pack(harness, fixtures.BasicComputerFixture)

	lab.NewSpec(t, harness).Run(
		lab.TestCase("errors on empty input", func(assert *require.Assertions, harness *lab.Harness) {
			apiClient, ok := lab.Unpack(harness, fixtures.BHAdminApiClientFixture)
			assert.True(ok)

			_, err := apiClient.CypherSearch(v2.CypherSearch{})
			assert.ErrorContains(err, frontend.ErrInvalidInput.Error())
		}),
		lab.TestCase("errors on syntax mistake", func(assert *require.Assertions, harness *lab.Harness) {
			apiClient, ok := lab.Unpack(harness, fixtures.BHAdminApiClientFixture)
			assert.True(ok)

			_, err := apiClient.CypherSearch(v2.CypherSearch{
				Query: "my syntax stinks",
			})
			assert.ErrorContains(err, "extraneous input")
		}),
		lab.TestCase("errors on queries that are not supported", func(assert *require.Assertions, harness *lab.Harness) {
			apiClient, ok := lab.Unpack(harness, fixtures.BHAdminApiClientFixture)
			assert.True(ok)

			queryWithUpdateClause := "match (b) where b.name = 'test' remove b.prop return b"
			_, err := apiClient.CypherSearch(v2.CypherSearch{
				Query: queryWithUpdateClause,
			})
			assert.ErrorContains(err, frontend.ErrUpdateClauseNotSupported.Error())
		}),
		lab.TestCase("succesfully runs cypher query", func(assert *require.Assertions, harness *lab.Harness) {
			apiClient, ok := lab.Unpack(harness, fixtures.BHAdminApiClientFixture)
			assert.True(ok)

			graphResponse, err := apiClient.CypherSearch(v2.CypherSearch{
				Query: "match (n:Computer) where n.objectid = '" + fixtures.BasicComputerSID.String() + "' return n",
			})
			assert.NoError(err)
			assert.Equal(1, len(graphResponse.Nodes))
			assert.Equal(0, len(graphResponse.Edges))

			expectedComputer, ok := lab.Unpack(harness, fixtures.BasicComputerFixture)
			assert.True(ok)

			expectedComputerId, _ := expectedComputer.Properties.Get(common.ObjectID.String()).String()
			actualComputerId := graphResponse.Nodes[expectedComputer.ID.String()].ObjectId
			assert.Equal(expectedComputerId, actualComputerId)
		}),
	)
}
