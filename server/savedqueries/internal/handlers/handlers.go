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
	"errors"
	"net/http"

	"github.com/specterops/bloodhound/cmd/api/src/api"
	"github.com/specterops/bloodhound/packages/go/params"
	"github.com/specterops/bloodhound/packages/go/responses"
	"github.com/specterops/bloodhound/server/savedqueries/internal/services"
	"github.com/specterops/bloodhound/server/users"
)

type SavedQueries interface {
	GetSavedQuery(ctx context.Context, savedQueryID int64, actor services.Actor) (services.SavedQuery, error)
}

type UserResolver interface {
	Resolve(ctx context.Context) (users.Principal, bool)
}

type Handlers struct {
	savedQueries SavedQueries
	userResolver UserResolver
}

func NewHandlersContainer(savedQueries SavedQueries, userResolver UserResolver) *Handlers {
	return &Handlers{
		savedQueries: savedQueries,
		userResolver: userResolver,
	}
}

func (s *Handlers) GetSavedQuery(response http.ResponseWriter, request *http.Request) {
	var ctx = request.Context()

	principal, isUser := s.userResolver.Resolve(ctx)
	if !isUser {
		responses.WriteError(ctx, http.StatusBadRequest, "no associated user found", response)
		return
	}

	savedQueryID, err := params.PathInt64(request, api.URIPathVariableSavedQueryID)
	if err != nil {
		responses.WriteError(ctx, http.StatusBadRequest, api.ErrorResponseDetailsIDMalformed, response)
		return
	}

	actor := services.Actor{
		UserID:          principal.ID.String(),
		IsAdministrator: principal.IsAdministrator(),
	}
	savedQuery, err := s.savedQueries.GetSavedQuery(ctx, savedQueryID, actor)
	if err != nil {
		handleSavedQueryError(request, response, err)
		return
	}

	responses.WriteBasic(ctx, BuildSavedQueryView(savedQuery), http.StatusOK, response)
}

func handleSavedQueryError(request *http.Request, response http.ResponseWriter, err error) {
	var ctx = request.Context()

	if errors.Is(err, services.ErrSavedQueryNotFound) {
		responses.WriteError(ctx, http.StatusNotFound, "query not found", response)
	} else if errors.Is(err, context.DeadlineExceeded) {
		responses.WriteError(ctx, http.StatusInternalServerError, api.ErrorResponseRequestTimeout, response)
	} else {
		responses.WriteInternalServerError(ctx, err, response)
	}
}
