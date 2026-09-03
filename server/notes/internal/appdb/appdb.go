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
	"github.com/specterops/bloodhound/server/notes/internal/services"
)

const tableRedTeamNotes = "red_team_notes"

// queryExecer is the minimal pgx surface the notes store relies on. It is
// satisfied by both *pgxpool.Pool and the test doubles used in unit tests.
type queryExecer interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
}

// pgxQuerier extends queryExecer with the ability to begin a transaction.
type pgxQuerier interface {
	queryExecer
	BeginTx(ctx context.Context, txOptions pgx.TxOptions) (pgx.Tx, error)
}

// noteRow holds the raw scanned values for a red_team_notes row. The db struct
// tags map column names to fields for pgx.RowToStructByName.
type noteRow struct {
	ID        int64       `db:"id"`
	CreatedAt null.Time   `db:"created_at"`
	UpdatedAt null.Time   `db:"updated_at"`
	DeletedAt null.Time   `db:"deleted_at"`
	UserID    null.String `db:"user_id"`
	Title     string      `db:"title"`
	Content   string      `db:"content"`
	Type      string      `db:"note_type"`
	Tags      []string    `db:"tags"`
	URL       string      `db:"url"`
	ObjectID  null.String `db:"object_id"`
	EdgeKind  null.String `db:"edge_kind"`
}

// toNote translates a raw DB row into the domain model.
func toNote(row noteRow) services.Note {
	return services.Note{
		ID:        row.ID,
		CreatedAt: row.CreatedAt.ValueOrZero(),
		UpdatedAt: row.UpdatedAt.ValueOrZero(),
		DeletedAt: row.DeletedAt.NullTime,
		UserID:    row.UserID.ValueOrZero(),
		Title:     row.Title,
		Content:   row.Content,
		Type:      row.Type,
		Tags:      row.Tags,
		URL:       row.URL,
		ObjectID:  row.ObjectID.ValueOrZero(),
		EdgeKind:  row.EdgeKind.ValueOrZero(),
	}
}

// Store performs red team note persistence directly against a PostgreSQL
// connection. It is the Database implementation.
type Store struct {
	db pgxQuerier
}

// attachmentRow holds the raw scanned values for a red_team_note_attachments row.
type attachmentRow struct {
	ID          int64     `db:"id"`
	CreatedAt   null.Time `db:"created_at"`
	Filename    string    `db:"filename"`
	ContentType string    `db:"content_type"`
	Token       string    `db:"token"`
	Data        []byte    `db:"data"`
}

// toAttachment translates a raw DB row into the domain model.
func toAttachment(row attachmentRow) services.Attachment {
	return services.Attachment{
		ID:          row.ID,
		CreatedAt:   row.CreatedAt.ValueOrZero(),
		Filename:    row.Filename,
		ContentType: row.ContentType,
		Token:       row.Token,
		Data:        row.Data,
	}
}

// NewStore returns a Store backed by the provided pgx querier.
func NewStore(db pgxQuerier) *Store {
	return &Store{db: db}
}

// CreateNote inserts a new note and returns it with the database-assigned id
// and timestamps.
func (s *Store) CreateNote(ctx context.Context, note services.Note) (services.Note, error) {
	var (
		timestamp = time.Now().UTC()
		row       noteRow
	)

	insertBuilder := sqlbuilder.PostgreSQL.NewInsertBuilder()
	insertBuilder.InsertInto(tableRedTeamNotes)
	insertBuilder.Cols(
		"created_at",
		"updated_at",
		"user_id",
		"title",
		"content",
		"note_type",
		"tags",
		"url",
		"object_id",
		"edge_kind",
	)
	insertBuilder.Values(
		timestamp,
		timestamp,
		null.StringFrom(note.UserID),
		note.Title,
		note.Content,
		note.Type,
		note.Tags,
		note.URL,
		null.StringFrom(note.ObjectID),
		null.StringFrom(note.EdgeKind),
	)

	sqlQuery, queryArgs := insertBuilder.Build()
	sqlQuery += " RETURNING id, created_at, updated_at, deleted_at, user_id, title, content, note_type, tags, url, object_id, edge_kind"

	rows, err := s.db.Query(ctx, sqlQuery, queryArgs...)
	if err != nil {
		return services.Note{}, fmt.Errorf("inserting note: %w", err)
	}

	row, err = pgx.CollectOneRow(rows, pgx.RowToStructByName[noteRow])
	if err != nil {
		return services.Note{}, fmt.Errorf("reading inserted note: %w", err)
	}

	return toNote(row), nil
}

// GetNote returns the note for the supplied id or services.ErrNotFound.
func (s *Store) GetNote(ctx context.Context, noteID int64) (services.Note, error) {
	return selectNote(ctx, s.db, func(sb *sqlbuilder.SelectBuilder) {
		sb.Where(sb.Equal("id", noteID))
	})
}

// selectNote builds and executes a single-row SELECT against the red_team_notes
// table, applying caller-supplied WHERE conditions. It returns
// services.ErrNotFound when no row matches.
func selectNote(ctx context.Context, querier queryExecer, applyConditions func(sb *sqlbuilder.SelectBuilder)) (services.Note, error) {
	var (
		row      noteRow
		sqlQuery string
	)

	selectBuilder := newNoteSelectBuilder()
	applyConditions(selectBuilder)
	selectBuilder.Where(selectBuilder.IsNull("deleted_at"))
	selectBuilder.Limit(1)

	sqlQuery, queryArgs := selectBuilder.Build()

	rows, err := querier.Query(ctx, sqlQuery, queryArgs...)
	if errors.Is(err, pgx.ErrNoRows) {
		return services.Note{}, services.ErrNotFound
	}
	if err != nil {
		return services.Note{}, err
	}

	row, err = pgx.CollectOneRow(rows, pgx.RowToStructByName[noteRow])
	if errors.Is(err, pgx.ErrNoRows) {
		return services.Note{}, services.ErrNotFound
	}
	if err != nil {
		return services.Note{}, fmt.Errorf("reading note row: %w", err)
	}

	return toNote(row), nil
}

// newNoteSelectBuilder returns a SELECT builder primed with the full note
// column list.
func newNoteSelectBuilder() *sqlbuilder.SelectBuilder {
	var selectBuilder = sqlbuilder.PostgreSQL.NewSelectBuilder()

	selectBuilder.Select(
		"id",
		"created_at",
		"updated_at",
		"deleted_at",
		"user_id",
		"title",
		"content",
		"note_type",
		"tags",
		"url",
		"object_id",
		"edge_kind",
	)
	selectBuilder.From(tableRedTeamNotes)

	return selectBuilder
}

// UpdateNote persists changes to an existing note and returns the updated row.
func (s *Store) UpdateNote(ctx context.Context, note services.Note) (services.Note, error) {
	var (
		updateBuilder = sqlbuilder.PostgreSQL.NewUpdateBuilder()
		commandTag    pgconn.CommandTag
	)

	updateBuilder.Update(tableRedTeamNotes)
	updateBuilder.Set(
		updateBuilder.Assign("updated_at", time.Now().UTC()),
		updateBuilder.Assign("title", note.Title),
		updateBuilder.Assign("content", note.Content),
		updateBuilder.Assign("note_type", note.Type),
		updateBuilder.Assign("tags", note.Tags),
		updateBuilder.Assign("url", note.URL),
		updateBuilder.Assign("object_id", null.StringFrom(note.ObjectID)),
		updateBuilder.Assign("edge_kind", null.StringFrom(note.EdgeKind)),
	)
	updateBuilder.Where(updateBuilder.Equal("id", note.ID), updateBuilder.IsNull("deleted_at"))

	sqlQuery, queryArgs := updateBuilder.Build()

	commandTag, err := s.db.Exec(ctx, sqlQuery, queryArgs...)
	if err != nil {
		return services.Note{}, fmt.Errorf("updating note: %w", err)
	}
	if commandTag.RowsAffected() == 0 {
		return services.Note{}, services.ErrNotFound
	}

	return s.GetNote(ctx, note.ID)
}

// DeleteNote soft-deletes the note for the supplied id by setting deleted_at.
func (s *Store) DeleteNote(ctx context.Context, noteID int64) error {
	var (
		updateBuilder = sqlbuilder.PostgreSQL.NewUpdateBuilder()
		commandTag    pgconn.CommandTag
	)

	updateBuilder.Update(tableRedTeamNotes)
	updateBuilder.Set(
		updateBuilder.Assign("deleted_at", time.Now().UTC()),
		updateBuilder.Assign("updated_at", time.Now().UTC()),
	)
	updateBuilder.Where(updateBuilder.Equal("id", noteID), updateBuilder.IsNull("deleted_at"))

	sqlQuery, queryArgs := updateBuilder.Build()

	commandTag, err := s.db.Exec(ctx, sqlQuery, queryArgs...)
	if err != nil {
		return fmt.Errorf("deleting note: %w", err)
	}
	if commandTag.RowsAffected() == 0 {
		return services.ErrNotFound
	}

	return nil
}

// ListNotes returns the notes matching the supplied filter along with the total
// number of matching rows prior to pagination.
func (s *Store) ListNotes(ctx context.Context, filter services.NoteFilter) ([]services.Note, int, error) {
	var (
		conditions = buildFilterConditions
	)

	count, err := s.countNotes(ctx, filter, conditions)
	if err != nil {
		return nil, 0, err
	}

	selectBuilder := newNoteSelectBuilder()
	conditions(selectBuilder, filter)
	for _, orderClause := range sortOrderClauses(filter.Sort) {
		selectBuilder.OrderBy(orderClause)
	}
	selectBuilder.OrderBy("id DESC")
	selectBuilder.Limit(filter.Limit)
	selectBuilder.Offset(filter.Skip)

	sqlQuery, queryArgs := selectBuilder.Build()

	rows, err := s.db.Query(ctx, sqlQuery, queryArgs...)
	if errors.Is(err, pgx.ErrNoRows) {
		return []services.Note{}, count, nil
	}
	if err != nil {
		return nil, 0, err
	}

	dbRows, err := pgx.CollectRows(rows, pgx.RowToStructByName[noteRow])
	if errors.Is(err, pgx.ErrNoRows) {
		return []services.Note{}, count, nil
	}
	if err != nil {
		return nil, 0, fmt.Errorf("reading note rows: %w", err)
	}

	notes := make([]services.Note, 0, len(dbRows))
	for _, row := range dbRows {
		notes = append(notes, toNote(row))
	}

	return notes, count, nil
}

// countNotes returns the number of rows matching the supplied filter prior to
// pagination.
func (s *Store) countNotes(ctx context.Context, filter services.NoteFilter, applyConditions func(sb *sqlbuilder.SelectBuilder, filter services.NoteFilter)) (int, error) {
	var (
		countBuilder = sqlbuilder.PostgreSQL.NewSelectBuilder()
		count        int64
	)

	countBuilder.Select("COUNT(*)")
	countBuilder.From(tableRedTeamNotes)
	applyConditions(countBuilder, filter)

	sqlQuery, queryArgs := countBuilder.Build()

	rows, err := s.db.Query(ctx, sqlQuery, queryArgs...)
	if err != nil {
		return 0, fmt.Errorf("counting notes: %w", err)
	}

	count, err = pgx.CollectOneRow(rows, pgx.RowTo[int64])
	if err != nil {
		return 0, fmt.Errorf("counting notes: %w", err)
	}

	return int(count), nil
}

// buildFilterConditions applies the shared WHERE clause used by both the list
// and count queries for a note filter.
func buildFilterConditions(sb *sqlbuilder.SelectBuilder, filter services.NoteFilter) {
	sb.Where(sb.IsNull("deleted_at"))

	if filter.ObjectID != "" {
		sb.Where(sb.Equal("object_id", filter.ObjectID))
	}

	if filter.EdgeKind != "" {
		sb.Where(sb.Equal("edge_kind", filter.EdgeKind))
	}

	if filter.Type != "" {
		sb.Where(sb.Equal("note_type", filter.Type))
	}

	for _, filterTag := range filter.Tags {
		if filterTag != "" {
			sb.Where("tags @> ARRAY[" + sb.Var(filterTag) + "]")
		}
	}
	if filter.Search != "" {
		searchPattern := "%" + filter.Search + "%"
		sb.Where(sb.Or(sb.ILike("title", searchPattern), sb.ILike("content", searchPattern)))
	}
}

// sortOrderClauses translates a validated sort token into SQL ORDER BY clauses.
// The id tiebreaker is appended by the caller to guarantee deterministic paging.
func sortOrderClauses(sort string) []string {
	switch sort {
	case services.SortUpdatedAtAsc:
		return []string{"updated_at ASC"}
	case services.SortTitleAsc:
		return []string{"title ASC"}
	case services.SortTitleDesc:
		return []string{"title DESC"}
	case services.SortUpdatedAtDesc:
		fallthrough
	default:
		return []string{"updated_at DESC"}
	}
}

// ListTags returns the distinct tags across live notes ordered by usage then name.
func (s *Store) ListTags(ctx context.Context) ([]services.TagCount, error) {
	var (
		selectBuilder = sqlbuilder.PostgreSQL.NewSelectBuilder()
	)

	selectBuilder.Select("unnest(tags) AS tag", "COUNT(*) AS usage_count")
	selectBuilder.From(tableRedTeamNotes)
	selectBuilder.Where(selectBuilder.IsNull("deleted_at"))
	selectBuilder.GroupBy("tag")
	selectBuilder.OrderBy("usage_count DESC", "tag ASC")

	sqlQuery, queryArgs := selectBuilder.Build()

	rows, err := s.db.Query(ctx, sqlQuery, queryArgs...)
	if err != nil {
		return nil, fmt.Errorf("listing tags: %w", err)
	}

	tagRows, err := pgx.CollectRows(rows, pgx.RowToStructByPos[struct {
		Tag   string
		Count int
	}])
	if err != nil {
		return nil, fmt.Errorf("reading tag rows: %w", err)
	}

	tags := make([]services.TagCount, 0, len(tagRows))
	for _, row := range tagRows {
		tags = append(tags, services.TagCount{Tag: row.Tag, Count: row.Count})
	}

	return tags, nil
}

// CreateAttachment inserts an attachment row and returns it with the assigned id.
func (s *Store) CreateAttachment(ctx context.Context, attachment services.Attachment) (services.Attachment, error) {
	var (
		insertBuilder = sqlbuilder.PostgreSQL.NewInsertBuilder()
	)

	insertBuilder.InsertInto("red_team_note_attachments")
	insertBuilder.Cols("created_at", "filename", "content_type", "token", "data")
	insertBuilder.Values(time.Now().UTC(), attachment.Filename, attachment.ContentType, attachment.Token, attachment.Data)

	sqlQuery, queryArgs := insertBuilder.Build()
	sqlQuery += " RETURNING id, created_at, filename, content_type, token, data"

	rows, err := s.db.Query(ctx, sqlQuery, queryArgs...)
	if err != nil {
		return services.Attachment{}, fmt.Errorf("inserting attachment: %w", err)
	}

	row, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[attachmentRow])
	if err != nil {
		return services.Attachment{}, fmt.Errorf("reading inserted attachment: %w", err)
	}

	return toAttachment(row), nil
}

// GetAttachment returns the attachment for the supplied id or services.ErrNotFound.
func (s *Store) GetAttachment(ctx context.Context, attachmentID int64) (services.Attachment, error) {
	var (
		selectBuilder = sqlbuilder.PostgreSQL.NewSelectBuilder()
	)

	selectBuilder.Select("id", "created_at", "filename", "content_type", "token", "data")
	selectBuilder.From("red_team_note_attachments")
	selectBuilder.Where(selectBuilder.Equal("id", attachmentID))
	selectBuilder.Limit(1)

	sqlQuery, queryArgs := selectBuilder.Build()

	rows, err := s.db.Query(ctx, sqlQuery, queryArgs...)
	if errors.Is(err, pgx.ErrNoRows) {
		return services.Attachment{}, services.ErrNotFound
	}
	if err != nil {
		return services.Attachment{}, err
	}

	row, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[attachmentRow])
	if errors.Is(err, pgx.ErrNoRows) {
		return services.Attachment{}, services.ErrNotFound
	}
	if err != nil {
		return services.Attachment{}, fmt.Errorf("reading attachment row: %w", err)
	}

	return toAttachment(row), nil
}

// GetAttachmentByToken returns the attachment for the supplied unguessable
// token or services.ErrNotFound.
func (s *Store) GetAttachmentByToken(ctx context.Context, token string) (services.Attachment, error) {
	var (
		selectBuilder = sqlbuilder.PostgreSQL.NewSelectBuilder()
	)

	selectBuilder.Select("id", "created_at", "filename", "content_type", "token", "data")
	selectBuilder.From("red_team_note_attachments")
	selectBuilder.Where(selectBuilder.Equal("token", token))
	selectBuilder.Limit(1)

	sqlQuery, queryArgs := selectBuilder.Build()

	rows, err := s.db.Query(ctx, sqlQuery, queryArgs...)
	if errors.Is(err, pgx.ErrNoRows) {
		return services.Attachment{}, services.ErrNotFound
	}
	if err != nil {
		return services.Attachment{}, err
	}

	row, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[attachmentRow])
	if errors.Is(err, pgx.ErrNoRows) {
		return services.Attachment{}, services.ErrNotFound
	}
	if err != nil {
		return services.Attachment{}, fmt.Errorf("reading attachment row: %w", err)
	}

	return toAttachment(row), nil
}
