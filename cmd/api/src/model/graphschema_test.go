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
	"math"
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
								{InfoKey: "overview", Title: "Overview", Position: 0, Content: json.RawMessage(`{"markdown":{"content":"# Overview\n\nThis is **bold**."}}`)},
								{InfoKey: "security-notes", Title: "Security Notes", Position: 1, Content: json.RawMessage(`{"markdown":{"content":"- Note 1\n- Note 2"}}`)},
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
						{InfoKey: "overview", Title: "Overview", Position: 0, Content: validContent},
						{InfoKey: "details", Title: "Details", Position: 1, Content: validContent},
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
			name: "error_-_empty_title",
			input: GraphExtensionInput{
				ExtensionInput: baseExtensionInput(),
				NodeKindsInput: NodesInput{{
					Name: "AD_Node",
					Info: KindInfoInputs{{InfoKey: "overview", Title: "", Position: 1, Content: validContent}},
				}},
			},
			wantErr: ErrInvalidKindInfoTitle,
		},
		{
			name: "error_-_whitespace_title",
			input: GraphExtensionInput{
				ExtensionInput: baseExtensionInput(),
				NodeKindsInput: NodesInput{{
					Name: "AD_Node",
					Info: KindInfoInputs{{InfoKey: "overview", Title: "   ", Position: 1, Content: validContent}},
				}},
			},
			wantErr: ErrInvalidKindInfoTitle,
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
					Info: KindInfoInputs{{InfoKey: "test", Title: "Title", Position: -1, Content: validContent}},
				}},
			},
			wantErr: ErrInvalidKindInfoPosition,
		},
		{
			name: "success_-_zero_position",
			input: GraphExtensionInput{
				ExtensionInput: baseExtensionInput(),
				NodeKindsInput: NodesInput{{
					Name: "AD_Node",
					Info: KindInfoInputs{{InfoKey: "test", Title: "Title", Position: 0, Content: validContent}},
				}},
			},
			wantErr: nil,
		},
		{
			name: "error_-_content_empty",
			input: GraphExtensionInput{
				ExtensionInput: baseExtensionInput(),
				NodeKindsInput: NodesInput{{
					Name: "AD_Node",
					Info: KindInfoInputs{{InfoKey: "test", Title: "Title", Position: 1, Content: json.RawMessage(``)}},
				}},
			},
			wantErr: ErrInvalidKindInfoContent,
		},
		{
			name: "error_-_content_missing_markdown_key",
			input: GraphExtensionInput{
				ExtensionInput: baseExtensionInput(),
				NodeKindsInput: NodesInput{{
					Name: "AD_Node",
					Info: KindInfoInputs{{InfoKey: "test", Title: "Title", Position: 1, Content: json.RawMessage(`{"foo":"bar"}`)}},
				}},
			},
			wantErr: ErrInvalidKindInfoContent,
		},
		{
			name: "error_-_content_missing_content_string",
			input: GraphExtensionInput{
				ExtensionInput: baseExtensionInput(),
				NodeKindsInput: NodesInput{{
					Name: "AD_Node",
					Info: KindInfoInputs{{InfoKey: "test", Title: "Title", Position: 1, Content: json.RawMessage(`{"markdown":{}}`)}},
				}},
			},
			wantErr: ErrInvalidKindInfoContent,
		},
		{
			name: "success_-_blank_content_string",
			input: GraphExtensionInput{
				ExtensionInput: baseExtensionInput(),
				NodeKindsInput: NodesInput{{
					Name: "AD_Node",
					Info: KindInfoInputs{{InfoKey: "test", Title: "Title", Position: 1, Content: json.RawMessage(`{"markdown":{"content":"   "}}`)}},
				}},
			},
			wantErr: nil,
		},
		{
			name: "success_-_empty_content_string",
			input: GraphExtensionInput{
				ExtensionInput: baseExtensionInput(),
				NodeKindsInput: NodesInput{{
					Name: "AD_Node",
					Info: KindInfoInputs{{InfoKey: "test", Title: "Title", Position: 1, Content: json.RawMessage(`{"markdown":{"content":""}}`)}},
				}},
			},
			wantErr: nil,
		},
		{
			name: "error_-_content_unknown_field",
			input: GraphExtensionInput{
				ExtensionInput: baseExtensionInput(),
				NodeKindsInput: NodesInput{{
					Name: "AD_Node",
					Info: KindInfoInputs{{InfoKey: "test", Title: "Title", Position: 1, Content: json.RawMessage(`{"markdown":{"content":"ok"},"extra":"x"}`)}},
				}},
			},
			wantErr: ErrInvalidKindInfoContent,
		},
		{
			name: "error_-_content_not_an_object",
			input: GraphExtensionInput{
				ExtensionInput: baseExtensionInput(),
				NodeKindsInput: NodesInput{{
					Name: "AD_Node",
					Info: KindInfoInputs{{InfoKey: "test", Title: "Title", Position: 1, Content: json.RawMessage(`"just a string"`)}},
				}},
			},
			wantErr: ErrInvalidKindInfoContent,
		},
		{
			name: "error_-_duplicate_info_key",
			input: GraphExtensionInput{
				ExtensionInput: baseExtensionInput(),
				NodeKindsInput: NodesInput{{
					Name: "AD_Node",
					Info: KindInfoInputs{
						{InfoKey: "overview", Title: "Overview", Position: 0, Content: validContent},
						{InfoKey: "overview", Title: "Details", Position: 1, Content: validContent},
					},
				}},
			},
			wantErr: ErrKindInfoDuplicateInfoKey,
		},
		{
			name: "error_-_duplicate_position",
			input: GraphExtensionInput{
				ExtensionInput: baseExtensionInput(),
				NodeKindsInput: NodesInput{{
					Name: "AD_Node",
					Info: KindInfoInputs{
						{InfoKey: "overview", Title: "Overview", Position: 0, Content: validContent},
						{InfoKey: "details", Title: "Details", Position: 0, Content: validContent},
					},
				}},
			},
			wantErr: ErrKindInfoDuplicatePosition,
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

func Test_GraphExtensionPayload_ToGraphExtensionInput(t *testing.T) {
	t.Parallel()

	type args struct {
		payload GraphExtensionPayload
	}
	tests := []struct {
		name    string
		args    args
		want    GraphExtensionInput
		wantErr bool
	}{
		{
			name: "success",
			args: args{
				payload: GraphExtensionPayload{
					GraphSchemaExtension: GraphSchemaExtensionPayload{
						Name:        "Test_Extension",
						DisplayName: "Test Extension",
						Version:     "1.0.0",
						Namespace:   "TEST",
					},
					GraphSchemaRelationshipKinds: []GraphSchemaRelationshipKindsPayload{
						{
							Name:          "GraphSchemaEdgeKind_1",
							Description:   "GraphSchemaRelationshipKind 1",
							IsTraversable: true,
						},
					},
					GraphSchemaNodeKinds: []GraphSchemaNodeKindsPayload{
						{
							Name:          "GraphSchemaNodeKind_1",
							DisplayName:   "GraphSchemaNodeKind 1",
							Description:   "a graph schema node",
							IsDisplayKind: true,
							Icon:          "User",
							IconColor:     "blue",
						},
					},
					GraphEnvironments: []EnvironmentPayload{
						{
							EnvironmentKind: "EnvironmentInput",
							SourceKind:      "Source_Kind_1",
							PrincipalKinds:  []string{"User"},
						},
					},
					GraphRelationshipFindings: []RelationshipFindingsPayload{
						{
							Name:             "Finding_1",
							DisplayName:      "Finding 1",
							RelationshipKind: "GraphSchemaEdgeKind_1",
							EnvironmentKind:  "EnvironmentInput",
							Remediation: RemediationPayload{
								ShortDescription: "remediation for Finding_1",
								LongDescription:  "a remediation for Finding 1",
								ShortRemediation: "do x",
								LongRemediation:  "do x but better",
							},
						},
					},
				},
			},
			want: GraphExtensionInput{
				ExtensionInput: ExtensionInput{
					Name:        "Test_Extension",
					DisplayName: "Test Extension",
					Version:     "1.0.0",
					Namespace:   "TEST",
				},
				RelationshipKindsInput: RelationshipsInput{
					{
						Name:          "GraphSchemaEdgeKind_1",
						Description:   "GraphSchemaRelationshipKind 1",
						IsTraversable: true,
						Info:          make(KindInfoInputs, 0),
					},
				},
				NodeKindsInput: NodesInput{
					{
						Name:          "GraphSchemaNodeKind_1",
						DisplayName:   "GraphSchemaNodeKind 1",
						Description:   "a graph schema node",
						IsDisplayKind: true,
						Icon:          "User",
						IconColor:     "blue",
						Info:          make(KindInfoInputs, 0),
					},
				},
				EnvironmentsInput: []EnvironmentInput{
					{
						EnvironmentKindName: "EnvironmentInput",
						SourceKindName:      "Source_Kind_1",
						PrincipalKinds:      []string{"User"},
					},
				},
				RelationshipFindingsInput: []RelationshipFindingInput{
					{
						Name:                 "Finding_1",
						DisplayName:          "Finding 1",
						RelationshipKindName: "GraphSchemaEdgeKind_1",
						EnvironmentKindName:  "EnvironmentInput",
						RemediationInput: RemediationInput{
							ShortDescription: "remediation for Finding_1",
							LongDescription:  "a remediation for Finding 1",
							ShortRemediation: "do x",
							LongRemediation:  "do x but better",
						},
					},
				},
			},
		},
		{
			name:    "error_-_invalid_node_info_markdown",
			wantErr: true,
			args: args{
				payload: GraphExtensionPayload{
					GraphSchemaNodeKinds: []GraphSchemaNodeKindsPayload{
						{
							Name: "Bad_Node",
							Info: map[string]KindInfoPayload{
								"bad": {Title: "Bad", Position: 1, Markdown: []byte(`{invalid json`)},
							},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := tt.args.payload.ToGraphExtensionInput()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equalf(t, tt.want, got, "ToGraphExtensionInput(%v)", tt.args.payload)
			}
		})
	}
}

func Test_parseInfoPayload(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   map[string]KindInfoPayload
		want    KindInfoInputs
		wantErr bool
	}{
		{
			name:  "empty_info",
			input: nil,
			want:  make(KindInfoInputs, 0),
		},
		{
			name: "valid_info",
			input: map[string]KindInfoPayload{
				"overview": {
					Title:    "Overview",
					Position: 1,
					Markdown: []byte(`{"content":"# Test"}`),
				},
			},
			want: KindInfoInputs{
				{
					InfoKey:  "overview",
					Title:    "Overview",
					Position: 1,
					Content:  []byte(`{"markdown":{"content":"# Test"}}`),
				},
			},
			wantErr: false,
		},
		{
			name: "multiple_info_entries",
			input: map[string]KindInfoPayload{
				"overview": {
					Title:    "Overview",
					Position: 0,
					Markdown: []byte(`{"content":"# Overview"}`),
				},
				"details": {
					Title:    "Details",
					Position: 1,
					Markdown: []byte(`{"content":"## Details"}`),
				},
			},
			wantErr: false,
		},
		{
			name: "blank_markdown_content",
			input: map[string]KindInfoPayload{
				"empty": {
					Title:    "Empty",
					Position: 1,
					Markdown: []byte(`{"content":""}`),
				},
			},
			want: KindInfoInputs{
				{
					InfoKey:  "empty",
					Title:    "Empty",
					Position: 1,
					Content:  []byte(`{"markdown":{"content":""}}`),
				},
			},
			wantErr: false,
		},
		{
			name: "empty_json_rawmessage_default",
			input: map[string]KindInfoPayload{
				"test": {
					Title:    "Test",
					Position: 1,
					Markdown: []byte{}, // Empty markdown gets default structure
				},
			},
			want: KindInfoInputs{
				{
					InfoKey:  "test",
					Title:    "Test",
					Position: 1,
					Content:  []byte(`{"markdown":{"content":""}}`), // Empty content
				},
			},
			wantErr: false,
		},
		{
			name: "nil_json_rawmessage_default",
			input: map[string]KindInfoPayload{
				"test": {
					Title:    "Test",
					Position: 1,
					Markdown: nil, // Nil markdown also gets default structure
				},
			},
			want: KindInfoInputs{
				{
					InfoKey:  "test",
					Title:    "Test",
					Position: 1,
					Content:  []byte(`{"markdown":{"content":""}}`), // Empty content
				},
			},
			wantErr: false,
		},
		{
			name: "error_-_invalid_markdown_json",
			input: map[string]KindInfoPayload{
				"bad": {
					Title:    "Bad",
					Position: 1,
					Markdown: []byte(`{invalid json`), // malformed JSON fails to wrap
				},
			},
			wantErr: true,
		},
		{
			name: "error_-_position_exceeds_int32",
			input: map[string]KindInfoPayload{
				"overflow": {
					Title:    "Overflow",
					Position: math.MaxInt32 + 1, // would silently wrap when cast to int32
					Markdown: []byte(`{"content":"# Test"}`),
				},
			},
			wantErr: true,
		},
		{
			name: "error_-_position_below_int32",
			input: map[string]KindInfoPayload{
				"underflow": {
					Title:    "Underflow",
					Position: math.MinInt32 - 1, // would silently wrap when cast to int32
					Markdown: []byte(`{"content":"# Test"}`),
				},
			},
			wantErr: true,
		},
		{
			name: "zero_position",
			input: map[string]KindInfoPayload{
				"zero": {
					Title:    "Zero",
					Position: 0, // within int32 range; passes the overflow guard
					Markdown: []byte(`{"content":"# Test"}`),
				},
			},
			want: KindInfoInputs{
				{
					InfoKey:  "zero",
					Title:    "Zero",
					Position: 0,
					Content:  []byte(`{"markdown":{"content":"# Test"}}`),
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := parseInfoPayload(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.want != nil {
					assert.Equal(t, tt.want, got)
				} else {
					assert.NotNil(t, got)
				}
			}
		})
	}
}
