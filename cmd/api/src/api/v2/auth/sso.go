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
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/specterops/bloodhound/cmd/api/src/api"
	"github.com/specterops/bloodhound/cmd/api/src/auth"
	"github.com/specterops/bloodhound/cmd/api/src/bhctx"
	"github.com/specterops/bloodhound/cmd/api/src/database/types/null"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/cmd/api/src/serde"
	"github.com/specterops/dawgs/cardinality"
)

// AuthProvider represents a unified SSO provider (either OIDC or SAML)
type AuthProvider struct {
	ID      int32                   `json:"id"`
	Name    string                  `json:"name"`
	Type    string                  `json:"type"`
	Slug    string                  `json:"slug"`
	Details interface{}             `json:"details"`
	Config  model.SSOProviderConfig `json:"config"`

	LoginUri    serde.URL `json:"login_uri"`
	CallbackUri serde.URL `json:"callback_uri"`
}

type getRoler interface {
	GetRole(ctx context.Context, roleID int32) (model.Role, error)
}

type getAllRoler interface {
	GetAllRoles(ctx context.Context, order string, filter model.SQLFilter) (model.Roles, error)
}

type jitUserUpserter interface {
	getAllRoler

	LookupUser(ctx context.Context, principalNameOrEmail string) (model.User, error)
	CreateUser(ctx context.Context, user model.User) (model.User, error)
	UpdateUser(ctx context.Context, user model.User) error
}

// DeleteSSOProviderResponse represents the response returned to the user from DeleteSSOProvider
type DeleteSSOProviderResponse struct {
	AffectedUsers model.Users `json:"affected_users"`
}

// DeleteSSOProvider deletes a sso_provider with the matching id
func (s ManagementResource) DeleteSSOProvider(response http.ResponseWriter, request *http.Request) {
	var (
		rawSSOProviderID = mux.Vars(request)[api.URIPathVariableSSOProviderID]
		requestContext   = bhctx.FromRequest(request)
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

func SanitizeAndGetRoles(ctx context.Context, autoProvisionConfig model.SSOProviderAutoProvisionConfig, maybeBHRoles []string, r getAllRoler) (model.Roles, error) {
	if dbRoles, err := r.GetAllRoles(ctx, "", model.SQLFilter{}); err != nil {
		return nil, err
	} else {
		var defaultRole model.Role
		dbRolesBySlug := make(map[string]*model.Role)
		// Make quick lookup by role slug -> lower cased, dashes for spaces, and prefixed by `bh` e.g. bh-power-user
		for _, r := range dbRoles {
			dbRolesBySlug[fmt.Sprintf("bh-%s", strings.ReplaceAll(strings.ToLower(r.Name), " ", "-"))] = &r
			if r.ID == autoProvisionConfig.DefaultRoleId {
				defaultRole = r
			}
		}

		if autoProvisionConfig.RoleProvision {
			var validRoles model.Roles
			validRolesSeen := cardinality.NewBitmap32() // Ensure no dupes
			// Only add valid roles
			for _, r := range maybeBHRoles {
				if dbRole := dbRolesBySlug[strings.ReplaceAll(strings.ToLower(r), " ", "-")]; dbRole != nil && !validRolesSeen.Contains(uint32(dbRole.ID)) {
					validRoles = append(validRoles, *dbRole)
					validRolesSeen.Add(uint32(dbRole.ID))
				}
			}
			switch {
			case len(validRoles) == 1:
				return validRoles, nil
			case len(validRoles) > 1:
				slog.WarnContext(
					ctx,
					"[SSO] JIT Role Provision detected multiple valid roles, falling back to default role",
					slog.String("valid_roles", strings.Join(validRoles.Names(), ",")),
					slog.String("default_role", defaultRole.Name),
				)
			default:
				slog.WarnContext(
					ctx,
					"[SSO] JIT Role Provision detected no valid roles, falling back to default role",
					slog.String("roles", strings.Join(maybeBHRoles, ",")),
					slog.String("default_role", defaultRole.Name),
				)
			}
		}

		/* Fallback to default role:
		- Role provision is disabled
		- Role provision is enabled but no valid roles are found
		- Role provision is enabled but multiple valid roles are found
		*/
		return model.Roles{defaultRole}, nil
	}
}
