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
	PasswordExpirationWindow            = "auth.password_expiration_window"
	DefaultPasswordExpirationWindow     = "P90D"
	PasswordExpirationWindowName        = "Local Auth Password Expiry Window"
	PasswordExpirationWindowDescription = "This configuration parameter sets the local auth password expiry window for users that have valid auth secrets. Values for this configuration must follow the duration specification of ISO-8601."

	Neo4jConfigs            = "neo4j.configuration"
	Neo4jConfigsName        = "Neo4j Configuration Parameters"
	Neo4jConfigsDescription = "This configuration parameter sets the BatchWriteSize and the BatchFlushSize for Neo4J."

	CitrixRDPSupportKey         = "analysis.citrix_rdp_support"
	CitrixRDPSupportName        = "Citrix RDP Support"
	CitrixRDPSupportDescription = "This configuration parameter toggles Citrix support during post-processing. When enabled, computers identified with a 'Direct Access Users' local group will assume that Citrix is installed and CanRDP edges will require membership of both 'Direct Access Users' and 'Remote Desktop Users' local groups on the computer."
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

type PasswordExpiration struct {
	Duration time.Duration `json:"duration"`
}

// Because PasswordExpiration are stored as ISO strings, but we want to use them as durations, we override UnmarshalJSON to handle the conversion
func (s *PasswordExpiration) UnmarshalJSON(data []byte) error {
	pDb := struct {
		Duration string `json:"duration,omitempty"`
	}{}

	if err := json.Unmarshal(data, &pDb); err != nil {
		return fmt.Errorf("error unmarshaling data for PruneTTLParameters: %w", err)
	} else {
		if duration, err := iso8601.FromString(pDb.Duration); err != nil {
			return err
		} else {
			s.Duration = duration.ToDuration()
		}

		return nil
	}
}

func GetPasswordExpiration(ctx context.Context, service ParameterService) (time.Duration, error) {
	var expiration PasswordExpiration

	if cfg, err := service.GetConfigurationParameter(ctx, PasswordExpirationWindow); err != nil {
		return 0, err
	} else if err := cfg.Map(&expiration); err != nil {
		return 0, err
	}

	return expiration.Duration, nil
}

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
