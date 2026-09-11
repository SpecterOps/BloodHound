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
	"encoding/json"
	"fmt"

	"github.com/huandu/go-sqlbuilder"
	"github.com/jackc/pgx/v5"
	"github.com/specterops/bloodhound/cmd/api/src/database/types/null"
	"github.com/specterops/bloodhound/server/graphdb/internal/services"
)

const tableSchemaKindInfo = "schema_kind_info"

// kindInfoRow is the package-local DB row type for a schema_kind_info entry.
type kindInfoRow struct {
	ID                 int32           `db:"id"`
	KindID             int32           `db:"kind_id"`
	NodeKindID         *int32          `db:"node_kind_id"`
	RelationshipKindID *int32          `db:"relationship_kind_id"`
	InfoKey            string          `db:"info_key"`
	Title              string          `db:"title"`
	Position           int32           `db:"position"`
	Content            json.RawMessage `db:"content"`
	CreatedAt          null.Time       `db:"created_at"`
	UpdatedAt          null.Time       `db:"updated_at"`
}

// toKindInfo translates a raw kind info row into the domain model.
func toKindInfo(row kindInfoRow) services.KindInfo {
	return services.KindInfo{
		ID:                 row.ID,
		KindID:             row.KindID,
		NodeKindID:         row.NodeKindID,
		RelationshipKindID: row.RelationshipKindID,
		InfoKey:            row.InfoKey,
		Title:              row.Title,
		Position:           row.Position,
		Content:            row.Content,
		CreatedAt:          row.CreatedAt.ValueOrZero(),
		UpdatedAt:          row.UpdatedAt.ValueOrZero(),
	}
}

// GetKindInfos returns all KindInfo's associated with the given kind name,
// ordered by position then title. An empty slice is returned when no rows match.
func (s *Store) GetKindInfos(ctx context.Context, kindName string) ([]services.KindInfo, error) {
	var (
		selectBuilder = sqlbuilder.PostgreSQL.NewSelectBuilder()
		sqlQuery      string
		args          []any
		rows          pgx.Rows
		kindInfoRows  []kindInfoRow
		err           error
	)

	selectBuilder.Select(
		"ki.id",
		"ki.kind_id",
		"ki.node_kind_id",
		"ki.relationship_kind_id",
		"ki.info_key",
		"ki.title",
		"ki.position",
		"ki.content",
		"ki.created_at",
		"ki.updated_at",
	)
	selectBuilder.From(selectBuilder.As(tableSchemaKindInfo, "ki"))
	selectBuilder.Join(selectBuilder.As(tableKind, "k"), "ki.kind_id = k.id")
	selectBuilder.Where(selectBuilder.Equal("k.name", kindName))
	selectBuilder.OrderBy("ki.position", "ki.title")

	sqlQuery, args = selectBuilder.Build()

	if rows, err = s.db.Query(ctx, sqlQuery, args...); err != nil {
		return nil, fmt.Errorf("fetching kind infos: %w", err)
	} else if kindInfoRows, err = pgx.CollectRows(rows, pgx.RowToStructByName[kindInfoRow]); err != nil {
		return nil, fmt.Errorf("reading rows: %w", err)
	}

	kindInfos := make([]services.KindInfo, 0, len(kindInfoRows))
	for _, row := range kindInfoRows {
		kindInfos = append(kindInfos, toKindInfo(row))
	}

	return kindInfos, nil
}
