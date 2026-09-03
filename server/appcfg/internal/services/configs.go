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
	"database/sql"
	"time"

	"github.com/specterops/bloodhound/cmd/api/src/database/types"
)

type ParameterKey string

// Parameter is a runtime configuration parameter that can be fetched from the appcfg.ParameterService interface. The
// Value member is a DB-safe JSON type wrapper that can store arbitrary JSON objects and map them to golang struct
// definitions.
type Parameter struct {
	Key         ParameterKey
	Name        string
	Description string
	Value       types.JSONBObject

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt sql.NullTime
}

// Map is a convenience function for mapping the data stored in the Value Parameter struct member onto
// a richer type provided by the given value.
func (s *Parameter) Map(value any) error {
	return s.Value.Map(value)
}

func (s *Parameter) IsValidKey(parameterKey ParameterKey) bool {
	switch parameterKey {
	case PasswordExpirationWindow, Neo4jConfigs, PruneTTL, CitrixRDPSupportKey, ReconciliationKey, ScheduledAnalysis, ClientMetricsKey, APITokenExpiration:
		return true
	default:
		return false
	}
}

// IsProtectedKey These keys should not be updatable by users
func (s *Parameter) IsProtectedKey(parameterKey ParameterKey) bool {
	switch parameterKey {
	case TrustedProxiesConfig, FedEULACustomTextKey, TierManagementParameterKey, SessionTTLHours, StaleClientUpdatedLogicKey, RetainIngestedFilesKey, AGTParameterKey, TimeoutLimit, APITokens, EnvironmentTargetedAccessControlKey, SupportAccountProvisioningKey, GraphStorageOptimizationKey:
		return true
	default:
		return false
	}
}

// Parameters is a collection of Parameter structs.
type Parameters []Parameter
