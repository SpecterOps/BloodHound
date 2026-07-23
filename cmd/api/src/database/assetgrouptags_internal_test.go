// Copyright 2026 Specter Ops, Inc.
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

package database

import (
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSelectorNodeBatchSQLArgs(t *testing.T) {
	t.Parallel()

	sqlArguments, err := newSelectorNodeBatchSQLArgs(make([]model.AssetGroupSelectorNode, 2))

	require.NoError(t, err)
	require.Len(t, sqlArguments, 9)
	assert.Contains(t, sqlArguments, "selector_ids")
	assert.Contains(t, sqlArguments, "node_ids")
	assert.Contains(t, sqlArguments, "certifications")
	assert.Contains(t, sqlArguments, "certified_by_values")
	assert.Contains(t, sqlArguments, "sources")
	assert.Contains(t, sqlArguments, "node_primary_kinds")
	assert.Contains(t, sqlArguments, "node_environment_ids")
	assert.Contains(t, sqlArguments, "node_object_ids")
	assert.Contains(t, sqlArguments, "node_names")
}

func TestNewStrictNamedArrayArgs(t *testing.T) {
	t.Parallel()

	t.Run("Success -- Builds strict named arguments", func(t *testing.T) {
		t.Parallel()

		sqlArguments, err := newStrictNamedArrayArgs(
			2,
			newArrayArgument("first", []int{1, 2}),
			newArrayArgument("second", []string{"one", "two"}),
		)

		require.NoError(t, err)
		assert.Equal(t, pgtype.FlatArray[int]{1, 2}, sqlArguments["first"])
		assert.Equal(t, pgtype.FlatArray[string]{"one", "two"}, sqlArguments["second"])
	})

	t.Run("Error -- Rejects argument length mismatch", func(t *testing.T) {
		t.Parallel()

		_, err := newStrictNamedArrayArgs(
			2,
			newArrayArgument("values", []int{1}),
		)

		assert.EqualError(t, err, "strict named array argument \"values\" has 1 values; expected 2")
	})
}
