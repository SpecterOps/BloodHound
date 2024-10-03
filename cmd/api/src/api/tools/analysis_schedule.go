package tools

import (
	"fmt"
	"github.com/specterops/bloodhound/src/api"
	v2 "github.com/specterops/bloodhound/src/api/v2"
	"github.com/specterops/bloodhound/src/database/types"
	"github.com/specterops/bloodhound/src/model/appcfg"
	"github.com/teambition/rrule-go"
	"net/http"
	"strings"
)

type ScheduledAnalysisConfiguration struct {
	Enabled bool   `json:"enabled"`
	RRule   string `json:"rrule"`
}

const ErrorInvalidRrule = "invalid rrule specified: %v"
const ErrorFailedRetrievingData = "error retrieving configuration data: %v"

func (s ToolContainer) GetScheduledAnalysisConfiguration(response http.ResponseWriter, request *http.Request) {
	if config, err := appcfg.GetScheduledAnalysisParameter(request.Context(), s.db); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusInternalServerError, fmt.Sprintf(ErrorFailedRetrievingData, err), request), response)
	} else {
		api.WriteJSONResponse(request.Context(), config, http.StatusOK, response)
	}
}

func (s ToolContainer) SetScheduledAnalysisConfiguration(response http.ResponseWriter, request *http.Request) {
	var config ScheduledAnalysisConfiguration

	if err := api.ReadJSONRequestPayloadLimited(&config, request); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, v2.ErrorParseParams, request), response)
	} else if !config.Enabled {
		nextParameter := appcfg.ScheduledAnalysisParameter{
			Enabled: false,
		}

		if val, err := types.NewJSONBObject(nextParameter); err != nil {
			api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusInternalServerError, fmt.Sprintf("failed to convert value to JSONBObject: %v", err), request), response)
		} else {
			updatedParameter := appcfg.Parameter{
				Key:   appcfg.ScheduledAnalysis,
				Value: val,
			}

			if err := s.db.SetConfigurationParameter(request.Context(), updatedParameter); err != nil {
				api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusInternalServerError, fmt.Sprintf("error updating database: %v", api.FormatDatabaseError(err)), request), response)
			}
		}
	} else {
		//Validate that the rrule is a good rule. We're going to require a DTSTART to keep scheduling consistent.
		//We're also going to reject UNTIL/COUNT because it will most likely break the pipeline once it's hit without being invalid
		if _, err := rrule.StrToRRule(config.RRule); err != nil {
			api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf(ErrorInvalidRrule, err), request), response)
		} else if strings.Contains(strings.ToUpper(config.RRule), "UNTIL") || strings.Contains(strings.ToUpper(config.RRule), "COUNT") {
			api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf(ErrorInvalidRrule, "count/until not supported"), request), response)
		} else if !strings.Contains(strings.ToUpper(config.RRule), "DTSTART") {
			api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf(ErrorInvalidRrule, "dtstart is required"), request), response)
		} else {
			nextParameter := appcfg.ScheduledAnalysisParameter{
				Enabled: true,
				RRule:   config.RRule,
			}

			if val, err := types.NewJSONBObject(nextParameter); err != nil {
				api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusInternalServerError, fmt.Sprintf("failed to convert value to JSONBObject: %v", err), request), response)
			} else {
				updatedParameter := appcfg.Parameter{
					Key:   appcfg.ScheduledAnalysis,
					Value: val,
				}

				if err := s.db.SetConfigurationParameter(request.Context(), updatedParameter); err != nil {
					api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusInternalServerError, fmt.Sprintf("error updating database: %v", api.FormatDatabaseError(err)), request), response)
				}
			}
		}
	}
}
