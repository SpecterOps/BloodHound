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

package migration_test

import (
	"context"
	"slices"
	"testing"

	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/cmd/api/src/test/integration"
	"github.com/stretchr/testify/require"
)

func TestMigration_AssetGroups(t *testing.T) {
	dbInst := integration.SetupDB(t)

	assetGroups, err := dbInst.GetAllAssetGroups(context.Background(), "", model.SQLFilter{})
	require.Nil(t, err)
	require.Len(t, assetGroups, 3)
	require.True(t, slices.ContainsFunc(assetGroups, func(assetGroup model.AssetGroup) bool {
		return assetGroup.Name == "Admin Tier Zero" && assetGroup.Tag == "admin_tier_0"
	}))
	require.True(t, slices.ContainsFunc(assetGroups, func(assetGroup model.AssetGroup) bool {
		return assetGroup.Name == "Owned" && assetGroup.Tag == "owned"
	}))
	require.True(t, slices.ContainsFunc(assetGroups, func(assetGroup model.AssetGroup) bool {
		return assetGroup.Name == "Decoy" && assetGroup.Tag == "decoy"
	}))
}
