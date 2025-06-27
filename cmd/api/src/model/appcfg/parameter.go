// Copyright 2023 Specter Ops, Inc.
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

package appcfg

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"reflect"
	"time"

	iso8601 "github.com/channelmeter/iso8601duration"
	"github.com/specterops/bloodhound/src/database/types"
	"github.com/specterops/bloodhound/src/model"
	"github.com/specterops/bloodhound/src/utils"
	"github.com/specterops/bloodhound/src/utils/validation"
	"github.com/specterops/dawgs/drivers/neo4j"
)

type ParameterKey string

const (
	PasswordExpirationWindow ParameterKey = "auth.password_expiration_window"
	SessionTTLHours          ParameterKey = "auth.session_ttl_hours"
	Neo4jConfigs             ParameterKey = "neo4j.configuration"
	CitrixRDPSupportKey      ParameterKey = "analysis.citrix_rdp_support"
	PruneTTL                 ParameterKey = "prune.ttl"
	ReconciliationKey        ParameterKey = "analysis.reconciliation"

	// The below keys are not intended to be user updateable, so should not be added to IsValidKey
	ScheduledAnalysis          ParameterKey = "analysis.scheduled"
	TrustedProxiesConfig       ParameterKey = "http.trusted_proxies"
	FedEULACustomTextKey       ParameterKey = "eula.custom_text"
	TierManagementParameterKey ParameterKey = "analysis.tiering"
)

const (
	DefaultPasswordExpirationWindow = time.Hour * 24 * 90

	RestrictOutboundNTLMDefaultValue = "analysis.restrict_outbound_ntlm_default_value"
	DefaultSessionTTLHours           = 8

	DefaultPruneBaseTTL           = time.Hour * 24 * 7
	DefaultPruneHasSessionEdgeTTL = time.Hour * 24 * 3

	DefaultTierLimit  = 1
	DefaultLabelLimit = 0
)

// Parameter is a runtime configuration parameter that can be fetched from the appcfg.ParameterService interface. The
// Value member is a DB-safe JSON type wrapper that can store arbitrary JSON objects and map them to golang struct
// definitions.
type Parameter struct {
	Key         ParameterKey      `json:"key" gorm:"unique"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Value       types.JSONBObject `json:"value"`

	model.Serial
}

// Map is a convenience function for mapping the data stored in the Value Parameter struct member onto
// a richer type provided by the given value.
func (s *Parameter) Map(value any) error {
	return s.Value.Map(value)
}

func (s *Parameter) IsValidKey(parameterKey ParameterKey) bool {
	switch parameterKey {
	case PasswordExpirationWindow, Neo4jConfigs, PruneTTL, CitrixRDPSupportKey, ReconciliationKey, RestrictOutboundNTLMDefaultValue:
		return true
	default:
		return false
	}
}

// IsProtectedKey These keys should not be updatable by users
func (s *Parameter) IsProtectedKey(parameterKey ParameterKey) bool {
	switch parameterKey {
	case ScheduledAnalysis, TrustedProxiesConfig, FedEULACustomTextKey, TierManagementParameterKey, SessionTTLHours:
		return true
	default:
		return false
	}
}

// Validate WARNING - This will not protect the protected keys, use IsValidKey for that, this validates the json payload matches the intended parameter values
func (s *Parameter) Validate() utils.Errors {
	// validate the base parameter
	var (
		objMap map[string]any
		ok     bool
	)
	if objMap, ok = s.Value.Object.(map[string]any); !ok || len(objMap) == 0 {
		return utils.Errors{errors.New("missing or invalid property: value")}
	}

	// validate the specific parameter value
	var v any
	switch s.Key {
	case PasswordExpirationWindow:
		v = &PasswordExpiration{}
	case Neo4jConfigs:
		v = &Neo4jParameters{}
	case PruneTTL:
		v = &PruneTTLParameters{}
	case CitrixRDPSupportKey:
		v = &CitrixRDPSupport{}
	case ReconciliationKey:
		v = &ReconciliationParameter{}
	case RestrictOutboundNTLMDefaultValue:
		v = &RestrictOutboundNTLMDefault{}
	case TierManagementParameterKey:
		v = &TieringParameters{}
	case ScheduledAnalysis:
		v = &ScheduledAnalysisParameter{}
	case TrustedProxiesConfig:
		v = &TrustedProxiesParameters{}
	case FedEULACustomTextKey:
		v = &FedEULACustomTextParameter{}
	case SessionTTLHours:
		v = &SessionTTLHoursParameter{}
	default:
		return utils.Errors{errors.New("invalid key")}
	}

	// numField panics when val is not a struct, so we need both checks
	if val := reflect.Indirect(reflect.ValueOf(v)); val.Kind() != reflect.Struct || val.NumField() != len(objMap) {
		return utils.Errors{errors.New("value property contains an invalid field")}
	} else if err := s.Map(&v); err != nil {
		return utils.Errors{err}
	} else if errs := validation.Validate(v); errs != nil {
		return errs
	}

	return nil
}

func (s *Parameter) AuditData() model.AuditData {
	return model.AuditData{
		"key":   s.Key,
		"value": s.Value,
	}
}

type AppConfigUpdateRequest struct {
	Key   string         `json:"key"`
	Value map[string]any `json:"value"`
}

func ConvertAppConfigUpdateRequestToParameter(appConfigUpdateRequest AppConfigUpdateRequest) (Parameter, error) {
	if value, err := types.NewJSONBObject(appConfigUpdateRequest.Value); err != nil {
		return Parameter{}, fmt.Errorf("failed to convert value to JSONBObject: %w", err)
	} else {
		return Parameter{
			Key:   ParameterKey(appConfigUpdateRequest.Key),
			Value: value,
		}, nil
	}
}

// Parameters is a collection of Parameter structs.
type Parameters []Parameter

// ParameterService is a contract which defines expected functionality for fetching and setting Parameter from an
// abstract backend storage.
type ParameterService interface {
	// GetAllConfigurationParameters gets all available runtime Parameters for the application.
	GetAllConfigurationParameters(ctx context.Context) (Parameters, error)

	// GetConfigurationParameter attempts to fetch a Parameter struct by its parameter name.
	GetConfigurationParameter(ctx context.Context, parameterKey ParameterKey) (Parameter, error)

	// SetConfigurationParameter attempts to store or update the given Parameter.
	SetConfigurationParameter(ctx context.Context, configurationParameter Parameter) error
}

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

func GetPasswordExpiration(ctx context.Context, service ParameterService) time.Duration {
	var expiration PasswordExpiration

	if cfg, err := service.GetConfigurationParameter(ctx, PasswordExpirationWindow); err != nil {
		slog.WarnContext(ctx, "Failed to fetch password expiratio configuration; returning default values")
		return DefaultPasswordExpirationWindow
	} else if err := cfg.Map(&expiration); err != nil {
		slog.WarnContext(ctx, "Invalid password expiration configuration supplied; returning default values")
		return DefaultPasswordExpirationWindow
	}

	return expiration.Duration
}

// Neo4jConfigs

type Neo4jParameters struct {
	WriteFlushSize int `json:"write_flush_size,omitempty"`
	BatchWriteSize int `json:"batch_write_size,omitempty"`
}

func GetNeo4jParameters(ctx context.Context, service ParameterService) Neo4jParameters {
	var result = Neo4jParameters{
		WriteFlushSize: neo4j.DefaultWriteFlushSize,
		BatchWriteSize: neo4j.DefaultBatchWriteSize,
	}

	if neo4jParametersCfg, err := service.GetConfigurationParameter(ctx, Neo4jConfigs); err != nil {
		slog.WarnContext(ctx, "Failed to fetch neo4j configuration; returning default values")
	} else if err = neo4jParametersCfg.Map(&result); err != nil {
		slog.WarnContext(ctx, "Invalid neo4j configuration supplied; returning default values")
	}

	return result
}

// CitrixRDP

type CitrixRDPSupport struct {
	Enabled bool `json:"enabled,omitempty"`
}

func GetCitrixRDPSupport(ctx context.Context, service ParameterService) bool {
	var result CitrixRDPSupport

	if cfg, err := service.GetConfigurationParameter(ctx, CitrixRDPSupportKey); err != nil {
		slog.WarnContext(ctx, "Failed to fetch CitrixRDPSupport configuration; returning default values")
	} else if err := cfg.Map(&result); err != nil {
		slog.WarnContext(ctx, fmt.Sprintf("Invalid CitrixRDPSupport configuration supplied, %v. returning default values.", err))
	}

	return result.Enabled
}

// RestrictOutboundNTLMDefault

type RestrictOutboundNTLMDefault struct {
	Enabled bool `json:"enabled,omitempty"`
}

func GetRestrictOutboundNTLMDefaultValue(ctx context.Context, service ParameterService) bool {
	var result RestrictOutboundNTLMDefault

	if cfg, err := service.GetConfigurationParameter(ctx, RestrictOutboundNTLMDefaultValue); err != nil {
		slog.WarnContext(ctx, "Failed to fetch RestrictOutboundNTLMDefaultValue configuration; returning default values")
	} else if err := cfg.Map(&result); err != nil {
		slog.WarnContext(ctx, fmt.Sprintf("Invalid RestrictOutboundNTLMDefaultValue configuration supplied, %v. returning default values.", err))
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

func GetPruneTTLParameters(ctx context.Context, service ParameterService) PruneTTLParameters {
	result := PruneTTLParameters{
		BaseTTL:           DefaultPruneBaseTTL,
		HasSessionEdgeTTL: DefaultPruneHasSessionEdgeTTL,
	}

	if pruneTTLParametersCfg, err := service.GetConfigurationParameter(ctx, PruneTTL); err != nil {
		slog.WarnContext(ctx, "Failed to fetch prune TTL configuration; returning default values")
	} else if err = pruneTTLParametersCfg.Map(&result); err != nil {
		slog.WarnContext(ctx, fmt.Sprintf("Invalid prune TTL configuration supplied; returning default values %+v", err))
	}

	return result
}

// Reconciliation

type ReconciliationParameter struct {
	Enabled bool `json:"enabled,omitempty"`
}

func GetReconciliationParameter(ctx context.Context, service ParameterService) bool {
	result := ReconciliationParameter{Enabled: true}

	if cfg, err := service.GetConfigurationParameter(ctx, ReconciliationKey); err != nil {
		slog.WarnContext(ctx, "Failed to fetch reconciliation configuration; returning default values")
	} else if err := cfg.Map(&result); err != nil {
		slog.WarnContext(ctx, fmt.Sprintf("Invalid reconciliation configuration supplied, %v. returning default values.", err))
	}

	return result.Enabled
}

type ScheduledAnalysisParameter struct {
	Enabled bool   `json:"enabled,omitempty"`
	RRule   string `json:"rrule,omitempty" validate:"rrule"`
}

func GetScheduledAnalysisParameter(ctx context.Context, service ParameterService) (ScheduledAnalysisParameter, error) {
	result := ScheduledAnalysisParameter{Enabled: false, RRule: ""}

	if cfg, err := service.GetConfigurationParameter(ctx, ScheduledAnalysis); err != nil {
		return result, err
	} else if err := cfg.Map(&result); err != nil {
		return result, err
	}

	return result, nil
}

type TrustedProxiesParameters struct {
	TrustedProxies int `json:"trusted_proxies,omitempty"`
}

func GetTrustedProxiesParameters(ctx context.Context, service ParameterService) int {
	var result = TrustedProxiesParameters{
		TrustedProxies: 0,
	}

	if trustedProxiesParametersCfg, err := service.GetConfigurationParameter(ctx, TrustedProxiesConfig); err != nil {
		slog.WarnContext(ctx, "Failed to fetch trusted proxies configuration; returning default values")
	} else if err = trustedProxiesParametersCfg.Map(&result); err != nil {
		slog.WarnContext(ctx, "Invalid trusted proxies configuration supplied; returning default values")
	}

	return result.TrustedProxies
}

type TieringParameters struct {
	TierLimit                int  `json:"tier_limit,omitempty"`
	LabelLimit               int  `json:"label_limit,omitempty"`
	MultiTierAnalysisEnabled bool `json:"multi_tier_analysis_enabled,omitempty"`
}

func GetTieringParameters(ctx context.Context, service ParameterService) TieringParameters {
	result := TieringParameters{
		TierLimit:                DefaultTierLimit,
		LabelLimit:               DefaultLabelLimit,
		MultiTierAnalysisEnabled: false,
	}

	if tieringParametersCfg, err := service.GetConfigurationParameter(ctx, TierManagementParameterKey); err != nil {
		slog.WarnContext(ctx, "Failed to fetch tiering configuration; returning default values")
	} else if err = tieringParametersCfg.Map(&result); err != nil {
		slog.WarnContext(ctx, fmt.Sprintf("Invalid tiering configuration supplied; returning default values %+v", err))
	}

	return result
}

type FedEULACustomTextParameter struct {
	CustomText string `json:"custom_text,omitempty"`
}

// GetFedRAMPCustomEULA Note this is not gated by the FedEULA FF and that should be checked alongside this
func GetFedRAMPCustomEULA(ctx context.Context, service ParameterService) string {
	var result FedEULACustomTextParameter

	if fedEulaCustomText, err := service.GetConfigurationParameter(ctx, FedEULACustomTextKey); err != nil {
		slog.WarnContext(ctx, "Failed to fetch eula custom text; returning default value")
	} else if err = fedEulaCustomText.Map(&result); err != nil {
		slog.WarnContext(ctx, "Invalid eula custom text supplied; returning default value")
	}

	return result.CustomText
}

type SessionTTLHoursParameter struct {
	Hours int `json:"hours,omitempty"`
}

func GetSessionTTLHours(ctx context.Context, service ParameterService) time.Duration {
	var result = SessionTTLHoursParameter{
		Hours: DefaultSessionTTLHours, // Default to a logged in auth session time to live of 8 hours
	}

	if sessionTTLHours, err := service.GetConfigurationParameter(ctx, SessionTTLHours); err != nil {
		slog.WarnContext(ctx, "Failed to fetch auth session ttl hours; returning default values")
	} else if err = sessionTTLHours.Map(&result); err != nil {
		slog.WarnContext(ctx, "Invalid auth session ttl hours supplied; returning default values")
	} else if result.Hours <= 0 {
		slog.WarnContext(ctx, "auth session ttl hours ≤ 0; returning default values")
		result.Hours = DefaultSessionTTLHours
	}

	return time.Hour * time.Duration(result.Hours)
}
