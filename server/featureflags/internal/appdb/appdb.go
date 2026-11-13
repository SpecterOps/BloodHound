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
	"time"

	"github.com/huandu/go-sqlbuilder"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/specterops/bloodhound/cmd/api/src/database/types/null"
	"github.com/specterops/bloodhound/server/featureflags/internal/services"
)

const tableFeatureFlags = "feature_flags"

// queryExecer is the minimal pgx surface the feature-flag store relies on. It is
// satisfied by both *pgxpool.Pool and the test doubles used in unit tests.
type queryExecer interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
}

// pgxQuerier extends queryExecer with the ability to begin a transaction.
// Only *pgxpool.Pool satisfies this full interface; pgx.Tx satisfies only queryExecer.
type pgxQuerier interface {
	queryExecer
	BeginTx(ctx context.Context, txOptions pgx.TxOptions) (pgx.Tx, error)
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
	return selectFeatureFlag(ctx, s.db, func(sb *sqlbuilder.SelectBuilder) {
		sb.Where(sb.Equal("key", key))
	})
}

// GetFlagByID returns the feature flag for the supplied id, or ErrNotFound
// when no matching flag exists.
func (s *Store) GetFlagByID(ctx context.Context, id int32) (services.FeatureFlag, error) {
	return selectFeatureFlag(ctx, s.db, func(sb *sqlbuilder.SelectBuilder) {
		sb.Where(sb.Equal("id", id))
	})
}

// selectFeatureFlag builds and executes a single-row SELECT against the
// feature_flags table, applying caller-supplied WHERE conditions via
// applyConditions. It returns ErrNotFound when no row matches.
func selectFeatureFlag(ctx context.Context, querier queryExecer, applyConditions func(sb *sqlbuilder.SelectBuilder)) (services.FeatureFlag, error) {
	var (
		row       featureFlagRow
		rows      pgx.Rows
		err       error
		sqlQuery  string
		queryArgs []any
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
	applyConditions(selectBuilder)
	selectBuilder.Limit(1)

	sqlQuery, queryArgs = selectBuilder.Build()

	rows, err = querier.Query(ctx, sqlQuery, queryArgs...)
	if errors.Is(err, pgx.ErrNoRows) {
		return services.FeatureFlag{}, services.ErrNotFound
	}
	if err != nil {
		return services.FeatureFlag{}, err
	}

	row, err = pgx.CollectOneRow(rows, pgx.RowToStructByName[featureFlagRow])
	if errors.Is(err, pgx.ErrNoRows) {
		return services.FeatureFlag{}, services.ErrNotFound
	}
	if err != nil {
		return services.FeatureFlag{}, fmt.Errorf("reading rows: %w", err)
	}
	return toFeatureFlag(row), nil
}

// GetAllFlags returns every feature flag in the feature_flags table.
func (s *Store) GetAllFlags(ctx context.Context) ([]services.FeatureFlag, error) {
	var (
		rows      pgx.Rows
		err       error
		sqlQuery  string
		queryArgs []any
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

	sqlQuery, queryArgs = selectBuilder.Build()

	rows, err = s.db.Query(ctx, sqlQuery, queryArgs...)
	if err != nil {
		return nil, err
	}

	dbRows, err := pgx.CollectRows(rows, pgx.RowToStructByName[featureFlagRow])
	if err != nil {
		return nil, fmt.Errorf("reading rows: %w", err)
	}

	flags := make([]services.FeatureFlag, 0, len(dbRows))
	for _, row := range dbRows {
		flags = append(flags, toFeatureFlag(row))
	}
	return flags, nil
}

// SetFlag updates a feature flag's enablement.
func (s *Store) SetFlag(ctx context.Context, flag services.FeatureFlag) error {
	var (
		updateBuilder = sqlbuilder.PostgreSQL.NewUpdateBuilder()
		tx            pgx.Tx
		commandTag    pgconn.CommandTag
		err           error
	)

	tx, err = s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	defer func() {
		// Rollback is a no-op once the transaction has been committed.
		_ = tx.Rollback(ctx)
	}()

	updateBuilder.Update(tableFeatureFlags)
	updateBuilder.Set(
		updateBuilder.Assign("enabled", flag.Enabled),
		updateBuilder.Assign("updated_at", time.Now().UTC()),
	)
	updateBuilder.Where(updateBuilder.Equal("id", flag.ID))

	sqlQuery, args := updateBuilder.Build()
	commandTag, err = tx.Exec(ctx, sqlQuery, args...)
	if err != nil {
		return err
	}
	if commandTag.RowsAffected() == 0 {
		return services.ErrNotFound
	}

	return tx.Commit(ctx)
}
