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
	"fmt"

	"github.com/gofrs/uuid"
	"github.com/huandu/go-sqlbuilder"
	"github.com/jackc/pgx/v5"
	"github.com/specterops/bloodhound/server/etac/internal/services"
)

const (
	tableEnvironmentTargetedAccessControl = "environment_targeted_access_control"
)

// pgxQuerier is the minimal pgx surface the ETAC store relies on. It is
// satisfied by both *pgxpool.Pool and the test doubles used in unit tests.
type pgxQuerier interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
}

// etacRow holds the raw scanned values for an environment_targeted_access_control
// row. The db struct tags map column names to fields for pgx.RowToStructByName.
type etacRow struct {
	UserID        string `db:"user_id"`
	EnvironmentID string `db:"environment_id"`
}

// toEnvironmentTargetedAccessControl translates a raw DB row into the domain model.
func toEnvironmentTargetedAccessControl(row etacRow) services.EnvironmentTargetedAccessControl {
	return services.EnvironmentTargetedAccessControl(row)
}

// Store performs ETAC reads directly against a PostgreSQL connection. It is the
// Database implementation.
type Store struct {
	db pgxQuerier
}

// NewStore returns a Store backed by the provided pgx querier.
func NewStore(db pgxQuerier) *Store {
	return &Store{db: db}
}

// GetEnvironmentTargetedAccessControlForUser returns every ETAC row associated
// with the supplied user. An empty slice is returned when the user has no
// environment-targeted access controls applied.
func (s *Store) GetEnvironmentTargetedAccessControlForUser(ctx context.Context, userID uuid.UUID) ([]services.EnvironmentTargetedAccessControl, error) {
	var (
		rows pgx.Rows
		err  error
	)

	selectBuilder := sqlbuilder.PostgreSQL.NewSelectBuilder()
	selectBuilder.Select("user_id", "environment_id")
	selectBuilder.From(tableEnvironmentTargetedAccessControl)
	selectBuilder.Where(selectBuilder.Equal("user_id", userID.String()))

	sqlQuery, args := selectBuilder.Build()

	rows, err = s.db.Query(ctx, sqlQuery, args...)
	if err != nil {
		return nil, err
	}

	accessControlList, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (services.EnvironmentTargetedAccessControl, error) {
		scanned, err := pgx.RowToStructByName[etacRow](row)
		return toEnvironmentTargetedAccessControl(scanned), err
	})
	if err != nil {
		return nil, fmt.Errorf("reading rows: %s", err)
	}

	return accessControlList, nil
}
