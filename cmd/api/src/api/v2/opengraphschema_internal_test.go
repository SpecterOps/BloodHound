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
	"math"
	"testing"

	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/stretchr/testify/assert"
)

func Test_convertGraphExtensionPayloadToGraphExtension(t *testing.T) {
	t.Parallel()

	type args struct {
		payload GraphExtensionPayload
	}
	tests := []struct {
		name    string
		args    args
		want    model.GraphExtensionInput
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
						Info:          make(model.KindInfoInputs, 0),
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
						Info:          make(model.KindInfoInputs, 0),
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

			got, err := convertGraphExtensionPayloadToGraphExtension(tt.args.payload)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equalf(t, tt.want, got, "ConvertGraphExtensionPayloadToGraphExtension(%v)", tt.args.payload)
			}
		})
	}
}

func Test_parseInfoPayload(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   map[string]KindInfoPayload
		want    model.KindInfoInputs
		wantErr bool
	}{
		{
			name:  "empty_info",
			input: nil,
			want:  make(model.KindInfoInputs, 0),
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
			want: model.KindInfoInputs{
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
					Position: 1,
					Markdown: []byte(`{"content":"# Overview"}`),
				},
				"details": {
					Title:    "Details",
					Position: 2,
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
			want: model.KindInfoInputs{
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
			want: model.KindInfoInputs{
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
			want: model.KindInfoInputs{
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
