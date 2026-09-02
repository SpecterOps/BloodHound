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

package handlers

//go:generate go tool mockery

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/specterops/bloodhound/cmd/api/src/api"
	"github.com/specterops/bloodhound/cmd/api/src/auth"
	"github.com/specterops/bloodhound/cmd/api/src/bhctx"
	"github.com/specterops/bloodhound/packages/go/bhlog/attr"
	"github.com/specterops/bloodhound/packages/go/responses"
	"github.com/specterops/bloodhound/server/notes/internal/services"
)

const (
	queryParameterObjectID = "object_id"
	queryParameterEdgeKind = "edge_kind"
	queryParameterType     = "type"
	queryParameterTag      = "tag"
	queryParameterSearch   = "search"
	queryParameterSort     = "sort"
	queryParameterSkip     = "skip"
	queryParameterLimit    = "limit"

	attachmentFormField    = "file"
	attachmentCacheControl = "private, max-age=31536000, immutable"
)

// Notes defines the notes service boundary for the notes handlers package.
type Notes interface {
	CreateNote(ctx context.Context, note services.Note) (services.Note, error)
	GetNote(ctx context.Context, noteID int64) (services.Note, error)
	UpdateNote(ctx context.Context, note services.Note) (services.Note, error)
	DeleteNote(ctx context.Context, noteID int64) error
	ListNotes(ctx context.Context, filter services.NoteFilter) ([]services.Note, int, error)
	ListTags(ctx context.Context) ([]services.TagCount, error)
	CreateAttachment(ctx context.Context, attachment services.Attachment) (services.Attachment, error)
	GetAttachment(ctx context.Context, attachmentID int64) (services.Attachment, error)
	GetAttachmentByToken(ctx context.Context, token string) (services.Attachment, error)
}

// Handlers is a dependency injection container for notes handlers.
type Handlers struct {
	notes Notes
}

// NewHandlersContainer initializes the notes Handlers dependency injection container.
func NewHandlersContainer(notes Notes) *Handlers {
	return &Handlers{
		notes: notes,
	}
}

// ListNotes returns red team notes matching the supplied query parameter filters
// as a paginated JSON response. Supported filters are object_id, edge_kind,
// type, tag and search along with skip/limit pagination.
func (s Handlers) ListNotes(response http.ResponseWriter, request *http.Request) {
	var (
		ctx         = request.Context()
		queryParams = request.URL.Query()
		filter      = services.NoteFilter{
			ObjectID: queryParams.Get(queryParameterObjectID),
			EdgeKind: queryParams.Get(queryParameterEdgeKind),
			Type:     queryParams.Get(queryParameterType),
			Tags:     queryParams[queryParameterTag],
			Search:   queryParams.Get(queryParameterSearch),
			Sort:     queryParams.Get(queryParameterSort),
		}
	)

	skip, err := parseIntQueryParameter(queryParams, queryParameterSkip, 0)
	if err != nil {
		responses.WriteError(ctx, http.StatusBadRequest, err.Error(), response)
		return
	}

	limit, err := parseIntQueryParameter(queryParams, queryParameterLimit, services.DefaultNoteListLimit)
	if err != nil {
		responses.WriteError(ctx, http.StatusBadRequest, err.Error(), response)
		return
	}

	filter.Skip = skip
	filter.Limit = limit

	notes, count, err := s.notes.ListNotes(ctx, filter)
	if err != nil {
		handleNotesError(request, response, err)
		return
	}

	responses.WritePaginated(ctx, BuildNotesView(notes), limit, skip, count, http.StatusOK, response)
}

// GetNote returns a single red team note identified by the note_id path variable.
func (s Handlers) GetNote(response http.ResponseWriter, request *http.Request) {
	var ctx = request.Context()

	noteID, err := parseNoteID(request)
	if err != nil {
		responses.WriteError(ctx, http.StatusBadRequest, api.ErrorResponseDetailsIDMalformed, response)
		return
	}

	note, err := s.notes.GetNote(ctx, noteID)
	if err != nil {
		handleNotesError(request, response, err)
		return
	}

	responses.WriteBasic(ctx, BuildNoteView(note), http.StatusOK, response)
}

// CreateNote persists a new red team note from the JSON request body. The
// creating user is taken from the auth context populated by the route
// middleware; a missing user indicates an unexpected internal state.
func (s Handlers) CreateNote(response http.ResponseWriter, request *http.Request) {
	var (
		ctx               = request.Context()
		createNoteRequest CreateNoteRequest
	)

	if err := json.NewDecoder(request.Body).Decode(&createNoteRequest); err != nil {
		responses.WriteError(ctx, http.StatusBadRequest, api.ErrorResponsePayloadUnmarshalError, response)
		return
	}

	user, isUser := auth.GetUserFromAuthCtx(bhctx.FromRequest(request).AuthCtx)
	if !isUser {
		responses.WriteError(ctx, http.StatusInternalServerError, api.ErrorResponseDetailsInternalServerError, response)
		return
	}

	note, err := s.notes.CreateNote(ctx, createNoteRequest.ToNote(user.ID.String()))
	if err != nil {
		handleNotesError(request, response, err)
		return
	}

	responses.WriteBasic(ctx, BuildNoteView(note), http.StatusCreated, response)
}

// UpdateNote persists changes to an existing red team note identified by the
// note_id path variable using the JSON request body.
func (s Handlers) UpdateNote(response http.ResponseWriter, request *http.Request) {
	var (
		ctx               = request.Context()
		updateNoteRequest UpdateNoteRequest
	)

	noteID, err := parseNoteID(request)
	if err != nil {
		responses.WriteError(ctx, http.StatusBadRequest, api.ErrorResponseDetailsIDMalformed, response)
		return
	}

	if err := json.NewDecoder(request.Body).Decode(&updateNoteRequest); err != nil {
		responses.WriteError(ctx, http.StatusBadRequest, api.ErrorResponsePayloadUnmarshalError, response)
		return
	}

	note, err := s.notes.UpdateNote(ctx, updateNoteRequest.ToNote(noteID))
	if err != nil {
		handleNotesError(request, response, err)
		return
	}

	responses.WriteBasic(ctx, BuildNoteView(note), http.StatusOK, response)
}

// DeleteNote soft-deletes the red team note identified by the note_id path
// variable.
func (s Handlers) DeleteNote(response http.ResponseWriter, request *http.Request) {
	var ctx = request.Context()

	noteID, err := parseNoteID(request)
	if err != nil {
		responses.WriteError(ctx, http.StatusBadRequest, api.ErrorResponseDetailsIDMalformed, response)
		return
	}

	if err := s.notes.DeleteNote(ctx, noteID); err != nil {
		handleNotesError(request, response, err)
		return
	}

	response.WriteHeader(http.StatusNoContent)
}

// parseNoteID extracts and parses the note_id URI path variable.
func parseNoteID(request *http.Request) (int64, error) {
	var rawNoteID = mux.Vars(request)[api.URIPathVariableNoteID]

	return strconv.ParseInt(rawNoteID, 10, 64)
}

// parseIntQueryParameter parses a non-negative integer query parameter or
// returns the supplied default when the parameter is absent.
func parseIntQueryParameter(queryParams map[string][]string, key string, defaultValue int) (int, error) {
	var rawValue = queryParams[key]

	if len(rawValue) == 0 || rawValue[0] == "" {
		return defaultValue, nil
	}

	parsedValue, err := strconv.Atoi(rawValue[0])
	if err != nil {
		return 0, errors.New("query parameter " + key + " must be an integer")
	}

	if parsedValue < 0 {
		return 0, errors.New("query parameter " + key + " must not be negative")
	}

	return parsedValue, nil
}

// handleNotesError maps service-layer errors to HTTP responses, translating
// known sentinel errors to their corresponding status codes and falling back to
// a logged 500 for anything unexpected.
func handleNotesError(request *http.Request, response http.ResponseWriter, err error) {
	if errors.Is(err, services.ErrNotFound) {
		responses.WriteError(request.Context(), http.StatusNotFound, services.ErrNotFound.Error(), response)
	} else if errors.Is(err, services.ErrTitleRequired) || errors.Is(err, services.ErrInvalidType) || errors.Is(err, services.ErrInvalidSort) ||
		errors.Is(err, services.ErrAttachmentTooLarge) || errors.Is(err, services.ErrAttachmentTypeUnsupported) || errors.Is(err, services.ErrAttachmentEmpty) {
		responses.WriteError(request.Context(), http.StatusBadRequest, err.Error(), response)
	} else {
		slog.Error("Unexpected notes database error", attr.Error(err))
		responses.WriteError(request.Context(), http.StatusInternalServerError, api.ErrorResponseDetailsInternalServerError, response)
	}
}

// ListTags returns the distinct tags in use across live notes so the UI can
// offer tag suggestions when filtering.
func (s Handlers) ListTags(response http.ResponseWriter, request *http.Request) {
	var ctx = request.Context()

	tags, err := s.notes.ListTags(ctx)
	if err != nil {
		handleNotesError(request, response, err)
		return
	}

	responses.WriteBasic(ctx, BuildTagCountsView(tags), http.StatusOK, response)
}

// UploadAttachment accepts a multipart image upload and returns the serving URL
// plus a ready-to-paste markdown image reference for the note editor.
func (s Handlers) UploadAttachment(response http.ResponseWriter, request *http.Request) {
	var ctx = request.Context()

	if err := request.ParseMultipartForm(services.MaxAttachmentSize); err != nil {
		responses.WriteError(ctx, http.StatusBadRequest, api.ErrorResponsePayloadUnmarshalError, response)
		return
	}

	file, header, err := request.FormFile(attachmentFormField)
	if err != nil {
		responses.WriteError(ctx, http.StatusBadRequest, "multipart field '"+attachmentFormField+"' is required", response)
		return
	}
	defer file.Close()

	fileData, err := io.ReadAll(io.LimitReader(file, services.MaxAttachmentSize+1))
	if err != nil {
		responses.WriteError(ctx, http.StatusInternalServerError, api.ErrorResponseDetailsInternalServerError, response)
		return
	}

	contentType := header.Header.Get("Content-Type")
	if sniffedType := http.DetectContentType(fileData); strings.HasPrefix(sniffedType, "image/") {
		contentType = sniffedType
	}

	attachment, err := s.notes.CreateAttachment(ctx, services.Attachment{
		Filename:    header.Filename,
		ContentType: contentType,
		Data:        fileData,
	})
	if err != nil {
		handleNotesError(request, response, err)
		return
	}

	responses.WriteBasic(ctx, BuildAttachmentView(attachment), http.StatusCreated, response)
}

// GetAttachment streams the stored attachment bytes with its original content
// type so markdown image references render in the browser.
func (s Handlers) GetAttachment(response http.ResponseWriter, request *http.Request) {
	var ctx = request.Context()

	rawAttachmentID := mux.Vars(request)[api.URIPathVariableAttachmentID]

	attachmentID, err := strconv.ParseInt(rawAttachmentID, 10, 64)
	if err != nil {
		responses.WriteError(ctx, http.StatusBadRequest, api.ErrorResponseDetailsIDMalformed, response)
		return
	}

	attachment, err := s.notes.GetAttachment(ctx, attachmentID)
	if err != nil {
		handleNotesError(request, response, err)
		return
	}

	response.Header().Set("Content-Type", attachment.ContentType)
	response.Header().Set("Content-Disposition", fmt.Sprintf(`inline; filename="%s"`, attachment.Filename))
	response.Header().Set("Cache-Control", attachmentCacheControl)
	response.WriteHeader(http.StatusOK)

	if _, err := response.Write(attachment.Data); err != nil {
		slog.Error("Failed writing attachment body", attr.Error(err))
	}
}

// GetMedia streams an attachment by its unguessable token without requiring
// authentication, so markdown image references render in the browser and in
// exported documents. Knowing the token is the only capability required.
func (s Handlers) GetMedia(response http.ResponseWriter, request *http.Request) {
	var ctx = request.Context()

	attachmentToken := mux.Vars(request)[api.URIPathVariableAttachmentToken]

	attachment, err := s.notes.GetAttachmentByToken(ctx, attachmentToken)
	if err != nil {
		handleNotesError(request, response, err)
		return
	}

	response.Header().Set("Content-Type", attachment.ContentType)
	response.Header().Set("Content-Disposition", fmt.Sprintf(`inline; filename="%s"`, attachment.Filename))
	response.Header().Set("Cache-Control", attachmentCacheControl)
	response.WriteHeader(http.StatusOK)

	if _, err := response.Write(attachment.Data); err != nil {
		slog.Error("Failed writing media body", attr.Error(err))
	}
}
