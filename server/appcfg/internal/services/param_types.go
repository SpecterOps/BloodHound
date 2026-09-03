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

// This file defines the external interface methods allowing
// access to specific parameter types' values
package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	iso8601 "github.com/channelmeter/iso8601duration"
	"github.com/specterops/bloodhound/packages/go/bhlog/attr"
	"github.com/specterops/dawgs/drivers/neo4j"
)

const (
	PasswordExpirationWindow ParameterKey = "auth.password_expiration_window"
	SessionTTLHours          ParameterKey = "auth.session_ttl_hours"
	Neo4jConfigs             ParameterKey = "neo4j.configuration"
	CitrixRDPSupportKey      ParameterKey = "analysis.citrix_rdp_support"
	PruneTTL                 ParameterKey = "prune.ttl"
	ReconciliationKey        ParameterKey = "analysis.reconciliation"
	ScheduledAnalysis        ParameterKey = "analysis.scheduled"
	ClientMetricsKey         ParameterKey = "pipeline.client_metrics"
	APITokenExpiration       ParameterKey = "auth.api_token_expiration"

	// The below keys are not intended to be user updatable, so should not be added to IsValidKey
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

// PasswordExpirationWindow

type PasswordExpiration struct {
	Duration time.Duration `json:"duration"`
}

// Because PasswordExpiration are stored as ISO strings, but we want to use them as durations, we override UnmarshalJSON to handle the conversion
func (s *PasswordExpiration) UnmarshalJSON(data []byte) error {
	pDb := struct {
		Duration string `json:"duration,omitempty"`
	}{}

	if err := json.Unmarshal(data, &pDb); err != nil {
		return fmt.Errorf("error unmarshaling data for PasswordExpiration: %w", err)
	} else {
		if duration, err := iso8601.FromString(pDb.Duration); err != nil {
			return err
		} else {
			s.Duration = duration.ToDuration()
		}

		return nil
	}

}

func (s *Service) GetPasswordExpiration(ctx context.Context) time.Duration {
	var expiration PasswordExpiration

	if cfg, err := s.db.GetConfigurationParameter(ctx, PasswordExpirationWindow); err != nil {
		slog.WarnContext(ctx, "Failed to fetch password expiration configuration; returning default values")
		return DefaultPasswordExpirationWindow
	} else if err := cfg.Map(&expiration); err != nil {
		slog.WarnContext(ctx, "Invalid password expiration configuration supplied; returning default values.",
			attr.Error(err),
			slog.String("parameter_key", string(PasswordExpirationWindow)))
		return DefaultPasswordExpirationWindow
	}

	return expiration.Duration
}

// Neo4jConfigs

type Neo4jParameters struct {
	WriteFlushSize int `json:"write_flush_size,omitempty"`
	BatchWriteSize int `json:"batch_write_size,omitempty"`
}

func (s *Service) GetNeo4jParameters(ctx context.Context) Neo4jParameters {
	var result = Neo4jParameters{
		WriteFlushSize: neo4j.DefaultWriteFlushSize,
		BatchWriteSize: neo4j.DefaultBatchWriteSize,
	}

	if neo4jParametersCfg, err := s.db.GetConfigurationParameter(ctx, Neo4jConfigs); err != nil {
		slog.WarnContext(ctx, "Failed to fetch neo4j configuration; returning default values")
	} else if err = neo4jParametersCfg.Map(&result); err != nil {
		slog.WarnContext(ctx, "Invalid neo4j configuration supplied; returning default values.",
			attr.Error(err),
			slog.String("parameter_key", string(Neo4jConfigs)))
	}

	return result
}

// CitrixRDP

type CitrixRDPSupport struct {
	Enabled bool `json:"enabled,omitempty"`
}

func (s *Service) GetCitrixRDPSupport(ctx context.Context) bool {
	var result CitrixRDPSupport

	if cfg, err := s.db.GetConfigurationParameter(ctx, CitrixRDPSupportKey); err != nil {
		slog.WarnContext(ctx, "Failed to fetch CitrixRDPSupport configuration; returning default values")
	} else if err := cfg.Map(&result); err != nil {
		slog.WarnContext(ctx, "Invalid CitrixRDPSupport configuration supplied, returning default values.",
			attr.Error(err),
			slog.String("parameter_key", string(CitrixRDPSupportKey)))
	}

	return result.Enabled
}

// PruneTTL

type PruneTTLParameters struct {
	BaseTTL           time.Duration `json:"base_ttl,omitempty" validate:"duration,min=P4D,max=P30D"`
	HasSessionEdgeTTL time.Duration `json:"has_session_edge_ttl,omitempty" validate:"duration,min=P2D,max=P7D"`
}

// Because PruneTTLs are stored as ISO strings, but we want to use them as durations, we override UnmarshalJSON to handle the conversion
func (s *PruneTTLParameters) UnmarshalJSON(data []byte) error {
	pTTL := struct {
		BaseTTL           string `json:"base_ttl,omitempty"`
		HasSessionEdgeTTL string `json:"has_session_edge_ttl,omitempty"`
	}{}

	if err := json.Unmarshal(data, &pTTL); err != nil {
		return fmt.Errorf("error unmarshaling data for PruneTTLParameters: %w", err)
	} else {
		if duration, err := iso8601.FromString(pTTL.BaseTTL); err != nil {
			return errors.New("missing or invalid base_ttl")
		} else {
			s.BaseTTL = duration.ToDuration()
		}
		if duration, err := iso8601.FromString(pTTL.HasSessionEdgeTTL); err != nil {
			return errors.New("missing or invalid has_session_edge_ttl")
		} else {

			s.HasSessionEdgeTTL = duration.ToDuration()
		}

		return nil
	}
}

func (s *Service) GetPruneTTLParameters(ctx context.Context) PruneTTLParameters {
	result := PruneTTLParameters{
		BaseTTL:           DefaultPruneBaseTTL,
		HasSessionEdgeTTL: DefaultPruneHasSessionEdgeTTL,
	}

	if pruneTTLParametersCfg, err := s.db.GetConfigurationParameter(ctx, PruneTTL); err != nil {
		slog.WarnContext(ctx, "Failed to fetch prune TTL configuration; returning default values")
	} else if err = pruneTTLParametersCfg.Map(&result); err != nil {
		slog.WarnContext(ctx, "Invalid prune TTL configuration supplied; returning default values.",
			attr.Error(err),
			slog.String("parameter_key", string(PruneTTL)))
	}

	return result
}

// Reconciliation

type ReconciliationParameter struct {
	Enabled bool `json:"enabled,omitempty"`
}

func (s *Service) GetReconciliationParameter(ctx context.Context) bool {
	result := ReconciliationParameter{Enabled: true}

	if cfg, err := s.db.GetConfigurationParameter(ctx, ReconciliationKey); err != nil {
		slog.WarnContext(ctx, "Failed to fetch reconciliation configuration; returning default values")
	} else if err := cfg.Map(&result); err != nil {
		slog.WarnContext(ctx, "Invalid reconciliation configuration supplied, returning default values.",
			attr.Error(err),
			slog.String("parameter_key", string(ReconciliationKey)))
	}

	return result.Enabled
}

type ScheduledAnalysisParameter struct {
	Enabled bool   `json:"enabled,omitempty"`
	RRule   string `json:"rrule,omitempty" validate:"rrule"`
}

func (s *Service) GetScheduledAnalysisParameter(ctx context.Context) (ScheduledAnalysisParameter, error) {
	result := ScheduledAnalysisParameter{Enabled: false, RRule: ""}

	if cfg, err := s.db.GetConfigurationParameter(ctx, ScheduledAnalysis); err != nil {
		return result, err
	} else if err := cfg.Map(&result); err != nil {
		return result, err
	}

	return result, nil
}

type TrustedProxiesParameters struct {
	TrustedProxies int `json:"trusted_proxies,omitempty"`
}

func (s *Service) GetTrustedProxiesParameters(ctx context.Context) int {
	var result = TrustedProxiesParameters{
		TrustedProxies: 0,
	}

	if trustedProxiesParametersCfg, err := s.db.GetConfigurationParameter(ctx, TrustedProxiesConfig); err != nil {
		slog.WarnContext(ctx, "Failed to fetch trusted proxies configuration; returning default values")
	} else if err = trustedProxiesParametersCfg.Map(&result); err != nil {
		slog.WarnContext(ctx, "Invalid trusted proxies configuration supplied; returning default values.",
			attr.Error(err),
			slog.String("parameter_key", string(TrustedProxiesConfig)))
	}

	return result.TrustedProxies
}

type TieringParameters struct {
	TierLimit                int  `json:"tier_limit,omitempty"`
	LabelLimit               int  `json:"label_limit,omitempty"`
	MultiTierAnalysisEnabled bool `json:"multi_tier_analysis_enabled,omitempty"`
}

type AGTParameters struct {
	DAWGsWorkerLimit     int `json:"dawgs_worker_limit,omitempty"`
	ExpansionWorkerLimit int `json:"expansion_worker_limit,omitempty"`
	SelectorWorkerLimit  int `json:"selector_worker_limit,omitempty"`
}

func (s *Service) GetAGTParameters(ctx context.Context) AGTParameters {
	result := AGTParameters{
		DAWGsWorkerLimit:     DefaultDawgsWorkerLimit,
		ExpansionWorkerLimit: DefaultExpansionWorkerLimit,
		SelectorWorkerLimit:  DefaultSelectorWorkerLimit,
	}

	if agtParametersCfg, err := s.db.GetConfigurationParameter(ctx, AGTParameterKey); err != nil {
		slog.WarnContext(ctx, "Failed to fetch agt configuration; returning default values")
	} else if err = agtParametersCfg.Map(&result); err != nil {
		slog.WarnContext(ctx, "Invalid agt configuration supplied; returning default values.",
			attr.Error(err),
			slog.String("parameter_key", string(AGTParameterKey)))
	}

	if result.DAWGsWorkerLimit <= 0 || result.DAWGsWorkerLimit > MaxDawgsWorkerLimit {
		slog.WarnContext(ctx, "Invalid agt configuration supplied for dawgs_worker_limit; setting to max value.", slog.Int("max_dawgs_worker_limit", MaxDawgsWorkerLimit))
		result.DAWGsWorkerLimit = MaxDawgsWorkerLimit
	}

	if result.SelectorWorkerLimit <= 0 {
		slog.WarnContext(ctx, "Invalid agt configuration supplied for selector_worker_limit; setting to default value.", slog.Int("default_selector_worker_limit", DefaultSelectorWorkerLimit))
		result.SelectorWorkerLimit = DefaultSelectorWorkerLimit
	}

	if result.ExpansionWorkerLimit <= 0 {
		slog.WarnContext(ctx, "Invalid agt configuration supplied for expansion_worker_limit; setting to default value.", slog.Int("default_expansion_worker_limit", DefaultExpansionWorkerLimit))
		result.ExpansionWorkerLimit = DefaultExpansionWorkerLimit
	}

	return result
}

type FedEULACustomTextParameter struct {
	CustomText string `json:"custom_text,omitempty"`
}

// GetFedRAMPCustomEULA Note this is not gated by the FedEULA FF and that should be checked alongside this
func (s *Service) GetFedRAMPCustomEULA(ctx context.Context) string {
	var result FedEULACustomTextParameter

	if fedEulaCustomText, err := s.db.GetConfigurationParameter(ctx, FedEULACustomTextKey); err != nil {
		slog.WarnContext(ctx, "Failed to fetch eula custom text; returning default value")
	} else if err = fedEulaCustomText.Map(&result); err != nil {
		slog.WarnContext(ctx, "Invalid eula custom text supplied; returning default value")
	}

	return result.CustomText
}

type SessionTTLHoursParameter struct {
	Hours int `json:"hours,omitempty"`
}

func (s *Service) GetSessionTTLHours(ctx context.Context) time.Duration {
	var result = SessionTTLHoursParameter{
		Hours: DefaultSessionTTLHours, // Default to a logged in auth session time to live of 8 hours
	}

	if sessionTTLHours, err := s.db.GetConfigurationParameter(ctx, SessionTTLHours); err != nil {
		slog.WarnContext(ctx, "Failed to fetch auth session ttl hours; returning default values")
	} else if err = sessionTTLHours.Map(&result); err != nil {
		slog.WarnContext(ctx, "Invalid auth session ttl hours supplied; returning default values")
	} else if result.Hours <= 0 {
		slog.WarnContext(ctx, "Auth session ttl hours ≤ 0; returning default values")
		result.Hours = DefaultSessionTTLHours
	}

	return time.Hour * time.Duration(result.Hours)
}

// StaleClientUpdatedLogic

type StaleClientUpdatedLogic struct {
	Enabled bool `json:"enabled,omitempty"`
}

func (s *Service) GetStaleClientUpdatedLogic(ctx context.Context) bool {
	var result StaleClientUpdatedLogic

	if cfg, err := s.db.GetConfigurationParameter(ctx, StaleClientUpdatedLogicKey); err != nil {
		slog.WarnContext(ctx, "Failed to fetch StaleClientLogic configuration; returning default values")
	} else if err := cfg.Map(&result); err != nil {
		slog.WarnContext(ctx, "Invalid StaleClientLogic configuration supplied. returning default values.",
			attr.Error(err),
			slog.String("parameter_key", string(StaleClientUpdatedLogicKey)))
	}

	return result.Enabled
}

// RetainIngestedFiles
type RetainIngestedFilesParameter struct {
	Enabled bool `json:"enabled,omitempty"`
}

func (s *Service) ShouldRetainIngestedFiles(ctx context.Context) bool {
	result := RetainIngestedFilesParameter{
		// Retention should always default to false in the case where the parameter may not be set
		Enabled: false,
	}

	if cfg, err := s.db.GetConfigurationParameter(ctx, RetainIngestedFilesKey); err != nil {
		slog.WarnContext(ctx, "Failed to fetch ShouldRetainIngestedFiles configuration; returning default values")
	} else if err := cfg.Map(&result); err != nil {
		slog.WarnContext(ctx, "Invalid ShouldRetainIngestedFiles configuration supplied, returning default values.",
			attr.Error(err),
			slog.String("parameter_key", string(RetainIngestedFilesKey)))
	}

	return result.Enabled
}

type TimeoutLimitParameter struct {
	Enabled bool `json:"enabled,omitempty"`
}

func (s *Service) GetTimeoutLimitParameter(ctx context.Context) bool {
	result := TimeoutLimitParameter{Enabled: true}

	if cfg, err := s.db.GetConfigurationParameter(ctx, TimeoutLimit); err != nil {
		slog.WarnContext(ctx, "Failed to fetch timeout limit configuration; returning default values")
	} else if err := cfg.Map(&result); err != nil {
		slog.WarnContext(ctx, "Invalid timeout limit configuration supplied, returning default values.",
			attr.Error(err),
			slog.String("parameter_key", string(TimeoutLimit)))
	}

	return result.Enabled
}

type APITokensParameter struct {
	Enabled bool `json:"enabled"`
}

func (s *Service) GetAPITokensParameter(ctx context.Context) bool {
	result := APITokensParameter{Enabled: true}

	if cfg, err := s.db.GetConfigurationParameter(ctx, APITokens); err != nil {
		slog.WarnContext(ctx, "Failed to fetch API tokens configuration; returning default values")
	} else if err := cfg.Map(&result); err != nil {
		slog.WarnContext(ctx, "Invalid API tokens configuration supplied, returning default values.",
			attr.Error(err),
			slog.String("parameter_key", string(APITokens)))
	}

	return result.Enabled
}

type EnvironmentTargetedAccessControlParameters struct {
	Enabled bool `json:"enabled,omitempty"`
}

func (s *Service) GetEnvironmentTargetedAccessControlParameters(ctx context.Context) EnvironmentTargetedAccessControlParameters {
	result := EnvironmentTargetedAccessControlParameters{
		Enabled: false,
	}

	if etacParametersCfg, err := s.db.GetConfigurationParameter(ctx, EnvironmentTargetedAccessControlKey); err != nil {
		slog.WarnContext(ctx, "Failed to fetch environment targeted access control configuration; returning default values")
	} else if err = etacParametersCfg.Map(&result); err != nil {
		slog.WarnContext(ctx, "Invalid environment targeted access control configuration supplied; returning default values",
			slog.String("parameter_key", string(EnvironmentTargetedAccessControlKey)),
			attr.Error(err))
	}

	return result
}

type SupportAccountProvisioningParameters struct {
	// Setting disabled as false means that you are explicitly opting into the feature
	Enabled    bool          `json:"enabled,omitempty"`
	SessionTTL time.Duration `json:"session_ttl,omitempty"`
}

func (s *SupportAccountProvisioningParameters) UnmarshalJSON(data []byte) error {
	pDb := struct {
		SessionTTL string `json:"session_ttl,omitempty"`
		Enabled    bool   `json:"enabled,omitempty"`
	}{}

	if err := json.Unmarshal(data, &pDb); err != nil {
		return fmt.Errorf("error unmarshaling data for SupportAccountProvisioningParameters: %w", err)
	} else {
		if duration, err := iso8601.FromString(pDb.SessionTTL); err != nil {
			return err
		} else {
			s.SessionTTL = duration.ToDuration()
			s.Enabled = pDb.Enabled
		}

		return nil
	}
}

func (s *Service) GetSupportAccountProvisioningParameters(ctx context.Context) SupportAccountProvisioningParameters {
	result := SupportAccountProvisioningParameters{
		Enabled:    true,
		SessionTTL: time.Hour * 2,
	}

	if jitParametersCfg, err := s.db.GetConfigurationParameter(ctx, SupportAccountProvisioningKey); err != nil {
		slog.WarnContext(ctx, "Failed to fetch support account provisioning configuration; returning default values")
	} else if err = jitParametersCfg.Map(&result); err != nil {
		slog.WarnContext(ctx, "Invalid support account provisioning configuration supplied; returning default values",
			attr.Error(err))
	}

	return result
}

type ClientMetricsParameter struct {
	Enabled bool `json:"enabled,omitempty"`
}

func (s *Service) GetClientMetricsParameter(ctx context.Context) ClientMetricsParameter {
	result := ClientMetricsParameter{
		Enabled: false,
	}

	if clientMetricsCfg, err := s.db.GetConfigurationParameter(ctx, ClientMetricsKey); err != nil {
		slog.WarnContext(ctx, "Failed to fetch client metrics configuration; returning default values")
	} else if err = clientMetricsCfg.Map(&result); err != nil {
		slog.WarnContext(ctx, "Invalid client metrics configuration supplied; returning default values",
			attr.Error(err))
	}

	return result
}

type APITokenExpirationParameter struct {
	Enabled          bool `json:"enabled"`
	ExpirationPeriod int  `json:"expiration_period" validate:"integer,min=1,max=365"`
}

func (s *Service) GetAPITokenExpirationParameter(ctx context.Context) APITokenExpirationParameter {
	result := APITokenExpirationParameter{Enabled: false, ExpirationPeriod: 90}

	if cfg, err := s.db.GetConfigurationParameter(ctx, APITokenExpiration); err != nil {
		slog.WarnContext(ctx, "Failed to fetch API tokens expiration configuration; returning default values.")
	} else if err := cfg.Map(&result); err != nil {
		slog.WarnContext(ctx, "Invalid API tokens expiration configuration supplied, returning default values.",
			attr.Error(err))
	} else if result.ExpirationPeriod <= 0 || result.ExpirationPeriod > 365 {
		slog.WarnContext(ctx, "Invalid API token expiration period supplied, returning default values.",
			slog.Int("invalid_expiration_period", result.ExpirationPeriod),
			slog.String("parameter_key", string(APITokenExpiration)))
		result.ExpirationPeriod = 90
	}

	return result
}

// GraphStorageOptimization
type GraphStorageOptimizationParameter struct {
	AfterBoot          bool `json:"after_boot"`
	AfterAnalysis      bool `json:"after_analysis"`
	MinIntervalSeconds int  `json:"min_interval_seconds"`
}

func (s *Service) GetGraphStorageOptimizationParameter(ctx context.Context) GraphStorageOptimizationParameter {
	result := GraphStorageOptimizationParameter{
		AfterBoot:          false,
		AfterAnalysis:      false,
		MinIntervalSeconds: 86400,
	}

	if cfg, err := s.db.GetConfigurationParameter(ctx, GraphStorageOptimizationKey); err != nil {
		slog.WarnContext(ctx, "Failed to fetch graph storage optimization configuration; returning default values")
	} else if err := cfg.Map(&result); err != nil {
		slog.WarnContext(ctx, "Invalid graph storage optimization configuration supplied; returning default values",
			attr.Error(err),
			slog.String("parameter_key", string(GraphStorageOptimizationKey)))
	} else if result.MinIntervalSeconds < 0 {
		slog.WarnContext(ctx, "Invalid negative min interval seconds supplied, returning default value.",
			slog.Int("invalid_min_interval_seconds", result.MinIntervalSeconds),
			slog.String("parameter_key", string(GraphStorageOptimizationKey)))
		result.MinIntervalSeconds = 86400
	}

	return result
}
