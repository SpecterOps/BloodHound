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

package ingest_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/specterops/bloodhound/src/model/ingest"
	ingest_service "github.com/specterops/bloodhound/src/services/ingest"
	"github.com/stretchr/testify/assert"
)

type metaTagAssertion struct {
	name         string
	rawString    string
	err          error
	expectedType ingest.DataType
}

func Test_ValidateMetaTag(t *testing.T) {
	assertions := []metaTagAssertion{
		{
			name:         "valid",
			rawString:    `{"meta": {"methods": 0, "type": "sessions", "count": 0, "version": 5}, "data": []}`,
			err:          nil,
			expectedType: ingest.DataTypeSession,
		},
		{
			name:      "No data tag",
			rawString: `{"meta": {"methods": 0, "type": "sessions", "count": 0, "version": 5}}`,
			err:       ingest.ErrDataTagNotFound,
		},
		{
			name:      "No meta tag",
			rawString: `{"data": []}`,
			err:       ingest.ErrMetaTagNotFound,
		},
		{
			name:      "No valid tags",
			rawString: `{}`,
			err:       ingest.ErrNoTagFound,
		},
		{
			name:         "ignore invalid tag but still find correct tag",
			rawString:    `{"meta": 0, "meta": {"methods": 0, "type": "sessions", "count": 0, "version": 5}, "data": []}`,
			err:          nil,
			expectedType: ingest.DataTypeSession,
		},
		{
			name:         "swapped order",
			rawString:    `{"data": [],"meta": {"methods": 0, "type": "sessions", "count": 0, "version": 5}}`,
			err:          nil,
			expectedType: ingest.DataTypeSession,
		},
		{
			name:      "invalid type",
			rawString: `{"data": [],"meta": {"methods": 0, "type": "invalid", "count": 0, "version": 5}}`,
			err:       ingest.ErrMetaTagNotFound,
		},
	}

	for _, assertion := range assertions {
		meta, err := ingest_service.ValidateMetaTag(strings.NewReader(assertion.rawString), false)
		assert.ErrorIs(t, err, assertion.err)
		if assertion.err == nil {
			assert.Equal(t, meta.Type, assertion.expectedType)
		}
	}
}

type genericAssertion struct {
	name       string
	err        error
	errMsgs    []string // one record may have multiple violations
	payload    *testPayload
	rawPayload string // for cases that require raw JSON to trip validation controls
}

type testNode struct {
	ID         string         `json:"id,omitempty"`
	Properties map[string]any `json:"properties"`
	Kinds      []string       `json:"kinds"`
}

type testEdge struct {
	// non-pointer nested structs are automatically initialized with the zero value of the struct type.
	// we want the edgePiece's to be pointers so that we can omit them in the request and test validation
	Start      *edgePiece     `json:"start"`
	End        *edgePiece     `json:"end"`
	Kind       string         `json:"kind,omitempty"`
	Properties map[string]any `json:"properties"`
}

type edgePiece struct {
	IDValue    string `json:"id_value,omitempty"`
	IDProperty string `json:"id_property,omitempty"`
}

type testGraph struct {
	Nodes []testNode `json:"nodes,omitempty"`
	Edges []testEdge `json:"edges,omitempty"`
}

type testPayload struct {
	Graph testGraph `json:"graph"`
}

// TODO: Error aggregation. ie multiple nodes that have errors

func Test_ValidateGenericIngest(t *testing.T) {

	var (
		positiveCases = []genericAssertion{
			{
				name: "payload contains one node",
				payload: &testPayload{
					Graph: testGraph{
						Nodes: []testNode{
							{
								ID:    "1234",
								Kinds: []string{"a"},
								Properties: map[string]any{
									"hello": "world",
									"one":   2,
									"true":  false,
								},
							},
						},
					},
				},
			},
			{
				name: "payload contains one edge",
				payload: &testPayload{
					Graph: testGraph{
						Edges: []testEdge{
							{
								Start: &edgePiece{
									IDValue: "1234",
								},
								End: &edgePiece{
									IDValue: "5678",
								},
								Kind: "kind A",
								Properties: map[string]any{
									"hello": "world",
									"true":  false,
									"one":   2,
								},
							},
						},
					},
				},
			},
		}

		negativeCases = []genericAssertion{}
	)

	negativeCases = append(negativeCases, decodingFailureCases()...)
	negativeCases = append(negativeCases, schemaFailureCases()...)
	negativeCases = append(negativeCases, itemsWithMultipleFailureCases()...)

	for _, assertion := range negativeCases {
		var (
			testMessage = fmt.Sprintf("negative case failed. test name: %s", assertion.name)
			reader      io.Reader
		)

		if assertion.payload != nil { // test case uses go structure
			payload, err := json.Marshal(assertion.payload)
			assert.Nil(t, err, testMessage)
			reader = bytes.NewReader(payload)
		} else if assertion.rawPayload != "" { // test cases uses raw json
			reader = strings.NewReader(assertion.rawPayload)
		}

		err := ingest_service.ValidateGenericIngest(reader, true)
		fmt.Println(err)
		if len(assertion.errMsgs) > 0 {
			for _, validationError := range assertion.errMsgs {
				assert.ErrorContains(t, err, validationError, testMessage)
			}
		}
	}

	for _, assertion := range positiveCases {
		testMessage := fmt.Sprintf("positive case failed. test name: %s", assertion.name)
		// marshal the test structure into json to simulate input
		payload, err := json.Marshal(assertion.payload)
		assert.Nil(t, err, testMessage)

		reader := bytes.NewReader(payload)

		err = ingest_service.ValidateGenericIngest(reader, true)
		assert.Nil(t, err, testMessage)
	}
}

// these cases exercise the json.Decoder in different ways to produce (recoverable) UnmarshalTypeErrors
// and (unrecoverable) SyntaxErrors
func decodingFailureCases() []genericAssertion {
	return []genericAssertion{
		{ // UnmarshalType error, is recoverable
			name:       "decoding error: node is not a JSON Object",
			rawPayload: `{"graph": {"nodes": ["this is a string"]}}`,
			errMsgs:    []string{"[0] decode error", "type mismatch", "json: cannot unmarshal string"},
		},
		{ // SyntaxError, try to advance but may have to abort.
			name:       "decoding error: trailing comma in object",
			rawPayload: `{"graph": {"nodes": [{"id":"123",}]}}`,
			errMsgs:    []string{"[0] decode error", "syntax error", "invalid character '}' looking for beginning of object key string"},
		},
		{
			name:       "decoding error: unclosed object",
			rawPayload: `{"graph": {"nodes": [{"id":"123"]}}`,
			errMsgs:    []string{"[0] decode error", "syntax error", "invalid character ']' after object key:value pair"},
		},
		{
			name:       "decoding error: unquoted keys",
			rawPayload: `{"graph": {"nodes": [{id:"123"}]}}`,
			errMsgs:    []string{"[0] decode error", "syntax error", "invalid character 'i' looking for beginning of object key string"},
		},
	}
}

// these test cases represent all the ways a node or an edge can fail schema validation
func schemaFailureCases() []genericAssertion {
	return []genericAssertion{
		{
			name: "payload doesn't contain atleast one of nodes or edges",
			payload: &testPayload{
				Graph: testGraph{},
			},
			err: ingest.ErrEmptyIngest,
		},
		{
			name: "node validation: ID is null",
			payload: &testPayload{
				Graph: testGraph{
					Nodes: []testNode{
						{
							Kinds: []string{"kind A", "kind b"},
						},
					},
				},
			},
			errMsgs: []string{"validation failed for nodes[0]", "at '': missing property 'id'"},
		},
		{
			name: "node validation: ID is empty string",
			payload: &testPayload{
				Graph: testGraph{
					Nodes: []testNode{
						{
							ID:    "",
							Kinds: []string{"kind A", "kind b"},
						},
					},
				},
			},
			errMsgs: []string{"validation failed for nodes[0]", "at '': missing property 'id'"},
		},
		{
			name: "node validation: > than 2 kinds supplied",
			payload: &testPayload{
				Graph: testGraph{
					Nodes: []testNode{
						{
							ID:    "1234",
							Kinds: []string{"kind A", "kind b", "kind c"},
						},
					},
				},
			},
			errMsgs: []string{"validation failed for nodes[0]", "at '/kinds': maxItems: got 3, want 2"},
		},
		{
			name: "node validation: atleast one kind must be specified",
			payload: &testPayload{
				Graph: testGraph{
					Nodes: []testNode{
						{
							ID:    "1234",
							Kinds: []string{},
						},
					},
				},
			},
			errMsgs: []string{"validation failed for nodes[0]", "at '/kinds': minItems: got 0, want 1"},
		},
		{
			name: "node validation: kinds cannot be a null array",
			payload: &testPayload{
				Graph: testGraph{
					Nodes: []testNode{
						{
							ID: "1234",
						},
					},
				},
			},
			errMsgs: []string{"validation failed for nodes[0]", "at '/kinds': got null, want array"},
		},
		{
			name: "node validation: multiple issues. no node id, > 2 kinds supplied",
			payload: &testPayload{
				Graph: testGraph{
					Nodes: []testNode{
						{
							Kinds: []string{"kind A", "kind b", "kind c"},
						},
					},
				},
			},
			errMsgs: []string{"validation failed for nodes[0]", "at '/kinds': maxItems: got 3, want 2", "at '': missing property 'id'"},
		},
		{
			name: "edge validation: start not provided",
			payload: &testPayload{
				Graph: testGraph{
					Edges: []testEdge{
						{
							End: &edgePiece{
								IDValue: "a5678",
							},
							Kind: "kind A",
						},
					},
				},
			},
			errMsgs: []string{
				"validation failed for edges[0]",
				"at '/start': got null, want object"},
		},
		{
			name: "edge validation: start id not provided",
			payload: &testPayload{
				Graph: testGraph{
					Edges: []testEdge{
						{
							Start: &edgePiece{},
							End: &edgePiece{
								IDValue: "a5678",
							},
							Kind: "kind A",
						},
					},
				},
			},
			err:     ingest.ErrEdgeValidation,
			errMsgs: []string{"validation failed for edges[0]", "at '/start': missing property 'id_value'"},
		},
		{
			name: "edge validation: end not provided",
			payload: &testPayload{
				Graph: testGraph{
					Edges: []testEdge{
						{
							Start: &edgePiece{
								IDValue: "1234",
							},
							Kind: "kind A",
						},
					},
				},
			},
			err:     ingest.ErrEdgeValidation,
			errMsgs: []string{"validation failed for edges[0]", "at '/end': got null, want object"},
		},
		{
			name: "edge validation: end id not provided",
			payload: &testPayload{
				Graph: testGraph{
					Edges: []testEdge{
						{
							Start: &edgePiece{IDValue: "1234"},
							End:   &edgePiece{},
							Kind:  "kind A",
						},
					},
				},
			},
			err:     ingest.ErrEdgeValidation,
			errMsgs: []string{"validation failed for edges[0]", "at '/end': missing property 'id_value'"},
		},
		{
			name: "edge validation: end id is empty",
			payload: &testPayload{
				Graph: testGraph{
					Edges: []testEdge{
						{
							Start: &edgePiece{IDValue: "1234"},
							End:   &edgePiece{IDValue: ""},
							Kind:  "kind A",
						},
					},
				},
			},
			err:     ingest.ErrEdgeValidation,
			errMsgs: []string{"validation failed for edges[0]", "at '/end': missing property 'id_value'"},
		},
		{
			name: "edge validation: kind not provided",
			payload: &testPayload{
				Graph: testGraph{
					Edges: []testEdge{
						{
							Start: &edgePiece{
								IDValue: "1234",
							},
							End: &edgePiece{
								IDValue: "5678",
							},
						},
					},
				},
			},
			err:     ingest.ErrEdgeValidation,
			errMsgs: []string{"validation failed for edges[0]", "at '': missing property 'kind'"},
		},
		{
			name: "edge validation: multiple errors. start and end not provided",
			payload: &testPayload{
				Graph: testGraph{
					Edges: []testEdge{
						{
							Kind: "kind A",
						},
					},
				},
			},
			err:     ingest.ErrEdgeValidation,
			errMsgs: []string{"validation failed for edges[0]", "at '/end': got null, want object", "at '/start': got null, want object"},
		},
	}
}

func itemsWithMultipleFailureCases() []genericAssertion {
	return []genericAssertion{
		{
			name: "multiple nodes don't have an an ID.",
			payload: &testPayload{
				Graph: testGraph{
					Nodes: []testNode{
						{
							Kinds: []string{"kind A"},
						},
						{
							Kinds: []string{"kind A"},
						},
						{
							Kinds: []string{"kind A"},
						},
						{
							Kinds: []string{"kind A"},
						},
					},
				},
			},
			errMsgs: []string{"[0] validation error", "[1] validation error", "[2] validation error", "[3] validation error", "at '': missing property 'id'"},
		},
		{
			name: "multiple nodes with mixed errors.",
			payload: &testPayload{
				Graph: testGraph{
					Nodes: []testNode{
						{ // no ID
							Kinds: []string{"kind A"},
						},
						{ // no kinds
							Kinds: []string{},
						},
					},
				},
			},
			errMsgs: []string{"validation failed for nodes[0]: at '': missing property 'id'",
				"validation failed for nodes[2]: at '': missing property 'id', at '/kinds': minItems: got 0, want 1"},
		},
	}
}
