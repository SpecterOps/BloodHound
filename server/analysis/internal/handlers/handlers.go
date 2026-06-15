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
	"log/slog"
	"net/http"

	"github.com/specterops/bloodhound/cmd/api/src/api"
	"github.com/specterops/bloodhound/cmd/api/src/auth"
	"github.com/specterops/bloodhound/cmd/api/src/bhctx"
	"github.com/specterops/bloodhound/packages/go/bhlog/attr"
	"github.com/specterops/bloodhound/packages/go/responses"
	"github.com/specterops/bloodhound/server/analysis/internal/services"
)

// Analysis defines the analysis service boundary for the analysis handlers package.
type Analysis interface {
	GetRequest(context.Context) (services.RequestedAnalysis, error)
	CreateRequest(ctx context.Context, requestedBy string) (services.RequestedAnalysis, bool, error)
	CancelAnalysisRequest(ctx context.Context) error
}

// Handlers is a dependency injection container for analysis handlers
type Handlers struct {
	analysis Analysis
}

// NewHandlersContainer initializes the Handlers dependency injection container
func NewHandlersContainer(analysis Analysis) *Handlers {
	return &Handlers{
		analysis: analysis,
	}
}

// GetAnalysisRequest returns the currently pending analysis request. Returns 200 with the
// request details when one is pending, or 204 No Content when no request is
// currently pending.
func (s Handlers) GetAnalysisRequest(response http.ResponseWriter, request *http.Request) {
	var ctx = request.Context()

	ra, err := s.analysis.GetRequest(ctx)
	if err != nil {
		handleAnalysisError(request, response, err)
		return
	}

	responses.WriteBasic(ctx, BuildRequestedAnalysisView(ra), http.StatusOK, response)
}

// CreateAnalysisRequest submits a new analysis request attributed to the authenticated user.
// Returns 202 Accepted when this call accepted the request, or 200 OK when a request
// was already pending (the body in both cases describes the currently pending request).
//
// Authentication is enforced by the route middleware (RequirePermissions); if no user
// is present on the auth context here it indicates an unexpected internal state and a
// 500 is returned.
func (s Handlers) CreateAnalysisRequest(response http.ResponseWriter, request *http.Request) {
	var ctx = request.Context()

	user, isUser := auth.GetUserFromAuthCtx(bhctx.FromRequest(request).AuthCtx)
	if !isUser {
		responses.WriteInternalServerError(ctx, errors.New("no user on auth context after authentication middleware"), response)
		return
	}

	current, created, err := s.analysis.CreateRequest(ctx, user.ID.String())
	if err != nil {
		handleAnalysisError(request, response, err)
		return
	}

	statusCode := http.StatusOK
	if created {
		statusCode = http.StatusAccepted
	}
	responses.WriteBasic(ctx, BuildRequestedAnalysisView(current), statusCode, response)
}

func (s Handlers) CancelAnalysisRequest(response http.ResponseWriter, request *http.Request) {
	var ctx = request.Context()

	_, isUser := auth.GetUserFromAuthCtx(bhctx.FromRequest(request).AuthCtx)
	if !isUser {
		responses.WriteInternalServerError(ctx, errors.New("no user on auth context after authentication middleware"), response)
		return
	}

	err := s.analysis.CancelAnalysisRequest(ctx)
	if err != nil {
		handleAnalysisError(request, response, err)
		return
	}

	response.WriteHeader(http.StatusAccepted)
}

func handleAnalysisError(request *http.Request, response http.ResponseWriter, err error) {
	if errors.Is(err, services.ErrNoPendingRequest) {
		response.WriteHeader(http.StatusNoContent)
	} else if errors.Is(err, services.ErrDeletionRequestPending) {
		response.WriteHeader(http.StatusConflict)
	} else if errors.Is(err, context.DeadlineExceeded) {
		responses.WriteError(request.Context(), http.StatusInternalServerError, api.ErrorResponseRequestTimeout, response)
	} else {
		slog.Error("Unexpected database error", attr.Error(err))
		responses.WriteError(request.Context(), http.StatusInternalServerError, api.ErrorResponseDetailsInternalServerError, response)
	}
}
