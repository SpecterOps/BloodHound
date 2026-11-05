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
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/pashagolub/pgxmock/v4"
	"github.com/specterops/bloodhound/server/savedqueries/internal/appdb"
	"github.com/specterops/bloodhound/server/savedqueries/internal/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const expectedGetSavedQuerySQL = `SELECT sq.id, sq.user_id, sq.name, sq.query, sq.description, sq.category, sq.schema_extension_id, sq.created_at, sq.updated_at, sq.deleted_at, EXISTS (SELECT 1 FROM saved_queries_permissions sqp WHERE sqp.query_id = sq.id AND (sqp.public = TRUE OR sqp.shared_to_user_id = $1)) AS public_or_shared FROM saved_queries sq WHERE sq.id = $2 LIMIT $3`

func TestStore_GetSavedQuery(t *testing.T) {
	type mock struct {
		pool pgxmock.PgxPoolIface
	}

	type expected struct {
		savedQuery  services.SavedQuery
		access      services.SavedQueryAccess
		err         error
		errContains string
	}

	type testData struct {
		name       string
		setupMocks func(mock)
		expected   expected
	}

	var (
		ctx               = context.Background()
		savedQueryID      = int64(42)
		actorUserID       = "11111111-1111-1111-1111-111111111111"
		schemaExtensionID = int32(7)
		createdAt         = time.Date(2026, 11, 1, 12, 0, 0, 0, time.UTC)
		updatedAt         = time.Date(2026, 11, 2, 12, 0, 0, 0, time.UTC)
		unexpectedErr     = errors.New("database unavailable")
		savedQuery        = services.SavedQuery{
			ID:                savedQueryID,
			UserID:            "00000000-0000-0000-0000-000000000000",
			Name:              "Extension query",
			Query:             "MATCH (n) RETURN n",
			Description:       "Provided by an extension",
			Category:          "extension",
			SchemaExtensionID: &schemaExtensionID,
			CreatedAt:         createdAt,
			UpdatedAt:         updatedAt,
			DeletedAt:         sql.NullTime{},
		}
	)

	rowColumns := []string{"id", "user_id", "name", "query", "description", "category", "schema_extension_id", "created_at", "updated_at", "deleted_at", "public_or_shared"}
	tt := []testData{
		{
			name: "Success: query and actor-relative access are returned - 200",
			setupMocks: func(mock mock) {
				mock.pool.ExpectQuery(expectedGetSavedQuerySQL).WithArgs(actorUserID, savedQueryID, 1).WillReturnRows(
					mock.pool.NewRows(rowColumns).AddRow(
						savedQuery.ID,
						savedQuery.UserID,
						savedQuery.Name,
						savedQuery.Query,
						savedQuery.Description,
						savedQuery.Category,
						savedQuery.SchemaExtensionID,
						savedQuery.CreatedAt,
						savedQuery.UpdatedAt,
						nil,
						true,
					),
				)
			},
			expected: expected{savedQuery: savedQuery, access: services.SavedQueryAccess{PublicOrShared: true}},
		},
		{
			name: "Error: missing query maps to domain sentinel - 404",
			setupMocks: func(mock mock) {
				mock.pool.ExpectQuery(expectedGetSavedQuerySQL).WithArgs(actorUserID, savedQueryID, 1).WillReturnRows(
					mock.pool.NewRows(rowColumns),
				)
			},
			expected: expected{err: services.ErrSavedQueryNotFound},
		},
		{
			name: "Error: database query failure is wrapped - 500",
			setupMocks: func(mock mock) {
				mock.pool.ExpectQuery(expectedGetSavedQuerySQL).WithArgs(actorUserID, savedQueryID, 1).WillReturnError(unexpectedErr)
			},
			expected: expected{errContains: "querying saved query: database unavailable"},
		},
		{
			name: "Error: row collection failure is wrapped - 500",
			setupMocks: func(mock mock) {
				mock.pool.ExpectQuery(expectedGetSavedQuerySQL).WithArgs(actorUserID, savedQueryID, 1).WillReturnRows(
					mock.pool.NewRows(rowColumns).AddRow(
						savedQuery.ID,
						savedQuery.UserID,
						savedQuery.Name,
						savedQuery.Query,
						savedQuery.Description,
						savedQuery.Category,
						savedQuery.SchemaExtensionID,
						savedQuery.CreatedAt,
						savedQuery.UpdatedAt,
						nil,
						true,
					).RowError(0, unexpectedErr),
				)
			},
			expected: expected{errContains: "collecting saved query: database unavailable"},
		},
	}

	for _, testCase := range tt {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			pool, err := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual))
			require.NoError(t, err)
			t.Cleanup(pool.Close)

			testCase.setupMocks(mock{pool: pool})
			store := appdb.NewStore(pool)
			actualSavedQuery, actualAccess, err := store.GetSavedQuery(ctx, savedQueryID, actorUserID)

			assert.Equal(t, testCase.expected.savedQuery, actualSavedQuery)
			assert.Equal(t, testCase.expected.access, actualAccess)
			if testCase.expected.err != nil {
				assert.ErrorIs(t, err, testCase.expected.err)
			} else if testCase.expected.errContains != "" {
				assert.EqualError(t, err, testCase.expected.errContains)
			} else {
				assert.NoError(t, err)
			}
			require.NoError(t, pool.ExpectationsWereMet())
		})
	}
}
