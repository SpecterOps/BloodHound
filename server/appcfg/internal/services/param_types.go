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

// This file has the setup needed for operating on specific config types

package services

import (
	"encoding/json"
	"fmt"
	"time"

	iso8601 "github.com/channelmeter/iso8601duration"
)

// # Param keys

const (
	PasswordExpirationWindow            ParameterKey = "auth.password_expiration_window"
	SessionTTLHours                     ParameterKey = "auth.session_ttl_hours"
	Neo4jConfigs                        ParameterKey = "neo4j.configuration"
	CitrixRDPSupportKey                 ParameterKey = "analysis.citrix_rdp_support"
	PruneTTL                            ParameterKey = "prune.ttl"
	ReconciliationKey                   ParameterKey = "analysis.reconciliation"
	ScheduledAnalysis                   ParameterKey = "analysis.scheduled"
	ClientMetricsKey                    ParameterKey = "pipeline.client_metrics"
	APITokenExpiration                  ParameterKey = "auth.api_token_expiration"
	TrustedProxiesConfig                ParameterKey = "http.trusted_proxies"
	FedEULACustomTextKey                ParameterKey = "eula.custom_text"
	TierManagementParameterKey          ParameterKey = "analysis.tiering"
	AGTParameterKey                     ParameterKey = "analysis.tagging"
	StaleClientUpdatedLogicKey          ParameterKey = "pipeline.updated_stale_client"
	RetainIngestedFilesKey              ParameterKey = "analysis.retain_ingest_files"
	APITokens                           ParameterKey = "auth.api_tokens"
	TimeoutLimit                        ParameterKey = "api.timeout_limit"
	EnvironmentTargetedAccessControlKey ParameterKey = "auth.environment_targeted_access_control"
	SupportAccountProvisioningKey       ParameterKey = "auth.support_account_provisioning"
	GraphStorageOptimizationKey         ParameterKey = "analysis.graph_storage_optimization"
)

const (
	DefaultPasswordExpirationWindow = time.Hour * 24 * 90

	DefaultSessionTTLHours = 8

	DefaultPruneBaseTTL           = time.Hour * 24 * 7
	DefaultPruneHasSessionEdgeTTL = time.Hour * 24 * 3

	MaxDawgsWorkerLimit         = 6 // This is the maximum analysis parallel workers during tagging
	DefaultDawgsWorkerLimit     = 2 // This is the parallel workers during tagging
	DefaultExpansionWorkerLimit = 3 // This is the size of the expansion worker pool during tagging
	DefaultSelectorWorkerLimit  = 7 // This is the size of the selector worker pool during tagging
)

// Map is a convenience function for mapping the data stored in the Value Parameter struct member onto
// a richer type provided by the given value.
func (s *Parameter) Map(value any) error {
	return s.Value.Map(value)
}

// # Param types

// ISODuration wraps time.Duration for json unmarshalling from ISO duration values.
// Note durations from the ISO duration format are inexact.
type ISODuration time.Duration

func (s *ISODuration) UnmarshalJSON(b []byte) error {
	var durationString string
	if err := json.Unmarshal(b, &durationString); err != nil {
		return err
	}

	// Parse the ISO 8601 duration string (e.g., "PT1H30M")
	parsed, err := iso8601.FromString(durationString)
	if err != nil {
		return fmt.Errorf("invalid ISO 8601 duration: %w", err)
	}

	// Convert the parsed struct into a standard time.Duration
	*s = ISODuration(parsed.ToDuration())
	return nil
}

// ## Param types list

type POCGraphStorageOptimizationParam struct {
	AfterBoot          bool `json:"after_boot"`
	AfterAnalysis      bool `json:"after_analysis"`
	MinIntervalSeconds int  `json:"min_interval_seconds"`
}

type POCSupportAccountProvisioningParam struct {
	Enabled    bool        `json:"enabled,omitempty"`
	SessionTTL ISODuration `json:"session_ttl,omitempty"`
}

// # Param definitions

//   - allowAPIAccess: determines whether http API get/set operations
//     will be allowed on this key
//   - hydrationRules: generic to an individual param type, and determine
//     how the param is read from the DB:
type ParamTypeDefinition struct {
	allowAPIAccess bool // framed such that default value is false = protected
	hydrationRules any
}

// ParamTypeHydrationRules determine how a param type is transformed after being read from storage
//   - Default: the value applied when the DB has no value, an incorrect
//     value type, or encounters an error
//   - Normalize: an optional function to alter the received value to (for
//     instance) correct an out-of-bounds value.
type ParamTypeHydrationRules[ParamType any] struct {
	Normalize func(*ParamType)
	Default   ParamType
}

var (
	// paramTypeDefinitions, together with each param's struct definition,
	// are a static representation of how instances of this parameter type
	// should be interpreted from stored raw json or what operations should be allowed.
	//
	// The individual param definitions above `validate` tags cover validation applied
	// on update value API requests, and will return an API error if failed.
	// Note that we should never assume DB values will comply with these validate
	// rules as the rules can be enacted or change over time. Best practice is to
	// include both a validate rule and a normalization function.
	//
	// NOTE: only parameters whose consumers have been moved to onion architecture
	// (and are actually calling GetConfig) will have hydrationRules. Other parameters
	// will only exist in here to specify API access rules.
	//
	// IMPORTANT: keep defined keys and allowAPIAccess values in sync with
	// bhce/cmd/api/src/model/appcfg/parameter.go `IsValidKey and `IsProtectedKey` until
	// PUT /config has been migrated to onion
	paramTypeDefinitions = map[ParameterKey]ParamTypeDefinition{
		GraphStorageOptimizationKey: {
			allowAPIAccess: false,
			hydrationRules: ParamTypeHydrationRules[POCGraphStorageOptimizationParam]{
				Default: POCGraphStorageOptimizationParam{
					AfterBoot:          false,
					AfterAnalysis:      false,
					MinIntervalSeconds: 86400,
				},
				Normalize: func(g *POCGraphStorageOptimizationParam) {
					if g.MinIntervalSeconds < 0 {
						g.MinIntervalSeconds = 86400
					}
				},
			},
		},
		SupportAccountProvisioningKey: {
			allowAPIAccess: true,
			hydrationRules: ParamTypeHydrationRules[POCSupportAccountProvisioningParam]{
				Default: POCSupportAccountProvisioningParam{
					Enabled:    true,
					SessionTTL: ISODuration(time.Hour * 2),
				},
			},
		},
	}
)

// getParamTypeHydrationRules returns the specified hydration rule and a bool indicating success
func getParamTypeHydrationRules[ParamType any](key ParameterKey) (ParamTypeHydrationRules[ParamType], bool) {
	def, ok := paramTypeDefinitions[key].hydrationRules.(ParamTypeHydrationRules[ParamType])
	return def, ok
}
