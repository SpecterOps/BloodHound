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
	"github.com/specterops/bloodhound/packages/go/analysis"
	"github.com/specterops/bloodhound/packages/go/analysis/tiering"
	"github.com/specterops/dawgs/graph"
)

func getNodeDisplayProperties(target *graph.Node) map[string]any {
	properties := target.Properties.Map

	// Set the node level. This is legacy behavior that should be eventually refactored.
	if tiering.IsTierZero(target) {
		properties["level"] = 0
		// Set tier zero state to control glyph
		properties["isTierZero"] = true
	} else {
		properties["isTierZero"] = false
	}

	// Set the legacy node type
	properties["nodetype"] = analysis.GetNodeKindDisplayLabel(target)

	return properties
}

func SetAssetGroupPropertiesForNode(node *graph.Node) *graph.Node {
	node.Properties.Set("category", "Asset Groups")
	node.Properties.Set("type", analysis.GetNodeKindDisplayLabel(node))
	node.Properties.Set("level", 0)
	return node
}
