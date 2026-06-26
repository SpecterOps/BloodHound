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
	"github.com/specterops/bloodhound/server/graphdb/internal/appdb"
	"github.com/specterops/bloodhound/server/graphdb/internal/appdb/mocks"
	"github.com/specterops/bloodhound/server/graphdb/internal/services"
	"github.com/specterops/dawgs/graph"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestStore_GetRelationship(t *testing.T) {
	ctx := context.Background()

	t.Run("maps not-found error to ErrRelationshipNotFound", func(t *testing.T) {
		graphReaderMock := mocks.NewMockgraphReader(t)
		graphReaderMock.EXPECT().ReadTransaction(ctx, mock.Anything).Return(graph.ErrNoResultsFound)
		store := appdb.NewStore(graphReaderMock, nil)

		_, err := store.GetRelationship(ctx, 1)
		assert.ErrorIs(t, err, services.ErrRelationshipNotFound)
	})

	t.Run("propagates real transaction error", func(t *testing.T) {
		txErr := errors.New("connection refused")
		graphReaderMock := mocks.NewMockgraphReader(t)
		graphReaderMock.EXPECT().ReadTransaction(ctx, mock.Anything).Return(txErr)
		store := appdb.NewStore(graphReaderMock, nil)

		_, err := store.GetRelationship(ctx, 1)
		assert.ErrorIs(t, err, txErr)
		assert.NotErrorIs(t, err, services.ErrRelationshipNotFound)
	})

	t.Run("returns ErrRelationshipNotFound when relationship is nil without error", func(t *testing.T) {
		graphReaderMock := mocks.NewMockgraphReader(t)
		graphReaderMock.EXPECT().ReadTransaction(ctx, mock.Anything).Return(nil)
		store := appdb.NewStore(graphReaderMock, nil)

		_, err := store.GetRelationship(ctx, 1)
		assert.ErrorIs(t, err, services.ErrRelationshipNotFound)
	})
}

func TestStore_GetKindByName(t *testing.T) {
	var (
		ctx            = context.Background()
		kindName       = "MemberOf"
		expectedSQL    = `SELECT rk.id, k.name FROM kind AS k LEFT JOIN schema_relationship_kinds AS rk ON rk.kind_id = k.id WHERE k.name = $1 LIMIT $2`
		expectedKindID = int32(42)
	)

	pool, err := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual))
	require.NoError(t, err)
	defer pool.Close()

	store := appdb.NewStore(nil, pool)

	t.Run("returns kind on success", func(t *testing.T) {
		pool.ExpectQuery(expectedSQL).
			WithArgs(kindName, 1).
			WillReturnRows(pgxmock.NewRows([]string{"id", "name"}).AddRow(&expectedKindID, kindName))

		kind, err := store.GetKindByName(ctx, kindName)
		require.NoError(t, err)
		require.NotNil(t, kind.ID)
		assert.Equal(t, expectedKindID, *kind.ID)
		assert.Equal(t, kindName, kind.Name)
	})

	t.Run("returns kind with nil id when no schema_relationship_kinds row", func(t *testing.T) {
		pool.ExpectQuery(expectedSQL).
			WithArgs(kindName, 1).
			WillReturnRows(pgxmock.NewRows([]string{"id", "name"}).AddRow(nil, kindName))

		kind, err := store.GetKindByName(ctx, kindName)
		require.NoError(t, err)
		assert.Nil(t, kind.ID)
		assert.Equal(t, kindName, kind.Name)
	})

	t.Run("returns ErrKindNotFound when no rows", func(t *testing.T) {
		pool.ExpectQuery(expectedSQL).
			WithArgs("Unknown", 1).
			WillReturnRows(pgxmock.NewRows([]string{"id", "name"}))

		_, err := store.GetKindByName(ctx, "Unknown")
		assert.ErrorIs(t, err, services.ErrKindNotFound)
	})

	t.Run("returns error on database error", func(t *testing.T) {
		dbErr := errors.New("db error")
		pool.ExpectQuery(expectedSQL).
			WithArgs(kindName, 1).
			WillReturnError(dbErr)

		_, err := store.GetKindByName(ctx, kindName)
		assert.ErrorIs(t, err, dbErr)
	})
}
