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

package opengraphschema

import (
	"fmt"
	"testing"

	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/stretchr/testify/require"
)

func Test_validateGraphSchemaModel(t *testing.T) {
	type args struct {
		graphExtension model.GraphExtensionInput
	}
	tests := []struct {
		name    string
		args    args
		wantErr error
	}{
		{
			name: "fail - empty extension name",
			args: args{
				graphExtension: model.GraphExtensionInput{},
			},
			wantErr: fmt.Errorf("graph schema extension name is required"),
		},
		{
			name: "fail - empty extension version",
			args: args{
				graphExtension: model.GraphExtensionInput{
					ExtensionInput: model.ExtensionInput{
						Name: "Test extension",
					},
				},
			},
			wantErr: fmt.Errorf("graph schema extension version is required"),
		},
		{
			name: "fail - empty extension namespace",
			args: args{
				graphExtension: model.GraphExtensionInput{
					ExtensionInput: model.ExtensionInput{
						Name:    "Test extension",
						Version: "1.0.0",
					},
				},
			},
			wantErr: fmt.Errorf("graph schema extension namespace is required"),
		},
		{
			name: "fail - empty graph schema nodes",
			args: args{
				graphExtension: model.GraphExtensionInput{
					ExtensionInput: model.ExtensionInput{
						Name:      "Test extension",
						Version:   "1.0.0",
						Namespace: "AD",
					},
				},
			},
			wantErr: fmt.Errorf("graph schema node kinds are required"),
		},
		{
			name: "fail - duplicate kinds - two node kinds",
			args: args{
				graphExtension: model.GraphExtensionInput{
					ExtensionInput: model.ExtensionInput{
						Name:      "Test extension",
						Version:   "1.0.0",
						Namespace: "AD",
					},
					NodeKindsInput: model.NodesInput{
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
				graphExtension: model.GraphExtensionInput{
					ExtensionInput: model.ExtensionInput{
						Name:      "Test extension",
						Version:   "1.0.0",
						Namespace: "AD",
					},
					NodeKindsInput: model.NodesInput{
						{
							Name: "node kind 1",
						},
						{
							Name: "AD_node kind 2",
						},
					},
				},
			},
			wantErr: fmt.Errorf("graph schema kind node kind 1 is missing extension namespace prefix"),
		},
		{
			name: "fail - duplicate kinds - two edge kinds",
			args: args{
				graphExtension: model.GraphExtensionInput{
					ExtensionInput: model.ExtensionInput{
						Name:      "Test extension",
						Version:   "1.0.0",
						Namespace: "AD",
					},
					NodeKindsInput: model.NodesInput{
						{
							Name: "AD_node kind 1",
						},
						{
							Name: "AD_node kind 2",
						},
					},
					RelationshipKindsInput: model.RelationshipsInput{
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
				graphExtension: model.GraphExtensionInput{
					ExtensionInput: model.ExtensionInput{
						Name:      "Test extension",
						Version:   "1.0.0",
						Namespace: "AD",
					},
					NodeKindsInput: model.NodesInput{
						{
							Name: "AD_node kind 1",
						},
						{
							Name: "AD_node kind 2",
						},
					},
					RelationshipKindsInput: model.RelationshipsInput{
						{
							Name: "AD_edge kind 1",
						},
						{
							Name: "edge kind 1",
						},
					},
				},
			},
			wantErr: fmt.Errorf("graph schema edge kind edge kind 1 is missing extension namespace prefix"),
		},
		{
			name: "fail - duplicate properties",
			args: args{
				graphExtension: model.GraphExtensionInput{
					ExtensionInput: model.ExtensionInput{
						Name:      "Test extension",
						Version:   "1.0.0",
						Namespace: "AD",
					},
					NodeKindsInput: model.NodesInput{
						{
							Name: "AD_node kind 1",
						},
						{
							Name: "AD_node kind 2",
						},
					},
					RelationshipKindsInput: model.RelationshipsInput{
						{
							Name: "AD_edge kind 1",
						},
						{
							Name: "AD_edge kind 2",
						},
					},
					PropertiesInput: model.PropertiesInput{
						{
							Name: "property 1",
						},
						{
							Name: "property 1",
						},
					},
				},
			},
			wantErr: fmt.Errorf("duplicate graph properties: property 1"),
		},
		{
			name: "fail - duplicate kinds - same edge and node kind",
			args: args{
				graphExtension: model.GraphExtensionInput{
					ExtensionInput: model.ExtensionInput{
						Name:      "Test extension",
						Version:   "1.0.0",
						Namespace: "AD",
					},
					NodeKindsInput: model.NodesInput{
						{
							Name: "AD_a_duplicate_graph_kind",
						},
						{
							Name: "AD_node kind 2",
						},
					},
					RelationshipKindsInput: model.RelationshipsInput{
						{
							Name: "AD_edge kind 1",
						},
						{
							Name: "AD_a_duplicate_graph_kind",
						},
					},
					PropertiesInput: model.PropertiesInput{
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
				graphExtension: model.GraphExtensionInput{
					ExtensionInput: model.ExtensionInput{
						Name:      "Test extension",
						Version:   "1.0.0",
						Namespace: "AD",
					},
					NodeKindsInput: model.NodesInput{
						{
							Name: "AD_node_kind_1",
						},
						{
							Name: "AD_node kind 2",
						},
					},
					RelationshipKindsInput: model.RelationshipsInput{
						{
							Name: "AD_edge kind 1",
						},
					},
					PropertiesInput: model.PropertiesInput{
						{
							Name: "property 1",
						},
						{
							Name: "property 2",
						},
					},
					EnvironmentsInput: model.EnvironmentsInput{
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
			name: "fail - environment kind not declared as a node kind",
			args: args{
				graphExtension: model.GraphExtensionInput{
					ExtensionInput: model.ExtensionInput{
						Name:      "Test extension",
						Version:   "1.0.0",
						Namespace: "AD",
					},
					NodeKindsInput: model.NodesInput{
						{
							Name: "AD_node_kind_1",
						},
						{
							Name: "AD_node kind 2",
						},
					},
					RelationshipKindsInput: model.RelationshipsInput{
						{
							Name: "AD_edge kind 1",
						},
					},
					PropertiesInput: model.PropertiesInput{
						{
							Name: "property 1",
						},
						{
							Name: "property 2",
						},
					},
					EnvironmentsInput: model.EnvironmentsInput{
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
			name: "fail - environment source kind is empty",
			args: args{
				graphExtension: model.GraphExtensionInput{
					ExtensionInput: model.ExtensionInput{
						Name:      "Test extension",
						Version:   "1.0.0",
						Namespace: "AD",
					},
					NodeKindsInput: model.NodesInput{
						{
							Name: "AD_node_kind_1",
						},
						{
							Name: "AD_env_kind",
						},
					},
					RelationshipKindsInput: model.RelationshipsInput{
						{
							Name: "AD_edge kind 1",
						},
					},
					PropertiesInput: model.PropertiesInput{
						{
							Name: "property 1",
						},
						{
							Name: "property 2",
						},
					},
					EnvironmentsInput: model.EnvironmentsInput{
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
			name: "fail - environment source kind cannot be declared as a node or relationship kind",
			args: args{
				graphExtension: model.GraphExtensionInput{
					ExtensionInput: model.ExtensionInput{
						Name:      "Test extension",
						Version:   "1.0.0",
						Namespace: "AD",
					},
					NodeKindsInput: model.NodesInput{
						{
							Name: "AD_node_kind_1",
						},
						{
							Name: "AD_env_kind",
						},
					},
					RelationshipKindsInput: model.RelationshipsInput{
						{
							Name: "AD_edge kind 1",
						},
					},
					PropertiesInput: model.PropertiesInput{
						{
							Name: "property 1",
						},
						{
							Name: "property 2",
						},
					},
					EnvironmentsInput: model.EnvironmentsInput{
						{
							EnvironmentKindName: "AD_env_kind",
							SourceKindName:      "AD_node_kind_1",
							PrincipalKinds:      []string{"AD_node_kind_1"},
						},
					},
				},
			},
			wantErr: fmt.Errorf("graph schema environment source kind %s should not be declared as a node kind", "AD_node_kind_1"),
		},
		{
			name: "fail - environment principal kind missing namespace prefix",
			args: args{
				graphExtension: model.GraphExtensionInput{
					ExtensionInput: model.ExtensionInput{
						Name:      "Test extension",
						Version:   "1.0.0",
						Namespace: "AD",
					},
					NodeKindsInput: model.NodesInput{
						{
							Name: "AD_node_kind_1",
						},
						{
							Name: "AD_env_kind",
						},
					},
					RelationshipKindsInput: model.RelationshipsInput{
						{
							Name: "AD_edge kind 1",
						},
					},
					PropertiesInput: model.PropertiesInput{
						{
							Name: "property 1",
						},
						{
							Name: "property 2",
						},
					},
					EnvironmentsInput: model.EnvironmentsInput{
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
			name: "fail - environment principal kind not declared as a node kind",
			args: args{
				graphExtension: model.GraphExtensionInput{
					ExtensionInput: model.ExtensionInput{
						Name:      "Test extension",
						Version:   "1.0.0",
						Namespace: "AD",
					},
					NodeKindsInput: model.NodesInput{
						{
							Name: "AD_node_kind_1",
						},
						{
							Name: "AD_env_kind",
						},
					},
					RelationshipKindsInput: model.RelationshipsInput{
						{
							Name: "AD_edge kind 1",
						},
					},
					PropertiesInput: model.PropertiesInput{
						{
							Name: "property 1",
						},
						{
							Name: "property 2",
						},
					},
					EnvironmentsInput: model.EnvironmentsInput{
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
				graphExtension: model.GraphExtensionInput{
					ExtensionInput: model.ExtensionInput{
						Name:      "Test extension",
						Version:   "1.0.0",
						Namespace: "AD",
					},
					NodeKindsInput: model.NodesInput{
						{
							Name: "AD_node_kind_1",
						},
						{
							Name: "AD_env_kind",
						},
					},
					RelationshipKindsInput: model.RelationshipsInput{
						{
							Name: "AD_edge kind 1",
						},
					},
					PropertiesInput: model.PropertiesInput{
						{
							Name: "property 1",
						},
						{
							Name: "property 2",
						},
					},
					EnvironmentsInput: model.EnvironmentsInput{
						{
							EnvironmentKindName: "AD_env_kind",
							SourceKindName:      "Base",
							PrincipalKinds:      []string{"AD_node_kind_1"},
						},
					},
					RelationshipFindingsInput: model.RelationshipFindingsInput{
						{
							Name: "finding_1",
						},
					},
				},
			},
			wantErr: fmt.Errorf("graph schema relationship finding %s is missing extension namespace prefix", "finding_1"),
		},
		{
			name: "fail - relationship finding environment kind name missing namespace prefix",
			args: args{
				graphExtension: model.GraphExtensionInput{
					ExtensionInput: model.ExtensionInput{
						Name:      "Test extension",
						Version:   "1.0.0",
						Namespace: "AD",
					},
					NodeKindsInput: model.NodesInput{
						{
							Name: "AD_node_kind_1",
						},
						{
							Name: "AD_env_kind",
						},
					},
					RelationshipKindsInput: model.RelationshipsInput{
						{
							Name: "AD_edge kind 1",
						},
					},
					PropertiesInput: model.PropertiesInput{
						{
							Name: "property 1",
						},
						{
							Name: "property 2",
						},
					},
					EnvironmentsInput: model.EnvironmentsInput{
						{
							EnvironmentKindName: "AD_env_kind",
							SourceKindName:      "Base",
							PrincipalKinds:      []string{"AD_node_kind_1"},
						},
					},
					RelationshipFindingsInput: model.RelationshipFindingsInput{
						{
							Name:                "AD_finding_1",
							EnvironmentKindName: "env_kind",
						},
					},
				},
			},
			wantErr: fmt.Errorf("graph schema relationship finding environment kind %s is missing extension namespace prefix", "env_kind"),
		},
		{
			name: "fail - relationship finding relationship kind name missing namespace prefix",
			args: args{
				graphExtension: model.GraphExtensionInput{
					ExtensionInput: model.ExtensionInput{
						Name:      "Test extension",
						Version:   "1.0.0",
						Namespace: "AD",
					},
					NodeKindsInput: model.NodesInput{
						{
							Name: "AD_node_kind_1",
						},
						{
							Name: "AD_env_kind",
						},
					},
					RelationshipKindsInput: model.RelationshipsInput{
						{
							Name: "AD_edge kind 1",
						},
					},
					PropertiesInput: model.PropertiesInput{
						{
							Name: "property 1",
						},
						{
							Name: "property 2",
						},
					},
					EnvironmentsInput: model.EnvironmentsInput{
						{
							EnvironmentKindName: "AD_env_kind",
							SourceKindName:      "Base",
							PrincipalKinds:      []string{"AD_node_kind_1"},
						},
					},
					RelationshipFindingsInput: model.RelationshipFindingsInput{
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
			name: "fail - relationship finding environment kind not declared as a node kind",
			args: args{
				graphExtension: model.GraphExtensionInput{
					ExtensionInput: model.ExtensionInput{
						Name:      "Test extension",
						Version:   "1.0.0",
						Namespace: "AD",
					},
					NodeKindsInput: model.NodesInput{
						{
							Name: "AD_node_kind_1",
						},
						{
							Name: "AD_env_kind",
						},
					},
					RelationshipKindsInput: model.RelationshipsInput{
						{
							Name: "AD_edge kind 1",
						},
					},
					PropertiesInput: model.PropertiesInput{
						{
							Name: "property 1",
						},
						{
							Name: "property 2",
						},
					},
					EnvironmentsInput: model.EnvironmentsInput{
						{
							EnvironmentKindName: "AD_env_kind",
							SourceKindName:      "Base",
							PrincipalKinds:      []string{"AD_node_kind_1"},
						},
					},
					RelationshipFindingsInput: model.RelationshipFindingsInput{
						{
							Name:                 "AD_finding_1",
							EnvironmentKindName:  "AD_env_kind_MISSING",
							RelationshipKindName: "AD_edge kind 1",
						},
					},
				},
			},
			wantErr: fmt.Errorf("graph schema relationship finding environment kind %s not declared as a node kind", "AD_env_kind_MISSING"),
		},
		{
			name: "fail - relationship finding relationship kind not declared as a relationship kind",
			args: args{
				graphExtension: model.GraphExtensionInput{
					ExtensionInput: model.ExtensionInput{
						Name:      "Test extension",
						Version:   "1.0.0",
						Namespace: "AD",
					},
					NodeKindsInput: model.NodesInput{
						{
							Name: "AD_node_kind_1",
						},
						{
							Name: "AD_env_kind",
						},
					},
					RelationshipKindsInput: model.RelationshipsInput{
						{
							Name: "AD_edge kind 1",
						},
					},
					PropertiesInput: model.PropertiesInput{
						{
							Name: "property 1",
						},
						{
							Name: "property 2",
						},
					},
					EnvironmentsInput: model.EnvironmentsInput{
						{
							EnvironmentKindName: "AD_env_kind",
							SourceKindName:      "Base",
							PrincipalKinds:      []string{"AD_node_kind_1"},
						},
					},
					RelationshipFindingsInput: model.RelationshipFindingsInput{
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
			name: "fail - relationship finding source kind cannot be empty",
			args: args{
				graphExtension: model.GraphExtensionInput{
					ExtensionInput: model.ExtensionInput{
						Name:      "Test extension",
						Version:   "1.0.0",
						Namespace: "AD",
					},
					NodeKindsInput: model.NodesInput{
						{
							Name: "AD_node_kind_1",
						},
						{
							Name: "AD_env_kind",
						},
					},
					RelationshipKindsInput: model.RelationshipsInput{
						{
							Name: "AD_edge kind 1",
						},
					},
					PropertiesInput: model.PropertiesInput{
						{
							Name: "property 1",
						},
						{
							Name: "property 2",
						},
					},
					EnvironmentsInput: model.EnvironmentsInput{
						{
							EnvironmentKindName: "AD_env_kind",
							SourceKindName:      "Base",
							PrincipalKinds:      []string{"AD_node_kind_1"},
						},
					},
					RelationshipFindingsInput: model.RelationshipFindingsInput{
						{
							Name:                 "AD_finding_1",
							EnvironmentKindName:  "AD_env_kind",
							RelationshipKindName: "AD_edge kind 1",
						},
					},
				},
			},
			wantErr: fmt.Errorf("graph schema relationship finding source kind cannot be empty"),
		},
		{
			name: "fail - relationship finding source kind cannot be declared as a node or relationship kind",
			args: args{
				graphExtension: model.GraphExtensionInput{
					ExtensionInput: model.ExtensionInput{
						Name:      "Test extension",
						Version:   "1.0.0",
						Namespace: "AD",
					},
					NodeKindsInput: model.NodesInput{
						{
							Name: "AD_node_kind_1",
						},
						{
							Name: "AD_env_kind",
						},
					},
					RelationshipKindsInput: model.RelationshipsInput{
						{
							Name: "AD_edge kind 1",
						},
					},
					PropertiesInput: model.PropertiesInput{
						{
							Name: "property 1",
						},
						{
							Name: "property 2",
						},
					},
					EnvironmentsInput: model.EnvironmentsInput{
						{
							EnvironmentKindName: "AD_env_kind",
							SourceKindName:      "Base",
							PrincipalKinds:      []string{"AD_node_kind_1"},
						},
					},
					RelationshipFindingsInput: model.RelationshipFindingsInput{
						{
							Name:                 "AD_finding_1",
							EnvironmentKindName:  "AD_env_kind",
							RelationshipKindName: "AD_edge kind 1",
							SourceKindName:       "AD_node_kind_1",
						},
					},
				},
			},
			wantErr: fmt.Errorf("graph schema relationship finding source kind %s should not be declared as a node kind", "AD_node_kind_1"),
		},
		{
			name: "success - valid ExtensionInput",
			args: args{
				graphExtension: model.GraphExtensionInput{
					ExtensionInput: model.ExtensionInput{
						Name:      "Test extension",
						Version:   "1.0.0",
						Namespace: "AD",
					},
					NodeKindsInput: model.NodesInput{{
						Name: "AD_node kind 1",
					}},
				},
			},
			wantErr: nil,
		},
		{
			name: "success - valid full ExtensionInput",
			args: args{
				graphExtension: model.GraphExtensionInput{
					ExtensionInput: model.ExtensionInput{
						Name:      "Test extension",
						Version:   "1.0.0",
						Namespace: "AD",
					},
					NodeKindsInput: model.NodesInput{
						{
							Name: "AD_node_kind_1",
						},
						{
							Name: "AD_env_kind_1",
						},
					},
					RelationshipKindsInput: model.RelationshipsInput{{
						Name: "AD_edge kind 1",
					}},
					EnvironmentsInput: model.EnvironmentsInput{
						{
							EnvironmentKindName: "AD_env_kind_1",
							SourceKindName:      "Base",
							PrincipalKinds: []string{
								"AD_node_kind_1",
							},
						},
					},
					RelationshipFindingsInput: model.RelationshipFindingsInput{
						{
							Name:                 "AD_finding_1",
							SourceKindName:       "Base",
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
			if err := validateGraphExtension(tt.args.graphExtension); tt.wantErr != nil {
				require.ErrorContains(t, err, tt.wantErr.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}
