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
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/huandu/go-sqlbuilder"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/specterops/bloodhound/server/extensions/internal/services"
)

const tableSchemaKindInfo = "schema_kind_info"

// entity panel vars
var (
	graphSchemaKindInfoConstraintKindIDFKey     = "schema_kind_info_kind_id_fkey"
	graphSchemaKindInfoConstraintUniquePosition = "schema_kind_info_unique_kind_position"
	graphSchemaKindInfoConstraintUniqueInfoKey  = "schema_kind_info_unique_kind_info_key"

	// defaultKindInfoContent is stored when a kind info entry has no content, preserving the
	// expected {"markdown":{"content":"..."}} structure rather than an empty object.
	defaultKindInfoContent = `{"markdown":{"content":""}}`
)

// kindInfoRow is the package-local DB row type for a schema_kind_info entry joined with
// the kind table for the kind name. db: tags drive pgx.RowToStructByName scanning.
type kindInfoRow struct {
	ID                 int32           `db:"id"`
	KindID             int32           `db:"kind_id"`
	NodeKindID         *int32          `db:"node_kind_id"`
	RelationshipKindID *int32          `db:"relationship_kind_id"`
	Name               string          `db:"name"`
	InfoKey            string          `db:"info_key"`
	Title              string          `db:"title"`
	Position           int32           `db:"position"`
	Content            json.RawMessage `db:"content"`
	CreatedAt          time.Time       `db:"created_at"`
	UpdatedAt          time.Time       `db:"updated_at"`
}

// toKindInfo translates a raw kind info row into the domain model.
func toKindInfo(row kindInfoRow) services.KindInfo {
	return services.KindInfo{
		ID:                 row.ID,
		KindID:             row.KindID,
		NodeKindID:         row.NodeKindID,
		RelationshipKindID: row.RelationshipKindID,
		Name:               row.Name,
		InfoKey:            row.InfoKey,
		Title:              row.Title,
		Position:           row.Position,
		Content:            row.Content,
		CreatedAt:          row.CreatedAt,
		UpdatedAt:          row.UpdatedAt,
	}
}

// GetKindInfos returns all KindInfo's associated with the given kind name,
// ordered by position then title. An empty slice is returned when no rows match.
func (s *Store) GetKindInfos(ctx context.Context, kindName string) ([]services.KindInfo, error) {
	selectBuilder := sqlbuilder.PostgreSQL.NewSelectBuilder()

	selectBuilder.Select(
		"ki.id",
		"ki.kind_id",
		"ki.node_kind_id",
		"ki.relationship_kind_id",
		"k.name",
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

	sqlQuery, args := selectBuilder.Build()

	if rows, err := s.db.Query(ctx, sqlQuery, args...); err != nil {
		return nil, fmt.Errorf("fetching kind infos: %w", err)
	} else if kindInfoRows, err := pgx.CollectRows(rows, pgx.RowToStructByName[kindInfoRow]); err != nil {
		return nil, fmt.Errorf("reading rows: %w", err)
	} else {
		kindInfos := make([]services.KindInfo, 0, len(kindInfoRows))
		for _, row := range kindInfoRows {
			kindInfos = append(kindInfos, toKindInfo(row))
		}

		return kindInfos, nil
	}
}

func (s *Store) CreateKindInfo(ctx context.Context, kindID int32, nodeKindID, relationshipKindID *int32, kindInfo services.KindInfoInput) (services.KindInfo, error) {
	var (
		content       = kindInfo.Content
		insertBuilder = sqlbuilder.PostgreSQL.NewInsertBuilder()
		selectBuilder = sqlbuilder.PostgreSQL.NewSelectBuilder()
	)

	if len(content) == 0 {
		content = []byte(defaultKindInfoContent)
	}

	insertBuilder.InsertInto(tableSchemaKindInfo)
	insertBuilder.Cols(
		"kind_id", "node_kind_id", "relationship_kind_id", "info_key", "title",
		"position", "content", "created_at", "updated_at",
	)
	insertBuilder.Values(
		kindID,
		nodeKindID,
		relationshipKindID,
		kindInfo.InfoKey,
		kindInfo.Title,
		kindInfo.Position,
		sqlbuilder.Build("$0::jsonb", string(content)),
		sqlbuilder.Raw("NOW()"),
		sqlbuilder.Raw("NOW()"),
	)
	insertBuilder.Returning(
		"id", "kind_id", "node_kind_id", "relationship_kind_id",
		"info_key", "title", "position", "content", "created_at", "updated_at",
	)

	selectBuilder.With(sqlbuilder.With(sqlbuilder.CTEQuery("inserted_kind_info").As(insertBuilder)))
	selectBuilder.Select(
		"ki.id", "ki.kind_id", "ki.node_kind_id", "ki.relationship_kind_id", "k.name",
		"ki.info_key", "ki.title", "ki.position", "ki.content", "ki.created_at", "ki.updated_at",
	)
	selectBuilder.From(selectBuilder.As("inserted_kind_info", "ki"))
	selectBuilder.Join(selectBuilder.As(tableKind, "k"), "ki.kind_id = k.id")

	sqlQuery, args := selectBuilder.Build()

	if rows, err := s.db.Query(ctx, sqlQuery, args...); err != nil {
		return services.KindInfo{}, fmt.Errorf("creating kind info: %w", checkKindInfoError(err))
	} else if row, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[kindInfoRow]); err != nil {
		return services.KindInfo{}, fmt.Errorf("reading created kind info: %w", checkKindInfoError(err))
	} else {
		return toKindInfo(row), nil
	}
}

func (s *Store) UpdateKindInfo(ctx context.Context, kindInfo services.KindInfo) (services.KindInfo, error) {
	var (
		content       = kindInfo.Content
		updateBuilder = sqlbuilder.PostgreSQL.NewUpdateBuilder()
		selectBuilder = sqlbuilder.PostgreSQL.NewSelectBuilder()
	)

	if len(content) == 0 {
		content = []byte(defaultKindInfoContent)
	}

	updateBuilder.Update(tableSchemaKindInfo)
	updateBuilder.Set(
		updateBuilder.Assign("kind_id", kindInfo.KindID),
		updateBuilder.Assign("node_kind_id", kindInfo.NodeKindID),
		updateBuilder.Assign("relationship_kind_id", kindInfo.RelationshipKindID),
		updateBuilder.Assign("info_key", kindInfo.InfoKey),
		updateBuilder.Assign("title", kindInfo.Title),
		updateBuilder.Assign("position", kindInfo.Position),
		updateBuilder.Assign("content", sqlbuilder.Build("$0::jsonb", string(content))),
		updateBuilder.Assign("updated_at", sqlbuilder.Raw("NOW()")),
	)
	updateBuilder.Where(updateBuilder.Equal("id", kindInfo.ID))
	updateBuilder.Returning(
		"id", "kind_id", "node_kind_id", "relationship_kind_id",
		"info_key", "title", "position", "content", "created_at", "updated_at",
	)

	selectBuilder.With(sqlbuilder.With(sqlbuilder.CTEQuery("updated_kind_info").As(updateBuilder)))
	selectBuilder.Select(
		"ki.id", "ki.kind_id", "ki.node_kind_id", "ki.relationship_kind_id", "k.name",
		"ki.info_key", "ki.title", "ki.position", "ki.content", "ki.created_at", "ki.updated_at",
	)
	selectBuilder.From(selectBuilder.As("updated_kind_info", "ki"))
	selectBuilder.Join(selectBuilder.As(tableKind, "k"), "ki.kind_id = k.id")

	sqlQuery, args := selectBuilder.Build()

	if rows, err := s.db.Query(ctx, sqlQuery, args...); err != nil {
		return services.KindInfo{}, fmt.Errorf("updating kind info: %w", checkKindInfoError(err))
	} else if row, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[kindInfoRow]); errors.Is(err, pgx.ErrNoRows) {
		return services.KindInfo{}, services.ErrKindInfoNotFound
	} else if err != nil {
		return services.KindInfo{}, fmt.Errorf("reading updated kind info: %w", checkKindInfoError(err))
	} else {
		return toKindInfo(row), nil
	}
}

func (s *Store) DeleteKindInfo(ctx context.Context, kindInfoID int32) error {
	deleteBuilder := sqlbuilder.PostgreSQL.NewDeleteBuilder()
	deleteBuilder.DeleteFrom(tableSchemaKindInfo)
	deleteBuilder.Where(deleteBuilder.Equal("id", kindInfoID))

	sqlQuery, args := deleteBuilder.Build()

	if commandTag, err := s.db.Exec(ctx, sqlQuery, args...); err != nil {
		return fmt.Errorf("deleting kind info: %w", err)
	} else if commandTag.RowsAffected() == 0 {
		return services.ErrKindInfoNotFound
	}

	return nil
}

func checkKindInfoError(err error) error {
	var pgErr *pgconn.PgError

	if errors.As(err, &pgErr) {
		switch pgErr.ConstraintName {
		case graphSchemaKindInfoConstraintKindIDFKey:
			return fmt.Errorf("%w: %w", services.ErrKindInfoKindNotFound, err)
		case graphSchemaKindInfoConstraintUniquePosition:
			return fmt.Errorf("%w: %w", services.ErrKindInfoDuplicatePosition, err)
		case graphSchemaKindInfoConstraintUniqueInfoKey:
			return fmt.Errorf("%w: %w", services.ErrKindInfoDuplicateInfoKey, err)
		}
	}

	return err
}
