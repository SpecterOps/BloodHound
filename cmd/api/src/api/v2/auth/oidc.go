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
	"net/http"

	"github.com/specterops/bloodhound/src/utils/validation"

	"github.com/specterops/bloodhound/src/api"
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
