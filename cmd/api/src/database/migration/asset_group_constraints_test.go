// Copyright 2024 Specter Ops, Inc.
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
// +build integration

package migration_test

import (
	"context"
	"testing"

	"github.com/specterops/bloodhound/src/test/integration"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAssetGroupConstraints(t *testing.T) {
	t.Parallel()
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
