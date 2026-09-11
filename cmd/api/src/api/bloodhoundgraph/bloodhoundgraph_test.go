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

package bloodhoundgraph

import (
	"testing"

	"github.com/specterops/bloodhound/packages/go/graphschema"
	"github.com/specterops/dawgs/graph"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetFontIcon(t *testing.T) {
	t.Run("sets meta image for Meta node without admintier", func(t *testing.T) {
		node := &BloodHoundGraphNode{
			BloodHoundGraphItem: &BloodHoundGraphItem{
				Data: map[string]any{},
			},
		}

		node.SetFontIcon("Meta", nil)

		assert.Equal(t, "#000", node.Color)
		assert.Equal(t, "/ui/meta.png", node.Image)
		assert.Nil(t, node.FontIcon)
	})

	t.Run("sets tier zero meta image when admintier is 0", func(t *testing.T) {
		node := &BloodHoundGraphNode{
			BloodHoundGraphItem: &BloodHoundGraphItem{
				Data: map[string]any{
					"admintier": int64(0),
				},
			},
		}

		node.SetFontIcon("Meta", nil)

		assert.Equal(t, "#000", node.Color)
		assert.Equal(t, "/ui/metat0.png", node.Image)
		assert.Nil(t, node.FontIcon)
	})

	t.Run("sets default meta image when admintier is non-zero", func(t *testing.T) {
		node := &BloodHoundGraphNode{
			BloodHoundGraphItem: &BloodHoundGraphItem{
				Data: map[string]any{
					"admintier": int64(1),
				},
			},
		}

		node.SetFontIcon("Meta", nil)

		assert.Equal(t, "#000", node.Color)
		assert.Equal(t, "/ui/meta.png", node.Image)
		assert.Nil(t, node.FontIcon)
	})

	t.Run("sets font icon and color for known schema node kind", func(t *testing.T) {
		nodeKindMap := graphschema.PrimaryDisplayKinds{
			graph.StringKind("User"): graphschema.DisplayKind{
				Name: "User",
				Icon: graphschema.DisplayNodeIcon{
					Name:  "user",
					Color: "#17E625",
					Type:  graphschema.DisplayNodeTypeFontAwesome,
				},
			},
		}

		node := &BloodHoundGraphNode{
			BloodHoundGraphItem: &BloodHoundGraphItem{
				Data: map[string]any{},
			},
		}

		node.SetFontIcon("User", nodeKindMap)

		require.NotNil(t, node.FontIcon)
		assert.Equal(t, "fas fa-user", node.FontIcon.Text)
		assert.Equal(t, "#17E625", node.Color)
		assert.Empty(t, node.Image)
	})

	t.Run("sets default unknown icon for unrecognized node kind", func(t *testing.T) {
		node := &BloodHoundGraphNode{
			BloodHoundGraphItem: &BloodHoundGraphItem{
				Data: map[string]any{},
			},
		}

		node.SetFontIcon("SomeUnknownKind", nil)

		require.NotNil(t, node.FontIcon)
		assert.Equal(t, "fas fa-question", node.FontIcon.Text)
		assert.Equal(t, "#EEE", node.Color)
		assert.Empty(t, node.Image)
	})

	t.Run("Meta takes priority over schema map entry", func(t *testing.T) {
		nodeKindMap := graphschema.PrimaryDisplayKinds{
			graphschema.Meta: graphschema.DisplayKind{
				Name: "Meta",
				Icon: graphschema.DisplayNodeIcon{
					Name:  "star",
					Color: "#FFF",
				},
			},
		}
		node := &BloodHoundGraphNode{
			BloodHoundGraphItem: &BloodHoundGraphItem{
				Data: map[string]any{},
			},
		}

		node.SetFontIcon("Meta", nodeKindMap)

		assert.Equal(t, "#000", node.Color)
		assert.Equal(t, "/ui/meta.png", node.Image)
		assert.Nil(t, node.FontIcon)
	})
}
