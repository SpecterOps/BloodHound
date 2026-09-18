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

	"github.com/pashagolub/pgxmock/v4"
	"github.com/specterops/bloodhound/server/opengraphschema/internal/appdb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestStore(t *testing.T) (*appdb.Store, pgxmock.PgxPoolIface) {
	t.Helper()

	pool, err := pgxmock.NewPool()
	require.NoError(t, err)

	t.Cleanup(func() {
		require.NoError(t, pool.ExpectationsWereMet())
		pool.Close()
	})

	return appdb.NewStore(pool), pool
}

func TestStore_GetEnvironmentsFiltered(t *testing.T) {
	t.Parallel()

	columns := []string{
		"id", "schema_extension_id", "schema_extension_display_name",
		"environment_kind_id", "environment_kind_name", "source_kind_id",
		"created_at", "updated_at", "deleted_at",
	}

	unexpectedErr := errors.New("db unavailable")

	tests := []struct {
		name        string
		onlyBuiltin bool
		setupMock   func(pool pgxmock.PgxPoolIface)
		wantKinds   []string
		wantErr     error
	}{
		{
			name:        "Success: returns all environments when not builtin only",
			onlyBuiltin: false,
			setupMock: func(pool pgxmock.PgxPoolIface) {
				pool.ExpectQuery(`INNER JOIN schema_extensions ext ON se\.schema_extension_id = ext\.id\s+ORDER BY se\.id`).
					WillReturnRows(pool.NewRows(columns).
						AddRow(int32(1), int32(10), "Ext A", int32(100), "AZBase", int32(200), nil, nil, nil).
						AddRow(int32(2), int32(11), "Ext B", int32(101), "Base", int32(201), nil, nil, nil))
			},
			wantKinds: []string{"AZBase", "Base"},
		},
		{
			name:        "Success: applies builtin filter",
			onlyBuiltin: true,
			setupMock: func(pool pgxmock.PgxPoolIface) {
				pool.ExpectQuery(`WHERE ext\.is_builtin = true`).
					WillReturnRows(pool.NewRows(columns).
						AddRow(int32(1), int32(10), "Ext A", int32(100), "AZBase", int32(200), nil, nil, nil))
			},
			wantKinds: []string{"AZBase"},
		},
		{
			name:        "Error: propagates query error",
			onlyBuiltin: false,
			setupMock: func(pool pgxmock.PgxPoolIface) {
				pool.ExpectQuery(`FROM schema_environments`).
					WillReturnError(unexpectedErr)
			},
			wantErr: unexpectedErr,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			store, pool := newTestStore(t)
			test.setupMock(pool)

			environments, err := store.GetEnvironmentsFiltered(context.Background(), test.onlyBuiltin)
			if test.wantErr != nil {
				assert.ErrorIs(t, err, test.wantErr)
				return
			}

			require.NoError(t, err)
			var kinds []string
			for _, environment := range environments {
				kinds = append(kinds, environment.EnvironmentKindName)
			}
			assert.Equal(t, test.wantKinds, kinds)
		})
	}
}
