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
	"sort"

	"github.com/specterops/bloodhound/src/api"
	"github.com/specterops/bloodhound/src/auth/bhsaml"
	"github.com/specterops/bloodhound/src/model"
)

// AuthProvider represents a unified SSO provider (either OIDC or SAML)
type AuthProvider struct {
	ID      int64       `json:"id"`
	Name    string      `json:"name"`
	Type    string      `json:"type"`
	Slug    string      `json:"slug"`
	Details interface{} `json:"details"`
}

// ListAuthProviders lists all available SSO providers (SAML and OIDC)
func (s ManagementResource) ListAuthProviders(response http.ResponseWriter, request *http.Request) {

	ctx := request.Context()

	providers := []AuthProvider{}

	// Fetch all SSO providers
	ssoProviders, err := s.db.GetAllSSOProviders(ctx)
	if err != nil {
		api.HandleDatabaseError(request, response, err)
		return
	}

	for _, ssoProvider := range ssoProviders {
		var details interface{}

		switch ssoProvider.Type {
		case model.SessionAuthProviderOIDC:
			oidcProvider, err := s.db.GetOIDCProviderBySSOProviderID(ctx, int(ssoProvider.ID))
			if err != nil {
				api.HandleDatabaseError(request, response, err)
				return
			}
			details = oidcProvider
		case model.SessionAuthProviderSAML:
			samlProvider, err := bhsaml.GetSAMLProviderBySSOProviderID(s.db, ssoProvider.ID, ctx)
			if err != nil {
				api.HandleDatabaseError(request, response, err)
				return
			}
			details = samlProvider
		default:
			continue
		}

		provider := AuthProvider{
			ID:      int64(ssoProvider.ID),
			Name:    ssoProvider.Name,
			Type:    ssoProvider.Type.String(),
			Slug:    ssoProvider.Slug,
			Details: details,
		}
		providers = append(providers, provider)
	}

	// Sort providers alphabetically by Name
	sort.Slice(providers, func(i, j int) bool {
		return providers[i].Name < providers[j].Name
	})

	// Return the combined list
	api.WriteBasicResponse(ctx, providers, http.StatusOK, response)
}
