package appcfg_test

import (
	"context"
	"testing"

	"github.com/specterops/bloodhound/dawgs/drivers/neo4j"
	"github.com/specterops/bloodhound/src/database/types"
	"github.com/specterops/bloodhound/src/model/appcfg"
	"github.com/specterops/bloodhound/src/test/integration"
	"github.com/stretchr/testify/require"
)

func TestParameters_IsValidKey(t *testing.T) {
	parameter := appcfg.Parameter{}
	t.Run("should return false on invalid keys", func(t *testing.T) {
		require.False(t, parameter.IsValidKey(""))
	})

	t.Run("should return true on valid keys", func(t *testing.T) {
		require.True(t, parameter.IsValidKey(appcfg.PruneTTL))
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
	require.False(t, appcfg.GetCitrixRDPSupport(context.Background(), integration.SetupDB(t)))
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
