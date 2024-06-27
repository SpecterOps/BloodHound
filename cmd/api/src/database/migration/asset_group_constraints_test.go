package migration_test

import (
    "context"
    "testing"
    "github.com/specterops/bloodhound/src/test/integration"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestAssetGroupConstraints(t *testing.T) {
    ctx := context.Background()
    dbInst := integration.SetupDB(t)

    t.Run("Duplicate Name", func(t *testing.T) {
        // Create initial asset group
        _, err := dbInst.CreateAssetGroup(ctx, "Test Group", "test_tag", false)
        require.NoError(t, err)

        // Attempt to create asset group with duplicate name
        _, err = dbInst.CreateAssetGroup(ctx, "Test Group", "different_tag", false)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "duplicate key value violates unique constraint")
		assert.Contains(t, err.Error(), "asset_groups_name_key")
    })

    t.Run("Duplicate Tag", func(t *testing.T) {
        // Create initial asset group
        _, err := dbInst.CreateAssetGroup(ctx, "Another Group", "another_tag", false)
        require.NoError(t, err)

        // Attempt to create asset group with duplicate tag
        _, err = dbInst.CreateAssetGroup(ctx, "Different Group", "another_tag", false)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "duplicate key value violates unique constraint")
		assert.Contains(t, err.Error(), "asset_groups_tag_key")
    })
}