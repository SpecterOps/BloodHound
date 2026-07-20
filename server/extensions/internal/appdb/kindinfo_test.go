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
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/pashagolub/pgxmock/v4"
	"github.com/specterops/bloodhound/server/extensions/internal/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Literal SQL expected by the Store, compared via pgxmock.QueryMatcherEqual (whitespace
// normalised). Column order, aliases, JOIN, WHERE and ORDER BY shape are load-bearing.
const expectedGetKindInfosSQL = `SELECT ki.id, ki.kind_id, ki.node_kind_id, ki.relationship_kind_id, k.name, ki.info_key, ki.title, ki.position, ki.content, ki.created_at, ki.updated_at FROM schema_kind_info AS ki JOIN kind AS k ON ki.kind_id = k.id WHERE ki.node_kind_id = $1 ORDER BY ki.position, ki.title`

func kindInfoColumns() []string {
	return []string{
		"id", "kind_id", "node_kind_id", "relationship_kind_id", "name",
		"info_key", "title", "position", "content", "created_at", "updated_at",
	}
}

func TestStore_GetKindInfosByNodeKindID(t *testing.T) {
	var (
		ctx        = context.Background()
		nodeKindID = int32(42)
		createdAt  = time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
		updatedAt  = time.Date(2026, 2, 3, 4, 5, 6, 0, time.UTC)
		content    = json.RawMessage(`{"markdown":"hi"}`)
		dbErr      = errors.New("db error")
	)

	tests := []struct {
		name          string
		expectations  func(pool pgxmock.PgxPoolIface)
		wantKindInfos []services.KindInfo
		wantErr       error
	}{
		{
			name: "success_-_returns_ordered_infos_with_name",
			expectations: func(pool pgxmock.PgxPoolIface) {
				pool.ExpectQuery(expectedGetKindInfosSQL).WithArgs(nodeKindID).WillReturnRows(
					pool.NewRows(kindInfoColumns()).
						AddRow(int32(1), int32(99), &nodeKindID, (*int32)(nil), "User",
							"overview", "Overview", int32(0), content, createdAt, updatedAt).
						AddRow(int32(2), int32(99), &nodeKindID, (*int32)(nil), "User",
							"details", "Details", int32(1), content, createdAt, updatedAt),
				)
			},
			wantKindInfos: []services.KindInfo{
				{ID: 1, KindID: 99, NodeKindID: &nodeKindID, Name: "User", InfoKey: "overview",
					Title: "Overview", Position: 0, Content: content, CreatedAt: createdAt, UpdatedAt: updatedAt},
				{ID: 2, KindID: 99, NodeKindID: &nodeKindID, Name: "User", InfoKey: "details",
					Title: "Details", Position: 1, Content: content, CreatedAt: createdAt, UpdatedAt: updatedAt},
			},
		},
		{
			name: "success_-_returns_empty_slice_when_no_infos",
			expectations: func(pool pgxmock.PgxPoolIface) {
				pool.ExpectQuery(expectedGetKindInfosSQL).WithArgs(nodeKindID).WillReturnRows(
					pool.NewRows(kindInfoColumns()),
				)
			},
			wantKindInfos: []services.KindInfo{},
		},
		{
			name: "error_-_propagates_database_error",
			expectations: func(pool pgxmock.PgxPoolIface) {
				pool.ExpectQuery(expectedGetKindInfosSQL).WithArgs(nodeKindID).WillReturnError(dbErr)
			},
			wantKindInfos: nil,
			wantErr:       dbErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store, pool := newTestStore(t)
			tt.expectations(pool)

			kindInfos, err := store.GetKindInfosByNodeKindID(ctx, nodeKindID)
			assert.Equal(t, tt.wantKindInfos, kindInfos)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
			}
			require.NoError(t, pool.ExpectationsWereMet())
		})
	}
}
