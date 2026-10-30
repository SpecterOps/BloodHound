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
	"time"

	"github.com/specterops/bloodhound/server/opengraphschema/internal/services"

	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const expectedGetEnvironmentsSQL = `SELECT
 se.id, se.schema_extension_id, ext.display_name AS schema_extension_display_name,
 se.environment_kind_id, k.name AS environment_kind_name, se.source_kind_id,
 se.created_at, se.updated_at, se.deleted_at
 FROM schema_environments se
 INNER JOIN kind k ON se.environment_kind_id = k.id
 INNER JOIN schema_extensions ext ON se.schema_extension_id = ext.id`

func TestStore_GetEnvironmentsFiltered(t *testing.T) {
	t.Parallel()

	columns := []string{
		"id", "schema_extension_id", "schema_extension_display_name",
		"environment_kind_id", "environment_kind_name", "source_kind_id",
		"created_at", "updated_at", "deleted_at",
	}

	var (
		unexpectedErr     = errors.New("db unavailable")
		timestamp         = time.Date(2026, 10, 1, 12, 0, 0, 0, time.UTC)
		firstEnvironment  = services.Environment{ID: 1, SchemaExtensionID: 10, SchemaExtensionDisplayName: "Ext A", EnvironmentKindID: 100, EnvironmentKindName: "AZBase", SourceKindID: 200, CreatedAt: timestamp, UpdatedAt: timestamp}
		secondEnvironment = services.Environment{ID: 2, SchemaExtensionID: 11, SchemaExtensionDisplayName: "Ext B", EnvironmentKindID: 101, EnvironmentKindName: "Base", SourceKindID: 201, CreatedAt: timestamp, UpdatedAt: timestamp, DeletedAt: &timestamp}
	)

	tests := []struct {
		name             string
		onlyBuiltin      bool
		setupMock        func(pool pgxmock.PgxPoolIface)
		wantEnvironments []services.Environment
		errorContext     string
	}{
		{
			name:        "Success: returns all environments when not builtin only",
			onlyBuiltin: false,
			setupMock: func(pool pgxmock.PgxPoolIface) {
				pool.ExpectQuery(expectedGetEnvironmentsSQL + " ORDER BY se.id").
					WillReturnRows(pool.NewRows(columns).
						AddRow(int32(1), int32(10), "Ext A", int32(100), "AZBase", int32(200), timestamp, timestamp, nil).
						AddRow(int32(2), int32(11), "Ext B", int32(101), "Base", int32(201), timestamp, timestamp, timestamp))
			},
			wantEnvironments: []services.Environment{firstEnvironment, secondEnvironment},
		},
		{
			name:        "Success: applies builtin filter",
			onlyBuiltin: true,
			setupMock: func(pool pgxmock.PgxPoolIface) {
				pool.ExpectQuery(expectedGetEnvironmentsSQL + " WHERE ext.is_builtin = true ORDER BY se.id").
					WillReturnRows(pool.NewRows(columns).
						AddRow(int32(1), int32(10), "Ext A", int32(100), "AZBase", int32(200), timestamp, timestamp, nil))
			},
			wantEnvironments: []services.Environment{firstEnvironment},
		},
		{
			name:        "Error: isolates query error",
			onlyBuiltin: false,
			setupMock: func(pool pgxmock.PgxPoolIface) {
				pool.ExpectQuery(expectedGetEnvironmentsSQL + " ORDER BY se.id").
					WillReturnError(unexpectedErr)
			},
			errorContext: "listing schema environments",
		},
		{
			name: "Success: returns initialized empty environments",
			setupMock: func(pool pgxmock.PgxPoolIface) {
				pool.ExpectQuery(expectedGetEnvironmentsSQL + " ORDER BY se.id").WillReturnRows(pool.NewRows(columns))
			},
			wantEnvironments: []services.Environment{},
		},
		{
			name: "Error: isolates row errors",
			setupMock: func(pool pgxmock.PgxPoolIface) {
				pool.ExpectQuery(expectedGetEnvironmentsSQL + " ORDER BY se.id").WillReturnRows(pool.NewRows(columns).AddRow(int32(1), int32(10), "Ext A", int32(100), "AZBase", int32(200), timestamp, timestamp, nil).RowError(0, unexpectedErr))
			},
			errorContext: "reading schema environments",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			store, pool := newTestStore(t)
			test.setupMock(pool)

			environments, err := store.GetEnvironmentsFiltered(context.Background(), test.onlyBuiltin)
			require.NoError(t, pool.ExpectationsWereMet())
			if test.errorContext != "" {
				require.ErrorContains(t, err, test.errorContext)
				assert.NotErrorIs(t, err, unexpectedErr)
				assert.Nil(t, errors.Unwrap(err))
				return
			}

			require.NoError(t, err)
			assert.Equal(t, test.wantEnvironments, environments)
		})
	}
}
