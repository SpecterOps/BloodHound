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
	"time"

	"github.com/specterops/bloodhound/packages/go/duration"

	"github.com/specterops/dawgs/drivers/neo4j"
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
	CustomSubClaimKey                   ParameterKey = "auth.custom_sub_claim_key"
)

// Map is a convenience function for mapping the data stored in the Value Parameter struct member onto
// a richer type provided by the given value.
func (s *Parameter) Map(value any) error {
	return s.Value.Map(value)
}

// # Param types

// ## Param types list

type PasswordExpirationParameter struct {
	Duration duration.ISODuration `json:"duration"`
}

type Neo4jParameter struct {
	WriteFlushSize int `json:"write_flush_size,omitempty"`
	BatchWriteSize int `json:"batch_write_size,omitempty"`
}

type CitrixRDPSupportParameter struct {
	Enabled bool `json:"enabled,omitempty"`
}

type PruneTTLParameter struct {
	BaseTTL           duration.ISODuration `json:"base_ttl,omitempty" validate:"duration,min=P4D,max=P30D"`
	HasSessionEdgeTTL duration.ISODuration `json:"has_session_edge_ttl,omitempty" validate:"duration,min=P2D,max=P7D"`
}

type ReconciliationParameter struct {
	Enabled bool `json:"enabled,omitempty"`
}

type ScheduledAnalysisParameter struct {
	Enabled bool   `json:"enabled,omitempty"`
	RRule   string `json:"rrule,omitempty" validate:"rrule"`
}

type TrustedProxiesParameters struct {
	TrustedProxies int `json:"trusted_proxies,omitempty"`
}

type TieringParameter struct {
	TierLimit                int  `json:"tier_limit,omitempty"`
	LabelLimit               int  `json:"label_limit,omitempty"`
	MultiTierAnalysisEnabled bool `json:"multi_tier_analysis_enabled,omitempty"`
}

type AGTParameter struct {
	DAWGsWorkerLimit     int `json:"dawgs_worker_limit,omitempty"`
	ExpansionWorkerLimit int `json:"expansion_worker_limit,omitempty"`
	SelectorWorkerLimit  int `json:"selector_worker_limit,omitempty"`
}

type FedEULACustomTextParameter struct {
	CustomText string `json:"custom_text,omitempty"`
}

type SessionTTLHoursParameter struct {
	Hours int `json:"hours,omitempty"`
}

type StaleClientUpdatedLogicParameter struct {
	Enabled bool `json:"enabled,omitempty"`
}

type RetainIngestedFilesParameter struct {
	Enabled bool `json:"enabled,omitempty"`
}

type TimeoutLimitParameter struct {
	Enabled bool `json:"enabled,omitempty"`
}

type APITokensParameter struct {
	Enabled bool `json:"enabled"`
}

type EnvironmentTargetedAccessControlParameter struct {
	Enabled bool `json:"enabled,omitempty"`
}

type SupportAccountProvisioningParameter struct {
	// Setting disabled as false means that you are explicitly opting into the feature
	Enabled    bool                 `json:"enabled,omitempty"`
	SessionTTL duration.ISODuration `json:"session_ttl,omitempty"`
}

type ClientMetricsParameter struct {
	Enabled bool `json:"enabled,omitempty"`
}

type APITokenExpirationParameter struct {
	Enabled          bool `json:"enabled"`
	ExpirationPeriod int  `json:"expiration_period" validate:"integer,min=1,max=365"`
}

type GraphStorageOptimizationParameter struct {
	AfterBoot          bool `json:"after_boot"`
	AfterAnalysis      bool `json:"after_analysis"`
	MinIntervalSeconds int  `json:"min_interval_seconds"`
}

type CustomSubClaimKeyParameter struct {
	ClaimKey string `json:"claim_key"`
}

const (
	defaultSessionTTLHours                         = 8
	defaultAGTDAWGsWorkerLimit                     = 2
	maxAGTDAWGsWorkerLimit                         = 6
	defaultAGTExpansionWorkerLimit                 = 3
	defaultAGTSelectorWorkerLimit                  = 7
	defaultAPITokenExpirationPeriod                = 90
	maxAPITokenExpirationPeriod                    = 365
	defaultGraphStorageOptimizationMinIntervalSecs = 86400
)

// # Param definitions

//   - AllowAPIAccess: determines whether http API get/set operations
//     will be allowed on this key
//   - GetType: returns an empty instance of the type corresponding to
//     this key
//   - HydrationRules: generic to an individual param type, and determines
//     how the param is read from the DB
type ParamTypeDefinition struct {
	AllowAPIAccess bool // framed such that default value is false = protected
	GetType        func() any
	HydrationRules any
}

// ParamTypeHydrationRules determine how a param type is transformed after being read from storage
//   - Default: the value applied when the DB has no value, an incorrect
//     value type, or encounters an error
//   - Normalize: an optional function to alter the received value to (for
//     instance) correct an out-of-bounds value
type ParamTypeHydrationRules[ParamType any] struct {
	Default   ParamType
	Normalize func(*ParamType)
}

var (
	// ParamTypeDefinitions, together with each param's struct definition,
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
	// This is exported to allow test overrides
	ParamTypeDefinitions = map[ParameterKey]ParamTypeDefinition{
		// Example:
		// GraphStorageOptimizationKey: {
		// 	AllowAPIAccess: false,
		//  GetType: func() any { return GraphStorageOptimizationParameter{} },
		// 	HydrationRules: ParamTypeHydrationRules[POCGraphStorageOptimizationParam]{
		// 		Default: POCGraphStorageOptimizationParam{
		// 			AfterBoot:          false,
		// 			AfterAnalysis:      false,
		// 			MinIntervalSeconds: 86400,
		// 		},
		// 		Normalize: func(g *POCGraphStorageOptimizationParam) {
		// 			if g.MinIntervalSeconds < 0 {
		// 				g.MinIntervalSeconds = 86400
		// 			}
		// 		},
		// 	},
		// },
		PasswordExpirationWindow: {
			AllowAPIAccess: true,
			GetType:        func() any { return PasswordExpirationParameter{} },
			HydrationRules: ParamTypeHydrationRules[PasswordExpirationParameter]{
				Default: PasswordExpirationParameter{
					Duration: duration.ISODuration(90 * 24 * time.Hour),
				},
			},
		},
		Neo4jConfigs: {
			AllowAPIAccess: true,
			GetType:        func() any { return Neo4jParameter{} },
			HydrationRules: ParamTypeHydrationRules[Neo4jParameter]{
				Default: Neo4jParameter{
					WriteFlushSize: neo4j.DefaultWriteFlushSize,
					BatchWriteSize: neo4j.DefaultBatchWriteSize,
				},
			},
		},
		PruneTTL: {
			AllowAPIAccess: true,
			GetType:        func() any { return PruneTTLParameter{} },
			HydrationRules: ParamTypeHydrationRules[PruneTTLParameter]{
				Default: PruneTTLParameter{
					BaseTTL:           duration.ISODuration(7 * 24 * time.Hour),
					HasSessionEdgeTTL: duration.ISODuration(3 * 24 * time.Hour),
				},
			},
		},
		CitrixRDPSupportKey: {
			AllowAPIAccess: true,
			GetType:        func() any { return CitrixRDPSupportParameter{} },
			HydrationRules: ParamTypeHydrationRules[CitrixRDPSupportParameter]{
				Default: CitrixRDPSupportParameter{},
			},
		},
		ReconciliationKey: {
			AllowAPIAccess: true,
			GetType:        func() any { return ReconciliationParameter{} },
			HydrationRules: ParamTypeHydrationRules[ReconciliationParameter]{
				Default: ReconciliationParameter{Enabled: true},
			},
		},
		ScheduledAnalysis: {
			AllowAPIAccess: true,
			GetType:        func() any { return ScheduledAnalysisParameter{} },
			HydrationRules: ParamTypeHydrationRules[ScheduledAnalysisParameter]{
				Default: ScheduledAnalysisParameter{Enabled: false, RRule: ""},
			},
		},
		ClientMetricsKey: {
			AllowAPIAccess: true,
			GetType:        func() any { return ClientMetricsParameter{} },
			HydrationRules: ParamTypeHydrationRules[ClientMetricsParameter]{
				Default: ClientMetricsParameter{Enabled: false},
			},
		},
		APITokenExpiration: {
			AllowAPIAccess: true,
			GetType:        func() any { return APITokenExpirationParameter{} },
			HydrationRules: ParamTypeHydrationRules[APITokenExpirationParameter]{
				Default: APITokenExpirationParameter{Enabled: false, ExpirationPeriod: defaultAPITokenExpirationPeriod},
				Normalize: func(parameter *APITokenExpirationParameter) {
					if parameter.ExpirationPeriod <= 0 || parameter.ExpirationPeriod > maxAPITokenExpirationPeriod {
						parameter.ExpirationPeriod = defaultAPITokenExpirationPeriod
					}
				},
			},
		},
		TrustedProxiesConfig: {
			GetType: func() any { return TrustedProxiesParameters{} },
			HydrationRules: ParamTypeHydrationRules[TrustedProxiesParameters]{
				Default: TrustedProxiesParameters{TrustedProxies: 0},
			},
		},
		FedEULACustomTextKey: {
			GetType: func() any { return FedEULACustomTextParameter{} },
			HydrationRules: ParamTypeHydrationRules[FedEULACustomTextParameter]{
				Default: FedEULACustomTextParameter{CustomText: ""},
			},
		},
		TierManagementParameterKey: {
			GetType: func() any { return TieringParameter{} },
			HydrationRules: ParamTypeHydrationRules[TieringParameter]{
				Default: TieringParameter{},
			},
		},
		SessionTTLHours: {
			GetType: func() any { return SessionTTLHoursParameter{} },
			HydrationRules: ParamTypeHydrationRules[SessionTTLHoursParameter]{
				Default: SessionTTLHoursParameter{Hours: defaultSessionTTLHours},
				Normalize: func(parameter *SessionTTLHoursParameter) {
					if parameter.Hours <= 0 {
						parameter.Hours = defaultSessionTTLHours
					}
				},
			},
		},
		StaleClientUpdatedLogicKey: {
			GetType: func() any { return StaleClientUpdatedLogicParameter{} },
			HydrationRules: ParamTypeHydrationRules[StaleClientUpdatedLogicParameter]{
				Default: StaleClientUpdatedLogicParameter{Enabled: false},
			},
		},
		RetainIngestedFilesKey: {
			GetType: func() any { return RetainIngestedFilesParameter{} },
			HydrationRules: ParamTypeHydrationRules[RetainIngestedFilesParameter]{
				Default: RetainIngestedFilesParameter{Enabled: false},
			},
		},
		AGTParameterKey: {
			GetType: func() any { return AGTParameter{} },
			HydrationRules: ParamTypeHydrationRules[AGTParameter]{
				Default: AGTParameter{
					DAWGsWorkerLimit:     defaultAGTDAWGsWorkerLimit,
					ExpansionWorkerLimit: defaultAGTExpansionWorkerLimit,
					SelectorWorkerLimit:  defaultAGTSelectorWorkerLimit,
				},
				Normalize: func(parameter *AGTParameter) {
					if parameter.DAWGsWorkerLimit <= 0 || parameter.DAWGsWorkerLimit > maxAGTDAWGsWorkerLimit {
						parameter.DAWGsWorkerLimit = maxAGTDAWGsWorkerLimit
					}
					if parameter.SelectorWorkerLimit <= 0 {
						parameter.SelectorWorkerLimit = defaultAGTSelectorWorkerLimit
					}
					if parameter.ExpansionWorkerLimit <= 0 {
						parameter.ExpansionWorkerLimit = defaultAGTExpansionWorkerLimit
					}
				},
			},
		},
		TimeoutLimit: {
			GetType: func() any { return TimeoutLimitParameter{} },
			HydrationRules: ParamTypeHydrationRules[TimeoutLimitParameter]{
				Default: TimeoutLimitParameter{Enabled: true},
			},
		},
		APITokens: {
			GetType: func() any { return APITokensParameter{} },
			HydrationRules: ParamTypeHydrationRules[APITokensParameter]{
				Default: APITokensParameter{Enabled: true},
			},
		},
		EnvironmentTargetedAccessControlKey: {
			GetType: func() any { return EnvironmentTargetedAccessControlParameter{} },
			HydrationRules: ParamTypeHydrationRules[EnvironmentTargetedAccessControlParameter]{
				Default: EnvironmentTargetedAccessControlParameter{Enabled: false},
			},
		},
		SupportAccountProvisioningKey: {
			GetType: func() any { return SupportAccountProvisioningParameter{} },
			HydrationRules: ParamTypeHydrationRules[SupportAccountProvisioningParameter]{
				Default: SupportAccountProvisioningParameter{
					Enabled:    true,
					SessionTTL: duration.ISODuration(2 * time.Hour),
				},
			},
		},
		GraphStorageOptimizationKey: {
			GetType: func() any { return GraphStorageOptimizationParameter{} },
			HydrationRules: ParamTypeHydrationRules[GraphStorageOptimizationParameter]{
				Default: GraphStorageOptimizationParameter{
					AfterBoot:          false,
					AfterAnalysis:      false,
					MinIntervalSeconds: defaultGraphStorageOptimizationMinIntervalSecs,
				},
				Normalize: func(parameter *GraphStorageOptimizationParameter) {
					if parameter.MinIntervalSeconds < 0 {
						parameter.MinIntervalSeconds = defaultGraphStorageOptimizationMinIntervalSecs
					}
				},
			},
		},
		CustomSubClaimKey: {
			GetType: func() any { return CustomSubClaimKeyParameter{} },
			HydrationRules: ParamTypeHydrationRules[CustomSubClaimKeyParameter]{
				Default: CustomSubClaimKeyParameter{ClaimKey: ""},
			},
		},
	}
)

// getParamTypeHydrationRules returns the specified hydration rule and a bool indicating success
func getParamTypeHydrationRules[ParamType any](key ParameterKey) (ParamTypeHydrationRules[ParamType], bool) {
	def, ok := ParamTypeDefinitions[key].HydrationRules.(ParamTypeHydrationRules[ParamType])
	return def, ok
}
