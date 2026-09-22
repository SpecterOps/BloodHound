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
	"github.com/specterops/bloodhound/server/audit"
	"github.com/stretchr/testify/require"
)

// mockAuditService is a hand-rolled mock of the middleware.AuditService
// port. It records every call and can be configured to return errors so the
// best-effort behavior of the middleware can be exercised.
type mockAuditService struct {
	commitID uuid.UUID

	intentErr   error
	successErr  error
	failureErr  error
	rejectedErr error

	intentEntries   []audit.Entry
	successCommits  []uuid.UUID
	failureCommits  []uuid.UUID
	rejectedEntries []audit.Entry

	// rejectedCtxErr captures ctx.Err() at the time the rejection write was
	// invoked so tests can assert the write is not tied to the request's
	// cancellation.
	rejectedCtxErr error

	// successCtxErr/failureCtxErr capture ctx.Err() at the time the outcome write
	// was invoked so tests can assert the write is not tied to the request's
	// cancellation.
	successCtxErr error
	failureCtxErr error

	// intentDeadlineSet/intentCtxErr capture whether the intent write received a
	// bounded context (a deadline was set) and ctx.Err() at the time the intent
	// write was invoked, so tests can assert the write is bounded and that request
	// cancellation still propagates.
	intentDeadlineSet bool
	intentCtxErr      error
}

func (s *mockAuditService) Intent(ctx context.Context, entry audit.Entry) (uuid.UUID, error) {
	_, s.intentDeadlineSet = ctx.Deadline()
	s.intentCtxErr = ctx.Err()
	s.intentEntries = append(s.intentEntries, entry)
	return s.commitID, s.intentErr
}

func (s *mockAuditService) Success(ctx context.Context, commitID uuid.UUID, _ audit.Entry) error {
	s.successCommits = append(s.successCommits, commitID)
	s.successCtxErr = ctx.Err()
	return s.successErr
}

func (s *mockAuditService) Failure(ctx context.Context, commitID uuid.UUID, _ audit.Entry) error {
	s.failureCommits = append(s.failureCommits, commitID)
	s.failureCtxErr = ctx.Err()
	return s.failureErr
}

func (s *mockAuditService) RecordRejected(ctx context.Context, entry audit.Entry) error {
	s.rejectedEntries = append(s.rejectedEntries, entry)
	s.rejectedCtxErr = ctx.Err()
	return s.rejectedErr
}

const (
	testRoute     = "/api/v2/things/{thing_id}"
	testActorID   = "22222222-2222-2222-2222-222222222222"
	testActorName = "actor"
	testActorMail = "actor@example.com"
	testRequestID = "request-id"
	testCommitID  = "11111111-1111-1111-1111-111111111111"
)

// newAuditTestRouter wires the AuditMiddleware onto a router with a single
// registered route that returns the supplied status code. The returned router
// is used to drive requests through the middleware.
func newAuditTestRouter(auditService middleware.AuditService, handlerStatus int) *mux.Router {
	router := mux.NewRouter()
	router.Use(middleware.AuditMiddleware(auditService, router, nil))
	router.HandleFunc(testRoute, func(response http.ResponseWriter, _ *http.Request) {
		response.WriteHeader(handlerStatus)
	})
	return router
}

// newAuditRequest builds a request targeting testRoute with an authenticated
// user and request id populated in the BloodHound context.
func newAuditRequest(method string) *http.Request {
	request := httptest.NewRequest(method, "/api/v2/things/abc", nil)
	bhCtx := &bhctx.Context{
		RequestID: testRequestID,
		AuthCtx: auth.Context{
			Owner: model.User{
				PrincipalName: testActorName,
				EmailAddress:  null.StringFrom(testActorMail),
				Unique:        model.Unique{ID: uuid.FromStringOrNil(testActorID)},
			},
		},
	}
	return request.WithContext(bhctx.Set(request.Context(), bhCtx))
}

// TestAuditMiddleware_Outcomes covers the intent + result rows written for a
// request that runs to completion: a 2xx handler produces a success row, a
// >=400 handler produces a failure row, and reads (GET) are audited like any
// other request.
func TestAuditMiddleware_Outcomes(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		handlerStatus  int
		wantStatus     int
		wantSuccessLen int
		wantFailureLen int
	}{
		{
			name:           "mutating success writes a success row",
			method:         http.MethodPost,
			handlerStatus:  http.StatusOK,
			wantStatus:     http.StatusOK,
			wantSuccessLen: 1,
		},
		{
			name:           "mutating failure writes a failure row",
			method:         http.MethodDelete,
			handlerStatus:  http.StatusBadRequest,
			wantStatus:     http.StatusBadRequest,
			wantFailureLen: 1,
		},
		{
			name:           "read is audited and writes a success row",
			method:         http.MethodGet,
			handlerStatus:  http.StatusOK,
			wantStatus:     http.StatusOK,
			wantSuccessLen: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				mock     = &mockAuditService{commitID: uuid.FromStringOrNil(testCommitID)}
				router   = newAuditTestRouter(mock, tt.handlerStatus)
				recorder = httptest.NewRecorder()
			)

			router.ServeHTTP(recorder, newAuditRequest(tt.method))

			require.Equal(t, tt.wantStatus, recorder.Code)
			require.Len(t, mock.intentEntries, 1)
			require.Len(t, mock.successCommits, tt.wantSuccessLen)
			require.Len(t, mock.failureCommits, tt.wantFailureLen)

			entry := mock.intentEntries[0]
			require.Equal(t, tt.method+testRoute, entry.Action)
			require.Equal(t, testActorID, entry.ActorID)
			require.Equal(t, testActorName, entry.ActorName)
			require.Equal(t, testActorMail, entry.ActorEmail)
			require.Equal(t, testRequestID, entry.RequestID)

			if tt.wantSuccessLen == 1 {
				require.Equal(t, mock.commitID, mock.successCommits[0])
			}
			if tt.wantFailureLen == 1 {
				require.Equal(t, mock.commitID, mock.failureCommits[0])
			}
		})
	}
}

func TestAuditMiddleware_IntentWriteIsBounded(t *testing.T) {
	var (
		mock     = &mockAuditService{commitID: uuid.FromStringOrNil(testCommitID)}
		router   = newAuditTestRouter(mock, http.StatusOK)
		recorder = httptest.NewRecorder()
	)

	router.ServeHTTP(recorder, newAuditRequest(http.MethodPost))

	// The intent write must be bounded so a slow/unavailable database cannot block
	// the request goroutine indefinitely before the handler runs.
	require.Len(t, mock.intentEntries, 1)
	require.True(t, mock.intentDeadlineSet, "intent write must receive a context with a deadline")
}

func TestAuditMiddleware_IntentWriteRespectsRequestCancellation(t *testing.T) {
	var (
		mock            = &mockAuditService{commitID: uuid.FromStringOrNil(testCommitID)}
		router          = newAuditTestRouter(mock, http.StatusOK)
		recorder        = httptest.NewRecorder()
		baseCtx, cancel = context.WithCancel(context.Background())
	)

	// Cancel before the request runs to simulate a client disconnecting before the
	// handler is reached.
	cancel()

	request := httptest.NewRequest(http.MethodPost, "/api/v2/things/abc", nil)
	request = request.WithContext(bhctx.Set(baseCtx, &bhctx.Context{RequestID: testRequestID}))

	router.ServeHTTP(recorder, request)

	// The intent context is derived from the request context, so request
	// cancellation still propagates to the intent write.
	require.Len(t, mock.intentEntries, 1)
	require.ErrorIs(t, mock.intentCtxErr, context.Canceled)
}

var errAudit = errors.New("audit write failed")

// TestAuditMiddleware_IntentErrorFailsRequest verifies the fail-closed behavior:
// a failed intent write rejects the request with a 500 and the handler never
// runs, for both writes and reads.
func TestAuditMiddleware_IntentErrorFailsRequest(t *testing.T) {
	tests := []struct {
		name   string
		method string
	}{
		{name: "mutating request fails closed", method: http.MethodPost},
		{name: "read request fails closed", method: http.MethodGet},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				mock     = &mockAuditService{intentErr: errAudit}
				router   = newAuditTestRouter(mock, http.StatusOK)
				recorder = httptest.NewRecorder()
			)

			router.ServeHTTP(recorder, newAuditRequest(tt.method))

			// The handler never runs (so the 200 it would set is not observed)
			// and no result row is written because the action did not proceed.
			require.Equal(t, http.StatusInternalServerError, recorder.Code)
			require.Len(t, mock.intentEntries, 1)
			require.Empty(t, mock.successCommits)
			require.Empty(t, mock.failureCommits)
		})
	}
}

func TestAuditMiddleware_ResultErrorSwallowed(t *testing.T) {
	var (
		mock     = &mockAuditService{successErr: errAudit, failureErr: errAudit}
		router   = newAuditTestRouter(mock, http.StatusCreated)
		recorder = httptest.NewRecorder()
	)

	router.ServeHTTP(recorder, newAuditRequest(http.MethodPut))

	// A failing result write is logged and swallowed; the underlying request
	// completes normally.
	require.Equal(t, http.StatusCreated, recorder.Code)
	require.Len(t, mock.successCommits, 1)
}

func TestAuditMiddleware_ExcludedRouteNotAudited(t *testing.T) {
	var (
		mock     = &mockAuditService{}
		router   = mux.NewRouter()
		recorder = httptest.NewRecorder()
		request  = httptest.NewRequest(http.MethodGet, "/health", nil)
	)
	router.Use(middleware.AuditMiddleware(mock, router, func(routeTemplate string) bool { return routeTemplate == "/health" }))
	router.HandleFunc("/health", func(response http.ResponseWriter, _ *http.Request) {
		response.WriteHeader(http.StatusOK)
	})
	request = request.WithContext(bhctx.Set(request.Context(), &bhctx.Context{RequestID: testRequestID}))

	router.ServeHTTP(recorder, request)

	// An excluded route runs its handler normally but produces no audit rows.
	require.Equal(t, http.StatusOK, recorder.Code)
	require.Empty(t, mock.intentEntries)
	require.Empty(t, mock.successCommits)
	require.Empty(t, mock.failureCommits)
}

func TestAuditMiddleware_OutcomeWriteSurvivesRequestCancellation(t *testing.T) {
	var (
		mock            = &mockAuditService{commitID: uuid.FromStringOrNil(testCommitID)}
		router          = mux.NewRouter()
		recorder        = httptest.NewRecorder()
		baseCtx, cancel = context.WithCancel(context.Background())
	)
	defer cancel()

	router.Use(middleware.AuditMiddleware(mock, router, nil))
	router.HandleFunc(testRoute, func(response http.ResponseWriter, _ *http.Request) {
		// Simulate the client disconnecting while the handler runs.
		cancel()
		response.WriteHeader(http.StatusOK)
	})

	request := httptest.NewRequest(http.MethodGet, "/api/v2/things/abc", nil)
	request = request.WithContext(bhctx.Set(baseCtx, &bhctx.Context{RequestID: testRequestID}))

	router.ServeHTTP(recorder, request)

	// The outcome write still happens and is handed a context that is not
	// cancelled even though the request's context was cancelled mid-handler.
	require.Len(t, mock.successCommits, 1)
	require.NoError(t, mock.successCtxErr)
}

func TestAuditMiddleware_HandlerPanicRecordsFailure(t *testing.T) {
	var (
		mock     = &mockAuditService{commitID: uuid.FromStringOrNil(testCommitID)}
		router   = mux.NewRouter()
		recorder = httptest.NewRecorder()
	)

	router.Use(middleware.AuditMiddleware(mock, router, nil))
	router.HandleFunc(testRoute, func(_ http.ResponseWriter, _ *http.Request) {
		panic("handler boom")
	})

	// The middleware must re-panic so the outer PanicHandler can handle the panic;
	// without a PanicHandler in this test router the panic propagates out of
	// ServeHTTP unchanged.
	require.PanicsWithValue(t, "handler boom", func() {
		router.ServeHTTP(recorder, newAuditRequest(http.MethodPost))
	})

	// A panicking handler still produces an intent row and a matching failure row
	// rather than leaving the intent dangling with no outcome. No success row is
	// written.
	require.Len(t, mock.intentEntries, 1)
	require.Len(t, mock.failureCommits, 1)
	require.Empty(t, mock.successCommits)
	require.Equal(t, mock.commitID, mock.failureCommits[0])

	// The panic-path failure write uses a context detached from the request's
	// cancellation, so it is not cancelled.
	require.NoError(t, mock.failureCtxErr)
}

func TestAuditMiddleware_UnauthenticatedActorEmpty(t *testing.T) {
	var (
		mock     = &mockAuditService{}
		router   = newAuditTestRouter(mock, http.StatusOK)
		recorder = httptest.NewRecorder()
		request  = httptest.NewRequest(http.MethodPost, "/api/v2/things/abc", nil)
	)
	request = request.WithContext(bhctx.Set(request.Context(), &bhctx.Context{RequestID: testRequestID}))

	router.ServeHTTP(recorder, request)

	require.Len(t, mock.intentEntries, 1)
	entry := mock.intentEntries[0]
	require.Empty(t, entry.ActorID)
	require.Empty(t, entry.ActorName, "middleware leaves the actor empty; the service applies the unknown default")
	require.Empty(t, entry.ActorEmail)
	require.Equal(t, testRequestID, entry.RequestID)
}

// authRejectMiddleware stands in for AuthMiddleware rejecting a request before it
// reaches the inner AuditMiddleware: it writes the supplied status and does not
// call next, so the inner stage never runs.
func authRejectMiddleware(status int) mux.MiddlewareFunc {
	return func(_ http.Handler) http.Handler {
		return http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) {
			response.WriteHeader(status)
		})
	}
}

// newPreAuthChain builds the two-stage chain as wired in production: the pre-auth
// outer stage wraps the mux, and the inner AuditMiddleware is attached inside the
// mux (post-routing). When authRejectStatus is non-zero an auth stub rejects the
// request before the inner stage runs; otherwise the registered handler runs and
// returns handlerStatus.
func newPreAuthChain(auditService middleware.AuditService, isExcluded func(routeTemplate string) bool, authRejectStatus, handlerStatus int) http.Handler {
	router := mux.NewRouter()
	if authRejectStatus != 0 {
		router.Use(authRejectMiddleware(authRejectStatus))
	}
	router.Use(middleware.AuditMiddleware(auditService, router, isExcluded))
	router.HandleFunc(testRoute, func(response http.ResponseWriter, _ *http.Request) {
		response.WriteHeader(handlerStatus)
	})
	return middleware.PreAuthAuditMiddleware(auditService, router, isExcluded)(router)
}

// newUnauthedRequest builds a request targeting path with only a request id in
// the BloodHound context (no authenticated actor), mirroring a request that fails
// authentication.
func newUnauthedRequest(method, path string) *http.Request {
	request := httptest.NewRequest(method, path, nil)
	return request.WithContext(bhctx.Set(request.Context(), &bhctx.Context{RequestID: testRequestID}))
}

// TestPreAuthAuditMiddleware_AuthFailureRecordsRejection covers the core reason
// the outer stage exists: a request rejected by auth before the inner stage runs
// produces exactly one best-effort rejection row attributed by request id and
// source IP, with the actor left empty.
func TestPreAuthAuditMiddleware_AuthFailureRecordsRejection(t *testing.T) {
	var (
		mock     = &mockAuditService{}
		chain    = newPreAuthChain(mock, nil, http.StatusUnauthorized, http.StatusOK)
		recorder = httptest.NewRecorder()
	)

	chain.ServeHTTP(recorder, newUnauthedRequest(http.MethodPost, "/api/v2/things/abc"))

	require.Equal(t, http.StatusUnauthorized, recorder.Code)

	// The inner stage never ran, so only the outer rejection row is written.
	require.Empty(t, mock.intentEntries)
	require.Empty(t, mock.successCommits)
	require.Empty(t, mock.failureCommits)
	require.Len(t, mock.rejectedEntries, 1)

	entry := mock.rejectedEntries[0]
	require.Equal(t, http.MethodPost+testRoute, entry.Action)
	require.Equal(t, testRequestID, entry.RequestID)
	require.Empty(t, entry.ActorID)
	require.Empty(t, entry.ActorName, "middleware leaves the actor empty; the service applies the unknown default")
	require.Empty(t, entry.ActorEmail)

	// The rejection write is detached from request cancellation.
	require.NoError(t, mock.rejectedCtxErr)
}

// TestPreAuthAuditMiddleware_HappyPathNoDoubleWrite verifies that when the inner
// stage audits a request the outer stage does not also write a rejection row.
func TestPreAuthAuditMiddleware_HappyPathNoDoubleWrite(t *testing.T) {
	var (
		mock     = &mockAuditService{commitID: uuid.FromStringOrNil(testCommitID)}
		chain    = newPreAuthChain(mock, nil, 0, http.StatusOK)
		recorder = httptest.NewRecorder()
	)

	chain.ServeHTTP(recorder, newAuditRequest(http.MethodPost))

	require.Equal(t, http.StatusOK, recorder.Code)

	// The inner stage took ownership: one intent + one success, no rejection row.
	require.Len(t, mock.intentEntries, 1)
	require.Len(t, mock.successCommits, 1)
	require.Empty(t, mock.rejectedEntries, "the outer stage must not double-write when the inner stage audited the request")
}

// TestPreAuthAuditMiddleware_SkippedRoutes verifies the outer stage does not write
// a rejection row for excluded or unmatched routes, even when the response is a
// rejection.
func TestPreAuthAuditMiddleware_SkippedRoutes(t *testing.T) {
	tests := []struct {
		name       string
		isExcluded func(routeTemplate string) bool
		path       string
		wantStatus int
	}{
		{
			name:       "excluded route is not audited",
			isExcluded: func(routeTemplate string) bool { return routeTemplate == testRoute },
			path:       "/api/v2/things/abc",
			wantStatus: http.StatusUnauthorized,
		},
		{
			// An unmatched route never reaches the auth stub (mux runs Use
			// middleware only for matched routes), so mux returns its 404.
			name:       "unmatched route is not audited",
			isExcluded: nil,
			path:       "/api/v2/does-not-match",
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				mock     = &mockAuditService{}
				chain    = newPreAuthChain(mock, tt.isExcluded, http.StatusUnauthorized, http.StatusOK)
				recorder = httptest.NewRecorder()
			)

			chain.ServeHTTP(recorder, newUnauthedRequest(http.MethodGet, tt.path))

			require.Equal(t, tt.wantStatus, recorder.Code)
			require.Empty(t, mock.rejectedEntries)
			require.Empty(t, mock.intentEntries)
		})
	}
}

// TestPreAuthAuditMiddleware_OnlyRecordsAuthStatuses verifies the outer stage
// records a rejection only for the pre-auth statuses it is scoped to (401 failed
// authentication, 400 malformed request) and ignores other pre-inner rejections.
func TestPreAuthAuditMiddleware_OnlyRecordsAuthStatuses(t *testing.T) {
	tests := []struct {
		name         string
		rejectStatus int
		wantRejected int
	}{
		{name: "401 unauthorized is recorded", rejectStatus: http.StatusUnauthorized, wantRejected: 1},
		{name: "400 bad request is recorded", rejectStatus: http.StatusBadRequest, wantRejected: 1},
		{name: "403 forbidden is not recorded", rejectStatus: http.StatusForbidden, wantRejected: 0},
		{name: "500 internal error is not recorded", rejectStatus: http.StatusInternalServerError, wantRejected: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				mock     = &mockAuditService{}
				chain    = newPreAuthChain(mock, nil, tt.rejectStatus, http.StatusOK)
				recorder = httptest.NewRecorder()
			)

			chain.ServeHTTP(recorder, newUnauthedRequest(http.MethodPost, "/api/v2/things/abc"))

			require.Equal(t, tt.rejectStatus, recorder.Code)
			require.Len(t, mock.rejectedEntries, tt.wantRejected)
		})
	}
}

// TestPreAuthAuditMiddleware_RejectionWriteErrorSwallowed verifies that a failing
// rejection write is logged and swallowed, never altering the response the client
// already received.
func TestPreAuthAuditMiddleware_RejectionWriteErrorSwallowed(t *testing.T) {
	var (
		mock     = &mockAuditService{rejectedErr: errAudit}
		chain    = newPreAuthChain(mock, nil, http.StatusUnauthorized, http.StatusOK)
		recorder = httptest.NewRecorder()
	)

	chain.ServeHTTP(recorder, newUnauthedRequest(http.MethodPost, "/api/v2/things/abc"))

	require.Equal(t, http.StatusUnauthorized, recorder.Code)
	require.Len(t, mock.rejectedEntries, 1)
}
