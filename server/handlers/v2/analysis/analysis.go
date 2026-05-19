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

package analysis

//go:generate go tool mockery

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/specterops/bloodhound/cmd/api/src/auth"
	"github.com/specterops/bloodhound/cmd/api/src/bhctx"
	"github.com/specterops/bloodhound/server/jsonapi/v2/responses"
	"github.com/specterops/bloodhound/server/models"
	"github.com/specterops/bloodhound/server/services/analysis"
)

// Analysis defines the analysis service boundary for the analysis handlers package.
type Analysis interface {
	GetRequest(context.Context) (models.RequestedAnalysis, error)
	CreateRequest(ctx context.Context, requestedBy string) error
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

// GetRequest returns the currently pending analysis request.
// A 200 with a zero-value body is returned when no request is pending; 200
// with the request details otherwise.
func (h Handlers) GetRequest(response http.ResponseWriter, request *http.Request) {
	var ctx = request.Context()

	ra, err := h.analysis.GetRequest(ctx)
	if err != nil && !errors.Is(err, analysis.ErrNoPendingRequest) {
		responses.WriteInternalServerError(request, err, response)
		return
	}

	responses.WriteBasic(ctx, BuildRequestedAnalysisView(ra), http.StatusOK, response)
}

// CreateRequest submits a new analysis request attributed to the authenticated user.
// Returns 202 Accepted on success. If no auth context is present, the request is
// attributed to "unknown-user" (matching legacy behaviour).
func (h Handlers) CreateRequest(response http.ResponseWriter, request *http.Request) {
	var (
		ctx    = request.Context()
		userID string
	)

	if user, isUser := auth.GetUserFromAuthCtx(bhctx.FromRequest(request).AuthCtx); !isUser {
		slog.WarnContext(ctx, "Encountered request analysis for unknown user, this shouldn't happen")
		userID = "unknown-user"
	} else {
		userID = user.ID.String()
	}

	err := h.analysis.CreateRequest(ctx, userID)
	if err != nil {
		responses.WriteInternalServerError(request, err, response)
		return
	}

	response.WriteHeader(http.StatusAccepted)
}
