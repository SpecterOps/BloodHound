// Copyright 2026 Specter Ops, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
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
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/specterops/bloodhound/server/extensions/internal/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type relationshipKindTestData struct {
	extensionID        int32
	kindID             int32
	relationshipKindID int32
	description        string
	isTraversable      bool
	createdAt          time.Time
	updatedAt          time.Time
}

func seedRelationshipKind(t *testing.T, ctx context.Context, pool *pgxpool.Pool) relationshipKindTestData {
	t.Helper()

	var extensionID int32
	require.NoError(t, pool.QueryRow(ctx, `
		INSERT INTO schema_extensions (name, display_name, version, is_builtin, namespace)
		VALUES ('TestRelationshipExtension', 'Test Relationship Extension', '1.0.0', false, 'TRK')
		RETURNING id`).Scan(&extensionID))

	var kindID int32
	require.NoError(t, pool.QueryRow(ctx, `
		INSERT INTO kind (name) VALUES ('TestRelationshipKind') RETURNING id`).Scan(&kindID))

	var relationshipKindID int32
	const (
		description   = "a test relationship kind"
		isTraversable = true
	)

	var createdAt, updatedAt time.Time
	require.NoError(t, pool.QueryRow(ctx, `
		INSERT INTO schema_relationship_kinds (schema_extension_id, kind_id, description, is_traversable)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at`, extensionID, kindID, description, isTraversable).
		Scan(&relationshipKindID, &createdAt, &updatedAt))

	t.Cleanup(func() {
		_, _ = pool.Exec(ctx, `DELETE FROM schema_extensions WHERE name = 'TestRelationshipExtension'`)
		_, _ = pool.Exec(ctx, `DELETE FROM kind WHERE name = 'TestRelationshipKind'`)
	})

	return relationshipKindTestData{
		extensionID:        extensionID,
		kindID:             kindID,
		relationshipKindID: relationshipKindID,
		description:        description,
		isTraversable:      isTraversable,
		createdAt:          createdAt,
		updatedAt:          updatedAt,
	}
}

func TestStore_GetRelationshipKind_Integration(t *testing.T) {
	t.Run("returns all relationship kind fields", func(t *testing.T) {
		var (
			store, pool = setupStore(t)
			ctx         = context.Background()
			data        = seedRelationshipKind(t, ctx, pool)
		)

		relationshipKind, err := store.GetRelationshipKind(ctx, data.relationshipKindID)

		require.NoError(t, err)
		assert.Equal(t, data.relationshipKindID, relationshipKind.ID)
		assert.Equal(t, int(data.extensionID), int(relationshipKind.Extension.ID))
		assert.Equal(t, data.kindID, relationshipKind.KindID)
		assert.Equal(t, "TestRelationshipKind", relationshipKind.Name)
		assert.Equal(t, data.description, relationshipKind.Description)
		assert.Equal(t, data.isTraversable, relationshipKind.IsTraversable)
		assert.Equal(t, data.createdAt, relationshipKind.CreatedAt)
		assert.Equal(t, data.updatedAt, relationshipKind.UpdatedAt)
	})

	t.Run("returns relationship kind not found for an unknown ID", func(t *testing.T) {
		var (
			store, _ = setupStore(t)
			ctx      = context.Background()
		)

		relationshipKind, err := store.GetRelationshipKind(ctx, int32(9999))

		assert.Equal(t, services.RelationshipKind{}, relationshipKind)
		assert.ErrorIs(t, err, services.ErrRelationshipKindNotFound)
	})
}
