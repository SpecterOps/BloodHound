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

package notes_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gofrs/uuid"
	"github.com/gorilla/mux"
	"github.com/peterldowns/pgtestdb"
	"github.com/specterops/bloodhound/cmd/api/src/auth"
	"github.com/specterops/bloodhound/cmd/api/src/bhctx"
	"github.com/specterops/bloodhound/cmd/api/src/config"
	"github.com/specterops/bloodhound/cmd/api/src/database"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/cmd/api/src/test/integration/utils"
	"github.com/specterops/bloodhound/server/notes/internal/appdb"
	"github.com/specterops/bloodhound/server/notes/internal/handlers"
	"github.com/specterops/bloodhound/server/notes/internal/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// injectAuthMiddleware wraps the given handler, attaching a bhctx.Context that
// identifies the supplied user as the request owner. This stands in for the
// production auth middleware so notes handlers can resolve a user without
// requiring the full auth stack.
func injectAuthMiddleware(handler http.HandlerFunc, userID uuid.UUID) http.HandlerFunc {
	return func(response http.ResponseWriter, request *http.Request) {
		var bhCtx = &bhctx.Context{
			AuthCtx: auth.Context{Owner: model.User{Unique: model.Unique{ID: userID}}},
		}
		handler(response, bhctx.SetRequestContext(request, bhCtx))
	}
}

// setupNotesRouter creates an isolated test database with all migrations
// applied, wires the notes handler stack and registers the production route
// patterns on a gorilla mux router.
func setupNotesRouter(t *testing.T) (*mux.Router, *database.BloodhoundDB) {
	t.Helper()

	var (
		ctx      = context.Background()
		connConf = pgtestdb.Custom(t, getNotesPostgresConfig(t), pgtestdb.NoopMigrator{})
	)

	cfg, err := config.NewDefaultConnectionConfiguration(connConf.URL())
	require.NoError(t, err)

	gormDB, dbPool, err := database.OpenDatabase(cfg.Database)
	require.NoError(t, err)

	db := database.NewBloodhoundDB(gormDB, dbPool, auth.NewIdentityResolver(), cfg)
	require.NoError(t, db.Migrate(ctx))

	t.Cleanup(func() { db.Close(ctx) })

	var (
		store      = appdb.NewStore(db.Pool())
		svc        = services.NewService(store)
		handlerSet = handlers.NewHandlersContainer(svc)
		routerInst = mux.NewRouter()
	)

	routerInst.HandleFunc("/api/v2/red-team-notes", handlerSet.ListNotes).Methods(http.MethodGet)
	routerInst.HandleFunc("/api/v2/red-team-notes/tags", handlerSet.ListTags).Methods(http.MethodGet)
	routerInst.HandleFunc("/api/v2/red-team-notes/attachments", injectAuthMiddleware(handlerSet.UploadAttachment, testUserID(t))).Methods(http.MethodPost)
	routerInst.HandleFunc("/api/v2/red-team-notes/attachments/{attachment_id}", handlerSet.GetAttachment).Methods(http.MethodGet)
	routerInst.HandleFunc("/api/v2/red-team-notes/media/{attachment_token}", handlerSet.GetMedia).Methods(http.MethodGet)
	routerInst.HandleFunc("/api/v2/red-team-notes", injectAuthMiddleware(handlerSet.CreateNote, testUserID(t))).Methods(http.MethodPost)
	routerInst.HandleFunc("/api/v2/red-team-notes/{note_id}", handlerSet.GetNote).Methods(http.MethodGet)
	routerInst.HandleFunc("/api/v2/red-team-notes/{note_id}", handlerSet.UpdateNote).Methods(http.MethodPut)
	routerInst.HandleFunc("/api/v2/red-team-notes/{note_id}", handlerSet.DeleteNote).Methods(http.MethodDelete)

	return routerInst, db
}

func testUserID(t *testing.T) uuid.UUID {
	t.Helper()

	userID, err := uuid.NewV4()
	require.NoError(t, err)

	return userID
}

// getNotesPostgresConfig reads the integration test configuration from the
// environment and returns a pgtestdb.Config for the notes e2e tests.
func getNotesPostgresConfig(t *testing.T) pgtestdb.Config {
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

type noteEnvelope struct {
	Data handlers.NoteView `json:"data"`
}

type notesEnvelope struct {
	Data  handlers.NotesView `json:"data"`
	Count int                `json:"count"`
}

func doRequest(t *testing.T, router *mux.Router, method, path string, body any) *httptest.ResponseRecorder {
	t.Helper()

	var reader *bytes.Reader

	if body != nil {
		rawBody, err := json.Marshal(body)
		require.NoError(t, err)
		reader = bytes.NewReader(rawBody)
	} else {
		reader = bytes.NewReader(nil)
	}

	request := httptest.NewRequest(method, path, reader)
	request.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)

	return recorder
}

func TestNotes_CRUD_Integration(t *testing.T) {
	var (
		router, _  = setupNotesRouter(t)
		createBody = handlers.CreateNoteRequest{
			Title:    "NTLM relaying notes",
			Content:  "Coerce authentication via printerbug, relay to LDAP/SMB.",
			Type:     "technique",
			Tags:     []string{"ntlm", "relay"},
			URL:      "https://github.com/fortra/impacket",
			ObjectID: "S-1-5-21-12345",
			EdgeKind: "CoerceAndRelayNTLMToSMB",
		}
	)

	recorder := doRequest(t, router, http.MethodPost, "/api/v2/red-team-notes", createBody)
	require.Equal(t, http.StatusCreated, recorder.Code, recorder.Body.String())

	var created noteEnvelope
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &created))
	assert.NotZero(t, created.Data.ID)
	assert.Equal(t, createBody.Title, created.Data.Title)
	assert.Equal(t, "technique", created.Data.Type)
	assert.Equal(t, []string{"ntlm", "relay"}, created.Data.Tags)
	assert.NotEmpty(t, created.Data.UserID, "creating user should be recorded")

	recorder = doRequest(t, router, http.MethodGet, fmt.Sprintf("/api/v2/red-team-notes/%d", created.Data.ID), nil)
	require.Equal(t, http.StatusOK, recorder.Code)

	var fetched noteEnvelope
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &fetched))
	assert.Equal(t, createBody.Title, fetched.Data.Title)

	updateBody := handlers.UpdateNoteRequest{
		Title:    "NTLM relaying notes (updated)",
		Content:  "Updated content",
		Type:     "tool",
		Tags:     []string{"impacket"},
		URL:      "https://github.com/fortra/impacket",
		ObjectID: "S-1-5-21-12345",
		EdgeKind: "CoerceAndRelayNTLMToSMB",
	}

	recorder = doRequest(t, router, http.MethodPut, fmt.Sprintf("/api/v2/red-team-notes/%d", created.Data.ID), updateBody)
	require.Equal(t, http.StatusOK, recorder.Code, recorder.Body.String())

	var updated noteEnvelope
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &updated))
	assert.Equal(t, "NTLM relaying notes (updated)", updated.Data.Title)
	assert.Equal(t, "tool", updated.Data.Type)
	assert.Equal(t, []string{"impacket"}, updated.Data.Tags)

	recorder = doRequest(t, router, http.MethodDelete, fmt.Sprintf("/api/v2/red-team-notes/%d", created.Data.ID), nil)
	require.Equal(t, http.StatusNoContent, recorder.Code)

	recorder = doRequest(t, router, http.MethodGet, fmt.Sprintf("/api/v2/red-team-notes/%d", created.Data.ID), nil)
	assert.Equal(t, http.StatusNotFound, recorder.Code)
}

func TestNotes_ListAndFilter_Integration(t *testing.T) {
	var (
		router, _  = setupNotesRouter(t)
		noteBodies = []handlers.CreateNoteRequest{
			{Title: "BloodHound sharp hound collector", Type: "tool", Tags: []string{"recon", "ad"}, ObjectID: "S-1-5-21-1"},
			{Title: "DCSync technique", Type: "technique", Tags: []string{"ad"}, ObjectID: "S-1-5-21-1", EdgeKind: "DCSync"},
			{Title: "Private repo notes", Type: "source", Tags: []string{"misc"}, ObjectID: "S-1-5-21-2"},
		}
	)

	for _, body := range noteBodies {
		recorder := doRequest(t, router, http.MethodPost, "/api/v2/red-team-notes", body)
		require.Equal(t, http.StatusCreated, recorder.Code, recorder.Body.String())
	}

	recorder := doRequest(t, router, http.MethodGet, "/api/v2/red-team-notes", nil)
	require.Equal(t, http.StatusOK, recorder.Code)

	var allNotes notesEnvelope
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &allNotes))
	assert.Equal(t, 3, allNotes.Count)
	assert.Len(t, allNotes.Data, 3)

	recorder = doRequest(t, router, http.MethodGet, "/api/v2/red-team-notes?object_id=S-1-5-21-1", nil)
	require.Equal(t, http.StatusOK, recorder.Code)

	var objectNotes notesEnvelope
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &objectNotes))
	assert.Equal(t, 2, objectNotes.Count)

	recorder = doRequest(t, router, http.MethodGet, "/api/v2/red-team-notes?type=technique", nil)
	require.Equal(t, http.StatusOK, recorder.Code)

	var techniqueNotes notesEnvelope
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &techniqueNotes))
	assert.Equal(t, 1, techniqueNotes.Count)
	assert.Equal(t, "DCSync technique", techniqueNotes.Data[0].Title)

	recorder = doRequest(t, router, http.MethodGet, "/api/v2/red-team-notes?tag=recon", nil)
	require.Equal(t, http.StatusOK, recorder.Code)

	var taggedNotes notesEnvelope
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &taggedNotes))
	assert.Equal(t, 1, taggedNotes.Count)

	recorder = doRequest(t, router, http.MethodGet, "/api/v2/red-team-notes?tag=recon&tag=ad", nil)
	require.Equal(t, http.StatusOK, recorder.Code)

	var multiTaggedNotes notesEnvelope
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &multiTaggedNotes))
	assert.Equal(t, 1, multiTaggedNotes.Count)
	assert.Equal(t, "BloodHound sharp hound collector", multiTaggedNotes.Data[0].Title)

	recorder = doRequest(t, router, http.MethodGet, "/api/v2/red-team-notes?tag=recon&tag=misc", nil)
	require.Equal(t, http.StatusOK, recorder.Code)

	var noMatchNotes notesEnvelope
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &noMatchNotes))
	assert.Equal(t, 0, noMatchNotes.Count)

	recorder = doRequest(t, router, http.MethodGet, "/api/v2/red-team-notes?search="+url.QueryEscape("dcsync"), nil)
	require.Equal(t, http.StatusOK, recorder.Code)

	var searchNotes notesEnvelope
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &searchNotes))
	assert.Equal(t, 1, searchNotes.Count)

	recorder = doRequest(t, router, http.MethodGet, "/api/v2/red-team-notes?limit=2&skip=0", nil)
	require.Equal(t, http.StatusOK, recorder.Code)

	var paginatedNotes notesEnvelope
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &paginatedNotes))
	assert.Equal(t, 3, paginatedNotes.Count)
	assert.Len(t, paginatedNotes.Data, 2)

	recorder = doRequest(t, router, http.MethodGet, "/api/v2/red-team-notes?sort=title", nil)
	require.Equal(t, http.StatusOK, recorder.Code)

	var sortedNotes notesEnvelope
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &sortedNotes))
	require.Len(t, sortedNotes.Data, 3)
	assert.Equal(t, "BloodHound sharp hound collector", sortedNotes.Data[0].Title)
	assert.Equal(t, "DCSync technique", sortedNotes.Data[1].Title)
	assert.Equal(t, "Private repo notes", sortedNotes.Data[2].Title)

	recorder = doRequest(t, router, http.MethodGet, "/api/v2/red-team-notes?sort=not-a-sort", nil)
	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestNotes_ValidationErrors_Integration(t *testing.T) {
	var router, _ = setupNotesRouter(t)

	recorder := doRequest(t, router, http.MethodPost, "/api/v2/red-team-notes", handlers.CreateNoteRequest{Title: "", Type: "general"})
	assert.Equal(t, http.StatusBadRequest, recorder.Code)

	recorder = doRequest(t, router, http.MethodPost, "/api/v2/red-team-notes", handlers.CreateNoteRequest{Title: "Valid title", Type: "not-a-type"})
	assert.Equal(t, http.StatusBadRequest, recorder.Code)

	recorder = doRequest(t, router, http.MethodPost, "/api/v2/red-team-notes", "{malformed json")
	assert.Equal(t, http.StatusBadRequest, recorder.Code)

	recorder = doRequest(t, router, http.MethodGet, "/api/v2/red-team-notes/not-a-number", nil)
	assert.Equal(t, http.StatusBadRequest, recorder.Code)

	recorder = doRequest(t, router, http.MethodGet, "/api/v2/red-team-notes/999999", nil)
	assert.Equal(t, http.StatusNotFound, recorder.Code)
}

type tagsEnvelope struct {
	Data []handlers.TagCountView `json:"data"`
}

type attachmentEnvelope struct {
	Data handlers.AttachmentView `json:"data"`
}

func TestNotes_TagsAndAttachments_Integration(t *testing.T) {
	var router, _ = setupNotesRouter(t)

	createRecorder := doRequest(t, router, http.MethodPost, "/api/v2/red-team-notes", handlers.CreateNoteRequest{
		Title: "Tagged note",
		Type:  "technique",
		Tags:  []string{"adcs"},
	})
	require.Equal(t, http.StatusCreated, createRecorder.Code)

	recorder := doRequest(t, router, http.MethodGet, "/api/v2/red-team-notes/tags", nil)
	require.Equal(t, http.StatusOK, recorder.Code)

	var tags tagsEnvelope
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &tags))
	require.Len(t, tags.Data, 1)
	assert.Equal(t, "adcs", tags.Data[0].Tag)
	assert.Equal(t, 1, tags.Data[0].Count)

	var (
		body       bytes.Buffer
		multipartW = multipart.NewWriter(&body)
		pngPayload = []byte{0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a, 0x00, 0x01}
	)

	part, err := multipartW.CreateFormFile("file", "poc.png")
	require.NoError(t, err)
	_, err = part.Write(pngPayload)
	require.NoError(t, err)
	require.NoError(t, multipartW.Close())

	uploadRequest := httptest.NewRequest(http.MethodPost, "/api/v2/red-team-notes/attachments", &body)
	uploadRequest.Header.Set("Content-Type", multipartW.FormDataContentType())

	uploadRecorder := httptest.NewRecorder()
	router.ServeHTTP(uploadRecorder, uploadRequest)
	require.Equal(t, http.StatusCreated, uploadRecorder.Code, uploadRecorder.Body.String())

	var attachment attachmentEnvelope
	require.NoError(t, json.Unmarshal(uploadRecorder.Body.Bytes(), &attachment))
	assert.NotZero(t, attachment.Data.ID)
	assert.NotEmpty(t, attachment.Data.Token)
	assert.Equal(t, "/api/v2/red-team-notes/media/"+attachment.Data.Token, attachment.Data.URL)
	assert.Equal(t, "![poc.png]("+attachment.Data.URL+")", attachment.Data.Markdown)

	serveRequest := httptest.NewRequest(http.MethodGet, attachment.Data.URL, nil)
	serveRecorder := httptest.NewRecorder()
	router.ServeHTTP(serveRecorder, serveRequest)
	require.Equal(t, http.StatusOK, serveRecorder.Code)
	assert.Equal(t, pngPayload, serveRecorder.Body.Bytes())
}
