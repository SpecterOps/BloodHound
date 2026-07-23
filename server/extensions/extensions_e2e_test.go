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

package extensions_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/peterldowns/pgtestdb"
	"github.com/specterops/bloodhound/cmd/api/src/auth"
	"github.com/specterops/bloodhound/cmd/api/src/config"
	"github.com/specterops/bloodhound/cmd/api/src/database"
	"github.com/specterops/bloodhound/cmd/api/src/test/integration/utils"
	"github.com/specterops/bloodhound/server/extensions/internal/appdb"
	"github.com/specterops/bloodhound/server/extensions/internal/handlers"
	"github.com/specterops/bloodhound/server/extensions/internal/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// getPostgresConfig reads the integration test configuration from the environment
// and returns a pgtestdb.Config for the extensions e2e tests.
func getPostgresConfig(t *testing.T) pgtestdb.Config {
	t.Helper()

	cfg, err := utils.LoadIntegrationTestConfig()
	require.NoError(t, err)

	environmentMap := make(map[string]string)
	for entry := range strings.FieldsSeq(cfg.Database.Connection) {
		if parts := strings.SplitN(entry, "=", 2); len(parts) == 2 {
			environmentMap[parts[0]] = parts[1]
		}
	}

	if strings.HasPrefix(environmentMap["host"], "/") {
		return pgtestdb.Config{
			DriverName: "pgx",
			User:       environmentMap["user"],
			Password:   environmentMap["password"],
			Database:   environmentMap["dbname"],
			Options:    fmt.Sprintf("host=%s", url.PathEscape(environmentMap["host"])),
			TestRole: &pgtestdb.Role{
				Username:     environmentMap["user"],
				Password:     environmentMap["password"],
				Capabilities: "NOSUPERUSER NOCREATEROLE",
			},
		}
	}

	return pgtestdb.Config{
		DriverName:                "pgx",
		Host:                      environmentMap["host"],
		Port:                      environmentMap["port"],
		User:                      environmentMap["user"],
		Password:                  environmentMap["password"],
		Database:                  environmentMap["dbname"],
		Options:                   "sslmode=disable",
		ForceTerminateConnections: true,
	}
}

// extensionsHarness bundles the wired HTTP handler and the seeded node kind id so
// E2E test cases can assert the full response contract.
type extensionsHarness struct {
	handler    *mux.Router
	nodeKindID int32
}

// newExtensionsHarness spins up an isolated postgres database via pgtestdb, applies
// the relational migrations and built-in extension data, seeds a node kind with two
// kind infos and wires the appdb -> services -> handlers stack onto a mux router.
func newExtensionsHarness(t *testing.T) extensionsHarness {
	t.Helper()

	var (
		ctx      = context.Background()
		connConf = pgtestdb.Custom(t, getPostgresConfig(t), pgtestdb.NoopMigrator{})
	)

	cfg, err := config.NewDefaultConnectionConfiguration(connConf.URL())
	require.NoError(t, err)

	gormDB, dbPool, err := database.OpenDatabase(cfg.Database)
	require.NoError(t, err)

	bhDB := database.NewBloodhoundDB(gormDB, dbPool, auth.NewIdentityResolver(), cfg)
	require.NoError(t, bhDB.Migrate(ctx))
	require.NoError(t, bhDB.PopulateExtensionData(ctx))
	t.Cleanup(func() { bhDB.Close(ctx) })

	harness := extensionsHarness{}
	kindID := seedNodeKind(t, ctx, dbPool, &harness)
	seedKindInfos(t, ctx, dbPool, kindID, harness.nodeKindID)

	handlerSet := handlers.NewHandlersContainer(services.NewService(appdb.NewStore(dbPool)))

	harness.handler = mux.NewRouter()
	harness.handler.HandleFunc(
		fmt.Sprintf("/api/v2/node-kinds/{%s}", handlers.URIPathVariableNodeKindID),
		handlerSet.GetNodeKindByID,
	).Methods(http.MethodGet)

	return harness
}

// seedNodeKind inserts a schema_extensions row, a kind row and a schema_node_kinds row,
// records the node kind id on the harness, registers a cascade-delete cleanup and
// returns the seeded kind id.
func seedNodeKind(t *testing.T, ctx context.Context, pool *pgxpool.Pool, harness *extensionsHarness) int32 {
	t.Helper()

	var extensionID int32
	require.NoError(t, pool.QueryRow(ctx, `
		INSERT INTO schema_extensions (name, display_name, version, is_builtin, namespace)
		VALUES ('TestExtension', 'Test Extension', '1.0.0', false, 'TST')
		RETURNING id`).Scan(&extensionID))

	var kindID int32
	require.NoError(t, pool.QueryRow(ctx, `
		INSERT INTO kind (name) VALUES ('TestNodeKind') RETURNING id`).Scan(&kindID))

	require.NoError(t, pool.QueryRow(ctx, `
		INSERT INTO schema_node_kinds (schema_extension_id, kind_id, display_name, description, is_display_kind, icon, icon_color)
		VALUES ($1, $2, 'Test Node Kind', 'a test node kind', true, 'user', '#fff')
		RETURNING id`, extensionID, kindID).Scan(&harness.nodeKindID))

	t.Cleanup(func() {
		_, _ = pool.Exec(ctx, `DELETE FROM schema_extensions WHERE name = 'TestExtension'`)
		_, _ = pool.Exec(ctx, `DELETE FROM kind WHERE name = 'TestNodeKind'`)
	})

	return kindID
}

// seedKindInfos inserts two schema_kind_info rows for the supplied kind/node kind,
// ordering the INSERT by position 1 then 0. Cleanup is handled by seedNodeKind's
// cascade delete.
func seedKindInfos(t *testing.T, ctx context.Context, pool *pgxpool.Pool, kindID, nodeKindID int32) {
	t.Helper()

	_, err := pool.Exec(ctx, `
		INSERT INTO schema_kind_info (kind_id, node_kind_id, info_key, title, position, content)
		VALUES ($1, $2, 'details', 'Details', 1, '{"markdown":{"content":"details md"}}'),
		       ($1, $2, 'overview', 'Overview', 0, '{"markdown":{"content":"overview md"}}')`,
		kindID, nodeKindID)
	require.NoError(t, err)
}

func TestGetNodeKindByID(t *testing.T) {
	var (
		harness  = newExtensionsHarness(t)
		basePath = "/api/v2/node-kinds/"
	)

	type testCase struct {
		name       string
		path       func(harness extensionsHarness) string
		wantStatus int
		assertBody func(t *testing.T, body []byte, harness extensionsHarness)
	}

	tests := []testCase{
		{
			name:       "success_-_returns_node_kind_view_with_info",
			path:       func(harness extensionsHarness) string { return basePath + strconv.Itoa(int(harness.nodeKindID)) },
			wantStatus: http.StatusOK,
			assertBody: func(t *testing.T, body []byte, harness extensionsHarness) {
				assert.JSONEq(t, expectedEnvelope(harness), string(body))
			},
		},
		{
			name:       "error_-_returns_400_when_id_is_malformed",
			path:       func(harness extensionsHarness) string { return basePath + "not-a-number" },
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "error_-_returns_404_when_node_kind_not_found",
			path:       func(harness extensionsHarness) string { return basePath + strconv.Itoa(int(harness.nodeKindID)+100000) },
			wantStatus: http.StatusNotFound,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			var (
				recorder = httptest.NewRecorder()
				request  = httptest.NewRequest(http.MethodGet, testCase.path(harness), nil)
			)

			harness.handler.ServeHTTP(recorder, request)

			assert.Equal(t, testCase.wantStatus, recorder.Code)
			if testCase.assertBody != nil {
				assert.Equal(t, "application/json", recorder.Header().Get("Content-Type"))
				testCase.assertBody(t, recorder.Body.Bytes(), harness)
			}
		})
	}
}

// expectedEnvelope returns the exact {"data":{...}} JSON the handler is expected to
// produce for the seeded node kind, including the info map keyed by info_key with
// flattened markdown content.
func expectedEnvelope(harness extensionsHarness) string {
	return fmt.Sprintf(`{"data":{
		"node_kind_id":%d,
		"name":"TestNodeKind",
		"display_name":"Test Node Kind",
		"description":"a test node kind",
		"is_display_kind":true,
		"icon":"user",
		"color":"#fff",
		"info":{
			"overview":{"title":"Overview","position":0,"markdown":{"content":"overview md"}},
			"details":{"title":"Details","position":1,"markdown":{"content":"details md"}}
		}
	}}`, harness.nodeKindID)
}
