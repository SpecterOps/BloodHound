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
	"errors"
	"fmt"

	"github.com/huandu/go-sqlbuilder"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/specterops/bloodhound/server/graphdb/internal/services"
)

const tableSchemaKindInfo = "schema_kind_info"

// constraint names from the schema_kind_info table (see the migration that creates it)
// that are translated into services-layer sentinels.
const (
	constraintKindInfoKindIDFKey     = "schema_kind_info_kind_id_fkey"
	constraintKindInfoUniquePosition = "schema_kind_info_unique_kind_position"
	constraintKindInfoUniqueInfoKey  = "schema_kind_info_unique_kind_info_key"
)

func (s *Store) CreateKindInfos(ctx context.Context, kindInfos []services.KindInfo) error {
	var (
		insertBuilder = sqlbuilder.PostgreSQL.NewInsertBuilder()
		sqlQuery      string
		args          []any
		err           error
	)

	if len(kindInfos) == 0 {
		return nil
	}

	insertBuilder.InsertInto(tableSchemaKindInfo)
	insertBuilder.Cols(
		"kind_id",
		"node_kind_id",
		"relationship_kind_id",
		"info_key",
		"title",
		"position",
		"content",
	)

	for _, kindInfo := range kindInfos {
		content := kindInfo.Content
		if len(content) == 0 {
			content = json.RawMessage("{}")
		}

		insertBuilder.Values(
			kindInfo.KindID,
			kindInfo.NodeKindID,
			kindInfo.RelationshipKindID,
			kindInfo.InfoKey,
			kindInfo.Title,
			kindInfo.Position,
			sqlbuilder.Build("$?::jsonb", string(content)),
		)
	}

	sqlQuery, args = insertBuilder.Build()

	if _, err = s.db.Exec(ctx, sqlQuery, args...); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.ConstraintName {
			case constraintKindInfoKindIDFKey:
				return fmt.Errorf("%w: %w", services.ErrKindInfoKindNotFound, err)
			case constraintKindInfoUniquePosition:
				return fmt.Errorf("%w: %w", services.ErrKindInfoDuplicatePosition, err)
			case constraintKindInfoUniqueInfoKey:
				return fmt.Errorf("%w: %w", services.ErrKindInfoDuplicateInfoKey, err)
			}
		}

		return fmt.Errorf("creating kind infos: %w", err)
	}

	return nil
}

// kindInfoRow is the package-local DB row type for a schema_kind_info entry.
type kindInfoRow struct {
	KindID             int32           `db:"kind_id"`
	NodeKindID         *int32          `db:"node_kind_id"`
	RelationshipKindID *int32          `db:"relationship_kind_id"`
	InfoKey            string          `db:"info_key"`
	Title              string          `db:"title"`
	Position           int32           `db:"position"`
	Content            json.RawMessage `db:"content"`
}

// toKindInfo translates a raw kind info row into the domain model.
func toKindInfo(row kindInfoRow) services.KindInfo {
	return services.KindInfo{
		KindID:             row.KindID,
		NodeKindID:         row.NodeKindID,
		RelationshipKindID: row.RelationshipKindID,
		InfoKey:            row.InfoKey,
		Title:              row.Title,
		Position:           row.Position,
		Content:            row.Content,
	}
}

// GetKindInfos returns all KindInfo's associated with the given kind_id,
// ordered by position then title. An empty slice is returned when no rows match.
func (s *Store) GetKindInfos(ctx context.Context, kindID int32) ([]services.KindInfo, error) {
	var (
		selectBuilder = sqlbuilder.PostgreSQL.NewSelectBuilder()
		sqlQuery      string
		args          []any
		rows          pgx.Rows
		kindInfoRows  []kindInfoRow
		err           error
	)

	selectBuilder.Select(
		"kind_id",
		"node_kind_id",
		"relationship_kind_id",
		"info_key",
		"title",
		"position",
		"content",
	)
	selectBuilder.From(tableSchemaKindInfo)
	selectBuilder.Where(selectBuilder.Equal("kind_id", kindID))
	selectBuilder.OrderBy("position", "title")

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
