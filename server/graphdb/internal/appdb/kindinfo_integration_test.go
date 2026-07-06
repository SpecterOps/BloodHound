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

//go:build integration

package appdb_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/specterops/bloodhound/server/graphdb/internal/appdb"
	"github.com/specterops/bloodhound/server/graphdb/internal/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	pgErrCodeForeignKeyViolation = "23503"
	pgErrCodeUniqueViolation     = "23505"
	pgErrCodeCheckViolation      = "23514"
)

type kindInfoSeed struct {
	nodeBackingKindID             int32
	nodeSchemaKindID              int32
	relationshipBackingKindID     int32
	relationshipSchemaKindID      int32
	missingBackingKindID          int32
	missingNodeSchemaKindID       int32
	missingRelationshipSchemaKind int32
}

func TestCreateKindInfos(t *testing.T) {
	var (
		ctx         = context.Background()
		store, pool = setupKindInfoStore(t)
		seed        = seedKindInfoPrerequisites(t, ctx, pool)
	)

	kindInfos := []services.KindInfo{
		{
			KindID:     seed.nodeBackingKindID,
			NodeKindID: &seed.nodeSchemaKindID,
			InfoKey:    "general",
			Title:      "General",
			Position:   0,
			Content:    json.RawMessage(`{"markdown":{"content":"hello"}}`),
		},
		{
			KindID:             seed.relationshipBackingKindID,
			RelationshipKindID: &seed.relationshipSchemaKindID,
			InfoKey:            "abuse",
			Title:              "Abuse",
			Position:           0,
			Content:            json.RawMessage(`{"query":{"cypher":"MATCH (n) RETURN n"}}`),
		},
	}

	require.NoError(t, store.CreateKindInfos(ctx, kindInfos))

	var rowCount int
	require.NoError(t, pool.QueryRow(ctx, `SELECT count(*) FROM schema_kind_info`).Scan(&rowCount))
	assert.Equal(t, 2, rowCount)

	var (
		title    string
		content  []byte
		position int32
	)
	require.NoError(t, pool.QueryRow(ctx, `
		SELECT title, content, position
		FROM schema_kind_info
		WHERE kind_id = $1 AND info_key = $2
	`, seed.nodeBackingKindID, "general").Scan(&title, &content, &position))
	assert.Equal(t, "General", title)
	assert.JSONEq(t, `{"markdown":{"content":"hello"}}`, string(content))
	assert.Equal(t, int32(0), position)
}

// these test cases exercise db-level constraints
func TestCreateKindInfos_Constraints(t *testing.T) {
	var (
		ctx         = context.Background()
		store, pool = setupKindInfoStore(t)
		seed        = seedKindInfoPrerequisites(t, ctx, pool)
	)

	tests := []struct {
		name           string
		setup          func(t *testing.T)
		kindInfo       services.KindInfo
		wantCode       string
		wantConstraint string
		wantSentinel   error
	}{
		{
			name: "constraint - requires exactly one origin",
			kindInfo: services.KindInfo{
				KindID:   seed.nodeBackingKindID,
				InfoKey:  "general",
				Title:    "General",
				Position: 0,
				Content:  json.RawMessage(`{"markdown":{"content":"hello"}}`),
			},
			wantCode:       pgErrCodeCheckViolation,
			wantConstraint: "schema_kind_info_kind_origin",
		},
		{
			name: "constraint - rejects multiple origins",
			kindInfo: services.KindInfo{
				KindID:             seed.nodeBackingKindID,
				NodeKindID:         &seed.nodeSchemaKindID,
				RelationshipKindID: &seed.relationshipSchemaKindID,
				InfoKey:            "general",
				Title:              "General",
				Position:           0,
				Content:            json.RawMessage(`{"markdown":{"content":"hello"}}`),
			},
			wantCode:       pgErrCodeCheckViolation,
			wantConstraint: "schema_kind_info_kind_origin",
		},
		{
			name: "constraint - rejects blank info key",
			kindInfo: services.KindInfo{
				KindID:     seed.nodeBackingKindID,
				NodeKindID: &seed.nodeSchemaKindID,
				InfoKey:    "   ",
				Title:      "General",
				Position:   0,
				Content:    json.RawMessage(`{"markdown":{"content":"hello"}}`),
			},
			wantCode:       pgErrCodeCheckViolation,
			wantConstraint: "schema_kind_info_info_key_not_empty",
		},
		{
			name: "constraint - rejects blank title",
			kindInfo: services.KindInfo{
				KindID:     seed.nodeBackingKindID,
				NodeKindID: &seed.nodeSchemaKindID,
				InfoKey:    "general",
				Title:      "   ",
				Position:   0,
				Content:    json.RawMessage(`{"markdown":{"content":"hello"}}`),
			},
			wantCode:       pgErrCodeCheckViolation,
			wantConstraint: "schema_kind_info_title_not_empty",
		},
		{
			name: "constraint - rejects negative position",
			kindInfo: services.KindInfo{
				KindID:     seed.nodeBackingKindID,
				NodeKindID: &seed.nodeSchemaKindID,
				InfoKey:    "general",
				Title:      "General",
				Position:   -1,
				Content:    json.RawMessage(`{"markdown":{"content":"hello"}}`),
			},
			wantCode:       pgErrCodeCheckViolation,
			wantConstraint: "schema_kind_info_position_nonnegative_nonzero",
		},
		{
			name: "constraint - rejects non object content",
			kindInfo: services.KindInfo{
				KindID:     seed.nodeBackingKindID,
				NodeKindID: &seed.nodeSchemaKindID,
				InfoKey:    "general",
				Title:      "General",
				Position:   0,
				Content:    json.RawMessage(`[]`),
			},
			wantCode:       pgErrCodeCheckViolation,
			wantConstraint: "schema_kind_info_content_is_object",
		},
		{
			name: "constraint - enforces unique position per kind",
			setup: func(t *testing.T) {
				t.Helper()
				insertValidKindInfo(t, ctx, store, services.KindInfo{
					KindID:     seed.nodeBackingKindID,
					NodeKindID: &seed.nodeSchemaKindID,
					InfoKey:    "general",
					Title:      "General",
					Position:   0,
					Content:    json.RawMessage(`{"markdown":{"content":"hello"}}`),
				})
			},
			kindInfo: services.KindInfo{
				KindID:     seed.nodeBackingKindID,
				NodeKindID: &seed.nodeSchemaKindID,
				InfoKey:    "details",
				Title:      "Details",
				Position:   0,
				Content:    json.RawMessage(`{"markdown":{"content":"hello again"}}`),
			},
			wantCode:       pgErrCodeUniqueViolation,
			wantConstraint: "schema_kind_info_unique_kind_position",
			wantSentinel:   services.ErrKindInfoDuplicatePosition,
		},
		{
			name: "constraint - enforces unique info key per kind",
			setup: func(t *testing.T) {
				t.Helper()
				insertValidKindInfo(t, ctx, store, services.KindInfo{
					KindID:     seed.nodeBackingKindID,
					NodeKindID: &seed.nodeSchemaKindID,
					InfoKey:    "general",
					Title:      "General",
					Position:   0,
					Content:    json.RawMessage(`{"markdown":{"content":"hello"}}`),
				})
			},
			kindInfo: services.KindInfo{
				KindID:     seed.nodeBackingKindID,
				NodeKindID: &seed.nodeSchemaKindID,
				InfoKey:    "general",
				Title:      "General Again",
				Position:   1,
				Content:    json.RawMessage(`{"markdown":{"content":"hello again"}}`),
			},
			wantCode:       pgErrCodeUniqueViolation,
			wantConstraint: "schema_kind_info_unique_kind_info_key",
			wantSentinel:   services.ErrKindInfoDuplicateInfoKey,
		},
		{
			name: "constraint - requires existing kind",
			kindInfo: services.KindInfo{
				KindID:     seed.missingBackingKindID,
				NodeKindID: &seed.nodeSchemaKindID,
				InfoKey:    "general",
				Title:      "General",
				Position:   0,
				Content:    json.RawMessage(`{"markdown":{"content":"hello"}}`),
			},
			wantCode:       pgErrCodeForeignKeyViolation,
			wantConstraint: "schema_kind_info_kind_id_fkey",
			wantSentinel:   services.ErrKindInfoKindNotFound,
		},
		{
			name: "constraint - requires existing node kind origin",
			kindInfo: services.KindInfo{
				KindID:     seed.nodeBackingKindID,
				NodeKindID: &seed.missingNodeSchemaKindID,
				InfoKey:    "general",
				Title:      "General",
				Position:   0,
				Content:    json.RawMessage(`{"markdown":{"content":"hello"}}`),
			},
			wantCode:       pgErrCodeForeignKeyViolation,
			wantConstraint: "schema_kind_info_node_kind_id_fkey",
		},
		{
			name: "constraint - requires existing relationship kind origin",
			kindInfo: services.KindInfo{
				KindID:             seed.relationshipBackingKindID,
				RelationshipKindID: &seed.missingRelationshipSchemaKind,
				InfoKey:            "general",
				Title:              "General",
				Position:           0,
				Content:            json.RawMessage(`{"markdown":{"content":"hello"}}`),
			},
			wantCode:       pgErrCodeForeignKeyViolation,
			wantConstraint: "schema_kind_info_relationship_kind_id_fkey",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			truncateKindInfo(t, ctx, pool)
			if testCase.setup != nil {
				testCase.setup(t)
			}

			err := store.CreateKindInfos(ctx, []services.KindInfo{testCase.kindInfo})

			assertPgConstraintError(t, err, testCase.wantCode, testCase.wantConstraint)
			if testCase.wantSentinel != nil {
				assert.ErrorIs(t, err, testCase.wantSentinel)
			}
		})
	}
}

func TestGetKindInfos_NoneExist(t *testing.T) {
	var (
		ctx         = context.Background()
		store, pool = setupKindInfoStore(t)
		seed        = seedKindInfoPrerequisites(t, ctx, pool)
	)

	kindInfos, err := store.GetKindInfos(ctx, seed.nodeBackingKindID)

	require.NoError(t, err)
	assert.Empty(t, kindInfos)
}

func TestGetKindInfos_HappyPath(t *testing.T) {
	var (
		ctx         = context.Background()
		store, pool = setupKindInfoStore(t)
		seed        = seedKindInfoPrerequisites(t, ctx, pool)
	)

	// belongs to a different kind and must not appear in the results
	insertValidKindInfo(t, ctx, store, services.KindInfo{
		KindID:             seed.relationshipBackingKindID,
		RelationshipKindID: &seed.relationshipSchemaKindID,
		InfoKey:            "other",
		Title:              "Other",
		Position:           0,
		Content:            json.RawMessage(`{"query":{"cypher":"MATCH (n) RETURN n"}}`),
	})

	insertValidKindInfo(t, ctx, store, services.KindInfo{
		KindID:     seed.nodeBackingKindID,
		NodeKindID: &seed.nodeSchemaKindID,
		InfoKey:    "general",
		Title:      "General",
		Position:   0,
		Content:    json.RawMessage(`{"markdown":{"content":"hello"}}`),
	})

	kindInfos, err := store.GetKindInfos(ctx, seed.nodeBackingKindID)

	require.NoError(t, err)
	require.Len(t, kindInfos, 1)

	assert.Equal(t, seed.nodeBackingKindID, kindInfos[0].KindID)
	assert.Equal(t, &seed.nodeSchemaKindID, kindInfos[0].NodeKindID)
	assert.Nil(t, kindInfos[0].RelationshipKindID)
	assert.Equal(t, "general", kindInfos[0].InfoKey)
	assert.Equal(t, "General", kindInfos[0].Title)
	assert.Equal(t, int32(0), kindInfos[0].Position)
	assert.JSONEq(t, `{"markdown":{"content":"hello"}}`, string(kindInfos[0].Content))
}

func setupKindInfoStore(t *testing.T) (*appdb.Store, *pgxpool.Pool) {
	t.Helper()

	_, pool, graphDB := setupStoreWithGraph(t)
	return appdb.NewStore(graphDB, pool), pool
}

func seedKindInfoPrerequisites(t *testing.T, ctx context.Context, pool *pgxpool.Pool) kindInfoSeed {
	t.Helper()

	var (
		extensionID               int32
		nodeBackingKindID         int32
		nodeSchemaKindID          int32
		relationshipBackingKindID int32
		relationshipSchemaKindID  int32
		maxBackingKindID          int32
		maxNodeSchemaKindID       int32
		maxRelationshipSchemaKind int32
	)

	require.NoError(t, pool.QueryRow(ctx, `
		INSERT INTO schema_extensions (name, display_name, version, is_builtin, namespace)
		VALUES ('KindInfoTestExtension', 'Kind Info Test Extension', '1.0.0', false, 'KIT')
		RETURNING id
	`).Scan(&extensionID))

	require.NoError(t, pool.QueryRow(ctx, `
		INSERT INTO kind (name)
		VALUES ('KindInfoTestNode')
		RETURNING id
	`).Scan(&nodeBackingKindID))

	require.NoError(t, pool.QueryRow(ctx, `
		INSERT INTO schema_node_kinds (schema_extension_id, kind_id, display_name, description, is_display_kind, icon, icon_color)
		VALUES ($1, $2, 'Kind Info Test Node', 'Test node kind', true, 'cube', '#336699')
		RETURNING id
	`, extensionID, nodeBackingKindID).Scan(&nodeSchemaKindID))

	require.NoError(t, pool.QueryRow(ctx, `
		INSERT INTO kind (name)
		VALUES ('KindInfoTestRelationship')
		RETURNING id
	`).Scan(&relationshipBackingKindID))

	require.NoError(t, pool.QueryRow(ctx, `
		INSERT INTO schema_relationship_kinds (schema_extension_id, kind_id, description, is_traversable)
		VALUES ($1, $2, 'Test relationship kind', true)
		RETURNING id
	`, extensionID, relationshipBackingKindID).Scan(&relationshipSchemaKindID))

	require.NoError(t, pool.QueryRow(ctx, `SELECT max(id) FROM kind`).Scan(&maxBackingKindID))
	require.NoError(t, pool.QueryRow(ctx, `SELECT max(id) FROM schema_node_kinds`).Scan(&maxNodeSchemaKindID))
	require.NoError(t, pool.QueryRow(ctx, `SELECT max(id) FROM schema_relationship_kinds`).Scan(&maxRelationshipSchemaKind))

	return kindInfoSeed{
		nodeBackingKindID:             nodeBackingKindID,
		nodeSchemaKindID:              nodeSchemaKindID,
		relationshipBackingKindID:     relationshipBackingKindID,
		relationshipSchemaKindID:      relationshipSchemaKindID,
		missingBackingKindID:          maxBackingKindID + 1,
		missingNodeSchemaKindID:       maxNodeSchemaKindID + 1,
		missingRelationshipSchemaKind: maxRelationshipSchemaKind + 1,
	}
}

func insertValidKindInfo(t *testing.T, ctx context.Context, store *appdb.Store, kindInfo services.KindInfo) {
	t.Helper()

	require.NoError(t, store.CreateKindInfos(ctx, []services.KindInfo{kindInfo}))
}

func truncateKindInfo(t *testing.T, ctx context.Context, pool *pgxpool.Pool) {
	t.Helper()

	_, err := pool.Exec(ctx, `TRUNCATE TABLE schema_kind_info`)
	require.NoError(t, err)
}

func assertPgConstraintError(t *testing.T, err error, expectedCode, expectedConstraint string) {
	t.Helper()

	var pgErr *pgconn.PgError
	require.Error(t, err)
	require.True(t, errors.As(err, &pgErr), "expected wrapped *pgconn.PgError, got %T: %v", err, err)
	assert.Equal(t, expectedCode, pgErr.Code)
	assert.Equal(t, expectedConstraint, pgErr.ConstraintName)
}
