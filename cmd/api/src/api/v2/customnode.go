package v2

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/specterops/bloodhound/src/api"
	"github.com/specterops/bloodhound/src/model"
	"net/http"
	"regexp"
)

const (
	CustomNodeKindParameter = "kind_name"
)

var (
	validIconName = regexp.MustCompile(`^fa-[a-z0-9-]+$`)
)

func (s *Resources) GetCustomNodeKinds(response http.ResponseWriter, request *http.Request) {
	if kinds, err := s.DB.GetCustomNodeKinds(request.Context()); err != nil {
		api.HandleDatabaseError(request, response, err)
	} else {
		api.WriteBasicResponse(request.Context(), kinds, http.StatusOK, response)
	}
}

func (s *Resources) GetCustomNodeKind(response http.ResponseWriter, request *http.Request) {
	var (
		paramId = mux.Vars(request)[CustomNodeKindParameter]
	)

	if kind, err := s.DB.GetCustomNodeKind(request.Context(), paramId); err != nil {
		api.HandleDatabaseError(request, response, err)
	} else {
		api.WriteBasicResponse(request.Context(), kind, http.StatusOK, response)
	}
}

type CreateCustomNodeRequest struct {
	CustomTypes map[string]model.CustomNodeKindConfig `json:"custom_types"`
}

func validateCreateCustomNodeRequest(customNodeKindRequest CreateCustomNodeRequest) error {
	for _, config := range customNodeKindRequest.CustomTypes {
		if err := validateConfig(config); err != nil {
			return err
		}
	}

	return nil
}

func validateConfig(config model.CustomNodeKindConfig) error {
	if config.Icon.Type != "font-awesome" {
		return fmt.Errorf("custom node kind config type (%s) is not supported", config.Icon.Type)
	} else if !validIconName.MatchString(config.Icon.Name) {
		return fmt.Errorf("custom node kind config name (%s) is not valid", config.Icon.Name)
	}

	return nil
}

func (s *Resources) CreateCustomNodeKind(response http.ResponseWriter, request *http.Request) {
	var (
		customNodeKindRequest CreateCustomNodeRequest
	)

	if err := json.NewDecoder(request.Body).Decode(&customNodeKindRequest); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, api.ErrorResponsePayloadUnmarshalError, request), response)
	} else if err := validateCreateCustomNodeRequest(customNodeKindRequest); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, api.ErrorResponseCodeBadRequest, request), response)
	} else if kinds, err := s.DB.CreateCustomNodeKinds(request.Context(), convertCreateCustomNodeRequest(customNodeKindRequest)); err != nil {
		api.HandleDatabaseError(request, response, err)
	} else {
		api.WriteBasicResponse(request.Context(), kinds, http.StatusCreated, response)
	}
}

func convertCreateCustomNodeRequest(request CreateCustomNodeRequest) []model.CustomNodeKind {
	var customNodeKinds []model.CustomNodeKind

	for key, val := range request.CustomTypes {
		customNodeKinds = append(customNodeKinds, model.CustomNodeKind{
			KindName: key,
			Config:   val,
		})
	}

	return customNodeKinds
}

type UpdateCustomNodeKindRequest struct {
	Config model.CustomNodeKindConfig `json:"config"`
}

func (s *Resources) UpdateCustomNodeKind(response http.ResponseWriter, request *http.Request) {
	var (
		paramId               = mux.Vars(request)[CustomNodeKindParameter]
		customNodeKindRequest UpdateCustomNodeKindRequest
	)

	if err := json.NewDecoder(request.Body).Decode(&customNodeKindRequest); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, api.ErrorResponsePayloadUnmarshalError, request), response)
	} else if err := validateConfig(customNodeKindRequest.Config); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, api.ErrorResponseCodeBadRequest, request), response)
	} else if kind, err := s.DB.UpdateCustomNodeKind(request.Context(), model.CustomNodeKind{KindName: paramId, Config: customNodeKindRequest.Config}); err != nil {
		api.HandleDatabaseError(request, response, err)
	} else {
		api.WriteBasicResponse(request.Context(), kind, http.StatusOK, response)
	}
}

func (s *Resources) DeleteCustomNodeKind(response http.ResponseWriter, request *http.Request) {
	var (
		paramId = mux.Vars(request)[CustomNodeKindParameter]
	)

	if err := s.DB.DeleteCustomNodeKind(request.Context(), paramId); err != nil {
		api.HandleDatabaseError(request, response, err)
	} else {
		response.WriteHeader(http.StatusOK)
	}
}
