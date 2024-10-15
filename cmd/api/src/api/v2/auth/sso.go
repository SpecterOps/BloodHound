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
	"strings"

	"github.com/specterops/bloodhound/src/api"
	"github.com/specterops/bloodhound/src/model"
	"gorm.io/gorm/utils"
)

// AuthProvider represents a unified SSO provider (either OIDC or SAML)
type AuthProvider struct {
	ID      int32       `json:"id"`
	Name    string      `json:"name"`
	Type    string      `json:"type"`
	Slug    string      `json:"slug"`
	Details interface{} `json:"details"`
}

// ListAuthProviders lists all available SSO providers (SAML and OIDC) with sorting and filtering
func (s ManagementResource) ListAuthProviders(response http.ResponseWriter, request *http.Request) {
	var (
		ctx               = request.Context()
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
			api.WriteErrorResponse(ctx, api.BuildErrorResponse(http.StatusBadRequest, api.ErrorResponseDetailsNotSortable, request), response)
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
		api.WriteErrorResponse(ctx, api.BuildErrorResponse(http.StatusBadRequest, api.ErrorResponseDetailsBadQueryParameterFilters, request), response)
		return
	} else {
		for name, filters := range queryFilters {
			if validPredicates, err := model.SSOProviderValidFilterPredicates(name); err != nil {
				api.WriteErrorResponse(ctx, api.BuildErrorResponse(http.StatusBadRequest, err.Error(), request), response)
				return
			} else {
				for i, filter := range filters {
					if !utils.Contains(validPredicates, string(filter.Operator)) {
						api.WriteErrorResponse(ctx, api.BuildErrorResponse(http.StatusBadRequest, api.ErrorResponseDetailsFilterPredicateNotSupported, request), response)
						return
					}
					queryFilters[name][i].IsStringData = model.SSOProviderIsStringField(filter.Name)
				}
			}
		}

		if sqlFilter, err = queryFilters.BuildSQLFilter(); err != nil {
			api.WriteErrorResponse(ctx, api.BuildErrorResponse(http.StatusBadRequest, "error building SQL for filter", request), response)
			return
		} else if ssoProviders, err = s.db.GetAllSSOProviders(ctx, strings.Join(order, ", "), sqlFilter); err != nil {
			api.HandleDatabaseError(request, response, err)
		} else {
			for _, ssoProvider := range ssoProviders {
				provider := AuthProvider{
					ID:   ssoProvider.ID,
					Name: ssoProvider.Name,
					Type: ssoProvider.Type.String(),
					Slug: ssoProvider.Slug,
				}

				switch ssoProvider.Type {
				case model.SessionAuthProviderOIDC:
					if ssoProvider.OIDCProvider != nil {
						provider.Details = ssoProvider.OIDCProvider
					}
				case model.SessionAuthProviderSAML:
					if ssoProvider.SAMLProvider != nil {
						provider.Details = ssoProvider.SAMLProvider
					}
				}

				providers = append(providers, provider)
			}

			api.WriteBasicResponse(ctx, providers, http.StatusOK, response)
		}
	}
}
