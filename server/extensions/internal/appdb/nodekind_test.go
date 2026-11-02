// Copyright 2026 Specter Ops, Inc.
//
// Licensed under the Apache License, Version 2.0
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
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

	"github.com/pashagolub/pgxmock/v4"
	"github.com/specterops/bloodhound/packages/go/params"
	"github.com/specterops/bloodhound/server/extensions/internal/appdb"
	"github.com/specterops/bloodhound/server/extensions/internal/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Literal SQL expected by the Store, compared via pgxmock.QueryMatcherEqual (whitespace
// normalised). Column order, aliases, JOIN and WHERE shape are load-bearing.
const expectedGetNodeKindSQL = `SELECT nk.id, nk.schema_extension_id, nk.kind_id, k.name, nk.display_name, nk.description, nk.is_display_kind, nk.icon, nk.icon_color, nk.created_at, nk.updated_at, nk.deleted_at FROM schema_node_kinds AS nk JOIN kind AS k ON nk.kind_id = k.id WHERE nk.id = $1`

func newTestStore(t *testing.T) (*appdb.Store, pgxmock.PgxPoolIface) {
	t.Helper()
	pool, err := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual))
	require.NoError(t, err)
	t.Cleanup(pool.Close)
	return appdb.NewStore(pool), pool
}

func nodeKindColumns() []string {
	return []string{
		"id", "schema_extension_id", "kind_id", "name", "display_name", "description",
		"is_display_kind", "icon", "icon_color", "created_at", "updated_at", "deleted_at",
	}
}

func TestStore_GetNodeKind(t *testing.T) {
	var (
		ctx        = context.Background()
		nodeKindID = int32(42)
		createdAt  = time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
		updatedAt  = time.Date(2026, 2, 3, 4, 5, 6, 0, time.UTC)
		dbErr      = errors.New("db error")
	)

	tests := []struct {
		name         string
		expectations func(pool pgxmock.PgxPoolIface)
		wantNodeKind services.NodeKind
		wantErr      error
	}{
		{
			name: "success_-_maps_all_columns",
			expectations: func(pool pgxmock.PgxPoolIface) {
				pool.ExpectQuery(expectedGetNodeKindSQL).WithArgs(nodeKindID).WillReturnRows(
					pool.NewRows(nodeKindColumns()).AddRow(
						int32(42), int32(7), int32(99), "User", "User", "a user",
						true, "user", "#fff", createdAt, updatedAt, nil,
					),
				)
			},
			wantNodeKind: services.NodeKind{
				ID: 42, SchemaExtensionID: 7, KindID: 99, Name: "User", DisplayName: "User",
				Description: "a user", IsDisplayKind: true, Icon: "user", Color: "#fff",
				CreatedAt: createdAt, UpdatedAt: updatedAt, DeletedAt: nil,
			},
		},
		{
			name: "error_-_maps_no_rows_to_ErrNodeKindNotFound",
			expectations: func(pool pgxmock.PgxPoolIface) {
				pool.ExpectQuery(expectedGetNodeKindSQL).WithArgs(nodeKindID).WillReturnRows(
					pool.NewRows(nodeKindColumns()),
				)
			},
			wantNodeKind: services.NodeKind{},
			wantErr:      services.ErrNodeKindNotFound,
		},
		{
			name: "error_-_propagates_database_error",
			expectations: func(pool pgxmock.PgxPoolIface) {
				pool.ExpectQuery(expectedGetNodeKindSQL).WithArgs(nodeKindID).WillReturnError(dbErr)
			},
			wantNodeKind: services.NodeKind{},
			wantErr:      dbErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store, pool := newTestStore(t)
			tt.expectations(pool)

			nodeKind, err := store.GetNodeKind(ctx, nodeKindID)
			assert.Equal(t, tt.wantNodeKind, nodeKind)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
			}
			require.NoError(t, pool.ExpectationsWereMet())
		})
	}
}

const expectedCreateGraphSchemaNodeKindSQL = `WITH dawgs_kind (id, name) AS (SELECT id, name FROM upsert_kind($1)), inserted_schema_node AS (INSERT INTO schema_node_kinds (kind_id, schema_extension_id, display_name, description, is_display_kind, icon, icon_color) SELECT dk.id, $2, $3, $4, $5, $6, $7 FROM dawgs_kind AS dk RETURNING id, kind_id, schema_extension_id, display_name, description, is_display_kind, icon, icon_color, created_at, updated_at, deleted_at) SELECT isn.id, isn.kind_id, isn.schema_extension_id, dk.name, isn.display_name, isn.description, isn.is_display_kind, isn.icon, isn.icon_color, isn.created_at, isn.updated_at, isn.deleted_at FROM inserted_schema_node AS isn JOIN dawgs_kind AS dk ON isn.kind_id = dk.id`

func TestStore_CreateGraphSchemaNodeKind(t *testing.T) {
	t.Parallel()

	type args struct {
		ctx         context.Context
		extensionID int32
		nodeKind    services.NodeInput
	}

	type testCase struct {
		name         string
		args         args
		setupMock    func(pgxmock.PgxPoolIface)
		wantNodeKind services.NodeKind
		wantErr      error
	}

	var (
		ctx      = context.Background()
		nodeKind = services.NodeKind{
			ID:                42,
			SchemaExtensionID: 7,
			KindID:            99,
			Name:              "User",
			DisplayName:       "User",
			Description:       "a user",
			IsDisplayKind:     true,
			Icon:              "user",
			Color:             "#fff",
			CreatedAt:         time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC),
			UpdatedAt:         time.Date(2026, 2, 3, 4, 5, 6, 0, time.UTC),
		}
		nodeKindInput = services.NodeInput{
			Name:          nodeKind.Name,
			DisplayName:   nodeKind.DisplayName,
			Description:   nodeKind.Description,
			IsDisplayKind: nodeKind.IsDisplayKind,
			Icon:          nodeKind.Icon,
			IconColor:     nodeKind.Color,
		}
		databaseError = errors.New("db error")
	)

	tests := []testCase{
		{
			name: "success_-_maps_all_columns",
			args: args{ctx: ctx, extensionID: nodeKind.SchemaExtensionID, nodeKind: nodeKindInput},
			setupMock: func(pool pgxmock.PgxPoolIface) {
				pool.ExpectQuery(expectedCreateGraphSchemaNodeKindSQL).
					WithArgs(nodeKindInput.Name, nodeKind.SchemaExtensionID, nodeKindInput.DisplayName,
						nodeKindInput.Description, nodeKindInput.IsDisplayKind, nodeKindInput.Icon, nodeKindInput.IconColor).
					WillReturnRows(pool.NewRows(nodeKindColumns()).AddRow(
						nodeKind.ID, nodeKind.SchemaExtensionID, nodeKind.KindID, nodeKind.Name, nodeKind.DisplayName,
						nodeKind.Description, nodeKind.IsDisplayKind, nodeKind.Icon, nodeKind.Color,
						nodeKind.CreatedAt, nodeKind.UpdatedAt, nil,
					)).RowsWillBeClosed()
			},
			wantNodeKind: nodeKind,
		},
		{
			name: "error_-_propagates_query_error",
			args: args{ctx: ctx, extensionID: nodeKind.SchemaExtensionID, nodeKind: nodeKindInput},
			setupMock: func(pool pgxmock.PgxPoolIface) {
				pool.ExpectQuery(expectedCreateGraphSchemaNodeKindSQL).
					WithArgs(nodeKindInput.Name, nodeKind.SchemaExtensionID, nodeKindInput.DisplayName,
						nodeKindInput.Description, nodeKindInput.IsDisplayKind, nodeKindInput.Icon, nodeKindInput.IconColor).
					WillReturnError(databaseError)
			},
			wantErr: databaseError,
		},
		{
			name: "error_-_propagates_row_completion_error",
			args: args{ctx: ctx, extensionID: nodeKind.SchemaExtensionID, nodeKind: nodeKindInput},
			setupMock: func(pool pgxmock.PgxPoolIface) {
				pool.ExpectQuery(expectedCreateGraphSchemaNodeKindSQL).
					WithArgs(nodeKindInput.Name, nodeKind.SchemaExtensionID, nodeKindInput.DisplayName,
						nodeKindInput.Description, nodeKindInput.IsDisplayKind, nodeKindInput.Icon, nodeKindInput.IconColor).
					WillReturnRows(pool.NewRows(nodeKindColumns()).CloseError(databaseError)).
					RowsWillBeClosed()
			},
			wantErr: databaseError,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			store, pool := newTestStore(t)
			test.setupMock(pool)

			nodeKind, err := store.CreateGraphSchemaNodeKind(
				test.args.ctx, test.args.extensionID, test.args.nodeKind,
			)
			assert.Equal(t, test.wantNodeKind, nodeKind)
			assert.ErrorIs(t, err, test.wantErr)
			require.NoError(t, pool.ExpectationsWereMet())
		})
	}
}

const (
	expectedGetGraphSchemaNodeKindsSQL   = `SELECT nk.id, nk.kind_id, k.name, nk.schema_extension_id, nk.display_name, nk.description, nk.is_display_kind, nk.icon, nk.icon_color, nk.created_at, nk.updated_at, nk.deleted_at FROM schema_node_kinds AS nk JOIN kind AS k ON nk.kind_id = k.id`
	expectedCountGraphSchemaNodeKindsSQL = `SELECT COUNT(*) FROM schema_node_kinds AS nk JOIN kind AS k ON nk.kind_id = k.id`
)

func TestStore_GetGraphSchemaNodeKinds(t *testing.T) {
	t.Parallel()

	type args struct {
		ctx     context.Context
		filters params.Filters
		sort    params.SortItems
		skip    int
		limit   int
	}

	type testCase struct {
		name          string
		args          args
		setupMock     func(pgxmock.PgxPoolIface)
		wantNodeKinds []services.NodeKind
		wantCount     int
		wantErr       error
	}

	var (
		ctx           = context.Background()
		databaseError = errors.New("db error")
		nodeKind      = services.NodeKind{
			ID:                42,
			SchemaExtensionID: 7,
			KindID:            99,
			Name:              "User",
			DisplayName:       "User",
			Description:       "a user",
			IsDisplayKind:     true,
			Icon:              "user",
			Color:             "#fff",
			CreatedAt:         time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC),
			UpdatedAt:         time.Date(2026, 2, 3, 4, 5, 6, 0, time.UTC),
		}
	)

	tests := []testCase{
		{
			name: "success_-_returns_node_kinds_and_total",
			args: args{ctx: ctx},
			setupMock: func(pool pgxmock.PgxPoolIface) {
				pool.ExpectQuery(expectedGetGraphSchemaNodeKindsSQL + " ORDER BY nk.id ASC").
					WillReturnRows(pool.NewRows(nodeKindColumns()).AddRow(
						nodeKind.ID, nodeKind.SchemaExtensionID, nodeKind.KindID, nodeKind.Name, nodeKind.DisplayName,
						nodeKind.Description, nodeKind.IsDisplayKind, nodeKind.Icon, nodeKind.Color,
						nodeKind.CreatedAt, nodeKind.UpdatedAt, nil,
					)).RowsWillBeClosed()
			},
			wantNodeKinds: []services.NodeKind{nodeKind},
			wantCount:     1,
		},
		{
			name: "error_-_propagates_query_error",
			args: args{ctx: ctx},
			setupMock: func(pool pgxmock.PgxPoolIface) {
				pool.ExpectQuery(expectedGetGraphSchemaNodeKindsSQL + " ORDER BY nk.id ASC").
					WillReturnError(databaseError)
			},
			wantErr: fmt.Errorf("error fetching node kinds: %w", databaseError),
		},
		{
			name: "error_-_propagates_row_completion_error",
			args: args{ctx: ctx},
			setupMock: func(pool pgxmock.PgxPoolIface) {
				pool.ExpectQuery(expectedGetGraphSchemaNodeKindsSQL + " ORDER BY nk.id ASC").
					WillReturnRows(pool.NewRows(nodeKindColumns()).CloseError(databaseError)).
					RowsWillBeClosed()
			},
			wantErr: fmt.Errorf("error reading rows: %w", databaseError),
		},
		{
			name: "success_-_preserves_total_for_empty_page",
			args: args{
				ctx: ctx,
				filters: params.Filters{
					"name": {
						{Operator: params.ApproximatelyEquals, Value: nodeKind.Name},
					},
				},
				sort: params.SortItems{
					{Field: "name", Direction: params.Descending},
					{Field: "id", Direction: params.Ascending},
				},
				skip:  3,
				limit: 1,
			},
			setupMock: func(pool pgxmock.PgxPoolIface) {
				pool.ExpectQuery(expectedGetGraphSchemaNodeKindsSQL+" WHERE (k.name ILIKE $1) ORDER BY k.name DESC, nk.id ASC LIMIT $2 OFFSET $3").
					WithArgs("%"+nodeKind.Name+"%", 1, 3).
					WillReturnRows(pool.NewRows(nodeKindColumns())).RowsWillBeClosed()
				pool.ExpectQuery(expectedCountGraphSchemaNodeKindsSQL + " WHERE (k.name ILIKE $1)").
					WithArgs("%" + nodeKind.Name + "%").
					WillReturnRows(pool.NewRows([]string{"count"}).AddRow(3)).RowsWillBeClosed()
			},
			wantNodeKinds: []services.NodeKind{},
			wantCount:     3,
		},
		{
			name: "error_-_propagates_count_query_error",
			args: args{ctx: ctx, limit: 1},
			setupMock: func(pool pgxmock.PgxPoolIface) {
				pool.ExpectQuery(expectedGetGraphSchemaNodeKindsSQL + " ORDER BY nk.id ASC LIMIT $1").
					WithArgs(1).
					WillReturnRows(pool.NewRows(nodeKindColumns())).RowsWillBeClosed()
				pool.ExpectQuery(expectedCountGraphSchemaNodeKindsSQL).WillReturnError(databaseError)
			},
			wantNodeKinds: []services.NodeKind{},
			wantErr:       fmt.Errorf("error counting node kinds: %w", databaseError),
		},
		{
			name: "error_-_propagates_count_row_completion_error",
			args: args{ctx: ctx, limit: 1},
			setupMock: func(pool pgxmock.PgxPoolIface) {
				pool.ExpectQuery(expectedGetGraphSchemaNodeKindsSQL + " ORDER BY nk.id ASC LIMIT $1").
					WithArgs(1).
					WillReturnRows(pool.NewRows(nodeKindColumns())).RowsWillBeClosed()
				pool.ExpectQuery(expectedCountGraphSchemaNodeKindsSQL).
					WillReturnRows(pool.NewRows([]string{"count"}).CloseError(databaseError)).
					RowsWillBeClosed()
			},
			wantNodeKinds: []services.NodeKind{},
			wantErr:       fmt.Errorf("error reading node kind count: %w", databaseError),
		},
		{
			name: "error_-_rejects_unknown_filter_field",
			args: args{
				ctx: ctx,
				filters: params.Filters{
					"unknown": {
						{Operator: params.Equals, Value: nodeKind.Name},
					},
				},
			},
			setupMock:     func(pgxmock.PgxPoolIface) {},
			wantNodeKinds: []services.NodeKind{},
			wantErr:       errors.New(`node kind filter references unknown field "unknown"`),
		},
		{
			name: "error_-_rejects_invalid_sort_direction",
			args: args{
				ctx: ctx,
				sort: params.SortItems{
					{Field: "name", Direction: "sideways"},
				},
			},
			setupMock:     func(pgxmock.PgxPoolIface) {},
			wantNodeKinds: []services.NodeKind{},
			wantErr:       fmt.Errorf("%w: %q", services.ErrInvalidSortDirection, "sideways"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			store, pool := newTestStore(t)
			test.setupMock(pool)

			nodeKinds, count, err := store.GetGraphSchemaNodeKinds(
				test.args.ctx, test.args.filters, test.args.sort, test.args.skip, test.args.limit,
			)
			assert.Equal(t, test.wantNodeKinds, nodeKinds)
			assert.Equal(t, test.wantCount, count)
			if test.wantErr == nil {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, test.wantErr.Error())
				if cause := errors.Unwrap(test.wantErr); cause != nil {
					assert.ErrorIs(t, err, cause)
				}
			}
			require.NoError(t, pool.ExpectationsWereMet())
		})
	}
}

func TestStore_GetGraphSchemaNodeKindsByExtensionId(t *testing.T) {
	t.Parallel()

	type args struct {
		ctx         context.Context
		extensionID int32
	}

	type testCase struct {
		name          string
		args          args
		setupMock     func(pgxmock.PgxPoolIface)
		wantNodeKinds []services.NodeKind
		wantErr       error
	}

	var (
		ctx           = context.Background()
		databaseError = errors.New("db error")
		nodeKind      = services.NodeKind{
			ID:                42,
			SchemaExtensionID: 7,
			KindID:            99,
			Name:              "User",
		}
	)

	tests := []testCase{
		{
			name: "success_-_returns_node_kinds_for_extension",
			args: args{ctx: ctx, extensionID: nodeKind.SchemaExtensionID},
			setupMock: func(pool pgxmock.PgxPoolIface) {
				pool.ExpectQuery(expectedGetGraphSchemaNodeKindsSQL + " WHERE (nk.schema_extension_id = $1) ORDER BY nk.id ASC").
					WithArgs(fmt.Sprint(nodeKind.SchemaExtensionID)).
					WillReturnRows(pool.NewRows(nodeKindColumns()).AddRow(
						nodeKind.ID, nodeKind.SchemaExtensionID, nodeKind.KindID, nodeKind.Name, nodeKind.DisplayName,
						nodeKind.Description, nodeKind.IsDisplayKind, nodeKind.Icon, nodeKind.Color,
						nodeKind.CreatedAt, nodeKind.UpdatedAt, nil,
					)).RowsWillBeClosed()
			},
			wantNodeKinds: []services.NodeKind{nodeKind},
		},
		{
			name: "error_-_propagates_database_error",
			args: args{ctx: ctx, extensionID: nodeKind.SchemaExtensionID},
			setupMock: func(pool pgxmock.PgxPoolIface) {
				pool.ExpectQuery(expectedGetGraphSchemaNodeKindsSQL + " WHERE (nk.schema_extension_id = $1) ORDER BY nk.id ASC").
					WithArgs(fmt.Sprint(nodeKind.SchemaExtensionID)).
					WillReturnError(databaseError)
			},
			wantErr: databaseError,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			store, pool := newTestStore(t)
			test.setupMock(pool)

			nodeKinds, err := store.GetGraphSchemaNodeKindsByExtensionId(
				test.args.ctx, test.args.extensionID,
			)
			assert.Equal(t, test.wantNodeKinds, nodeKinds)
			assert.ErrorIs(t, err, test.wantErr)
			require.NoError(t, pool.ExpectationsWereMet())
		})
	}
}

const expectedUpdateGraphSchemaNodeKindSQL = `UPDATE schema_node_kinds AS nk SET schema_extension_id = $1, display_name = $2, description = $3, is_display_kind = $4, icon = $5, icon_color = $6, updated_at = NOW() FROM kind AS k WHERE nk.id = $7 AND nk.kind_id = k.id RETURNING nk.id, nk.kind_id, nk.schema_extension_id, k.name, nk.display_name, nk.description, nk.is_display_kind, nk.icon, nk.icon_color, nk.created_at, nk.updated_at, nk.deleted_at`

func TestStore_UpdateGraphSchemaNodeKind(t *testing.T) {
	t.Parallel()

	type args struct {
		ctx      context.Context
		nodeKind services.NodeKind
	}

	type testCase struct {
		name         string
		args         args
		setupMock    func(pgxmock.PgxPoolIface)
		wantNodeKind services.NodeKind
		wantErr      error
	}

	var (
		ctx      = context.Background()
		nodeKind = services.NodeKind{
			ID:                42,
			SchemaExtensionID: 7,
			KindID:            99,
			Name:              "User",
			DisplayName:       "User",
			Description:       "a user",
			IsDisplayKind:     true,
			Icon:              "user",
			Color:             "#fff",
			CreatedAt:         time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC),
			UpdatedAt:         time.Date(2026, 2, 3, 4, 5, 6, 0, time.UTC),
		}
		databaseError = errors.New("db error")
	)

	tests := []testCase{
		{
			name: "success_-_maps_all_columns",
			args: args{ctx: ctx, nodeKind: nodeKind},
			setupMock: func(pool pgxmock.PgxPoolIface) {
				pool.ExpectQuery(expectedUpdateGraphSchemaNodeKindSQL).
					WithArgs(
						nodeKind.SchemaExtensionID, nodeKind.DisplayName, nodeKind.Description,
						nodeKind.IsDisplayKind, nodeKind.Icon, nodeKind.Color, nodeKind.ID,
					).
					WillReturnRows(pool.NewRows(nodeKindColumns()).AddRow(
						nodeKind.ID, nodeKind.SchemaExtensionID, nodeKind.KindID, nodeKind.Name,
						nodeKind.DisplayName, nodeKind.Description, nodeKind.IsDisplayKind,
						nodeKind.Icon, nodeKind.Color, nodeKind.CreatedAt, nodeKind.UpdatedAt, nil,
					)).RowsWillBeClosed()
			},
			wantNodeKind: nodeKind,
		},
		{
			name: "error_-_node_kind_not_found",
			args: args{ctx: ctx, nodeKind: nodeKind},
			setupMock: func(pool pgxmock.PgxPoolIface) {
				pool.ExpectQuery(expectedUpdateGraphSchemaNodeKindSQL).
					WithArgs(
						nodeKind.SchemaExtensionID, nodeKind.DisplayName, nodeKind.Description,
						nodeKind.IsDisplayKind, nodeKind.Icon, nodeKind.Color, nodeKind.ID,
					).
					WillReturnRows(pool.NewRows(nodeKindColumns())).RowsWillBeClosed()
			},
			wantErr: services.ErrNodeKindNotFound,
		},
		{
			name: "error_-_propagates_query_error",
			args: args{ctx: ctx, nodeKind: nodeKind},
			setupMock: func(pool pgxmock.PgxPoolIface) {
				pool.ExpectQuery(expectedUpdateGraphSchemaNodeKindSQL).
					WithArgs(
						nodeKind.SchemaExtensionID, nodeKind.DisplayName, nodeKind.Description,
						nodeKind.IsDisplayKind, nodeKind.Icon, nodeKind.Color, nodeKind.ID,
					).
					WillReturnError(databaseError)
			},
			wantErr: databaseError,
		},
		{
			name: "error_-_propagates_row_completion_error",
			args: args{ctx: ctx, nodeKind: nodeKind},
			setupMock: func(pool pgxmock.PgxPoolIface) {
				pool.ExpectQuery(expectedUpdateGraphSchemaNodeKindSQL).
					WithArgs(
						nodeKind.SchemaExtensionID, nodeKind.DisplayName, nodeKind.Description,
						nodeKind.IsDisplayKind, nodeKind.Icon, nodeKind.Color, nodeKind.ID,
					).
					WillReturnRows(pool.NewRows(nodeKindColumns()).CloseError(databaseError)).
					RowsWillBeClosed()
			},
			wantErr: databaseError,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			store, pool := newTestStore(t)
			test.setupMock(pool)

			nodeKind, err := store.UpdateGraphSchemaNodeKind(
				test.args.ctx, test.args.nodeKind,
			)
			assert.Equal(t, test.wantNodeKind, nodeKind)
			assert.ErrorIs(t, err, test.wantErr)
			require.NoError(t, pool.ExpectationsWereMet())
		})
	}
}

const expectedDeleteGraphSchemaNodeKindSQL = `DELETE FROM schema_node_kinds WHERE id = $1`

func TestStore_DeleteGraphSchemaNodeKind(t *testing.T) {
	t.Parallel()

	type args struct {
		ctx        context.Context
		nodeKindID int32
	}

	type testCase struct {
		name      string
		args      args
		setupMock func(pgxmock.PgxPoolIface)
		wantErr   error
	}

	var (
		ctx           = context.Background()
		nodeKindID    = int32(42)
		databaseError = errors.New("db error")
	)

	tests := []testCase{
		{
			name: "success_-_deletes_node_kind",
			args: args{ctx: ctx, nodeKindID: nodeKindID},
			setupMock: func(pool pgxmock.PgxPoolIface) {
				pool.ExpectExec(expectedDeleteGraphSchemaNodeKindSQL).
					WithArgs(nodeKindID).
					WillReturnResult(pgxmock.NewResult("DELETE", 1))
			},
		},
		{
			name: "error_-_node_kind_not_found",
			args: args{ctx: ctx, nodeKindID: nodeKindID},
			setupMock: func(pool pgxmock.PgxPoolIface) {
				pool.ExpectExec(expectedDeleteGraphSchemaNodeKindSQL).
					WithArgs(nodeKindID).
					WillReturnResult(pgxmock.NewResult("DELETE", 0))
			},
			wantErr: services.ErrNodeKindNotFound,
		},
		{
			name: "error_-_propagates_database_error",
			args: args{ctx: ctx, nodeKindID: nodeKindID},
			setupMock: func(pool pgxmock.PgxPoolIface) {
				pool.ExpectExec(expectedDeleteGraphSchemaNodeKindSQL).
					WithArgs(nodeKindID).
					WillReturnError(databaseError)
			},
			wantErr: databaseError,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			store, pool := newTestStore(t)
			test.setupMock(pool)

			err := store.DeleteGraphSchemaNodeKind(test.args.ctx, test.args.nodeKindID)
			assert.ErrorIs(t, err, test.wantErr)
			require.NoError(t, pool.ExpectationsWereMet())
		})
	}
}
