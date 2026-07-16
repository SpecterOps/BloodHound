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

package appdb_test

import (
	"context"
	"errors"
	"testing"

	"github.com/gofrs/uuid"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/specterops/bloodhound/cmd/api/src/auth"
	"github.com/specterops/bloodhound/cmd/api/src/bhctx"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/server/featureflags/internal/appdb"
	"github.com/specterops/bloodhound/server/featureflags/internal/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Literal SQL strings expected by the Store. These are compared via
// pgxmock.QueryMatcherEqual, which whitespace-normalises both sides, so
// column order, table name, WHERE predicate and parameter shape are
// load-bearing.
const (
	expectedSelectByKeySQL = `SELECT id, created_at, updated_at, key, name, description, enabled, user_updatable FROM feature_flags WHERE key = $1 LIMIT $2`

	expectedSelectByIDSQL = `SELECT id, created_at, updated_at, key, name, description, enabled, user_updatable FROM feature_flags WHERE id = $1 LIMIT $2`

	expectedSelectAllSQL = `SELECT id, created_at, updated_at, key, name, description, enabled, user_updatable FROM feature_flags`

	expectedUpdateSQL = `UPDATE feature_flags SET enabled = $1, updated_at = $2 WHERE id = $3`

	expectedAuditInsertSQL = `INSERT INTO audit_logs (created_at, actor_id, actor_name, actor_email, action, fields, request_id, source_ip_address, status, commit_id) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`
)

func newTestStore(t *testing.T) (*appdb.Store, pgxmock.PgxPoolIface) {
	t.Helper()
	pool, err := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual))
	require.NoError(t, err)
	t.Cleanup(pool.Close)
	return appdb.NewStore(pool), pool
}

func flagColumns() []string {
	return []string{"id", "created_at", "updated_at", "key", "name", "description", "enabled", "user_updatable"}
}

// authenticatedContext attaches a bhctx.Context carrying the supplied user as
// the auth owner, mirroring what the auth middleware does on real requests.
// SetFlag's audit-log path reads the actor from this context.
func authenticatedContext(userID uuid.UUID) context.Context {
	return bhctx.Set(context.Background(), &bhctx.Context{
		RequestID: "test-request",
		RequestIP: "127.0.0.1",
		AuthCtx: auth.Context{
			Owner: model.User{
				Unique:        model.Unique{ID: userID},
				PrincipalName: "test-user",
			},
		},
	})
}

func TestStore_GetFlagByKey(t *testing.T) {
	var (
		ctx   = context.Background()
		dbErr = errors.New("connection refused")
	)

	tests := []struct {
		name         string
		expectations func(pool pgxmock.PgxPoolIface)
		wantFlag     services.FeatureFlag
		wantErr      error
	}{
		{
			name: "returns the feature flag on success",
			expectations: func(pool pgxmock.PgxPoolIface) {
				pool.ExpectQuery(expectedSelectByKeySQL).WithArgs(services.FeatureOpenHoundSupport, 1).WillReturnRows(
					pool.NewRows(flagColumns()).AddRow(
						int32(7), nil, nil, services.FeatureOpenHoundSupport, "OpenHound Support", "desc", true, false,
					),
				)
			},
			wantFlag: services.FeatureFlag{ID: 7, Key: services.FeatureOpenHoundSupport, Name: "OpenHound Support", Description: "desc", Enabled: true},
		},
		{
			name: "maps zero rows to ErrNotFound",
			expectations: func(pool pgxmock.PgxPoolIface) {
				pool.ExpectQuery(expectedSelectByKeySQL).WithArgs(services.FeatureOpenHoundSupport, 1).WillReturnRows(
					pool.NewRows(flagColumns()),
				)
			},
			wantErr: services.ErrNotFound,
		},
		{
			name: "propagates other database errors",
			expectations: func(pool pgxmock.PgxPoolIface) {
				pool.ExpectQuery(expectedSelectByKeySQL).WithArgs(services.FeatureOpenHoundSupport, 1).WillReturnError(dbErr)
			},
			wantErr: dbErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store, pool := newTestStore(t)
			tt.expectations(pool)

			flag, err := store.GetFlagByKey(ctx, services.FeatureOpenHoundSupport)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantFlag, flag)
			}
			require.NoError(t, pool.ExpectationsWereMet())
		})
	}
}

func TestStore_GetFlagByID(t *testing.T) {
	var (
		ctx   = context.Background()
		dbErr = errors.New("connection refused")
	)

	tests := []struct {
		name         string
		expectations func(pool pgxmock.PgxPoolIface)
		wantFlag     services.FeatureFlag
		wantErr      error
	}{
		{
			name: "returns the feature flag on success",
			expectations: func(pool pgxmock.PgxPoolIface) {
				pool.ExpectQuery(expectedSelectByIDSQL).WithArgs(int32(11), 1).WillReturnRows(
					pool.NewRows(flagColumns()).AddRow(
						int32(11), nil, nil, services.FeatureAlerts, "Alerts", "desc", false, true,
					),
				)
			},
			wantFlag: services.FeatureFlag{ID: 11, Key: services.FeatureAlerts, Name: "Alerts", Description: "desc", Enabled: false, UserUpdatable: true},
		},
		{
			name: "maps zero rows to ErrNotFound",
			expectations: func(pool pgxmock.PgxPoolIface) {
				pool.ExpectQuery(expectedSelectByIDSQL).WithArgs(int32(11), 1).WillReturnRows(
					pool.NewRows(flagColumns()),
				)
			},
			wantErr: services.ErrNotFound,
		},
		{
			name: "propagates other database errors",
			expectations: func(pool pgxmock.PgxPoolIface) {
				pool.ExpectQuery(expectedSelectByIDSQL).WithArgs(int32(11), 1).WillReturnError(dbErr)
			},
			wantErr: dbErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store, pool := newTestStore(t)
			tt.expectations(pool)

			flag, err := store.GetFlagByID(ctx, 11)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantFlag, flag)
			}
			require.NoError(t, pool.ExpectationsWereMet())
		})
	}
}

func TestStore_GetAllFlags(t *testing.T) {
	var (
		ctx   = context.Background()
		dbErr = errors.New("connection refused")
	)

	tests := []struct {
		name         string
		expectations func(pool pgxmock.PgxPoolIface)
		wantFlags    []services.FeatureFlag
		wantErr      error
	}{
		{
			name: "returns every flag from the result set",
			expectations: func(pool pgxmock.PgxPoolIface) {
				pool.ExpectQuery(expectedSelectAllSQL).WillReturnRows(
					pool.NewRows(flagColumns()).
						AddRow(int32(1), nil, nil, services.FeatureOpenHoundSupport, "OpenHound", "", true, false).
						AddRow(int32(2), nil, nil, services.FeatureAlerts, "Alerts", "", false, true),
				)
			},
			wantFlags: []services.FeatureFlag{
				{ID: 1, Key: services.FeatureOpenHoundSupport, Name: "OpenHound", Enabled: true, UserUpdatable: false},
				{ID: 2, Key: services.FeatureAlerts, Name: "Alerts", Enabled: false, UserUpdatable: true},
			},
		},
		{
			name: "returns an empty slice when no flags are configured",
			expectations: func(pool pgxmock.PgxPoolIface) {
				pool.ExpectQuery(expectedSelectAllSQL).WillReturnRows(pool.NewRows(flagColumns()))
			},
			wantFlags: []services.FeatureFlag{},
		},
		{
			name: "propagates database errors",
			expectations: func(pool pgxmock.PgxPoolIface) {
				pool.ExpectQuery(expectedSelectAllSQL).WillReturnError(dbErr)
			},
			wantErr: dbErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store, pool := newTestStore(t)
			tt.expectations(pool)

			flags, err := store.GetAllFlags(ctx)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantFlags, flags)
			}
			require.NoError(t, pool.ExpectationsWereMet())
		})
	}
}

func TestStore_SetFlag(t *testing.T) {
	var (
		userID   = uuid.Must(uuid.NewV4())
		authCtx  = authenticatedContext(userID)
		flag     = services.FeatureFlag{ID: 42, Key: services.FeatureAlerts, Name: "Alerts", Enabled: true, UserUpdatable: false}
		userFlag = services.FeatureFlag{ID: 42, Key: services.FeatureAlerts, Name: "Alerts", Enabled: true, UserUpdatable: true}
		dbErr    = errors.New("connection refused")
	)

	tests := []struct {
		name         string
		ctx          context.Context
		flag         services.FeatureFlag
		expectations func(pool pgxmock.PgxPoolIface)
		wantErr      error
	}{
		{
			name: "commits the update without an audit entry when the flag is not user-updatable",
			ctx:  context.Background(),
			flag: flag,
			expectations: func(pool pgxmock.PgxPoolIface) {
				pool.ExpectBegin()
				pool.ExpectExec(expectedUpdateSQL).
					WithArgs(true, pgxmock.AnyArg(), int32(42)).
					WillReturnResult(pgxmock.NewResult("UPDATE", 1))
				pool.ExpectCommit()
			},
		},
		{
			name: "commits both the update and an audit entry when the flag is user-updatable",
			ctx:  authCtx,
			flag: userFlag,
			expectations: func(pool pgxmock.PgxPoolIface) {
				pool.ExpectBegin()
				pool.ExpectExec(expectedUpdateSQL).
					WithArgs(true, pgxmock.AnyArg(), int32(42)).
					WillReturnResult(pgxmock.NewResult("UPDATE", 1))
				pool.ExpectExec(expectedAuditInsertSQL).
					WithArgs(
						pgxmock.AnyArg(), // created_at
						userID.String(),  // actor_id
						"test-user",      // actor_name
						"",               // actor_email
						string(model.AuditLogActionToggleEarlyAccessFeatureFlag), // action
						pgxmock.AnyArg(),                    // fields (json)
						"test-request",                      // request_id
						"127.0.0.1",                         // source_ip_address
						string(model.AuditLogStatusSuccess), // status
						pgxmock.AnyArg(),                    // commit_id
					).
					WillReturnResult(pgxmock.NewResult("INSERT", 1))
				pool.ExpectCommit()
			},
		},
		{
			name: "rolls back and returns the error when BeginTx fails",
			ctx:  context.Background(),
			flag: flag,
			expectations: func(pool pgxmock.PgxPoolIface) {
				pool.ExpectBegin().WillReturnError(dbErr)
			},
			wantErr: dbErr,
		},
		{
			name: "rolls back when the UPDATE fails",
			ctx:  context.Background(),
			flag: flag,
			expectations: func(pool pgxmock.PgxPoolIface) {
				pool.ExpectBegin()
				pool.ExpectExec(expectedUpdateSQL).
					WithArgs(true, pgxmock.AnyArg(), int32(42)).
					WillReturnError(dbErr)
				pool.ExpectRollback()
			},
			wantErr: dbErr,
		},
		{
			name: "rolls back and returns ErrNotFound when the UPDATE matches no rows",
			ctx:  authCtx,
			flag: userFlag,
			expectations: func(pool pgxmock.PgxPoolIface) {
				pool.ExpectBegin()
				pool.ExpectExec(expectedUpdateSQL).
					WithArgs(true, pgxmock.AnyArg(), int32(42)).
					WillReturnResult(pgxmock.NewResult("UPDATE", 0))
				pool.ExpectRollback()
			},
			wantErr: services.ErrNotFound,
		},
		{
			name: "rolls back when the audit insert fails for a user-updatable flag",
			ctx:  authCtx,
			flag: userFlag,
			expectations: func(pool pgxmock.PgxPoolIface) {
				pool.ExpectBegin()
				pool.ExpectExec(expectedUpdateSQL).
					WithArgs(true, pgxmock.AnyArg(), int32(42)).
					WillReturnResult(pgxmock.NewResult("UPDATE", 1))
				pool.ExpectExec(expectedAuditInsertSQL).
					WithArgs(
						pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(),
						pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(),
					).
					WillReturnError(dbErr)
				pool.ExpectRollback()
			},
			wantErr: dbErr,
		},
		{
			name: "returns an error when no authenticated user is on the context",
			ctx:  context.Background(),
			flag: userFlag,
			expectations: func(pool pgxmock.PgxPoolIface) {
				pool.ExpectBegin()
				pool.ExpectExec(expectedUpdateSQL).
					WithArgs(true, pgxmock.AnyArg(), int32(42)).
					WillReturnResult(pgxmock.NewResult("UPDATE", 1))
				pool.ExpectRollback()
			},
			wantErr: errors.New("no authenticated user on context"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store, pool := newTestStore(t)
			tt.expectations(pool)

			err := store.SetFlag(tt.ctx, tt.flag)
			if tt.wantErr != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr.Error())
			} else {
				require.NoError(t, err)
			}
			require.NoError(t, pool.ExpectationsWereMet())
		})
	}
}
