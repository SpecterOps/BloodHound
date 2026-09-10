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

package database_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/specterops/bloodhound/cmd/api/src/database"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/cmd/api/src/test/integration"
)

func TestSavedQueries_ListSavedQueries(t *testing.T) {
	var (
		testCtx = context.Background()
		dbInst  = integration.SetupDB(t)

		savedQueriesFilter = model.QueryParameterFilter{
			Name:         "id",
			Operator:     model.GreaterThan,
			Value:        "4",
			IsStringData: false,
		}
		savedQueriesFilterMap = model.QueryParameterFilterMap{savedQueriesFilter.Name: model.QueryParameterFilters{savedQueriesFilter}}
	)

	userUUID, err := uuid.NewV4()
	require.Nil(t, err)

	for i := 0; i < 7; i++ {
		if _, err := dbInst.CreateSavedQuery(testCtx, userUUID, fmt.Sprintf("saved_query_%d", i), "", "", nil, nil, ""); err != nil {
			t.Fatalf("Error creating audit log: %v", err)
		}
	}

	// no filtering - expect all queries
	if _, count, err := dbInst.ListSavedQueries(testCtx, string(model.SavedQueryScopeOwned), userUUID, "", model.SQLFilter{}, 0, 10); err != nil {
		t.Fatalf("Failed to list all saved queries: %v", err)
	} else if count != 7 {
		t.Fatalf("Expected 7 saved queries to be returned, received %d", count)
	} else if filter, err := savedQueriesFilterMap.BuildSQLFilter(); err != nil {
		t.Fatalf("Failed to generate SQL Filter: %v", err)
		// Limit is set to 1 to verify that count is total filtered count, not response size
	} else if _, count, err = dbInst.ListSavedQueries(testCtx, string(model.SavedQueryScopeOwned), userUUID, "", filter, 0, 1); err != nil {
		t.Fatalf("Failed to list filtered saved queries: %v", err)
	} else if count != 3 {
		t.Fatalf("Expected 3 saved queries to be returned, received %d", count)
	}
}

func assertSavedQueryConstraintError(t *testing.T, err error, expectedConstraint string) {
	t.Helper()

	var pgErr *pgconn.PgError
	require.Error(t, err)
	require.True(t, errors.As(err, &pgErr), "expected wrapped *pgconn.PgError, got %T: %v", err, err)
	assert.Equal(t, expectedConstraint, pgErr.ConstraintName)
}

func TestSavedQueries_CreateSavedQuery(t *testing.T) {
	t.Parallel()

	var (
		suite          = setupIntegrationTestSuite(t)
		userUUID, uErr = uuid.NewV4()
	)
	t.Cleanup(func() {
		teardownIntegrationTestSuite(t, &suite)
	})
	require.NoError(t, uErr)

	type testSetupData struct {
		name              string
		savedQueryID      int64
		query             string
		description       string
		schemaExtensionID *int32
		queryKey          *string
		category          string
	}
	type testCase struct {
		name     string
		setup    func(t *testing.T, suite IntegrationTestSuite) testSetupData
		assert   func(t *testing.T, suite IntegrationTestSuite, setupData testSetupData) model.SavedQuery
		teardown func(t *testing.T, suite IntegrationTestSuite, setupData testSetupData, created model.SavedQuery)
	}

	stringPtr := func(value string) *string {
		return &value
	}

	tests := []testCase{
		{
			name: "success_-_create_extension_query",
			setup: func(t *testing.T, suite IntegrationTestSuite) testSetupData {
				t.Helper()
				ext, err := suite.BHDatabase.CreateGraphSchemaExtension(suite.Context, "CreateSQPersistExt", "Create SQ Persist Ext", "v1.0.0", "create_sq_persist_ns")
				require.NoError(t, err)
				return testSetupData{
					name:              "ext_query",
					query:             "MATCH (n) RETURN n",
					description:       "desc",
					schemaExtensionID: &ext.ID,
					queryKey:          stringPtr("ext_query_key"),
					category:          "extension",
				}
			},
			assert: func(t *testing.T, suite IntegrationTestSuite, setupData testSetupData) model.SavedQuery {
				t.Helper()
				created, err := suite.BHDatabase.CreateSavedQuery(suite.Context, uuid.Nil, setupData.name, setupData.query, setupData.description, setupData.schemaExtensionID, setupData.queryKey, setupData.category)
				require.NoError(t, err)
				assert.Equal(t, setupData.schemaExtensionID, created.SchemaExtensionID)
				assert.Equal(t, setupData.queryKey, created.QueryKey)
				assert.Equal(t, setupData.category, created.Category)

				fetched, err := suite.BHDatabase.GetSavedQuery(suite.Context, created.ID)
				require.NoError(t, err)
				assert.Equal(t, setupData.schemaExtensionID, fetched.SchemaExtensionID)
				assert.Equal(t, setupData.queryKey, fetched.QueryKey)
				assert.Equal(t, setupData.category, fetched.Category)
				return created
			},
			teardown: func(t *testing.T, suite IntegrationTestSuite, setupData testSetupData, created model.SavedQuery) {
				t.Helper()
				require.NoError(t, suite.BHDatabase.DeleteGraphSchemaExtension(suite.Context, *setupData.schemaExtensionID))
			},
		},
		{
			name: "success_-_create_user_query",
			setup: func(t *testing.T, suite IntegrationTestSuite) testSetupData {
				t.Helper()
				return testSetupData{
					name:        "user_query",
					query:       "MATCH (n) RETURN n",
					description: "desc",
					category:    "",
				}
			},
			assert: func(t *testing.T, suite IntegrationTestSuite, setupData testSetupData) model.SavedQuery {
				t.Helper()
				created, err := suite.BHDatabase.CreateSavedQuery(suite.Context, userUUID, setupData.name, setupData.query, setupData.description, setupData.schemaExtensionID, setupData.queryKey, setupData.category)
				require.NoError(t, err)
				assert.Nil(t, created.SchemaExtensionID)
				assert.Nil(t, created.QueryKey)
				assert.Equal(t, setupData.category, created.Category)

				fetched, err := suite.BHDatabase.GetSavedQuery(suite.Context, created.ID)
				require.NoError(t, err)
				assert.Nil(t, fetched.SchemaExtensionID)
				assert.Nil(t, fetched.QueryKey)
				assert.Equal(t, setupData.category, fetched.Category)
				return created
			},
			teardown: func(t *testing.T, suite IntegrationTestSuite, setupData testSetupData, created model.SavedQuery) {
				t.Helper()
				require.NoError(t, suite.BHDatabase.DeleteSavedQuery(suite.Context, created.ID))
			},
		},
		{
			name: "success_-_deleting_extension_cascades_to_saved_query",
			setup: func(t *testing.T, suite IntegrationTestSuite) testSetupData {
				t.Helper()
				ext, err := suite.BHDatabase.CreateGraphSchemaExtension(suite.Context, "CascadeExt", "Cascade Ext", "v1.0.0", "cascade_ns")
				require.NoError(t, err)
				return testSetupData{
					name:              "cascade_query",
					query:             "MATCH (n) RETURN n",
					description:       "desc",
					schemaExtensionID: &ext.ID,
					queryKey:          stringPtr("cascade_query_key"),
				}
			},
			assert: func(t *testing.T, suite IntegrationTestSuite, setupData testSetupData) model.SavedQuery {
				t.Helper()
				created, err := suite.BHDatabase.CreateSavedQuery(suite.Context, uuid.Nil, setupData.name, setupData.query, setupData.description, setupData.schemaExtensionID, setupData.queryKey, setupData.category)
				require.NoError(t, err)
				require.NoError(t, suite.BHDatabase.DeleteGraphSchemaExtension(suite.Context, *setupData.schemaExtensionID))

				_, err = suite.BHDatabase.GetSavedQuery(suite.Context, created.ID)
				assert.ErrorIs(t, err, database.ErrNotFound)
				return created
			},
			teardown: func(t *testing.T, suite IntegrationTestSuite, setupData testSetupData, created model.SavedQuery) {
				t.Helper()
				// The extension (and its cascaded saved query) is already removed by the assert step.
			},
		},
		{
			name: "error_-_extension_set_but_query_key_nil",
			setup: func(t *testing.T, suite IntegrationTestSuite) testSetupData {
				t.Helper()
				ext, err := suite.BHDatabase.CreateGraphSchemaExtension(suite.Context, "ExtNoKeyExt", "Ext No Key Ext", "v1.0.0", "ext_no_key_ns")
				require.NoError(t, err)
				return testSetupData{
					name:              "ext_no_key",
					query:             "MATCH (n) RETURN n",
					description:       "desc",
					schemaExtensionID: &ext.ID,
				}
			},
			assert: func(t *testing.T, suite IntegrationTestSuite, setupData testSetupData) model.SavedQuery {
				t.Helper()
				created, err := suite.BHDatabase.CreateSavedQuery(suite.Context, uuid.Nil, setupData.name, setupData.query, setupData.description, setupData.schemaExtensionID, setupData.queryKey, setupData.category)
				assertSavedQueryConstraintError(t, err, "chk_saved_queries_extension_shape")
				return created
			},
			teardown: func(t *testing.T, suite IntegrationTestSuite, setupData testSetupData, created model.SavedQuery) {
				t.Helper()
				require.NoError(t, suite.BHDatabase.DeleteGraphSchemaExtension(suite.Context, *setupData.schemaExtensionID))
			},
		},
		{
			name: "error_-_query_key_set_but_extension_nil",
			setup: func(t *testing.T, suite IntegrationTestSuite) testSetupData {
				t.Helper()
				return testSetupData{
					name:        "key_no_ext",
					query:       "MATCH (n) RETURN n",
					description: "desc",
					queryKey:    stringPtr("k"),
				}
			},
			assert: func(t *testing.T, suite IntegrationTestSuite, setupData testSetupData) model.SavedQuery {
				t.Helper()
				created, err := suite.BHDatabase.CreateSavedQuery(suite.Context, userUUID, setupData.name, setupData.query, setupData.description, setupData.schemaExtensionID, setupData.queryKey, setupData.category)
				assertSavedQueryConstraintError(t, err, "chk_saved_queries_extension_shape")
				return created
			},
			teardown: func(t *testing.T, suite IntegrationTestSuite, setupData testSetupData, created model.SavedQuery) {
				t.Helper()
				// Nothing was persisted: the insert is rejected and no extension is created.
			},
		},
		{
			name: "error_-_duplicate_user_saved_query_name",
			setup: func(t *testing.T, suite IntegrationTestSuite) testSetupData {
				t.Helper()
				name := "duplicate_user_name"
				savedQuery, err := suite.BHDatabase.CreateSavedQuery(suite.Context, userUUID, name, "MATCH (n) RETURN n", "desc", nil, nil, "")
				require.NoError(t, err)
				return testSetupData{
					name:         name,
					savedQueryID: savedQuery.ID,
					query:        "MATCH (n) RETURN n",
					description:  "desc",
				}
			},
			assert: func(t *testing.T, suite IntegrationTestSuite, setupData testSetupData) model.SavedQuery {
				t.Helper()
				created, err := suite.BHDatabase.CreateSavedQuery(suite.Context, userUUID, setupData.name, setupData.query, setupData.description, nil, nil, setupData.category)
				assertSavedQueryConstraintError(t, err, "idx_saved_queries_user_id_name")
				return created
			},
			teardown: func(t *testing.T, suite IntegrationTestSuite, setupData testSetupData, created model.SavedQuery) {
				t.Helper()
				require.NoError(t, suite.BHDatabase.DeleteSavedQuery(suite.Context, setupData.savedQueryID))
			},
		},
		{
			name: "error_-_duplicate_extension_saved_query_name",
			setup: func(t *testing.T, suite IntegrationTestSuite) testSetupData {
				t.Helper()
				ext, err := suite.BHDatabase.CreateGraphSchemaExtension(suite.Context, "DupNameExt", "Dup Name Ext", "v1.0.0", "dup_name_ns")
				require.NoError(t, err)

				name := "duplicate_extension_name"
				_, err = suite.BHDatabase.CreateSavedQuery(suite.Context, uuid.Nil, name, "MATCH (n) RETURN n", "desc", &ext.ID, stringPtr("first"), "")
				require.NoError(t, err)

				return testSetupData{
					name:              name,
					query:             "MATCH (n) RETURN n",
					description:       "desc",
					schemaExtensionID: &ext.ID,
					queryKey:          stringPtr("second"),
				}
			},
			assert: func(t *testing.T, suite IntegrationTestSuite, setupData testSetupData) model.SavedQuery {
				t.Helper()
				created, err := suite.BHDatabase.CreateSavedQuery(suite.Context, uuid.Nil, setupData.name, setupData.query, setupData.description, setupData.schemaExtensionID, setupData.queryKey, setupData.category)
				assertSavedQueryConstraintError(t, err, "idx_saved_queries_schema_extension_id_name")
				return created
			},
			teardown: func(t *testing.T, suite IntegrationTestSuite, setupData testSetupData, created model.SavedQuery) {
				t.Helper()
				require.NoError(t, suite.BHDatabase.DeleteGraphSchemaExtension(suite.Context, *setupData.schemaExtensionID))
			},
		},
		{
			name: "error_-_duplicate_extension_query_key",
			setup: func(t *testing.T, suite IntegrationTestSuite) testSetupData {
				t.Helper()
				ext, err := suite.BHDatabase.CreateGraphSchemaExtension(suite.Context, "DupKeyExt", "Dup Key Ext", "v1.0.0", "dup_key_ns")
				require.NoError(t, err)

				_, err = suite.BHDatabase.CreateSavedQuery(suite.Context, uuid.Nil, "dup_first", "MATCH (n) RETURN n", "desc", &ext.ID, stringPtr("dup"), "")
				require.NoError(t, err)

				return testSetupData{
					name:              "dup_second",
					query:             "MATCH (n) RETURN n",
					description:       "desc",
					schemaExtensionID: &ext.ID,
					queryKey:          stringPtr("dup"),
				}
			},
			assert: func(t *testing.T, suite IntegrationTestSuite, setupData testSetupData) model.SavedQuery {
				t.Helper()
				created, err := suite.BHDatabase.CreateSavedQuery(suite.Context, uuid.Nil, setupData.name, setupData.query, setupData.description, setupData.schemaExtensionID, setupData.queryKey, setupData.category)
				assertSavedQueryConstraintError(t, err, "idx_saved_queries_extension_query_key")
				return created
			},
			teardown: func(t *testing.T, suite IntegrationTestSuite, setupData testSetupData, created model.SavedQuery) {
				t.Helper()
				require.NoError(t, suite.BHDatabase.DeleteGraphSchemaExtension(suite.Context, *setupData.schemaExtensionID))
			},
		},
		{
			name: "error_-_extension_with_real_user_id_rejected",
			setup: func(t *testing.T, suite IntegrationTestSuite) testSetupData {
				t.Helper()
				ext, err := suite.BHDatabase.CreateGraphSchemaExtension(suite.Context, "ExtRealUserExt", "Ext Real User Ext", "v1.0.0", "ext_real_user_ns")
				require.NoError(t, err)
				return testSetupData{
					name:              "ext_real_user",
					query:             "MATCH (n) RETURN n",
					description:       "desc",
					schemaExtensionID: &ext.ID,
					queryKey:          stringPtr("ext_real_user_key"),
				}
			},
			assert: func(t *testing.T, suite IntegrationTestSuite, setupData testSetupData) model.SavedQuery {
				t.Helper()
				created, err := suite.BHDatabase.CreateSavedQuery(suite.Context, userUUID, setupData.name, setupData.query, setupData.description, setupData.schemaExtensionID, setupData.queryKey, setupData.category)
				assertSavedQueryConstraintError(t, err, "chk_saved_queries_extension_shape")
				return created
			},
			teardown: func(t *testing.T, suite IntegrationTestSuite, setupData testSetupData, created model.SavedQuery) {
				t.Helper()
				require.NoError(t, suite.BHDatabase.DeleteGraphSchemaExtension(suite.Context, *setupData.schemaExtensionID))
			},
		},
		{
			name: "error_-_system_owner_without_extension_rejected",
			setup: func(t *testing.T, suite IntegrationTestSuite) testSetupData {
				t.Helper()
				return testSetupData{
					name:        "system_owner_no_ext",
					query:       "MATCH (n) RETURN n",
					description: "desc",
				}
			},
			assert: func(t *testing.T, suite IntegrationTestSuite, setupData testSetupData) model.SavedQuery {
				t.Helper()
				created, err := suite.BHDatabase.CreateSavedQuery(suite.Context, uuid.Nil, setupData.name, setupData.query, setupData.description, setupData.schemaExtensionID, setupData.queryKey, setupData.category)
				assertSavedQueryConstraintError(t, err, "chk_saved_queries_extension_shape")
				return created
			},
			teardown: func(t *testing.T, suite IntegrationTestSuite, setupData testSetupData, created model.SavedQuery) {
				t.Helper()
				// Nothing was persisted: the insert is rejected by chk_saved_queries_extension_shape.
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			setupData := testCase.setup(t, suite)
			created := testCase.assert(t, suite, setupData)
			testCase.teardown(t, suite, setupData, created)
		})
	}
}
