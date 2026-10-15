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

package main

import (
	"encoding/json"
	"testing"

	"github.com/specterops/bloodhound/packages/go/ein"
	"github.com/specterops/bloodhound/packages/go/graphschema/azure"
	"github.com/specterops/bloodhound/packages/go/graphschema/common"
	"github.com/specterops/dawgs/graph"
	"github.com/stretchr/testify/require"
)

func TestExportNode(t *testing.T) {
	var (
		properties    = graph.AsProperties(map[string]any{"objectid": "mixed-Case", "displayname": "Example", "enabled": false})
		node          = graph.PrepareNode(properties, azure.Entity, azure.User)
		exported, err = exportNode(node, azure.Entity)
	)
	require.NoError(t, err)
	require.Equal(t, "mixed-Case", exported.ID)
	require.Equal(t, []string{"AZUser", "AZBase"}, exported.Kinds)
	require.Equal(t, false, exported.Properties["enabled"])
	require.NotContains(t, exported.Properties, "objectid")
	require.Contains(t, node.Properties.Map, "objectid", "export must not mutate the database properties")

	properties.Set(common.PrimaryKind.String(), azure.Entity.String())
	exported, err = exportNode(node, azure.Entity)
	require.NoError(t, err)
	require.Equal(t, "AZBase", exported.Kinds[0])

	properties.Set(common.PrimaryKind.String(), "MissingKind")
	_, err = exportNode(node, azure.Entity)
	require.ErrorContains(t, err, "absent from its kinds")

	_, err = exportNode(graph.PrepareNode(graph.NewProperties(), azure.User), azure.Entity)
	require.ErrorContains(t, err, "no usable objectid")
}

func TestValidateDocument(t *testing.T) {
	var (
		result = document{
			Metadata: ein.GenericMetadata{SourceKind: "AZBase"},
			Graph: graphData{
				Nodes: []ein.GenericNode{{ID: "user", Kinds: []string{"AZUser", "AZBase"}, Properties: map[string]any{"enabled": false}}},
				Edges: []exportedEdge{{
					Start: exportedEndpoint{MatchBy: "id", Value: "user"},
					End:   exportedEndpoint{MatchBy: "id", Value: "user"},
					Kind:  "AZMemberOf", Properties: map[string]any{},
				}},
			},
		}
		content, err = json.Marshal(result)
	)
	require.NoError(t, err)
	require.NoError(t, validateDocument(content))

	result.Graph.Nodes[0].Properties["nested"] = map[string]any{"unsupported": true}
	content, err = json.Marshal(result)
	require.NoError(t, err)
	require.ErrorContains(t, validateDocument(content), "valid OpenGraph")
}
