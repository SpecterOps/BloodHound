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

	"github.com/pashagolub/pgxmock/v6"
	"github.com/specterops/bloodhound/server/opengraphschema/internal/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// expectedGetExtensionSQL is the literal SQL the Store builds for GetExtension.
const expectedGetExtensionSQL = `SELECT id, name, display_name, version, is_builtin, namespace FROM schema_extensions WHERE id = $1`

func extensionColumns() []string {
	return []string{"id", "name", "display_name", "version", "is_builtin", "namespace"}
}

func TestStore_GetExtension(t *testing.T) {
	var (
		ctx   = context.Background()
		extID = int32(7)
		dbErr = errors.New("db error")
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
						int32(7), "TestExtension", "Test Extension", "1.0.0", false, "TST",
					),
				)
			},
			wantExtension: services.Extension{
				ID: 7, Name: "TestExtension", DisplayName: "Test Extension",
				Version: "1.0.0", IsBuiltin: false, Namespace: "TST",
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
				pool.ExpectQuery(expectedGetExtensionSQL).WithArgs(extID).WillReturnRows(pool.NewRows(extensionColumns()).AddRow(int32(7), "TestExtension", "Test Extension", "1.0.0", false, "TST").RowError(0, dbErr))
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

func TestStore_ListExtensions(t *testing.T) {
	t.Parallel()
	var (
		databaseError = errors.New("driver failure")
		tests         = []struct {
			name         string
			setupMock    func(pgxmock.PgxPoolIface)
			want         []services.Extension
			errorContext string
		}{
			{name: "maps built-in and custom extensions in query order", setupMock: func(pool pgxmock.PgxPoolIface) {
				pool.ExpectQuery(expectedListExtensionsSQL).WillReturnRows(pool.NewRows(extensionColumns()).
					AddRow(int32(1), "AD", "Active Directory", "v1", true, "AD").
					AddRow(int32(2), "custom", "Custom Extension", "v2", false, "CUSTOM"))
			}, want: []services.Extension{
				{ID: 1, Name: "AD", DisplayName: "Active Directory", Version: "v1", IsBuiltin: true, Namespace: "AD"},
				{ID: 2, Name: "custom", DisplayName: "Custom Extension", Version: "v2", Namespace: "CUSTOM"},
			}},
			{name: "empty", setupMock: func(pool pgxmock.PgxPoolIface) {
				pool.ExpectQuery(expectedListExtensionsSQL).WillReturnRows(pool.NewRows(extensionColumns()))
			}, want: []services.Extension{}},
			{name: "query error", setupMock: func(pool pgxmock.PgxPoolIface) {
				pool.ExpectQuery(expectedListExtensionsSQL).WillReturnError(databaseError)
			}, errorContext: "listing extensions"},
			{name: "row error", setupMock: func(pool pgxmock.PgxPoolIface) {
				pool.ExpectQuery(expectedListExtensionsSQL).WillReturnRows(pool.NewRows(extensionColumns()).
					AddRow(int32(1), "AD", "Active Directory", "v1", true, "AD").RowError(0, databaseError))
			}, errorContext: "reading extensions"},
			{name: "scan error", setupMock: func(pool pgxmock.PgxPoolIface) {
				pool.ExpectQuery(expectedListExtensionsSQL).WillReturnRows(pool.NewRows(extensionColumns()).
					AddRow("invalid ID", "AD", "Active Directory", "v1", true, "AD"))
			}, errorContext: "reading extensions"},
		}
	)
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			var store, pool = newTestStore(t)
			testCase.setupMock(pool)
			extensions, err := store.ListExtensions(context.Background())
			assert.Equal(t, testCase.want, extensions)
			if testCase.errorContext != "" {
				require.ErrorContains(t, err, testCase.errorContext)
				assert.NotErrorIs(t, err, databaseError)
				assert.Nil(t, errors.Unwrap(err))
			} else {
				require.NoError(t, err)
			}
			require.NoError(t, pool.ExpectationsWereMet())
		})
	}
}

const expectedListExtensionsSQL = `SELECT id, name, display_name, version, is_builtin, namespace FROM schema_extensions ORDER BY name ASC`
