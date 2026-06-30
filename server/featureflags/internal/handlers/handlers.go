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
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/specterops/bloodhound/cmd/api/src/api"
	"github.com/specterops/bloodhound/packages/go/bhlog/attr"
	"github.com/specterops/bloodhound/packages/go/responses"
	"github.com/specterops/bloodhound/server/featureflags/internal/services"
)

// FeatureFlag defines the feature flag service boundary for the feature flag handlers package.
type FeatureFlag interface {
	GetAllFlags(ctx context.Context) ([]services.FeatureFlag, error)
	ToggleFlag(ctx context.Context, id int32) (services.FeatureFlag, error)
	IsEnabled(ctx context.Context, key string) (bool, error)
}

// Handlers is a dependency injection container for featureflags handlers
type Handlers struct {
	featureFlag FeatureFlag
}

// NewHandlersContainer initializes the featureflags Handlers dependency injection container
func NewHandlersContainer(featureFlag FeatureFlag) *Handlers {
	return &Handlers{
		featureFlag: featureFlag,
	}
}

// GetAllFlags returns the full list of feature flags as a JSON response. The
// handler delegates to the service layer to load the flags and serializes the
// result using the package's view builder.
func (s Handlers) GetAllFlags(response http.ResponseWriter, request *http.Request) {
	var ctx = request.Context()

	flags, err := s.featureFlag.GetAllFlags(ctx)
	if err != nil {
		handleFeatureFlagError(request, response, err)
		return
	}

	responses.WriteBasic(ctx, BuildFeatureFlagsView(flags), http.StatusOK, response)
}

// ToggleFlag toggles the Enabled state of the feature flag identified by the
// feature_id path parameter. The handler delegates to the service layer, which
// loads the existing flag, validates that it is user-updatable, flips its
// Enabled value and persists the result.
//
// Authentication is enforced by the route middleware (RequirePermissions); if no
// user is present on the auth context here it indicates an unexpected internal
// state and a 500 is returned.
func (s Handlers) ToggleFlag(response http.ResponseWriter, request *http.Request) {
	var ctx = request.Context()

	rawFeatureID := mux.Vars(request)[api.URIPathVariableFeatureID]

	featureID, err := strconv.ParseInt(rawFeatureID, 10, 32)
	if err != nil {
		responses.WriteError(ctx, http.StatusBadRequest, api.ErrorResponseDetailsIDMalformed, response)
		return
	}

	flag, err := s.featureFlag.ToggleFlag(ctx, int32(featureID))
	if err != nil {
		handleFeatureFlagError(request, response, err)
		return
	}

	responses.WriteBasic(ctx, BuildFeatureFlagView(flag), http.StatusOK, response)
}

// IsEnabled reports whether the feature flag identified by key is currently
// enabled. It is intended for in-process callers that need to gate behavior on
// a feature flag without going through the HTTP layer.
func (s Handlers) IsEnabled(ctx context.Context, key string) (bool, error) {
	return s.featureFlag.IsEnabled(ctx, key)
}

// handleFeatureFlagError maps service-layer errors to HTTP responses, translating
// known sentinel errors to their corresponding status codes and falling back to
// a logged 500 for anything unexpected.
func handleFeatureFlagError(request *http.Request, response http.ResponseWriter, err error) {
	if errors.Is(err, services.ErrNotFound) {
		responses.WriteError(request.Context(), http.StatusNotFound, services.ErrNotFound.Error(), response)
	} else if errors.Is(err, services.ErrNotUserUpdatable) {
		responses.WriteError(request.Context(), http.StatusForbidden, services.ErrNotUserUpdatable.Error(), response)
	} else {
		slog.Error("Unexpected database error", attr.Error(err))
		responses.WriteError(request.Context(), http.StatusInternalServerError, api.ErrorResponseDetailsInternalServerError, response)
	}
}
