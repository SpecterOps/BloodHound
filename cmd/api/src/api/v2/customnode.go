package v2

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/specterops/bloodhound/src/api"
	"github.com/specterops/bloodhound/src/model"
	"net/http"
	"strconv"
)

const (
	CustomNodeKindParameter = "kind_id"
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

	if id, err := strconv.ParseInt(paramId, 10, 32); err == nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusNotFound, api.ErrorResponseDetailsResourceNotFound, request), response)
	} else if kind, err := s.DB.GetCustomNodeKind(request.Context(), int32(id)); err != nil {
		api.HandleDatabaseError(request, response, err)
	} else {
		api.WriteBasicResponse(request.Context(), kind, http.StatusOK, response)
	}
}

type CreateCustomNodeRequest struct {
	KindID int16                      `json:"kind_id"`
	Config model.CustomNodeKindConfig `json:"config"`
}

func (s *Resources) CreateCustomNodeKind(response http.ResponseWriter, request *http.Request) {
	var (
		customNodeKindRequest CreateCustomNodeRequest
	)

	if err := json.NewDecoder(request.Body).Decode(&customNodeKindRequest); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, api.ErrorResponsePayloadUnmarshalError, request), response)
	} else if kind, err := s.DB.CreateCustomNodeKind(request.Context(), model.CustomNodeKind{KindID: customNodeKindRequest.KindID, Config: customNodeKindRequest.Config}); err != nil {
		api.HandleDatabaseError(request, response, err)
	} else {
		api.WriteBasicResponse(request.Context(), kind, http.StatusCreated, response)
	}
}

type UpdateCustomNodeKindRequest struct {
	KindID int16                      `json:"kind_id"`
	Config model.CustomNodeKindConfig `json:"config"`
}

func (s *Resources) UpdateCustomNodeKind(response http.ResponseWriter, request *http.Request) {
	var (
		paramId               = mux.Vars(request)[CustomNodeKindParameter]
		customNodeKindRequest UpdateCustomNodeKindRequest
	)

	if id, err := strconv.ParseInt(paramId, 10, 32); err == nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusNotFound, api.ErrorResponseDetailsResourceNotFound, request), response)
	} else if err := json.NewDecoder(request.Body).Decode(&customNodeKindRequest); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, api.ErrorResponsePayloadUnmarshalError, request), response)
	} else if kind, err := s.DB.UpdateCustomNodeKind(request.Context(), model.CustomNodeKind{ID: int32(id), KindID: customNodeKindRequest.KindID, Config: customNodeKindRequest.Config}); err != nil {
		api.HandleDatabaseError(request, response, err)
	} else {
		api.WriteBasicResponse(request.Context(), kind, http.StatusOK, response)
	}
}

func (s *Resources) DeleteCustomNodeKind(response http.ResponseWriter, request *http.Request) {
	var (
		paramId = mux.Vars(request)[CustomNodeKindParameter]
	)

	if id, err := strconv.ParseInt(paramId, 10, 32); err == nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusNotFound, api.ErrorResponseDetailsResourceNotFound, request), response)
	} else if err := s.DB.DeleteCustomNodeKind(request.Context(), int32(id)); err != nil {
		api.HandleDatabaseError(request, response, err)
	} else {
		response.WriteHeader(http.StatusOK)
	}
}
