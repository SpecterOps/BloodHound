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
package v2

import (
	"testing"

	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/stretchr/testify/assert"
)

func Test_convertGraphExtensionPayloadToGraphExtension(t *testing.T) {
	type args struct {
		payload GraphExtensionPayload
	}
	tests := []struct {
		name string
		args args
		want model.GraphExtensionInput
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
					/*GraphSchemaProperties: []GraphSchemaPropertiesPayload{
						{
							Name:        "Property_1",
							DisplayName: "Property 1",
							DataType:    "string",
							Description: "a property",
						},
					},*/
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
							SourceKind:       "Source_Kind_1",
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
			want: model.GraphExtensionInput{
				ExtensionInput: model.ExtensionInput{
					Name:        "Test_Extension",
					DisplayName: "Test Extension",
					Version:     "1.0.0",
					Namespace:   "TEST",
				},
				PropertiesInput: make(model.PropertiesInput, 0),
				/*PropertiesInput: model.PropertiesInput{
					{
						Name:        "Property_1",
						DisplayName: "Property 1",
						DataType:    "string",
						Description: "a property",
					},
				},*/
				RelationshipKindsInput: model.RelationshipsInput{
					{
						Name:          "GraphSchemaEdgeKind_1",
						Description:   "GraphSchemaRelationshipKind 1",
						IsTraversable: true,
					},
				},
				NodeKindsInput: model.NodesInput{
					{
						Name:          "GraphSchemaNodeKind_1",
						DisplayName:   "GraphSchemaNodeKind 1",
						Description:   "a graph schema node",
						IsDisplayKind: true,
						Icon:          "User",
						IconColor:     "blue",
					},
				},
				EnvironmentsInput: []model.EnvironmentInput{
					{
						EnvironmentKindName: "EnvironmentInput",
						SourceKindName:      "Source_Kind_1",
						PrincipalKinds:      []string{"User"},
					},
				},
				RelationshipFindingsInput: []model.RelationshipFindingInput{
					{
						Name:                 "Finding_1",
						DisplayName:          "Finding 1",
						SourceKindName:       "Source_Kind_1",
						RelationshipKindName: "GraphSchemaEdgeKind_1",
						EnvironmentKindName:  "EnvironmentInput",
						RemediationInput: model.RemediationInput{
							ShortDescription: "remediation for Finding_1",
							LongDescription:  "a remediation for Finding 1",
							ShortRemediation: "do x",
							LongRemediation:  "do x but better",
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, convertGraphExtensionPayloadToGraphExtension(tt.args.payload), "ConvertGraphExtensionPayloadToGraphExtension(%v)", tt.args.payload)
		})
	}
}
