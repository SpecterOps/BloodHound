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

// This file contains appconfig functionality generic to all parameter types

package services

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	"github.com/specterops/bloodhound/cmd/api/src/database/types"
	"github.com/specterops/bloodhound/packages/go/bhlog/attr"
)

type ParameterKey string

// Parameter is a runtime configuration parameter that can be fetched from the appcfg.ParameterService interface. The
// Value member is a DB-safe JSON type wrapper that can store arbitrary JSON objects and map them to golang struct
// definitions.
type Parameter struct {
	ID          int32 // should not be necessary, but part of the API interface so must include
	Key         ParameterKey
	Name        string
	Description string
	Value       types.JSONBObject

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt sql.NullTime
}

func (s Service) IsValidKey(parameterKey ParameterKey) bool {
	switch parameterKey {
	case PasswordExpirationWindow, Neo4jConfigs, PruneTTL, CitrixRDPSupportKey, ReconciliationKey, ScheduledAnalysis, ClientMetricsKey, APITokenExpiration:
		return true
	default:
		return false
	}
}

// IsProtectedKey These keys should not be updatable by users
func (s Service) IsProtectedKey(parameterKey ParameterKey) bool {
	switch parameterKey {
	case TrustedProxiesConfig, FedEULACustomTextKey, TierManagementParameterKey, SessionTTLHours, StaleClientUpdatedLogicKey, RetainIngestedFilesKey, AGTParameterKey, TimeoutLimit, APITokens, EnvironmentTargetedAccessControlKey, SupportAccountProvisioningKey, GraphStorageOptimizationKey:
		return true
	default:
		return false
	}
}

// Parameters is a collection of Parameter structs.
type Parameters []Parameter

func (s Service) GetApplicationConfiguration(ctx context.Context, parameterKey ParameterKey) (Parameter, error) {
	return s.db.GetConfigurationParameter(ctx, parameterKey)
}

func (s Service) GetAllApplicationConfigurations(ctx context.Context) (Parameters, error) {
	return s.db.GetAllConfigurationParameters(ctx)
}

type GetConfigError struct {
	Err            error
	ParameterKey   ParameterKey
	AppliedDefault bool
}

func (s GetConfigError) Error() string {
	return fmt.Sprintf("failed to fetch configuration %s, applied default: %t, error: %v", string(s.ParameterKey), s.AppliedDefault, s.Err)
}

func (s GetConfigError) Unwrap() error { return s.Err }

func (s GetConfigError) slogWarn(ctx context.Context) {
	slog.WarnContext(ctx, "Failed to fetch configuration",
		attr.Error(s.Err),
		slog.String("parameter_key", string(s.ParameterKey)),
		slog.Bool("applied_default", s.AppliedDefault),
	)
}

// GetConfig returns a parameter of the specified type with the given key.
// GetConfig returns a GetConfigError without a default if the key is not defined
// or of the wrong type.
// If any other fetch problem happens, GetConfig logs a warning and returns the
// parameter's default value and a GetConfigError.
// Note once we update to Go 1.27 this can have a Service receiver
func GetConfig[ParamType any](ctx context.Context, s *Service, key ParameterKey) (ParamType, error) {
	var result ParamType

	paramDefinition, ok := getParamTypeHydrationRules[ParamType](key)
	if !ok {
		return result, GetConfigError{
			Err:            fmt.Errorf("key did not exist or did not match ParamType"),
			ParameterKey:   key,
			AppliedDefault: false,
		}
	}

	result = paramDefinition.Default

	// read parameter from the database based on Key
	// get value, read into ParamType
	if cfg, err := s.db.GetConfigurationParameter(ctx, key); err != nil {
		getConfigErr := GetConfigError{
			Err:            err,
			ParameterKey:   key,
			AppliedDefault: true,
		}
		getConfigErr.slogWarn(ctx)
		return result, getConfigErr
	} else if err := cfg.Map(&result); err != nil {
		getConfigErr := GetConfigError{
			Err:            err,
			ParameterKey:   key,
			AppliedDefault: true,
		}
		getConfigErr.slogWarn(ctx)
		return result, getConfigErr
	}

	if paramDefinition.Normalize != nil {
		paramDefinition.Normalize(&result)
	}

	return result, nil
}
