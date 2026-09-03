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

//go:build integration

package appdb_test

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/peterldowns/pgtestdb"
	"github.com/specterops/bloodhound/cmd/api/src/auth"
	"github.com/specterops/bloodhound/cmd/api/src/config"
	"github.com/specterops/bloodhound/cmd/api/src/database"
	"github.com/specterops/bloodhound/cmd/api/src/test/integration/utils"
	"github.com/specterops/bloodhound/server/notes/internal/appdb"
	"github.com/specterops/bloodhound/server/notes/internal/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupStore(t *testing.T) (*appdb.Store, *pgxpool.Pool) {
	t.Helper()

	var (
		ctx      = context.Background()
		connConf = pgtestdb.Custom(t, getPostgresConfig(t), pgtestdb.NoopMigrator{})
	)

	cfg, err := config.NewDefaultConnectionConfiguration(connConf.URL())
	require.NoError(t, err)
	cfg.Database.Connection = connConf.URL()

	gormDB, dbPool, err := database.OpenDatabase(cfg.Database)
	require.NoError(t, err)

	db := database.NewBloodhoundDB(gormDB, dbPool, auth.NewIdentityResolver(), cfg)

	err = db.Migrate(ctx)
	require.NoError(t, err)

	t.Cleanup(func() { db.Close(ctx) })

	return appdb.NewStore(db.Pool()), db.Pool()
}

func getPostgresConfig(t *testing.T) pgtestdb.Config {
	t.Helper()

	cfg, err := utils.LoadIntegrationTestConfig()
	require.NoError(t, err)

	environmentMap := make(map[string]string)
	for _, entry := range strings.Fields(cfg.Database.Connection) {
		if parts := strings.SplitN(entry, "=", 2); len(parts) == 2 {
			environmentMap[parts[0]] = parts[1]
		}
	}

	if strings.HasPrefix(environmentMap["host"], "/") {
		return pgtestdb.Config{
			DriverName: "pgx",
			User:       environmentMap["user"],
			Password:   environmentMap["password"],
			Database:   environmentMap["dbname"],
			Options:    fmt.Sprintf("host=%s", url.PathEscape(environmentMap["host"])),
			TestRole: &pgtestdb.Role{
				Username:     environmentMap["user"],
				Password:     environmentMap["password"],
				Capabilities: "NOSUPERUSER NOCREATEROLE",
			},
		}
	}

	return pgtestdb.Config{
		DriverName:                "pgx",
		Host:                      environmentMap["host"],
		Port:                      environmentMap["port"],
		User:                      environmentMap["user"],
		Password:                  environmentMap["password"],
		Database:                  environmentMap["dbname"],
		Options:                   "sslmode=disable",
		ForceTerminateConnections: true,
	}
}

func TestStore_CreateNote_Integration(t *testing.T) {
	var (
		store, _ = setupStore(t)
		ctx      = context.Background()
	)

	created, err := store.CreateNote(ctx, services.Note{
		UserID:   "integration-user",
		Title:    "Kerberoasting reference",
		Content:  "Request TGS for SPN-enabled accounts, crack offline.",
		Type:     services.NoteTypeTechnique,
		Tags:     []string{"kerberos", "ad"},
		URL:      "https://example.com/impacket",
		ObjectID: "S-1-5-21-123",
		EdgeKind: "",
	})
	require.NoError(t, err)

	assert.NotZero(t, created.ID)
	assert.Equal(t, "Kerberoasting reference", created.Title)
	assert.Equal(t, services.NoteTypeTechnique, created.Type)
	assert.Equal(t, []string{"kerberos", "ad"}, created.Tags)
	assert.Equal(t, "S-1-5-21-123", created.ObjectID)
	assert.False(t, created.CreatedAt.IsZero())
	assert.False(t, created.DeletedAt.Valid)

	fetched, err := store.GetNote(ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, created.Title, fetched.Title)
	assert.Equal(t, created.Tags, fetched.Tags)
}

func TestStore_GetNote_Integration(t *testing.T) {
	var (
		store, _ = setupStore(t)
		ctx      = context.Background()
	)

	_, err := store.GetNote(ctx, 999999)
	assert.ErrorIs(t, err, services.ErrNotFound)
}

func TestStore_UpdateNote_Integration(t *testing.T) {
	var (
		store, _ = setupStore(t)
		ctx      = context.Background()
	)

	created, err := store.CreateNote(ctx, services.Note{
		Title:   "Original title",
		Content: "Original content",
		Type:    services.NoteTypeGeneral,
		Tags:    []string{},
	})
	require.NoError(t, err)

	created.Title = "Updated title"
	created.Content = "Updated content"
	created.Type = services.NoteTypeTool
	created.Tags = []string{"updated"}
	created.URL = "https://example.com/tool"
	created.ObjectID = "S-1-5-21-456"
	created.EdgeKind = "CoerceAndRelayNTLMToSMB"

	updated, err := store.UpdateNote(ctx, created)
	require.NoError(t, err)
	assert.Equal(t, "Updated title", updated.Title)
	assert.Equal(t, services.NoteTypeTool, updated.Type)
	assert.Equal(t, []string{"updated"}, updated.Tags)
	assert.Equal(t, "S-1-5-21-456", updated.ObjectID)
	assert.Equal(t, "CoerceAndRelayNTLMToSMB", updated.EdgeKind)
	assert.True(t, updated.UpdatedAt.After(created.CreatedAt) || updated.UpdatedAt.Equal(created.CreatedAt))

	_, err = store.UpdateNote(ctx, services.Note{ID: 999999, Title: "ghost", Type: services.NoteTypeGeneral})
	assert.ErrorIs(t, err, services.ErrNotFound)
}

func TestStore_DeleteNote_Integration(t *testing.T) {
	var (
		store, _ = setupStore(t)
		ctx      = context.Background()
	)

	created, err := store.CreateNote(ctx, services.Note{
		Title: "To be deleted",
		Type:  services.NoteTypeGeneral,
		Tags:  []string{},
	})
	require.NoError(t, err)

	require.NoError(t, store.DeleteNote(ctx, created.ID))

	_, err = store.GetNote(ctx, created.ID)
	assert.ErrorIs(t, err, services.ErrNotFound)

	assert.ErrorIs(t, store.DeleteNote(ctx, created.ID), services.ErrNotFound)
}

func TestStore_ListNotes_Integration(t *testing.T) {
	var (
		store, _ = setupStore(t)
		ctx      = context.Background()
	)

	techniqueNote, err := store.CreateNote(ctx, services.Note{
		Title:    "DCSync walkthrough",
		Type:     services.NoteTypeTechnique,
		Tags:     []string{"ad", "replication"},
		ObjectID: "S-1-5-21-111",
	})
	require.NoError(t, err)

	toolNote, err := store.CreateNote(ctx, services.Note{
		Title:    "Impacket collection",
		Type:     services.NoteTypeTool,
		Tags:     []string{"python"},
		URL:      "https://github.com/fortra/impacket",
		ObjectID: "S-1-5-21-111",
		EdgeKind: "DCSync",
	})
	require.NoError(t, err)

	unrelatedNote, err := store.CreateNote(ctx, services.Note{
		Title:    "Unrelated note",
		Type:     services.NoteTypeGeneral,
		Tags:     []string{"misc"},
		ObjectID: "S-1-5-21-999",
	})
	require.NoError(t, err)

	t.Run("lists all notes without filters", func(t *testing.T) {
		notes, count, err := store.ListNotes(ctx, services.NoteFilter{Limit: 50})
		require.NoError(t, err)
		assert.Equal(t, 3, count)
		assert.Len(t, notes, 3)
	})

	t.Run("filters by object id", func(t *testing.T) {
		notes, count, err := store.ListNotes(ctx, services.NoteFilter{ObjectID: "S-1-5-21-111", Limit: 50})
		require.NoError(t, err)
		assert.Equal(t, 2, count)
		assert.Len(t, notes, 2)
	})

	t.Run("filters by edge kind", func(t *testing.T) {
		notes, count, err := store.ListNotes(ctx, services.NoteFilter{EdgeKind: "DCSync", Limit: 50})
		require.NoError(t, err)
		assert.Equal(t, 1, count)
		require.Len(t, notes, 1)
		assert.Equal(t, toolNote.ID, notes[0].ID)
	})

	t.Run("filters by type", func(t *testing.T) {
		notes, count, err := store.ListNotes(ctx, services.NoteFilter{Type: services.NoteTypeTechnique, Limit: 50})
		require.NoError(t, err)
		assert.Equal(t, 1, count)
		require.Len(t, notes, 1)
		assert.Equal(t, techniqueNote.ID, notes[0].ID)
	})

	t.Run("filters by tag containment", func(t *testing.T) {
		notes, count, err := store.ListNotes(ctx, services.NoteFilter{Tags: []string{"ad"}, Limit: 50})
		require.NoError(t, err)
		assert.Equal(t, 1, count)
		require.Len(t, notes, 1)
		assert.Equal(t, techniqueNote.ID, notes[0].ID)
	})

	t.Run("filters by multiple tags with AND semantics", func(t *testing.T) {
		notes, count, err := store.ListNotes(ctx, services.NoteFilter{Tags: []string{"ad", "replication"}, Limit: 50})
		require.NoError(t, err)
		assert.Equal(t, 1, count)
		require.Len(t, notes, 1)
		assert.Equal(t, techniqueNote.ID, notes[0].ID)

		notes, count, err = store.ListNotes(ctx, services.NoteFilter{Tags: []string{"ad", "python"}, Limit: 50})
		require.NoError(t, err)
		assert.Equal(t, 0, count)
		assert.Empty(t, notes)
	})

	t.Run("filters by search pattern across title and content", func(t *testing.T) {
		notes, count, err := store.ListNotes(ctx, services.NoteFilter{Search: "impacket", Limit: 50})
		require.NoError(t, err)
		assert.Equal(t, 1, count)
		require.Len(t, notes, 1)
		assert.Equal(t, toolNote.ID, notes[0].ID)
	})

	t.Run("paginates results", func(t *testing.T) {
		notes, count, err := store.ListNotes(ctx, services.NoteFilter{Limit: 2, Skip: 0})
		require.NoError(t, err)
		assert.Equal(t, 3, count)
		assert.Len(t, notes, 2)

		notes, count, err = store.ListNotes(ctx, services.NoteFilter{Limit: 2, Skip: 2})
		require.NoError(t, err)
		assert.Equal(t, 3, count)
		assert.Len(t, notes, 1)
	})

	t.Run("excludes soft-deleted notes", func(t *testing.T) {
		require.NoError(t, store.DeleteNote(ctx, unrelatedNote.ID))

		notes, count, err := store.ListNotes(ctx, services.NoteFilter{Limit: 50})
		require.NoError(t, err)
		assert.Equal(t, 2, count)
		assert.Len(t, notes, 2)
	})
}

func TestStore_Tags_Integration(t *testing.T) {
	var (
		store, _ = setupStore(t)
		ctx      = context.Background()
	)

	_, err := store.CreateNote(ctx, services.Note{Title: "a", Type: services.NoteTypeGeneral, Tags: []string{"ad", "kerberos"}})
	require.NoError(t, err)
	_, err = store.CreateNote(ctx, services.Note{Title: "b", Type: services.NoteTypeGeneral, Tags: []string{"ad"}})
	require.NoError(t, err)
	deletedNote, err := store.CreateNote(ctx, services.Note{Title: "c", Type: services.NoteTypeGeneral, Tags: []string{"ghost"}})
	require.NoError(t, err)
	require.NoError(t, store.DeleteNote(ctx, deletedNote.ID))

	tags, err := store.ListTags(ctx)
	require.NoError(t, err)

	tagCounts := make(map[string]int)
	for _, tag := range tags {
		tagCounts[tag.Tag] = tag.Count
	}

	assert.Equal(t, 2, tagCounts["ad"])
	assert.Equal(t, 1, tagCounts["kerberos"])
	assert.NotContains(t, tagCounts, "ghost", "tags of soft-deleted notes must not be suggested")
}

func TestStore_Attachments_Integration(t *testing.T) {
	var (
		store, _ = setupStore(t)
		ctx      = context.Background()
		payload  = []byte{0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a}
	)

	created, err := store.CreateAttachment(ctx, services.Attachment{
		Filename:    "poc.png",
		ContentType: "image/png",
		Data:        payload,
	})
	require.NoError(t, err)
	assert.NotZero(t, created.ID)

	fetched, err := store.GetAttachment(ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, "poc.png", fetched.Filename)
	assert.Equal(t, "image/png", fetched.ContentType)
	assert.Equal(t, payload, fetched.Data)

	_, err = store.GetAttachment(ctx, 999999)
	assert.ErrorIs(t, err, services.ErrNotFound)
}
