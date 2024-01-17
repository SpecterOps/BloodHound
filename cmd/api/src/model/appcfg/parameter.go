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
	"fmt"
	"time"

	"github.com/specterops/bloodhound/src/database/types"
	"github.com/specterops/bloodhound/src/model"
	iso8601 "github.com/channelmeter/iso8601duration"
	"github.com/specterops/bloodhound/dawgs/drivers/neo4j"
	"github.com/specterops/bloodhound/log"
)

const (
	PasswordExpirationWindow            = "auth.password_expiration_window"
	DefaultPasswordExpirationWindow     = "P90D"
	Neo4jConfigs                        = "neo4j.configuration"
	PasswordExpirationWindowName        = "Local Auth Password Expiry Window"
	Neo4jConfigsName                    = "Neo4j Configuration Parameters"
	PasswordExpirationWindowDescription = "This configuration parameter sets the local auth password expiry window for users that have valid auth secrets. Values for this configuration must follow the duration specification of ISO-8601."
	Neo4jConfigsDescription             = "This configuration parameter sets the BatchWriteSize and the BatchFlushSize for Neo4J."
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
	availParams, err := AvailableParameters()
	if err != nil {
		log.Errorf("Error occurred getting AvailableParamters: %v", err)
		return false
	}

	_, valid := availParams[parameter]
	return valid
}

// Parameters is a collection of Parameter structs.
type Parameters []Parameter

type ParameterSet map[string]Parameter

// ParameterService is a contract which defines expected functionality for fetching and setting Parameter from an
// abstract backend storage.
type ParameterService interface {
	// GetAllConfigurationParameters gets all available runtime Parameters for the application.
	GetAllConfigurationParameters() (Parameters, error)

	// GetConfigurationParameter attempts to fetch a Parameter struct by its parameter name.
	GetConfigurationParameter(parameter string) (Parameter, error)

	// GetConfigurationParametersByPrefix attempts to fetch all Parameters that have a parameter name that
	// starts with the given prefix.
	GetConfigurationParametersByPrefix(prefix string) (Parameters, error)

	// SetConfigurationParameter attempts to store or update the given Parameter.
	SetConfigurationParameter(configurationParameter Parameter) error
}

func AvailableParameters() (ParameterSet, error) {
	if passwordExpirationValue, err := types.NewJSONBObject(PasswordExpiration{
		Duration: DefaultPasswordExpirationWindow,
	}); err != nil {
		return ParameterSet{}, fmt.Errorf("error creating PasswordExpiration parameter: %w", err)
	} else if neo4jExpirationValue, err := types.NewJSONBObject(Neo4jParameters{
		BatchWriteSize: neo4j.DefaultBatchWriteSize,
		WriteFlushSize: neo4j.DefaultWriteFlushSize,
	}); err != nil {
		return ParameterSet{}, fmt.Errorf("error creating neo4jExpirationValue parameter: %w", err)
	} else {
		return ParameterSet{
			PasswordExpirationWindow: {
				Key:         PasswordExpirationWindow,
				Name:        PasswordExpirationWindowName,
				Description: PasswordExpirationWindowDescription,
				Value:       passwordExpirationValue,
				Serial:      model.Serial{},
			},
			Neo4jConfigs: {
				Key:         Neo4jConfigs,
				Name:        Neo4jConfigsName,
				Description: Neo4jConfigsDescription,
				Value:       neo4jExpirationValue,
			},
		}, nil
	}
}

type PasswordExpiration struct {
	Duration string `json:"duration"`
}

func (s PasswordExpiration) ParseDuration() (time.Duration, error) {
	if duration, err := iso8601.FromString(s.Duration); err != nil {
		return 0, err
	} else {
		return duration.ToDuration(), nil
	}
}

func GetPasswordExpiration(service ParameterService) (time.Duration, error) {
	var expiration PasswordExpiration

	if cfg, err := service.GetConfigurationParameter(PasswordExpirationWindow); err != nil {
		return 0, err
	} else if err := cfg.Map(&expiration); err != nil {
		return 0, err
	} else {
		return expiration.ParseDuration()
	}
}

type Neo4jParameters struct {
	WriteFlushSize int `json:"write_flush_size,omitempty"`
	BatchWriteSize int `json:"batch_write_size,omitempty"`
}

func GetNeo4jParameters(service ParameterService) Neo4jParameters {
	var result Neo4jParameters

	if neo4jParametersCfg, err := service.GetConfigurationParameter(Neo4jConfigs); err != nil {
		log.Errorf("failed to fetch neo4j configuration; returning default values")
		result = Neo4jParameters{
			WriteFlushSize: neo4j.DefaultWriteFlushSize,
			BatchWriteSize: neo4j.DefaultBatchWriteSize,
		}
	} else if err = neo4jParametersCfg.Map(result); err != nil {
		log.Errorf("invalid neo4j configuration supplied; returning default values")
		result = Neo4jParameters{
			WriteFlushSize: neo4j.DefaultWriteFlushSize,
			BatchWriteSize: neo4j.DefaultBatchWriteSize,
		}
	}

	return result
}
