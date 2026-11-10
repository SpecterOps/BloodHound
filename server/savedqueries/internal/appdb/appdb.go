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

package appdb

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/huandu/go-sqlbuilder"
	"github.com/jackc/pgx/v5"
	"github.com/specterops/bloodhound/server/savedqueries/internal/services"
)

const (
	tableSavedQueries            = "saved_queries"
	tableSavedQueriesPermissions = "saved_queries_permissions"
)

type pgxQuerier interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
}

type savedQueryRow struct {
	ID                int64          `db:"id"`
	UserID            sql.NullString `db:"user_id"`
	Name              sql.NullString `db:"name"`
	Query             sql.NullString `db:"query"`
	Description       sql.NullString `db:"description"`
	Category          sql.NullString `db:"category"`
	SchemaExtensionID *int32         `db:"schema_extension_id"`
	CreatedAt         sql.NullTime   `db:"created_at"`
	UpdatedAt         sql.NullTime   `db:"updated_at"`
	PublicOrShared    bool           `db:"public_or_shared"`
}

type Store struct {
	database pgxQuerier
}

func NewStore(databaseInterface pgxQuerier) *Store {
	return &Store{database: databaseInterface}
}

func (s *Store) GetSavedQuery(ctx context.Context, savedQueryID int64, actorUserID string) (services.SavedQuery, services.SavedQueryAccess, error) {
	var (
		queryBuilder = sqlbuilder.PostgreSQL.NewSelectBuilder()
		rows         pgx.Rows
		row          savedQueryRow
		err          error
	)

	publicOrSharedExpression := fmt.Sprintf(
		"EXISTS (SELECT 1 FROM %s sqp WHERE sqp.query_id = sq.id AND (sqp.public = TRUE OR sqp.shared_to_user_id = %s)) AS public_or_shared",
		tableSavedQueriesPermissions,
		queryBuilder.Var(actorUserID),
	)
	queryBuilder.Select(
		"sq.id",
		"sq.user_id",
		"sq.name",
		"sq.query",
		"sq.description",
		"sq.category",
		"sq.schema_extension_id",
		"sq.created_at",
		"sq.updated_at",
		publicOrSharedExpression,
	).From(tableSavedQueries + " sq").Where(queryBuilder.Equal("sq.id", savedQueryID)).Limit(1)

	query, args := queryBuilder.Build()
	rows, err = s.database.Query(ctx, query, args...)
	if err != nil {
		return services.SavedQuery{}, services.SavedQueryAccess{}, fmt.Errorf("querying saved query: %w", err)
	}

	row, err = pgx.CollectOneRow(rows, pgx.RowToStructByName[savedQueryRow])
	if errors.Is(err, pgx.ErrNoRows) {
		return services.SavedQuery{}, services.SavedQueryAccess{}, services.ErrSavedQueryNotFound
	}
	if err != nil {
		return services.SavedQuery{}, services.SavedQueryAccess{}, fmt.Errorf("collecting saved query: %w", err)
	}

	return services.SavedQuery{
		ID:                row.ID,
		UserID:            row.UserID.String,
		Name:              row.Name.String,
		Query:             row.Query.String,
		Description:       row.Description.String,
		Category:          row.Category.String,
		SchemaExtensionID: row.SchemaExtensionID,
		CreatedAt:         row.CreatedAt.Time,
		UpdatedAt:         row.UpdatedAt.Time,
	}, services.SavedQueryAccess{PublicOrShared: row.PublicOrShared}, nil
}
