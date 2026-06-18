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

	"github.com/specterops/bloodhound/cmd/api/src/api"
	"github.com/specterops/bloodhound/packages/go/bhlog/attr"
	"github.com/specterops/bloodhound/packages/go/responses"
	"github.com/specterops/bloodhound/server/appcfg/internal/services"
)

// Appcfg defines the appcfg service boundary for the appcfg handlers package.
type Appcfg interface {
	GetDatapipeStatus(context.Context) (services.DatapipeStatus, error)
}

// Handlers is a dependency injection container for appcfg handlers
type Handlers struct {
	appcfg Appcfg
}

// NewHandlersContainer initializes the Handlers dependency injection container
func NewHandlersContainer(appcfg Appcfg) *Handlers {
	return &Handlers{
		appcfg: appcfg,
	}
}

// GetDatapipeStatus returns the current status of the datapipe including analysis
// timestamps and next scheduled analysis time (Enterprise only).
func (s Handlers) GetDatapipeStatus(response http.ResponseWriter, request *http.Request) {
	var ctx = request.Context()

	status, err := s.appcfg.GetDatapipeStatus(ctx)
	if err != nil {
		handleAppcfgError(request, response, err)
		return
	}

	responses.WriteBasic(ctx, BuildDatapipeStatusView(status), http.StatusOK, response)
}

// handleAppcfgError maps sentinel errors from the service layer to HTTP status codes.
func handleAppcfgError(request *http.Request, response http.ResponseWriter, err error) {
	var ctx = request.Context()

	if errors.Is(err, services.ErrNotFound) {
		responses.WriteError(ctx, http.StatusNotFound, api.ErrorResponseDetailsResourceNotFound, response)
	} else if errors.Is(err, context.DeadlineExceeded) {
		responses.WriteError(ctx, http.StatusInternalServerError, api.ErrorResponseRequestTimeout, response)
	} else {
		slog.ErrorContext(ctx, "Unexpected database error", attr.Error(err))
		responses.WriteError(ctx, http.StatusInternalServerError, api.ErrorResponseDetailsInternalServerError, response)
	}
}
