// Copyright 2024 Specter Ops, Inc.
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

//go:build integration

package appcfg_test

import (
	"context"
	"testing"

	"github.com/specterops/dawgs/drivers/neo4j"
	"github.com/stretchr/testify/require"

	"github.com/specterops/bloodhound/cmd/api/src/database/types"
	"github.com/specterops/bloodhound/cmd/api/src/model/appcfg"
	"github.com/specterops/bloodhound/cmd/api/src/test/integration"
)

func TestParameters_IsValidKey(t *testing.T) {
	parameter := appcfg.Parameter{}
	t.Run("should return false on invalid keys", func(t *testing.T) {
		require.False(t, parameter.IsValidKey(""))
	})

	t.Run("should return true on valid keys", func(t *testing.T) {
		require.True(t, parameter.IsValidKey(appcfg.PruneTTL))
	})

	t.Run("should return true for scheduled analysis key", func(t *testing.T) {
		require.True(t, parameter.IsValidKey(appcfg.ScheduledAnalysis))
	})
}

func TestParameters_Validate(t *testing.T) {
	t.Run("should error on missing parent value key", func(t *testing.T) {
		parameter := appcfg.Parameter{}
		errs := parameter.Validate()
		require.Len(t, errs, 1)
		require.Equal(t, "missing or invalid property: value", errs[0].Error())
	})

	t.Run("should error on additional fields", func(t *testing.T) {
		val, err := types.NewJSONBObject(map[string]any{"enabled": true, "added_field": true})
		require.Nil(t, err)
		parameter := appcfg.Parameter{Value: val, Key: appcfg.ReconciliationKey}
		errs := parameter.Validate()
		require.Len(t, errs, 1)
		require.Equal(t, "value property contains an invalid field", errs[0].Error())
	})

	t.Run("should error on invalid parameter key", func(t *testing.T) {
		val, err := types.NewJSONBObject(map[string]any{"enabled": true})
		require.Nil(t, err)
		parameter := appcfg.Parameter{Value: val, Key: "invalid key"}
		errs := parameter.Validate()
		require.Len(t, errs, 1)
		require.Equal(t, "invalid key", errs[0].Error())
	})

	t.Run("should error for missing field", func(t *testing.T) {
		val, err := types.NewJSONBObject(map[string]any{"base_ttl": "P7D", "added_field": true})
		require.Nil(t, err)
		parameter := appcfg.Parameter{Value: val, Key: appcfg.PruneTTL}
		errs := parameter.Validate()
		require.Len(t, errs, 1)
		require.Equal(t, "missing or invalid has_session_edge_ttl", errs[0].Error())
	})

	t.Run("should error on unmarshal error for incorrect field value", func(t *testing.T) {
		val, err := types.NewJSONBObject(map[string]any{"base_ttl": true, "has_session_edge_ttl": "P7D"})
		require.Nil(t, err)
		parameter := appcfg.Parameter{Value: val, Key: appcfg.PruneTTL}
		errs := parameter.Validate()
		require.Len(t, errs, 1)
		require.Contains(t, errs[0].Error(), "base_ttl of type string")
	})

	t.Run("should error on invalid field per validation", func(t *testing.T) {
		val, err := types.NewJSONBObject(map[string]any{"base_ttl": "P7D", "has_session_edge_ttl": "P14D"})
		require.Nil(t, err)
		parameter := appcfg.Parameter{Value: val, Key: appcfg.PruneTTL}
		errs := parameter.Validate()
		require.Len(t, errs, 1)
		require.Equal(t, "HasSessionEdgeTTL: must be <= P7D", errs[0].Error())
	})

	t.Run("should pass validation", func(t *testing.T) {
		val, err := types.NewJSONBObject(map[string]any{"base_ttl": "P7D", "has_session_edge_ttl": "P7D"})
		require.Nil(t, err)
		parameter := appcfg.Parameter{Value: val, Key: appcfg.PruneTTL}
		errs := parameter.Validate()
		require.Len(t, errs, 0)
	})

	t.Run("should validate scheduled analysis with valid rrule", func(t *testing.T) {
		val, err := types.NewJSONBObject(map[string]any{"enabled": true, "rrule": "DTSTART:20240101T000000Z\nRRULE:FREQ=DAILY;INTERVAL=1"})
		require.Nil(t, err)
		parameter := appcfg.Parameter{Value: val, Key: appcfg.ScheduledAnalysis}
		errs := parameter.Validate()
		require.Len(t, errs, 0)
	})

	t.Run("should error on scheduled analysis with invalid rrule", func(t *testing.T) {
		val, err := types.NewJSONBObject(map[string]any{"enabled": true, "rrule": "invalid"})
		require.Nil(t, err)
		parameter := appcfg.Parameter{Value: val, Key: appcfg.ScheduledAnalysis}
		errs := parameter.Validate()
		require.Len(t, errs, 1)
		require.Contains(t, errs[0].Error(), "invalid rrule")
	})

	t.Run("should error on scheduled analysis missing dtstart", func(t *testing.T) {
		val, err := types.NewJSONBObject(map[string]any{"enabled": true, "rrule": "RRULE:FREQ=DAILY;INTERVAL=1"})
		require.Nil(t, err)
		parameter := appcfg.Parameter{Value: val, Key: appcfg.ScheduledAnalysis}
		errs := parameter.Validate()
		require.Len(t, errs, 1)
		require.Contains(t, errs[0].Error(), "dtstart is required")
	})

	t.Run("should validate scheduled analysis when disabled with empty rrule", func(t *testing.T) {
		val, err := types.NewJSONBObject(map[string]any{"enabled": false, "rrule": ""})
		require.Nil(t, err)
		parameter := appcfg.Parameter{Value: val, Key: appcfg.ScheduledAnalysis}
		errs := parameter.Validate()
		require.Len(t, errs, 0)
	})
}

func TestParameters_GetPasswordExpiration(t *testing.T) {
	require.Equal(t, appcfg.DefaultPasswordExpirationWindow, appcfg.GetPasswordExpiration(context.Background(), integration.SetupDB(t)))
}

func TestParameters_GetNeo4jParameters(t *testing.T) {
	result := appcfg.Neo4jParameters{
		WriteFlushSize: neo4j.DefaultWriteFlushSize,
		BatchWriteSize: neo4j.DefaultBatchWriteSize,
	}
	require.Equal(t, result, appcfg.GetNeo4jParameters(context.Background(), integration.SetupDB(t)))
}

func TestParameters_GetCitrixRDPSupport(t *testing.T) {
	require.True(t, appcfg.GetCitrixRDPSupport(context.Background(), integration.SetupDB(t)))
}

func TestParameters_GetPruneTTLParameters(t *testing.T) {
	result := appcfg.PruneTTLParameters{
		BaseTTL:           appcfg.DefaultPruneBaseTTL,
		HasSessionEdgeTTL: appcfg.DefaultPruneHasSessionEdgeTTL,
	}
	require.Equal(t, result, appcfg.GetPruneTTLParameters(context.Background(), integration.SetupDB(t)))
}

func TestParameters_GetReconciliationParameter(t *testing.T) {
	require.True(t, appcfg.GetReconciliationParameter(context.Background(), integration.SetupDB(t)))
}

func TestParameters_GetTimeoutLimitParameter(t *testing.T) {
	require.True(t, appcfg.GetTimeoutLimitParameter(context.Background(), integration.SetupDB(t)))
}

func TestParameters_GetAPITokensParameter(t *testing.T) {
	require.True(t, appcfg.GetAPITokensParameter(context.Background(), integration.SetupDB(t)))
}

func TestParameters_GetEnvironmentTargetedAccessControlParameters(t *testing.T) {
	result := appcfg.EnvironmentTargetedAccessControlParameters{
		Enabled: false,
	}
	require.Equal(t, result, appcfg.GetEnvironmentTargetedAccessControlParameters(context.Background(), integration.SetupDB(t)))
}

func TestParameters_GetScheduledAnalysisParameter(t *testing.T) {
	t.Run("should return default values when parameter not found", func(t *testing.T) {
		result, err := appcfg.GetScheduledAnalysisParameter(context.Background(), integration.SetupDB(t))
		require.NoError(t, err)
		require.False(t, result.Enabled)
		require.Equal(t, "", result.RRule)
	})

	t.Run("should return configured values when parameter exists", func(t *testing.T) {
		var (
			db  = integration.SetupDB(t)
			ctx = context.Background()
		)

		// Set a scheduled analysis configuration
		val, err := types.NewJSONBObject(map[string]any{
			"enabled": true,
			"rrule":   "DTSTART:20240101T000000Z\nRRULE:FREQ=DAILY;INTERVAL=1",
		})
		require.NoError(t, err)

		parameter := appcfg.Parameter{
			Key:   appcfg.ScheduledAnalysis,
			Value: val,
		}

		err = db.SetConfigurationParameter(ctx, parameter)
		require.NoError(t, err)

		// Retrieve and verify
		result, err := appcfg.GetScheduledAnalysisParameter(ctx, db)
		require.NoError(t, err)
		require.True(t, result.Enabled)
		require.Equal(t, "DTSTART:20240101T000000Z\nRRULE:FREQ=DAILY;INTERVAL=1", result.RRule)
	})

	t.Run("should return disabled state when configured as disabled", func(t *testing.T) {
		var (
			db  = integration.SetupDB(t)
			ctx = context.Background()
		)

		// Set a disabled scheduled analysis configuration
		val, err := types.NewJSONBObject(map[string]any{
			"enabled": false,
			"rrule":   "",
		})
		require.NoError(t, err)

		parameter := appcfg.Parameter{
			Key:   appcfg.ScheduledAnalysis,
			Value: val,
		}

		err = db.SetConfigurationParameter(ctx, parameter)
		require.NoError(t, err)

		// Retrieve and verify
		result, err := appcfg.GetScheduledAnalysisParameter(ctx, db)
		require.NoError(t, err)
		require.False(t, result.Enabled)
		require.Equal(t, "", result.RRule)
	})
}
