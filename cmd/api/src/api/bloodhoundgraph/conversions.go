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

package bloodhoundgraph

import (
	"maps"

	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/packages/go/graphschema"
	"github.com/specterops/bloodhound/packages/go/graphschema/common"
	"github.com/specterops/dawgs/graph"
)

const (
	defaultNodeBorderColor     = "black"
	defaultNodeBackgroundColor = "rgba(255,255,255,0.9)"
	defaultNodeFontSize        = 14
	defaultRelationshipColor   = "3a5464"
)

func NodeToBloodHoundGraph(graphSchemaNodeValidDisplayKinds model.GraphSchemaNodeKindMap, customNodeKinds model.CustomNodeKindMap, node *graph.Node) BloodHoundGraphNode {
	validPrimaryKinds := graphSchemaNodeValidDisplayKinds.ToKindsMap()
	// Add custom node kinds to valid primary kinds
	maps.Copy(validPrimaryKinds, customNodeKinds.ValidKinds())
	var (
		nodeKindLabel       = model.GetNodeKindDisplayLabel(validPrimaryKinds, node)
		name, _             = node.Properties.GetWithFallback(common.Name.String(), graphschema.DefaultMissingName, common.DisplayName.String(), common.ObjectID.String()).String()
		bloodHoundGraphNode = BloodHoundGraphNode{
			BloodHoundGraphItem: &BloodHoundGraphItem{
				Data: getNodeDisplayProperties(validPrimaryKinds, node),
			},
			Size: 1,
			Border: &BloodHoundGraphNodeBorder{
				Color: defaultNodeBorderColor,
			},
			Label: &BloodHoundGraphNodeLabel{
				Text:            name,
				BackgroundColor: defaultNodeBackgroundColor,
				FontSize:        defaultNodeFontSize,
				Center:          true,
			},
		}
	)

	bloodHoundGraphNode.SetFontIcon(nodeKindLabel, graphSchemaNodeValidDisplayKinds, customNodeKinds)

	return bloodHoundGraphNode
}

func RelationshipToBloodHoundGraph(rel *graph.Relationship) BloodHoundGraphLink {
	var relProperties map[string]any

	if rel.Properties != nil {
		relProperties = rel.Properties.Map
	}

	return BloodHoundGraphLink{
		BloodHoundGraphItem: &BloodHoundGraphItem{
			Color: defaultRelationshipColor,
			Data:  relProperties,
		},
		Label: &BloodHoundGraphLinkLabel{
			Text: rel.Kind.String(),
		},
		End2: &BloodHoundGraphLinkEnd{
			Arrow: true,
		},
		ID1: rel.StartID.String(),
		ID2: rel.EndID.String(),
	}
}

func PathSetToBloodHoundGraph(graphSchemaNodeValidDisplayKinds model.GraphSchemaNodeKindMap, customNodeKinds model.CustomNodeKindMap, paths graph.PathSet) map[string]any {
	result := make(map[string]any)

	for _, path := range paths.Paths() {
		for _, rel := range path.Edges {
			result["rel_"+rel.ID.String()] = RelationshipToBloodHoundGraph(rel)
		}
	}

	for _, node := range paths.AllNodes() {
		result[node.ID.String()] = NodeToBloodHoundGraph(graphSchemaNodeValidDisplayKinds, customNodeKinds, node)
	}

	return result
}
