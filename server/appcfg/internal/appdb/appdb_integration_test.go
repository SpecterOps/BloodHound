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

package appdb_test

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/peterldowns/pgtestdb"
	"github.com/specterops/bloodhound/cmd/api/src/auth"
	"github.com/specterops/bloodhound/cmd/api/src/config"
	"github.com/specterops/bloodhound/cmd/api/src/database"
	"github.com/specterops/bloodhound/cmd/api/src/database/types/null"
	"github.com/specterops/bloodhound/cmd/api/src/test/integration/utils"
	"github.com/specterops/bloodhound/server/appcfg/internal/appdb"
	"github.com/specterops/bloodhound/server/appcfg/internal/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupStoreAndPool is like setupStore but also returns the underlying pgxpool.Pool so
// callers that need to seed rows not exposed by the Store API can do so directly.
func setupStoreAndPool(t *testing.T) (*appdb.Store, *pgxpool.Pool) {
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

	t.Cleanup(func() { bhDB.Close(ctx) })

	return appdb.NewStore(bhDB.Pool()), bhDB.Pool()
}

// getPostgresConfig reads the integration test connection details from the
// configured environment and returns a pgtestdb.Config suitable for spinning
// up isolated databases. Supports both TCP and unix-socket host values.
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

func TestStore_GetDatapipeStatus_Integration(t *testing.T) {
	t.Run("returns ErrNotFound when no datapipe_status row exists", func(t *testing.T) {
		var (
			ctx         = context.Background()
			store, pool = setupStoreAndPool(t)
		)

		// Ensure the table is empty
		_, err := pool.Exec(ctx, "DELETE FROM datapipe_status")
		require.NoError(t, err)

		_, err = store.GetDatapipeStatus(ctx)
		assert.ErrorIs(t, err, services.ErrNotFound)
	})

	t.Run("returns the datapipe status when a row exists", func(t *testing.T) {
		var (
			ctx         = context.Background()
			store, pool = setupStoreAndPool(t)
		)

		// Update the datapipe_status row (migrations insert a default row)
		var (
			expectedStatus                  = services.DatapipeStatusIdle
			expectedUpdatedAt               = time.Now().UTC().Truncate(time.Microsecond)
			expectedLastCompleteAnalysisAt  = null.TimeFrom(time.Now().UTC().Add(-1 * time.Hour).Truncate(time.Microsecond))
			expectedLastAnalysisRunAt       = null.TimeFrom(time.Now().UTC().Add(-30 * time.Minute).Truncate(time.Microsecond))
			expectedNextScheduledAnalysisAt = null.TimeFrom(time.Now().UTC().Add(2 * time.Hour).Truncate(time.Microsecond))
		)

		_, err := pool.Exec(ctx, `
			UPDATE datapipe_status
			SET status = $1,
			    updated_at = $2,
			    last_complete_analysis_at = $3,
			    last_analysis_run_at = $4,
			    next_scheduled_analysis_at = $5
		`,
			expectedStatus,
			expectedUpdatedAt,
			expectedLastCompleteAnalysisAt,
			expectedLastAnalysisRunAt,
			expectedNextScheduledAnalysisAt,
		)
		require.NoError(t, err)

		status, err := store.GetDatapipeStatus(ctx)
		require.NoError(t, err)

		assert.Equal(t, expectedStatus, status.Status)
		assert.WithinDuration(t, expectedUpdatedAt, status.UpdatedAt, 1*time.Second)
		assert.True(t, status.LastCompleteAnalysisAt.Valid)
		assert.WithinDuration(t, expectedLastCompleteAnalysisAt.Time, status.LastCompleteAnalysisAt.Time, 1*time.Second)
		assert.True(t, status.LastAnalysisRunAt.Valid)
		assert.WithinDuration(t, expectedLastAnalysisRunAt.Time, status.LastAnalysisRunAt.Time, 1*time.Second)
		assert.True(t, status.NextScheduledAnalysisAt.Valid)
		assert.WithinDuration(t, expectedNextScheduledAnalysisAt.Time, status.NextScheduledAnalysisAt.Time, 1*time.Second)
	})

	t.Run("handles NULL next_scheduled_analysis_at correctly", func(t *testing.T) {
		var (
			ctx         = context.Background()
			store, pool = setupStoreAndPool(t)
		)

		// Update the row with NULL next_scheduled_analysis_at
		_, err := pool.Exec(ctx, `
			UPDATE datapipe_status
			SET status = $1,
			    updated_at = $2,
			    last_complete_analysis_at = $3,
			    last_analysis_run_at = $4,
			    next_scheduled_analysis_at = NULL
		`,
			services.DatapipeStatusIngesting,
			time.Now().UTC(),
			time.Now().UTC(),
			time.Now().UTC(),
		)
		require.NoError(t, err)

		status, err := store.GetDatapipeStatus(ctx)
		require.NoError(t, err)

		assert.Equal(t, services.DatapipeStatusIngesting, status.Status)
		assert.False(t, status.NextScheduledAnalysisAt.Valid, "NextScheduledAnalysisAt should be NULL")
	})

	t.Run("validates all DatapipeStatusType values are readable", func(t *testing.T) {
		testCases := []struct {
			name   string
			status services.DatapipeStatusType
		}{
			{"idle", services.DatapipeStatusIdle},
			{"ingesting", services.DatapipeStatusIngesting},
			{"analyzing", services.DatapipeStatusAnalyzing},
			{"purging", services.DatapipeStatusPurging},
			{"pruning", services.DatapipeStatusPruning},
			{"starting", services.DatapipeStatusStarting},
		}

		for _, tc := range testCases {
			tc := tc
			t.Run(tc.name, func(t *testing.T) {
				var (
					ctx         = context.Background()
					store, pool = setupStoreAndPool(t)
				)

				// Update the row with the test status
				_, err := pool.Exec(ctx, `
					UPDATE datapipe_status
					SET status = $1,
					    updated_at = $2,
					    last_complete_analysis_at = $3,
					    last_analysis_run_at = $4
				`,
					tc.status,
					time.Now().UTC(),
					time.Now().UTC(),
					time.Now().UTC(),
				)
				require.NoError(t, err)

				status, err := store.GetDatapipeStatus(ctx)
				require.NoError(t, err)
				assert.Equal(t, tc.status, status.Status)
			})
		}
	})

	t.Run("validates the LIMIT 1 clause returns a result", func(t *testing.T) {
		var (
			ctx         = context.Background()
			store, pool = setupStoreAndPool(t)
		)

		// The migration inserts a default row with NULL timestamps, so update it first
		_, err := pool.Exec(ctx, `
			UPDATE datapipe_status
			SET status = $1,
			    updated_at = $2,
			    last_complete_analysis_at = $3,
			    last_analysis_run_at = $4
		`,
			services.DatapipeStatusIdle,
			time.Now().UTC(),
			time.Now().UTC(),
			time.Now().UTC(),
		)
		require.NoError(t, err)

		// The singleton constraint prevents multiple rows. This test validates
		// that LIMIT 1 works as expected.
		status, err := store.GetDatapipeStatus(ctx)
		require.NoError(t, err)
		assert.Equal(t, services.DatapipeStatusIdle, status.Status)
		assert.NotZero(t, status.UpdatedAt)
	})
}
