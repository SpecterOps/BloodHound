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
	"errors"
	"log/slog"

	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/packages/go/analysis"
	"github.com/specterops/bloodhound/packages/go/graphschema"
	"github.com/specterops/bloodhound/packages/go/graphschema/ad"
	azureSchema "github.com/specterops/bloodhound/packages/go/graphschema/azure"
	"github.com/specterops/bloodhound/packages/go/graphschema/common"
	"github.com/specterops/dawgs/graph"
)

const (
	defaultNodeBorderColor     = "black"
	defaultNodeBackgroundColor = "rgba(255,255,255,0.9)"
	defaultNodeFontSize        = 14
	defaultRelationshipColor   = "3a5464"
)

func NodeToBloodHoundGraph(node *graph.Node) BloodHoundGraphNode {
	var (
		nodeKindLabel       = analysis.GetNodeKindDisplayLabel(node)
		name, _             = node.Properties.GetWithFallback(common.Name.String(), graphschema.DefaultMissingName, common.DisplayName.String(), common.ObjectID.String()).String()
		bloodHoundGraphNode = BloodHoundGraphNode{
			BloodHoundGraphItem: &BloodHoundGraphItem{
				Data: getNodeDisplayProperties(node),
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

	bloodHoundGraphNode.SetIcon(nodeKindLabel)
	bloodHoundGraphNode.SetBackground(nodeKindLabel)

	return bloodHoundGraphNode
}

func NodeToBloodHoundGraphWithOpenGraph(node *graph.Node) BloodHoundGraphNode {
	// TODO DRY this up
	var (
		nodeKindLabel       = analysis.GetNodeKindDisplayLabel(node)
		name, _             = node.Properties.GetWithFallback(common.Name.String(), graphschema.DefaultMissingName, common.DisplayName.String(), common.ObjectID.String()).String()
		bloodHoundGraphNode = BloodHoundGraphNode{
			BloodHoundGraphItem: &BloodHoundGraphItem{
				Data: getNodeDisplayProperties(node),
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

	// somehow, check if it's an opengraph node -- if it is call the DB to get the custom icon and color
	// will this work to identify OG nodes?
	if !node.Kinds.ContainsOneOf(ad.Entity, azureSchema.Entity) {
		if customNodeConfig, err := getOpenGraphNodeCustomIconConfig(node.Kinds); err != nil {
			// log error, default to defaults
			slog.Error("Error fetching custom icons from database", err)
			bloodHoundGraphNode.SetNodeStyle(nodeKindLabel)
		} else {
			bloodHoundGraphNode.FontIcon = &BloodHoundGraphFontIcon{
				//Text: "fas fa-window-restore",
				// TODO -- do I need to add the fas prefix here?
				Text: customNodeConfig.Icon.Name,
			}
			bloodHoundGraphNode.Color = customNodeConfig.Icon.Color
		}

	} else {
		bloodHoundGraphNode.SetNodeStyle(nodeKindLabel)
	}

	return bloodHoundGraphNode
}

func getOpenGraphNodeCustomIconConfig(kinds graph.Kinds) (model.CustomNodeKindConfig, error) {
	var customKindConfig model.CustomNodeKindConfig // TODO set this to the default
	for _, kind := range kinds {
		// see if the DB has an entry for that kind
		resources.DB.GetCustomNodeKind()
		// if it does, break and return

	}

	return nil, errors.New("no custom icons found for this kind")
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

func NodeSetToBloodHoundGraph(nodes graph.NodeSet, openGraphSearchEnabled bool) map[string]any {
	result := make(map[string]any, nodes.Len())

	if openGraphSearchEnabled {
		for _, node := range nodes {
			result[node.ID.String()] = NodeToBloodHoundGraphWithOpenGraph(node)
		}
	} else {
		for _, node := range nodes {
			result[node.ID.String()] = NodeToBloodHoundGraph(node)
		}

	}
	return result
}

func PathSetToBloodHoundGraph(paths graph.PathSet) map[string]any {
	result := make(map[string]any)

	for _, path := range paths.Paths() {
		for _, rel := range path.Edges {
			result["rel_"+rel.ID.String()] = RelationshipToBloodHoundGraph(rel)
		}
	}

	for _, node := range paths.AllNodes() {
		result[node.ID.String()] = NodeToBloodHoundGraph(node)
	}

	return result
}
