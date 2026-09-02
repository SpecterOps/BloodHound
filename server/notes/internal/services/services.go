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
package services

//go:generate go tool mockery

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/gofrs/uuid"
)

const (
	NoteTypeGeneral   = "general"
	NoteTypeTechnique = "technique"
	NoteTypeTool      = "tool"
	NoteTypeSource    = "source"

	DefaultNoteListLimit = 50
	MaxNoteListLimit     = 500

	MaxAttachmentSize = 5 * 1024 * 1024

	SortUpdatedAtDesc = "-updated_at"
	SortUpdatedAtAsc  = "updated_at"
	SortTitleAsc      = "title"
	SortTitleDesc     = "-title"
)

var (
	ErrNotFound                  = errors.New("note not found")
	ErrTitleRequired             = errors.New("note title is required")
	ErrInvalidType               = errors.New("note type must be one of: general, technique, tool, source")
	ErrInvalidSort               = errors.New("sort must be one of: updated_at, -updated_at, title, -title")
	ErrAttachmentTooLarge        = errors.New("attachment exceeds the 5MB limit")
	ErrAttachmentTypeUnsupported = errors.New("attachment must be a png, jpeg, gif, webp or svg image")
	ErrAttachmentEmpty           = errors.New("attachment data is required")
)

var validAttachmentContentTypes = map[string]bool{
	"image/png":     true,
	"image/jpeg":    true,
	"image/gif":     true,
	"image/webp":    true,
	"image/svg+xml": true,
}

var validNoteTypes = map[string]bool{
	NoteTypeGeneral:   true,
	NoteTypeTechnique: true,
	NoteTypeTool:      true,
	NoteTypeSource:    true,
}

var validNoteSorts = map[string]bool{
	SortUpdatedAtDesc: true,
	SortUpdatedAtAsc:  true,
	SortTitleAsc:      true,
	SortTitleDesc:     true,
}

// Note is the domain representation of a row in the red_team_notes table. Notes
// are free-form red team knowledge entries (techniques, tooling, source code
// references) that may optionally be linked to a graph node via ObjectID or to
// an edge kind via EdgeKind.
type Note struct {
	ID        int64
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt sql.NullTime
	UserID    string
	Title     string
	Content   string
	Type      string
	Tags      []string
	URL       string
	ObjectID  string
	EdgeKind  string
}

// NoteFilter narrows the result set of ListNotes. Empty fields are ignored.
// Tags uses AND semantics: a note must contain every supplied tag to match.
type NoteFilter struct {
	ObjectID string
	EdgeKind string
	Type     string
	Tags     []string
	Search   string
	Sort     string
	Skip     int
	Limit    int
}

// Attachment is an uploaded binary (image) referenced from note markdown via
// the attachment serving endpoint.
type Attachment struct {
	ID          int64
	CreatedAt   time.Time
	Filename    string
	ContentType string
	Token       string
	Data        []byte
}

// TagCount is a distinct tag together with the number of live notes using it.
type TagCount struct {
	Tag   string
	Count int
}

// Validate normalizes defaults and enforces invariants on a note prior to
// persistence.
func ValidateNote(note Note) error {
	if strings.TrimSpace(note.Title) == "" {
		return ErrTitleRequired
	}

	if note.Type == "" {
		note.Type = NoteTypeGeneral
	}

	if !validNoteTypes[note.Type] {
		return ErrInvalidType
	}

	return nil
}

// Database describes the persistence capabilities the notes Service requires.
// Implementations translate driver-level not-found errors into ErrNotFound so
// the Service can reason about them in domain terms.
type Database interface {
	CreateNote(ctx context.Context, note Note) (Note, error)
	GetNote(ctx context.Context, noteID int64) (Note, error)
	UpdateNote(ctx context.Context, note Note) (Note, error)
	DeleteNote(ctx context.Context, noteID int64) error
	ListNotes(ctx context.Context, filter NoteFilter) ([]Note, int, error)
	ListTags(ctx context.Context) ([]TagCount, error)
	CreateAttachment(ctx context.Context, attachment Attachment) (Attachment, error)
	GetAttachment(ctx context.Context, attachmentID int64) (Attachment, error)
	GetAttachmentByToken(ctx context.Context, token string) (Attachment, error)
}

// Service implements red team note use cases on top of a Database implementation.
type Service struct {
	db Database
}

// NewService constructs a Service from the supplied Database port.
func NewService(db Database) *Service {
	if db == nil {
		panic("notes: service requires a non-nil Database")
	}

	return &Service{db: db}
}

// CreateNote validates and persists a new note.
func (s *Service) CreateNote(ctx context.Context, note Note) (Note, error) {
	if note.Type == "" {
		note.Type = NoteTypeGeneral
	}

	if note.Tags == nil {
		note.Tags = []string{}
	}

	if err := ValidateNote(note); err != nil {
		return Note{}, err
	}

	return s.db.CreateNote(ctx, note)
}

// GetNote returns the note for the supplied id or ErrNotFound.
func (s *Service) GetNote(ctx context.Context, noteID int64) (Note, error) {
	return s.db.GetNote(ctx, noteID)
}

// UpdateNote validates and persists changes to an existing note.
func (s *Service) UpdateNote(ctx context.Context, note Note) (Note, error) {
	if note.Tags == nil {
		note.Tags = []string{}
	}

	if err := ValidateNote(note); err != nil {
		return Note{}, err
	}

	return s.db.UpdateNote(ctx, note)
}

// DeleteNote soft-deletes the note for the supplied id.
func (s *Service) DeleteNote(ctx context.Context, noteID int64) error {
	return s.db.DeleteNote(ctx, noteID)
}

// ListNotes returns the notes matching the supplied filter along with the total
// number of matching rows prior to pagination.
func (s *Service) ListNotes(ctx context.Context, filter NoteFilter) ([]Note, int, error) {
	if filter.Limit <= 0 {
		filter.Limit = DefaultNoteListLimit
	}

	if filter.Limit > MaxNoteListLimit {
		filter.Limit = MaxNoteListLimit
	}

	if filter.Skip < 0 {
		filter.Skip = 0
	}

	if filter.Sort == "" {
		filter.Sort = SortUpdatedAtDesc
	}

	if !validNoteSorts[filter.Sort] {
		return nil, 0, ErrInvalidSort
	}

	return s.db.ListNotes(ctx, filter)
}

// ListTags returns the distinct tags in use across live notes, most used first.
func (s *Service) ListTags(ctx context.Context) ([]TagCount, error) {
	return s.db.ListTags(ctx)
}

// CreateAttachment validates and persists an uploaded image attachment.
func (s *Service) CreateAttachment(ctx context.Context, attachment Attachment) (Attachment, error) {
	if len(attachment.Data) == 0 {
		return Attachment{}, ErrAttachmentEmpty
	}

	if len(attachment.Data) > MaxAttachmentSize {
		return Attachment{}, ErrAttachmentTooLarge
	}

	if !validAttachmentContentTypes[attachment.ContentType] {
		return Attachment{}, ErrAttachmentTypeUnsupported
	}

	if attachment.Token == "" {
		attachmentToken, err := uuid.NewV4()
		if err != nil {
			return Attachment{}, err
		}
		attachment.Token = attachmentToken.String()
	}

	return s.db.CreateAttachment(ctx, attachment)
}

// GetAttachment returns the attachment for the supplied id or ErrNotFound.
func (s *Service) GetAttachment(ctx context.Context, attachmentID int64) (Attachment, error) {
	return s.db.GetAttachment(ctx, attachmentID)
}

// GetAttachmentByToken returns the attachment for the supplied unguessable
// token or ErrNotFound. Token look back the public media endpoint.
func (s *Service) GetAttachmentByToken(ctx context.Context, token string) (Attachment, error) {
	return s.db.GetAttachmentByToken(ctx, token)
}
