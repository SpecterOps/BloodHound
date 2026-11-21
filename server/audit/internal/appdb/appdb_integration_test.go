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
	cfg.Database.Connection = connConf.URL()

	gormDB, databasePool, err := database.OpenDatabase(cfg.Database)
	require.NoError(t, err)

	bloodhoundDB := database.NewBloodhoundDB(gormDB, databasePool, auth.NewIdentityResolver(), cfg)
	require.NoError(t, bloodhoundDB.Migrate(ctx))

	t.Cleanup(func() { bloodhoundDB.Close(ctx) })

	return appdb.NewStore(bloodhoundDB.Pool()), bloodhoundDB.Pool()
}

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

func TestStore_CreateAuditLog_Integration(t *testing.T) {
	type expected struct {
		auditLog services.AuditLog
	}
	type testData struct {
		name     string
		auditLog services.AuditLog
		expected expected
	}

	var (
		ctx         = context.Background()
		store, pool = setupStore(t)
		commitID    = uuid.Must(uuid.NewV4())
		expectedLog = services.AuditLog{
			ActorID:         "actor-id",
			ActorName:       "Integration Test Actor",
			ActorEmail:      "actor@example.com",
			Action:          services.ActionUpdateParameter,
			Fields:          map[string]any{"key": "feature.enabled", "value": true},
			RequestID:       "request-id",
			SourceIPAddress: "127.0.0.1",
			Status:          services.StatusIntent,
			CommitID:        commitID,
		}
		tests = []testData{
			{
				name:     "Success: inserts an audit log",
				auditLog: expectedLog,
				expected: expected{auditLog: expectedLog},
			},
		}
	)

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			var (
				actualCreatedAt       time.Time
				actualActorID         string
				actualActorName       string
				actualActorEmail      string
				actualAction          string
				actualFields          []byte
				actualRequestID       string
				actualSourceIPAddress string
				actualStatus          string
				actualCommitID        string
			)

			require.NoError(t, store.CreateAuditLog(ctx, testCase.auditLog))

			err := pool.QueryRow(ctx, `
				SELECT created_at, actor_id, actor_name, actor_email, action, fields,
				       request_id, source_ip_address, status, commit_id
				FROM audit_logs
				WHERE commit_id = $1
			`, testCase.auditLog.CommitID.String()).Scan(
				&actualCreatedAt,
				&actualActorID,
				&actualActorName,
				&actualActorEmail,
				&actualAction,
				&actualFields,
				&actualRequestID,
				&actualSourceIPAddress,
				&actualStatus,
				&actualCommitID,
			)
			require.NoError(t, err)

			assert.WithinDuration(t, time.Now().UTC(), actualCreatedAt, 5*time.Second)
			assert.Equal(t, testCase.expected.auditLog.ActorID, actualActorID)
			assert.Equal(t, testCase.expected.auditLog.ActorName, actualActorName)
			assert.Equal(t, testCase.expected.auditLog.ActorEmail, actualActorEmail)
			assert.Equal(t, string(testCase.expected.auditLog.Action), actualAction)
			assert.JSONEq(t, `{"key":"feature.enabled","value":true}`, string(actualFields))
			assert.Equal(t, testCase.expected.auditLog.RequestID, actualRequestID)
			assert.Equal(t, testCase.expected.auditLog.SourceIPAddress, actualSourceIPAddress)
			assert.Equal(t, string(testCase.expected.auditLog.Status), actualStatus)
			assert.Equal(t, testCase.expected.auditLog.CommitID.String(), actualCommitID)
		})
	}
}
