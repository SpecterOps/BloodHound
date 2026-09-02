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
package services_test

import (
	"context"
	"errors"
	"testing"

	"github.com/specterops/bloodhound/server/notes/internal/services"
	"github.com/specterops/bloodhound/server/notes/internal/services/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newServiceUnderTest(t *testing.T) (*services.Service, *mocks.MockDatabase) {
	t.Helper()

	var (
		database = mocks.NewMockDatabase(t)
		svc      = services.NewService(database)
	)

	return svc, database
}

func TestNewService(t *testing.T) {
	t.Parallel()

	svc, _ := newServiceUnderTest(t)
	assert.NotNil(t, svc)
}

func TestNewService_NilDatabasePanics(t *testing.T) {
	t.Parallel()

	assert.PanicsWithValue(t, "notes: service requires a non-nil Database", func() {
		services.NewService(nil)
	})
}

func TestValidateNote(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		note    services.Note
		wantErr error
	}{
		{
			name: "Success: valid note passes validation",
			note: services.Note{
				Title: "Kerberoasting",
				Type:  services.NoteTypeTechnique,
			},
		},
		{
			name: "Error: empty title fails validation",
			note: services.Note{
				Title: "   ",
				Type:  services.NoteTypeGeneral,
			},
			wantErr: services.ErrTitleRequired,
		},
		{
			name: "Error: invalid type fails validation",
			note: services.Note{
				Title: "Some note",
				Type:  "not-a-real-type",
			},
			wantErr: services.ErrInvalidType,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			err := services.ValidateNote(test.note)
			if test.wantErr != nil {
				assert.ErrorIs(t, err, test.wantErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestService_CreateNote(t *testing.T) {
	t.Parallel()

	var ctx = context.Background()

	tests := []struct {
		name       string
		note       services.Note
		setupMocks func(m *mocks.MockDatabase, note services.Note)
		wantErr    error
	}{
		{
			name: "Success: defaults type and tags before persisting",
			note: services.Note{
				Title:  "Impacket secretsdump",
				UserID: "user-1",
			},
			setupMocks: func(m *mocks.MockDatabase, note services.Note) {
				expectedNote := note
				expectedNote.Type = services.NoteTypeGeneral
				expectedNote.Tags = []string{}
				m.EXPECT().CreateNote(ctx, expectedNote).Return(expectedNote, nil)
			},
		},
		{
			name: "Error: validation failure does not reach the database",
			note: services.Note{
				Title: "",
			},
			setupMocks: func(m *mocks.MockDatabase, note services.Note) {},
			wantErr:    services.ErrTitleRequired,
		},
		{
			name: "Error: invalid note type does not reach the database",
			note: services.Note{
				Title: "Some note",
				Type:  "invalid",
			},
			setupMocks: func(m *mocks.MockDatabase, note services.Note) {},
			wantErr:    services.ErrInvalidType,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			svc, database := newServiceUnderTest(t)
			test.setupMocks(database, test.note)

			got, err := svc.CreateNote(ctx, test.note)
			if test.wantErr != nil {
				assert.ErrorIs(t, err, test.wantErr)
				assert.Equal(t, services.Note{}, got)
			} else {
				require.NoError(t, err)
				assert.Equal(t, test.note.Title, got.Title)
			}
		})
	}
}

func TestService_UpdateNote(t *testing.T) {
	t.Parallel()

	var (
		ctx       = context.Background()
		validNote = services.Note{
			ID:    42,
			Title: "Updated title",
			Type:  services.NoteTypeTool,
		}
	)

	tests := []struct {
		name       string
		note       services.Note
		setupMocks func(m *mocks.MockDatabase)
		wantErr    error
	}{
		{
			name: "Success: persists the updated note",
			note: validNote,
			setupMocks: func(m *mocks.MockDatabase) {
				expectedNote := validNote
				expectedNote.Tags = []string{}
				m.EXPECT().UpdateNote(ctx, expectedNote).Return(expectedNote, nil)
			},
		},
		{
			name:       "Error: empty title fails validation",
			note:       services.Note{ID: 42, Title: ""},
			setupMocks: func(m *mocks.MockDatabase) {},
			wantErr:    services.ErrTitleRequired,
		},
		{
			name: "Error: propagates database errors",
			note: validNote,
			setupMocks: func(m *mocks.MockDatabase) {
				expectedNote := validNote
				expectedNote.Tags = []string{}
				m.EXPECT().UpdateNote(ctx, expectedNote).Return(services.Note{}, services.ErrNotFound)
			},
			wantErr: services.ErrNotFound,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			svc, database := newServiceUnderTest(t)
			test.setupMocks(database)

			got, err := svc.UpdateNote(ctx, test.note)
			if test.wantErr != nil {
				assert.ErrorIs(t, err, test.wantErr)
			} else {
				require.NoError(t, err)
				assert.Equal(t, test.note.ID, got.ID)
			}
		})
	}
}

func TestService_ListNotes(t *testing.T) {
	t.Parallel()

	var (
		ctx           = context.Background()
		databaseError = errors.New("connection refused")
		expectedNotes = []services.Note{{ID: 1, Title: "note"}}
	)

	tests := []struct {
		name        string
		filter      services.NoteFilter
		expectLimit int
		setupMocks  func(m *mocks.MockDatabase, filter services.NoteFilter)
		wantErr     error
	}{
		{
			name:        "Success: applies default limit",
			filter:      services.NoteFilter{},
			expectLimit: services.DefaultNoteListLimit,
			setupMocks: func(m *mocks.MockDatabase, filter services.NoteFilter) {
				m.EXPECT().ListNotes(ctx, filter).Return(expectedNotes, 1, nil)
			},
		},
		{
			name:        "Success: clamps limit to the maximum",
			filter:      services.NoteFilter{Limit: 100000},
			expectLimit: services.MaxNoteListLimit,
			setupMocks: func(m *mocks.MockDatabase, filter services.NoteFilter) {
				m.EXPECT().ListNotes(ctx, filter).Return(expectedNotes, 1, nil)
			},
		},
		{
			name:        "Success: resets negative skip",
			filter:      services.NoteFilter{Skip: -5, Limit: 10},
			expectLimit: 10,
			setupMocks: func(m *mocks.MockDatabase, filter services.NoteFilter) {
				m.EXPECT().ListNotes(ctx, filter).Return(expectedNotes, 1, nil)
			},
		},
		{
			name:        "Error: propagates database errors",
			filter:      services.NoteFilter{},
			expectLimit: services.DefaultNoteListLimit,
			setupMocks: func(m *mocks.MockDatabase, filter services.NoteFilter) {
				m.EXPECT().ListNotes(ctx, filter).Return(nil, 0, databaseError)
			},
			wantErr: databaseError,
		},
		{
			name:        "Error: rejects invalid sort values",
			filter:      services.NoteFilter{Sort: "not-a-sort"},
			expectLimit: services.DefaultNoteListLimit,
			setupMocks:  func(m *mocks.MockDatabase, filter services.NoteFilter) {},
			wantErr:     services.ErrInvalidSort,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			svc, database := newServiceUnderTest(t)

			expectedFilter := test.filter
			if expectedFilter.Limit <= 0 {
				expectedFilter.Limit = services.DefaultNoteListLimit
			}
			if expectedFilter.Limit > services.MaxNoteListLimit {
				expectedFilter.Limit = services.MaxNoteListLimit
			}
			if expectedFilter.Skip < 0 {
				expectedFilter.Skip = 0
			}
			if expectedFilter.Sort == "" {
				expectedFilter.Sort = services.SortUpdatedAtDesc
			}

			test.setupMocks(database, expectedFilter)

			notes, count, err := svc.ListNotes(ctx, test.filter)
			if test.wantErr != nil {
				assert.ErrorIs(t, err, test.wantErr)
			} else {
				require.NoError(t, err)
				assert.Equal(t, expectedNotes, notes)
				assert.Equal(t, 1, count)
			}
		})
	}
}

func TestService_GetNote_DeleteNote(t *testing.T) {
	t.Parallel()

	var (
		ctx          = context.Background()
		expectedNote = services.Note{ID: 7, Title: "note"}
	)

	t.Run("GetNote delegates to the database", func(t *testing.T) {
		t.Parallel()

		svc, database := newServiceUnderTest(t)
		database.EXPECT().GetNote(ctx, int64(7)).Return(expectedNote, nil)

		got, err := svc.GetNote(ctx, 7)
		require.NoError(t, err)
		assert.Equal(t, expectedNote, got)
	})

	t.Run("DeleteNote delegates to the database", func(t *testing.T) {
		t.Parallel()

		svc, database := newServiceUnderTest(t)
		database.EXPECT().DeleteNote(ctx, int64(7)).Return(nil)

		assert.NoError(t, svc.DeleteNote(ctx, 7))
	})
}
