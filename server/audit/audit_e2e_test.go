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
	"context"
	"io"
	"net/http"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/specterops/bloodhound/cmd/api/src/api/middleware"
	"github.com/specterops/bloodhound/cmd/api/src/api/router"
	"github.com/specterops/bloodhound/cmd/api/src/database"
	"github.com/specterops/bloodhound/cmd/api/src/database/types/null"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/server/audit"
	"github.com/specterops/bloodhound/server/internal/servertest"
	"github.com/stretchr/testify/require"
)

const (
	auditSuccessRoute = "/api/v2/audit-e2e/success"
	auditFailureRoute = "/api/v2/audit-e2e/failure"
)

type auditHarness struct {
	server  *servertest.Harness
	pool    *pgxpool.Pool
	baseURL string
	token   string
}

func newAuditHarness(t *testing.T) *auditHarness {
	t.Helper()

	var (
		ctx    = context.Background()
		result = &auditHarness{}
	)

	result.server = servertest.NewHarness(t, func(routerInst *router.Router, database *database.BloodhoundDB) {
		auditRuntime := audit.Register(database.Pool())
		routerInst.UsePostrouting(middleware.AuditMiddleware(auditRuntime.Service, routerInst.MuxRouter(), routerInst.IsAuditExcluded))
		routerInst.POST(auditSuccessRoute, func(response http.ResponseWriter, _ *http.Request) {
			response.WriteHeader(http.StatusCreated)
		}).RequireAuth()
		routerInst.POST(auditFailureRoute, func(response http.ResponseWriter, _ *http.Request) {
			response.WriteHeader(http.StatusBadRequest)
		}).RequireAuth()
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

func (s *auditHarness) request(t *testing.T, route string) *http.Request {
	t.Helper()

	request, err := http.NewRequestWithContext(context.Background(), http.MethodPost, s.baseURL+route, nil)
	require.NoError(t, err)
	request.Header.Set("Authorization", "Bearer "+s.token)
	return request
}

func countAuditRows(t *testing.T, pool *pgxpool.Pool, action, status string) int {
	t.Helper()

	var count int
	require.NoError(t, pool.QueryRow(context.Background(),
		`SELECT COUNT(*) FROM audit_logs WHERE action = $1 AND status = $2`,
		action,
		status,
	).Scan(&count))

	return count
}

func TestAuditMiddleware_PersistsLifecycle(t *testing.T) {
	type testData struct {
		name             string
		route            string
		expectedCode     int
		expectedStatus   string
		unexpectedStatus string
	}

	tests := []testData{
		{
			name:             "Success: handler result persists intent and success - 201",
			route:            auditSuccessRoute,
			expectedCode:     http.StatusCreated,
			expectedStatus:   "success",
			unexpectedStatus: "failure",
		},
		{
			name:             "Error: handler result persists intent and failure - 400",
			route:            auditFailureRoute,
			expectedCode:     http.StatusBadRequest,
			expectedStatus:   "failure",
			unexpectedStatus: "success",
		},
	}

	harness := newAuditHarness(t)
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			response, err := http.DefaultClient.Do(harness.request(t, testCase.route))
			require.NoError(t, err)
			defer response.Body.Close()
			_, _ = io.Copy(io.Discard, response.Body)

			action := http.MethodPost + testCase.route
			require.Equal(t, testCase.expectedCode, response.StatusCode)
			require.Equal(t, 1, countAuditRows(t, harness.pool, action, "intent"))
			require.Equal(t, 1, countAuditRows(t, harness.pool, action, testCase.expectedStatus))
			require.Zero(t, countAuditRows(t, harness.pool, action, testCase.unexpectedStatus))
		})
	}
}
