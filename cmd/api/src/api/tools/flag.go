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
	"net/http"
	"strconv"

	"github.com/specterops/bloodhound/src/api"
	"github.com/specterops/bloodhound/src/database"
	"github.com/go-chi/chi/v5"
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
	if flags, err := s.db.GetAllFlags(); err != nil {
		api.HandleDatabaseError(request, response, err)
	} else {
		api.WriteBasicResponse(request.Context(), flags, http.StatusOK, response)
	}
}

func (s ToolContainer) ToggleFlag(response http.ResponseWriter, request *http.Request) {
	rawFeatureID := chi.URLParam(request, URIPathVariableFeatureID)

	if featureID, err := strconv.ParseInt(rawFeatureID, 10, 32); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, api.ErrorResponseDetailsIDMalformed, request), response)
	} else if featureFlag, err := s.db.GetFlag(int32(featureID)); err != nil {
		api.HandleDatabaseError(request, response, err)
	} else {
		featureFlag.Enabled = !featureFlag.Enabled

		if err := s.db.SetFlag(featureFlag); err != nil {
			api.HandleDatabaseError(request, response, err)
		} else {
			api.WriteBasicResponse(request.Context(), ToggleFlagResponse{
				Enabled: featureFlag.Enabled,
			}, http.StatusOK, response)
		}
	}
}
