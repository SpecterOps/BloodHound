// Copyright 2023 Specter Ops, Inc.
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

package tools

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/specterops/bloodhound/cmd/api/src/api"
	"github.com/specterops/bloodhound/cmd/api/src/database"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/cmd/api/src/model/appcfg"
)

const (
	URIPathVariableFeatureID = "feature_id"
)

type ToolContainer struct {
	db database.Database
}

type ToggleFlagResponse struct {
	Enabled bool `json:"enabled"`
}

func NewToolContainer(db database.Database) ToolContainer {
	return ToolContainer{db: db}
}

func (s ToolContainer) GetFlags(response http.ResponseWriter, request *http.Request) {
	if flags, err := s.db.GetAllFlags(request.Context()); err != nil {
		api.HandleDatabaseError(request, response, err)
	} else {
		api.WriteBasicResponse(request.Context(), flags, http.StatusOK, response)
	}
}

func shouldRequestAnalysisOnEnable(previouslyEnabled bool, currentlyEnabled bool) bool {
	return !previouslyEnabled && currentlyEnabled
}

func (s ToolContainer) ToggleFlag(response http.ResponseWriter, request *http.Request) {
	var (
		ctx          = request.Context()
		rawFeatureID = chi.URLParam(request, URIPathVariableFeatureID)
	)

	featureID, err := strconv.ParseInt(rawFeatureID, 10, 32)
	if err != nil {
		api.WriteErrorResponse(ctx, api.BuildErrorResponse(http.StatusBadRequest, api.ErrorResponseDetailsIDMalformed, request), response)
		return
	}

	featureFlag, err := s.db.GetFlag(ctx, int32(featureID))
	if err != nil {
		api.HandleDatabaseError(request, response, err)
		return
	}

	previouslyEnabled := featureFlag.Enabled
	featureFlag.Enabled = !featureFlag.Enabled

	if err := s.db.SetFlag(ctx, featureFlag); err != nil {
		api.HandleDatabaseError(request, response, err)
		return
	}

	if featureFlag.Key == appcfg.FeatureFindingsPrioritizationV0 &&
		shouldRequestAnalysisOnEnable(previouslyEnabled, featureFlag.Enabled) {
		if err := s.db.RequestAnalysis(ctx, appcfg.PrioritizationFlagRequestSource, model.AnalysisModeNoPostProcessing); err != nil {
			featureFlag.Enabled = previouslyEnabled

			if rollbackErr := s.db.SetFlag(ctx, featureFlag); rollbackErr != nil {
				api.HandleDatabaseError(request, response, errors.Join(err, rollbackErr))
				return
			}

			api.HandleDatabaseError(request, response, err)
			return
		}
	}

	api.WriteBasicResponse(ctx, ToggleFlagResponse{
		Enabled: featureFlag.Enabled,
	}, http.StatusOK, response)
}
