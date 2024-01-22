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

//go:build serial_integration
// +build serial_integration

package v2_test

import (
	"testing"

	"github.com/specterops/bloodhound/dawgs/drivers/neo4j"
	v2 "github.com/specterops/bloodhound/src/api/v2"
	"github.com/specterops/bloodhound/src/api/v2/integration"
	"github.com/specterops/bloodhound/src/model/appcfg"
	"github.com/stretchr/testify/require"
)

func Test_GetAppConfigs(t *testing.T) {
	var (
		passwordExpirationWindowFound = false
		neo4jConfigsFound             = false
		passwordExpirationValue       appcfg.PasswordExpiration
		neo4jParametersValue          appcfg.Neo4jParameters
		testCtx                       = integration.NewFOSSContext(t)
	)

	config, err := testCtx.AdminClient().GetAppConfigs()
	require.Nilf(t, err, "Error while getting app config: %v", err)

	for _, parameter := range config {
		switch parameter.Key {
		case appcfg.PasswordExpirationWindow:
			mapParameter(t, &passwordExpirationValue, parameter)
			require.Equal(t, appcfg.PasswordExpirationWindowName, parameter.Name)
			require.Equal(t, appcfg.PasswordExpirationWindowDescription, parameter.Description)
			require.Equal(t, appcfg.DefaultPasswordExpirationWindow, passwordExpirationValue.Duration)
			passwordExpirationWindowFound = true
		case appcfg.Neo4jConfigs:
			mapParameter(t, &neo4jParametersValue, parameter)
			require.Equal(t, appcfg.Neo4jConfigsName, parameter.Name)
			require.Equal(t, appcfg.Neo4jConfigsDescription, parameter.Description)
			require.Equal(t, neo4j.DefaultBatchWriteSize, neo4jParametersValue.BatchWriteSize)
			require.Equal(t, neo4j.DefaultWriteFlushSize, neo4jParametersValue.WriteFlushSize)
			neo4jConfigsFound = true
		}
	}

	require.True(t, passwordExpirationWindowFound, "Failed to find Password Expiration Window in response")
	require.True(t, neo4jConfigsFound, "Failed to find Neo4J Configs in response")
}

func Test_GetAppConfigWithParameter(t *testing.T) {
	var (
		passwordExpirationValue appcfg.PasswordExpiration
		testCtx                 = integration.NewFOSSContext(t)
	)

	config, err := testCtx.AdminClient().GetAppConfig(appcfg.PasswordExpirationWindow)
	require.Nilf(t, err, "Error while getting app config: %v", err)
	require.True(t, len(config) == 1, "Response contains too many results")
	require.Equal(t, appcfg.PasswordExpirationWindow, config[0].Key)
	mapParameter(t, &passwordExpirationValue, config[0])
	require.Equal(t, appcfg.PasswordExpirationWindowName, config[0].Name)
	require.Equal(t, appcfg.PasswordExpirationWindowDescription, config[0].Description)
	require.Equal(t, appcfg.DefaultPasswordExpirationWindow, passwordExpirationValue.Duration)
}

func Test_PutAppConfig(t *testing.T) {
	const updatedDuration = "P30D"

	var (
		updatedPasswordExpiration                appcfg.PasswordExpiration
		updatedPasswordExpirationWindowParameter = v2.AppConfigUpdateRequest{
			Key: appcfg.PasswordExpirationWindow,
			Value: map[string]any{
				"duration": updatedDuration,
			},
		}
		testCtx = integration.NewFOSSContext(t)
	)

	parameter, err := testCtx.AdminClient().PutAppConfig(updatedPasswordExpirationWindowParameter)
	require.Nilf(t, err, "Error while updating app config: %v", err)

	mapParameter(t, &updatedPasswordExpiration, parameter)
	require.Equal(t, updatedDuration, updatedPasswordExpiration.Duration)

	// Check that our change really is in the database
	config, err := testCtx.AdminClient().GetAppConfig(appcfg.PasswordExpirationWindow)
	require.Nilf(t, err, "Error while getting updated app config: %v", err)

	mapParameter(t, &updatedPasswordExpiration, config[0])
	require.Equal(t, updatedDuration, updatedPasswordExpiration.Duration)
}

func mapParameter[T *appcfg.PasswordExpiration | *appcfg.Neo4jParameters](t *testing.T, value T, parameter appcfg.Parameter) {
	err := parameter.Value.Map(&value)
	require.Nilf(t, err, "Failed to map parameter value to %T type: %v", value, err)
}
