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
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/specterops/bloodhound/server/opengraphschema/internal/services"
)

// selectEnvironmentsBase resolves the environment kind and owning extension.
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
const filterBuiltinEnvironments = ` WHERE ext.is_builtin = true`

type environmentRow struct {
	ID                         int32        `db:"id"`
	SchemaExtensionID          int32        `db:"schema_extension_id"`
	SchemaExtensionDisplayName string       `db:"schema_extension_display_name"`
	EnvironmentKindID          int32        `db:"environment_kind_id"`
	EnvironmentKindName        string       `db:"environment_kind_name"`
	SourceKindID               int32        `db:"source_kind_id"`
	CreatedAt                  time.Time    `db:"created_at"`
	UpdatedAt                  time.Time    `db:"updated_at"`
	DeletedAt                  sql.NullTime `db:"deleted_at"`
}

func toEnvironment(row environmentRow) services.Environment {
	var deletedAt *time.Time
	if row.DeletedAt.Valid {
		deletedAt = &row.DeletedAt.Time
	}
	return services.Environment{
		ID: row.ID, SchemaExtensionID: row.SchemaExtensionID,
		SchemaExtensionDisplayName: row.SchemaExtensionDisplayName,
		EnvironmentKindID:          row.EnvironmentKindID, EnvironmentKindName: row.EnvironmentKindName,
		SourceKindID: row.SourceKindID, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
		DeletedAt: deletedAt,
	}
}

// GetEnvironmentsFiltered returns schema environments ordered by id. When
// onlyBuiltin is true only environments belonging to builtin extensions are returned.
func (s *Store) GetEnvironmentsFiltered(ctx context.Context, onlyBuiltin bool) ([]services.Environment, error) {
	var query = selectEnvironmentsBase
	if onlyBuiltin {
		query += filterBuiltinEnvironments
	}
	query += orderEnvironments
	rows, err := s.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("listing schema environments: %s", err)
	}
	records, err := pgx.CollectRows(rows, pgx.RowToStructByName[environmentRow])
	if err != nil {
		return nil, fmt.Errorf("reading schema environments: %s", err)
	}
	environments := make([]services.Environment, 0, len(records))
	for _, record := range records {
		environments = append(environments, toEnvironment(record))
	}
	return environments, nil
}
