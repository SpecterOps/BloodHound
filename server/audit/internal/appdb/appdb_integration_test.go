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
	"strings"
	"testing"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/peterldowns/pgtestdb"
	"github.com/specterops/bloodhound/cmd/api/src/auth"
	"github.com/specterops/bloodhound/cmd/api/src/config"
	"github.com/specterops/bloodhound/cmd/api/src/database"
	"github.com/specterops/bloodhound/cmd/api/src/test/integration/utils"
	"github.com/specterops/bloodhound/server/audit/internal/appdb"
	"github.com/specterops/bloodhound/server/audit/internal/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupStore(t *testing.T) (*appdb.Store, *pgxpool.Pool) {
	t.Helper()

	var (
		ctx      = context.Background()
		connConf = pgtestdb.Custom(t, getPostgresConfig(t), pgtestdb.NoopMigrator{})
	)

	cfg, err := config.NewDefaultConnectionConfiguration(connConf.URL())
	require.NoError(t, err)

	gormDB, dbPool, err := database.OpenDatabase(cfg.Database)
	require.NoError(t, err)

	db := database.NewBloodhoundDB(gormDB, dbPool, auth.NewIdentityResolver(), cfg)
	require.NoError(t, db.Migrate(ctx))

	t.Cleanup(func() { db.Close(ctx) })

	return appdb.NewStore(dbPool), dbPool
}

func getPostgresConfig(t *testing.T) pgtestdb.Config {
	t.Helper()

	cfg, err := utils.LoadIntegrationTestConfig()
	require.NoError(t, err)

	environmentMap := make(map[string]string)
	for _, entry := range strings.Fields(cfg.Database.Connection) {
		if parts := strings.SplitN(entry, "=", 2); len(parts) == 2 {
			environmentMap[parts[0]] = parts[1]
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

func TestStore_InsertAuditLog_PersistsRow(t *testing.T) {
	var (
		ctx          = context.Background()
		store, pool  = setupStore(t)
		commitID     = uuid.Must(uuid.NewV4())
		record       = services.AuditRecord{
			Action:          "POST /api/v2/roles/{role_id}",
			ActorID:         "actor-id",
			ActorName:       "actor-name",
			ActorEmail:      "actor@example.com",
			RequestID:       "req-1",
			SourceIPAddress: "10.0.0.1",
			Status:          services.StatusSuccess,
			CommitID:        commitID,
			Fields:          map[string]any{"key": "value"},
			Source:          services.SourceMiddleware,
		}
	)

	require.NoError(t, store.InsertAuditLog(ctx, record))

	var (
		action, status, source string
		gotCommitID            string
		fields                 map[string]any
		partition              string
	)
	require.NoError(t, pool.QueryRow(ctx,
		`SELECT action, status, source, commit_id, fields, tableoid::regclass::text
		 FROM audit_logs WHERE commit_id = $1`, commitID.String(),
	).Scan(&action, &status, &source, &gotCommitID, &fields, &partition))

	assert.Equal(t, record.Action, action)
	assert.Equal(t, string(services.StatusSuccess), status)
	assert.Equal(t, string(services.SourceMiddleware), source)
	assert.Equal(t, commitID.String(), gotCommitID)
	assert.Equal(t, map[string]any{"key": "value"}, fields)
	// The row must land in a concrete monthly partition, not the parent name.
	assert.True(t, strings.HasPrefix(partition, "audit_logs_"), "row should route to a child partition, got %q", partition)
	assert.NotEqual(t, "audit_logs", partition)
}

func TestStore_InsertAuditLog_EmptyFieldsStoredAsNull(t *testing.T) {
	var (
		ctx         = context.Background()
		store, pool = setupStore(t)
		commitID    = uuid.Must(uuid.NewV4())
	)

	require.NoError(t, store.InsertAuditLog(ctx, services.AuditRecord{
		Action:   "GET /x",
		Status:   services.StatusIntent,
		CommitID: commitID,
		Source:   services.SourceMiddleware,
	}))

	var fieldsIsNull bool
	require.NoError(t, pool.QueryRow(ctx,
		`SELECT fields IS NULL FROM audit_logs WHERE commit_id = $1`, commitID.String(),
	).Scan(&fieldsIsNull))
	assert.True(t, fieldsIsNull, "empty fields should persist as SQL NULL, not jsonb 'null'")
}

func TestStore_InsertAuditLog_InvalidStatusMapsToSentinel(t *testing.T) {
	var (
		ctx      = context.Background()
		store, _ = setupStore(t)
	)

	err := store.InsertAuditLog(ctx, services.AuditRecord{
		Action:   "GET /x",
		Status:   services.Status("bogus"), // violates the status CHECK constraint
		CommitID: uuid.Must(uuid.NewV4()),
		Source:   services.SourceMiddleware,
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, services.ErrInvalidAuditRecord)
}
