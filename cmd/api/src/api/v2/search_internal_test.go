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

	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/packages/go/graphschema"
	"github.com/specterops/bloodhound/packages/go/graphschema/ad"
	"github.com/specterops/bloodhound/packages/go/graphschema/azure"
	"github.com/specterops/bloodhound/packages/go/graphschema/common"
	"github.com/specterops/dawgs/graph"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_SetNodeProperties(t *testing.T) {
	var (
		schemaExtensionID            int32 = 7
		schemaEnvironmentKindID      int32 = 11
		adSchemaExtensionID          int32 = 21
		adSchemaEnvironmentKindID    int32 = 31
		azureSchemaExtensionID       int32 = 22
		azureSchemaEnvironmentKindID int32 = 32
	)

	tests := []struct {
		name     string
		nodes    []*graph.Node
		expected model.EnvironmentSelectors
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
					Kinds: graph.Kinds{ad.Domain},
				},
			},
			expected: model.EnvironmentSelectors{
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
			expected: model.EnvironmentSelectors{
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
			expected: model.EnvironmentSelectors{
				{
					Type:      "",
					Name:      graphschema.DefaultMissingName,
					ObjectID:  graphschema.DefaultMissingObjectId,
					Collected: false,
				},
			},
		},
		{
			name: "opengraph environment includes schema extension id",
			nodes: []*graph.Node{
				{
					Properties: &graph.Properties{
						Map: map[string]any{
							common.ObjectID.String():     "3",
							common.Name.String():         "Node3",
							common.Collected.String():    true,
							graphschema.EnvironmentIDKey: "environment-3",
						},
					},
					Kinds: graph.Kinds{graph.StringKind("OpenGraphEnvironment")},
				},
			},
			expected: model.EnvironmentSelectors{
				{
					Type:                    "OpenGraph Extension",
					Name:                    "Node3",
					ObjectID:                "environment-3",
					Collected:               true,
					SchemaExtensionID:       &schemaExtensionID,
					SchemaEnvironmentKindID: &schemaEnvironmentKindID,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BuildEnvironmentSelectors(tt.nodes, map[string]string{
				"OpenGraphEnvironment": "OpenGraph Extension",
				ad.Domain.String():     "Active Directory",
				azure.Tenant.String():  "Azure",
			}, map[string]int32{
				"OpenGraphEnvironment": schemaExtensionID,
				ad.Domain.String():     adSchemaExtensionID,
				azure.Tenant.String():  azureSchemaExtensionID,
			}, map[string]int32{
				"OpenGraphEnvironment": schemaEnvironmentKindID,
				ad.Domain.String():     adSchemaEnvironmentKindID,
				azure.Tenant.String():  azureSchemaEnvironmentKindID,
			})
			assert.Equal(t, tt.expected, got, tt.name)
		})
	}
}

func Test_filterAndFormatSearchResults(t *testing.T) {
	var (
		inputNodeProps = graph.NewProperties().
				Set("name", "this is a name").
				Set("objectid", "object id").
				Set("distinguishedname", "ze most distinguished")

		input = []*graph.Node{
			{Properties: inputNodeProps},
		}
		primaryDisplayKinds = make(graphschema.PrimaryDisplayKinds)
	)
	primaryDisplayKinds.Add("Person", "person-half-dress", "ff91af", graphschema.DisplayNodeTypeFontAwesome)

	actual := filterAndFormatSearchResults(input, nil, primaryDisplayKinds)

	expectedName, _ := inputNodeProps.Get("name").String()
	expectedObjectId, _ := inputNodeProps.Get("objectid").String()
	expectedDistinguishedName, _ := inputNodeProps.Get("distinguishedname").String()

	require.Equal(t, 1, len(actual))
	require.Equal(t, expectedName, "this is a name")
	require.Equal(t, expectedObjectId, "object id")
	require.Equal(t, expectedDistinguishedName, "ze most distinguished")
}

func Test_filterAndFormatSearchResults_default(t *testing.T) {
	var (
		input = []*graph.Node{
			{Properties: graph.NewProperties()},
		}
		expectedName              = graphschema.DefaultMissingName
		expectedObjectId          = graphschema.DefaultMissingObjectId
		expectedDistinguishedName = ""

		primaryDisplayKinds = make(graphschema.PrimaryDisplayKinds)
	)
	primaryDisplayKinds.Add("Person", "person-half-dress", "ff91af", graphschema.DisplayNodeTypeFontAwesome)

	actual := filterAndFormatSearchResults(input, nil, primaryDisplayKinds)

	require.Equal(t, 1, len(actual))
	require.Equal(t, expectedName, actual[0].Name)
	require.Equal(t, expectedObjectId, actual[0].ObjectID)
	require.Equal(t, expectedDistinguishedName, actual[0].DistinguishedName)
}

func Test_filterAndFormatSearchResults_includeOpenGraphNodes(t *testing.T) {
	var (
		customKind     = "CustomKind"
		inputNodeProps = graph.NewProperties().
				Set("name", "this is a name").
				Set("objectid", "object id")
		input = []*graph.Node{
			{Kinds: []graph.Kind{graph.StringKind("OtherKind"), graph.StringKind(customKind)},
				Properties: inputNodeProps},
		}
		primaryDisplayKinds = make(graphschema.PrimaryDisplayKinds)
	)
	primaryDisplayKinds.Add("CustomKind", "person-half-dress", "ff91af", graphschema.DisplayNodeTypeFontAwesome)

	actual := filterAndFormatSearchResults(input, nil, primaryDisplayKinds)

	require.Equal(t, 1, len(actual))
	require.Equal(t, customKind, actual[0].Type)
}

func Test_filterAndFormatSearchResults_filterEnvironments(t *testing.T) {
	var (
		inputNodeProp1 = graph.Node{
			ID:    1,
			Kinds: graph.Kinds{ad.Entity, ad.Domain},
			Properties: &graph.Properties{
				Map: map[string]any{common.ObjectID.String(): "objectid1", common.Name.String(): "name1", ad.DomainSID.String(): "12345"},
			},
		}
		inputNodeProp2 = graph.Node{
			ID:    2,
			Kinds: graph.Kinds{ad.Entity, ad.Domain},
			Properties: &graph.Properties{
				Map: map[string]any{common.ObjectID.String(): "objectid2", common.Name.String(): "name2", ad.DomainSID.String(): "54321"},
			},
		}
		inputNodeProp3 = graph.Node{
			ID:    3,
			Kinds: graph.Kinds{azure.Entity, azure.Tenant},
			Properties: &graph.Properties{
				Map: map[string]any{common.ObjectID.String(): "objectid3", common.Name.String(): "name3", azure.TenantID.String(): "azure12345"},
			},
		}

		input = []*graph.Node{&inputNodeProp1, &inputNodeProp2, &inputNodeProp3}

		primaryDisplayKinds = make(graphschema.PrimaryDisplayKinds)
	)
	primaryDisplayKinds.Add("Person", "person-half-dress", "ff91af", graphschema.DisplayNodeTypeFontAwesome)

	actual := filterAndFormatSearchResults(input, []string{"54321"}, primaryDisplayKinds)

	expectedName, _ := inputNodeProp2.Properties.Get(common.Name.String()).String()
	expectedObjectId, _ := inputNodeProp2.Properties.Get(common.ObjectID.String()).String()

	require.Equal(t, 1, len(actual))
	actualResult := actual[0]
	require.Equal(t, expectedName, actualResult.Name)
	require.Equal(t, expectedObjectId, actualResult.ObjectID)
}

func Test_filterAndFormatSearchResults_filterEnvironmentsEmpty(t *testing.T) {
	var (
		inputNodeProp1 = graph.Node{
			ID:    1,
			Kinds: graph.Kinds{ad.Entity, ad.Domain},
			Properties: &graph.Properties{
				Map: map[string]any{common.ObjectID.String(): "objectid1", common.Name.String(): "name1", ad.DomainSID.String(): "12345"},
			},
		}
		inputNodeProp2 = graph.Node{
			ID:    2,
			Kinds: graph.Kinds{ad.Entity, ad.Domain},
			Properties: &graph.Properties{
				Map: map[string]any{common.ObjectID.String(): "objectid2", common.Name.String(): "name2", ad.DomainSID.String(): "54321"},
			},
		}
		inputNodeProp3 = graph.Node{
			ID:    3,
			Kinds: graph.Kinds{azure.Entity, azure.Tenant},
			Properties: &graph.Properties{
				Map: map[string]any{common.ObjectID.String(): "objectid3", common.Name.String(): "name3", azure.TenantID.String(): "azure12345"},
			},
		}

		input = []*graph.Node{&inputNodeProp1, &inputNodeProp2, &inputNodeProp3}

		primaryDisplayKinds = make(graphschema.PrimaryDisplayKinds)
	)
	primaryDisplayKinds.Add("Person", "person-half-dress", "ff91af", graphschema.DisplayNodeTypeFontAwesome)

	actual := filterAndFormatSearchResults(input, []string{}, primaryDisplayKinds)

	require.Empty(t, actual)
}

func Test_filterAndFormatSearchResults_filterEnvironments_domainSIDFail(t *testing.T) {
	var (
		inputNodeProp1 = graph.Node{
			ID:    1,
			Kinds: graph.Kinds{ad.Entity, ad.Domain},
			Properties: &graph.Properties{
				Map: map[string]any{common.ObjectID.String(): "objectid1", common.Name.String(): "name1", ad.DomainSID.String(): "12345"},
			},
		}
		inputNodeProp2 = graph.Node{
			ID:    2,
			Kinds: graph.Kinds{ad.Entity, ad.Domain},
			Properties: &graph.Properties{
				Map: map[string]any{common.ObjectID.String(): "objectid2", common.Name.String(): "name2"},
			},
		}
		inputNodeProp3 = graph.Node{
			ID:    3,
			Kinds: graph.Kinds{azure.Entity, azure.Tenant},
			Properties: &graph.Properties{
				Map: map[string]any{common.ObjectID.String(): "objectid3", common.Name.String(): "name3", azure.TenantID.String(): "azure12345"},
			},
		}

		input               = []*graph.Node{&inputNodeProp1, &inputNodeProp2, &inputNodeProp3}
		primaryDisplayKinds = make(graphschema.PrimaryDisplayKinds)
	)
	primaryDisplayKinds.Add("Person", "person-half-dress", "ff91af", graphschema.DisplayNodeTypeFontAwesome)

	result := filterAndFormatSearchResults(input, []string{"54321"}, primaryDisplayKinds)
	require.Len(t, result, 0)
}

func Test_filterAndFormatSearchResults_filterEnvironments_tenantIDFail(t *testing.T) {
	var (
		inputNodeProp1 = graph.Node{
			ID:    1,
			Kinds: graph.Kinds{ad.Entity, ad.Domain},
			Properties: &graph.Properties{
				Map: map[string]any{common.ObjectID.String(): "objectid1", common.Name.String(): "name1", ad.DomainSID.String(): "12345"},
			},
		}
		inputNodeProp2 = graph.Node{
			ID:    2,
			Kinds: graph.Kinds{ad.Entity, ad.Domain},
			Properties: &graph.Properties{
				Map: map[string]any{common.ObjectID.String(): "objectid2", common.Name.String(): "name2", ad.DomainSID.String(): "54321"},
			},
		}
		inputNodeProp3 = graph.Node{
			ID:    3,
			Kinds: graph.Kinds{azure.Entity, azure.Tenant},
			Properties: &graph.Properties{
				Map: map[string]any{common.ObjectID.String(): "objectid3", common.Name.String(): "name3"},
			},
		}

		input = []*graph.Node{&inputNodeProp1, &inputNodeProp2, &inputNodeProp3}

		primaryDisplayKinds = make(graphschema.PrimaryDisplayKinds)
	)
	primaryDisplayKinds.Add("Person", "person-half-dress", "ff91af", graphschema.DisplayNodeTypeFontAwesome)

	result := filterAndFormatSearchResults(input, []string{"azure12345"}, primaryDisplayKinds)
	require.Len(t, result, 0)
}

func Test_filterAndFormatSearchResults_filterEnvironmentsOG(t *testing.T) {
	var (
		inputNodeProp1 = graph.Node{
			ID:    1,
			Kinds: graph.Kinds{ad.Entity, ad.Domain},
			Properties: &graph.Properties{
				Map: map[string]any{common.ObjectID.String(): "objectid1", common.Name.String(): "name1", ad.DomainSID.String(): "12345"},
			},
		}
		inputNodeProp2 = graph.Node{
			ID:    2,
			Kinds: graph.Kinds{ad.Entity, ad.Domain},
			Properties: &graph.Properties{
				Map: map[string]any{common.ObjectID.String(): "objectid2", common.Name.String(): "name2", ad.DomainSID.String(): "54321"},
			},
		}
		inputNodeProp3 = graph.Node{
			ID:    3,
			Kinds: graph.Kinds{graph.StringKind("OtherKind")},
			Properties: &graph.Properties{
				Map: map[string]any{common.ObjectID.String(): "objectid3", common.Name.String(): "name3", graphschema.EnvironmentIDKey: "og-12345"},
			},
		}

		input = []*graph.Node{&inputNodeProp1, &inputNodeProp2, &inputNodeProp3}

		primaryDisplayKinds = make(graphschema.PrimaryDisplayKinds)
	)
	primaryDisplayKinds.Add("Person", "person-half-dress", "ff91af", graphschema.DisplayNodeTypeFontAwesome)

	actual := filterAndFormatSearchResults(input, []string{"og-12345"}, primaryDisplayKinds)

	require.Len(t, actual, 1)
	require.Equal(t, "objectid3", actual[0].ObjectID)
}

func Test_getSearchableNodeKinds(t *testing.T) {
	tests := []struct {
		name                   string
		openGraphSearchEnabled bool
		primaryDisplayKinds    graphschema.PrimaryDisplayKinds
		typeParams             graph.Kinds
		expected               graph.Kinds
		expectErr              bool
	}{
		{
			name:                   "returns unconstrained kinds when open graph search is enabled and no types are provided",
			openGraphSearchEnabled: true,
			primaryDisplayKinds:    graphschema.PrimaryDisplayKinds{graph.StringKind("CustomKind"): graphschema.DisplayKind{}},
			typeParams:             nil,
			expected:               nil,
		},
		{
			name:                   "returns default entity kinds when open graph search is disabled and no types are provided",
			openGraphSearchEnabled: false,
			typeParams:             nil,
			expected:               graph.Kinds{ad.Entity, azure.Entity},
		},
		{
			name:                   "returns custom kinds when open graph search is enabled",
			openGraphSearchEnabled: true,
			primaryDisplayKinds:    graphschema.PrimaryDisplayKinds{graph.StringKind("CustomKind"): graphschema.DisplayKind{}},
			typeParams:             graph.Kinds{graph.StringKind("CustomKind")},
			expected:               graph.Kinds{graph.StringKind("CustomKind")},
		},
		{
			name:                   "returns an error when all provided types are invalid",
			openGraphSearchEnabled: false,
			typeParams:             graph.Kinds{graph.StringKind("CustomKind")},
			expectErr:              true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := getSearchableNodeKinds(tt.openGraphSearchEnabled, tt.primaryDisplayKinds, tt.typeParams)

			if tt.expectErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expected, actual)
		})
	}
}
