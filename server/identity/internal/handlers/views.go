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
