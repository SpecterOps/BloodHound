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
	t.Parallel()

	type args struct {
		ctx context.Context
		key string
	}
	type want struct {
		flag services.FeatureFlag
		err  error
	}

	dbErr := errors.New("connection refused")

	tests := []struct {
		name      string
		args      args
		setupMock func(pool pgxmock.PgxPoolIface)
		want      want
	}{
		{
			name: "Success: returns the feature flag",
			args: args{ctx: context.Background(), key: services.FeatureOpenHoundSupport},
			setupMock: func(pool pgxmock.PgxPoolIface) {
				pool.ExpectQuery(expectedSelectByKeySQL).WithArgs(services.FeatureOpenHoundSupport, 1).WillReturnRows(
					pool.NewRows(flagColumns()).AddRow(
						int32(7), nil, nil, services.FeatureOpenHoundSupport, "OpenHound Support", "desc", true, false,
					),
				)
			},
			want: want{flag: services.FeatureFlag{ID: 7, Key: services.FeatureOpenHoundSupport, Name: "OpenHound Support", Description: "desc", Enabled: true}},
		},
		{
			name: "Error: maps zero rows to ErrNotFound",
			args: args{ctx: context.Background(), key: services.FeatureOpenHoundSupport},
			setupMock: func(pool pgxmock.PgxPoolIface) {
				pool.ExpectQuery(expectedSelectByKeySQL).WithArgs(services.FeatureOpenHoundSupport, 1).WillReturnRows(
					pool.NewRows(flagColumns()),
				)
			},
			want: want{err: services.ErrNotFound},
		},
		{
			name: "Error: propagates other database errors",
			args: args{ctx: context.Background(), key: services.FeatureOpenHoundSupport},
			setupMock: func(pool pgxmock.PgxPoolIface) {
				pool.ExpectQuery(expectedSelectByKeySQL).WithArgs(services.FeatureOpenHoundSupport, 1).WillReturnError(dbErr)
			},
			want: want{err: dbErr},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			store, pool := newTestStore(t)
			if test.setupMock != nil {
				test.setupMock(pool)
			}

			flag, err := store.GetFlagByKey(test.args.ctx, test.args.key)
			if test.want.err != nil {
				assert.ErrorIs(t, err, test.want.err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, test.want.flag, flag)
			}
			require.NoError(t, pool.ExpectationsWereMet())
		})
	}
}

func TestStore_GetFlagByID(t *testing.T) {
	t.Parallel()

	type args struct {
		ctx context.Context
		id  int32
	}
	type want struct {
		flag services.FeatureFlag
		err  error
	}

	dbErr := errors.New("connection refused")

	tests := []struct {
		name      string
		args      args
		setupMock func(pool pgxmock.PgxPoolIface)
		want      want
	}{
		{
			name: "Success: returns the feature flag",
			args: args{ctx: context.Background(), id: 11},
			setupMock: func(pool pgxmock.PgxPoolIface) {
				pool.ExpectQuery(expectedSelectByIDSQL).WithArgs(int32(11), 1).WillReturnRows(
					pool.NewRows(flagColumns()).AddRow(
						int32(11), nil, nil, services.FeatureAlerts, "Alerts", "desc", false, true,
					),
				)
			},
			want: want{flag: services.FeatureFlag{ID: 11, Key: services.FeatureAlerts, Name: "Alerts", Description: "desc", Enabled: false, UserUpdatable: true}},
		},
		{
			name: "Error: maps zero rows to ErrNotFound",
			args: args{ctx: context.Background(), id: 11},
			setupMock: func(pool pgxmock.PgxPoolIface) {
				pool.ExpectQuery(expectedSelectByIDSQL).WithArgs(int32(11), 1).WillReturnRows(
					pool.NewRows(flagColumns()),
				)
			},
			want: want{err: services.ErrNotFound},
		},
		{
			name: "Error: propagates other database errors",
			args: args{ctx: context.Background(), id: 11},
			setupMock: func(pool pgxmock.PgxPoolIface) {
				pool.ExpectQuery(expectedSelectByIDSQL).WithArgs(int32(11), 1).WillReturnError(dbErr)
			},
			want: want{err: dbErr},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			store, pool := newTestStore(t)
			if test.setupMock != nil {
				test.setupMock(pool)
			}

			flag, err := store.GetFlagByID(test.args.ctx, test.args.id)
			if test.want.err != nil {
				assert.ErrorIs(t, err, test.want.err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, test.want.flag, flag)
			}
			require.NoError(t, pool.ExpectationsWereMet())
		})
	}
}

func TestStore_GetAllFlags(t *testing.T) {
	t.Parallel()

	type args struct {
		ctx context.Context
	}
	type want struct {
		flags []services.FeatureFlag
		err   error
	}

	dbErr := errors.New("connection refused")

	tests := []struct {
		name      string
		args      args
		setupMock func(pool pgxmock.PgxPoolIface)
		want      want
	}{
		{
			name: "Success: returns every flag from the result set",
			args: args{ctx: context.Background()},
			setupMock: func(pool pgxmock.PgxPoolIface) {
				pool.ExpectQuery(expectedSelectAllSQL).WillReturnRows(
					pool.NewRows(flagColumns()).
						AddRow(int32(1), nil, nil, services.FeatureOpenHoundSupport, "OpenHound", "", true, false).
						AddRow(int32(2), nil, nil, services.FeatureAlerts, "Alerts", "", false, true),
				)
			},
			want: want{flags: []services.FeatureFlag{
				{ID: 1, Key: services.FeatureOpenHoundSupport, Name: "OpenHound", Enabled: true, UserUpdatable: false},
				{ID: 2, Key: services.FeatureAlerts, Name: "Alerts", Enabled: false, UserUpdatable: true},
			}},
		},
		{
			name: "Success: returns an empty slice when no flags are configured",
			args: args{ctx: context.Background()},
			setupMock: func(pool pgxmock.PgxPoolIface) {
				pool.ExpectQuery(expectedSelectAllSQL).WillReturnRows(pool.NewRows(flagColumns()))
			},
			want: want{flags: []services.FeatureFlag{}},
		},
		{
			name: "Error: propagates database errors",
			args: args{ctx: context.Background()},
			setupMock: func(pool pgxmock.PgxPoolIface) {
				pool.ExpectQuery(expectedSelectAllSQL).WillReturnError(dbErr)
			},
			want: want{err: dbErr},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			store, pool := newTestStore(t)
			if test.setupMock != nil {
				test.setupMock(pool)
			}

			flags, err := store.GetAllFlags(test.args.ctx)
			if test.want.err != nil {
				assert.ErrorIs(t, err, test.want.err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, test.want.flags, flags)
			}
			require.NoError(t, pool.ExpectationsWereMet())
		})
	}
}

func TestStore_SetFlag(t *testing.T) {
	t.Parallel()

	type args struct {
		ctx  context.Context
		flag services.FeatureFlag
	}
	type want struct {
		err         error
		errContains string
	}

	var (
		userID   = uuid.Must(uuid.NewV4())
		authCtx  = authenticatedContext(userID)
		flag     = services.FeatureFlag{ID: 42, Key: services.FeatureAlerts, Name: "Alerts", Enabled: true, UserUpdatable: false}
		userFlag = services.FeatureFlag{ID: 42, Key: services.FeatureAlerts, Name: "Alerts", Enabled: true, UserUpdatable: true}
		dbErr    = errors.New("connection refused")
	)

	tests := []struct {
		name      string
		args      args
		setupMock func(pool pgxmock.PgxPoolIface)
		want      want
	}{
		{
			name: "Success: commits the update without an audit entry when the flag is not user-updatable",
			args: args{ctx: context.Background(), flag: flag},
			setupMock: func(pool pgxmock.PgxPoolIface) {
				pool.ExpectBegin()
				pool.ExpectExec(expectedUpdateSQL).
					WithArgs(true, pgxmock.AnyArg(), int32(42)).
					WillReturnResult(pgxmock.NewResult("UPDATE", 1))
				pool.ExpectCommit()
			},
		},
		{
			name: "Success: commits both the update and an audit entry when the flag is user-updatable",
			args: args{ctx: authCtx, flag: userFlag},
			setupMock: func(pool pgxmock.PgxPoolIface) {
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
			name: "Error: rolls back and returns the error when BeginTx fails",
			args: args{ctx: context.Background(), flag: flag},
			setupMock: func(pool pgxmock.PgxPoolIface) {
				pool.ExpectBegin().WillReturnError(dbErr)
			},
			want: want{err: dbErr},
		},
		{
			name: "Error: rolls back when the UPDATE fails",
			args: args{ctx: context.Background(), flag: flag},
			setupMock: func(pool pgxmock.PgxPoolIface) {
				pool.ExpectBegin()
				pool.ExpectExec(expectedUpdateSQL).
					WithArgs(true, pgxmock.AnyArg(), int32(42)).
					WillReturnError(dbErr)
				pool.ExpectRollback()
			},
			want: want{err: dbErr},
		},
		{
			name: "Error: rolls back and returns ErrNotFound when the UPDATE matches no rows",
			args: args{ctx: authCtx, flag: userFlag},
			setupMock: func(pool pgxmock.PgxPoolIface) {
				pool.ExpectBegin()
				pool.ExpectExec(expectedUpdateSQL).
					WithArgs(true, pgxmock.AnyArg(), int32(42)).
					WillReturnResult(pgxmock.NewResult("UPDATE", 0))
				pool.ExpectRollback()
			},
			want: want{err: services.ErrNotFound},
		},
		{
			name: "Error: rolls back when the audit insert fails for a user-updatable flag",
			args: args{ctx: authCtx, flag: userFlag},
			setupMock: func(pool pgxmock.PgxPoolIface) {
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
			want: want{err: dbErr},
		},
		{
			name: "Error: returns an error when no authenticated user is on the context",
			args: args{ctx: context.Background(), flag: userFlag},
			setupMock: func(pool pgxmock.PgxPoolIface) {
				pool.ExpectBegin()
				pool.ExpectExec(expectedUpdateSQL).
					WithArgs(true, pgxmock.AnyArg(), int32(42)).
					WillReturnResult(pgxmock.NewResult("UPDATE", 1))
				pool.ExpectRollback()
			},
			want: want{errContains: "no authenticated user on context"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			store, pool := newTestStore(t)
			if test.setupMock != nil {
				test.setupMock(pool)
			}

			err := store.SetFlag(test.args.ctx, test.args.flag)
			switch {
			case test.want.err != nil:
				require.Error(t, err)
				assert.ErrorIs(t, err, test.want.err)
			case test.want.errContains != "":
				require.Error(t, err)
				assert.ErrorContains(t, err, test.want.errContains)
			default:
				require.NoError(t, err)
			}
			require.NoError(t, pool.ExpectationsWereMet())
		})
	}
}
