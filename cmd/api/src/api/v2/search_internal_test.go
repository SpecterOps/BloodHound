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

package v2

import (
	"testing"

	"github.com/specterops/dawgs/graph"
	"github.com/specterops/bloodhound/graphschema/azure"
	"github.com/specterops/bloodhound/graphschema/common"
	"github.com/specterops/bloodhound/src/model"

	"github.com/stretchr/testify/assert"
)

func Test_SetNodeProperties(t *testing.T) {
	tests := []struct {
		name     string
		nodes    []*graph.Node
		expected model.DomainSelectors
	}{
		{
			name: "basic case",
			nodes: []*graph.Node{
				{
					ID: 1,
					Properties: &graph.Properties{
						Map: map[string]any{
							common.ObjectID.String():  "1",
							common.Name.String():      "Node1",
							common.Collected.String(): false},
					},
				},
			},
			expected: model.DomainSelectors{
				{
					Type:      "active-directory",
					Name:      "Node1",
					ObjectID:  "1",
					Collected: false,
				},
			},
		},
		{
			name: "azure tenant",
			nodes: []*graph.Node{
				{
					Properties: &graph.Properties{
						Map: map[string]any{
							common.ObjectID.String():  "2",
							common.Name.String():      "Node2",
							common.Collected.String(): true,
						},
					},
					Kinds: graph.Kinds{azure.Tenant},
				},
			},
			expected: model.DomainSelectors{
				{
					Type:      "azure",
					Name:      "Node2",
					ObjectID:  "2",
					Collected: true,
				},
			},
		},
		{
			name: "missing properties",
			nodes: []*graph.Node{
				{
					Properties: &graph.Properties{
						Map: map[string]any{},
					},
				},
			},
			expected: model.DomainSelectors{
				{
					Type:      "active-directory",
					Name:      "NO NAME",
					ObjectID:  "NO OBJECT ID",
					Collected: false,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := setNodeProperties(tt.nodes)
			assert.Equal(t, tt.expected, got)
		})
	}
}
