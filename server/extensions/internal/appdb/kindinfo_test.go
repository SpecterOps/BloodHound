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
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/specterops/bloodhound/server/extensions/internal/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Literal SQL expected by the Store, compared via pgxmock.QueryMatcherEqual (whitespace
// normalised).
const expectedGetKindInfosSQL = `SELECT ki.id, ki.kind_id, ki.node_kind_id, ki.relationship_kind_id, k.name, ki.info_key, ki.title, ki.position, ki.content, ki.created_at, ki.updated_at FROM schema_kind_info AS ki JOIN kind AS k ON ki.kind_id = k.id WHERE k.name = $1 ORDER BY ki.position, ki.title`

func kindInfoColumns() []string {
	return []string{
		"id", "kind_id", "node_kind_id", "relationship_kind_id", "name",
		"info_key", "title", "position", "content", "created_at", "updated_at",
	}
}

func TestStore_GetKindInfos(t *testing.T) {
	t.Parallel()

	var (
		ctx        = context.Background()
		nodeKindID = int32(42)
		kindName   = "User"
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
			name: "returns ordered infos for a kind name",
			expectations: func(pool pgxmock.PgxPoolIface) {
				pool.ExpectQuery(expectedGetKindInfosSQL).WithArgs(kindName).WillReturnRows(
					pool.NewRows(kindInfoColumns()).
						AddRow(int32(1), int32(99), &nodeKindID, (*int32)(nil), kindName,
							"overview", "Overview", int32(0), content, createdAt, updatedAt).
						AddRow(int32(2), int32(99), &nodeKindID, (*int32)(nil), kindName,
							"details", "Details", int32(1), content, createdAt, updatedAt),
				)
			},
			wantKindInfos: []services.KindInfo{
				{ID: 1, KindID: 99, NodeKindID: &nodeKindID, Name: kindName, InfoKey: "overview",
					Title: "Overview", Position: 0, Content: content, CreatedAt: createdAt, UpdatedAt: updatedAt},
				{ID: 2, KindID: 99, NodeKindID: &nodeKindID, Name: kindName, InfoKey: "details",
					Title: "Details", Position: 1, Content: content, CreatedAt: createdAt, UpdatedAt: updatedAt},
			},
		},
		{
			name: "returns an empty slice when no infos exist",
			expectations: func(pool pgxmock.PgxPoolIface) {
				pool.ExpectQuery(expectedGetKindInfosSQL).WithArgs(kindName).WillReturnRows(
					pool.NewRows(kindInfoColumns()),
				)
			},
			wantKindInfos: []services.KindInfo{},
		},
		{
			name: "propagates database errors",
			expectations: func(pool pgxmock.PgxPoolIface) {
				pool.ExpectQuery(expectedGetKindInfosSQL).WithArgs(kindName).WillReturnError(dbErr)
			},
			wantKindInfos: nil,
			wantErr:       dbErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			store, pool := newTestStore(t)
			tt.expectations(pool)

			kindInfos, err := store.GetKindInfos(ctx, kindName)
			if assert.Equal(t, tt.wantKindInfos, kindInfos); tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
			}
			require.NoError(t, pool.ExpectationsWereMet())
		})
	}
}

const expectedCreateKindInfoSQL = `WITH inserted_kind_info AS (INSERT INTO schema_kind_info (kind_id, node_kind_id, relationship_kind_id, info_key, title, position, content, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7::jsonb, NOW(), NOW()) RETURNING id, kind_id, node_kind_id, relationship_kind_id, info_key, title, position, content, created_at, updated_at) SELECT ki.id, ki.kind_id, ki.node_kind_id, ki.relationship_kind_id, k.name, ki.info_key, ki.title, ki.position, ki.content, ki.created_at, ki.updated_at FROM inserted_kind_info AS ki JOIN kind AS k ON ki.kind_id = k.id`

func TestStore_CreateKindInfo(t *testing.T) {
	t.Parallel()

	type args struct {
		ctx                context.Context
		kindID             int32
		nodeKindID         *int32
		relationshipKindID *int32
		kindInfo           services.KindInfoInput
	}
	type want struct {
		kindInfo services.KindInfo
		err      error
	}
	type testCase struct {
		name      string
		args      args
		setupMock func(pgxmock.PgxPoolIface)
		want      want
	}

	var (
		ctx          = context.Background()
		nodeKindInfo = services.KindInfo{
			ID:         7,
			KindID:     99,
			NodeKindID: new(int32(42)),
			Name:       "User",
			InfoKey:    "overview",
			Title:      "Overview",
			Position:   2,
			Content:    json.RawMessage(`{"markdown":{"content":"Overview"}}`),
			CreatedAt:  time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC),
			UpdatedAt:  time.Date(2026, 2, 3, 4, 5, 6, 0, time.UTC),
		}
		relationshipKindInfo = nodeKindInfo
		defaultKindInfo      = nodeKindInfo
		kindInfoInput        = services.KindInfoInput{
			InfoKey:  nodeKindInfo.InfoKey,
			Title:    nodeKindInfo.Title,
			Position: nodeKindInfo.Position,
		}
		databaseError           = errors.New("database error")
		unmappedConstraintError = &pgconn.PgError{
			Code:           "23514",
			ConstraintName: "unknown_constraint",
		}
	)

	relationshipKindInfo.NodeKindID = nil
	relationshipKindInfo.RelationshipKindID = new(int32(43))
	defaultKindInfo.Content = json.RawMessage(`{"markdown":{"content":""}}`)

	tests := []testCase{
		{
			name: "success_-_creates_node_kind_info_with_explicit_content",
			args: args{
				ctx:        ctx,
				kindID:     nodeKindInfo.KindID,
				nodeKindID: nodeKindInfo.NodeKindID,
				kindInfo: services.KindInfoInput{
					InfoKey:  kindInfoInput.InfoKey,
					Title:    kindInfoInput.Title,
					Position: kindInfoInput.Position,
					Content:  nodeKindInfo.Content,
				},
			},
			setupMock: func(pool pgxmock.PgxPoolIface) {
				pool.ExpectQuery(expectedCreateKindInfoSQL).
					WithArgs(
						nodeKindInfo.KindID, nodeKindInfo.NodeKindID, nodeKindInfo.RelationshipKindID,
						nodeKindInfo.InfoKey, nodeKindInfo.Title, nodeKindInfo.Position, string(nodeKindInfo.Content),
					).
					WillReturnRows(pool.NewRows(kindInfoColumns()).AddRow(
						nodeKindInfo.ID, nodeKindInfo.KindID, nodeKindInfo.NodeKindID, nodeKindInfo.RelationshipKindID, nodeKindInfo.Name,
						nodeKindInfo.InfoKey, nodeKindInfo.Title, nodeKindInfo.Position, nodeKindInfo.Content,
						nodeKindInfo.CreatedAt, nodeKindInfo.UpdatedAt,
					)).RowsWillBeClosed()
			},
			want: want{kindInfo: nodeKindInfo},
		},
		{
			name: "success_-_creates_relationship_kind_info",
			args: args{
				ctx:                ctx,
				kindID:             nodeKindInfo.KindID,
				relationshipKindID: relationshipKindInfo.RelationshipKindID,
				kindInfo: services.KindInfoInput{
					InfoKey:  kindInfoInput.InfoKey,
					Title:    kindInfoInput.Title,
					Position: kindInfoInput.Position,
					Content:  nodeKindInfo.Content,
				},
			},
			setupMock: func(pool pgxmock.PgxPoolIface) {
				pool.ExpectQuery(expectedCreateKindInfoSQL).
					WithArgs(
						relationshipKindInfo.KindID, relationshipKindInfo.NodeKindID, relationshipKindInfo.RelationshipKindID,
						relationshipKindInfo.InfoKey, relationshipKindInfo.Title, relationshipKindInfo.Position, string(relationshipKindInfo.Content),
					).
					WillReturnRows(pool.NewRows(kindInfoColumns()).AddRow(
						relationshipKindInfo.ID, relationshipKindInfo.KindID, relationshipKindInfo.NodeKindID, relationshipKindInfo.RelationshipKindID, relationshipKindInfo.Name,
						relationshipKindInfo.InfoKey, relationshipKindInfo.Title, relationshipKindInfo.Position, relationshipKindInfo.Content,
						relationshipKindInfo.CreatedAt, relationshipKindInfo.UpdatedAt,
					)).RowsWillBeClosed()
			},
			want: want{kindInfo: relationshipKindInfo},
		},
		{
			name: "success_-_defaults_nil_content",
			args: args{
				ctx:        ctx,
				kindID:     nodeKindInfo.KindID,
				nodeKindID: nodeKindInfo.NodeKindID,
				kindInfo:   kindInfoInput,
			},
			setupMock: func(pool pgxmock.PgxPoolIface) {
				pool.ExpectQuery(expectedCreateKindInfoSQL).
					WithArgs(
						defaultKindInfo.KindID, defaultKindInfo.NodeKindID, defaultKindInfo.RelationshipKindID,
						defaultKindInfo.InfoKey, defaultKindInfo.Title, defaultKindInfo.Position, string(defaultKindInfo.Content),
					).
					WillReturnRows(pool.NewRows(kindInfoColumns()).AddRow(
						defaultKindInfo.ID, defaultKindInfo.KindID, defaultKindInfo.NodeKindID, defaultKindInfo.RelationshipKindID, defaultKindInfo.Name,
						defaultKindInfo.InfoKey, defaultKindInfo.Title, defaultKindInfo.Position, defaultKindInfo.Content,
						defaultKindInfo.CreatedAt, defaultKindInfo.UpdatedAt,
					)).RowsWillBeClosed()
			},
			want: want{kindInfo: defaultKindInfo},
		},
		{
			name: "success_-_defaults_empty_content",
			args: args{
				ctx:        ctx,
				kindID:     nodeKindInfo.KindID,
				nodeKindID: nodeKindInfo.NodeKindID,
				kindInfo: services.KindInfoInput{
					InfoKey:  kindInfoInput.InfoKey,
					Title:    kindInfoInput.Title,
					Position: kindInfoInput.Position,
					Content:  json.RawMessage{},
				},
			},
			setupMock: func(pool pgxmock.PgxPoolIface) {
				pool.ExpectQuery(expectedCreateKindInfoSQL).
					WithArgs(
						defaultKindInfo.KindID, defaultKindInfo.NodeKindID, defaultKindInfo.RelationshipKindID,
						defaultKindInfo.InfoKey, defaultKindInfo.Title, defaultKindInfo.Position, string(defaultKindInfo.Content),
					).
					WillReturnRows(pool.NewRows(kindInfoColumns()).AddRow(
						defaultKindInfo.ID, defaultKindInfo.KindID, defaultKindInfo.NodeKindID, defaultKindInfo.RelationshipKindID, defaultKindInfo.Name,
						defaultKindInfo.InfoKey, defaultKindInfo.Title, defaultKindInfo.Position, defaultKindInfo.Content,
						defaultKindInfo.CreatedAt, defaultKindInfo.UpdatedAt,
					)).RowsWillBeClosed()
			},
			want: want{kindInfo: defaultKindInfo},
		},
		{
			name: "error_-_preserves_query_errors",
			args: args{
				ctx:        ctx,
				kindID:     nodeKindInfo.KindID,
				nodeKindID: nodeKindInfo.NodeKindID,
				kindInfo:   kindInfoInput,
			},
			setupMock: func(pool pgxmock.PgxPoolIface) {
				pool.ExpectQuery(expectedCreateKindInfoSQL).
					WithArgs(
						defaultKindInfo.KindID, defaultKindInfo.NodeKindID, defaultKindInfo.RelationshipKindID,
						defaultKindInfo.InfoKey, defaultKindInfo.Title, defaultKindInfo.Position, string(defaultKindInfo.Content),
					).
					WillReturnError(databaseError)
			},
			want: want{err: databaseError},
		},
		{
			name: "error_-_preserves_row_errors",
			args: args{
				ctx:        ctx,
				kindID:     nodeKindInfo.KindID,
				nodeKindID: nodeKindInfo.NodeKindID,
				kindInfo:   kindInfoInput,
			},
			setupMock: func(pool pgxmock.PgxPoolIface) {
				pool.ExpectQuery(expectedCreateKindInfoSQL).
					WithArgs(
						defaultKindInfo.KindID, defaultKindInfo.NodeKindID, defaultKindInfo.RelationshipKindID,
						defaultKindInfo.InfoKey, defaultKindInfo.Title, defaultKindInfo.Position, string(defaultKindInfo.Content),
					).
					WillReturnRows(pool.NewRows(kindInfoColumns()).AddRow(
						defaultKindInfo.ID, defaultKindInfo.KindID, defaultKindInfo.NodeKindID, defaultKindInfo.RelationshipKindID, defaultKindInfo.Name,
						defaultKindInfo.InfoKey, defaultKindInfo.Title, defaultKindInfo.Position, defaultKindInfo.Content,
						defaultKindInfo.CreatedAt, defaultKindInfo.UpdatedAt,
					).CloseError(databaseError)).RowsWillBeClosed()
			},
			want: want{err: databaseError},
		},
		{
			name: "error_-_maps_missing_kind_from_rows",
			args: args{
				ctx:        ctx,
				kindID:     nodeKindInfo.KindID,
				nodeKindID: nodeKindInfo.NodeKindID,
				kindInfo:   kindInfoInput,
			},
			setupMock: func(pool pgxmock.PgxPoolIface) {
				pool.ExpectQuery(expectedCreateKindInfoSQL).
					WithArgs(
						defaultKindInfo.KindID, defaultKindInfo.NodeKindID, defaultKindInfo.RelationshipKindID,
						defaultKindInfo.InfoKey, defaultKindInfo.Title, defaultKindInfo.Position, string(defaultKindInfo.Content),
					).
					WillReturnRows(pool.NewRows(kindInfoColumns()).CloseError(
						fmt.Errorf("driver: %w", &pgconn.PgError{
							Code:           "23503",
							ConstraintName: "schema_kind_info_kind_id_fkey",
						}),
					)).RowsWillBeClosed()
			},
			want: want{err: services.ErrKindInfoKindNotFound},
		},
		{
			name: "error_-_maps_duplicate_info_key_from_rows",
			args: args{
				ctx:        ctx,
				kindID:     nodeKindInfo.KindID,
				nodeKindID: nodeKindInfo.NodeKindID,
				kindInfo:   kindInfoInput,
			},
			setupMock: func(pool pgxmock.PgxPoolIface) {
				pool.ExpectQuery(expectedCreateKindInfoSQL).
					WithArgs(
						defaultKindInfo.KindID, defaultKindInfo.NodeKindID, defaultKindInfo.RelationshipKindID,
						defaultKindInfo.InfoKey, defaultKindInfo.Title, defaultKindInfo.Position, string(defaultKindInfo.Content),
					).
					WillReturnRows(pool.NewRows(kindInfoColumns()).CloseError(&pgconn.PgError{
						Code:           "23505",
						ConstraintName: "schema_kind_info_unique_kind_info_key",
					})).RowsWillBeClosed()
			},
			want: want{err: services.ErrKindInfoDuplicateInfoKey},
		},
		{
			name: "error_-_maps_duplicate_position",
			args: args{
				ctx:        ctx,
				kindID:     nodeKindInfo.KindID,
				nodeKindID: nodeKindInfo.NodeKindID,
				kindInfo:   kindInfoInput,
			},
			setupMock: func(pool pgxmock.PgxPoolIface) {
				pool.ExpectQuery(expectedCreateKindInfoSQL).
					WithArgs(
						defaultKindInfo.KindID, defaultKindInfo.NodeKindID, defaultKindInfo.RelationshipKindID,
						defaultKindInfo.InfoKey, defaultKindInfo.Title, defaultKindInfo.Position, string(defaultKindInfo.Content),
					).
					WillReturnError(&pgconn.PgError{
						Code:           "23505",
						ConstraintName: "schema_kind_info_unique_kind_position",
					})
			},
			want: want{err: services.ErrKindInfoDuplicatePosition},
		},
		{
			name: "error_-_preserves_unmapped_constraint",
			args: args{
				ctx:        ctx,
				kindID:     nodeKindInfo.KindID,
				nodeKindID: nodeKindInfo.NodeKindID,
				kindInfo:   kindInfoInput,
			},
			setupMock: func(pool pgxmock.PgxPoolIface) {
				pool.ExpectQuery(expectedCreateKindInfoSQL).
					WithArgs(
						defaultKindInfo.KindID, defaultKindInfo.NodeKindID, defaultKindInfo.RelationshipKindID,
						defaultKindInfo.InfoKey, defaultKindInfo.Title, defaultKindInfo.Position, string(defaultKindInfo.Content),
					).
					WillReturnError(unmappedConstraintError)
			},
			want: want{err: unmappedConstraintError},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			store, pool := newTestStore(t)
			test.setupMock(pool)

			kindInfo, err := store.CreateKindInfo(
				test.args.ctx, test.args.kindID, test.args.nodeKindID, test.args.relationshipKindID, test.args.kindInfo,
			)

			assert.Equal(t, test.want.kindInfo, kindInfo)
			assert.ErrorIs(t, err, test.want.err)
			require.NoError(t, pool.ExpectationsWereMet())
		})
	}
}

const expectedUpdateKindInfoSQL = `WITH updated_kind_info AS (UPDATE schema_kind_info SET kind_id = $1, node_kind_id = $2, relationship_kind_id = $3, info_key = $4, title = $5, position = $6, content = $7::jsonb, updated_at = NOW() WHERE id = $8 RETURNING id, kind_id, node_kind_id, relationship_kind_id, info_key, title, position, content, created_at, updated_at) SELECT ki.id, ki.kind_id, ki.node_kind_id, ki.relationship_kind_id, k.name, ki.info_key, ki.title, ki.position, ki.content, ki.created_at, ki.updated_at FROM updated_kind_info AS ki JOIN kind AS k ON ki.kind_id = k.id`

func TestStore_UpdateKindInfo(t *testing.T) {
	t.Parallel()

	type args struct {
		ctx      context.Context
		kindInfo services.KindInfo
	}
	type want struct {
		kindInfo services.KindInfo
		err      error
	}
	type testCase struct {
		name      string
		args      args
		setupMock func(pgxmock.PgxPoolIface)
		want      want
	}

	var (
		ctx      = context.Background()
		kindInfo = services.KindInfo{
			ID:         7,
			KindID:     99,
			NodeKindID: new(int32(42)),
			Name:       "User",
			InfoKey:    "overview",
			Title:      "Updated",
			Position:   2,
			Content:    json.RawMessage(`{"markdown":{"content":"Updated overview"}}`),
			CreatedAt:  time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC),
			UpdatedAt:  time.Date(2026, 2, 3, 4, 5, 6, 0, time.UTC),
		}
		relationshipKindInfo = kindInfo
		defaultKindInfo      = kindInfo
		databaseError        = errors.New("database error")
	)

	relationshipKindInfo.KindID = 100
	relationshipKindInfo.NodeKindID = nil
	relationshipKindInfo.RelationshipKindID = new(int32(43))
	relationshipKindInfo.Name = "MemberOf"
	defaultKindInfo.Content = json.RawMessage(`{"markdown":{"content":""}}`)

	tests := []testCase{
		{
			name: "success_-_updates_node_kind_info_fields",
			args: args{ctx: ctx, kindInfo: kindInfo},
			setupMock: func(pool pgxmock.PgxPoolIface) {
				pool.ExpectQuery(expectedUpdateKindInfoSQL).
					WithArgs(
						kindInfo.KindID, kindInfo.NodeKindID, kindInfo.RelationshipKindID,
						kindInfo.InfoKey, kindInfo.Title, kindInfo.Position, string(kindInfo.Content), kindInfo.ID,
					).
					WillReturnRows(pool.NewRows(kindInfoColumns()).AddRow(
						kindInfo.ID, kindInfo.KindID, kindInfo.NodeKindID, kindInfo.RelationshipKindID, kindInfo.Name,
						kindInfo.InfoKey, kindInfo.Title, kindInfo.Position, kindInfo.Content, kindInfo.CreatedAt, kindInfo.UpdatedAt,
					)).RowsWillBeClosed()
			},
			want: want{kindInfo: kindInfo},
		},
		{
			name: "success_-_updates_relationship_kind_info_fields",
			args: args{ctx: ctx, kindInfo: relationshipKindInfo},
			setupMock: func(pool pgxmock.PgxPoolIface) {
				pool.ExpectQuery(expectedUpdateKindInfoSQL).
					WithArgs(
						relationshipKindInfo.KindID, relationshipKindInfo.NodeKindID, relationshipKindInfo.RelationshipKindID,
						relationshipKindInfo.InfoKey, relationshipKindInfo.Title, relationshipKindInfo.Position,
						string(relationshipKindInfo.Content), relationshipKindInfo.ID,
					).
					WillReturnRows(pool.NewRows(kindInfoColumns()).AddRow(
						relationshipKindInfo.ID, relationshipKindInfo.KindID, relationshipKindInfo.NodeKindID,
						relationshipKindInfo.RelationshipKindID, relationshipKindInfo.Name,
						relationshipKindInfo.InfoKey, relationshipKindInfo.Title, relationshipKindInfo.Position,
						relationshipKindInfo.Content, relationshipKindInfo.CreatedAt, relationshipKindInfo.UpdatedAt,
					)).RowsWillBeClosed()
			},
			want: want{kindInfo: relationshipKindInfo},
		},
		{
			name: "success_-_defaults_nil_content",
			args: args{
				ctx: ctx,
				kindInfo: services.KindInfo{
					ID: kindInfo.ID, KindID: kindInfo.KindID, NodeKindID: kindInfo.NodeKindID,
					InfoKey: kindInfo.InfoKey, Title: kindInfo.Title, Position: kindInfo.Position,
				},
			},
			setupMock: func(pool pgxmock.PgxPoolIface) {
				pool.ExpectQuery(expectedUpdateKindInfoSQL).
					WithArgs(
						defaultKindInfo.KindID, defaultKindInfo.NodeKindID, defaultKindInfo.RelationshipKindID,
						defaultKindInfo.InfoKey, defaultKindInfo.Title, defaultKindInfo.Position,
						string(defaultKindInfo.Content), defaultKindInfo.ID,
					).
					WillReturnRows(pool.NewRows(kindInfoColumns()).AddRow(
						defaultKindInfo.ID, defaultKindInfo.KindID, defaultKindInfo.NodeKindID,
						defaultKindInfo.RelationshipKindID, defaultKindInfo.Name, defaultKindInfo.InfoKey,
						defaultKindInfo.Title, defaultKindInfo.Position, defaultKindInfo.Content,
						defaultKindInfo.CreatedAt, defaultKindInfo.UpdatedAt,
					)).RowsWillBeClosed()
			},
			want: want{kindInfo: defaultKindInfo},
		},
		{
			name: "success_-_defaults_empty_content",
			args: args{
				ctx: ctx,
				kindInfo: services.KindInfo{
					ID: kindInfo.ID, KindID: kindInfo.KindID, NodeKindID: kindInfo.NodeKindID,
					InfoKey: kindInfo.InfoKey, Title: kindInfo.Title, Position: kindInfo.Position,
					Content: json.RawMessage{},
				},
			},
			setupMock: func(pool pgxmock.PgxPoolIface) {
				pool.ExpectQuery(expectedUpdateKindInfoSQL).
					WithArgs(
						defaultKindInfo.KindID, defaultKindInfo.NodeKindID, defaultKindInfo.RelationshipKindID,
						defaultKindInfo.InfoKey, defaultKindInfo.Title, defaultKindInfo.Position,
						string(defaultKindInfo.Content), defaultKindInfo.ID,
					).
					WillReturnRows(pool.NewRows(kindInfoColumns()).AddRow(
						defaultKindInfo.ID, defaultKindInfo.KindID, defaultKindInfo.NodeKindID,
						defaultKindInfo.RelationshipKindID, defaultKindInfo.Name, defaultKindInfo.InfoKey,
						defaultKindInfo.Title, defaultKindInfo.Position, defaultKindInfo.Content,
						defaultKindInfo.CreatedAt, defaultKindInfo.UpdatedAt,
					)).RowsWillBeClosed()
			},
			want: want{kindInfo: defaultKindInfo},
		},
		{
			name: "error_-_returns_not_found_when_no_kind_info_matches",
			args: args{ctx: ctx, kindInfo: kindInfo},
			setupMock: func(pool pgxmock.PgxPoolIface) {
				pool.ExpectQuery(expectedUpdateKindInfoSQL).
					WithArgs(
						kindInfo.KindID, kindInfo.NodeKindID, kindInfo.RelationshipKindID,
						kindInfo.InfoKey, kindInfo.Title, kindInfo.Position, string(kindInfo.Content), kindInfo.ID,
					).
					WillReturnRows(pool.NewRows(kindInfoColumns())).RowsWillBeClosed()
			},
			want: want{err: services.ErrKindInfoNotFound},
		},
		{
			name: "error_-_preserves_database_errors",
			args: args{ctx: ctx, kindInfo: kindInfo},
			setupMock: func(pool pgxmock.PgxPoolIface) {
				pool.ExpectQuery(expectedUpdateKindInfoSQL).
					WithArgs(
						kindInfo.KindID, kindInfo.NodeKindID, kindInfo.RelationshipKindID,
						kindInfo.InfoKey, kindInfo.Title, kindInfo.Position, string(kindInfo.Content), kindInfo.ID,
					).
					WillReturnError(databaseError)
			},
			want: want{err: databaseError},
		},
		{
			name: "error_-_maps_wrapped_missing_kind_from_rows",
			args: args{ctx: ctx, kindInfo: kindInfo},
			setupMock: func(pool pgxmock.PgxPoolIface) {
				pool.ExpectQuery(expectedUpdateKindInfoSQL).
					WithArgs(
						kindInfo.KindID, kindInfo.NodeKindID, kindInfo.RelationshipKindID,
						kindInfo.InfoKey, kindInfo.Title, kindInfo.Position, string(kindInfo.Content), kindInfo.ID,
					).
					WillReturnRows(pool.NewRows(kindInfoColumns()).CloseError(
						fmt.Errorf("driver: %w", &pgconn.PgError{
							Code:           "23503",
							ConstraintName: "schema_kind_info_kind_id_fkey",
						}),
					)).RowsWillBeClosed()
			},
			want: want{err: services.ErrKindInfoKindNotFound},
		},
		{
			name: "error_-_maps_duplicate_info_key_from_rows",
			args: args{ctx: ctx, kindInfo: kindInfo},
			setupMock: func(pool pgxmock.PgxPoolIface) {
				pool.ExpectQuery(expectedUpdateKindInfoSQL).
					WithArgs(
						kindInfo.KindID, kindInfo.NodeKindID, kindInfo.RelationshipKindID,
						kindInfo.InfoKey, kindInfo.Title, kindInfo.Position, string(kindInfo.Content), kindInfo.ID,
					).
					WillReturnRows(pool.NewRows(kindInfoColumns()).CloseError(&pgconn.PgError{
						Code:           "23505",
						ConstraintName: "schema_kind_info_unique_kind_info_key",
					})).RowsWillBeClosed()
			},
			want: want{err: services.ErrKindInfoDuplicateInfoKey},
		},
		{
			name: "error_-_maps_duplicate_position",
			args: args{ctx: ctx, kindInfo: kindInfo},
			setupMock: func(pool pgxmock.PgxPoolIface) {
				pool.ExpectQuery(expectedUpdateKindInfoSQL).
					WithArgs(
						kindInfo.KindID, kindInfo.NodeKindID, kindInfo.RelationshipKindID,
						kindInfo.InfoKey, kindInfo.Title, kindInfo.Position, string(kindInfo.Content), kindInfo.ID,
					).
					WillReturnError(&pgconn.PgError{
						Code:           "23505",
						ConstraintName: "schema_kind_info_unique_kind_position",
					})
			},
			want: want{err: services.ErrKindInfoDuplicatePosition},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			store, pool := newTestStore(t)
			test.setupMock(pool)

			updatedKindInfo, err := store.UpdateKindInfo(test.args.ctx, test.args.kindInfo)

			assert.Equal(t, test.want.kindInfo, updatedKindInfo)
			assert.ErrorIs(t, err, test.want.err)
			require.NoError(t, pool.ExpectationsWereMet())
		})
	}
}

const expectedDeleteKindInfoSQL = `DELETE FROM schema_kind_info WHERE id = $1`

func TestStore_DeleteKindInfo(t *testing.T) {
	t.Parallel()

	type args struct {
		ctx        context.Context
		kindInfoID int32
	}
	type want struct {
		err error
	}

	var (
		ctx           = context.Background()
		kindInfoID    = int32(7)
		databaseError = errors.New("database error")
		tests         = []struct {
			name      string
			args      args
			setupMock func(pgxmock.PgxPoolIface)
			want      want
		}{
			{
				name: "success_-_deletes_kind_info",
				args: args{ctx: ctx, kindInfoID: kindInfoID},
				setupMock: func(pool pgxmock.PgxPoolIface) {
					pool.ExpectExec(expectedDeleteKindInfoSQL).
						WithArgs(kindInfoID).
						WillReturnResult(pgxmock.NewResult("DELETE", 1))
				},
			},
			{
				name: "error_-_kind_info_not_found",
				args: args{ctx: ctx, kindInfoID: kindInfoID},
				setupMock: func(pool pgxmock.PgxPoolIface) {
					pool.ExpectExec(expectedDeleteKindInfoSQL).
						WithArgs(kindInfoID).
						WillReturnResult(pgxmock.NewResult("DELETE", 0))
				},
				want: want{err: services.ErrKindInfoNotFound},
			},
			{
				name: "error_-_propagates_database_error",
				args: args{ctx: ctx, kindInfoID: kindInfoID},
				setupMock: func(pool pgxmock.PgxPoolIface) {
					pool.ExpectExec(expectedDeleteKindInfoSQL).
						WithArgs(kindInfoID).
						WillReturnError(databaseError)
				},
				want: want{err: databaseError},
			},
		}
	)

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			store, pool := newTestStore(t)
			test.setupMock(pool)

			err := store.DeleteKindInfo(test.args.ctx, test.args.kindInfoID)

			assert.ErrorIs(t, err, test.want.err)
			require.NoError(t, pool.ExpectationsWereMet())
		})
	}
}
