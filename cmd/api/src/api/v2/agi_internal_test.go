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

package v2

import (
	"net/http"
	"testing"

	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/graphschema/ad"
	"github.com/specterops/bloodhound/graphschema/azure"
	"github.com/specterops/bloodhound/graphschema/common"
	"github.com/specterops/bloodhound/src/api"
	"github.com/specterops/bloodhound/src/model"
	"github.com/stretchr/testify/require"
)

func TestGetLatestQueryParameter_LatestMalformed(t *testing.T) {
	req, err := http.NewRequest("GET", "/api/v2/asset-groups/1/collections", nil)
	require.Nil(t, err)

	q := req.URL.Query()
	q.Add("latest", "foo")
	req.URL.RawQuery = q.Encode()

	_, err = getLatestQueryParameter(req.URL.Query())
	require.NotNil(t, err)
	require.Equal(t, api.ErrorResponseDetailsLatestMalformed, err.Error())
}

func TestGetLatestQueryParameter(t *testing.T) {
	req, err := http.NewRequest("GET", "/api/v2/asset-groups/1/collections?latest", nil)
	require.Nil(t, err)

	latest, err := getLatestQueryParameter(req.URL.Query())
	require.Nil(t, err)
	require.True(t, latest)
}

func TestParseAGMembersFromNodes_(t *testing.T) {
	nodes := graph.NodeSet{
		1: &graph.Node{
			ID:    1,
			Kinds: graph.Kinds{ad.Entity, ad.Domain},
			Properties: &graph.Properties{
				Map: map[string]any{common.ObjectID.String(): "a", common.Name.String(): "a", ad.DomainSID.String(): "a"},
			},
		},
		2: &graph.Node{
			ID:    2,
			Kinds: graph.Kinds{azure.Entity, azure.Tenant},
			Properties: &graph.Properties{
				Map: map[string]any{common.ObjectID.String(): "b", common.Name.String(): "b", azure.TenantID.String(): "b"},
			},
		},
		3: &graph.Node{
			ID:    3,
			Kinds: graph.Kinds{ad.Entity, ad.Domain},
			Properties: &graph.Properties{
				Map: map[string]any{common.ObjectID.String(): "c", common.Name.String(): "c", ad.DomainSID.String(): "c"},
			},
		}}
	members := parseAGMembersFromNodes(nodes, model.AssetGroupSelectors{model.AssetGroupSelector{
		AssetGroupID:   1,
		Name:           "a",
		Selector:       "a",
		SystemSelector: false,
	}}, 1)
	require.Equal(t, 3, len(members))

	customMembersExpected, azureNodesExpected := 1, 1
	azureNodesFound, customMembersFound := 0, 0

	for _, member := range members {
		if member.CustomMember {
			customMembersFound++
		}
		if member.EnvironmentKind != ad.Domain.String() {
			azureNodesFound++
		}
	}

	require.Equal(t, customMembersExpected, customMembersFound)
	require.Equal(t, azureNodesExpected, azureNodesFound)
}

func TestParseAGMembersFromNodes_MissingNodeProperties(t *testing.T) {
	nodes := graph.NodeSet{
		// the parse fn should handle nodes with missing name and missing properties with warnings and no output
		1: &graph.Node{
			ID:    1,
			Kinds: graph.Kinds{ad.Entity, ad.Domain},
			Properties: &graph.Properties{
				Map: map[string]any{},
			},
		},
		2: &graph.Node{
			ID:    2,
			Kinds: graph.Kinds{azure.Entity, azure.Tenant},
			Properties: &graph.Properties{
				Map: map[string]any{},
			},
		},
	}

	members := parseAGMembersFromNodes(nodes,
		model.AssetGroupSelectors{model.AssetGroupSelector{
			AssetGroupID:   1,
			Name:           "a",
			Selector:       "a",
			SystemSelector: false,
		}}, 1)

	require.Equal(t, 0, len(members))
}
