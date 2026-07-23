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

	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSelectorNodeBatchValuesSQLArgs(t *testing.T) {
	t.Parallel()

	t.Run("Success -- Builds strict named arguments", func(t *testing.T) {
		t.Parallel()
		batchValues := newSelectorNodeBatchValues(make([]model.AssetGroupSelectorNode, 2))

		sqlArguments, err := batchValues.sqlArgs(2)

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
	})

	for _, test := range []struct {
		name              string
		argumentName      string
		removeColumnValue func(batchValues *selectorNodeBatchValues)
	}{
		{
			name:         "Rejects selector ID length mismatch",
			argumentName: "selector_ids",
			removeColumnValue: func(batchValues *selectorNodeBatchValues) {
				batchValues.selectorIds = batchValues.selectorIds[:1]
			},
		},
		{
			name:         "Rejects node ID length mismatch",
			argumentName: "node_ids",
			removeColumnValue: func(batchValues *selectorNodeBatchValues) {
				batchValues.nodeIds = batchValues.nodeIds[:1]
			},
		},
		{
			name:         "Rejects certification length mismatch",
			argumentName: "certifications",
			removeColumnValue: func(batchValues *selectorNodeBatchValues) {
				batchValues.certifications = batchValues.certifications[:1]
			},
		},
		{
			name:         "Rejects certified by length mismatch",
			argumentName: "certified_by_values",
			removeColumnValue: func(batchValues *selectorNodeBatchValues) {
				batchValues.certifiedByValues = batchValues.certifiedByValues[:1]
			},
		},
		{
			name:         "Rejects source length mismatch",
			argumentName: "sources",
			removeColumnValue: func(batchValues *selectorNodeBatchValues) {
				batchValues.sources = batchValues.sources[:1]
			},
		},
		{
			name:         "Rejects primary kind length mismatch",
			argumentName: "node_primary_kinds",
			removeColumnValue: func(batchValues *selectorNodeBatchValues) {
				batchValues.nodePrimaryKinds = batchValues.nodePrimaryKinds[:1]
			},
		},
		{
			name:         "Rejects environment ID length mismatch",
			argumentName: "node_environment_ids",
			removeColumnValue: func(batchValues *selectorNodeBatchValues) {
				batchValues.nodeEnvironmentIds = batchValues.nodeEnvironmentIds[:1]
			},
		},
		{
			name:         "Rejects object ID length mismatch",
			argumentName: "node_object_ids",
			removeColumnValue: func(batchValues *selectorNodeBatchValues) {
				batchValues.nodeObjectIds = batchValues.nodeObjectIds[:1]
			},
		},
		{
			name:         "Rejects node name length mismatch",
			argumentName: "node_names",
			removeColumnValue: func(batchValues *selectorNodeBatchValues) {
				batchValues.nodeNames = batchValues.nodeNames[:1]
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			batchValues := newSelectorNodeBatchValues(make([]model.AssetGroupSelectorNode, 2))
			test.removeColumnValue(&batchValues)

			_, err := batchValues.sqlArgs(2)

			assert.EqualError(t, err, "selector node batch argument \""+test.argumentName+"\" has 1 values; expected 2")
		})
	}
}
