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

// This file contains functions generic to all parameter types
package services

import (
	"context"
	"database/sql"
	"time"

	"github.com/specterops/bloodhound/cmd/api/src/database/types"
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
