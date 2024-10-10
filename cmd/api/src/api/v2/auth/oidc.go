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

package auth

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/specterops/bloodhound/src/api"
	"github.com/specterops/bloodhound/src/database"
	"github.com/specterops/bloodhound/src/utils/validation"
)

// CreateOIDCProviderRequest represents the body of the CreateOIDCProvider endpoint
type CreateOIDCProviderRequest struct {
	Name     string `json:"name" validate:"required"`
	Issuer   string `json:"issuer"  validate:"url"`
	ClientID string `json:"client_id" validate:"required"`
}

// CreateOIDCProvider creates an OIDC provider entry given a valid request
func (s ManagementResource) CreateOIDCProvider(response http.ResponseWriter, request *http.Request) {
	var (
		createRequest = CreateOIDCProviderRequest{}
	)

	if err := api.ReadJSONRequestPayloadLimited(&createRequest, request); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, err.Error(), request), response)
	} else if validated := validation.Validate(createRequest); validated != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, validated.Error(), request), response)
	} else {
		if oidcProvider, err := s.db.CreateOIDCProvider(request.Context(), createRequest.Name, createRequest.Issuer, createRequest.ClientID); err != nil {
			api.HandleDatabaseError(request, response, err)
		} else {
			api.WriteBasicResponse(request.Context(), oidcProvider, http.StatusCreated, response)
		}
	}
}

// DeleteOIDCProvider deletes an OIDC Provider entry
func (s ManagementResource) DeleteOIDCProvider(response http.ResponseWriter, request *http.Request) {
	var (
		rawOIDCProviderID = mux.Vars(request)[api.URIPathVariableOIDCProviderID]
	)

	// Convert the incoming string url param to an int
	if oidcProviderID, err := strconv.Atoi(rawOIDCProviderID); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, err.Error(), request), response)
	} else if err = s.db.DeleteOIDCProvider(request.Context(), oidcProviderID); errors.Is(err, database.ErrNotFound) {
		// Handle error if requested record could not be found
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusNotFound, err.Error(), request), response)
	} else if err != nil {
		api.HandleDatabaseError(request, response, err)
	} else {
		api.WriteBasicResponse(request.Context(), "", http.StatusOK, response)
	}
}
