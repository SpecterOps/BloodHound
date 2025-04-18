// Copyright 2025 Specter Ops, Inc.
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
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/specterops/bloodhound/src/model/ingest"
	ingest_service "github.com/specterops/bloodhound/src/services/ingest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
			name:         "succesful generic payload",
			rawString:    `{"graph": {"nodes":[]}}`,
			err:          nil,
			expectedType: ingest.DataTypeGeneric,
		},
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

	schema, err := ingest_service.LoadIngestSchema()
	require.Nil(t, err)

	t.Run("Verify ValidateMetaTag for classic collection types-- {meta:{}, data:[]}", func(t *testing.T) {
		for _, assertion := range assertions {
			meta, err := ingest_service.ValidateMetaTag(strings.NewReader(assertion.rawString), schema, false)
			assert.ErrorIs(t, err, assertion.err)
			if assertion.err == nil {
				assert.Equal(t, assertion.expectedType, meta.Type)
			}
		}
	})
}

type genericIngestAssertion struct {
	name       string
	payload    *testPayload
	rawPayload string // for cases that require raw JSON to trip validation controls

	criticalErrMsgs       []string
	validationErrContains [][]string // do substring matches so that cosmetic changes to formatted error strings don't break test cases
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
	Value   string `json:"value,omitempty"`
	MatchBy string `json:"match_by,omitempty"`
	Kind    string `json:"kind,omitempty"`
}

type testPayload struct {
	// Graph testGraph `json:"graph"`
	Nodes []testNode `json:"nodes,omitempty"`
	Edges []testEdge `json:"edges,omitempty"`
}

func prepareReader(assertion genericIngestAssertion) (io.Reader, error) {
	if assertion.payload != nil { // test case uses go structure
		payload, err := json.Marshal(assertion.payload)
		if err != nil {
			return nil, err
		}
		return bytes.NewReader(payload), nil
	} else if assertion.rawPayload != "" { // test cases uses raw json
		return strings.NewReader(assertion.rawPayload), nil
	}

	return nil, errors.New("no payload defined in test case")
}

func Test_ValidateGenericIngest(t *testing.T) {
	var (
		positiveCases = []genericIngestAssertion{}
		negativeCases = []genericIngestAssertion{}
	)

	positiveCases = append(positiveCases, positiveGenericIngestCases()...)

	negativeCases = append(negativeCases, complexNestedPropertyCases()...)
	negativeCases = append(negativeCases, decodingFailureCases()...)
	negativeCases = append(negativeCases, criticalFailureCases()...)
	negativeCases = append(negativeCases, nodeSchemaFailureCases()...)
	negativeCases = append(negativeCases, edgeSchemaFailureCases()...)
	negativeCases = append(negativeCases, itemsWithMultipleFailureCases()...)

	ingestSchema, err := ingest_service.LoadIngestSchema()
	require.Nil(t, err, fmt.Sprintf("failed to load ingest schema: %s", err))

	for _, assertion := range negativeCases {
		t.Run(fmt.Sprintf("negative case: %s", assertion.name), func(t *testing.T) {
			reader, err := prepareReader(assertion)
			require.Nil(t, err)

			decoder := json.NewDecoder(reader)
			err = ingest_service.ValidateGenericIngest(decoder, ingestSchema)

			report, ok := err.(ingest_service.ValidationReport)
			require.True(t, ok)

			assert.Equal(t, len(assertion.criticalErrMsgs), len(report.CriticalErrors))

			// check critical errors
			for index, actualCriticalErr := range report.CriticalErrors {
				assert.Equal(t, assertion.criticalErrMsgs[index], actualCriticalErr.Message)
			}

			// check non-critical validation errors
			for i, expectedSubstrings := range assertion.validationErrContains {
				actual := report.ValidationErrors[i].Message
				for _, substr := range expectedSubstrings {
					assert.Contains(t, actual, substr)
				}
			}

		})
	}

	for _, assertion := range positiveCases {
		t.Run(fmt.Sprintf("positive case: %s", assertion.name), func(t *testing.T) {
			// marshal the test structure into json to simulate input
			payload, err := json.Marshal(assertion.payload)
			assert.Nil(t, err)

			reader := bytes.NewReader(payload)
			decoder := json.NewDecoder(reader)

			err = ingest_service.ValidateGenericIngest(decoder, ingestSchema)
			assert.Nil(t, err)
		})
	}
}

func positiveGenericIngestCases() []genericIngestAssertion {
	return []genericIngestAssertion{
		{
			name: "payload contains one node",
			payload: &testPayload{
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
		{
			name: "node has an array in property bag",
			payload: &testPayload{
				Nodes: []testNode{
					{
						ID:    "1234",
						Kinds: []string{"a"},
						Properties: map[string]any{
							"arr": []int{1, 2, 3},
						},
					},
				},
			},
		},
		{
			name: "edge has an array in property bag",
			payload: &testPayload{
				Edges: []testEdge{
					{
						Start: &edgePiece{Value: "1"},
						End:   &edgePiece{Value: "2"},
						Kind:  "a",
						Properties: map[string]any{
							"arr": []string{"one", "two"},
						},
					},
				},
			},
		},
		{
			name: "payload contains one edge",
			payload: &testPayload{
				Edges: []testEdge{
					{
						Start: &edgePiece{
							Value: "1234",
						},
						End: &edgePiece{
							Value: "5678",
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
		{
			name: "edge specifies match_by",
			payload: &testPayload{
				Edges: []testEdge{
					{
						Start: &edgePiece{
							Value:   "a name",
							MatchBy: "name",
						},
						End: &edgePiece{
							Value:   "another name",
							MatchBy: "name",
						},
						Kind: "kindA",
					},
				},
			},
		},
		{
			name: "edge specifies kind filter",
			payload: &testPayload{
				Edges: []testEdge{
					{
						Start: &edgePiece{
							Value: "a name",
							Kind:  "kindA",
						},
						End: &edgePiece{
							Value: "another name",
							Kind:  "kindB",
						},
						Kind: "kindA",
					},
				},
			},
		},
	}
}

func complexNestedPropertyCases() []genericIngestAssertion {
	return []genericIngestAssertion{
		{
			name: "node cannot have an object in properties",
			payload: &testPayload{
				Nodes: []testNode{
					{
						ID:    "1234",
						Kinds: []string{"one"},
						Properties: map[string]any{
							"nested": map[string]any{
								"hello": "world",
							},
						},
					},
				},
			},
			validationErrContains: [][]string{{"nodes[0] schema validation", "nested object cannot be stored as property", "remove \"nested\""}},
		},
		{
			name: "node cannot have objects inside an array in the property bag",
			payload: &testPayload{
				Nodes: []testNode{
					{
						ID:    "1234",
						Kinds: []string{"one"},
						Properties: map[string]any{
							"arr": []any{
								map[string]any{ // object nested inside of arr
									"1": 2,
								},
							},
						},
					},
				},
			},
			validationErrContains: [][]string{{"nodes[0] schema validation", "nested object cannot be stored as property", "remove \"arr\""}},
		},
		{
			name: "node cannot have an array with mixed types",
			payload: &testPayload{
				Nodes: []testNode{
					{
						ID:    "1234",
						Kinds: []string{"one"},
						Properties: map[string]any{
							"arr": []any{1, "two"},
						},
					},
				},
			},
			validationErrContains: [][]string{{"nodes[0] schema validation", "contains a mixed-type array"}},
		},
		{
			name: "edge cannot have an array with mixed types",
			payload: &testPayload{
				Edges: []testEdge{
					{
						Kind:  "a",
						Start: &edgePiece{Value: "hey"},
						End:   &edgePiece{Value: "sup"},
						Properties: map[string]any{
							"arr": []any{1, "two"},
						},
					},
				},
			},
			validationErrContains: [][]string{{"edges[0] schema validation", "contains a mixed-type array"}},
		},
		{
			name: "edge cannot have a nested object in properties",
			payload: &testPayload{
				Edges: []testEdge{
					{
						Kind:  "a",
						Start: &edgePiece{Value: "hey"},
						End:   &edgePiece{Value: "sup"},
						Properties: map[string]any{
							"nested": map[string]any{
								"hello": 1,
							},
						},
					},
				},
			},
			validationErrContains: [][]string{{"edges[0] schema validation", "nested object cannot be stored as property"}},
		},
		{
			name: "edge cannot have an array with a nested object",
			payload: &testPayload{
				Edges: []testEdge{
					{
						Kind:  "a",
						Start: &edgePiece{Value: "hey"},
						End:   &edgePiece{Value: "sup"},
						Properties: map[string]any{
							"array_with_nest": []any{map[string]any{"hello": "hi"}},
						},
					},
				},
			},
			validationErrContains: [][]string{{"edges[0] schema validation", "nested object cannot be stored as property"}},
		},
	}
}

// these cases exercise the json.Decoder in different ways to produce (recoverable) UnmarshalTypeErrors and (unrecoverable) SyntaxErrors
func decodingFailureCases() []genericIngestAssertion {
	return []genericIngestAssertion{
		{ // UnmarshalType error, is recoverable
			name:       "decoding error: node is not a JSON Object",
			rawPayload: `{"nodes": ["this is a string"]}`,
			validationErrContains: [][]string{
				{"nodes[0] type mismatch", "cannot unmarshal string into Go value of type map[string]interface {}"},
			},
		},
		{
			name:            "decoding error: trailing comma in object",
			rawPayload:      `{"nodes": [{"id":"123",}]}`,
			criticalErrMsgs: []string{"nodes[0] syntax error: invalid character '}' looking for beginning of object key string"},
		},
		{
			name:            "decoding error: unclosed object",
			rawPayload:      `{"nodes": [{"id":"123"]}`,
			criticalErrMsgs: []string{"nodes[0] syntax error: invalid character ']' after object key:value pair"},
		},
		{
			name:            "decoding error: unquoted keys",
			rawPayload:      `{"nodes": [{id:"123"}]}`,
			criticalErrMsgs: []string{"nodes[0] syntax error: invalid character 'i' looking for beginning of object key string"},
		},
		{
			name:            "validation and critical errors",
			rawPayload:      `{"nodes":[{"id":1234}], "edges":[}`,
			criticalErrMsgs: []string{"error decoding edges array: invalid character '}' looking for beginning of value"},
			validationErrContains: [][]string{
				{"nodes[0] schema validation", "at '': missing property 'kinds'", "at '/id': got number, want string"},
			},
		},
	}
}

// these cases exercise top-level mistakes that will halt the parse and return early
func criticalFailureCases() []genericIngestAssertion {
	return []genericIngestAssertion{
		{
			name:            "no opening { on payload",
			rawPayload:      "a",
			criticalErrMsgs: []string{"error decoding graph object: invalid character 'a' looking for beginning of value"},
		},
		{
			name:            "no closing } on payload",
			rawPayload:      `{"nodes": []`,
			criticalErrMsgs: []string{"error decoding graph object: EOF"},
		},
		{
			name:            "nodes array is not opened properly with '['",
			rawPayload:      `{"nodes": "some string"}`,
			criticalErrMsgs: []string{"error opening nodes array: expected '[', got some string"},
		},
		{
			name:            "nodes array is not closed properly with ']'",
			rawPayload:      `{"nodes": [{"id":"1234"}}`,
			criticalErrMsgs: []string{"error decoding nodes array: invalid character '}' after array element"},
			validationErrContains: [][]string{
				{"nodes[0] schema validation", "at '': missing property 'kinds'"},
			},
		},
		{
			name:            "edges array is not opened properly with '['",
			rawPayload:      `{"nodes": [], "edges": "a string value"}`,
			criticalErrMsgs: []string{"error opening edges array: expected '[', got a string value"},
		},
		{
			name:            "edges array is not closed properly with ']'",
			rawPayload:      `{"nodes": [], "edges": [{"id":"1234"}}`,
			criticalErrMsgs: []string{"error decoding edges array: invalid character '}' after array element"},
			validationErrContains: [][]string{
				{"edges[0] schema validation", "at '': missing properties 'start', 'end', 'kind'"},
			},
		},
	}
}

func nodeSchemaFailureCases() []genericIngestAssertion {
	return []genericIngestAssertion{
		{
			name:            "payload doesn't contain atleast one of nodes or edges",
			payload:         &testPayload{},
			criticalErrMsgs: []string{"graph tag is empty. atleast one of nodes: [] or edges: [] is required"},
		},
		{
			name: "node validation: ID is null",
			payload: &testPayload{
				Nodes: []testNode{
					{
						Kinds: []string{"kind A", "kind b"},
					},
				},
			},
			validationErrContains: [][]string{
				{"nodes[0]", "at '': missing property 'id'"},
			},
		},
		{
			name: "node validation: ID is empty string",
			payload: &testPayload{
				Nodes: []testNode{
					{
						ID:    "",
						Kinds: []string{"kind A", "kind b"},
					},
				},
			},
			validationErrContains: [][]string{
				{"nodes[0]", "at '': missing property 'id'"},
			},
		},
		{
			name: "node validation: > than 3 kinds supplied",
			payload: &testPayload{
				Nodes: []testNode{
					{
						ID:    "1234",
						Kinds: []string{"kind A", "kind b", "kind c", "kind d"},
					},
				},
			},
			validationErrContains: [][]string{
				{"nodes[0]", "at '/kinds': maxItems: got 4, want 3"},
			},
		},
		{
			name: "node validation: atleast one kind must be specified",
			payload: &testPayload{
				Nodes: []testNode{
					{
						ID:    "1234",
						Kinds: []string{},
					},
				},
			},
			validationErrContains: [][]string{
				{"nodes[0]", "at '/kinds': minItems: got 0, want 1"},
			},
		},
		{
			name: "node validation: kinds cannot be a null array",
			payload: &testPayload{
				Nodes: []testNode{
					{
						ID: "1234",
					},
				},
			},
			validationErrContains: [][]string{
				{"nodes[0]", "at '/kinds': got null, want array"},
			},
		},
		{
			name: "node validation: multiple issues. no node id, > 3 kinds supplied",
			payload: &testPayload{
				Nodes: []testNode{
					{
						Kinds: []string{"kind A", "kind b", "kind c", "kind d"},
					},
				},
			},
			validationErrContains: [][]string{
				{"nodes[0]", "at '': missing property 'id'", "at '/kinds': maxItems: got 4, want 3"},
			},
		},
	}
}

// these test cases represent all the ways an edge can fail schema validation.
func edgeSchemaFailureCases() []genericIngestAssertion {
	return []genericIngestAssertion{
		{
			name: "edge validation: start not provided",
			payload: &testPayload{
				Edges: []testEdge{
					{
						End: &edgePiece{
							Value: "a5678",
						},
						Kind: "kind A",
					},
				},
			},
			validationErrContains: [][]string{
				{"edges[0]", "at '/start': got null, want object"},
			},
		},
		{
			name: "edge validation: start id not provided",
			payload: &testPayload{
				Edges: []testEdge{
					{
						Start: &edgePiece{},
						End: &edgePiece{
							Value: "a5678",
						},
						Kind: "kind A",
					},
				},
			},
			validationErrContains: [][]string{
				{"edges[0]", "at '/start': missing property 'value'"},
			},
		},
		{
			name: "edge validation: end not provided",
			payload: &testPayload{
				Edges: []testEdge{
					{
						Start: &edgePiece{
							Value: "1234",
						},
						Kind: "kind A",
					},
				},
			},
			validationErrContains: [][]string{
				{"edges[0]", "at '/end': got null, want object"},
			},
		},
		{
			name: "edge validation: end id not provided",
			payload: &testPayload{
				Edges: []testEdge{
					{
						Start: &edgePiece{Value: "1234"},
						End:   &edgePiece{},
						Kind:  "kind A",
					},
				},
			},
			validationErrContains: [][]string{
				{"edges[0]", "at '/end': missing property 'value'"},
			},
		},
		{
			name: "edge validation: end id is empty",
			payload: &testPayload{
				Edges: []testEdge{
					{
						Start: &edgePiece{Value: "1234"},
						End:   &edgePiece{Value: ""},
						Kind:  "kind A",
					},
				},
			},
			validationErrContains: [][]string{
				{"edges[0]", "at '/end': missing property 'value'"},
			},
		},
		{
			name: "edge validation: kind not provided",
			payload: &testPayload{
				Edges: []testEdge{
					{
						Start: &edgePiece{
							Value: "1234",
						},
						End: &edgePiece{
							Value: "5678",
						},
					},
				},
			},
			validationErrContains: [][]string{
				{"edges[0]", "at '': missing property 'kind'"},
			},
		},
		{
			name: "edge validation: multiple errors. start and end not provided",
			payload: &testPayload{
				Edges: []testEdge{
					{
						Kind: "kind A",
					},
				},
			},
			validationErrContains: [][]string{
				{"edges[0]", "at '/start': got null, want object", "at '/end': got null, want object"},
			},
		},
		{
			name: "edge validation: start node match_by not valid enum value",
			payload: &testPayload{
				Edges: []testEdge{
					{
						Start: &edgePiece{
							Value:   "1234",
							MatchBy: "something bad",
						},
						End: &edgePiece{
							Value: "1234",
						},
						Kind: "kind A",
					},
				},
			},
			validationErrContains: [][]string{
				{"edges[0]", "at '/start/match_by'", "value must be one of 'id', 'name'"},
			},
		},
	}
}

func itemsWithMultipleFailureCases() []genericIngestAssertion {
	return []genericIngestAssertion{
		{
			name: "multiple nodes don't have an an ID.",
			payload: &testPayload{
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
			validationErrContains: [][]string{
				{"nodes[0]", "at '': missing property 'id'"},
				{"nodes[1]", "at '': missing property 'id'"},
				{"nodes[2]", "at '': missing property 'id'"},
				{"nodes[3]", "at '': missing property 'id'"},
			},
		},
		{
			name: "multiple nodes with mixed errors.",
			payload: &testPayload{
				Nodes: []testNode{
					{ // no ID
						Kinds: []string{"kind A"},
					},
					{ // no kinds
						Kinds: []string{},
					},
				},
			},
			validationErrContains: [][]string{
				{"nodes[0]", "at '': missing property 'id'"},
				{"nodes[1]", "at '': missing property 'id'", "at '/kinds': minItems: got 0, want 1"},
			},
		},
		{
			name: "nodes and edges both have errors",
			payload: &testPayload{
				Nodes: []testNode{
					{
						Kinds: []string{"a"},
					},
				},
				Edges: []testEdge{
					{
						Kind: "a",
					},
				},
			},
			validationErrContains: [][]string{
				{"nodes[0]", "at '': missing property 'id'"},
				{"edges[0]", "at '/start': got null, want object", "at '/end': got null, want object"},
			},
		},
		{
			name: "edges have errors",
			payload: &testPayload{
				Edges: []testEdge{
					{
						Kind: "a",
					},
					{
						Kind: "b",
					},
				},
			},
			validationErrContains: [][]string{
				{"edges[0]", "'/start': got null, want object", "'/end': got null, want object"},
				{"edges[1]", "'/start': got null, want object", "'/end': got null, want object"},
			},
		},
	}
}
