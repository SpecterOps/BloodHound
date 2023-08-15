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
	"strings"

	"github.com/specterops/bloodhound/analysis"
	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/graphschema/ad"
	"github.com/specterops/bloodhound/graphschema/common"
	"github.com/specterops/bloodhound/log"
)

// We ignore the property lookup errors here since there's no clear path for a caller to handle it. Logging is also
// restricted to debug verbosity as anything more permissive could potentially flood centralized infrastructure in
// production

func getNodeName(target *graph.Node) string {
	const noNameOrObjectID = "NO NAME OR ID"

	props := target.Properties

	if name, err := props.Get(common.Name.String()).String(); err == nil {
		return name
	} else if objectID, err := props.Get(common.ObjectID.String()).String(); err == nil {
		return objectID
	}

	return noNameOrObjectID
}

func getNodeLevel(target *graph.Node) (int, bool) {
	if startSystemTags, err := target.Properties.Get(common.SystemTags.String()).String(); err == nil {
		log.Debugf("Unable to find a %s property for node %d with kinds %v", common.SystemTags.String(), target.ID, target.Kinds)
	} else if strings.Contains(startSystemTags, ad.AdminTierZero) {
		return 0, true
	}

	return -1, false
}

func getNodeDisplayProperties(target *graph.Node) map[string]any {
	properties := target.Properties.Map

	// Set the node level. This is legacy behavior that should be eventually refactored. The UI should be able to
	// consume the system_tags and user_tags properties directly.
	if nodeLevel, hasLevel := getNodeLevel(target); hasLevel {
		properties["level"] = nodeLevel
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
