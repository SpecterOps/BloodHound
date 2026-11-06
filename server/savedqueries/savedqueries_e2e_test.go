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

package savedqueries_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/gofrs/uuid"
	"github.com/specterops/bloodhound/cmd/api/src/api/router"
	"github.com/specterops/bloodhound/cmd/api/src/auth"
	"github.com/specterops/bloodhound/cmd/api/src/database"
	"github.com/specterops/bloodhound/cmd/api/src/database/types/null"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/server/internal/servertest"
	"github.com/specterops/bloodhound/server/savedqueries"
	"github.com/specterops/bloodhound/server/savedqueries/internal/handlers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetSavedQuery(t *testing.T) {
	type expected struct {
		responseCode int
		queryName    string
		extensionID  *int32
	}

	type testData struct {
		name         string
		token        string
		savedQueryID string
		expected     expected
	}

	var (
		ctx     = context.Background()
		harness = servertest.NewHarness(t, func(routerInst *router.Router, databaseInterface *database.BloodhoundDB) {
			savedqueries.Register(routerInst, databaseInterface.Pool())
		})
		databaseInterface = harness.DB
		server            = harness.Server
		ownerID           = uuid.Must(uuid.NewV4())
		readerID          = uuid.Must(uuid.NewV4())
		adminID           = uuid.Must(uuid.NewV4())
	)

	allRoles, err := databaseInterface.GetAllRoles(ctx, "", model.SQLFilter{})
	require.NoError(t, err)
	userRole, found := allRoles.FindByName(auth.RoleUser)
	require.True(t, found)

	ownerToken := servertest.MintJWT(t, ctx, databaseInterface, harness.Auther, model.User{
		Unique:        model.Unique{ID: ownerID},
		PrincipalName: "saved-query-owner@example.com",
		EmailAddress:  null.StringFrom("saved-query-owner@example.com"),
		EULAAccepted:  true,
		Roles:         model.Roles{userRole},
	})
	readerToken := servertest.MintJWT(t, ctx, databaseInterface, harness.Auther, model.User{
		Unique:        model.Unique{ID: readerID},
		PrincipalName: "saved-query-reader@example.com",
		EmailAddress:  null.StringFrom("saved-query-reader@example.com"),
		EULAAccepted:  true,
		Roles:         model.Roles{userRole},
	})
	adminToken := servertest.MintJWT(t, ctx, databaseInterface, harness.Auther, model.User{
		Unique:        model.Unique{ID: adminID},
		PrincipalName: "saved-query-admin@example.com",
		EmailAddress:  null.StringFrom("saved-query-admin@example.com"),
		EULAAccepted:  true,
		Roles:         model.Roles{servertest.AdminRole(t, ctx, databaseInterface)},
	})

	privateQuery, err := databaseInterface.CreateSavedQuery(ctx, ownerID, "Private query", "MATCH (n) RETURN n", "Private", nil, nil, "")
	require.NoError(t, err)
	sharedQuery, err := databaseInterface.CreateSavedQuery(ctx, ownerID, "Shared query", "MATCH (n) RETURN n", "Shared", nil, nil, "")
	require.NoError(t, err)
	_, err = databaseInterface.CreateSavedQueryPermissionsToUsers(ctx, sharedQuery.ID, readerID)
	require.NoError(t, err)
	publicQuery, err := databaseInterface.CreateSavedQuery(ctx, ownerID, "Public query", "MATCH (n) RETURN n", "Public", nil, nil, "")
	require.NoError(t, err)
	_, err = databaseInterface.CreateSavedQueryPermissionToPublic(ctx, publicQuery.ID)
	require.NoError(t, err)

	extension, err := databaseInterface.CreateGraphSchemaExtension(ctx, "SavedQueryE2E", "Saved Query E2E", "1.0.0", "saved_query_e2e")
	require.NoError(t, err)
	queryKey := "extension-query"
	extensionQuery, err := databaseInterface.CreateSavedQuery(ctx, uuid.Nil, "Extension query", "MATCH (n) RETURN n", "Extension", &extension.ID, &queryKey, "extension")
	require.NoError(t, err)
	_, err = databaseInterface.CreateSavedQueryPermissionToPublic(ctx, extensionQuery.ID)
	require.NoError(t, err)

	tt := []testData{
		{
			name:         "Success: owner reads private query - 200",
			token:        ownerToken,
			savedQueryID: fmt.Sprint(privateQuery.ID),
			expected:     expected{responseCode: http.StatusOK, queryName: privateQuery.Name},
		},
		{
			name:         "Success: administrator reads another user's private query - 200",
			token:        adminToken,
			savedQueryID: fmt.Sprint(privateQuery.ID),
			expected:     expected{responseCode: http.StatusOK, queryName: privateQuery.Name},
		},
		{
			name:         "Success: recipient reads shared query - 200",
			token:        readerToken,
			savedQueryID: fmt.Sprint(sharedQuery.ID),
			expected:     expected{responseCode: http.StatusOK, queryName: sharedQuery.Name},
		},
		{
			name:         "Success: user reads public query - 200",
			token:        readerToken,
			savedQueryID: fmt.Sprint(publicQuery.ID),
			expected:     expected{responseCode: http.StatusOK, queryName: publicQuery.Name},
		},
		{
			name:         "Success: user reads extension-provided query - 200",
			token:        readerToken,
			savedQueryID: fmt.Sprint(extensionQuery.ID),
			expected:     expected{responseCode: http.StatusOK, queryName: extensionQuery.Name, extensionID: &extension.ID},
		},
		{
			name:         "Error: inaccessible private query is hidden - 404",
			token:        readerToken,
			savedQueryID: fmt.Sprint(privateQuery.ID),
			expected:     expected{responseCode: http.StatusNotFound},
		},
		{
			name:         "Error: missing query - 404",
			token:        readerToken,
			savedQueryID: "9223372036854775807",
			expected:     expected{responseCode: http.StatusNotFound},
		},
		{
			name:         "Error: malformed query ID - 400",
			token:        readerToken,
			savedQueryID: "invalid",
			expected:     expected{responseCode: http.StatusBadRequest},
		},
		{
			name:         "Error: unauthenticated request - 401",
			savedQueryID: fmt.Sprint(publicQuery.ID),
			expected:     expected{responseCode: http.StatusUnauthorized},
		},
	}

	for _, testCase := range tt {
		t.Run(testCase.name, func(t *testing.T) {
			request, err := http.NewRequestWithContext(ctx, http.MethodGet, server.URL+"/api/v2/saved-queries/"+testCase.savedQueryID, nil)
			require.NoError(t, err)
			if testCase.token != "" {
				request.Header.Set("Authorization", "Bearer "+testCase.token)
			}

			response, err := http.DefaultClient.Do(request)
			require.NoError(t, err)
			defer response.Body.Close()
			assert.Equal(t, testCase.expected.responseCode, response.StatusCode)

			if testCase.expected.queryName != "" {
				var envelope struct {
					Data handlers.SavedQueryView `json:"data"`
				}
				require.NoError(t, json.NewDecoder(response.Body).Decode(&envelope))
				assert.Equal(t, testCase.expected.queryName, envelope.Data.Name)
				assert.Equal(t, testCase.expected.extensionID, envelope.Data.SchemaExtensionID)
			}
		})
	}
}
