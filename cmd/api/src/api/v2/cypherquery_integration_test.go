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
	"bytes"
	"fmt"
	"github.com/specterops/bloodhound/cypher/backend/cypher"
	"github.com/specterops/bloodhound/src/auth"
	"github.com/specterops/bloodhound/src/model"
	"github.com/specterops/bloodhound/src/utils/test"
	"net/http"
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

			_, err := apiClient.CypherQuery(v2.CypherQueryPayload{})
			assert.ErrorContains(err, frontend.ErrInvalidInput.Error())
		}),
		lab.TestCase("errors on syntax mistake", func(assert *require.Assertions, harness *lab.Harness) {
			apiClient, ok := lab.Unpack(harness, fixtures.BHAdminApiClientFixture)
			assert.True(ok)

			_, err := apiClient.CypherQuery(v2.CypherQueryPayload{
				Query: "my syntax stinks",
			})
			assert.ErrorContains(err, "extraneous input")
		}),
		lab.TestCase("errors on queries that are not supported", func(assert *require.Assertions, harness *lab.Harness) {
			apiClient, ok := lab.Unpack(harness, fixtures.BHAdminApiClientFixture)
			assert.True(ok)

			queryWithUserSpecifiedParameters := "match (n:Guardian {name: $name}) return n"
			_, err := apiClient.CypherQuery(v2.CypherQueryPayload{
				Query: queryWithUserSpecifiedParameters,
			})
			assert.ErrorContains(err, frontend.ErrUserSpecifiedParametersNotSupported.Error())
		}),
		lab.TestCase("successfully runs cypher query", func(assert *require.Assertions, harness *lab.Harness) {
			apiClient, ok := lab.Unpack(harness, fixtures.BHAdminApiClientFixture)
			assert.True(ok)

			graphResponse, err := apiClient.CypherQuery(v2.CypherQueryPayload{
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

			_, err := apiClient.CypherQuery(v2.CypherQueryPayload{Query: "match (w) where w.name = 'voldemort' remove w.name return w"})
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
	userApiClientFixture := fixtures.NewUserApiClientFixture(customCfgFixture, adminApiClientFixture, auth.RoleUser)
	lab.Pack(harness, userApiClientFixture)

	var (
		parseCtx = frontend.NewContext(
			&frontend.ExplicitProcedureInvocationFilter{},
			&frontend.ImplicitProcedureInvocationFilter{},
			&frontend.SpecifiedParametersFilter{},
		)
		stripper = cypher.NewCypherEmitter(true)
	)

	lab.NewSpec(t, harness).Run(
		lab.TestCase("allows mutations with enable cypher_mutations true", func(assert *require.Assertions, harness *lab.Harness) {
			apiClient, ok := lab.Unpack(harness, adminApiClientFixture)
			assert.True(ok)

			var (
				query         = "match (w) where w.name = 'vldmrt' remove w.name return w"
				strippedQuery = &bytes.Buffer{}
			)
			parsedQuery, err := frontend.ParseCypher(parseCtx, query)
			require.Nil(t, err)
			err = stripper.Write(parsedQuery, strippedQuery)
			require.Nil(t, err)

			_, err = apiClient.CypherQuery(v2.CypherQueryPayload{Query: query})
			assert.Nil(err)

			auditLogs, err := apiClient.GetLatestAuditLogs()
			assert.Nil(err)

			err = test.AssertAuditLogs(auditLogs.Logs, model.AuditLogActionMutateGraph, model.AuditLogStatusSuccess, model.AuditData{"query": strippedQuery.String()})
			assert.Nil(err)
		}),

		lab.TestCase("fails unauthorized mutation attempts and adds it to audit log", func(assert *require.Assertions, harness *lab.Harness) {
			adminApiClient, ok := lab.Unpack(harness, adminApiClientFixture)
			assert.True(ok)
			userApiClient, ok := lab.Unpack(harness, userApiClientFixture)
			assert.True(ok)

			var (
				query         = "match (w) where w.name = 'harryp' delete w"
				strippedQuery = &bytes.Buffer{}
			)
			parsedQuery, err := frontend.ParseCypher(parseCtx, query)
			require.Nil(t, err)
			err = stripper.Write(parsedQuery, strippedQuery)
			require.Nil(t, err)

			_, err = userApiClient.CypherQuery(v2.CypherQueryPayload{Query: query})
			assert.ErrorContains(err, "Permission denied: User may not modify the graph")

			auditLogs, err := adminApiClient.GetLatestAuditLogs()
			assert.Nil(err)

			found := false
			for _, al := range auditLogs.Logs {
				if al.Action == model.AuditLogActionUnauthorizedAccessAttempt {
					found = true
				}
			}
			assert.True(found)
		}),

		lab.TestCase("adds failed mutation attempts to audit log", func(assert *require.Assertions, harness *lab.Harness) {
			apiClient, ok := lab.Unpack(harness, adminApiClientFixture)
			assert.True(ok)

			var (
				query         = "match (w) set w.wizard = true return w.wizard"
				strippedQuery = &bytes.Buffer{}
			)

			parsedQuery, err := frontend.ParseCypher(parseCtx, query)
			require.Nil(t, err)
			err = stripper.Write(parsedQuery, strippedQuery)
			require.Nil(t, err)

			_, err = apiClient.CypherQuery(v2.CypherQueryPayload{Query: query})
			assert.ErrorContains(err, fmt.Sprintf("%d", http.StatusInternalServerError))

			auditLogs, err := apiClient.GetLatestAuditLogs()
			assert.Nil(err)

			err = test.AssertAuditLogs(auditLogs.Logs, model.AuditLogActionMutateGraph, model.AuditLogStatusFailure, model.AuditData{"query": strippedQuery.String()})
			assert.Nil(err)
		}),
	)
}
