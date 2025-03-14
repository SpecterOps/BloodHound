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
	"time"

	"github.com/specterops/bloodhound/src/model"
	"github.com/specterops/bloodhound/src/test/integration"
	"github.com/stretchr/testify/require"
)

func TestDatabase_CreateAssetGroupLabelSelector(t *testing.T) {
	var (
		dbInst          = integration.SetupDB(t)
		testCtx         = context.Background()
		testActor       = "test_actor"
		testName        = "test selector name"
		testDescription = "test description"
		isDefault       = false
		allowDisable    = true
		autoCertify     = false
		testSeeds       = []model.SelectorSeed{
			{Type: model.SelectorTypeObjectId, Value: "ObjectID1234"},
			{Type: model.SelectorTypeObjectId, Value: "ObjectID5678"},
		}
	)

	selector, err := dbInst.CreateAssetGroupLabelSelector(testCtx, 1, testActor, testName, testDescription, isDefault, allowDisable, autoCertify, testSeeds)
	require.NoError(t, err)
	require.Equal(t, 1, selector.AssetGroupLabelId)
	require.WithinDuration(t, time.Now(), selector.CreatedAt, time.Second)
	require.Equal(t, testActor, selector.CreatedBy)
	require.WithinDuration(t, time.Now(), selector.UpdatedAt, time.Second)
	require.Equal(t, testActor, selector.UpdatedBy)
	require.Empty(t, selector.DisabledAt)
	require.Empty(t, selector.DisabledBy)
	require.Equal(t, testName, selector.Name)
	require.Equal(t, testDescription, selector.Description)
	require.Equal(t, autoCertify, selector.AutoCertify)
	require.Equal(t, isDefault, selector.IsDefault)

	for idx, seed := range testSeeds {
		require.Equal(t, seed.Type, selector.Seeds[idx].Type)
		require.Equal(t, seed.Value, selector.Seeds[idx].Value)
	}
}

func TestDatabase_CreateAssetGroupLabel(t *testing.T) {
	var (
		dbInst          = integration.SetupDB(t)
		testCtx         = context.Background()
		tierId          = 1
		testActor       = "test_actor"
		testName        = "test label name"
		testDescription = "test label description"
	)

	label, err := dbInst.CreateAssetGroupLabel(testCtx, tierId, testActor, testName, testDescription)
	require.NoError(t, err)
	require.Equal(t, tierId, int(label.AssetGroupTierId.Int32))
	require.WithinDuration(t, time.Now(), label.CreatedAt, time.Second)
	require.Equal(t, testActor, label.CreatedBy)
	require.WithinDuration(t, time.Now(), label.UpdatedAt, time.Second)
	require.Equal(t, testActor, label.UpdatedBy)
	require.Empty(t, label.DeletedAt)
	require.Empty(t, label.DeletedBy)
	require.Equal(t, testName, label.Name)
	require.Equal(t, testDescription, label.Description)

	label, err = dbInst.GetAssetGroupLabel(testCtx, label.ID)
	require.NoError(t, err)
	require.Equal(t, tierId, int(label.AssetGroupTierId.Int32))
	require.WithinDuration(t, time.Now(), label.CreatedAt, time.Second)
	require.Equal(t, testActor, label.CreatedBy)
	require.WithinDuration(t, time.Now(), label.UpdatedAt, time.Second)
	require.Equal(t, testActor, label.UpdatedBy)
	require.Empty(t, label.DeletedAt)
	require.Empty(t, label.DeletedBy)
	require.Equal(t, testName, label.Name)
	require.Equal(t, testDescription, label.Description)

	label, err = dbInst.GetAssetGroupLabel(testCtx, 1234)
	require.Error(t, err)
}
