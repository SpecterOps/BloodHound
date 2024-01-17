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
// +build serial_integration

package migration_test

import (
	"github.com/specterops/bloodhound/src/model"
	"github.com/specterops/bloodhound/src/test/integration"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestMigration_AssetGroups(t *testing.T) {
	// We expect a new DB to have the T0 group and the Owned group
	expectedNumAssetGroups := 2
	dbInst := integration.OpenDatabase(t)
	if err := integration.Prepare(dbInst); err != nil {
		t.Fatalf("Failed preparing DB: %v", err)
	}

	assetGroups, err := dbInst.GetAllAssetGroups("", model.SQLFilter{})
	require.Nil(t, err)
	require.Equal(t, expectedNumAssetGroups, len(assetGroups))
}
