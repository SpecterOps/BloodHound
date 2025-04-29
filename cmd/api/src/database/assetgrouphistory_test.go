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

	"github.com/gofrs/uuid"
	"github.com/specterops/bloodhound/src/database/types/null"
	"github.com/specterops/bloodhound/src/model"
	"github.com/specterops/bloodhound/src/test/integration"
	"github.com/stretchr/testify/require"
)

func TestDatabase_CreateAndGetAssetGroupHistory(t *testing.T) {
	var (
		dbInst    = integration.SetupDB(t)
		testCtx   = context.Background()
		testActor = model.User{
			EmailAddress: null.StringFrom("user@example.com"),
			Unique:       model.Unique{ID: uuid.FromStringOrNil("01234567-9012-4567-9012-456789012345")},
		}
		testTarget        = "test target"
		testAssetGroupTag = 1
	)

	err := dbInst.CreateAssetGroupHistoryRecord(testCtx, testActor, testTarget, model.AssetGroupHistoryActionDeleteSelector, testAssetGroupTag, null.String{}, null.String{})
	require.NoError(t, err)

	record, err := dbInst.GetAssetGroupHistoryRecords(testCtx)
	require.NoError(t, err)
	require.Len(t, record, 1)
	require.Equal(t, model.AssetGroupHistoryActionDeleteSelector, record[0].Action)
	require.Equal(t, testActor.ID.String(), record[0].Actor)
	require.Equal(t, testActor.EmailAddress, record[0].Email)
	require.Equal(t, testTarget, record[0].Target)
	require.Equal(t, testAssetGroupTag, record[0].AssetGroupTagId)
	require.Equal(t, null.String{}, record[0].EnvironmentId)
	require.Equal(t, null.String{}, record[0].Note)
	require.False(t, record[0].CreatedAt.IsZero())
}
