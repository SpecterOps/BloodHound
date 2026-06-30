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

//go:generate go tool mockery

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/specterops/bloodhound/cmd/api/src/api"
	"github.com/specterops/bloodhound/packages/go/responses"
	"github.com/specterops/bloodhound/server/identity/internal/services"
)

// Identity defines the identity service boundary for the identity handlers package.
type Identity interface {
	GetRole(ctx context.Context, id int32) (services.Role, error)
	GetPermission(ctx context.Context, id int) (services.Permission, error)
}

// Handlers is a dependency injection container for identity handlers.
type Handlers struct {
	identity Identity
}

// NewHandlersContainer initializes the Handlers dependency injection container
func NewHandlersContainer(identity Identity) *Handlers {
	return &Handlers{
		identity: identity,
	}
}

// GetPermission returns the permission for the id in the request path.
func (s *Handlers) GetPermission(response http.ResponseWriter, request *http.Request) {
	var (
		ctx             = request.Context()
		rawPermissionID = mux.Vars(request)[api.URIPathVariablePermissionID]
	)

	permissionID, err := strconv.Atoi(rawPermissionID)
	if err != nil {
		responses.WriteError(ctx, http.StatusBadRequest, api.ErrorResponseDetailsIDMalformed, response)
		return
	}

	permission, err := s.identity.GetPermission(ctx, permissionID)
	if err != nil {
		handleIdentityError(request, response, err)
		return
	}

	responses.WriteBasic(ctx, BuildPermissionView(permission), http.StatusOK, response)
}

// GetRole returns the role for the id in the request path.
func (s *Handlers) GetRole(response http.ResponseWriter, request *http.Request) {
	var (
		ctx       = request.Context()
		rawRoleID = mux.Vars(request)[api.URIPathVariableRoleID]
	)

	roleID, err := strconv.ParseInt(rawRoleID, 10, 32)
	if err != nil {
		responses.WriteError(ctx, http.StatusBadRequest, api.ErrorResponseDetailsIDMalformed, response)
		return
	}

	role, err := s.identity.GetRole(ctx, int32(roleID))
	if err != nil {
		handleIdentityError(request, response, err)
		return
	}

	responses.WriteBasic(ctx, BuildRoleView(role), http.StatusOK, response)
}

func handleIdentityError(request *http.Request, response http.ResponseWriter, err error) {
	var ctx = request.Context()

	if errors.Is(err, services.ErrNoRoleFound) || errors.Is(err, services.ErrNoPermissionFound) {
		responses.WriteError(ctx, http.StatusNotFound, api.ErrorResponseDetailsResourceNotFound, response)
	} else if errors.Is(err, context.DeadlineExceeded) {
		responses.WriteError(ctx, http.StatusInternalServerError, api.ErrorResponseRequestTimeout, response)
	} else {
		responses.WriteInternalServerError(ctx, err, response)
	}
}
