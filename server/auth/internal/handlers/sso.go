// Copyright 2026 Specter Ops, Inc.
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

package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"path"

	"github.com/specterops/bloodhound/cmd/api/src/bhctx"
	"github.com/specterops/bloodhound/cmd/api/src/serde"
	"github.com/specterops/bloodhound/packages/go/params"
	"github.com/specterops/bloodhound/packages/go/responses"
	"github.com/specterops/bloodhound/server/auth/internal/services"
)

type AuthProviderView struct {
	ID          int32              `json:"id"`
	Name        string             `json:"name"`
	Type        string             `json:"type"`
	Slug        string             `json:"slug"`
	Details     any                `json:"details"`
	Config      ProviderConfigView `json:"config"`
	LoginURI    serde.URL          `json:"login_uri"`
	CallbackURI serde.URL          `json:"callback_uri"`
}

type SAMLProviderView struct {
	SAMLProviderDetailsView
	ServiceProviderIssuerURI     serde.URL `json:"sp_issuer_uri"`
	ServiceProviderInitiationURI serde.URL `json:"sp_sso_uri"`
	ServiceProviderMetadataURI   serde.URL `json:"sp_metadata_uri"`
	ServiceProviderACSURI        serde.URL `json:"sp_acs_uri"`
}

type AuthProviderListView []AuthProviderView

func (s AuthProviderListView) JSONView() ([]byte, error) {
	return json.Marshal(s)
}

func (s AuthProviderListView) ValidFilters() map[string]params.FilterableField {
	var (
		comparisonOperators = []params.FilterOperator{params.Equals, params.NotEquals, params.GreaterThan, params.GreaterThanOrEquals, params.LessThan, params.LessThanOrEquals}
		stringOperators     = []params.FilterOperator{params.Equals, params.NotEquals, params.ApproximatelyEquals}
	)

	return map[string]params.FilterableField{
		"id":         {Operators: comparisonOperators},
		"type":       {Operators: comparisonOperators},
		"name":       {Operators: stringOperators, IsStringData: true},
		"slug":       {Operators: stringOperators, IsStringData: true},
		"created_at": {Operators: comparisonOperators},
		"updated_at": {Operators: comparisonOperators},
	}
}

func (s AuthProviderListView) IsSortable(field string) bool {
	switch field {
	case "id", "name", "slug", "type", "created_at", "updated_at":
		return true
	default:
		return false
	}
}

func (s *Handlers) ListAuthProviders(response http.ResponseWriter, request *http.Request) {
	var (
		ctx   = request.Context()
		bhCtx = bhctx.Get(ctx)
	)

	if ssoProviders, err := s.service.ListAuthProviders(ctx, bhCtx.Filters, bhCtx.Sort); errors.Is(err, services.ErrInvalidProviderQuery) {
		responses.WriteError(ctx, http.StatusBadRequest, err.Error(), response)
	} else if err != nil {
		responses.WriteInternalServerError(ctx, err, response)
	} else {
		responses.WriteBasic(ctx, buildAuthProviderListView(ssoProviders, *bhCtx.Host), http.StatusOK, response)
	}
}

func buildAuthProviderListView(ssoProviders []services.SSOProvider, host url.URL) AuthProviderListView {
	var views AuthProviderListView

	for _, ssoProvider := range ssoProviders {
		var view = AuthProviderView{
			ID:     ssoProvider.ID,
			Name:   ssoProvider.Name,
			Slug:   ssoProvider.Slug,
			Type:   ssoProvider.Type.String(),
			Config: ProviderConfigView{AutoProvision: AutoProvisionConfigView(ssoProvider.Config.AutoProvision)},
		}

		view.formatProviderURLs(host)

		switch ssoProvider.Type {
		case services.ProviderOIDC:
			if details := ssoProvider.OIDCProvider; details != nil {
				view.Details = OIDCProviderView{
					ClientID:               details.ClientID,
					Issuer:                 details.Issuer,
					SSOProviderID:          details.SSOProviderID,
					ProviderTimestampsView: ProviderTimestampsView(details.ProviderTimestamps),
				}
			}
		case services.ProviderSAML:
			if details := ssoProvider.SAMLProvider; details != nil {
				var samlView = SAMLProviderView{
					SAMLProviderDetailsView: SAMLProviderDetailsView{
						Name:                       details.Name,
						DisplayName:                details.DisplayName,
						IssuerURI:                  details.IssuerURI,
						SingleSignOnURI:            details.SingleSignOnURI,
						RootURIVersion:             details.RootURIVersion,
						PrincipalAttributeMappings: details.PrincipalAttributeMappings,
						SSOProviderID:              details.SSOProviderID,
						ProviderTimestampsView:     ProviderTimestampsView(details.ProviderTimestamps),
					},
				}

				samlView.formatSAMLProviderURLs(host)
				view.Details = samlView
			}
		}

		views = append(views, view)
	}

	return views
}

func (s *AuthProviderView) formatProviderURLs(host url.URL) {
	var root = host

	root.Path = path.Join("/api/v2/sso", s.Slug)
	s.LoginURI = serde.FromURL(*root.JoinPath("login"))
	s.CallbackURI = serde.FromURL(*root.JoinPath("callback"))
}

func (s *SAMLProviderView) formatSAMLProviderURLs(host url.URL) {
	var root = host

	root.Path = path.Join(map[int]string{1: "/api/v1/login/saml", 2: "/api/v2/sso"}[s.RootURIVersion], s.Name)
	s.ServiceProviderIssuerURI = serde.FromURL(root)
	s.ServiceProviderInitiationURI = serde.FromURL(*root.JoinPath("login"))
	s.ServiceProviderMetadataURI = serde.FromURL(*root.JoinPath("metadata"))

	switch s.RootURIVersion {
	case 1:
		s.ServiceProviderACSURI = serde.FromURL(*root.JoinPath("acs"))
	case 2:
		s.ServiceProviderACSURI = serde.FromURL(*root.JoinPath("callback"))
	}
}
