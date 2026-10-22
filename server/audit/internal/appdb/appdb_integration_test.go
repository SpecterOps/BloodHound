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
		ctx           = context.Background()
		configuration = integrationConfiguration(t)
		connection    = pgtestdb.Custom(t, configuration, pgtestdb.NoopMigrator{})
	)

	databaseConfiguration, err := config.NewDefaultConnectionConfiguration(connection.URL())
	require.NoError(t, err)

	gormDatabase, databasePool, err := database.OpenDatabase(databaseConfiguration.Database)
	require.NoError(t, err)

	bloodhoundDatabase := database.NewBloodhoundDB(gormDatabase, databasePool, auth.NewIdentityResolver(), databaseConfiguration)
	require.NoError(t, bloodhoundDatabase.Migrate(ctx))
	t.Cleanup(func() { bloodhoundDatabase.Close(ctx) })

	return appdb.NewStore(databasePool), databasePool
}

func integrationConfiguration(t *testing.T) pgtestdb.Config {
	t.Helper()

	configuration, err := utils.LoadIntegrationTestConfig()
	require.NoError(t, err)

	var values = make(map[string]string)
	for _, entry := range strings.Fields(configuration.Database.Connection) {
		if parts := strings.SplitN(entry, "=", 2); len(parts) == 2 {
			values[parts[0]] = parts[1]
		}
	}

	return pgtestdb.Config{
		DriverName:                "pgx",
		Host:                      values["host"],
		Port:                      values["port"],
		User:                      values["user"],
		Password:                  values["password"],
		Database:                  values["dbname"],
		Options:                   "sslmode=disable",
		ForceTerminateConnections: true,
	}
}

func TestStore_InsertAuditLog(t *testing.T) {
	var (
		ctx            = context.Background()
		store, pool    = setupStore(t)
		commitID       = uuid.Must(uuid.NewV4())
		expectedFields = map[string]any{"key": "value"}
	)

	require.NoError(t, store.InsertAuditLog(ctx, services.AuditRecord{
		Action:          "POST/api/v2/roles/{role_id}",
		ActorID:         "actor-id",
		ActorName:       "actor-name",
		ActorEmail:      "actor@example.com",
		RequestID:       "request-id",
		SourceIPAddress: "10.0.0.1",
		Status:          services.StatusSuccess,
		CommitID:        commitID,
		Fields:          expectedFields,
	}))

	var (
		status    string
		fields    map[string]any
		partition string
	)
	require.NoError(t, pool.QueryRow(ctx, `
		SELECT status, fields, tableoid::regclass::text
		FROM audit_logs
		WHERE commit_id = $1
	`, commitID.String()).Scan(&status, &fields, &partition))

	assert.Equal(t, string(services.StatusSuccess), status)
	assert.Equal(t, expectedFields, fields)
	assert.True(t, strings.HasPrefix(partition, "audit_logs_"))
}
