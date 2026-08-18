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

// Package appdb stores open graph schema data in PostgreSQL. It owns the SQL for
// this package and reads schema environments directly from the pgx pool.
package appdb

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/specterops/bloodhound/cmd/api/src/model"
)

// selectEnvironmentsBase reads schema environments joined to their kind name and
// owning extension. It mirrors the legacy GetEnvironmentsFiltered projection so
// callers observe the same environment data.
const selectEnvironmentsBase = `
	SELECT
		se.id,
		se.schema_extension_id,
		ext.display_name AS schema_extension_display_name,
		se.environment_kind_id,
		k.name AS environment_kind_name,
		se.source_kind_id,
		se.created_at,
		se.updated_at,
		se.deleted_at
	FROM schema_environments se
	INNER JOIN kind k ON se.environment_kind_id = k.id
	INNER JOIN schema_extensions ext ON se.schema_extension_id = ext.id`

const orderEnvironments = ` ORDER BY se.id`

// filterBuiltinEnvironments narrows the base query to builtin extensions only.
const filterBuiltinEnvironments = ` WHERE ext.is_builtin = true`

// pgxQuerier lists only the pgx methods this package actually calls against
// PostgreSQL, so the same Store code can run against either a pool or a
// transaction.
type pgxQuerier interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
}

// Store reads open graph schema data from PostgreSQL.
type Store struct {
	db pgxQuerier
}

// NewStore returns a Store backed by the provided pgx connection pool.
func NewStore(db pgxQuerier) *Store {
	return &Store{db: db}
}

// GetEnvironmentsFiltered returns schema environments ordered by id. When
// onlyBuiltin is true only environments belonging to builtin extensions are
// returned.
func (s *Store) GetEnvironmentsFiltered(ctx context.Context, onlyBuiltin bool) ([]model.SchemaEnvironment, error) {
	var query = selectEnvironmentsBase
	if onlyBuiltin {
		query += filterBuiltinEnvironments
	}
	query += orderEnvironments

	rows, err := s.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var environments = []model.SchemaEnvironment{}
	for rows.Next() {
		var environment model.SchemaEnvironment
		if err := rows.Scan(
			&environment.ID,
			&environment.SchemaExtensionId,
			&environment.SchemaExtensionDisplayName,
			&environment.EnvironmentKindId,
			&environment.EnvironmentKindName,
			&environment.SourceKindId,
			&environment.CreatedAt,
			&environment.UpdatedAt,
			&environment.DeletedAt,
		); err != nil {
			return nil, err
		}
		environments = append(environments, environment)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return environments, nil
}
