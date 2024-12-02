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
	"net/url"
	"path"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/specterops/bloodhound/src/api"
	"github.com/specterops/bloodhound/src/auth"
	"github.com/specterops/bloodhound/src/ctx"
	"github.com/specterops/bloodhound/src/database/types/null"
	"github.com/specterops/bloodhound/src/model"
	"github.com/specterops/bloodhound/src/serde"
	"gorm.io/gorm/utils"
)

// AuthProvider represents a unified SSO provider (either OIDC or SAML)
type AuthProvider struct {
	ID      int32       `json:"id"`
	Name    string      `json:"name"`
	Type    string      `json:"type"`
	Slug    string      `json:"slug"`
	Details interface{} `json:"details"`

	LoginUri    serde.URL `json:"login_uri"`
	CallbackUri serde.URL `json:"callback_uri"`
}

func (s *AuthProvider) FormatProviderURLs(hostUrl url.URL) {
	root := hostUrl
	root.Path = path.Join("/api/v2/sso/", s.Slug)

	s.LoginUri = serde.FromURL(*root.JoinPath("login"))
	s.CallbackUri = serde.FromURL(*root.JoinPath("callback"))
}

// ListAuthProviders lists all available SSO providers (SAML and OIDC) with sorting and filtering
func (s ManagementResource) ListAuthProviders(response http.ResponseWriter, request *http.Request) {
	var (
		requestCtx        = request.Context()
		queryParams       = request.URL.Query()
		sortByColumns     = queryParams[api.QueryParameterSortBy]
		order             []string
		queryFilters      model.QueryParameterFilterMap
		sqlFilter         model.SQLFilter
		ssoProviders      []model.SSOProvider
		providers         []AuthProvider
		err               error
		queryFilterParser = model.NewQueryParameterFilterParser()
	)

	for _, column := range sortByColumns {
		var descending bool
		if strings.HasPrefix(column, "-") {
			descending = true
			column = column[1:]
		}

		if !model.SSOProviderSortableFields(column) {
			api.WriteErrorResponse(requestCtx, api.BuildErrorResponse(http.StatusBadRequest, api.ErrorResponseDetailsNotSortable, request), response)
			return
		}

		if descending {
			order = append(order, column+" desc")
		} else {
			order = append(order, column)
		}
	}

	// Set default order by created_at if no sorting is specified
	if len(order) == 0 {
		order = append(order, "created_at")
	}

	if queryFilters, err = queryFilterParser.ParseQueryParameterFilters(request); err != nil {
		api.WriteErrorResponse(requestCtx, api.BuildErrorResponse(http.StatusBadRequest, api.ErrorResponseDetailsBadQueryParameterFilters, request), response)
	} else {
		for name, filters := range queryFilters {
			if validPredicates, err := model.SSOProviderValidFilterPredicates(name); err != nil {
				api.WriteErrorResponse(requestCtx, api.BuildErrorResponse(http.StatusBadRequest, err.Error(), request), response)
				return
			} else {
				for i, filter := range filters {
					if !utils.Contains(validPredicates, string(filter.Operator)) {
						api.WriteErrorResponse(requestCtx, api.BuildErrorResponse(http.StatusBadRequest, api.ErrorResponseDetailsFilterPredicateNotSupported, request), response)
						return
					}
					queryFilters[name][i].IsStringData = model.SSOProviderIsStringField(filter.Name)
				}
			}
		}

		if sqlFilter, err = queryFilters.BuildSQLFilter(); err != nil {
			api.WriteErrorResponse(requestCtx, api.BuildErrorResponse(http.StatusBadRequest, "error building SQL for filter", request), response)
		} else if ssoProviders, err = s.db.GetAllSSOProviders(requestCtx, strings.Join(order, ", "), sqlFilter); err != nil {
			api.HandleDatabaseError(request, response, err)
		} else {
			for _, ssoProvider := range ssoProviders {
				provider := AuthProvider{
					ID:   ssoProvider.ID,
					Name: ssoProvider.Name,
					Type: ssoProvider.Type.String(),
					Slug: ssoProvider.Slug,
				}

				// Format callback url from host
				provider.FormatProviderURLs(*ctx.Get(requestCtx).Host)

				switch ssoProvider.Type {
				case model.SessionAuthProviderOIDC:
					if ssoProvider.OIDCProvider != nil {
						provider.Details = ssoProvider.OIDCProvider
					}
				case model.SessionAuthProviderSAML:
					if ssoProvider.SAMLProvider != nil {
						ssoProvider.SAMLProvider.FormatSAMLProviderURLs(*ctx.Get(requestCtx).Host)
						provider.Details = ssoProvider.SAMLProvider
					}
				}

				providers = append(providers, provider)
			}

			api.WriteBasicResponse(requestCtx, providers, http.StatusOK, response)
		}
	}
}

// DeleteSSOProviderResponse represents the response returned to the user from DeleteSSOProvider
type DeleteSSOProviderResponse struct {
	AffectedUsers model.Users `json:"affected_users"`
}

// DeleteSSOProvider deletes a sso_provider with the matching id
func (s ManagementResource) DeleteSSOProvider(response http.ResponseWriter, request *http.Request) {
	var (
		rawSSOProviderID = mux.Vars(request)[api.URIPathVariableSSOProviderID]
		requestContext   = ctx.FromRequest(request)
	)

	// Convert the incoming string url param to an int
	if ssoProviderID, err := strconv.Atoi(rawSSOProviderID); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, err.Error(), request), response)
	} else if user, isUser := auth.GetUserFromAuthCtx(requestContext.AuthCtx); isUser && user.SSOProviderID.Equal(null.Int32From(int32(ssoProviderID))) {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusConflict, "user may not delete their own SSO auth provider", request), response)
	} else if providerUsers, err := s.db.GetSSOProviderUsers(request.Context(), ssoProviderID); err != nil {
		api.HandleDatabaseError(request, response, err)
	} else if err = s.db.DeleteSSOProvider(request.Context(), ssoProviderID); err != nil {
		api.HandleDatabaseError(request, response, err)
	} else {
		api.WriteJSONResponse(request.Context(), DeleteSSOProviderResponse{
			AffectedUsers: providerUsers,
		}, http.StatusOK, response)
	}
}

// UpdateSSOProvider updates a sso_provider with the matching id
func (s ManagementResource) UpdateSSOProvider(response http.ResponseWriter, request *http.Request) {
	rawSSOProviderID := mux.Vars(request)[api.URIPathVariableSSOProviderID]

	// Convert the incoming string url param to an int
	if ssoProviderID, err := strconv.Atoi(rawSSOProviderID); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, err.Error(), request), response)
	} else if ssoProvider, err := s.db.GetSSOProviderById(request.Context(), int32(ssoProviderID)); err != nil {
		api.HandleDatabaseError(request, response, err)
	} else {
		switch ssoProvider.Type {
		case model.SessionAuthProviderSAML:
			s.UpdateSAMLProviderRequest(response, request, ssoProvider)
		case model.SessionAuthProviderOIDC:
			s.UpdateOIDCProviderRequest(response, request, ssoProvider)
		default:
			api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusNotImplemented, api.ErrorResponseDetailsNotImplemented, request), response)
		}
	}
}

func (s ManagementResource) SSOLoginHandler(response http.ResponseWriter, request *http.Request) {
	ssoProviderSlug := mux.Vars(request)[api.URIPathVariableSSOProviderSlug]

	if ssoProvider, err := s.db.GetSSOProviderBySlug(request.Context(), ssoProviderSlug); err != nil {
		api.HandleDatabaseError(request, response, err)
	} else {
		switch ssoProvider.Type {
		case model.SessionAuthProviderSAML:
			s.SAMLLoginHandler(response, request, ssoProvider)
		case model.SessionAuthProviderOIDC:
			s.OIDCLoginHandler(response, request, ssoProvider)
		default:
			api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusNotImplemented, api.ErrorResponseDetailsNotImplemented, request), response)
		}
	}
}

func (s ManagementResource) SSOCallbackHandler(response http.ResponseWriter, request *http.Request) {
	ssoProviderSlug := mux.Vars(request)[api.URIPathVariableSSOProviderSlug]

	if ssoProvider, err := s.db.GetSSOProviderBySlug(request.Context(), ssoProviderSlug); err != nil {
		api.HandleDatabaseError(request, response, err)
	} else {
		switch ssoProvider.Type {
		case model.SessionAuthProviderSAML:
			s.SAMLCallbackHandler(response, request, ssoProvider)
		case model.SessionAuthProviderOIDC:
			s.OIDCCallbackHandler(response, request, ssoProvider)
		default:
			api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusNotImplemented, api.ErrorResponseDetailsNotImplemented, request), response)
		}
	}
}
