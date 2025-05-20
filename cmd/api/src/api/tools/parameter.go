package tools

import (
	"fmt"
	"net/http"

	"github.com/specterops/bloodhound/src/api"
	"github.com/specterops/bloodhound/src/database/types"
	"github.com/specterops/bloodhound/src/model/appcfg"
)

type AppConfigUpdateRequest struct {
	Key   string         `json:"key"`
	Value map[string]any `json:"value"`
}

func convertAppConfigUpdateRequestToParameter(appConfigUpdateRequest AppConfigUpdateRequest) (appcfg.Parameter, error) {
	if value, err := types.NewJSONBObject(appConfigUpdateRequest.Value); err != nil {
		return appcfg.Parameter{}, fmt.Errorf("failed to convert value to JSONBObject: %w", err)
	} else {
		return appcfg.Parameter{
			Key:   appConfigUpdateRequest.Key,
			Value: value,
		}, nil
	}
}

func (s ToolContainer) SetApplicationParameter(response http.ResponseWriter, request *http.Request) {
	var (
		appConfig AppConfigUpdateRequest
	)

	if err := api.ReadJSONRequestPayloadLimited(&appConfig, request); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, api.ErrorResponsePayloadUnmarshalError, request), response)
	} else if parameter, err := convertAppConfigUpdateRequestToParameter(appConfig); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf("Configuration update request not converted to a parameter: %v", parameter), request), response)
	} else if err = s.db.SetConfigurationParameter(request.Context(), parameter); err != nil {
		api.HandleDatabaseError(request, response, err)
	} else {
		api.WriteBasicResponse(request.Context(), appConfig, http.StatusOK, response)
	}
}
