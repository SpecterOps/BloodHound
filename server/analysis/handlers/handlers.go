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

	"github.com/specterops/bloodhound/cmd/api/src/auth"
	"github.com/specterops/bloodhound/cmd/api/src/bhctx"
	"github.com/specterops/bloodhound/server/analysis/services"
	"github.com/specterops/bloodhound/server/jsonapiv2/responses"
)

const (
	errMessageUnauthenticated = "authentication is required to submit an analysis request"
)

// Analysis defines the analysis service boundary for the analysis handlers package.
type Analysis interface {
	GetRequest(context.Context) (services.RequestedAnalysis, error)
	CreateRequest(ctx context.Context, requestedBy string) (services.RequestedAnalysis, bool, error)
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

// GetRequest returns the currently pending analysis request. Returns 200 with the
// request details when one is pending, or 204 No Content when no request is
// currently pending.
func (s Handlers) GetRequest(response http.ResponseWriter, request *http.Request) {
	var ctx = request.Context()

	ra, err := s.analysis.GetRequest(ctx)
	if errors.Is(err, services.ErrNoPendingRequest) {
		responses.WriteNoContent(response)
		return
	}
	if err != nil {
		responses.WriteInternalServerError(ctx, err, response)
		return
	}

	responses.WriteBasic(ctx, BuildRequestedAnalysisView(ra), http.StatusOK, response)
}

// CreateRequest submits a new analysis request attributed to the authenticated user.
// Returns 202 Accepted when this call accepted the request, 200 OK when a request was
// already pending (the body in both cases describes the currently pending request), or
// 401 Unauthorized when no authenticated user can be resolved from the request.
func (s Handlers) CreateRequest(response http.ResponseWriter, request *http.Request) {
	var ctx = request.Context()

	user, isUser := auth.GetUserFromAuthCtx(bhctx.FromRequest(request).AuthCtx)
	if !isUser {
		responses.WriteError(ctx, http.StatusUnauthorized, errMessageUnauthenticated, response)
		return
	}

	current, created, err := s.analysis.CreateRequest(ctx, user.ID.String())
	if err != nil {
		responses.WriteInternalServerError(ctx, err, response)
		return
	}

	statusCode := http.StatusOK
	if created {
		statusCode = http.StatusAccepted
	}
	responses.WriteBasic(ctx, BuildRequestedAnalysisView(current), statusCode, response)
}
