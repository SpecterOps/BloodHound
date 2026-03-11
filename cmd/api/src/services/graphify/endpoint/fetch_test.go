// Copyright 2026 Specter Ops, Inc.
//
// Licensed under the Apache License, Version 2.0
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0
package endpoint_test

import (
	"testing"

	"github.com/specterops/bloodhound/cmd/api/src/services/graphify/endpoint"
	"github.com/specterops/bloodhound/packages/go/ein"
	"github.com/stretchr/testify/assert"
)

func TestCopyMatchExpressionsSorted(t *testing.T) {
	tests := []struct {
		name     string
		input    []ein.MatchExpression
		expected []ein.MatchExpression
	}{
		{
			name:     "empty slice",
			input:    []ein.MatchExpression{},
			expected: []ein.MatchExpression{},
		},
		{
			name:     "nil slice",
			input:    nil,
			expected: []ein.MatchExpression{},
		},
		{
			name: "single element",
			input: []ein.MatchExpression{
				{Key: "Name", Operator: "equals", Value: "test"},
			},
			expected: []ein.MatchExpression{
				{Key: "Name", Operator: "equals", Value: "test"},
			},
		},
		{
			name: "already sorted by property",
			input: []ein.MatchExpression{
				{Key: "A", Operator: "equals", Value: 1},
				{Key: "B", Operator: "equals", Value: 2},
				{Key: "C", Operator: "equals", Value: 3},
			},
			expected: []ein.MatchExpression{
				{Key: "A", Operator: "equals", Value: 1},
				{Key: "B", Operator: "equals", Value: 2},
				{Key: "C", Operator: "equals", Value: 3},
			},
		},
		{
			name: "reverse sorted by property",
			input: []ein.MatchExpression{
				{Key: "Z", Operator: "equals", Value: 1},
				{Key: "M", Operator: "equals", Value: 2},
				{Key: "A", Operator: "equals", Value: 3},
			},
			expected: []ein.MatchExpression{
				{Key: "A", Operator: "equals", Value: 3},
				{Key: "M", Operator: "equals", Value: 2},
				{Key: "Z", Operator: "equals", Value: 1},
			},
		},
		{
			name: "same property, different operators (sorted by operator)",
			input: []ein.MatchExpression{
				{Key: "Name", Operator: "z_operator", Value: 1},
				{Key: "Name", Operator: "a_operator", Value: 2},
				{Key: "Name", Operator: "m_operator", Value: 3},
			},
			expected: []ein.MatchExpression{
				{Key: "Name", Operator: "a_operator", Value: 2},
				{Key: "Name", Operator: "m_operator", Value: 3},
				{Key: "Name", Operator: "z_operator", Value: 1},
			},
		},
		{
			name: "mixed sorting: primary property, secondary operator",
			input: []ein.MatchExpression{
				{Key: "B", Operator: "equals", Value: 1},
				{Key: "A", Operator: "equals", Value: 2},
				{Key: "B", Operator: "contains", Value: 3}, // Should come after B+equals if "contains" > "equals"
				{Key: "A", Operator: "contains", Value: 4},
			},
			expected: []ein.MatchExpression{
				{Key: "A", Operator: "contains", Value: 4},
				{Key: "A", Operator: "equals", Value: 2},
				{Key: "B", Operator: "contains", Value: 3},
				{Key: "B", Operator: "equals", Value: 1},
			},
		},
		{
			name: "complex unsorted mix",
			input: []ein.MatchExpression{
				{Key: "Name", Operator: "equals", Value: "Z"},
				{Key: "ID", Operator: "equals", Value: "100"},
				{Key: "Name", Operator: "contains", Value: "A"},
				{Key: "ID", Operator: "contains", Value: "50"},
			},
			expected: []ein.MatchExpression{
				{Key: "ID", Operator: "contains", Value: "50"},
				{Key: "ID", Operator: "equals", Value: "100"},
				{Key: "Name", Operator: "contains", Value: "A"},
				{Key: "Name", Operator: "equals", Value: "Z"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := endpoint.CopyMatchExpressionsSorted(tt.input)

			// Assert length
			assert.Equal(t, len(tt.expected), len(result), "Result slice length mismatch")

			// Assert content equality
			assert.Equal(t, tt.expected, result, "Result slice content mismatch")

			// Assert that the returned slice is a different memory address (copy)
			if len(tt.input) > 0 {
				assert.NotSame(t, &tt.input, &result, "Returned slice should be a new allocation, not the same pointer")
			}
		})
	}
}
