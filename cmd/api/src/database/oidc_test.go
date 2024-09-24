//go:build integration
// +build integration

package database_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/specterops/bloodhound/src/test/integration"
	"github.com/stretchr/testify/require"
)

func TestBloodhoundDB_CreateOIDCProvider(t *testing.T) {
	var (
		testCtx = context.Background()
		dbInst  = integration.SetupDB(t)
	)
	defer dbInst.Close(testCtx)

	t.Run("successfully create an OIDC provider", func(t *testing.T) {
		provider, err := dbInst.CreateOIDCProvider(testCtx, "test", "https://test.localhost.com", "bloodhound")
		require.NoError(t, err)

		assert.Equal(t, "test", provider.Name)
		assert.Equal(t, "https://test.localhost.com", provider.LoginURL)
		assert.Equal(t, "bloodhound", provider.ClientID)
	})
}
