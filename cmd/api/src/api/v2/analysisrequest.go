// Copyright 2024 Specter Ops, Inc.
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

package v2

import (
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"

	"github.com/specterops/bloodhound/cmd/api/src/api"
	"github.com/specterops/bloodhound/cmd/api/src/auth"
	"github.com/specterops/bloodhound/cmd/api/src/ctx"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/packages/go/analysis"
	"github.com/specterops/bloodhound/packages/go/bhlog/measure"
)

type AnalysisRequestPayload struct {
	AnalysisStep int `json:"analysis_step"`
}

func (s Resources) GetAnalysisRequest(response http.ResponseWriter, request *http.Request) {
	if analysisRequest, err := s.DB.GetAnalysisRequest(request.Context()); err != nil && !errors.Is(err, sql.ErrNoRows) {
		api.HandleDatabaseError(request, response, err)
	} else {
		api.WriteBasicResponse(request.Context(), analysisRequest, http.StatusOK, response)
	}
}

func (s Resources) RequestAnalysis(response http.ResponseWriter, request *http.Request) {
	defer measure.ContextMeasureWithThreshold(request.Context(), slog.LevelDebug, "Requesting analysis")()

	var userId string
	if user, isUser := auth.GetUserFromAuthCtx(ctx.FromRequest(request).AuthCtx); !isUser {
		slog.WarnContext(request.Context(), "Encountered request analysis for unknown user, this shouldn't happen")
		userId = "unknown-user"
	} else {
		userId = user.ID.String()
	}

	step := int(analysis.AnalysisStepAll)
	if request.Body != nil {
		if body, err := io.ReadAll(request.Body); err == nil && len(body) > 0 {
			var payload AnalysisRequestPayload
			if err := json.Unmarshal(body, &payload); err != nil {
				api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, "JSON malformed", request), response)
				return
			}
			if payload.AnalysisStep != 0 {
				step = payload.AnalysisStep
			}
		}
	}

	if err := s.DB.RequestAnalysis(request.Context(), userId, step); err != nil {
		api.HandleDatabaseError(request, response, err)
		return
	}

	response.WriteHeader(http.StatusAccepted)
}

func (s Resources) CancelAnalysisRequest(response http.ResponseWriter, request *http.Request) {
	defer measure.ContextMeasure(request.Context(), slog.LevelDebug, "Cancelling analysis request")()

	if _, isUser := auth.GetUserFromAuthCtx(ctx.FromRequest(request).AuthCtx); !isUser {
		slog.ErrorContext(request.Context(), "Unable to get user from auth context")
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusUnauthorized, api.ErrorResponseUnknownUser.Error(), request), response)
	} else if analysisRequest, err := s.DB.GetAnalysisRequest(request.Context()); errors.Is(err, sql.ErrNoRows) {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusNotFound, api.ErrorResponseDetailsResourceNotFound, request), response)
	} else if err != nil {
		api.HandleDatabaseError(request, response, err)
	} else if analysisRequest.RequestType == model.AnalysisRequestDeletion {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusConflict, api.ErrorResponseAnalysisRequestTypeDeletionPending, request), response)
	} else if err := s.DB.DeleteAnalysisRequest(request.Context()); err != nil {
		api.HandleDatabaseError(request, response, err)
	} else {
		response.WriteHeader(http.StatusAccepted)
	}
}
