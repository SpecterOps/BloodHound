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
	"fmt"
	"time"

	iso8601 "github.com/channelmeter/iso8601duration"
	"github.com/specterops/bloodhound/dawgs/drivers/neo4j"
	"github.com/specterops/bloodhound/log"
	"github.com/specterops/bloodhound/src/database/types"
	"github.com/specterops/bloodhound/src/model"
)

const (
	PasswordExpirationWindow        = "auth.password_expiration_window"
	DefaultPasswordExpirationWindow = time.Hour * 24 * 90

	Neo4jConfigs        = "neo4j.configuration"
	PruneTTL            = "prune.ttl"
	CitrixRDPSupportKey = "analysis.citrix_rdp_support"
)

// Parameter is a runtime configuration parameter that can be fetched from the appcfg.ParameterService interface. The
// Value member is a DB-safe JSON type wrapper that can store arbitrary JSON objects and map them to golang struct
// definitions.
type Parameter struct {
	Key         string            `json:"key" gorm:"unique"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Value       types.JSONBObject `json:"value"`

	model.Serial
}

// Map is a convenience function for mapping the data stored in the Value Parameter struct member onto
// a richer type provided by the given value.
func (s Parameter) Map(value any) error {
	return s.Value.Map(value)
}

func (s Parameter) IsValid(parameter string) bool {
	validKeys := map[string]bool{
		PasswordExpirationWindow: true,
		Neo4jConfigs:             true,
		PruneTTL:                 true,
	}

	return validKeys[parameter]
}

// Parameters is a collection of Parameter structs.
type Parameters []Parameter

// ParameterService is a contract which defines expected functionality for fetching and setting Parameter from an
// abstract backend storage.
type ParameterService interface {
	// GetAllConfigurationParameters gets all available runtime Parameters for the application.
	GetAllConfigurationParameters(ctx context.Context) (Parameters, error)

	// GetConfigurationParameter attempts to fetch a Parameter struct by its parameter name.
	GetConfigurationParameter(ctx context.Context, parameter string) (Parameter, error)

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
		log.Warnf("Failed to fetch password expiratio configuration; returning default values")
		return DefaultPasswordExpirationWindow
	} else if err := cfg.Map(&expiration); err != nil {
		log.Warnf("Invalid password expiration configuration supplied; returning default values")
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
		log.Warnf("Failed to fetch neo4j configuration; returning default values")
	} else if err = neo4jParametersCfg.Map(&result); err != nil {
		log.Warnf("Invalid neo4j configuration supplied; returning default values")
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
		log.Warnf("Failed to fetch CitrixRDPSupport configuration; returning default values")
	} else if err := cfg.Map(&result); err != nil {
		log.Warnf("Invalid CitrixRDPSupport configuration supplied, %v. returning default values.", err)
	}

	return result.Enabled
}

// PruneTTL

type PruneTTLParameters struct {
	BaseTTL           time.Duration `json:"base_ttl,omitempty"`
	HasSessionEdgeTTL time.Duration `json:"has_session_edge_ttl,omitempty"`
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
			return err
		} else {
			s.BaseTTL = duration.ToDuration()
		}
		if duration, err := iso8601.FromString(pTTL.HasSessionEdgeTTL); err != nil {
			return err
		} else {
			s.HasSessionEdgeTTL = duration.ToDuration()
		}

		return nil
	}
}

func GetPruneTTLParameters(ctx context.Context, service ParameterService) PruneTTLParameters {
	var (
		day    = time.Hour * 24
		result = PruneTTLParameters{
			BaseTTL:           7 * day,
			HasSessionEdgeTTL: 3 * day,
		}
	)

	if pruneTTLParametersCfg, err := service.GetConfigurationParameter(ctx, PruneTTL); err != nil {
		log.Warnf("Failed to fetch prune TTL configuration; returning default values")
	} else if err = pruneTTLParametersCfg.Map(&result); err != nil {
		log.Warnf("Invalid prune TTL configuration supplied; returning default values %+v", err)
	}

	return result
}
