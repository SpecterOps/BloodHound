// Copyright 2026 Specter Ops, Inc.
//
// Licensed under the Apache License, Version 2.0
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
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
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/specterops/bloodhound/cmd/api/src/api"
	"github.com/specterops/bloodhound/cmd/api/src/bhctx"
	"github.com/specterops/bloodhound/cmd/api/src/database/types"
	"github.com/specterops/bloodhound/cmd/api/src/database/types/null"
	"github.com/specterops/bloodhound/cmd/api/src/utils"
	"github.com/specterops/bloodhound/packages/go/bhlog/attr"
	"github.com/specterops/bloodhound/packages/go/params"
	"github.com/specterops/bloodhound/packages/go/responses"
	"github.com/specterops/bloodhound/server/appcfg/internal/services"
	"github.com/teambition/rrule-go"
)

//go:generate go tool mockery

type Service interface {
	GetDatapipeStatus(context.Context) (services.DatapipeStatus, error)
	GetApplicationConfiguration(ctx context.Context, parameterKey services.ParameterKey) (services.Parameter, error)
	GetAllApplicationConfigurations(ctx context.Context) (services.Parameters, error)
	UpsertApplicationConfiguration(ctx context.Context, parameterKey services.ParameterKey, value types.JSONBObject) error

	IsAPIAllowedKey(parameterKey services.ParameterKey) bool
	ValidateValue(key services.ParameterKey, value types.JSONBObject) utils.Errors
}

type FeatureFlagReader interface {
	IsEnabled(ctx context.Context, key string) (bool, error)
}

type DataPipeSideEffects interface {
	SetNextScheduledAnalysisStartTime(ctx context.Context, time null.Time) error
}

type AuthTokensSideEffects interface {
	UpdateAuthTokenExpiration(ctx context.Context) error
}

// Note ffReader, dataPipeSideEffects, and authTokensSideEffects are included
// for custom Update Config special handling of particular configs.
// I chose to inject them directly into handler rather than into service and having handler call
// service for this functionality. I did this to avoid adding extra dependencies
// to service when only this Update Config handler needs it, and other business
// logic (such as getting a config value) does not.
// This may not be a good pattern to repeat elsewhere.
type Handlers struct {
	service               Service
	ffReader              FeatureFlagReader
	dataPipeSideEffects   DataPipeSideEffects
	authTokensSideEffects AuthTokensSideEffects
}

func NewHandlers(service Service, ffReader FeatureFlagReader, dataPipeSideEffects DataPipeSideEffects, authTokensSideEffects AuthTokensSideEffects) *Handlers {
	return &Handlers{
		service:               service,
		ffReader:              ffReader,
		dataPipeSideEffects:   dataPipeSideEffects,
		authTokensSideEffects: authTokensSideEffects,
	}
}

func (s *Handlers) GetDatapipeStatus(response http.ResponseWriter, request *http.Request) {
	var ctx = request.Context()

	status, err := s.service.GetDatapipeStatus(ctx)
	if err != nil {
		handleServiceError(ctx, response, err)
		return
	}

	responses.WriteBasic(ctx, BuildDatapipeStatusView(status), http.StatusOK, response)
}

func (s *Handlers) GetApplicationConfiguration(response http.ResponseWriter, request *http.Request) {
	var (
		ctx   = request.Context()
		bhCtx = bhctx.Get(ctx)

		filterKey = ""
	)

	if paramFilter, ok := bhCtx.Filters["parameter"]; ok {
		if len(paramFilter) > 0 && paramFilter[0].Operator == params.Equals {
			filterKey = paramFilter[0].Value
		}
	}

	if filterKey != "" {
		paramKey := services.ParameterKey(filterKey)

		// Preserving existing behavior, where IsProtectedKey is only applied to
		// individual config get requests
		if !s.service.IsAPIAllowedKey(paramKey) {
			responses.WriteError(ctx, http.StatusBadRequest, fmt.Sprintf("Configuration parameter %s is not valid.", paramKey), response)
			return
		}
		param, err := s.service.GetApplicationConfiguration(ctx, services.ParameterKey(filterKey))
		if err != nil {
			handleServiceError(ctx, response, err)
			return
		}

		responses.WriteBasic(ctx, BuildParameterListView(services.Parameters{param}), http.StatusOK, response)
		return
	}

	params, err := s.service.GetAllApplicationConfigurations(ctx)
	if err != nil {
		handleServiceError(ctx, response, err)
		return
	}

	responses.WriteBasic(ctx, BuildParameterListView(params), http.StatusOK, response)
}

func (s *Handlers) SetApplicationConfiguration(response http.ResponseWriter, request *http.Request) {
	var ctx = request.Context()

	// read request body
	var setConfigBody UpdateParameterView
	err := json.NewDecoder(request.Body).Decode(&setConfigBody)
	if err != nil {
		responses.WriteError(ctx, http.StatusBadRequest, "error unmarshalling JSON payload", response)
		return
	}

	// validate writeable key
	if allowed := s.service.IsAPIAllowedKey(setConfigBody.Key); !allowed {
		responses.WriteError(ctx, http.StatusBadRequest, fmt.Sprintf("Configuration parameter %s is not valid.", setConfigBody.Key), response)
		return
	}

	// validate according to parameter type
	if errs := s.service.ValidateValue(setConfigBody.Key, setConfigBody.Value); errs != nil {
		responses.WriteError(ctx, http.StatusBadRequest, errs.Error(), response)
		return
	}

	// validate feature flag special case
	if setConfigBody.Key == services.APITokenExpiration {
		if enabled, err := s.ffReader.IsEnabled(ctx, "api_key_expiration_support"); err != nil {
			handleServiceError(ctx, response, err)
			return
		} else if !enabled {
			responses.WriteError(ctx, http.StatusForbidden, fmt.Sprintf("Configuration parameter %s is not available.", setConfigBody.Key), response)
			return
		}
	}

	// update the parameter
	if err := s.service.UpsertApplicationConfiguration(ctx, setConfigBody.Key, setConfigBody.Value); err != nil {
		handleServiceError(ctx, response, err)
		return
	}

	// execute custom key-specific side-effects
	if err := s.doSideEffects(ctx, setConfigBody.Key, setConfigBody.Value); err != nil {
		handleServiceError(ctx, response, err)
		return
	}

	// return new data
	responses.WriteBasic(ctx, setConfigBody, http.StatusOK, response)
}

// to preserve existing behavior, update scheduled analysis logs on error
// and auth token expiration will return an error code
func (s *Handlers) doSideEffects(ctx context.Context, key services.ParameterKey, value types.JSONBObject) error {
	switch key {
	case services.ScheduledAnalysis:
		s.updateNextScheduledAnalysisStartTime(ctx, value)
		return nil
	case services.APITokenExpiration:
		// Update the auth_tokens if the Parameter being changed/set is API Token Expiration
		if err := s.authTokensSideEffects.UpdateAuthTokenExpiration(ctx); err != nil {
			return fmt.Errorf("updating auth token expiration: %w", err)
		}
	}
	return nil
}

func (s *Handlers) updateNextScheduledAnalysisStartTime(ctx context.Context, value types.JSONBObject) {
	var scheduledAnalysis services.ScheduledAnalysisParameter
	if err := value.Map(&scheduledAnalysis); err != nil {
		slog.ErrorContext(ctx, "Error mapping scheduled analysis parameter", attr.Error(err))
	} else if scheduledAnalysis.Enabled {
		if rule, err := rrule.StrToRRule(scheduledAnalysis.RRule); err != nil {
			slog.ErrorContext(ctx, "Error parsing scheduled analysis rrule", attr.Error(err))
		} else if err := s.dataPipeSideEffects.SetNextScheduledAnalysisStartTime(ctx, null.TimeFrom(rule.After(time.Now(), true))); err != nil {
			slog.ErrorContext(ctx, "Error setting next scheduled analysis start time", attr.Error(err))
		}
	} else {
		if err := s.dataPipeSideEffects.SetNextScheduledAnalysisStartTime(ctx, null.Time{}); err != nil {
			slog.ErrorContext(ctx, "Error clearing the next scheduled analysis start time", attr.Error(err))
		}
	}
}

func handleServiceError(ctx context.Context, response http.ResponseWriter, err error) {
	if errors.Is(err, services.ErrNotFound) {
		responses.WriteError(ctx, http.StatusNotFound, api.ErrorResponseDetailsResourceNotFound, response)
	} else if errors.Is(err, context.DeadlineExceeded) {
		responses.WriteError(ctx, http.StatusInternalServerError, api.ErrorResponseRequestTimeout, response)
	} else {
		slog.ErrorContext(ctx, "Unexpected database error", attr.Error(err))
		responses.WriteError(ctx, http.StatusInternalServerError, api.ErrorResponseDetailsInternalServerError, response)
	}
}
