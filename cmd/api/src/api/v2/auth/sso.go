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

	"github.com/gorilla/mux"
	"github.com/specterops/bloodhound/log"
	"github.com/specterops/bloodhound/src/api"
	"github.com/specterops/bloodhound/src/model"
)

func (s ManagementResource) SSOLoginHandler(response http.ResponseWriter, request *http.Request) {
	ssoProviderSlug := mux.Vars(request)[api.URIPathVariableSSOProviderSlug]
	log.Debugf("HERE I AM IN LOGIN - provider %s", ssoProviderSlug)

	if ssoProvider, err := s.db.GetSSOProviderBySlug(request.Context(), ssoProviderSlug); err != nil {
		api.HandleDatabaseError(request, response, err)
	} else {
		switch ssoProvider.Type {
		case model.SessionAuthProviderSAML:
			//todo handle saml login
			api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusNotFound, api.ErrorResponseDetailsResourceNotFound, request), response)
		case model.SessionAuthProviderOIDC:
			if oidcProvider, err := s.db.GetOIDCProviderBySSOProviderID(request.Context(), ssoProvider.ID); err != nil {
				api.HandleDatabaseError(request, response, err)
			} else {
				s.OIDCLoginHandler(response, request, ssoProvider, oidcProvider)
			}
		default:
			api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusNotFound, api.ErrorResponseDetailsResourceNotFound, request), response)
		}
	}
}

func (s ManagementResource) SSOCallbackHandler(response http.ResponseWriter, request *http.Request) {
	ssoProviderSlug := mux.Vars(request)[api.URIPathVariableSSOProviderSlug]
	log.Debugf("HERE I AM IN CALLBACK - provider %s", ssoProviderSlug)

	if ssoProvider, err := s.db.GetSSOProviderBySlug(request.Context(), ssoProviderSlug); err != nil {
		api.HandleDatabaseError(request, response, err)
	} else {
		switch ssoProvider.Type {
		case model.SessionAuthProviderSAML:
			//todo handle saml callback
			api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusNotFound, api.ErrorResponseDetailsResourceNotFound, request), response)
		case model.SessionAuthProviderOIDC:
			if oidcProvider, err := s.db.GetOIDCProviderBySSOProviderID(request.Context(), ssoProvider.ID); err != nil {
				api.HandleDatabaseError(request, response, err)
			} else {
				s.OIDCCallbackHandler(response, request, ssoProvider, oidcProvider)
			}
		default:
			api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusNotFound, api.ErrorResponseDetailsResourceNotFound, request), response)
		}
	}
}
