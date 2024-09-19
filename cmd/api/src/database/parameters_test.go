package database_test

import (
	"context"
	"testing"

	"github.com/specterops/bloodhound/src/database/types"
	"github.com/specterops/bloodhound/src/model/appcfg"
	"github.com/specterops/bloodhound/src/test/integration"
	"github.com/stretchr/testify/require"
)

func TestParameters_SetConfigurationParameter(t *testing.T) {
	var (
		testCtx = context.Background()
		dbInst  = integration.SetupDB(t)
	)
	newVal, err := types.NewJSONBObject(map[string]any{"enabled": false})
	require.Nil(t, err)

	updatedParameter := appcfg.Parameter{
		Key:   appcfg.ReconciliationKey,
		Value: newVal,
	}

	err = dbInst.SetConfigurationParameter(testCtx, updatedParameter)
	require.Nil(t, err)

	parameter, err := dbInst.GetConfigurationParameter(testCtx, appcfg.ReconciliationKey)
	require.Nil(t, err)

	result := appcfg.ReconciliationParameter{}
	err = parameter.Map(&result)
	require.Nil(t, err)

	require.Equal(t, updatedParameter.Key, parameter.Key)
	require.False(t, result.Enabled)
}

func TestParameters_GetConfigurationParameter(t *testing.T) {
	var (
		testCtx = context.Background()
		dbInst  = integration.SetupDB(t)
	)

	t.Run("get password expiration parameter", func(t *testing.T) {
		parameter, err := dbInst.GetConfigurationParameter(testCtx, appcfg.PasswordExpirationWindow)
		require.Nil(t, err)
		expected := &appcfg.Parameter{
			Key:         appcfg.PasswordExpirationWindow,
			Name:        "Local Auth Password Expiry Window",
			Description: "This configuration parameter sets the local auth password expiry window for users that have valid auth secrets. Values for this configuration must follow the duration specification of ISO-8601.",
		}
		require.Equal(t, expected.Key, parameter.Key)
		require.Equal(t, expected.Name, parameter.Name)
		require.Equal(t, expected.Description, parameter.Description)
	})

	t.Run("get neo4j configs parameter", func(t *testing.T) {
		parameter, err := dbInst.GetConfigurationParameter(testCtx, appcfg.Neo4jConfigs)
		require.Nil(t, err)
		expected := &appcfg.Parameter{
			Key:         appcfg.Neo4jConfigs,
			Name:        "Neo4j Configuration Parameters",
			Description: "This configuration parameter sets the BatchWriteSize and the BatchFlushSize for Neo4J.",
		}
		require.Equal(t, expected.Key, parameter.Key)
		require.Equal(t, expected.Name, parameter.Name)
		require.Equal(t, expected.Description, parameter.Description)
	})

	t.Run("get citrix rdp support parameter", func(t *testing.T) {
		parameter, err := dbInst.GetConfigurationParameter(testCtx, appcfg.CitrixRDPSupportKey)
		require.Nil(t, err)
		expected := &appcfg.Parameter{
			Key:         appcfg.CitrixRDPSupportKey,
			Name:        "Citrix RDP Support",
			Description: "This configuration parameter toggles Citrix support during post-processing. When enabled, computers identified with a 'Direct Access Users' local group will assume that Citrix is installed and CanRDP edges will require membership of both 'Direct Access Users' and 'Remote Desktop Users' local groups on the computer.",
		}
		require.Equal(t, expected.Key, parameter.Key)
		require.Equal(t, expected.Name, parameter.Name)
		require.Equal(t, expected.Description, parameter.Description)
	})

	t.Run("get prune ttl parameter", func(t *testing.T) {
		parameter, err := dbInst.GetConfigurationParameter(testCtx, appcfg.PruneTTL)
		require.Nil(t, err)
		expected := &appcfg.Parameter{
			Key:         appcfg.PruneTTL,
			Name:        "Prune Retention TTL Configuration Parameters",
			Description: "This configuration parameter sets the retention TTLs during analysis pruning.",
		}
		require.Equal(t, expected.Key, parameter.Key)
		require.Equal(t, expected.Name, parameter.Name)
		require.Equal(t, expected.Description, parameter.Description)
	})

	t.Run("get reconciliation parameter", func(t *testing.T) {
		parameter, err := dbInst.GetConfigurationParameter(testCtx, appcfg.ReconciliationKey)
		require.Nil(t, err)
		expected := &appcfg.Parameter{
			Key:         appcfg.ReconciliationKey,
			Name:        "Reconciliation",
			Description: "This configuration parameter enables / disables reconciliation during analysis.",
		}
		require.Equal(t, expected.Key, parameter.Key)
		require.Equal(t, expected.Name, parameter.Name)
		require.Equal(t, expected.Description, parameter.Description)
	})
}

func TestParameters_GetAllConfigurationParameter(t *testing.T) {
	var (
		testCtx = context.Background()
		dbInst  = integration.SetupDB(t)
	)
	parameters, err := dbInst.GetAllConfigurationParameters(testCtx)
	require.Nil(t, err)
	require.Len(t, parameters, 5)
	for _, parameter := range parameters {
		require.True(t, parameter.IsValidKey(parameter.Key))
	}
}
