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

package tools

import (
	"fmt"
	"net/http"
	"time"

	"github.com/specterops/bloodhound/cmd/api/src/api"
	"github.com/specterops/bloodhound/cmd/api/src/database/types"
	"github.com/specterops/bloodhound/cmd/api/src/model/appcfg"
	"github.com/specterops/bloodhound/cmd/api/src/utils/validation"
)

type ScheduledAnalysisConfiguration struct {
	Enabled bool   `json:"enabled"`
	RRule   string `json:"rrule"`
}

const ErrFailedRetrievingData = "error retrieving configuration data: %v"

func (s ToolContainer) GetScheduledAnalysisConfiguration(response http.ResponseWriter, request *http.Request) {
	if config, err := appcfg.GetScheduledAnalysisParameter(request.Context(), s.db); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusInternalServerError, fmt.Sprintf(ErrFailedRetrievingData, err), request), response)
	} else {
		api.WriteJSONResponse(request.Context(), config, http.StatusOK, response)
	}
}

func (s ToolContainer) SetScheduledAnalysisConfiguration(response http.ResponseWriter, request *http.Request) {
	var config ScheduledAnalysisConfiguration
	ctx := request.Context()

	if err := api.ReadJSONRequestPayloadLimited(&config, request); err != nil {
		api.WriteErrorResponse(ctx, api.BuildErrorResponse(http.StatusBadRequest, api.ErrorParseParams, request), response)
	} else if !config.Enabled {
		nextParameter := appcfg.ScheduledAnalysisParameter{
			Enabled: false,
		}

		if val, err := types.NewJSONBObject(nextParameter); err != nil {
			api.WriteErrorResponse(ctx, api.BuildErrorResponse(http.StatusInternalServerError, fmt.Sprintf("failed to convert value to JSONBObject: %v", err), request), response)
		} else {
			updatedParameter := appcfg.Parameter{
				Key:   appcfg.ScheduledAnalysis,
				Value: val,
			}

			if err := s.db.SetConfigurationParameter(ctx, updatedParameter); err != nil {
				api.WriteErrorResponse(ctx, api.BuildErrorResponse(http.StatusInternalServerError, fmt.Sprintf("error updating database: %v", api.FormatDatabaseError(err)), request), response)
			}
		}
	} else {
		if rule, err := validation.ValidateRRule(config.RRule); err != nil {
			api.WriteErrorResponse(ctx, api.BuildErrorResponse(http.StatusBadRequest, err.Error(), request), response)
		} else {
			nextParameter := appcfg.ScheduledAnalysisParameter{
				Enabled: true,
				RRule:   config.RRule,
			}

			if val, err := types.NewJSONBObject(nextParameter); err != nil {
				api.WriteErrorResponse(ctx, api.BuildErrorResponse(http.StatusInternalServerError, fmt.Sprintf("failed to convert value to JSONBObject: %v", err), request), response)
			} else {
				updatedParameter := appcfg.Parameter{
					Key:   appcfg.ScheduledAnalysis,
					Value: val,
				}

				if err := s.db.SetConfigurationParameter(ctx, updatedParameter); err != nil {
					api.WriteErrorResponse(ctx, api.BuildErrorResponse(http.StatusInternalServerError, fmt.Sprintf("error setting analysis schedule: %v", api.FormatDatabaseError(err)), request), response)
				} else if err = s.db.SetNextScheduledAnalysisStartTime(ctx, rule.After(time.Now(), true)); err != nil {
					api.WriteErrorResponse(ctx, api.BuildErrorResponse(http.StatusInternalServerError, fmt.Sprintf("scheduled analysis updated, but there was an error updating the next scheduled analysis start time: %v", api.FormatDatabaseError(err)), request), response)
				}

			}
		}
	}
}
