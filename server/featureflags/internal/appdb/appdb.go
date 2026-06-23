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
package appdb

import (
	"context"
	"errors"
	"fmt"

	"github.com/huandu/go-sqlbuilder"
	"github.com/jackc/pgx/v5"
	"github.com/specterops/bloodhound/cmd/api/src/database/types/null"
	"github.com/specterops/bloodhound/server/featureflags/internal/services"
)

const (
	tableFeatureFlags = "feature_flags"
)

// pgxQuerier is the minimal pgx surface the feature-flag store relies on. It is
// satisfied by both *pgxpool.Pool and the test doubles used in unit tests.
type pgxQuerier interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
}

// featureFlagRow holds the raw scanned values for a feature_flags row. The db
// struct tags map column names to fields for pgx.RowToStructByName.
type featureFlagRow struct {
	ID            int32     `db:"id"`
	CreatedAt     null.Time `db:"created_at"`
	UpdatedAt     null.Time `db:"updated_at"`
	Key           string    `db:"key"`
	Name          string    `db:"name"`
	Description   string    `db:"description"`
	Enabled       bool      `db:"enabled"`
	UserUpdatable bool      `db:"user_updatable"`
}

// toFeatureFlag translates a raw DB row into the domain model.
func toFeatureFlag(row featureFlagRow) services.FeatureFlag {
	return services.FeatureFlag{
		ID:            row.ID,
		CreatedAt:     row.CreatedAt.ValueOrZero(),
		UpdatedAt:     row.UpdatedAt.ValueOrZero(),
		Key:           row.Key,
		Name:          row.Name,
		Description:   row.Description,
		Enabled:       row.Enabled,
		UserUpdatable: row.UserUpdatable,
	}
}

// Store performs feature-flag reads directly against a PostgreSQL connection. It
// is the Database implementation.
type Store struct {
	db pgxQuerier
}

// NewStore returns a Store backed by the provided pgx querier.
func NewStore(db pgxQuerier) *Store {
	return &Store{db: db}
}

// GetFlagByKey returns the feature flag for the supplied key, or ErrNotFound
// when no matching flag exists.
func (s *Store) GetFlagByKey(ctx context.Context, key string) (services.FeatureFlag, error) {
	var (
		row  featureFlagRow
		rows pgx.Rows
		err  error
	)

	selectBuilder := sqlbuilder.PostgreSQL.NewSelectBuilder()
	selectBuilder.Select(
		"id",
		"created_at",
		"updated_at",
		"key",
		"name",
		"description",
		"enabled",
		"user_updatable",
	)
	selectBuilder.From(tableFeatureFlags)
	selectBuilder.Where(selectBuilder.Equal("key", key))
	selectBuilder.Limit(1)

	sqlQuery, args := selectBuilder.Build()

	rows, err = s.db.Query(ctx, sqlQuery, args...)
	if err != nil {
		return services.FeatureFlag{}, err
	}

	row, err = pgx.CollectOneRow(rows, pgx.RowToStructByName[featureFlagRow])
	if errors.Is(err, pgx.ErrNoRows) {
		return services.FeatureFlag{}, services.ErrNotFound
	}
	if err != nil {
		return services.FeatureFlag{}, fmt.Errorf("reading rows: %s", err)
	}
	return toFeatureFlag(row), nil
}
