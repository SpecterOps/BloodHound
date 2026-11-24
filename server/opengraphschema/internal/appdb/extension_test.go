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
	"fmt"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/pashagolub/pgxmock/v6"
	"github.com/specterops/bloodhound/packages/go/params"
	"github.com/specterops/bloodhound/server/opengraphschema/internal/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// expectedGetExtensionSQL is the literal SQL the Store builds for GetExtension.
const expectedGetExtensionSQL = `SELECT id, name, display_name, version, is_builtin, namespace, created_at, updated_at FROM schema_extensions WHERE id = $1`

func extensionColumns() []string {
	return []string{"id", "name", "display_name", "version", "is_builtin", "namespace", "created_at", "updated_at"}
}

func TestStore_GetExtension(t *testing.T) {
	var (
		createdAt = time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
		updatedAt = time.Date(2026, time.January, 2, 0, 0, 0, 0, time.UTC)
		ctx       = context.Background()
		extID     = int32(7)
		dbErr     = errors.New("db error")
	)

	tests := []struct {
		name          string
		expectations  func(pool pgxmock.PgxPoolIface)
		wantExtension services.Extension
		wantErr       error
		errorContext  string
	}{
		{
			name: "success_-_maps_all_columns",
			expectations: func(pool pgxmock.PgxPoolIface) {
				pool.ExpectQuery(expectedGetExtensionSQL).WithArgs(extID).WillReturnRows(
					pool.NewRows(extensionColumns()).AddRow(
						int32(7), "TestExtension", "Test Extension", "1.0.0", false, "TST", createdAt, updatedAt,
					),
				)
			},
			wantExtension: services.Extension{
				ID: 7, Name: "TestExtension", DisplayName: "Test Extension",
				Version: "1.0.0", IsBuiltin: false, Namespace: "TST",
				CreatedAt: createdAt, UpdatedAt: updatedAt,
			},
		},
		{
			name: "error_-_maps_no_rows_to_ErrExtensionNotFound",
			expectations: func(pool pgxmock.PgxPoolIface) {
				pool.ExpectQuery(expectedGetExtensionSQL).WithArgs(extID).WillReturnRows(
					pool.NewRows(extensionColumns()),
				)
			},
			wantExtension: services.Extension{},
			wantErr:       services.ErrExtensionNotFound,
		},
		{
			name: "error_-_isolates_database_error",
			expectations: func(pool pgxmock.PgxPoolIface) {
				pool.ExpectQuery(expectedGetExtensionSQL).WithArgs(extID).WillReturnError(dbErr)
			},
			wantExtension: services.Extension{},
			errorContext:  "fetching extension",
		},
		{
			name: "error_-_isolates_row_error",
			expectations: func(pool pgxmock.PgxPoolIface) {
				pool.ExpectQuery(expectedGetExtensionSQL).WithArgs(extID).WillReturnRows(pool.NewRows(extensionColumns()).AddRow(int32(7), "TestExtension", "Test Extension", "1.0.0", false, "TST", createdAt, updatedAt).RowError(0, dbErr))
			},
			errorContext: "reading rows",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store, pool := newTestStore(t)
			tt.expectations(pool)

			extension, err := store.GetExtension(ctx, extID)
			assert.Equal(t, tt.wantExtension, extension)
			if tt.errorContext != "" {
				require.ErrorContains(t, err, tt.errorContext)
				assert.NotErrorIs(t, err, dbErr)
				assert.Nil(t, errors.Unwrap(err))
			} else if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
			}
			require.NoError(t, pool.ExpectationsWereMet())
		})
	}
}

const (
	expectedListExtensionsSQL         = `SELECT id, name, display_name, version, is_builtin, namespace, created_at, updated_at FROM schema_extensions ORDER BY id ASC`
	expectedCreateExtensionSQL        = `INSERT INTO schema_extensions (name, display_name, version, is_builtin, namespace, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id, name, display_name, version, is_builtin, namespace, created_at, updated_at`
	expectedUpdateExtensionSQL        = `UPDATE schema_extensions SET name = $1, display_name = $2, version = $3, namespace = $4, updated_at = $5 WHERE id = $6 RETURNING id, name, display_name, version, is_builtin, namespace, created_at, updated_at`
	expectedDeleteExtensionBuiltinSQL = `SELECT is_builtin FROM schema_extensions WHERE id = $1 FOR UPDATE`
	expectedDeleteExtensionSQL        = `DELETE FROM schema_extensions WHERE id = $1 RETURNING id`
)

func TestStore_ListExtensions(t *testing.T) {
	t.Parallel()

	var (
		ctx           = context.Background()
		databaseError = errors.New("db error")
		createdAt     = time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
		updatedAt     = time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC)
	)

	tests := []struct {
		name         string
		filters      params.Filters
		setupMock    func(pgxmock.PgxPoolIface)
		want         []services.Extension
		wantCount    int
		wantErr      error
		errorContext string
	}{
		{
			name: "error_-_rejects_unknown_filter_field",
			filters: params.Filters{
				"unknown": {{Operator: params.Equals, Value: "value"}},
			},
			want:         nil,
			wantCount:    0,
			errorContext: "filter references unknown field",
		},
		{
			name: "error_-_isolates_query_error",
			setupMock: func(pool pgxmock.PgxPoolIface) {
				pool.ExpectQuery(expectedListExtensionsSQL).WillReturnError(databaseError)
			},
			want:         nil,
			errorContext: "error fetching data",
		},
		{
			name: "error_-_isolates_row_error",
			setupMock: func(pool pgxmock.PgxPoolIface) {
				pool.ExpectQuery(expectedListExtensionsSQL).WillReturnRows(
					pool.NewRows(extensionColumns()).AddRow(
						int32(7), "TestExtension", "Test Extension", "1.0.0", false, "TST", createdAt, updatedAt,
					).RowError(0, databaseError),
				)
			},
			want:         nil,
			errorContext: "error reading rows",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			store, pool := newTestStore(t)
			if test.setupMock != nil {
				test.setupMock(pool)
			}

			extensions, count, err := store.ListExtensions(ctx, test.filters, params.SortItems{}, 0, 0)
			assert.Equal(t, test.want, extensions)
			assert.Equal(t, test.wantCount, count)
			if test.errorContext != "" {
				require.ErrorContains(t, err, test.errorContext)
			} else {
				require.ErrorIs(t, err, test.wantErr)
			}
			require.NoError(t, pool.ExpectationsWereMet())
		})
	}
}

func TestStore_CreateExtension(t *testing.T) {
	t.Parallel()

	var (
		ctx           = context.Background()
		databaseError = errors.New("db error")
		input         = services.Extension{
			Name:        "TestExtension",
			DisplayName: "Test Extension",
			Version:     "1.0.0",
			Namespace:   "TST",
		}
	)

	tests := []struct {
		name         string
		setupMock    func(pgxmock.PgxPoolIface)
		want         services.Extension
		wantErr      error
		errorContext string
	}{
		{
			name: "error_-_rejects_duplicate_namespace",
			setupMock: func(pool pgxmock.PgxPoolIface) {
				pool.ExpectQuery(expectedCreateExtensionSQL).
					WithArgs(input.Name, input.DisplayName, input.Version, false, input.Namespace, pgxmock.AnyArg(), pgxmock.AnyArg()).
					WillReturnError(&pgconn.PgError{ConstraintName: "schema_extensions_namespace_unique"})
			},
			want:    services.Extension{},
			wantErr: services.ErrDuplicateExtensionNamespace,
		},
		{
			name: "error_-_rejects_duplicate_name",
			setupMock: func(pool pgxmock.PgxPoolIface) {
				pool.ExpectQuery(expectedCreateExtensionSQL).
					WithArgs(input.Name, input.DisplayName, input.Version, false, input.Namespace, pgxmock.AnyArg(), pgxmock.AnyArg()).
					WillReturnError(&pgconn.PgError{ConstraintName: "schema_extensions_name_key"})
			},
			want:    services.Extension{},
			wantErr: services.ErrDuplicateExtensionName,
		},
		{
			name: "error_-_isolates_query_error",
			setupMock: func(pool pgxmock.PgxPoolIface) {
				pool.ExpectQuery(expectedCreateExtensionSQL).
					WithArgs(input.Name, input.DisplayName, input.Version, false, input.Namespace, pgxmock.AnyArg(), pgxmock.AnyArg()).
					WillReturnError(databaseError)
			},
			want:         services.Extension{},
			errorContext: "creating extension",
		},
		{
			name: "error_-_isolates_row_error",
			setupMock: func(pool pgxmock.PgxPoolIface) {
				pool.ExpectQuery(expectedCreateExtensionSQL).
					WithArgs(input.Name, input.DisplayName, input.Version, false, input.Namespace, pgxmock.AnyArg(), pgxmock.AnyArg()).
					WillReturnRows(pool.NewRows(extensionColumns()).CloseError(databaseError))
			},
			want:         services.Extension{},
			errorContext: "reading created extension",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			store, pool := newTestStore(t)
			test.setupMock(pool)

			extension, err := store.CreateExtension(ctx, input)
			assert.Equal(t, test.want, extension)
			if test.errorContext != "" {
				require.ErrorContains(t, err, test.errorContext)
			} else {
				require.ErrorIs(t, err, test.wantErr)
			}
			require.NoError(t, pool.ExpectationsWereMet())
		})
	}
}

func TestStore_UpdateExtension(t *testing.T) {
	t.Parallel()

	var (
		ctx           = context.Background()
		databaseError = errors.New("db error")
		input         = services.Extension{
			ID:          7,
			Name:        "UpdatedExtension",
			DisplayName: "Updated Extension",
			Version:     "2.0.0",
			Namespace:   "UPD",
		}
	)

	tests := []struct {
		name         string
		setupMock    func(pgxmock.PgxPoolIface)
		wantErr      error
		errorContext string
	}{
		{
			name: "error_-_returns_extension_not_found",
			setupMock: func(pool pgxmock.PgxPoolIface) {
				pool.ExpectQuery(expectedUpdateExtensionSQL).
					WithArgs(input.Name, input.DisplayName, input.Version, input.Namespace, pgxmock.AnyArg(), input.ID).
					WillReturnRows(pool.NewRows(extensionColumns()))
			},
			wantErr: services.ErrExtensionNotFound,
		},
		{
			name: "error_-_rejects_duplicate_namespace",
			setupMock: func(pool pgxmock.PgxPoolIface) {
				pool.ExpectQuery(expectedUpdateExtensionSQL).
					WithArgs(input.Name, input.DisplayName, input.Version, input.Namespace, pgxmock.AnyArg(), input.ID).
					WillReturnError(&pgconn.PgError{ConstraintName: "schema_extensions_namespace_unique"})
			},
			wantErr: services.ErrDuplicateExtensionNamespace,
		},
		{
			name: "error_-_rejects_duplicate_name",
			setupMock: func(pool pgxmock.PgxPoolIface) {
				pool.ExpectQuery(expectedUpdateExtensionSQL).
					WithArgs(input.Name, input.DisplayName, input.Version, input.Namespace, pgxmock.AnyArg(), input.ID).
					WillReturnError(&pgconn.PgError{ConstraintName: "schema_extensions_name_key"})
			},
			wantErr: services.ErrDuplicateExtensionName,
		},
		{
			name: "error_-_isolates_query_error",
			setupMock: func(pool pgxmock.PgxPoolIface) {
				pool.ExpectQuery(expectedUpdateExtensionSQL).
					WithArgs(input.Name, input.DisplayName, input.Version, input.Namespace, pgxmock.AnyArg(), input.ID).
					WillReturnError(databaseError)
			},
			errorContext: "updating extension",
		},
		{
			name: "error_-_isolates_row_error",
			setupMock: func(pool pgxmock.PgxPoolIface) {
				pool.ExpectQuery(expectedUpdateExtensionSQL).
					WithArgs(input.Name, input.DisplayName, input.Version, input.Namespace, pgxmock.AnyArg(), input.ID).
					WillReturnRows(pool.NewRows(extensionColumns()).CloseError(databaseError))
			},
			errorContext: "reading updated extension",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			store, pool := newTestStore(t)
			test.setupMock(pool)

			extension, err := store.UpdateExtension(ctx, input)
			assert.Equal(t, services.Extension{}, extension)
			if test.errorContext != "" {
				require.ErrorContains(t, err, test.errorContext)
			} else {
				require.ErrorIs(t, err, test.wantErr)
			}
			require.NoError(t, pool.ExpectationsWereMet())
		})
	}
}

func TestStore_DeleteExtension(t *testing.T) {
	t.Parallel()

	var (
		ctx           = context.Background()
		extensionID   = int32(7)
		databaseError = errors.New("db error")
	)

	tests := []struct {
		name      string
		setupMock func(pgxmock.PgxPoolIface)
		wantErr   error
		context   string
	}{
		{
			name: "error_-_isolates_begin_transaction_error",
			setupMock: func(pool pgxmock.PgxPoolIface) {
				pool.ExpectBeginTx(pgx.TxOptions{}).WillReturnError(databaseError)
			},
			context: "beginning transaction",
		},
		{
			name: "error_-_returns_extension_not_found",
			setupMock: func(pool pgxmock.PgxPoolIface) {
				pool.ExpectBeginTx(pgx.TxOptions{})
				pool.ExpectQuery(expectedDeleteExtensionBuiltinSQL).
					WithArgs(extensionID).
					WillReturnRows(pool.NewRows([]string{"is_builtin"}))
				pool.ExpectRollback()
			},
			wantErr: services.ErrExtensionNotFound,
		},
		{
			name: "error_-_isolates_select_error",
			setupMock: func(pool pgxmock.PgxPoolIface) {
				pool.ExpectBeginTx(pgx.TxOptions{})
				pool.ExpectQuery(expectedDeleteExtensionBuiltinSQL).
					WithArgs(extensionID).
					WillReturnError(databaseError)
				pool.ExpectRollback()
			},
			context: "failure preparing to delete extension",
		},
		{
			name: "error_-_rejects_builtin_extension",
			setupMock: func(pool pgxmock.PgxPoolIface) {
				pool.ExpectBeginTx(pgx.TxOptions{})
				pool.ExpectQuery(expectedDeleteExtensionBuiltinSQL).
					WithArgs(extensionID).
					WillReturnRows(pool.NewRows([]string{"is_builtin"}).AddRow(true))
				pool.ExpectRollback()
			},
			wantErr: services.ErrGraphExtensionBuiltIn,
		},
		{
			name: "error_-_isolates_node_kind_fetch_error",
			setupMock: func(pool pgxmock.PgxPoolIface) {
				pool.ExpectBeginTx(pgx.TxOptions{})
				pool.ExpectQuery(expectedDeleteExtensionBuiltinSQL).
					WithArgs(extensionID).
					WillReturnRows(pool.NewRows([]string{"is_builtin"}).AddRow(false))
				pool.ExpectQuery(`SELECT nk.id, nk.kind_id, k.name, nk.schema_extension_id, nk.display_name, nk.description, nk.is_display_kind, nk.icon, nk.icon_color, nk.created_at, nk.updated_at, nk.deleted_at FROM schema_node_kinds AS nk JOIN kind AS k ON nk.kind_id = k.id WHERE (nk.schema_extension_id = $1) ORDER BY nk.id ASC`).
					WithArgs(fmt.Sprint(extensionID)).
					WillReturnError(databaseError)
				pool.ExpectRollback()
			},
			context: "failed to fetch node kinds",
		},
		{
			name: "error_-_returns_extension_not_found_when_delete_has_no_row",
			setupMock: func(pool pgxmock.PgxPoolIface) {
				pool.ExpectBeginTx(pgx.TxOptions{})
				pool.ExpectQuery(expectedDeleteExtensionBuiltinSQL).
					WithArgs(extensionID).
					WillReturnRows(pool.NewRows([]string{"is_builtin"}).AddRow(false))
				pool.ExpectQuery(`SELECT nk.id, nk.kind_id, k.name, nk.schema_extension_id, nk.display_name, nk.description, nk.is_display_kind, nk.icon, nk.icon_color, nk.created_at, nk.updated_at, nk.deleted_at FROM schema_node_kinds AS nk JOIN kind AS k ON nk.kind_id = k.id WHERE (nk.schema_extension_id = $1) ORDER BY nk.id ASC`).
					WithArgs(fmt.Sprint(extensionID)).
					WillReturnRows(pool.NewRows([]string{"id", "kind_id", "name", "schema_extension_id", "display_name", "description", "is_display_kind", "icon", "icon_color", "created_at", "updated_at", "deleted_at"}))
				pool.ExpectQuery(expectedDeleteExtensionSQL).
					WithArgs(extensionID).
					WillReturnRows(pool.NewRows([]string{"id"}))
				pool.ExpectRollback()
			},
			wantErr: services.ErrExtensionNotFound,
		},
		{
			name: "error_-_isolates_commit_error",
			setupMock: func(pool pgxmock.PgxPoolIface) {
				pool.ExpectBeginTx(pgx.TxOptions{})
				pool.ExpectQuery(expectedDeleteExtensionBuiltinSQL).
					WithArgs(extensionID).
					WillReturnRows(pool.NewRows([]string{"is_builtin"}).AddRow(false))
				pool.ExpectQuery(`SELECT nk.id, nk.kind_id, k.name, nk.schema_extension_id, nk.display_name, nk.description, nk.is_display_kind, nk.icon, nk.icon_color, nk.created_at, nk.updated_at, nk.deleted_at FROM schema_node_kinds AS nk JOIN kind AS k ON nk.kind_id = k.id WHERE (nk.schema_extension_id = $1) ORDER BY nk.id ASC`).
					WithArgs(fmt.Sprint(extensionID)).
					WillReturnRows(pool.NewRows([]string{"id", "kind_id", "name", "schema_extension_id", "display_name", "description", "is_display_kind", "icon", "icon_color", "created_at", "updated_at", "deleted_at"}))
				pool.ExpectQuery(expectedDeleteExtensionSQL).
					WithArgs(extensionID).
					WillReturnRows(pool.NewRows([]string{"id"}).AddRow(extensionID))
				pool.ExpectCommit().WillReturnError(databaseError)
			},
			context: "committing extension deletion",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			store, pool := newTestStore(t)
			test.setupMock(pool)

			err := store.DeleteExtension(ctx, extensionID)
			if test.context != "" {
				require.ErrorContains(t, err, test.context)
			} else {
				require.ErrorIs(t, err, test.wantErr)
			}
			require.NoError(t, pool.ExpectationsWereMet())
		})
	}
}
