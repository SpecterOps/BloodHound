// Copyright 2026 Specter Ops, Inc.
//
// Licensed under the Apache License, Version 2.0
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

package model

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func baseExtensionInput() ExtensionInput {
	return ExtensionInput{
		Name:        "Test extension",
		DisplayName: "Test extension",
		Version:     "v1.0.0",
		Namespace:   "AD",
	}
}

func Test_validateGraphExtension(t *testing.T) {
	type args struct {
		graphExtension GraphExtensionInput
	}
	tests := []struct {
		name    string
		args    args
		wantErr error
	}{
		{
			name: "fail - empty extension name",
			args: args{
				graphExtension: GraphExtensionInput{},
			},
			wantErr: fmt.Errorf("graph schema extension name is required"),
		},
		{
			name: "fail - empty extension version",
			args: args{
				graphExtension: GraphExtensionInput{
					ExtensionInput: ExtensionInput{
						Name:        "Test extension",
						DisplayName: "Test extension",
					},
				},
			},
			wantErr: fmt.Errorf("graph schema extension version is required"),
		},
		{
			name: "fail - invalid extension version",
			args: args{
				graphExtension: GraphExtensionInput{
					ExtensionInput: ExtensionInput{
						Name:        "Test extension",
						DisplayName: "Test extension",
						Version:     "1.0",
					},
				},
			},
			wantErr: fmt.Errorf("graph schema extension version is not valid semver: prefix `v` is missing"),
		},
		{
			name: "fail - empty extension namespace",
			args: args{
				graphExtension: GraphExtensionInput{
					ExtensionInput: ExtensionInput{
						Name:        "Test extension",
						DisplayName: "Test extension",
						Version:     "v1.0.0",
					},
				},
			},
			wantErr: fmt.Errorf("graph schema extension namespace is required"),
		},
		{
			name: "fail - extension namespace cannot be Tag",
			args: args{
				graphExtension: GraphExtensionInput{
					ExtensionInput: ExtensionInput{
						Name:        "Test extension",
						DisplayName: "Test extension",
						Version:     "v1.0.0",
						Namespace:   "Tag",
					},
				},
			},
			wantErr: fmt.Errorf("graph schema extension namespace 'Tag' uses reserved namespace 'tag'"),
		},
		{
			name: "fail - extension namespace cannot be tag (lowercase)",
			args: args{
				graphExtension: GraphExtensionInput{
					ExtensionInput: ExtensionInput{
						Name:        "Test extension",
						DisplayName: "Test extension",
						Version:     "v1.0.0",
						Namespace:   "tag",
					},
				},
			},
			wantErr: fmt.Errorf("graph schema extension namespace 'tag' uses reserved namespace 'tag'"),
		},
		{
			name: "fail - extension namespace cannot be TAG (uppercase)",
			args: args{
				graphExtension: GraphExtensionInput{
					ExtensionInput: ExtensionInput{
						Name:        "Test extension",
						DisplayName: "Test extension",
						Version:     "v1.0.0",
						Namespace:   "TAG",
					},
				},
			},
			wantErr: fmt.Errorf("graph schema extension namespace 'TAG' uses reserved namespace 'tag'"),
		},
		{
			name: "fail - extension namespace cannot be tAg (mixed case)",
			args: args{
				graphExtension: GraphExtensionInput{
					ExtensionInput: ExtensionInput{
						Name:        "Test extension",
						DisplayName: "Test extension",
						Version:     "v1.0.0",
						Namespace:   "tAg",
					},
				},
			},
			wantErr: fmt.Errorf("graph schema extension namespace 'tAg' uses reserved namespace 'tag'"),
		},
		{
			name: "fail - extension namespace cannot be prefixed by reserved namespace (tag_sub)",
			args: args{
				graphExtension: GraphExtensionInput{
					ExtensionInput: ExtensionInput{
						Name:        "Test extension",
						DisplayName: "Test extension",
						Version:     "v1.0.0",
						Namespace:   "tag_sub",
					},
				},
			},
			wantErr: fmt.Errorf("graph schema extension namespace 'tag_sub' uses reserved namespace 'tag'"),
		},
		{
			name: "fail - extension namespace cannot be prefixed by reserved namespace mixed-case (Tag_audit)",
			args: args{
				graphExtension: GraphExtensionInput{
					ExtensionInput: ExtensionInput{
						Name:        "Test extension",
						DisplayName: "Test extension",
						Version:     "v1.0.0",
						Namespace:   "Tag_audit",
					},
				},
			},
			wantErr: fmt.Errorf("graph schema extension namespace 'Tag_audit' uses reserved namespace 'tag'"),
		},
		{
			name: "fail - extension namespace cannot be prefixed by reserved namespace uppercase (TAG_x)",
			args: args{
				graphExtension: GraphExtensionInput{
					ExtensionInput: ExtensionInput{
						Name:        "Test extension",
						DisplayName: "Test extension",
						Version:     "v1.0.0",
						Namespace:   "TAG_x",
					},
				},
			},
			wantErr: fmt.Errorf("graph schema extension namespace 'TAG_x' uses reserved namespace 'tag'"),
		},
		{
			name: "pass reserved-namespace check - similar-looking namespace 'tagged' is not reserved (fails at later check)",
			args: args{
				graphExtension: GraphExtensionInput{
					ExtensionInput: ExtensionInput{
						Name:        "Test extension",
						DisplayName: "Test extension",
						Version:     "v1.0.0",
						Namespace:   "tagged",
					},
				},
			},
			wantErr: fmt.Errorf("graph schema node kinds are required"),
		},
		{
			name: "fail - empty graph schema nodes",
			args: args{
				graphExtension: GraphExtensionInput{
					ExtensionInput: baseExtensionInput(),
				},
			},
			wantErr: fmt.Errorf("graph schema node kinds are required"),
		},
		{
			name: "fail - duplicate kinds - two node kinds",
			args: args{
				graphExtension: GraphExtensionInput{
					ExtensionInput: baseExtensionInput(),
					NodeKindsInput: NodesInput{
						{
							Name: "AD_node kind 1",
						},
						{
							Name: "AD_node kind 1",
						},
					},
				},
			},
			wantErr: fmt.Errorf("duplicate graph kinds: AD_node kind 1"),
		},
		{
			name: "fail - node kind missing namespace",
			args: args{
				graphExtension: GraphExtensionInput{
					ExtensionInput: baseExtensionInput(),
					NodeKindsInput: NodesInput{
						{
							Name: "node_kind_1",
						},
						{
							Name: "AD_node kind 2",
						},
					},
				},
			},
			wantErr: fmt.Errorf("graph schema node kind node_kind_1 is missing extension namespace prefix"),
		},
		{
			name: "fail - node kind missing name after extension namespace",
			args: args{
				graphExtension: GraphExtensionInput{
					ExtensionInput: baseExtensionInput(),
					NodeKindsInput: NodesInput{
						{
							Name: "AD_",
						},
						{
							Name: "AD_node kind 2",
						},
					},
				},
			},
			wantErr: fmt.Errorf("graph schema node kind cannot be empty after the namespace prefix"),
		},
		{
			name: "fail - node kind icon color is not valid hex code",
			args: args{
				graphExtension: GraphExtensionInput{
					ExtensionInput: baseExtensionInput(),
					NodeKindsInput: NodesInput{
						{
							Name:      "AD_Kind1",
							IconColor: "#1234567",
						},
					},
				},
			},
			wantErr: fmt.Errorf("invalid hex color string #1234567 for node kind AD_Kind1"),
		},
		{
			name: "fail - duplicate kinds - two edge kinds",
			args: args{
				graphExtension: GraphExtensionInput{
					ExtensionInput: baseExtensionInput(),
					NodeKindsInput: NodesInput{
						{
							Name: "AD_node kind 1",
						},
						{
							Name: "AD_node kind 2",
						},
					},
					RelationshipKindsInput: RelationshipsInput{
						{
							Name: "AD_edge kind 1",
						},
						{
							Name: "AD_edge kind 1",
						},
					},
				},
			},
			wantErr: fmt.Errorf("duplicate graph kinds: AD_edge kind 1"),
		},
		{
			name: "fail - edge kind missing namespace",
			args: args{
				graphExtension: GraphExtensionInput{
					ExtensionInput: baseExtensionInput(),
					NodeKindsInput: NodesInput{
						{
							Name: "AD_node kind 1",
						},
						{
							Name: "AD_node kind 2",
						},
					},
					RelationshipKindsInput: RelationshipsInput{
						{
							Name: "AD_edge kind 1",
						},
						{
							Name: "edge_kind_1",
						},
					},
				},
			},
			wantErr: fmt.Errorf("graph schema edge kind edge_kind_1 is missing extension namespace prefix"),
		},
		{
			name: "fail - edge kind missing name after extension namespace",
			args: args{
				graphExtension: GraphExtensionInput{
					ExtensionInput: baseExtensionInput(),
					NodeKindsInput: NodesInput{
						{
							Name: "AD_node kind 1",
						},
						{
							Name: "AD_node kind 2",
						},
					},
					RelationshipKindsInput: RelationshipsInput{
						{
							Name: "AD_edge_kind_1",
						},
						{
							Name: "AD_",
						},
					},
				},
			},
			wantErr: fmt.Errorf("graph schema edge kind cannot be empty after the namespace prefix"),
		},
		{
			name: "fail - duplicate kinds - same edge and node kind",
			args: args{
				graphExtension: GraphExtensionInput{
					ExtensionInput: baseExtensionInput(),
					NodeKindsInput: NodesInput{
						{
							Name: "AD_a_duplicate_graph_kind",
						},
						{
							Name: "AD_node kind 2",
						},
					},
					RelationshipKindsInput: RelationshipsInput{
						{
							Name: "AD_edge kind 1",
						},
						{
							Name: "AD_a_duplicate_graph_kind",
						},
					},
					PropertiesInput: PropertiesInput{
						{
							Name: "property 1",
						},
						{
							Name: "property 2",
						},
					},
				},
			},
			wantErr: fmt.Errorf("duplicate graph kinds: %s", "AD_a_duplicate_graph_kind"),
		},
		{
			name: "fail - environment kind name missing namespace prefix",
			args: args{
				graphExtension: GraphExtensionInput{
					ExtensionInput: baseExtensionInput(),
					NodeKindsInput: NodesInput{
						{
							Name: "AD_node_kind_1",
						},
						{
							Name: "AD_node kind 2",
						},
					},
					RelationshipKindsInput: RelationshipsInput{
						{
							Name: "AD_edge kind 1",
						},
					},
					PropertiesInput: PropertiesInput{
						{
							Name: "property 1",
						},
						{
							Name: "property 2",
						},
					},
					EnvironmentsInput: EnvironmentsInput{
						{
							EnvironmentKindName: "environment",
							SourceKindName:      "",
							PrincipalKinds:      nil,
						},
					},
				},
			},
			wantErr: fmt.Errorf("graph schema environment kind %s is missing extension namespace prefix", "environment"),
		},
		{
			name: "fail - EnvironmentKindName cannot be empty after the namespace prefix",
			args: args{
				graphExtension: GraphExtensionInput{
					ExtensionInput: baseExtensionInput(),
					NodeKindsInput: NodesInput{
						{
							Name: "AD_node_kind_1",
						},
						{
							Name: "AD_node kind 2",
						},
					},
					RelationshipKindsInput: RelationshipsInput{
						{
							Name: "AD_edge kind 1",
						},
					},
					PropertiesInput: PropertiesInput{
						{
							Name: "property 1",
						},
						{
							Name: "property 2",
						},
					},
					EnvironmentsInput: EnvironmentsInput{
						{
							EnvironmentKindName: "AD_",
							SourceKindName:      "",
							PrincipalKinds:      nil,
						},
					},
				},
			},
			wantErr: fmt.Errorf("graph schema environment kind cannot be empty after the namespace prefix"),
		},
		{
			name: "fail - environment kind not declared as a node kind",
			args: args{
				graphExtension: GraphExtensionInput{
					ExtensionInput: baseExtensionInput(),
					NodeKindsInput: NodesInput{
						{
							Name: "AD_node_kind_1",
						},
						{
							Name: "AD_node kind 2",
						},
					},
					RelationshipKindsInput: RelationshipsInput{
						{
							Name: "AD_edge kind 1",
						},
					},
					PropertiesInput: PropertiesInput{
						{
							Name: "property 1",
						},
						{
							Name: "property 2",
						},
					},
					EnvironmentsInput: EnvironmentsInput{
						{
							EnvironmentKindName: "AD_environment",
							SourceKindName:      "",
							PrincipalKinds:      nil,
						},
					},
				},
			},
			wantErr: fmt.Errorf("graph schema environment %s not declared as a node kind", "AD_environment"),
		},
		{
			name: "fail - duplicate environments",
			args: args{
				graphExtension: GraphExtensionInput{
					ExtensionInput: baseExtensionInput(),
					NodeKindsInput: NodesInput{
						{
							Name: "AD_node_kind_1",
						},
						{
							Name: "AD_environment",
						},
					},
					RelationshipKindsInput: RelationshipsInput{
						{
							Name: "AD_edge kind 1",
						},
					},
					PropertiesInput: PropertiesInput{
						{
							Name: "property 1",
						},
						{
							Name: "property 2",
						},
					},
					EnvironmentsInput: EnvironmentsInput{
						{
							EnvironmentKindName: "AD_environment",
							SourceKindName:      "source_kind",
							PrincipalKinds:      []string{"AD_node_kind_1"},
						},
						{
							EnvironmentKindName: "AD_environment",
							SourceKindName:      "source_kind",
							PrincipalKinds:      nil,
						},
					},
				},
			},
			wantErr: fmt.Errorf("duplicate graph environments: AD_environment"),
		},
		{
			name: "fail - environment source kind is empty",
			args: args{
				graphExtension: GraphExtensionInput{
					ExtensionInput: baseExtensionInput(),
					NodeKindsInput: NodesInput{
						{
							Name: "AD_node_kind_1",
						},
						{
							Name: "AD_env_kind",
						},
					},
					RelationshipKindsInput: RelationshipsInput{
						{
							Name: "AD_edge kind 1",
						},
					},
					PropertiesInput: PropertiesInput{
						{
							Name: "property 1",
						},
						{
							Name: "property 2",
						},
					},
					EnvironmentsInput: EnvironmentsInput{
						{
							EnvironmentKindName: "AD_env_kind",
							SourceKindName:      "",
							PrincipalKinds:      nil,
						},
					},
				},
			},
			wantErr: fmt.Errorf("graph schema environment source kind cannot be empty"),
		},
		{
			name: "fail - environment source kind name conflicts with existing node kind name",
			args: args{
				graphExtension: GraphExtensionInput{
					ExtensionInput: baseExtensionInput(),
					NodeKindsInput: NodesInput{
						{
							Name: "AD_node_kind_1",
						},
						{
							Name: "AD_env_kind",
						},
					},
					RelationshipKindsInput: RelationshipsInput{
						{
							Name: "AD_edge kind 1",
						},
					},
					PropertiesInput: PropertiesInput{
						{
							Name: "property 1",
						},
						{
							Name: "property 2",
						},
					},
					EnvironmentsInput: EnvironmentsInput{
						{
							EnvironmentKindName: "AD_env_kind",
							SourceKindName:      "AD_node_kind_1",
							PrincipalKinds:      []string{"AD_node_kind_1"},
						},
					},
				},
			},
			wantErr: fmt.Errorf("graph schema environment source kind name %s conflicts with existing node kind", "AD_node_kind_1"),
		},
		{
			name: "fail - environment source kind name conflicts with existing relationship kind name",
			args: args{
				graphExtension: GraphExtensionInput{
					ExtensionInput: baseExtensionInput(),
					NodeKindsInput: NodesInput{
						{
							Name: "AD_node_kind_1",
						},
						{
							Name: "AD_env_kind",
						},
					},
					RelationshipKindsInput: RelationshipsInput{
						{
							Name: "AD_edge kind 1",
						},
					},
					PropertiesInput: PropertiesInput{
						{
							Name: "property 1",
						},
						{
							Name: "property 2",
						},
					},
					EnvironmentsInput: EnvironmentsInput{
						{
							EnvironmentKindName: "AD_env_kind",
							SourceKindName:      "AD_edge kind 1",
							PrincipalKinds:      []string{"AD_node_kind_1"},
						},
					},
				},
			},
			wantErr: fmt.Errorf("graph schema environment source kind name %s conflicts with existing relationship kind", "AD_edge kind 1"),
		},
		{
			name: "fail - environment principal kind missing namespace prefix",
			args: args{
				graphExtension: GraphExtensionInput{
					ExtensionInput: baseExtensionInput(),
					NodeKindsInput: NodesInput{
						{
							Name: "AD_node_kind_1",
						},
						{
							Name: "AD_env_kind",
						},
					},
					RelationshipKindsInput: RelationshipsInput{
						{
							Name: "AD_edge kind 1",
						},
					},
					PropertiesInput: PropertiesInput{
						{
							Name: "property 1",
						},
						{
							Name: "property 2",
						},
					},
					EnvironmentsInput: EnvironmentsInput{
						{
							EnvironmentKindName: "AD_env_kind",
							SourceKindName:      "Base",
							PrincipalKinds:      []string{"node_kind_1"},
						},
					},
				},
			},
			wantErr: fmt.Errorf("graph schema environment principal kind %s is missing extension namespace prefix", "node_kind_1"),
		},
		{
			name: "fail - environment principal kind cannot be empty after the namespace prefix",
			args: args{
				graphExtension: GraphExtensionInput{
					ExtensionInput: baseExtensionInput(),
					NodeKindsInput: NodesInput{
						{
							Name: "AD_node_kind_1",
						},
						{
							Name: "AD_env_kind",
						},
					},
					RelationshipKindsInput: RelationshipsInput{
						{
							Name: "AD_edge kind 1",
						},
					},
					PropertiesInput: PropertiesInput{
						{
							Name: "property 1",
						},
						{
							Name: "property 2",
						},
					},
					EnvironmentsInput: EnvironmentsInput{
						{
							EnvironmentKindName: "AD_env_kind",
							SourceKindName:      "Base",
							PrincipalKinds:      []string{"AD_"},
						},
					},
				},
			},
			wantErr: fmt.Errorf("graph schema environment principal kind cannot be empty after the namespace prefix"),
		},
		{
			name: "fail - environment principal kind not declared as a node kind",
			args: args{
				graphExtension: GraphExtensionInput{
					ExtensionInput: baseExtensionInput(),
					NodeKindsInput: NodesInput{
						{
							Name: "AD_node_kind_1",
						},
						{
							Name: "AD_env_kind",
						},
					},
					RelationshipKindsInput: RelationshipsInput{
						{
							Name: "AD_edge kind 1",
						},
					},
					PropertiesInput: PropertiesInput{
						{
							Name: "property 1",
						},
						{
							Name: "property 2",
						},
					},
					EnvironmentsInput: EnvironmentsInput{
						{
							EnvironmentKindName: "AD_env_kind",
							SourceKindName:      "Base",
							PrincipalKinds:      []string{"AD_node_kind_MISSING"},
						},
					},
				},
			},
			wantErr: fmt.Errorf("graph schema environment principal kind %s not declared node kind", "AD_node_kind_MISSING"),
		},
		{
			name: "fail - relationship finding name missing namespace prefix",
			args: args{
				graphExtension: GraphExtensionInput{
					ExtensionInput: baseExtensionInput(),
					NodeKindsInput: NodesInput{
						{
							Name: "AD_node_kind_1",
						},
						{
							Name: "AD_env_kind",
						},
					},
					RelationshipKindsInput: RelationshipsInput{
						{
							Name: "AD_edge kind 1",
						},
					},
					PropertiesInput: PropertiesInput{
						{
							Name: "property 1",
						},
						{
							Name: "property 2",
						},
					},
					EnvironmentsInput: EnvironmentsInput{
						{
							EnvironmentKindName: "AD_env_kind",
							SourceKindName:      "Base",
							PrincipalKinds:      []string{"AD_node_kind_1"},
						},
					},
					RelationshipFindingsInput: RelationshipFindingsInput{
						{
							Name: "finding_1",
						},
					},
				},
			},
			wantErr: fmt.Errorf("graph schema relationship finding %s is missing extension namespace prefix", "finding_1"),
		},
		{
			name: "fail - relationship finding name cannot be empty after the namespace prefix",
			args: args{
				graphExtension: GraphExtensionInput{
					ExtensionInput: baseExtensionInput(),
					NodeKindsInput: NodesInput{
						{
							Name: "AD_node_kind_1",
						},
						{
							Name: "AD_env_kind",
						},
					},
					RelationshipKindsInput: RelationshipsInput{
						{
							Name: "AD_edge kind 1",
						},
					},
					PropertiesInput: PropertiesInput{
						{
							Name: "property 1",
						},
						{
							Name: "property 2",
						},
					},
					EnvironmentsInput: EnvironmentsInput{
						{
							EnvironmentKindName: "AD_env_kind",
							SourceKindName:      "Base",
							PrincipalKinds:      []string{"AD_node_kind_1"},
						},
					},
					RelationshipFindingsInput: RelationshipFindingsInput{
						{
							Name: "AD_",
						},
					},
				},
			},
			wantErr: fmt.Errorf("graph schema relationship finding cannot be empty after the namespace prefix"),
		},
		{
			name: "fail - relationship finding relationship kind name missing namespace prefix",
			args: args{
				graphExtension: GraphExtensionInput{
					ExtensionInput: baseExtensionInput(),
					NodeKindsInput: NodesInput{
						{
							Name: "AD_node_kind_1",
						},
						{
							Name: "AD_env_kind",
						},
					},
					RelationshipKindsInput: RelationshipsInput{
						{
							Name: "AD_edge kind 1",
						},
					},
					PropertiesInput: PropertiesInput{
						{
							Name: "property 1",
						},
						{
							Name: "property 2",
						},
					},
					EnvironmentsInput: EnvironmentsInput{
						{
							EnvironmentKindName: "AD_env_kind",
							SourceKindName:      "Base",
							PrincipalKinds:      []string{"AD_node_kind_1"},
						},
					},
					RelationshipFindingsInput: RelationshipFindingsInput{
						{
							Name:                 "AD_finding_1",
							EnvironmentKindName:  "AD_env_kind",
							RelationshipKindName: "edge kind 1",
						},
					},
				},
			},
			wantErr: fmt.Errorf("graph schema relationship finding relationship kind %s is missing extension namespace prefix", "edge kind 1"),
		},
		{
			name: "fail - relationship finding relationship kind not declared as a relationship kind",
			args: args{
				graphExtension: GraphExtensionInput{
					ExtensionInput: baseExtensionInput(),
					NodeKindsInput: NodesInput{
						{
							Name: "AD_node_kind_1",
						},
						{
							Name: "AD_env_kind",
						},
					},
					RelationshipKindsInput: RelationshipsInput{
						{
							Name: "AD_edge kind 1",
						},
					},
					PropertiesInput: PropertiesInput{
						{
							Name: "property 1",
						},
						{
							Name: "property 2",
						},
					},
					EnvironmentsInput: EnvironmentsInput{
						{
							EnvironmentKindName: "AD_env_kind",
							SourceKindName:      "Base",
							PrincipalKinds:      []string{"AD_node_kind_1"},
						},
					},
					RelationshipFindingsInput: RelationshipFindingsInput{
						{
							Name:                 "AD_finding_1",
							EnvironmentKindName:  "AD_env_kind",
							RelationshipKindName: "AD_edge kind 2",
						},
					},
				},
			},
			wantErr: fmt.Errorf("graph schema relationship finding relationship kind %s not declared as a relationship kind", "AD_edge kind 2"),
		},
		{
			name: "fail - duplicate relationship findings",
			args: args{
				graphExtension: GraphExtensionInput{
					ExtensionInput: baseExtensionInput(),
					NodeKindsInput: NodesInput{
						{
							Name: "AD_node_kind_1",
						},
						{
							Name: "AD_env_kind_1",
						},
					},
					RelationshipKindsInput: RelationshipsInput{
						{
							Name: "AD_edge_kind_1",
						},
						{
							Name: "AD_edge_kind_2",
						},
					},
					PropertiesInput: PropertiesInput{
						{
							Name: "property 1",
						},
						{
							Name: "property 2",
						},
					},
					EnvironmentsInput: EnvironmentsInput{
						{
							EnvironmentKindName: "AD_env_kind_1",
							SourceKindName:      "Base",
							PrincipalKinds:      []string{"AD_node_kind_1"},
						},
					},
					RelationshipFindingsInput: RelationshipFindingsInput{
						{
							Name:                 "AD_finding_1",
							EnvironmentKindName:  "AD_env_kind_1",
							RelationshipKindName: "AD_edge_kind_1",
						},
						{
							Name:                 "AD_finding_1",
							EnvironmentKindName:  "AD_env_kind_1",
							RelationshipKindName: "AD_edge_kind_1",
						},
					},
				},
			},
			wantErr: fmt.Errorf("duplicate graph schema relationship finding: AD_finding_1"),
		},
		{
			name: "success - valid ExtensionInput",
			args: args{
				graphExtension: GraphExtensionInput{
					ExtensionInput: baseExtensionInput(),
					NodeKindsInput: NodesInput{{
						Name: "AD_node kind 1",
					}},
				},
			},
			wantErr: nil,
		},
		{
			name: "success - valid full ExtensionInput",
			args: args{
				graphExtension: GraphExtensionInput{
					ExtensionInput: baseExtensionInput(),
					NodeKindsInput: NodesInput{
						{
							Name:          "AD_node_kind_1",
							IconColor:     "#123456",
							IsDisplayKind: true,
							DisplayName:   "AD_node_kind_1",
							Icon:          "person",
							Info: KindInfoInputs{
								{InfoKey: "overview", Title: "Overview", Position: 1, Content: json.RawMessage(`{"markdown":{"content":"# Overview\n\nThis is **bold**."}}`)},
								{InfoKey: "security-notes", Title: "Security Notes", Position: 2, Content: json.RawMessage(`{"markdown":{"content":"- Note 1\n- Note 2"}}`)},
							},
						},
						{
							Name: "AD_env_kind_1",
						},
					},
					RelationshipKindsInput: RelationshipsInput{{
						Name: "AD_edge kind 1",
						Info: KindInfoInputs{
							{InfoKey: "details", Title: "Details", Position: 1, Content: json.RawMessage(`{"markdown":{"content":"Relationship details"}}`)},
						},
					}},
					EnvironmentsInput: EnvironmentsInput{
						{
							EnvironmentKindName: "AD_env_kind_1",
							SourceKindName:      "Base",
							PrincipalKinds: []string{
								"AD_node_kind_1",
							},
						},
					},
					RelationshipFindingsInput: RelationshipFindingsInput{
						{
							Name:                 "AD_finding_1",
							RelationshipKindName: "AD_edge kind 1",
							EnvironmentKindName:  "AD_env_kind_1",
						},
					},
				},
			},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.args.graphExtension.Validate(); tt.wantErr != nil {
				require.ErrorContains(t, err, tt.wantErr.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateMarkdownSafety(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		markdown string
		wantErr  error
	}{
		{name: "success_-_headings_and_formatting", markdown: "# Heading\n\n**bold** and *italic*", wantErr: nil},
		{name: "success_-_lists", markdown: "- List item 1\n- List item 2", wantErr: nil},
		{name: "success_-_links", markdown: "[Link](https://example.com)", wantErr: nil},
		{name: "success_-_code_blocks", markdown: "`code` and ```code block```", wantErr: nil},
		{name: "success_-_allowed_html", markdown: "Text with <b>bold</b> and <i>italic</i>", wantErr: nil},
		{name: "success_-_benign_angle_brackets", markdown: "Run: if a < b && c > d then exit", wantErr: nil},
		{name: "success_-_quotes_and_apostrophes", markdown: "He said \"hi\" and it's fine", wantErr: nil},
		{name: "success_-_mailto_link", markdown: "[email us](mailto:team@example.com)", wantErr: nil},
		{name: "error_-_script_tag", markdown: "<script>alert('xss')</script>", wantErr: ErrUnsafeMarkdownContent},
		{name: "error_-_img_onerror", markdown: "<img src=x onerror='alert(1)'>", wantErr: ErrUnsafeMarkdownContent},
		{name: "error_-_iframe", markdown: "<iframe src='evil.com'></iframe>", wantErr: ErrUnsafeMarkdownContent},
		{name: "error_-_javascript_href", markdown: "<a href='javascript:alert()'>Click</a>", wantErr: ErrUnsafeMarkdownContent},
		{name: "error_-_markdown_javascript_link", markdown: "[click me](javascript:alert(document.cookie))", wantErr: ErrUnsafeMarkdownContent},
		{name: "error_-_markdown_javascript_image", markdown: "![img](javascript:alert(1))", wantErr: ErrUnsafeMarkdownContent},
		{name: "error_-_markdown_vbscript_link", markdown: "[x](vbscript:msgbox(1))", wantErr: ErrUnsafeMarkdownContent},
		{name: "error_-_markdown_data_uri_link", markdown: "[x](data:text/html;base64,PHNjcmlwdD4=)", wantErr: ErrUnsafeMarkdownContent},
		{name: "error_-_markdown_reference_javascript_link", markdown: "[ref]: javascript:alert(1)", wantErr: ErrUnsafeMarkdownContent},

		// The following cases document real-world, benign content that the current naive
		// bluemonday sanitize-and-compare implementation wrongly rejects. The markdown is
		// sourced from AD/Azure entity-panel help text and repository .md files. They assert
		// the correct behavior (wantErr: nil) and therefore FAIL against the current
		// implementation, capturing the false positives to be addressed (BED-8764).
		{name: "false_positive_-_ntlm_hash_placeholder_in_prose", markdown: "Provide the <ntlm hash> value directly", wantErr: nil},
		{name: "false_positive_-_service_principal_app_id_placeholder", markdown: "Use the \"<service principal's app id>\" value", wantErr: nil},
		{name: "false_positive_-_generic_type_in_prose", markdown: "This returns a List<String> value", wantErr: nil},
		{name: "false_positive_-_superseded_rfc_identifier", markdown: "SUPERSEDED BY <RFC identifier> - deprecated in favor of another", wantErr: nil},
		{name: "false_positive_-_reference_link_with_amp_entity", markdown: "See [ATT&amp;CK T1098](https://attack.mitre.org/techniques/T1098/).", wantErr: nil},
		{name: "false_positive_-_placeholder_in_code_span", markdown: "See the folder `cmd/ui/tests/<suite>` for details", wantErr: nil},

		// The following cases pin down the exact boundary of what a markdown-aware
		// convert-then-sanitize pipeline (goldmark + bluemonday) would and would not fix,
		// so the ADR author can see the tradeoff concretely (BED-8764).
		//
		// An angle-bracket token inside a code span or fenced code block is unambiguously
		// literal text; goldmark escapes it (e.g. <code>&lt;blah&gt;</code>), so it becomes
		// safe and should be accepted. These currently FAIL (the bluemonday-only impl is not
		// markdown-aware and rejects the stripped tag) and would PASS once goldmark is added.
		{name: "boundary_-_angle_token_in_code_span_should_pass", markdown: "Provide the `<blah>` value", wantErr: nil},
		{name: "boundary_-_angle_token_in_fenced_code_should_pass", markdown: "```\n<blah>\n```", wantErr: nil},
		// A bare angle-bracket token in prose is, per the markdown/HTML spec, an actual HTML
		// tag and is indistinguishable from a real tag like <script>. Reject semantics MUST
		// reject it; goldmark does NOT change this. This case documents that goldmark is not
		// a fix for bare-prose tokens - authors must wrap them in backticks. Rejection here is
		// correct both before and after goldmark, so it PASSES against the current impl.
		{name: "boundary_-_bare_angle_token_in_prose_stays_rejected", markdown: "Provide the <blah> value", wantErr: ErrUnsafeMarkdownContent},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var (
				wrappedContent = struct {
					Markdown struct {
						Content string `json:"content"`
					} `json:"markdown"`
				}{}
			)
			wrappedContent.Markdown.Content = tc.markdown
			content, err := json.Marshal(wrappedContent)
			require.NoError(t, err)

			if err := ValidateMarkdownSafety(json.RawMessage(content)); tc.wantErr != nil {
				assert.ErrorIs(t, err, tc.wantErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestKindInfo_Validation(t *testing.T) {
	t.Parallel()

	var (
		validContent = json.RawMessage(`{"markdown":{"content":"Valid content"}}`)
	)

	testCases := []struct {
		name    string
		input   GraphExtensionInput
		wantErr error
	}{
		{
			name: "success_-_valid_info",
			input: GraphExtensionInput{
				ExtensionInput: baseExtensionInput(),
				NodeKindsInput: NodesInput{{
					Name: "AD_Node",
					Info: KindInfoInputs{
						{InfoKey: "overview", Title: "Overview", Position: 1, Content: validContent},
						{InfoKey: "details", Title: "Details", Position: 2, Content: validContent},
					},
				}},
			},
			wantErr: nil,
		},
		{
			name: "error_-_uppercase_key",
			input: GraphExtensionInput{
				ExtensionInput: baseExtensionInput(),
				NodeKindsInput: NodesInput{{
					Name: "AD_Node",
					Info: KindInfoInputs{{InfoKey: "INVALID_KEY", Title: "Title", Position: 1, Content: validContent}},
				}},
			},
			wantErr: ErrInvalidKindInfoKey,
		},
		{
			name: "error_-_key_with_spaces",
			input: GraphExtensionInput{
				ExtensionInput: baseExtensionInput(),
				NodeKindsInput: NodesInput{{
					Name: "AD_Node",
					Info: KindInfoInputs{{InfoKey: "invalid key", Title: "Title", Position: 1, Content: validContent}},
				}},
			},
			wantErr: ErrInvalidKindInfoKey,
		},
		{
			name: "error_-_key_with_special_chars",
			input: GraphExtensionInput{
				ExtensionInput: baseExtensionInput(),
				NodeKindsInput: NodesInput{{
					Name: "AD_Node",
					Info: KindInfoInputs{{InfoKey: "invalid@key", Title: "Title", Position: 1, Content: validContent}},
				}},
			},
			wantErr: ErrInvalidKindInfoKey,
		},
		{
			name: "error_-_key_too_long",
			input: GraphExtensionInput{
				ExtensionInput: baseExtensionInput(),
				NodeKindsInput: NodesInput{{
					Name: "AD_Node",
					Info: KindInfoInputs{{InfoKey: strings.Repeat("a", 129), Title: "Title", Position: 1, Content: validContent}},
				}},
			},
			wantErr: ErrInvalidKindInfoKey,
		},
		{
			name: "error_-_empty_key",
			input: GraphExtensionInput{
				ExtensionInput: baseExtensionInput(),
				NodeKindsInput: NodesInput{{
					Name: "AD_Node",
					Info: KindInfoInputs{{InfoKey: "", Title: "Title", Position: 1, Content: validContent}},
				}},
			},
			wantErr: ErrInvalidKindInfoKey,
		},
		{
			name: "error_-_too_many_entries",
			input: GraphExtensionInput{
				ExtensionInput: baseExtensionInput(),
				NodeKindsInput: NodesInput{{
					Name: "AD_Node",
					Info: func() KindInfoInputs {
						var entries KindInfoInputs
						for i := 0; i < 101; i++ {
							entries = append(entries, KindInfoInput{
								InfoKey:  fmt.Sprintf("key%d", i),
								Title:    "Title",
								Position: int32(i + 1),
								Content:  validContent,
							})
						}
						return entries
					}(),
				}},
			},
			wantErr: ErrTooManyKindInfoEntries,
		},
		{
			name: "error_-_invalid_position",
			input: GraphExtensionInput{
				ExtensionInput: baseExtensionInput(),
				NodeKindsInput: NodesInput{{
					Name: "AD_Node",
					Info: KindInfoInputs{{InfoKey: "test", Title: "Title", Position: 0, Content: validContent}},
				}},
			},
			wantErr: ErrInvalidKindInfoPosition,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if err := tc.input.Validate(); tc.wantErr != nil {
				require.ErrorIs(t, err, tc.wantErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
