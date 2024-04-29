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
	"github.com/specterops/bloodhound/src/test/integration/utils"
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

			queryWithUserSpecifiedParameters := "match (n:Guardian {name: $name}) return n"
			_, err := apiClient.CypherSearch(v2.CypherSearch{
				Query: queryWithUserSpecifiedParameters,
			})
			assert.ErrorContains(err, frontend.ErrUserSpecifiedParametersNotSupported.Error())
		}),
		lab.TestCase("successfully runs cypher query", func(assert *require.Assertions, harness *lab.Harness) {
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

func Test_CypherSearch_WithoutCypherMutationsEnabled(t *testing.T) {
	harness := lab.NewHarness()

	// Default ConfigFixture does not have cypher mutations enabled
	adminApiClientFixture := fixtures.NewAdminApiClientFixture(fixtures.ConfigFixture, fixtures.NewApiFixture())
	lab.Pack(harness, adminApiClientFixture)

	lab.NewSpec(t, harness).Run(
		lab.TestCase("errors on mutations with enable cypher_mutations false", func(assert *require.Assertions, harness *lab.Harness) {
			apiClient, ok := lab.Unpack(harness, adminApiClientFixture)
			assert.True(ok)

			_, err := apiClient.CypherSearch(v2.CypherSearch{Query: "match (w) where w.name = 'voldemort' remove w.name return w"})
			assert.ErrorContains(err, frontend.ErrUpdateClauseNotSupported.Error())
		}),
	)
}

func Test_CypherSearch_WithCypherMutationsEnabled(t *testing.T) {
	customCfg, err := utils.LoadIntegrationTestConfig()
	require.Nil(t, err)

	// Enable cypher mutations in API
	customCfg.EnableCypherMutations = true

	harness := lab.NewHarness()
	customCfgFixture := fixtures.NewCustomConfigFixture(customCfg)
	lab.Pack(harness, customCfgFixture)
	customApiFixture := fixtures.NewCustomApiFixture(customCfgFixture)
	lab.Pack(harness, customApiFixture)
	adminApiClientFixture := fixtures.NewAdminApiClientFixture(customCfgFixture, customApiFixture)
	lab.Pack(harness, adminApiClientFixture)

	lab.NewSpec(t, harness).Run(
		lab.TestCase("allows mutations with enable cypher_mutations true", func(assert *require.Assertions, harness *lab.Harness) {
			apiClient, ok := lab.Unpack(harness, adminApiClientFixture)
			assert.True(ok)

			_, err := apiClient.CypherSearch(v2.CypherSearch{Query: "match (w) where w.name = 'voldemort' remove w.name return w"})
			assert.Nil(err)
		}),
	)
}
