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

// fakeAuditService is a hand-rolled test double for the middleware.AuditService
// port. It records every call and can be configured to return errors so the
// best-effort behavior of the middleware can be exercised.
type fakeAuditService struct {
	commitID uuid.UUID

	intentErr  error
	successErr error
	failureErr error

	intentEntries  []audit.Entry
	successCommits []uuid.UUID
	failureCommits []uuid.UUID
}

func (s *fakeAuditService) Intent(_ context.Context, entry audit.Entry) (uuid.UUID, error) {
	s.intentEntries = append(s.intentEntries, entry)
	return s.commitID, s.intentErr
}

func (s *fakeAuditService) Success(_ context.Context, commitID uuid.UUID, _ audit.Entry) error {
	s.successCommits = append(s.successCommits, commitID)
	return s.successErr
}

func (s *fakeAuditService) Failure(_ context.Context, commitID uuid.UUID, _ audit.Entry) error {
	s.failureCommits = append(s.failureCommits, commitID)
	return s.failureErr
}

const (
	testRoute     = "/api/v2/things/{thing_id}"
	testActorID   = "22222222-2222-2222-2222-222222222222"
	testActorName = "actor"
	testActorMail = "actor@example.com"
	testRequestID = "request-id"
)

// newAuditTestRouter wires the AuditMiddleware onto a router with a single
// registered route that returns the supplied status code. The returned router
// is used to drive requests through the middleware.
func newAuditTestRouter(auditService middleware.AuditService, handlerStatus int) *mux.Router {
	router := mux.NewRouter()
	router.Use(middleware.AuditMiddleware(auditService, router))
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

func TestAuditMiddleware_MutatingSuccess(t *testing.T) {
	var (
		fake     = &fakeAuditService{commitID: uuid.FromStringOrNil("11111111-1111-1111-1111-111111111111")}
		router   = newAuditTestRouter(fake, http.StatusOK)
		recorder = httptest.NewRecorder()
	)

	router.ServeHTTP(recorder, newAuditRequest(http.MethodPost))

	require.Equal(t, http.StatusOK, recorder.Code)
	require.Len(t, fake.intentEntries, 1)
	require.Len(t, fake.successCommits, 1)
	require.Empty(t, fake.failureCommits)
	require.Equal(t, fake.commitID, fake.successCommits[0])

	entry := fake.intentEntries[0]
	require.Equal(t, http.MethodPost+" "+testRoute, entry.Action)
	require.Equal(t, testActorID, entry.ActorID)
	require.Equal(t, testActorName, entry.ActorName)
	require.Equal(t, testActorMail, entry.ActorEmail)
	require.Equal(t, testRequestID, entry.RequestID)
}

func TestAuditMiddleware_MutatingFailure(t *testing.T) {
	var (
		fake     = &fakeAuditService{commitID: uuid.FromStringOrNil("11111111-1111-1111-1111-111111111111")}
		router   = newAuditTestRouter(fake, http.StatusBadRequest)
		recorder = httptest.NewRecorder()
	)

	router.ServeHTTP(recorder, newAuditRequest(http.MethodDelete))

	require.Equal(t, http.StatusBadRequest, recorder.Code)
	require.Len(t, fake.intentEntries, 1)
	require.Len(t, fake.failureCommits, 1)
	require.Empty(t, fake.successCommits)
	require.Equal(t, fake.commitID, fake.failureCommits[0])
}

var errAudit = errors.New("audit write failed")

func TestAuditMiddleware_NonMutatingSkipped(t *testing.T) {
	var (
		fake     = &fakeAuditService{}
		router   = newAuditTestRouter(fake, http.StatusOK)
		recorder = httptest.NewRecorder()
	)

	router.ServeHTTP(recorder, newAuditRequest(http.MethodGet))

	require.Equal(t, http.StatusOK, recorder.Code)
	require.Empty(t, fake.intentEntries)
	require.Empty(t, fake.successCommits)
	require.Empty(t, fake.failureCommits)
}

func TestAuditMiddleware_IntentErrorSkipsResult(t *testing.T) {
	var (
		fake     = &fakeAuditService{intentErr: errAudit}
		router   = newAuditTestRouter(fake, http.StatusOK)
		recorder = httptest.NewRecorder()
	)

	router.ServeHTTP(recorder, newAuditRequest(http.MethodPost))

	// The request is unaffected by the audit failure and no result row is
	// written because there is no commit id to link it to.
	require.Equal(t, http.StatusOK, recorder.Code)
	require.Len(t, fake.intentEntries, 1)
	require.Empty(t, fake.successCommits)
	require.Empty(t, fake.failureCommits)
}

func TestAuditMiddleware_ResultErrorSwallowed(t *testing.T) {
	var (
		fake     = &fakeAuditService{successErr: errAudit, failureErr: errAudit}
		router   = newAuditTestRouter(fake, http.StatusCreated)
		recorder = httptest.NewRecorder()
	)

	router.ServeHTTP(recorder, newAuditRequest(http.MethodPut))

	// A failing result write is logged and swallowed; the underlying request
	// completes normally.
	require.Equal(t, http.StatusCreated, recorder.Code)
	require.Len(t, fake.successCommits, 1)
}

func TestAuditMiddleware_UnauthenticatedActorEmpty(t *testing.T) {
	var (
		fake     = &fakeAuditService{}
		router   = newAuditTestRouter(fake, http.StatusOK)
		recorder = httptest.NewRecorder()
		request  = httptest.NewRequest(http.MethodPost, "/api/v2/things/abc", nil)
	)
	request = request.WithContext(bhctx.Set(request.Context(), &bhctx.Context{RequestID: testRequestID}))

	router.ServeHTTP(recorder, request)

	require.Len(t, fake.intentEntries, 1)
	entry := fake.intentEntries[0]
	require.Empty(t, entry.ActorID)
	require.Equal(t, "anonymous", entry.ActorName)
	require.Empty(t, entry.ActorEmail)
	require.Equal(t, testRequestID, entry.RequestID)
}
