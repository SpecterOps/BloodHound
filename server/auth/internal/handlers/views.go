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
	"database/sql"
	"encoding/json"
	"time"

	"github.com/gofrs/uuid"
	"github.com/specterops/bloodhound/cmd/api/src/auth"
	"github.com/specterops/bloodhound/packages/go/responses"
)

type UserView struct {
	ID                               uuid.UUID               `json:"id"`
	SSOProviderID                    *int32                  `json:"sso_provider_id"`
	AuthSecret                       *AuthSecretView         `json:"AuthSecret"`
	Roles                            []RoleView              `json:"roles"`
	FirstName                        *string                 `json:"first_name"`
	LastName                         *string                 `json:"last_name"`
	EmailAddress                     *string                 `json:"email_address"`
	PrincipalName                    string                  `json:"principal_name"`
	LastLogin                        time.Time               `json:"last_login"`
	IsDisabled                       bool                    `json:"is_disabled"`
	AllEnvironments                  bool                    `json:"all_environments"`
	EnvironmentTargetedAccessControl []EnvironmentAccessView `json:"environment_targeted_access_control"`
	EULAAccepted                     bool                    `json:"eula_accepted"`
	TimestampsView
}

// buildSelfView projects users onto the explicit user response schema and
// preserves other authentication owners such as Enterprise clients/collectors
func buildSelfView(authContext auth.Context) responses.JSONViewer {
	var (
		authSecret        *AuthSecretView
		roles             []RoleView
		environmentAccess []EnvironmentAccessView
	)

	if user, found := auth.GetUserFromAuthCtx(authContext); !found {
		return ownerView{authContext: authContext}
	} else {
		if user.AuthSecret != nil {
			authSecret = &AuthSecretView{
				ID:            user.AuthSecret.ID,
				CreatedAt:     user.AuthSecret.CreatedAt,
				UpdatedAt:     user.AuthSecret.UpdatedAt,
				DeletedAt:     user.AuthSecret.DeletedAt,
				DigestMethod:  user.AuthSecret.DigestMethod,
				ExpiresAt:     user.AuthSecret.ExpiresAt,
				TOTPActivated: user.AuthSecret.TOTPActivated,
			}
		}

		if user.Roles != nil {
			roles = make([]RoleView, 0, len(user.Roles))

			for _, role := range user.Roles {
				var permissions []PermissionView

				if role.Permissions != nil {
					permissions = make([]PermissionView, 0, len(role.Permissions))

					for _, permission := range role.Permissions {
						permissions = append(permissions, PermissionView{
							ID:        permission.ID,
							CreatedAt: permission.CreatedAt,
							UpdatedAt: permission.UpdatedAt,
							DeletedAt: permission.DeletedAt,
							Authority: permission.Authority,
							Name:      permission.Name,
						})
					}
				}

				roles = append(roles, RoleView{
					ID:          role.ID,
					CreatedAt:   role.CreatedAt,
					UpdatedAt:   role.UpdatedAt,
					DeletedAt:   role.DeletedAt,
					Name:        role.Name,
					Description: role.Description,
					Permissions: permissions,
				})
			}
		}

		if user.EnvironmentTargetedAccessControl != nil {
			environmentAccess = make([]EnvironmentAccessView, 0, len(user.EnvironmentTargetedAccessControl))

			for _, access := range user.EnvironmentTargetedAccessControl {
				environmentAccess = append(environmentAccess, EnvironmentAccessView{
					ID:            access.ID,
					CreatedAt:     access.CreatedAt,
					UpdatedAt:     access.UpdatedAt,
					DeletedAt:     access.DeletedAt,
					UserID:        access.UserID,
					EnvironmentID: access.EnvironmentID,
				})
			}
		}

		return &UserView{
			ID:                               user.ID,
			CreatedAt:                        user.CreatedAt,
			UpdatedAt:                        user.UpdatedAt,
			DeletedAt:                        user.DeletedAt,
			SSOProviderID:                    user.SSOProviderID.Ptr(),
			AuthSecret:                       authSecret,
			Roles:                            roles,
			FirstName:                        user.FirstName.Ptr(),
			LastName:                         user.LastName.Ptr(),
			EmailAddress:                     user.EmailAddress.Ptr(),
			PrincipalName:                    user.PrincipalName,
			LastLogin:                        user.LastLogin,
			IsDisabled:                       user.IsDisabled,
			AllEnvironments:                  user.AllEnvironments,
			EnvironmentTargetedAccessControl: environmentAccess,
			EULAAccepted:                     user.EULAAccepted,
		}
	}
}

// ownerView preserves non-user owner JSON, including null, without Enterprise dependencies
type ownerView struct {
	authContext auth.Context
}

func (s ownerView) JSONView() ([]byte, error) {
	return json.Marshal(s.authContext.Owner)
}

func (s *UserView) JSONView() ([]byte, error) {
	return json.Marshal(s)
}

type UserMinimalView struct {
	ID        int32        `json:"id"`
	CreatedAt time.Time    `json:"created_at"`
	UpdatedAt time.Time    `json:"updated_at"`
	DeletedAt sql.NullTime `json:"deleted_at"`
}

type TimestampsView struct {
	CreatedAt time.Time    `json:"created_at"`
	UpdatedAt time.Time    `json:"updated_at"`
	DeletedAt sql.NullTime `json:"deleted_at"`
}

type AuthSecretView struct {
	UserMinimalView
	DigestMethod  string    `json:"digest_method"`
	ExpiresAt     time.Time `json:"expires_at"`
	TOTPActivated bool      `json:"totp_activated"`
}

type RoleView struct {
	UserMinimalView
	Name        string           `json:"name"`
	Description string           `json:"description"`
	Permissions []PermissionView `json:"permissions"`
}

type PermissionView struct {
	UserMinimalView
	Authority string `json:"authority"`
	Name      string `json:"name"`
}

type EnvironmentAccessView struct {
	ID int64 `json:"id"`
	TimestampsView
	UserID        string `json:"user_id"`
	EnvironmentID string `json:"environment_id"`
}
