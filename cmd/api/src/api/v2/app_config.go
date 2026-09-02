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
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/specterops/bloodhound/cmd/api/src/api"
	"github.com/specterops/bloodhound/cmd/api/src/database"
	"github.com/specterops/bloodhound/cmd/api/src/database/types/null"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/cmd/api/src/model/appcfg"
	"github.com/specterops/bloodhound/packages/go/bhlog/attr"
	"github.com/teambition/rrule-go"
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
		parameterFilterValue := appcfg.ParameterKey(parameterFilter.Value)
		if parameterFilter.Operator != model.Equals {
			api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf("%s: %s %s", api.ErrorResponseDetailsFilterPredicateNotSupported, parameterFilter.Name, parameterFilter.Operator), request), response)
		} else if !cfgParameter.IsValidKey(parameterFilterValue) || cfgParameter.IsProtectedKey(parameterFilterValue) {
			api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf("Configuration parameter %s is not valid.", parameterFilter.Value), request), response)
		} else if cfgParameter, err = s.DB.GetConfigurationParameter(request.Context(), parameterFilterValue); err != nil {
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
		ctx       = request.Context()
	)

	if err := api.ReadJSONRequestPayloadLimited(&appConfig, request); err != nil {
		api.WriteErrorResponse(ctx, api.BuildErrorResponse(http.StatusBadRequest, api.ErrorResponsePayloadUnmarshalError, request), response)
	} else if parameter, err := appcfg.ConvertAppConfigUpdateRequestToParameter(appConfig); err != nil {
		api.WriteErrorResponse(ctx, api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf("Configuration update request not converted to a parameter: %s", appConfig), request), response)
	} else if !parameter.IsValidKey(parameter.Key) {
		api.WriteErrorResponse(ctx, api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf("Configuration parameter %s is not valid.", parameter.Key), request), response)
	} else if errs := parameter.Validate(); errs != nil {
		api.WriteErrorResponse(ctx, api.BuildErrorResponse(http.StatusBadRequest, errs.Error(), request), response)
	} else if !checkApiKeyExpirationParamAvailable(request.Context(), parameter, s.DB) {
		api.WriteErrorResponse(ctx, api.BuildErrorResponse(http.StatusForbidden, fmt.Sprintf("Configuration parameter %s is not available.", parameter.Key), request), response)
	} else if err := s.DB.SetConfigurationParameter(ctx, parameter); err != nil {
		api.HandleDatabaseError(request, response, err)
	} else {
		if parameter.Key == appcfg.ScheduledAnalysis {
			updateNextScheduledAnalysisStartTime(ctx, s.DB, parameter)
		}

		api.WriteBasicResponse(request.Context(), appConfig, http.StatusOK, response)
	}
}

func updateNextScheduledAnalysisStartTime(ctx context.Context, db database.DatapipeStatusData, parameter appcfg.Parameter) {
	var scheduledAnalysis appcfg.ScheduledAnalysisParameter
	if err := parameter.Map(&scheduledAnalysis); err != nil {
		slog.ErrorContext(ctx, "Error mapping scheduled analysis parameter", attr.Error(err))
	} else if scheduledAnalysis.Enabled {
		if rule, err := rrule.StrToRRule(scheduledAnalysis.RRule); err != nil {
			slog.ErrorContext(ctx, "Error parsing scheduled analysis rrule", attr.Error(err))
		} else if err := db.SetNextScheduledAnalysisStartTime(ctx, null.TimeFrom(rule.After(time.Now(), true))); err != nil {
			slog.ErrorContext(ctx, "Error setting next scheduled analysis start time", attr.Error(err))
		}
	} else {
		if err := db.SetNextScheduledAnalysisStartTime(ctx, null.Time{}); err != nil {
			slog.ErrorContext(ctx, "Error clearing the next scheduled analysis start time", attr.Error(err))
		}
	}
}

func checkApiKeyExpirationParamAvailable(ctx context.Context, param appcfg.Parameter, flag appcfg.GetFlagByKeyer) bool {
	if param.Key == appcfg.APITokenExpiration {
		if f, err := flag.GetFlagByKey(ctx, appcfg.FeatureAPIKeyExpirationSupport); err != nil || !f.Enabled {
			return false
		}
	}

	return true
}
