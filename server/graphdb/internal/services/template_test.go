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

package services

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRenderNodeKindInfoMarkdown(t *testing.T) {
	var (
		kindID  = int32(7)
		context = NodeContext{
			NodeID: 123,
			Kinds: []KindContext{
				{KindID: &kindID, Name: "User"},
				{Name: "Entity"},
			},
			Properties: map[string]any{
				"name": "alice",
			},
		}
	)

	tests := []struct {
		name    string
		content string
		want    string
	}{
		{
			name:    "easy mode: can render a single field from the NodeContext",
			content: `{"markdown":{"content":"{{.NodeID}}"}}`,
			want:    "123",
		},
		{
			name:    "renders node context and curated sprig functions",
			content: `{"markdown":{"content":"# {{ .Properties.name | upper }} ({{ .NodeID }})\\n{{ join \",\" .Kinds }}"}}`,
			want:    "# ALICE (123)\\nUser,Entity",
		},
		{
			name:    "renders a default value for missing property",
			content: `{"markdown":{"content":"{{ .Properties.missing | default \"Unknown\" }}"}}`,
			want:    "Unknown",
		},
		{
			name:    "renders empty output for empty content",
			content: `{"markdown":{"content":""}}`,
			want:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := renderKindInfoMarkdown(json.RawMessage(tt.content), context)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestRenderNodeKindInfoMarkdown_ReturnsErrorForUnknownField(t *testing.T) {
	content := json.RawMessage(`{"markdown":{"content":"{{ .UnknownField }}"}}`)

	_, err := renderKindInfoMarkdown(content, NodeContext{})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "can't evaluate field UnknownField")
}

func TestRenderMarkdownRejectsRemovedFunction(t *testing.T) {
	content := json.RawMessage(`{"markdown":{"content":"{{ regexMatch \"a\" \"a\" }}"}}`)

	_, err := renderKindInfoMarkdown(content, NodeContext{})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "function")
}

func TestNewNodeContext(t *testing.T) {
	var kindID = int32(7)
	node := Node{
		ID: 123,
		Kinds: []Kind{
			{ID: &kindID, Name: "User"},
		},
		Properties: map[string]any{"name": "alice"},
		KindInfos:  []KindInfo{{InfoKey: "should-not-be-exposed"}},
	}

	context := newNodeContext(node)

	assert.Equal(t, int64(123), context.NodeID)
	assert.Equal(t, []KindContext{{KindID: &kindID, Name: "User"}}, context.Kinds)
	assert.Equal(t, map[string]any{"name": "alice"}, context.Properties)
}
