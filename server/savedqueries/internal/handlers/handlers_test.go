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

package handlers_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofrs/uuid"
	"github.com/gorilla/mux"
	"github.com/specterops/bloodhound/cmd/api/src/api"
	"github.com/specterops/bloodhound/packages/go/responses"
	"github.com/specterops/bloodhound/server/savedqueries/internal/handlers"
	"github.com/specterops/bloodhound/server/savedqueries/internal/handlers/mocks"
	"github.com/specterops/bloodhound/server/savedqueries/internal/services"
	"github.com/specterops/bloodhound/server/users"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandlers_GetSavedQuery(t *testing.T) {
	type mock struct {
		savedQueries *mocks.MockSavedQueries
		userResolver *mocks.MockUserResolver
	}

	type expected struct {
		responseCode int
		responseBody string
		errorMessage string
	}

	type testData struct {
		name         string
		buildRequest func(*testing.T) *http.Request
		setupMocks   func(mock, context.Context)
		expected     expected
	}

	var (
		userID            = uuid.Must(uuid.NewV4())
		adminID           = uuid.Must(uuid.NewV4())
		schemaExtensionID = int32(7)
		savedQuery        = services.SavedQuery{
			ID:                42,
			UserID:            uuid.Nil.String(),
			Name:              "Extension query",
			Query:             "MATCH (n) RETURN n",
			Description:       "Provided by an extension",
			Category:          "extension",
			SchemaExtensionID: &schemaExtensionID,
			CreatedAt:         time.Date(2026, 11, 1, 12, 0, 0, 0, time.UTC),
			UpdatedAt:         time.Date(2026, 11, 2, 12, 0, 0, 0, time.UTC),
		}
		unexpectedErr = errors.New("database unavailable")
	)

	buildRequest := func(t *testing.T, pathID string) *http.Request {
		t.Helper()
		request := httptest.NewRequest(http.MethodGet, "/api/v2/saved-queries/"+pathID, nil)
		request = mux.SetURLVars(request, map[string]string{api.URIPathVariableSavedQueryID: pathID})
		return request
	}

	tt := []testData{
		{
			name: "Success: user query omits extension-only fields - 200",
			buildRequest: func(t *testing.T) *http.Request {
				return buildRequest(t, "41")
			},
			setupMocks: func(mock mock, ctx context.Context) {
				mock.userResolver.EXPECT().Resolve(ctx).Return(users.AuthenticatedUser{ID: userID}, true)
				mock.savedQueries.EXPECT().GetSavedQuery(ctx, int64(41), services.Actor{UserID: userID.String()}).Return(services.SavedQuery{
					ID:          41,
					UserID:      userID.String(),
					Name:        "User query",
					Query:       "MATCH (n) RETURN n",
					Description: "Created by a user",
				}, nil)
			},
			expected: expected{
				responseCode: http.StatusOK,
				responseBody: `{"data":{"user_id":"` + userID.String() + `","name":"User query","query":"MATCH (n) RETURN n","description":"Created by a user","id":41,"created_at":"0001-01-01T00:00:00Z","updated_at":"0001-01-01T00:00:00Z","deleted_at":{"Time":"0001-01-01T00:00:00Z","Valid":false}}}`,
			},
		},
		{
			name: "Success: extension query is returned with the existing response contract - 200",
			buildRequest: func(t *testing.T) *http.Request {
				return buildRequest(t, "42")
			},
			setupMocks: func(mock mock, ctx context.Context) {
				mock.userResolver.EXPECT().Resolve(ctx).Return(users.AuthenticatedUser{ID: userID}, true)
				mock.savedQueries.EXPECT().GetSavedQuery(ctx, int64(42), services.Actor{UserID: userID.String()}).Return(savedQuery, nil)
			},
			expected: expected{
				responseCode: http.StatusOK,
				responseBody: `{"data":{"user_id":"00000000-0000-0000-0000-000000000000","name":"Extension query","query":"MATCH (n) RETURN n","description":"Provided by an extension","category":"extension","extension_id":7,"id":42,"created_at":"2026-11-01T12:00:00Z","updated_at":"2026-11-02T12:00:00Z","deleted_at":{"Time":"0001-01-01T00:00:00Z","Valid":false}}}`,
			},
		},
		{
			name: "Success: administrator role is projected into the actor - 200",
			buildRequest: func(t *testing.T) *http.Request {
				return buildRequest(t, "42")
			},
			setupMocks: func(mock mock, ctx context.Context) {
				mock.userResolver.EXPECT().Resolve(ctx).Return(users.AuthenticatedUser{ID: adminID, Administrator: true}, true)
				mock.savedQueries.EXPECT().GetSavedQuery(ctx, int64(42), services.Actor{UserID: adminID.String(), IsAdministrator: true}).Return(savedQuery, nil)
			},
			expected: expected{responseCode: http.StatusOK},
		},
		{
			name: "Error: missing authenticated user - 400",
			buildRequest: func(t *testing.T) *http.Request {
				return buildRequest(t, "42")
			},
			setupMocks: func(mock mock, ctx context.Context) {
				mock.userResolver.EXPECT().Resolve(ctx).Return(users.AuthenticatedUser{}, false)
			},
			expected: expected{responseCode: http.StatusBadRequest, errorMessage: "no associated user found"},
		},
		{
			name: "Error: malformed saved query ID - 400",
			buildRequest: func(t *testing.T) *http.Request {
				return buildRequest(t, "invalid")
			},
			setupMocks: func(mock mock, ctx context.Context) {
				mock.userResolver.EXPECT().Resolve(ctx).Return(users.AuthenticatedUser{ID: userID}, true)
			},
			expected: expected{responseCode: http.StatusBadRequest, errorMessage: api.ErrorResponseDetailsIDMalformed},
		},
		{
			name: "Error: missing or inaccessible saved query - 404",
			buildRequest: func(t *testing.T) *http.Request {
				return buildRequest(t, "42")
			},
			setupMocks: func(mock mock, ctx context.Context) {
				mock.userResolver.EXPECT().Resolve(ctx).Return(users.AuthenticatedUser{ID: userID}, true)
				mock.savedQueries.EXPECT().GetSavedQuery(ctx, int64(42), services.Actor{UserID: userID.String()}).Return(services.SavedQuery{}, services.ErrSavedQueryNotFound)
			},
			expected: expected{responseCode: http.StatusNotFound, errorMessage: "query not found"},
		},
		{
			name: "Error: service deadline is mapped to request timeout - 500",
			buildRequest: func(t *testing.T) *http.Request {
				return buildRequest(t, "42")
			},
			setupMocks: func(mock mock, ctx context.Context) {
				mock.userResolver.EXPECT().Resolve(ctx).Return(users.AuthenticatedUser{ID: userID}, true)
				mock.savedQueries.EXPECT().GetSavedQuery(ctx, int64(42), services.Actor{UserID: userID.String()}).Return(services.SavedQuery{}, context.DeadlineExceeded)
			},
			expected: expected{responseCode: http.StatusInternalServerError, errorMessage: api.ErrorResponseRequestTimeout},
		},
		{
			name: "Error: unexpected service failure - 500",
			buildRequest: func(t *testing.T) *http.Request {
				return buildRequest(t, "42")
			},
			setupMocks: func(mock mock, ctx context.Context) {
				mock.userResolver.EXPECT().Resolve(ctx).Return(users.AuthenticatedUser{ID: userID}, true)
				mock.savedQueries.EXPECT().GetSavedQuery(ctx, int64(42), services.Actor{UserID: userID.String()}).Return(services.SavedQuery{}, unexpectedErr)
			},
			expected: expected{responseCode: http.StatusInternalServerError, errorMessage: "an internal error has occurred that is preventing the service from servicing this request"},
		},
	}

	for _, testCase := range tt {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var (
				savedQueriesMock = mocks.NewMockSavedQueries(t)
				userResolverMock = mocks.NewMockUserResolver(t)
				handlerSet       = handlers.NewHandlersContainer(savedQueriesMock, userResolverMock)
				request          = testCase.buildRequest(t)
				recorder         = httptest.NewRecorder()
			)

			testCase.setupMocks(mock{savedQueries: savedQueriesMock, userResolver: userResolverMock}, request.Context())
			handlerSet.GetSavedQuery(recorder, request)

			assert.Equal(t, testCase.expected.responseCode, recorder.Code)
			if testCase.expected.responseBody != "" {
				assert.JSONEq(t, testCase.expected.responseBody, recorder.Body.String())
			} else if testCase.expected.responseCode >= http.StatusBadRequest && recorder.Body.Len() > 0 {
				var errorResponse responses.ErrorWrapper
				require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &errorResponse))
				assert.Equal(t, testCase.expected.responseCode, errorResponse.HTTPStatus)
				require.Len(t, errorResponse.Errors, 1)
				assert.Equal(t, testCase.expected.errorMessage, errorResponse.Errors[0].Message)
			}
		})
	}
}
