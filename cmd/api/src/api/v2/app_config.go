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

package v2

import (
	"fmt"
	"net/http"

	"github.com/specterops/bloodhound/src/api"
	"github.com/specterops/bloodhound/src/model"
	"github.com/specterops/bloodhound/src/model/appcfg"
)

type ListAppConfigParametersResponse struct {
	Data appcfg.Parameters `json:"data"`
}

func (s Resources) GetApplicationConfigurations(response http.ResponseWriter, request *http.Request) {
	var cfgParameter appcfg.Parameter

	const queryParameterName = "parameter"

	if queryFilters, err := s.QueryParameterFilterParser.ParseQueryParameterFilters(request); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, api.ErrorResponseDetailsBadQueryParameterFilters, request), response)
	} else if parameterFilter, hasParameterFilter := queryFilters.FirstFilter(queryParameterName); hasParameterFilter {
		if parameterFilter.Operator != model.Equals {
			api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf("%s: %s %s", api.ErrorResponseDetailsFilterPredicateNotSupported, parameterFilter.Name, parameterFilter.Operator), request), response)
		} else if !cfgParameter.IsValidKey(appcfg.ParameterKey(parameterFilter.Value)) {
			api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf("Configuration parameter %s is not valid.", parameterFilter.Value), request), response)
		} else if cfgParameter, err = s.DB.GetConfigurationParameter(request.Context(), appcfg.ParameterKey(parameterFilter.Value)); err != nil {
			api.HandleDatabaseError(request, response, err)
		} else {
			api.WriteBasicResponse(request.Context(), appcfg.Parameters{cfgParameter}, http.StatusOK, response)
		}
	} else if cfgParameters, err := s.DB.GetAllConfigurationParameters(request.Context()); err != nil {
		api.HandleDatabaseError(request, response, err)
	} else {
		api.WriteBasicResponse(request.Context(), cfgParameters, http.StatusOK, response)
	}
}

func (s Resources) SetApplicationConfiguration(response http.ResponseWriter, request *http.Request) {
	var (
		appConfig appcfg.AppConfigUpdateRequest
	)

	if err := api.ReadJSONRequestPayloadLimited(&appConfig, request); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, api.ErrorResponsePayloadUnmarshalError, request), response)
	} else if parameter, err := appcfg.ConvertAppConfigUpdateRequestToParameter(appConfig); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf("Configuration update request not converted to a parameter: %s", parameter.Key), request), response)
	} else if !parameter.IsValidKey(parameter.Key) {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf("Configuration parameter %s is not valid.", parameter.Key), request), response)
	} else if errs := parameter.Validate(); errs != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, errs.Error(), request), response)
	} else if err = s.DB.SetConfigurationParameter(request.Context(), parameter); err != nil {
		api.HandleDatabaseError(request, response, err)
	} else {
		api.WriteBasicResponse(request.Context(), appConfig, http.StatusOK, response)
	}
}
