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
	"github.com/specterops/bloodhound/server/extensions/internal/services"
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
			name: "error_-_propagates_database_error",
			expectations: func(pool pgxmock.PgxPoolIface) {
				pool.ExpectQuery(expectedGetExtensionSQL).WithArgs(extID).WillReturnError(dbErr)
			},
			wantExtension: services.Extension{},
			wantErr:       dbErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store, pool := newTestStore(t)
			tt.expectations(pool)

			extension, err := store.GetExtension(ctx, extID)
			assert.Equal(t, tt.wantExtension, extension)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
			}
			require.NoError(t, pool.ExpectationsWereMet())
		})
	}
}
