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
	"encoding/json"
	"net/http"

	"github.com/specterops/bloodhound/src/api"
	// "github.com/specterops/bloodhound/src/model"
)

type CreateEnvironmentConfigurationRequest struct {
	Name string          `json:"name"`
	Data json.RawMessage `json:"data"`
}

func (s Resources) CreateEnvironmentConfiguration(response http.ResponseWriter, request *http.Request) {
	var createRequest CreateEnvironmentConfigurationRequest

	if err := api.ReadJSONRequestPayloadLimited(&createRequest, request); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, err.Error(), request), response)
		return
	}

	if createRequest.Name == "" || len(createRequest.Data) == 0 {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, "name and data fields are required", request), response)
		return
	}

	// Validate the data structure
	var dataMap map[string]interface{}
	if err := json.Unmarshal(createRequest.Data, &dataMap); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, "invalid data format", request), response)
		return
	}

	if _, hasMeta := dataMap["meta"]; !hasMeta {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, "missing meta field in data", request), response)
		return
	}

	if _, hasNodeTypes := dataMap["nodeTypes"]; !hasNodeTypes {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, "missing nodeTypes field in data", request), response)
		return
	}

	if _, hasRelationshipTypes := dataMap["relationshipTypes"]; !hasRelationshipTypes {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, "missing relationshipTypes field in data", request), response)
		return
	}

	envConfig, err := s.DB.CreateEnvironmentConfiguration(request.Context(), createRequest.Name, createRequest.Data)
	if err != nil {
		api.HandleDatabaseError(request, response, err)
		return
	}

	api.WriteBasicResponse(request.Context(), envConfig, http.StatusCreated, response)
}
