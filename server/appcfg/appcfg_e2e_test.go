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

package appcfg_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/peterldowns/pgtestdb"
	v2 "github.com/specterops/bloodhound/cmd/api/src/api/v2"
	"github.com/specterops/bloodhound/cmd/api/src/auth"
	"github.com/specterops/bloodhound/cmd/api/src/config"
	"github.com/specterops/bloodhound/cmd/api/src/database"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/cmd/api/src/test/integration/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// datapipeStatusResponseEnvelope is the full JSON envelope shape returned by
// the GET /api/v2/datapipe/status handler. All five documented fields are included.
type datapipeStatusResponseEnvelope struct {
	Data model.DatapipeStatusWrapper `json:"data"`
}

// setupAppcfgDB creates an isolated test database with all migrations applied.
// The database is automatically closed when the test ends.
func setupAppcfgDB(t *testing.T) *database.BloodhoundDB {
	t.Helper()

	var (
		ctx      = context.Background()
		connConf = pgtestdb.Custom(t, getAppcfgPostgresConfig(t), pgtestdb.NoopMigrator{})
	)

	cfg, err := config.NewDefaultConnectionConfiguration(connConf.URL())
	require.NoError(t, err)

	gormDB, dbPool, err := database.OpenDatabase(cfg.Database)
	require.NoError(t, err)

	db := database.NewBloodhoundDB(gormDB, dbPool, auth.NewIdentityResolver(), cfg)
	require.NoError(t, db.Migrate(ctx))

	t.Cleanup(func() { db.Close(ctx) })

	return db
}

// getAppcfgPostgresConfig reads the integration test configuration from the
// environment and returns a pgtestdb.Config for the appcfg e2e tests.
func getAppcfgPostgresConfig(t *testing.T) pgtestdb.Config {
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

// newDatapipeStatusHandler wires the old v2.Resources.GetDatapipeStatus handler.
func newDatapipeStatusHandler(db *database.BloodhoundDB) http.HandlerFunc {
	return v2.Resources{DB: db}.GetDatapipeStatus
}

func TestGetDatapipeStatus(t *testing.T) {
	var (
		db        = setupAppcfgDB(t)
		ctx       = context.Background()
		muxRouter = mux.NewRouter()
		server    = httptest.NewServer(muxRouter)
	)
	muxRouter.HandleFunc("/api/v2/datapipe/status", newDatapipeStatusHandler(db)).Methods(http.MethodGet)
	t.Cleanup(server.Close)

	newGetRequest := func(t *testing.T) *http.Request {
		t.Helper()
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, server.URL+"/api/v2/datapipe/status", nil)
		require.NoError(t, err)
		return req
	}

	t.Run("returns 200 OK with datapipe status in idle state", func(t *testing.T) {
		resp, err := http.DefaultClient.Do(newGetRequest(t))
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))

		var envelope datapipeStatusResponseEnvelope
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&envelope))

		// Verify all contract fields are present and valid
		assert.Equal(t, model.DatapipeStatusIdle, envelope.Data.Status)
		assert.IsType(t, time.Time{}, envelope.Data.UpdatedAt)
		assert.IsType(t, time.Time{}, envelope.Data.LastCompleteAnalysisAt)
		assert.IsType(t, time.Time{}, envelope.Data.LastAnalysisRunAt)
	})
}
