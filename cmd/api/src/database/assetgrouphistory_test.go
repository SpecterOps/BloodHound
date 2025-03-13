//go:build integration

package database_test

import (
	"context"
	"testing"

	"github.com/specterops/bloodhound/src/model"
	"github.com/specterops/bloodhound/src/test/integration"
	"github.com/stretchr/testify/require"
)

func TestDatabase_CreateAndGetAssetGroupHistory(t *testing.T) {
	var (
		dbInst  = integration.SetupDB(t)
		testCtx = context.Background()
	)

	err := dbInst.CreateAssetGroupHistoryRecord(testCtx, model.AssetGroupHistoryActorSystem, "", model.AssetGroupHistoryActionDeleteSelector, 1, "", "")
	require.NoError(t, err)

	record, err := dbInst.GetAssetGroupHistoryRecords(testCtx)
	require.NoError(t, err)
	require.Len(t, record, 1)
	require.Equal(t, model.AssetGroupHistoryActionDeleteSelector, record[0].Action)
	require.Equal(t, model.AssetGroupHistoryActorSystem, record[0].Actor)
}
