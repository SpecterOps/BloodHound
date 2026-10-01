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

import (
	"errors"
	"io"
	"log/slog"
	"mime"
	"net/http"

	"github.com/gofrs/uuid"
	"github.com/gorilla/mux"
	"github.com/specterops/bloodhound/cmd/api/src/api"
	"github.com/specterops/bloodhound/cmd/api/src/auth"
	"github.com/specterops/bloodhound/cmd/api/src/bhctx"
	"github.com/specterops/bloodhound/packages/go/bhlog/attr"
	"github.com/specterops/bloodhound/packages/go/responses"
	"github.com/specterops/bloodhound/server/identity/internal/services"
)

// preferenceOwner validates the path before comparing UUID values, allowing
// equivalent UUID representations while denying access to any other user.
// The route separately requires AuthManageSelf; AuthManageUsers is no override.
func preferenceOwner(response http.ResponseWriter, request *http.Request) (uuid.UUID, bool) {
	var (
		ctx           = request.Context()
		actor, isUser = auth.GetUserFromAuthCtx(bhctx.Get(ctx).AuthCtx)
		userID, err   = uuid.FromString(mux.Vars(request)[api.URIPathVariableUserID])
	)
	response.Header().Set("Cache-Control", "no-store")
	if err != nil {
		responses.WriteError(ctx, http.StatusBadRequest, api.ErrorResponseDetailsIDMalformed, response)
		return uuid.Nil, false
	}
	if !isUser || actor.ID != userID {
		responses.WriteError(ctx, http.StatusForbidden, "preferences are only accessible to their owner", response)
		return uuid.Nil, false
	}
	return userID, true
}

// GetUserPreferences returns opaque bytes directly, without a JSON envelope.
func (s *Handlers) GetUserPreferences(response http.ResponseWriter, request *http.Request) {
	userID, authorized := preferenceOwner(response, request)
	if !authorized {
		return
	}
	storage, err := s.identity.GetUserPreferences(request.Context(), userID)
	if err != nil {
		handleIdentityError(request, response, err)
		return
	}
	response.Header().Set("Content-Type", "application/octet-stream")
	response.WriteHeader(http.StatusOK)
	if _, err = response.Write(storage); err != nil {
		slog.ErrorContext(request.Context(), "Writing user preferences response", attr.Error(err))
	}
}

// UpsertUserPreferences reads a bounded binary body before replacing storage.
// Failed and oversized reads never result in partial writes.
func (s *Handlers) UpsertUserPreferences(response http.ResponseWriter, request *http.Request) {
	var (
		ctx       = request.Context()
		sizeError *http.MaxBytesError
	)
	userID, authorized := preferenceOwner(response, request)
	if !authorized {
		return
	}
	contentType, _, err := mime.ParseMediaType(request.Header.Get("Content-Type"))
	if err != nil || contentType != "application/octet-stream" {
		responses.WriteError(ctx, http.StatusUnsupportedMediaType, "Content-Type must be application/octet-stream", response)
		return
	}
	request.Body = http.MaxBytesReader(response, request.Body, services.MaxUserPreferencesBytes)
	storage, err := io.ReadAll(request.Body)
	if errors.As(err, &sizeError) {
		handleIdentityError(request, response, services.ErrUserPreferencesTooLarge)
		return
	}
	if err != nil {
		responses.WriteError(ctx, http.StatusBadRequest, "unable to read user preferences", response)
		return
	}
	if err = s.identity.UpsertUserPreferences(ctx, userID, storage); err != nil {
		handleIdentityError(request, response, err)
		return
	}
	response.WriteHeader(http.StatusNoContent)
}
