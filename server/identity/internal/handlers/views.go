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
	"github.com/specterops/bloodhound/packages/go/params"
	"github.com/specterops/bloodhound/server/identity/internal/services"
)

// PermissionView is the JSON shape returned by the identity handlers for a
// permission. It is decoupled from services.Permission so the wire format can
// evolve independently of the domain model.
type PermissionView struct {
	Authority string       `json:"authority"`
	Name      string       `json:"name"`
	ID        int32        `json:"id"`
	CreatedAt time.Time    `json:"created_at"`
	UpdatedAt time.Time    `json:"updated_at"`
	DeletedAt sql.NullTime `json:"deleted_at"`
}

// BuildPermissionView projects a services.Permission into the view type the
// handlers return in their JSON envelope.
func BuildPermissionView(permission services.Permission) PermissionView {
	return PermissionView{
		Authority: permission.Authority,
		Name:      permission.Name,
		ID:        permission.ID,
		CreatedAt: permission.CreatedAt,
		UpdatedAt: permission.UpdatedAt,
		DeletedAt: permission.DeletedAt,
	}
}

// JSONView marshals the view to the byte slice expected by responses.WriteBasic,
// satisfying the responses.JSONViewer contract.
func (s PermissionView) JSONView() ([]byte, error) {
	return json.Marshal(s)
}

// RoleView is the JSON shape returned by the identity handlers for a role. It is
// decoupled from services.Role so the wire format can evolve independently of
// the domain model.
type RoleView struct {
	ID          int32            `json:"id"`
	Name        string           `json:"name"`
	Description string           `json:"description"`
	Permissions []PermissionView `json:"permissions"`
	CreatedAt   time.Time        `json:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at"`
	DeletedAt   sql.NullTime     `json:"deleted_at"`
}

// BuildRoleView projects a services.Role into the view type the handlers return
// in their JSON envelope.
func BuildRoleView(role services.Role) RoleView {
	var permissions = make([]PermissionView, 0, len(role.Permissions))
	for _, permission := range role.Permissions {
		permissions = append(permissions, BuildPermissionView(permission))
	}

	return RoleView{
		ID:          role.ID,
		Name:        role.Name,
		Description: role.Description,
		Permissions: permissions,
		CreatedAt:   role.CreatedAt,
		UpdatedAt:   role.UpdatedAt,
		DeletedAt:   role.DeletedAt,
	}
}

// JSONView marshals the view to the byte slice expected by responses.WriteBasic,
// satisfying the responses.JSONViewer contract.
func (s RoleView) JSONView() ([]byte, error) {
	return json.Marshal(s)
}

// RoleListView is the JSON shape returned by the identity handlers for a list of
// roles. It wraps the roles under a "roles" key so the payload matches the
// data.roles envelope the GET /api/v2/roles endpoint has always returned.
type RoleListView struct {
	Roles []RoleView `json:"roles"`
}

// BuildRoleListView projects a slice of services.Role into the list view the
// handlers return in their JSON envelope.
func BuildRoleListView(roles []services.Role) RoleListView {
	var views = make([]RoleView, 0, len(roles))
	for _, role := range roles {
		views = append(views, BuildRoleView(role))
	}

	return RoleListView{Roles: views}
}

// JSONView marshals the view to the byte slice expected by responses.WriteBasic,
// satisfying the responses.JSONViewer contract.
func (s RoleListView) JSONView() ([]byte, error) {
	return json.Marshal(s)
}

// ValidFilters implements params.Filterable, describing the role fields that may
// be filtered on and the operators each supports. It reproduces the legacy
// GET /api/v2/roles contract so the filter middleware validates identically.
func (s RoleListView) ValidFilters() map[string]params.FilterableField {
	var numericOperators = []params.FilterOperator{
		params.Equals,
		params.NotEquals,
		params.GreaterThan,
		params.GreaterThanOrEquals,
		params.LessThan,
		params.LessThanOrEquals,
	}

	return map[string]params.FilterableField{
		"name":       {Operators: []params.FilterOperator{params.Equals, params.NotEquals}, IsStringData: true},
		"id":         {Operators: numericOperators},
		"created_at": {Operators: numericOperators},
		"updated_at": {Operators: numericOperators},
	}
}

// IsSortable implements params.Sortable, reporting the role fields the sort
// middleware may order on. It reproduces the legacy GET /api/v2/roles contract.
func (s RoleListView) IsSortable(field string) bool {
	switch field {
	case "name", "description", "id", "created_at", "updated_at":
		return true
	default:
		return false
	}
}

// AuthSecretView is the JSON shape returned by the identity handlers for a user's
// auth secret. It reproduces the legacy model.AuthSecret contract, omitting the
// secret material (digest and TOTP secret) that was never on the wire.
type AuthSecretView struct {
	DigestMethod  string       `json:"digest_method"`
	ExpiresAt     time.Time    `json:"expires_at"`
	TOTPActivated bool         `json:"totp_activated"`
	ID            int32        `json:"id"`
	CreatedAt     time.Time    `json:"created_at"`
	UpdatedAt     time.Time    `json:"updated_at"`
	DeletedAt     sql.NullTime `json:"deleted_at"`
}

// BuildAuthSecretView projects a services.AuthSecret into the view type nested
// under a user in the handlers' JSON envelope.
func BuildAuthSecretView(authSecret services.AuthSecret) AuthSecretView {
	return AuthSecretView{
		DigestMethod:  authSecret.DigestMethod,
		ExpiresAt:     authSecret.ExpiresAt,
		TOTPActivated: authSecret.TOTPActivated,
		ID:            authSecret.ID,
		CreatedAt:     authSecret.CreatedAt,
		UpdatedAt:     authSecret.UpdatedAt,
		DeletedAt:     authSecret.DeletedAt,
	}
}

// EnvironmentAccessControlView is the JSON shape returned by the identity
// handlers for a single environment-targeted access control entry. It reproduces
// the legacy model.EnvironmentTargetedAccessControl contract.
type EnvironmentAccessControlView struct {
	UserID        string       `json:"user_id"`
	EnvironmentID string       `json:"environment_id"`
	ID            int64        `json:"id"`
	CreatedAt     time.Time    `json:"created_at"`
	UpdatedAt     time.Time    `json:"updated_at"`
	DeletedAt     sql.NullTime `json:"deleted_at"`
}

// BuildEnvironmentAccessControlView projects a services.EnvironmentAccessControl
// into the view type nested under a user in the handlers' JSON envelope.
func BuildEnvironmentAccessControlView(control services.EnvironmentAccessControl) EnvironmentAccessControlView {
	return EnvironmentAccessControlView{
		UserID:        control.UserID,
		EnvironmentID: control.EnvironmentID,
		ID:            control.ID,
		CreatedAt:     control.CreatedAt,
		UpdatedAt:     control.UpdatedAt,
		DeletedAt:     control.DeletedAt,
	}
}

// UserView is the JSON shape returned by the identity handlers for a user. It is
// decoupled from services.User so the wire format can evolve independently of the
// domain model. The nullable fields use pointer types so they marshal to a bare
// value or null, matching the legacy null.String/null.Int32 contract without
// coupling the slice to the cmd/api null package. The AuthSecret key retains its
// legacy capitalization because the legacy model.User field carried no json tag.
type UserView struct {
	SSOProviderID                    *int32                         `json:"sso_provider_id"`
	AuthSecret                       *AuthSecretView                `json:"AuthSecret"`
	Roles                            []RoleView                     `json:"roles"`
	FirstName                        *string                        `json:"first_name"`
	LastName                         *string                        `json:"last_name"`
	EmailAddress                     *string                        `json:"email_address"`
	PrincipalName                    string                         `json:"principal_name"`
	LastLogin                        time.Time                      `json:"last_login"`
	IsDisabled                       bool                           `json:"is_disabled"`
	AllEnvironments                  bool                           `json:"all_environments"`
	EnvironmentTargetedAccessControl []EnvironmentAccessControlView `json:"environment_targeted_access_control"`
	EULAAccepted                     bool                           `json:"eula_accepted"`
	ID                               uuid.UUID                      `json:"id"`
	CreatedAt                        time.Time                      `json:"created_at"`
	UpdatedAt                        time.Time                      `json:"updated_at"`
	DeletedAt                        sql.NullTime                   `json:"deleted_at"`
}

// BuildUserView projects a services.User into the view type the handlers return
// in their JSON envelope.
func BuildUserView(user services.User) UserView {
	var (
		roles      = make([]RoleView, 0, len(user.Roles))
		controls   = make([]EnvironmentAccessControlView, 0, len(user.EnvironmentTargetedAccessControl))
		authSecret *AuthSecretView
	)

	for _, role := range user.Roles {
		roles = append(roles, BuildRoleView(role))
	}

	for _, control := range user.EnvironmentTargetedAccessControl {
		controls = append(controls, BuildEnvironmentAccessControlView(control))
	}

	if user.AuthSecret != nil {
		var view = BuildAuthSecretView(*user.AuthSecret)
		authSecret = &view
	}

	return UserView{
		SSOProviderID:                    nullInt32ToPtr(user.SSOProviderID),
		AuthSecret:                       authSecret,
		Roles:                            roles,
		FirstName:                        nullStringToPtr(user.FirstName),
		LastName:                         nullStringToPtr(user.LastName),
		EmailAddress:                     nullStringToPtr(user.EmailAddress),
		PrincipalName:                    user.PrincipalName,
		LastLogin:                        user.LastLogin,
		IsDisabled:                       user.IsDisabled,
		AllEnvironments:                  user.AllEnvironments,
		EnvironmentTargetedAccessControl: controls,
		EULAAccepted:                     user.EULAAccepted,
		ID:                               user.ID,
		CreatedAt:                        user.CreatedAt,
		UpdatedAt:                        user.UpdatedAt,
		DeletedAt:                        user.DeletedAt,
	}
}

// JSONView marshals the view to the byte slice expected by responses.WriteBasic,
// satisfying the responses.JSONViewer contract.
func (s UserView) JSONView() ([]byte, error) {
	return json.Marshal(s)
}

// nullStringToPtr converts a sql.NullString into a *string so it marshals to a
// bare string or null, matching the legacy null.String wire contract.
func nullStringToPtr(value sql.NullString) *string {
	if !value.Valid {
		return nil
	}
	return &value.String
}

// nullInt32ToPtr converts a sql.NullInt32 into a *int32 so it marshals to a bare
// number or null, matching the legacy null.Int32 wire contract.
func nullInt32ToPtr(value sql.NullInt32) *int32 {
	if !value.Valid {
		return nil
	}
	return &value.Int32
}

// UserListView is the JSON shape returned by the identity handlers for a list of
// users. It wraps the users under a "users" key so the payload matches the
// data.users envelope the GET /api/v2/bloodhound-users and /api/v2/bhe-users
// endpoints have always returned.
type UserListView struct {
	Users []UserView `json:"users"`
}

// BuildUserListView projects a slice of services.User into the list view the
// handlers return in their JSON envelope.
func BuildUserListView(users []services.User) UserListView {
	var views = make([]UserView, 0, len(users))
	for _, user := range users {
		views = append(views, BuildUserView(user))
	}

	return UserListView{Users: views}
}

// JSONView marshals the view to the byte slice expected by responses.WriteBasic,
// satisfying the responses.JSONViewer contract.
func (s UserListView) JSONView() ([]byte, error) {
	return json.Marshal(s)
}

// ValidFilters implements params.Filterable, describing the user fields that may
// be filtered on and the operators each supports. It reproduces the legacy
// GET /api/v2/bloodhound-users contract so the filter middleware validates
// identically, with the exception of deleted_at: the users table has no
// deleted_at column, so filtering on it is rejected with a 400 rather than
// silently accepted.
func (s UserListView) ValidFilters() map[string]params.FilterableField {
	var (
		equalityOperators = []params.FilterOperator{params.Equals, params.NotEquals}
		numericOperators  = []params.FilterOperator{
			params.Equals,
			params.NotEquals,
			params.GreaterThan,
			params.GreaterThanOrEquals,
			params.LessThan,
			params.LessThanOrEquals,
		}
	)

	return map[string]params.FilterableField{
		"first_name":     {Operators: equalityOperators, IsStringData: true},
		"last_name":      {Operators: equalityOperators, IsStringData: true},
		"email_address":  {Operators: equalityOperators, IsStringData: true},
		"principal_name": {Operators: equalityOperators, IsStringData: true},
		"id":             {Operators: equalityOperators},
		"last_login":     {Operators: numericOperators},
		"created_at":     {Operators: numericOperators},
		"updated_at":     {Operators: numericOperators},
	}
}

// IsSortable implements params.Sortable, reporting the user fields the sort
// middleware may order on. It reproduces the legacy GET /api/v2/bloodhound-users
// contract, with the exception of deleted_at: the users table has no deleted_at
// column, so sorting on it is rejected with a 400 rather than silently accepted.
func (s UserListView) IsSortable(field string) bool {
	switch field {
	case "first_name", "last_name", "email_address", "principal_name", "last_login", "created_at", "updated_at":
		return true
	default:
		return false
	}
}
