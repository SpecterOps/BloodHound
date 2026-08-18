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
	"fmt"
	"testing"

	"github.com/gofrs/uuid"
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
		if _, err := dbInst.CreateSavedQuery(testCtx, userUUID, fmt.Sprintf("saved_query_%d", i), "", "", 0); err != nil {
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

func TestSavedQueries_SchemaExtensionID(t *testing.T) {

	type testSetupData struct {
		created model.SavedQuery
		extID   int32
	}
	type testCase struct {
		name     string
		setup    func(t *testing.T, ctx context.Context, dbInst database.Database) testSetupData
		assert   func(t *testing.T, ctx context.Context, dbInst database.Database, data testSetupData)
		teardown func(t *testing.T, ctx context.Context, dbInst database.Database, data testSetupData)
	}

	var (
		testCtx       = context.Background()
		userUUID, err = uuid.NewV4()
	)
	require.NoError(t, err)

	tests := []testCase{
		{
			name: "success_-_extension_linked_query_persists_schema_extension_id",
			setup: func(t *testing.T, ctx context.Context, dbInst database.Database) testSetupData {
				t.Helper()
				ext, err := dbInst.CreateGraphSchemaExtension(ctx, "SavedQueryExt", "Saved Query Ext", "v1.0.0", "sqext_ns")
				require.NoError(t, err)
				created, err := dbInst.CreateSavedQuery(ctx, userUUID, "ext_query", "MATCH (n) RETURN n", "desc", ext.ID)
				require.NoError(t, err)
				return testSetupData{created: created, extID: ext.ID}
			},
			assert: func(t *testing.T, ctx context.Context, dbInst database.Database, data testSetupData) {
				t.Helper()
				assert.Equal(t, data.extID, data.created.SchemaExtensionID)
				fetched, err := dbInst.GetSavedQuery(ctx, data.created.ID)
				require.NoError(t, err)
				assert.Equal(t, data.extID, fetched.SchemaExtensionID)
			},
		},
		{
			name: "success_-_zero_schema_extension_id_persists_as_null",
			setup: func(t *testing.T, ctx context.Context, dbInst database.Database) testSetupData {
				t.Helper()
				created, err := dbInst.CreateSavedQuery(ctx, userUUID, "user_query", "MATCH (n) RETURN n", "desc", 0)
				require.NoError(t, err)
				return testSetupData{created: created}
			},
			assert: func(t *testing.T, ctx context.Context, dbInst database.Database, data testSetupData) {
				t.Helper()
				assert.Zero(t, data.created.SchemaExtensionID)
				fetched, err := dbInst.GetSavedQuery(ctx, data.created.ID)
				require.NoError(t, err)
				assert.Zero(t, fetched.SchemaExtensionID)
			},
		},
		{
			name: "success_-_deleting_extension_cascades_to_saved_query",
			setup: func(t *testing.T, ctx context.Context, dbInst database.Database) testSetupData {
				t.Helper()
				ext, err := dbInst.CreateGraphSchemaExtension(ctx, "CascadeExt", "Cascade Ext", "v1.0.0", "cascade_ns")
				require.NoError(t, err)
				created, err := dbInst.CreateSavedQuery(ctx, userUUID, "cascade_query", "MATCH (n) RETURN n", "desc", ext.ID)
				require.NoError(t, err)
				return testSetupData{created: created, extID: ext.ID}
			},
			assert: func(t *testing.T, ctx context.Context, dbInst database.Database, data testSetupData) {
				t.Helper()
				require.NoError(t, dbInst.DeleteGraphSchemaExtension(ctx, data.extID))
				_, err := dbInst.GetSavedQuery(ctx, data.created.ID)
				assert.ErrorIs(t, err, database.ErrNotFound)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			dbInst := integration.SetupDB(t)
			data := tc.setup(t, testCtx, dbInst)
			if tc.teardown != nil {
				defer tc.teardown(t, testCtx, dbInst, data)
			}
			tc.assert(t, testCtx, dbInst, data)
		})
	}
}
