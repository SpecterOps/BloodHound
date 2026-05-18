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

import (
	"context"
	"errors"
	"net/http"

	"github.com/specterops/bloodhound/server/jsonapi/v2/responses"
	"github.com/specterops/bloodhound/server/models"
	"github.com/specterops/bloodhound/server/services/analysis"
)

// Analysis defines the analysis service boundary for the analysis handlers package
type Analysis interface {
	GetRequest(context.Context) (models.RequestedAnalysis, error)
	CreateRequest(context.Context, string) error
	CancelRequest(context.Context) error
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

// GetRequest gets any currently requested analysis that is pending. A 200 response with the
// zero-value AnalysisRequest is returned when no request is pending.
func (s Handlers) GetRequest(response http.ResponseWriter, request *http.Request) {
	var ctx = request.Context()

	ra, err := s.analysis.GetRequest(ctx)

	if err != nil && !errors.Is(err, analysis.ErrNoPendingRequest) {
		responses.WriteInternalServerError(request, err, response)
		return
	}

	requestedAnalysis := BuildRequestedAnalysisView(ra)

	responses.WriteBasic(ctx, requestedAnalysis, http.StatusOK, response)
}

// // CreateRequest creates a new requested analysis run for the current user
// func (s Handlers) CreateRequest(response http.ResponseWriter, request *http.Request) {
// 	defer measure.ContextMeasureWithThreshold(request.Context(), slog.LevelDebug, "Requesting analysis")()

// 	var userId string
// 	if user, isUser := auth.GetUserFromAuthCtx(ctx.FromRequest(request).AuthCtx); !isUser {
// 		slog.WarnContext(request.Context(), "Encountered request analysis for unknown user, this shouldn't happen")
// 		userId = "unknown-user"
// 	} else {
// 		userId = user.ID.String()
// 	}

// 	if err := s.analysis.CreateRequest(request.Context(), userId); err != nil {
// 		api.HandleDatabaseError(request, response, err)
// 		return
// 	}

// 	response.WriteHeader(http.StatusAccepted)
// }

// // CancelRequest removes any pending requested analysis
// func (s Handlers) CancelRequest(response http.ResponseWriter, request *http.Request) {
// 	defer measure.ContextMeasure(request.Context(), slog.LevelDebug, "Cancelling analysis request")()

// 	if _, isUser := auth.GetUserFromAuthCtx(ctx.FromRequest(request).AuthCtx); !isUser {
// 		slog.ErrorContext(request.Context(), "Unable to get user from auth context")
// 		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusUnauthorized, api.ErrorResponseUnknownUser.Error(), request), response)
// 	} else if analysisRequest, err := s.analysis.GetRequest(request.Context()); errors.Is(err, sql.ErrNoRows) {
// 		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusNotFound, api.ErrorResponseDetailsResourceNotFound, request), response)
// 	} else if err != nil {
// 		api.HandleDatabaseError(request, response, err)
// 	} else if analysisRequest.RequestType == model.AnalysisRequestDeletion {
// 		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusConflict, api.ErrorResponseAnalysisRequestTypeDeletionPending, request), response)
// 	} else if err := s.analysis.CancelRequest(request.Context()); err != nil {
// 		api.HandleDatabaseError(request, response, err)
// 	} else {
// 		response.WriteHeader(http.StatusAccepted)
// 	}
// }
