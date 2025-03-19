// Copyright 2025 Specter Ops, Inc.
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
