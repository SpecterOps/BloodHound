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

// GetAnalysisRequest returns the currently pending analysis request. Always returns 200 OK
// with the request details. If no request exists, returns a zero-valued struct.
func (s Handlers) GetAnalysisRequest(response http.ResponseWriter, request *http.Request) {
	var ctx = request.Context()

	ra, err := s.analysis.GetRequest(ctx)
	if err != nil && !errors.Is(err, services.ErrNoPendingRequest) {
		handleAnalysisError(request, response, err)
		return
	}

	// Always return 200 OK with the request (even if zero-valued when no request exists)
	// This matches the behavior in main where sql.ErrNoRows is ignored and zero-valued struct is returned
	responses.WriteBasic(ctx, BuildRequestedAnalysisView(ra), http.StatusOK, response)
}

// CreateAnalysisRequest submits a new analysis request attributed to the authenticated user.
// Always returns 202 Accepted with no response body.
//
// Authentication is enforced by the route middleware (RequirePermissions); if no user
// is present on the auth context, logs a warning and uses "unknown-user" as a fallback.
func (s Handlers) CreateAnalysisRequest(response http.ResponseWriter, request *http.Request) {
	var (
		ctx    = request.Context()
		userId string
	)

	user, isUser := auth.GetUserFromAuthCtx(bhctx.FromRequest(request).AuthCtx)
	if !isUser {
		slog.WarnContext(ctx, "Encountered request analysis for unknown user, this shouldn't happen")
		userId = "unknown-user"
	} else {
		userId = user.ID.String()
	}

	_, _, err := s.analysis.CreateRequest(ctx, userId)
	if err != nil {
		handleAnalysisError(request, response, err)
		return
	}

	// Always return 202 Accepted with no response body, matching main behavior
	response.WriteHeader(http.StatusAccepted)
}

// CancelAnalysisRequest cancels a pending analysis request.
// Returns 401 if user is missing, 404 if no request exists, 409 if deletion request is pending,
// or 202 on success.
func (s Handlers) CancelAnalysisRequest(response http.ResponseWriter, request *http.Request) {
	var ctx = request.Context()

	_, isUser := auth.GetUserFromAuthCtx(bhctx.FromRequest(request).AuthCtx)
	if !isUser {
		slog.ErrorContext(ctx, "Unable to get user from auth context")
		responses.WriteError(ctx, http.StatusUnauthorized, api.ErrorResponseUnknownUser.Error(), response)
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
	var ctx = request.Context()

	if errors.Is(err, services.ErrNoPendingRequest) {
		// For DELETE requests, return 404 Not Found (request doesn't exist to cancel)
		// For other requests, this is handled in the caller
		if request.Method == http.MethodDelete {
			responses.WriteError(ctx, http.StatusNotFound, api.ErrorResponseDetailsResourceNotFound, response)
		} else {
			response.WriteHeader(http.StatusNoContent)
		}
	} else if errors.Is(err, services.ErrDeletionRequestPending) {
		responses.WriteError(ctx, http.StatusConflict, api.ErrorResponseAnalysisRequestTypeDeletionPending, response)
	} else if errors.Is(err, context.DeadlineExceeded) {
		responses.WriteError(ctx, http.StatusInternalServerError, api.ErrorResponseRequestTimeout, response)
	} else {
		slog.Error("Unexpected database error", attr.Error(err))
		responses.WriteError(ctx, http.StatusInternalServerError, api.ErrorResponseDetailsInternalServerError, response)
	}
}
