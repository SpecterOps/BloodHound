// Copyright 2026 Specter Ops, Inc.
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
	"strings"
	"testing"

	"github.com/gofrs/uuid"
	"github.com/specterops/bloodhound/cmd/api/src/database"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func savedQuery(queryKey string, name string, query string, description string) model.SavedQueryInput {
	return model.SavedQueryInput{
		QueryKey:    queryKey,
		Name:        name,
		Query:       query,
		Description: description,
	}
}

func upsertExtensionSavedQueries(t *testing.T, testSuite IntegrationTestSuite, extensionName string, savedQueries ...model.SavedQueryInput) int32 {
	t.Helper()
	input := model.GraphExtensionInput{
		ExtensionInput:    model.ExtensionInput{Name: extensionName, DisplayName: "Saved Query Extension", Version: "1.0.0", Namespace: "SQ"},
		NodeKindsInput:    model.NodesInput{{Name: "SQ_Node"}},
		SavedQueriesInput: savedQueries,
	}
	_, err := testSuite.BHDatabase.UpsertOpenGraphExtension(testSuite.Context, input)
	require.NoError(t, err)
	extensions, _, err := testSuite.BHDatabase.GetGraphSchemaExtensions(testSuite.Context, model.Filters{"name": []model.Filter{{Operator: model.Equals, Value: extensionName, SetOperator: model.FilterAnd}}}, model.Sort{}, 0, 1)
	require.NoError(t, err)
	require.Len(t, extensions, 1)
	return extensions[0].ID
}

func requireExtensionSavedQueries(t *testing.T, testSuite IntegrationTestSuite, extensionID int32, expectedSavedQueries ...model.SavedQueryInput) map[string]model.SavedQuery {
	t.Helper()
	var (
		expectedSavedQueriesByKey = map[string]model.SavedQueryInput{}
		savedQueriesByKey         = map[string]model.SavedQuery{}
	)
	for _, expectedSavedQuery := range expectedSavedQueries {
		expectedSavedQueriesByKey[strings.ToLower(expectedSavedQuery.QueryKey)] = expectedSavedQuery
	}
	savedQueries, err := testSuite.BHDatabase.GetSavedQueriesByExtensionID(testSuite.Context, extensionID)
	require.NoError(t, err)
	require.Len(t, savedQueries, len(expectedSavedQueries))
	for _, savedQuery := range savedQueries {
		require.NotNil(t, savedQuery.QueryKey)
		normalizedQueryKey := strings.ToLower(*savedQuery.QueryKey)
		expectedSavedQuery, found := expectedSavedQueriesByKey[normalizedQueryKey]
		require.True(t, found)
		require.Equal(t, expectedSavedQuery.QueryKey, *savedQuery.QueryKey)
		require.Equal(t, uuid.Nil.String(), savedQuery.UserID)
		require.NotNil(t, savedQuery.SchemaExtensionID)
		require.Equal(t, extensionID, *savedQuery.SchemaExtensionID)
		require.Equal(t, expectedSavedQuery.Name, savedQuery.Name)
		require.Equal(t, expectedSavedQuery.Query, savedQuery.Query)
		require.Equal(t, expectedSavedQuery.Description, savedQuery.Description)
		isPublic, err := testSuite.BHDatabase.IsSavedQueryPublic(testSuite.Context, savedQuery.ID)
		require.NoError(t, err)
		require.True(t, isPublic)
		savedQueriesByKey[normalizedQueryKey] = savedQuery
	}
	return savedQueriesByKey
}

func requireSavedQueriesDeleted(t *testing.T, testSuite IntegrationTestSuite, savedQueryIDs ...int64) {
	t.Helper()
	for _, savedQueryID := range savedQueryIDs {
		_, err := testSuite.BHDatabase.GetSavedQuery(testSuite.Context, savedQueryID)
		require.ErrorIs(t, err, database.ErrNotFound)
	}
}

func TestBloodhoundDB_UpsertOpenGraphExtensionSavedQueries(t *testing.T) {
	type testSetupData struct {
		graphExtensionInput       model.GraphExtensionInput
		extensionID               int32
		expectedSavedQueries      model.SavedQueriesInput
		existingSavedQueriesByKey map[string]model.SavedQuery
		deletedSavedQueryIDs      []int64
	}
	type testCase struct {
		name     string
		setup    func(t *testing.T, testSuite IntegrationTestSuite) testSetupData
		assert   func(t *testing.T, testSuite IntegrationTestSuite, setupData *testSetupData, updated bool, err error)
		teardown func(t *testing.T, testSuite IntegrationTestSuite, setupData testSetupData)
	}
	graphExtensionWithSavedQueries := func(extensionName string, savedQueries model.SavedQueriesInput) model.GraphExtensionInput {
		return model.GraphExtensionInput{
			ExtensionInput:    model.ExtensionInput{Name: extensionName, DisplayName: "Saved Query Extension", Version: "1.0.0", Namespace: "SQ"},
			NodeKindsInput:    model.NodesInput{{Name: "SQ_Node"}},
			SavedQueriesInput: savedQueries,
		}
	}
	teardownExtension := func(t *testing.T, testSuite IntegrationTestSuite, setupData testSetupData) {
		t.Helper()
		if setupData.extensionID != 0 {
			assert.NoError(t, testSuite.BHDatabase.DeleteGraphSchemaExtension(testSuite.Context, setupData.extensionID))
		}
	}
	t.Parallel()
	testSuite := setupIntegrationTestSuite(t)
	defer teardownIntegrationTestSuite(t, &testSuite)
	tests := []testCase{
		{
			name: "Create",
			setup: func(t *testing.T, testSuite IntegrationTestSuite) testSetupData {
				t.Helper()
				expectedSavedQueries := model.SavedQueriesInput{
					savedQuery("one", "One", "RETURN 1", "first"),
					savedQuery("two", "Two", "RETURN 2", "second"),
				}
				return testSetupData{
					graphExtensionInput:  graphExtensionWithSavedQueries("SavedQueryReconcileCreate", expectedSavedQueries),
					expectedSavedQueries: expectedSavedQueries,
				}
			},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, setupData *testSetupData, updated bool, err error) {
				t.Helper()
				require.NoError(t, err)
				assert.False(t, updated)
				graphSchemaExtensions, _, lookupErr := testSuite.BHDatabase.GetGraphSchemaExtensions(
					testSuite.Context,
					model.Filters{"name": []model.Filter{{Operator: model.Equals, Value: setupData.graphExtensionInput.ExtensionInput.Name, SetOperator: model.FilterAnd}}},
					model.Sort{},
					0,
					1,
				)
				require.NoError(t, lookupErr)
				require.Len(t, graphSchemaExtensions, 1)
				setupData.extensionID = graphSchemaExtensions[0].ID
				requireExtensionSavedQueries(t, testSuite, setupData.extensionID, setupData.expectedSavedQueries...)
			},
			teardown: teardownExtension,
		},
		{
			name: "Update",
			setup: func(t *testing.T, testSuite IntegrationTestSuite) testSetupData {
				t.Helper()
				var (
					extensionName             = "SavedQueryReconcileUpdate"
					initialSavedQueries       = model.SavedQueriesInput{savedQuery("keep", "Keep", "RETURN 1", "old")}
					expectedSavedQueries      = model.SavedQueriesInput{savedQuery("keep", "Updated", "RETURN 2", "new")}
					updatedSavedQueries       = model.SavedQueriesInput{savedQuery("KEEP", "Updated", "RETURN 2", "new")}
					extensionID               = upsertExtensionSavedQueries(t, testSuite, extensionName, initialSavedQueries...)
					existingSavedQueriesByKey = requireExtensionSavedQueries(t, testSuite, extensionID, initialSavedQueries...)
				)
				return testSetupData{
					graphExtensionInput:       graphExtensionWithSavedQueries(extensionName, updatedSavedQueries),
					extensionID:               extensionID,
					expectedSavedQueries:      expectedSavedQueries,
					existingSavedQueriesByKey: existingSavedQueriesByKey,
				}
			},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, setupData *testSetupData, updated bool, err error) {
				t.Helper()
				require.NoError(t, err)
				assert.True(t, updated)
				currentSavedQueriesByKey := requireExtensionSavedQueries(t, testSuite, setupData.extensionID, setupData.expectedSavedQueries...)
				require.Equal(t, setupData.existingSavedQueriesByKey["keep"].ID, currentSavedQueriesByKey["keep"].ID)
			},
			teardown: teardownExtension,
		},
		{
			name: "Delete",
			setup: func(t *testing.T, testSuite IntegrationTestSuite) testSetupData {
				t.Helper()
				var (
					extensionName        = "SavedQueryReconcileDelete"
					initialSavedQueries  = model.SavedQueriesInput{savedQuery("keep", "Keep", "RETURN 1", "keep"), savedQuery("drop", "Drop", "RETURN 2", "drop")}
					expectedSavedQueries = model.SavedQueriesInput{savedQuery("keep", "Keep", "RETURN 1", "keep")}
					extensionID          = upsertExtensionSavedQueries(t, testSuite, extensionName, initialSavedQueries...)
					savedQueriesByKey    = requireExtensionSavedQueries(t, testSuite, extensionID, initialSavedQueries...)
				)
				return testSetupData{
					graphExtensionInput:  graphExtensionWithSavedQueries(extensionName, expectedSavedQueries),
					extensionID:          extensionID,
					expectedSavedQueries: expectedSavedQueries,
					deletedSavedQueryIDs: []int64{savedQueriesByKey["drop"].ID},
				}
			},
			assert: func(t *testing.T, testSuite IntegrationTestSuite, setupData *testSetupData, updated bool, err error) {
				t.Helper()
				require.NoError(t, err)
				assert.True(t, updated)
				requireExtensionSavedQueries(t, testSuite, setupData.extensionID, setupData.expectedSavedQueries...)
				requireSavedQueriesDeleted(t, testSuite, setupData.deletedSavedQueryIDs...)
			},
			teardown: teardownExtension,
		},
	}
	for _, currentTestCase := range tests {
		t.Run(currentTestCase.name, func(t *testing.T) {
			setupData := currentTestCase.setup(t, testSuite)
			if currentTestCase.teardown != nil {
				t.Cleanup(func() {
					currentTestCase.teardown(t, testSuite, setupData)
				})
			}

			updated, err := testSuite.BHDatabase.UpsertOpenGraphExtension(testSuite.Context, setupData.graphExtensionInput)
			currentTestCase.assert(t, testSuite, &setupData, updated, err)
		})
	}
}
