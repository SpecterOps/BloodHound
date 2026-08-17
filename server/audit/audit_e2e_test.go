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

package audit_test

import (
	"bufio"
	"compress/gzip"
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/specterops/bloodhound/cmd/api/src/api/middleware"
	"github.com/specterops/bloodhound/cmd/api/src/api/router"
	"github.com/specterops/bloodhound/cmd/api/src/database"
	"github.com/specterops/bloodhound/cmd/api/src/database/types/null"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/server/audit"
	"github.com/specterops/bloodhound/server/internal/servertest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Route templates the e2e cases register and drive through the production chain.
// They are namespaced under /api/v2/ so they mimic real audited endpoints.
const (
	auditSuccessRoute = "/api/v2/audit-e2e/success"
	auditFailureRoute = "/api/v2/audit-e2e/failure"
	auditStreamRoute  = "/api/v2/audit-e2e/stream"
)

// successPayload is returned by the success handler. It is large enough that the
// production CompressionMiddleware compresses it, so the client-side decode
// exercises the full gzip round trip through the audit responseRecorder.
var successPayload = []byte(`{"data":{"message":"` +
	"audit-e2e-success-audit-e2e-success-audit-e2e-success-audit-e2e-success-" +
	"audit-e2e-success-audit-e2e-success-audit-e2e-success-audit-e2e-success" +
	`"}}`)

// auditHarness bundles the wired server, database, and connection pool shared by
// the e2e cases so individual cases only declare what differs.
type auditHarness struct {
	server  *servertest.Harness
	pool    *pgxpool.Pool
	baseURL string
	token   string
}

// newAuditHarness provisions an isolated migrated database, wires the production
// FOSS middleware plus the audit middleware (matching entrypoint.go ordering:
// Panic -> Auth -> Compression -> Audit), registers the audited test routes, and
// mints an admin bearer token so requests resolve to an authenticated actor.
func newAuditHarness(t *testing.T) *auditHarness {
	t.Helper()

	var (
		ctx    = context.Background()
		result = &auditHarness{}
	)

	result.server = servertest.NewHarness(t, func(routerInst *router.Router, db *database.BloodhoundDB) {
		// Wire audit exactly as entrypoint.go does: post-routing, after the FOSS
		// global middleware (which the harness registered first), so the audit
		// recorder wraps the compression writer and the authenticated actor is
		// already resolved onto the request context.
		auditService, _ := audit.Register(db.Pool())
		routerInst.UsePostrouting(middleware.AuditMiddleware(auditService, routerInst.MuxRouter(), "/health"))

		registerAuditTestRoutes(routerInst)
	})

	adminRole := servertest.AdminRole(t, ctx, result.server.DB)
	result.token = servertest.MintJWT(t, ctx, result.server.DB, result.server.Auther, model.User{
		PrincipalName: "audit-e2e-admin@example.com",
		EmailAddress:  null.StringFrom("audit-e2e-admin@example.com"),
		Roles:         model.Roles{adminRole},
	})

	result.baseURL = result.server.Server.URL
	result.pool = result.server.DB.Pool()

	return result
}

// registerAuditTestRoutes mounts the three audited endpoints used by the e2e
// cases: a 200 success route returning a compressible JSON body, a 500 failure
// route, and a chunked streaming route that flushes mid-response.
func registerAuditTestRoutes(routerInst *router.Router) {
	routerInst.POST(auditSuccessRoute, func(response http.ResponseWriter, _ *http.Request) {
		response.Header().Set("Content-Type", "application/json")
		response.WriteHeader(http.StatusOK)
		_, _ = response.Write(successPayload)
	}).RequireAuth()

	routerInst.POST(auditFailureRoute, func(response http.ResponseWriter, _ *http.Request) {
		response.Header().Set("Content-Type", "application/json")
		response.WriteHeader(http.StatusInternalServerError)
		_, _ = response.Write([]byte(`{"errors":[{"message":"boom"}]}`))
	}).RequireAuth()

	routerInst.GET(auditStreamRoute, func(response http.ResponseWriter, _ *http.Request) {
		response.Header().Set("Content-Type", "text/event-stream")
		response.WriteHeader(http.StatusOK)
		flusher, ok := response.(http.Flusher)
		if !ok {
			http.Error(response, "streaming unsupported", http.StatusInternalServerError)
			return
		}
		for _, chunk := range []string{"data: one\n\n", "data: two\n\n"} {
			_, _ = response.Write([]byte(chunk))
			flusher.Flush()
		}
	}).RequireAuth()
}

// countAuditRows returns the number of audit_logs rows written for the given
// commit status against the supplied action (method + route template).
func countAuditRows(t *testing.T, pool *pgxpool.Pool, action, status string) int {
	t.Helper()

	var count int
	require.NoError(t, pool.QueryRow(context.Background(),
		`SELECT COUNT(*) FROM audit_logs WHERE action = $1 AND status = $2`,
		action, status,
	).Scan(&count))
	return count
}

// authedRequest builds an authenticated request against the harness server.
func (s *auditHarness) authedRequest(t *testing.T, method, route string, body io.Reader) *http.Request {
	t.Helper()

	request, err := http.NewRequestWithContext(context.Background(), method, s.baseURL+route, body)
	require.NoError(t, err)
	request.Header.Set("Authorization", "Bearer "+s.token)
	return request
}

// TestE2E_AuditMiddleware_SuccessPersistsIntentAndSuccess drives a 200 request
// through the production chain and asserts (a) the gzip-compressed body decodes
// to exactly the handler payload (compressed once, decoded once) and (b) intent
// and success rows land in audit_logs, with no failure row.
func TestE2E_AuditMiddleware_SuccessPersistsIntentAndSuccess(t *testing.T) {
	var (
		harness = newAuditHarness(t)
		action  = http.MethodPost + auditSuccessRoute
	)

	// Manually request gzip and decode it ourselves so the assertion proves the
	// body was compressed exactly once (the double-gzip regression produced a
	// payload that failed this single decode).
	request := harness.authedRequest(t, http.MethodPost, auditSuccessRoute, nil)
	request.Header.Set("Accept-Encoding", "gzip")

	response, err := http.DefaultClient.Do(request)
	require.NoError(t, err)
	defer response.Body.Close()

	require.Equal(t, http.StatusOK, response.StatusCode)
	require.Equal(t, "gzip", response.Header.Get("Content-Encoding"))

	gzipReader, err := gzip.NewReader(response.Body)
	require.NoError(t, err, "response body must be valid single-pass gzip")
	decoded, err := io.ReadAll(gzipReader)
	require.NoError(t, err)
	require.Equal(t, string(successPayload), string(decoded))

	assert.Equal(t, 1, countAuditRows(t, harness.pool, action, "intent"))
	assert.Equal(t, 1, countAuditRows(t, harness.pool, action, "success"))
	assert.Equal(t, 0, countAuditRows(t, harness.pool, action, "failure"))
}

// TestE2E_AuditMiddleware_FailurePersistsIntentAndFailure drives a 500 request
// through the production chain and asserts intent and failure rows are written
// with no success row.
func TestE2E_AuditMiddleware_FailurePersistsIntentAndFailure(t *testing.T) {
	var (
		harness = newAuditHarness(t)
		action  = http.MethodPost + auditFailureRoute
	)

	response, err := http.DefaultClient.Do(harness.authedRequest(t, http.MethodPost, auditFailureRoute, nil))
	require.NoError(t, err)
	defer response.Body.Close()
	_, _ = io.Copy(io.Discard, response.Body)

	require.Equal(t, http.StatusInternalServerError, response.StatusCode)

	assert.Equal(t, 1, countAuditRows(t, harness.pool, action, "intent"))
	assert.Equal(t, 1, countAuditRows(t, harness.pool, action, "failure"))
	assert.Equal(t, 0, countAuditRows(t, harness.pool, action, "success"))
}

// TestE2E_AuditMiddleware_StreamingFlushesThroughChain verifies that a streaming
// handler can flush through the full middleware chain (audit responseRecorder ->
// compression writer -> transport) without the recorder hiding http.Flusher, and
// that the streamed body is received intact and the request is still audited.
func TestE2E_AuditMiddleware_StreamingFlushesThroughChain(t *testing.T) {
	var (
		harness = newAuditHarness(t)
		action  = http.MethodGet + auditStreamRoute
	)

	// Explicitly disable gzip so the transport does not buffer the whole body,
	// letting us read the flushed chunks as they arrive.
	request := harness.authedRequest(t, http.MethodGet, auditStreamRoute, nil)
	request.Header.Set("Accept-Encoding", "identity")

	response, err := http.DefaultClient.Do(request)
	require.NoError(t, err)
	defer response.Body.Close()

	require.Equal(t, http.StatusOK, response.StatusCode)

	var (
		reader    = bufio.NewReader(response.Body)
		collected strings.Builder
	)
	for {
		line, readErr := reader.ReadString('\n')
		collected.WriteString(line)
		if readErr == io.EOF {
			break
		}
		require.NoError(t, readErr)
	}
	require.Equal(t, "data: one\n\ndata: two\n\n", collected.String())

	// The streamed request is audited like any other: intent + success.
	require.Eventually(t, func() bool {
		return countAuditRows(t, harness.pool, action, "intent") == 1 &&
			countAuditRows(t, harness.pool, action, "success") == 1
	}, 5*time.Second, 50*time.Millisecond)
}
