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
package handlers

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/specterops/bloodhound/server/notes/internal/services"
)

type NoteView struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	UserID    string    `json:"user_id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	Type      string    `json:"type"`
	Tags      []string  `json:"tags"`
	URL       string    `json:"url"`
	ObjectID  string    `json:"object_id"`
	EdgeKind  string    `json:"edge_kind"`
}

func BuildNoteView(note services.Note) NoteView {
	return NoteView{
		ID:        note.ID,
		CreatedAt: note.CreatedAt,
		UpdatedAt: note.UpdatedAt,
		UserID:    note.UserID,
		Title:     note.Title,
		Content:   note.Content,
		Type:      note.Type,
		Tags:      note.Tags,
		URL:       note.URL,
		ObjectID:  note.ObjectID,
		EdgeKind:  note.EdgeKind,
	}
}

func (s NoteView) JSONView() ([]byte, error) { return json.Marshal(s) }

type NotesView []NoteView

func BuildNotesView(notes []services.Note) NotesView {
	views := make(NotesView, 0, len(notes))
	for _, note := range notes {
		views = append(views, BuildNoteView(note))
	}
	return views
}

func (s NotesView) JSONView() ([]byte, error) { return json.Marshal(s) }

// CreateNoteRequest is the request body accepted by the create note endpoint.
type CreateNoteRequest struct {
	Title    string   `json:"title"`
	Content  string   `json:"content"`
	Type     string   `json:"type"`
	Tags     []string `json:"tags"`
	URL      string   `json:"url"`
	ObjectID string   `json:"object_id"`
	EdgeKind string   `json:"edge_kind"`
}

func (s CreateNoteRequest) ToNote(userID string) services.Note {
	return services.Note{
		UserID:   userID,
		Title:    s.Title,
		Content:  s.Content,
		Type:     s.Type,
		Tags:     s.Tags,
		URL:      s.URL,
		ObjectID: s.ObjectID,
		EdgeKind: s.EdgeKind,
	}
}

// UpdateNoteRequest is the request body accepted by the update note endpoint.
type UpdateNoteRequest struct {
	Title    string   `json:"title"`
	Content  string   `json:"content"`
	Type     string   `json:"type"`
	Tags     []string `json:"tags"`
	URL      string   `json:"url"`
	ObjectID string   `json:"object_id"`
	EdgeKind string   `json:"edge_kind"`
}

func (s UpdateNoteRequest) ToNote(noteID int64) services.Note {
	return services.Note{
		ID:       noteID,
		Title:    s.Title,
		Content:  s.Content,
		Type:     s.Type,
		Tags:     s.Tags,
		URL:      s.URL,
		ObjectID: s.ObjectID,
		EdgeKind: s.EdgeKind,
	}
}

type TagCountView struct {
	Tag   string `json:"tag"`
	Count int    `json:"count"`
}

type TagCountsView []TagCountView

func BuildTagCountsView(tags []services.TagCount) TagCountsView {
	views := make(TagCountsView, 0, len(tags))
	for _, tag := range tags {
		views = append(views, TagCountView{Tag: tag.Tag, Count: tag.Count})
	}
	return views
}

func (s TagCountsView) JSONView() ([]byte, error) { return json.Marshal(s) }

// AttachmentView carries the serving URL and a ready-to-paste markdown image
// reference for the uploaded attachment.
type AttachmentView struct {
	ID          int64  `json:"id"`
	Filename    string `json:"filename"`
	ContentType string `json:"content_type"`
	Token       string `json:"token"`
	URL         string `json:"url"`
	Markdown    string `json:"markdown"`
}

func BuildAttachmentView(attachment services.Attachment) AttachmentView {
	var (
		url      = fmt.Sprintf("/api/v2/red-team-notes/media/%s", attachment.Token)
		altText  = attachment.Filename
		markdown = fmt.Sprintf("![%s](%s)", altText, url)
	)

	return AttachmentView{
		ID:          attachment.ID,
		Filename:    attachment.Filename,
		ContentType: attachment.ContentType,
		Token:       attachment.Token,
		URL:         url,
		Markdown:    markdown,
	}
}

func (s AttachmentView) JSONView() ([]byte, error) { return json.Marshal(s) }
