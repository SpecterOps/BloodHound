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

package middleware_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofrs/uuid"
	"github.com/gorilla/mux"
	"github.com/specterops/bloodhound/cmd/api/src/api/middleware"
	"github.com/specterops/bloodhound/cmd/api/src/auth"
	"github.com/specterops/bloodhound/cmd/api/src/bhctx"
	"github.com/specterops/bloodhound/cmd/api/src/database/types/null"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	audittypes "github.com/specterops/bloodhound/server/audit/types"
	"github.com/stretchr/testify/require"
)

const (
	testAuditRoute     = "/api/v2/things/{thing_id}"
	testAuditActorID   = "22222222-2222-2222-2222-222222222222"
	testAuditActorName = "actor"
	testAuditActorMail = "actor@example.com"
	testAuditRequestID = "request-id"
	testAuditRequestIP = "192.0.2.10"
	testAuditCommitID  = "11111111-1111-1111-1111-111111111111"
)

var errAuditWrite = errors.New("audit write failed")

type mockAuditService struct {
	commitID uuid.UUID

	intentErr  error
	successErr error
	failureErr error

	intentEntries  []audittypes.Entry
	successCommits []uuid.UUID
	failureCommits []uuid.UUID

	intentContextErr   error
	intentHasDeadline  bool
	successContextErr  error
	successHasDeadline bool
	failureContextErr  error
	failureHasDeadline bool
}

func (s *mockAuditService) Intent(ctx context.Context, entry audittypes.Entry) (uuid.UUID, error) {
	_, s.intentHasDeadline = ctx.Deadline()
	s.intentContextErr = ctx.Err()
	s.intentEntries = append(s.intentEntries, entry)
	return s.commitID, s.intentErr
}

func (s *mockAuditService) Success(ctx context.Context, commitID uuid.UUID, _ audittypes.Entry) error {
	_, s.successHasDeadline = ctx.Deadline()
	s.successContextErr = ctx.Err()
	s.successCommits = append(s.successCommits, commitID)
	return s.successErr
}

func (s *mockAuditService) Failure(ctx context.Context, commitID uuid.UUID, _ audittypes.Entry) error {
	_, s.failureHasDeadline = ctx.Deadline()
	s.failureContextErr = ctx.Err()
	s.failureCommits = append(s.failureCommits, commitID)
	return s.failureErr
}

type flushResponseRecorder struct {
	*httptest.ResponseRecorder
	flushed bool
}

func (s *flushResponseRecorder) Flush() {
	s.flushed = true
	s.ResponseRecorder.Flush()
}

func newAuditTestRequest(ctx context.Context, method string) *http.Request {
	request := httptest.NewRequest(method, "/api/v2/things/abc", nil)
	requestContext := &bhctx.Context{
		RequestID: testAuditRequestID,
		RequestIP: testAuditRequestIP,
		AuthCtx: auth.Context{
			Owner: model.User{
				PrincipalName: testAuditActorName,
				EmailAddress:  null.StringFrom(testAuditActorMail),
				Unique:        model.Unique{ID: uuid.FromStringOrNil(testAuditActorID)},
			},
		},
	}

	return request.WithContext(bhctx.Set(ctx, requestContext))
}

func newAuditTestRouter(auditService middleware.AuditService, handler http.HandlerFunc) *mux.Router {
	router := mux.NewRouter()
	router.Use(middleware.AuditMiddleware(auditService, router, nil))
	router.HandleFunc(testAuditRoute, handler)
	return router
}

func TestAuditMiddleware_Outcomes(t *testing.T) {
	type testData struct {
		name            string
		method          string
		handler         http.HandlerFunc
		expectedCode    int
		expectedSuccess int
		expectedFailure int
	}

	tests := []testData{
		{
			name:   "Success: explicit status - 201",
			method: http.MethodPost,
			handler: func(response http.ResponseWriter, _ *http.Request) {
				response.WriteHeader(http.StatusCreated)
			},
			expectedCode:    http.StatusCreated,
			expectedSuccess: 1,
		},
		{
			name:   "Error: handler response - 400",
			method: http.MethodDelete,
			handler: func(response http.ResponseWriter, _ *http.Request) {
				response.WriteHeader(http.StatusBadRequest)
			},
			expectedCode:    http.StatusBadRequest,
			expectedFailure: 1,
		},
		{
			name:            "Success: implicit status - 200",
			method:          http.MethodGet,
			handler:         func(http.ResponseWriter, *http.Request) {},
			expectedCode:    http.StatusOK,
			expectedSuccess: 1,
		},
		{
			name:   "Success: first status controls outcome - 204",
			method: http.MethodPut,
			handler: func(response http.ResponseWriter, _ *http.Request) {
				response.WriteHeader(http.StatusNoContent)
				response.WriteHeader(http.StatusInternalServerError)
			},
			expectedCode:    http.StatusNoContent,
			expectedSuccess: 1,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			var (
				auditService = &mockAuditService{commitID: uuid.FromStringOrNil(testAuditCommitID)}
				router       = newAuditTestRouter(auditService, testCase.handler)
				response     = httptest.NewRecorder()
			)

			router.ServeHTTP(response, newAuditTestRequest(context.Background(), testCase.method))

			require.Equal(t, testCase.expectedCode, response.Code)
			require.Len(t, auditService.intentEntries, 1)
			require.Len(t, auditService.successCommits, testCase.expectedSuccess)
			require.Len(t, auditService.failureCommits, testCase.expectedFailure)
			require.True(t, auditService.intentHasDeadline)

			entry := auditService.intentEntries[0]
			require.Equal(t, testCase.method+testAuditRoute, entry.Action)
			require.Equal(t, testAuditActorID, entry.ActorID)
			require.Equal(t, testAuditActorName, entry.ActorName)
			require.Equal(t, testAuditActorMail, entry.ActorEmail)
			require.Equal(t, testAuditRequestID, entry.RequestID)
			require.Equal(t, testAuditRequestIP, entry.SourceIPAddress)
		})
	}
}

func TestAuditMiddleware_IntentFailure(t *testing.T) {
	for _, method := range []string{http.MethodGet, http.MethodPost} {
		t.Run("Error: "+method+" intent write - 500", func(t *testing.T) {
			var (
				auditService = &mockAuditService{intentErr: errAuditWrite}
				handlerRan   bool
				router       = newAuditTestRouter(auditService, func(http.ResponseWriter, *http.Request) { handlerRan = true })
				response     = httptest.NewRecorder()
			)

			router.ServeHTTP(response, newAuditTestRequest(context.Background(), method))

			require.Equal(t, http.StatusInternalServerError, response.Code)
			require.False(t, handlerRan)
			require.Empty(t, auditService.successCommits)
			require.Empty(t, auditService.failureCommits)
		})
	}
}

func TestAuditMiddleware_RequestCancellation(t *testing.T) {
	t.Run("Error: cancelled before intent - 500", func(t *testing.T) {
		var (
			auditService       = &mockAuditService{commitID: uuid.FromStringOrNil(testAuditCommitID), intentErr: context.Canceled}
			requestCtx, cancel = context.WithCancel(context.Background())
			response           = httptest.NewRecorder()
		)
		cancel()

		newAuditTestRouter(auditService, func(http.ResponseWriter, *http.Request) {}).ServeHTTP(response, newAuditTestRequest(requestCtx, http.MethodPost))

		require.ErrorIs(t, auditService.intentContextErr, context.Canceled)
	})

	t.Run("Success: cancelled during handler still records outcome - 200", func(t *testing.T) {
		var (
			auditService       = &mockAuditService{commitID: uuid.FromStringOrNil(testAuditCommitID)}
			requestCtx, cancel = context.WithCancel(context.Background())
			response           = httptest.NewRecorder()
			router             = newAuditTestRouter(auditService, func(response http.ResponseWriter, _ *http.Request) {
				cancel()
				response.WriteHeader(http.StatusOK)
			})
		)
		defer cancel()

		router.ServeHTTP(response, newAuditTestRequest(requestCtx, http.MethodGet))

		require.Len(t, auditService.successCommits, 1)
		require.NoError(t, auditService.successContextErr)
		require.True(t, auditService.successHasDeadline)
	})
}

func TestAuditMiddleware_SkipsExcludedAndUnmatchedRoutes(t *testing.T) {
	type testData struct {
		name       string
		requestURL string
		excluded   bool
		statusCode int
	}

	tests := []testData{
		{name: "Success: excluded route - 200", requestURL: "/health", excluded: true, statusCode: http.StatusOK},
		{name: "Error: unmatched route - 404", requestURL: "/missing", statusCode: http.StatusNotFound},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			var (
				auditService = &mockAuditService{}
				router       = mux.NewRouter()
				response     = httptest.NewRecorder()
				request      = httptest.NewRequest(http.MethodGet, testCase.requestURL, nil)
			)
			router.Use(middleware.AuditMiddleware(auditService, router, func(routeTemplate string) bool {
				return testCase.excluded && routeTemplate == "/health"
			}))
			router.HandleFunc("/health", func(response http.ResponseWriter, _ *http.Request) {
				response.WriteHeader(http.StatusOK)
			})

			router.ServeHTTP(response, request.WithContext(bhctx.Set(request.Context(), &bhctx.Context{})))

			require.Equal(t, testCase.statusCode, response.Code)
			require.Empty(t, auditService.intentEntries)
		})
	}
}

func TestAuditMiddleware_HandlerPanic(t *testing.T) {
	var (
		auditService = &mockAuditService{commitID: uuid.FromStringOrNil(testAuditCommitID)}
		router       = newAuditTestRouter(auditService, func(http.ResponseWriter, *http.Request) { panic("handler boom") })
		response     = httptest.NewRecorder()
	)

	require.PanicsWithValue(t, "handler boom", func() {
		router.ServeHTTP(response, newAuditTestRequest(context.Background(), http.MethodPost))
	})
	require.Len(t, auditService.failureCommits, 1)
	require.Empty(t, auditService.successCommits)
	require.NoError(t, auditService.failureContextErr)
	require.True(t, auditService.failureHasDeadline)
}

func TestAuditMiddleware_OutcomeFailureDoesNotChangeResponse(t *testing.T) {
	var (
		auditService = &mockAuditService{
			commitID:   uuid.FromStringOrNil(testAuditCommitID),
			successErr: errAuditWrite,
		}
		router   = newAuditTestRouter(auditService, func(response http.ResponseWriter, _ *http.Request) { response.WriteHeader(http.StatusCreated) })
		response = httptest.NewRecorder()
	)

	router.ServeHTTP(response, newAuditTestRequest(context.Background(), http.MethodPost))

	require.Equal(t, http.StatusCreated, response.Code)
	require.Len(t, auditService.successCommits, 1)
}

func TestAuditMiddleware_ResponseControllerUsesUnderlyingWriter(t *testing.T) {
	var (
		auditService = &mockAuditService{commitID: uuid.FromStringOrNil(testAuditCommitID)}
		response     = &flushResponseRecorder{ResponseRecorder: httptest.NewRecorder()}
		handlerErr   error
		router       = newAuditTestRouter(auditService, func(response http.ResponseWriter, _ *http.Request) {
			handlerErr = http.NewResponseController(response).Flush()
		})
	)

	router.ServeHTTP(response, newAuditTestRequest(context.Background(), http.MethodGet))

	require.NoError(t, handlerErr)
	require.True(t, response.flushed)
	require.Len(t, auditService.successCommits, 1)
}
