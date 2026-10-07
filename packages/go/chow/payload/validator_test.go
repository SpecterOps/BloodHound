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
package payload_test

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/specterops/bloodhound/packages/go/chow/ingest"
	"github.com/specterops/bloodhound/packages/go/chow/payload"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var emptyValidationReport = payload.ValidationReport{CriticalErrors: []payload.CriticalError{}, ValidationErrors: []payload.ValidationError{}}

type parseAndValidateAssertion struct {
	name               string
	payload            string
	expectedParsedData payload.ParsedData
	errValidationFunc  func(t *testing.T, report payload.ValidationReport, err error)
}

func repeatedInvalidNodesPayload(count int) string {
	invalidNodes := make([]string, count)
	for i := range invalidNodes {
		invalidNodes[i] = `{"id":"1","kinds":["A","A","A","A"]}`
	}

	return `{"metadata":{"source_kind":"hellobase"},"graph":{"nodes":[` + strings.Join(invalidNodes, ",") + `]}}`
}

func TestSitePayloadCompatibility(t *testing.T) {
	var siteDataTypes = []ingest.DataType{ingest.DataTypeSite, ingest.DataTypeSiteServer, ingest.DataTypeSiteSubnet}

	for _, siteDataType := range siteDataTypes {
		t.Run(string(siteDataType), func(t *testing.T) {
			payloadJSON := fmt.Sprintf(`{"meta":{"type":%q,"methods":0,"version":6},"data":[]}`, siteDataType)

			for _, parseMode := range []struct {
				name  string
				parse func(*payload.Validator) (payload.ParsedData, error)
			}{
				{
					name: "full validation",
					parse: func(validator *payload.Validator) (payload.ParsedData, error) {
						parsedData, _, err := validator.ParseAndValidate()
						return parsedData, err
					},
				},
				{
					name:  "metadata only",
					parse: (*payload.Validator).ParseMetadata,
				},
			} {
				t.Run(parseMode.name, func(t *testing.T) {
					schema, err := payload.LoadSchema()
					require.NoError(t, err)

					validator := payload.NewValidator(strings.NewReader(payloadJSON), schema)
					parsedData, err := parseMode.parse(&validator)
					require.NoError(t, err)
					assert.Equal(t, siteDataType, parsedData.PayloadType)
					assert.True(t, parsedData.OriginalData.MetadataFound)
					assert.Equal(t, siteDataType, parsedData.OriginalData.Metadata.Type)
					assert.Equal(t, uint64(0), uint64(parsedData.OriginalData.Metadata.Methods))
					assert.Equal(t, 6, parsedData.OriginalData.Metadata.Version)
				})
			}
		})
	}
}

func TestIncompleteGraphObjectAtEOFIsReported(t *testing.T) {
	var malformedPayload = `{"metadata":{"source_kind":"hellobase"},"graph":{"nodes":[]`

	schema, err := payload.LoadSchema()
	require.NoError(t, err)

	validator := payload.NewValidator(strings.NewReader(malformedPayload), schema)
	_, report, err := validator.ParseAndValidate()
	require.Error(t, err)
	require.NotEmpty(t, report.CriticalErrors)
}

func runParseAndValidateAssertions(t *testing.T, assertions []parseAndValidateAssertion) {
	t.Helper()

	schema, err := payload.LoadSchema()
	require.NoError(t, err)

	for _, assertion := range assertions {
		t.Run(assertion.name, func(t *testing.T) {
			v := payload.NewValidator(strings.NewReader(assertion.payload), schema)

			parsedData, validationReport, err := v.ParseAndValidate()
			assert.Equal(t, assertion.expectedParsedData, parsedData)
			assertion.errValidationFunc(t, validationReport, err)
		})
	}
}

func Test_ParseAndValidateOpenGraphPayloads(t *testing.T) {
	runParseAndValidateAssertions(t, []parseAndValidateAssertion{
		{
			name:               "successful opengraph payload",
			payload:            `{"metadata":{},"graph":{"nodes":[]}}`,
			expectedParsedData: payload.ParsedData{PayloadType: ingest.DataTypeOpenGraph, OpengraphData: payload.ParsedOpenGraphData{MetadataFound: true}},
			errValidationFunc: func(t *testing.T, report payload.ValidationReport, err error) {
				assert.Equal(t, emptyValidationReport, report)
				assert.NoError(t, err)
			},
		},
		{
			name:               "successful opengraph payload with no metadata",
			payload:            `{"graph":{"nodes":[]}}`,
			expectedParsedData: payload.ParsedData{PayloadType: ingest.DataTypeOpenGraph},
			errValidationFunc: func(t *testing.T, report payload.ValidationReport, err error) {
				assert.Equal(t, emptyValidationReport, report)
				assert.NoError(t, err)
			},
		},
		{
			name:               "successful opengraph metadata",
			payload:            `{"metadata":{"source_kind":"hellobase"},"graph":{"nodes":[]}}`,
			expectedParsedData: payload.ParsedData{PayloadType: ingest.DataTypeOpenGraph, OpengraphData: payload.ParsedOpenGraphData{MetadataFound: true, Metadata: ingest.OpengraphMetadata{SourceKind: "hellobase"}}},
			errValidationFunc: func(t *testing.T, report payload.ValidationReport, err error) {
				assert.Equal(t, emptyValidationReport, report)
				assert.NoError(t, err)
			},
		},
		{
			name:               "successful opengraph payload with $schema",
			payload:            `{"$schema":"test","metadata":{"source_kind":"hellobase"},"graph":{"nodes":[]}}`,
			expectedParsedData: payload.ParsedData{PayloadType: ingest.DataTypeOpenGraph, OpengraphData: payload.ParsedOpenGraphData{MetadataFound: true, Metadata: ingest.OpengraphMetadata{SourceKind: "hellobase"}}},
			errValidationFunc: func(t *testing.T, report payload.ValidationReport, err error) {
				assert.Equal(t, emptyValidationReport, report)
				assert.NoError(t, err)
			},
		},
		{
			name:               "unsuccessful opengraph metadata",
			payload:            `{"metadata":{"source_kind":1},"graph":{"nodes":[]}}`,
			expectedParsedData: payload.ParsedData{OpengraphData: payload.ParsedOpenGraphData{MetadataFound: true}},
			errValidationFunc: func(t *testing.T, report payload.ValidationReport, err error) {
				assert.ErrorIs(t, err, payload.ErrOpengraphMetadataValidation)

				assert.ElementsMatch(t, report.CriticalErrors, []payload.CriticalError{{Message: "opengraph metadata failed validation", Error: payload.ErrOpengraphMetadataValidation}})
			},
		},
		{
			name:               "unsuccessful opengraph no child tags",
			payload:            `{"graph":{}}`,
			expectedParsedData: payload.ParsedData{PayloadType: ingest.DataTypeOpenGraph},
			errValidationFunc: func(t *testing.T, report payload.ValidationReport, err error) {
				assert.ErrorIs(t, err, payload.ErrInvalidFileConfiguration)

				assert.ElementsMatch(t, report.CriticalErrors, []payload.CriticalError{{Message: "graph tag requires child nodes or edges tag", Error: payload.ErrInvalidFileConfiguration}})
			},
		},
		{
			name:               "unsuccessful opengraph metadata, invalid field",
			payload:            `{"metadata":{"random field":"hello"},"graph":{"nodes":[]}}`,
			expectedParsedData: payload.ParsedData{OpengraphData: payload.ParsedOpenGraphData{MetadataFound: true}},
			errValidationFunc: func(t *testing.T, report payload.ValidationReport, err error) {
				assert.ErrorIs(t, err, payload.ErrOpengraphMetadataValidation)

				assert.ElementsMatch(t, report.CriticalErrors, []payload.CriticalError{{Message: "opengraph metadata failed validation", Error: payload.ErrOpengraphMetadataValidation}})
			},
		},
	})
}

func Test_ParseAndValidateOpenGraphNodes(t *testing.T) {
	runParseAndValidateAssertions(t, []parseAndValidateAssertion{
		{
			name:               "successful opengraph payload with node",
			payload:            `{"metadata":{"source_kind":"hellobase"},"graph":{"nodes":[{"id":"TESTNODE","kinds":["User"],"properties":{"items":["hi"]}}]}}`,
			expectedParsedData: payload.ParsedData{PayloadType: ingest.DataTypeOpenGraph, OpengraphData: payload.ParsedOpenGraphData{MetadataFound: true, Metadata: ingest.OpengraphMetadata{SourceKind: "hellobase"}, NodesValidated: 1}},
			errValidationFunc: func(t *testing.T, report payload.ValidationReport, err error) {
				assert.Equal(t, emptyValidationReport, report)
				assert.NoError(t, err)
			},
		},
		{
			name:               "successful opengraph payload with uppercase node property name",
			payload:            `{"metadata":{"source_kind":"hellobase"},"graph":{"nodes":[{"id":"TESTNODE","kinds":["User"],"properties":{"DisplayName":"Alice"}}]}}`,
			expectedParsedData: payload.ParsedData{PayloadType: ingest.DataTypeOpenGraph, OpengraphData: payload.ParsedOpenGraphData{MetadataFound: true, Metadata: ingest.OpengraphMetadata{SourceKind: "hellobase"}, NodesValidated: 1}},
			errValidationFunc: func(t *testing.T, report payload.ValidationReport, err error) {
				assert.Equal(t, emptyValidationReport, report)
				assert.NoError(t, err)
			},
		},
		{
			name:               "unsuccessful opengraph payload, node id validation error",
			payload:            `{"metadata":{"source_kind":"hellobase"},"graph":{"nodes":[{"id":1,"kinds":["User"]}]}}`,
			expectedParsedData: payload.ParsedData{PayloadType: ingest.DataTypeOpenGraph, OpengraphData: payload.ParsedOpenGraphData{MetadataFound: true, Metadata: ingest.OpengraphMetadata{SourceKind: "hellobase"}, NodesValidated: 1}},
			errValidationFunc: func(t *testing.T, report payload.ValidationReport, err error) {
				assert.ErrorIs(t, err, payload.ErrValidationErrors)

				assert.ElementsMatch(t, report.ValidationErrors, []payload.ValidationError{
					{
						Location:  "/graph/nodes[0]",
						RawObject: `{"id":1,"kinds":["User"]}`,
						Errors:    []payload.ValidationErrorDetail{{Location: "/id", Error: "got number, want string"}},
					},
				})
			},
		},
		{
			name:               "unsuccessful opengraph payload, node kinds validation error",
			payload:            `{"metadata":{"source_kind":"hellobase"},"graph":{"nodes":[{"id":"TESTNODE","kinds":["User", 1]}]}}`,
			expectedParsedData: payload.ParsedData{PayloadType: ingest.DataTypeOpenGraph, OpengraphData: payload.ParsedOpenGraphData{MetadataFound: true, Metadata: ingest.OpengraphMetadata{SourceKind: "hellobase"}, NodesValidated: 1}},
			errValidationFunc: func(t *testing.T, report payload.ValidationReport, err error) {
				assert.ErrorIs(t, err, payload.ErrValidationErrors)

				assert.ElementsMatch(t, report.ValidationErrors, []payload.ValidationError{
					{
						Location:  "/graph/nodes[0]",
						RawObject: `{"id":"TESTNODE","kinds":["User", 1]}`,
						Errors:    []payload.ValidationErrorDetail{{Location: "/kinds/1", Error: "got number, want string"}},
					},
				})
			},
		},
		{
			name:               "unsuccessful opengraph payload, node kind tag prefix validation error",
			payload:            `{"metadata":{"source_kind":"hellobase"},"graph":{"nodes":[{"id":"TESTNODE","kinds":["Tag_Admin"]}]}}`,
			expectedParsedData: payload.ParsedData{PayloadType: ingest.DataTypeOpenGraph, OpengraphData: payload.ParsedOpenGraphData{MetadataFound: true, Metadata: ingest.OpengraphMetadata{SourceKind: "hellobase"}, NodesValidated: 1}},
			errValidationFunc: func(t *testing.T, report payload.ValidationReport, err error) {
				assert.ErrorIs(t, err, payload.ErrValidationErrors)

				assert.ElementsMatch(t, report.ValidationErrors, []payload.ValidationError{
					{
						Location:  "/graph/nodes[0]",
						RawObject: `{"id":"TESTNODE","kinds":["Tag_Admin"]}`,
						Errors:    []payload.ValidationErrorDetail{{Location: "/kinds/0", Error: "'not' failed"}},
					},
				})
			},
		},
		{
			name:               "unsuccessful opengraph payload, node kind standalone tag validation error",
			payload:            `{"metadata":{"source_kind":"hellobase"},"graph":{"nodes":[{"id":"TESTNODE","kinds":["tAg"]}]}}`,
			expectedParsedData: payload.ParsedData{PayloadType: ingest.DataTypeOpenGraph, OpengraphData: payload.ParsedOpenGraphData{MetadataFound: true, Metadata: ingest.OpengraphMetadata{SourceKind: "hellobase"}, NodesValidated: 1}},
			errValidationFunc: func(t *testing.T, report payload.ValidationReport, err error) {
				assert.ErrorIs(t, err, payload.ErrValidationErrors)

				assert.ElementsMatch(t, report.ValidationErrors, []payload.ValidationError{
					{
						Location:  "/graph/nodes[0]",
						RawObject: `{"id":"TESTNODE","kinds":["tAg"]}`,
						Errors:    []payload.ValidationErrorDetail{{Location: "/kinds/0", Error: "'not' failed"}},
					},
				})
			},
		},
		{
			name:               "unsuccessful opengraph payload, node properties validation error",
			payload:            `{"metadata":{"source_kind":"hellobase"},"graph":{"nodes":[{"id":"TESTNODE","kinds":["User"],"properties":{"items":{}}}]}}`,
			expectedParsedData: payload.ParsedData{PayloadType: ingest.DataTypeOpenGraph, OpengraphData: payload.ParsedOpenGraphData{MetadataFound: true, Metadata: ingest.OpengraphMetadata{SourceKind: "hellobase"}, NodesValidated: 1}},
			errValidationFunc: func(t *testing.T, report payload.ValidationReport, err error) {
				assert.ErrorIs(t, err, payload.ErrValidationErrors)

				assert.ElementsMatch(t, report.ValidationErrors, []payload.ValidationError{
					{
						Location:  "/graph/nodes[0]",
						RawObject: `{"id":"TESTNODE","kinds":["User"],"properties":{"items":{}}}`,
						Errors:    []payload.ValidationErrorDetail{{Location: "/properties/items", Error: "invalid type"}},
					},
				})
			},
		},
		{
			name:               "successful opengraph payload, null node properties",
			payload:            `{"metadata":{"source_kind":"hellobase"},"graph":{"nodes":[{"id":"TESTNODE","kinds":["User"],"properties":null}]}}`,
			expectedParsedData: payload.ParsedData{PayloadType: ingest.DataTypeOpenGraph, OpengraphData: payload.ParsedOpenGraphData{MetadataFound: true, Metadata: ingest.OpengraphMetadata{SourceKind: "hellobase"}, NodesValidated: 1}},
			errValidationFunc: func(t *testing.T, report payload.ValidationReport, err error) {
				assert.Equal(t, emptyValidationReport, report)
				assert.NoError(t, err)
			},
		},
		{
			name:               "unsuccessful opengraph payload, reserved objectid node property",
			payload:            `{"metadata":{"source_kind":"hellobase"},"graph":{"nodes":[{"id":"TESTNODE","kinds":["User"],"properties":{"objectid":"node-1"}}]}}`,
			expectedParsedData: payload.ParsedData{PayloadType: ingest.DataTypeOpenGraph, OpengraphData: payload.ParsedOpenGraphData{MetadataFound: true, Metadata: ingest.OpengraphMetadata{SourceKind: "hellobase"}, NodesValidated: 1}},
			errValidationFunc: func(t *testing.T, report payload.ValidationReport, err error) {
				assert.ErrorIs(t, err, payload.ErrValidationErrors)

				assert.ElementsMatch(t, report.ValidationErrors, []payload.ValidationError{
					{
						Location:  "/graph/nodes[0]",
						RawObject: `{"id":"TESTNODE","kinds":["User"],"properties":{"objectid":"node-1"}}`,
						Errors:    []payload.ValidationErrorDetail{{Location: "/properties", Error: "'not' failed"}},
					},
				})
			},
		},
		{
			name:               "unsuccessful opengraph payload, node multiple validation errors",
			payload:            `{"metadata":{"source_kind":"hellobase"},"graph":{"nodes":[{"id":1,"kinds":["User"],"properties":{"items":{}}}]}}`,
			expectedParsedData: payload.ParsedData{PayloadType: ingest.DataTypeOpenGraph, OpengraphData: payload.ParsedOpenGraphData{MetadataFound: true, Metadata: ingest.OpengraphMetadata{SourceKind: "hellobase"}, NodesValidated: 1}},
			errValidationFunc: func(t *testing.T, report payload.ValidationReport, err error) {
				assert.ErrorIs(t, err, payload.ErrValidationErrors)

				require.Len(t, report.ValidationErrors, 1)
				require.Equal(t, "/graph/nodes[0]", report.ValidationErrors[0].Location)
				require.Equal(t, `{"id":1,"kinds":["User"],"properties":{"items":{}}}`, report.ValidationErrors[0].RawObject)
				assert.ElementsMatch(t, report.ValidationErrors[0].Errors, []payload.ValidationErrorDetail{{Location: "/id", Error: "got number, want string"}, {Location: "/properties/items", Error: "invalid type"}})
			},
		},
		{
			name:               "unsuccessful opengraph payload, exceeds max validation errors",
			payload:            repeatedInvalidNodesPayload(17),
			expectedParsedData: payload.ParsedData{PayloadType: ingest.DataTypeOpenGraph, OpengraphData: payload.ParsedOpenGraphData{MetadataFound: true, Metadata: ingest.OpengraphMetadata{SourceKind: "hellobase"}, NodesValidated: 15}},
			errValidationFunc: func(t *testing.T, report payload.ValidationReport, err error) {
				assert.ErrorIs(t, err, payload.ErrMaxValidationErrors)

				require.Len(t, report.ValidationErrors, 15)
				for i, validationErr := range report.ValidationErrors {
					assert.Equal(t, fmt.Sprintf("/graph/nodes[%d]", i), validationErr.Location)
					assert.Equal(t, `{"id":"1","kinds":["A","A","A","A"]}`, validationErr.RawObject)
					assert.Equal(t, []payload.ValidationErrorDetail{{Location: "/kinds", Error: "maxItems: got 4, want 3"}}, validationErr.Errors)
				}
			},
		},
	})
}

func Test_ParseAndValidateOpenGraphEdges(t *testing.T) {
	runParseAndValidateAssertions(t, []parseAndValidateAssertion{
		{
			name:               "successful opengraph payload with edge",
			payload:            `{"metadata":{"source_kind":"hellobase"},"graph":{"nodes":[],"edges":[{"start":{"value":"TESTNODE"},"end":{"value":"TESTNODE2"},"kind":"RELATED","properties":{"items":["hi"]}}]}}`,
			expectedParsedData: payload.ParsedData{PayloadType: ingest.DataTypeOpenGraph, OpengraphData: payload.ParsedOpenGraphData{MetadataFound: true, Metadata: ingest.OpengraphMetadata{SourceKind: "hellobase"}, EdgesValidated: 1}},
			errValidationFunc: func(t *testing.T, report payload.ValidationReport, err error) {
				assert.Equal(t, emptyValidationReport, report)
				assert.NoError(t, err)
			},
		},
		{
			name:               "successful opengraph payload with uppercase edge property name",
			payload:            `{"metadata":{"source_kind":"hellobase"},"graph":{"nodes":[],"edges":[{"start":{"value":"TESTNODE"},"end":{"value":"TESTNODE2"},"kind":"RELATED","properties":{"DisplayName":"Alice"}}]}}`,
			expectedParsedData: payload.ParsedData{PayloadType: ingest.DataTypeOpenGraph, OpengraphData: payload.ParsedOpenGraphData{MetadataFound: true, Metadata: ingest.OpengraphMetadata{SourceKind: "hellobase"}, EdgesValidated: 1}},
			errValidationFunc: func(t *testing.T, report payload.ValidationReport, err error) {
				assert.Equal(t, emptyValidationReport, report)
				assert.NoError(t, err)
			},
		},
		{
			name:               "successful opengraph payload with edge property matching",
			payload:            `{"metadata":{"source_kind":"hellobase"},"graph":{"nodes":[],"edges":[{"start":{"match_by":"property","property_matchers":[{"key":"prop_1","operator":"equals","value":"ROHAN"}]},"end":{"match_by":"property","property_matchers":[{"key":"prop_1","operator":"equals","value":"WES"}]},"kind":"RELATED","properties":{"items":["hi"]}}]}}`,
			expectedParsedData: payload.ParsedData{PayloadType: ingest.DataTypeOpenGraph, OpengraphData: payload.ParsedOpenGraphData{MetadataFound: true, Metadata: ingest.OpengraphMetadata{SourceKind: "hellobase"}, EdgesValidated: 1}},
			errValidationFunc: func(t *testing.T, report payload.ValidationReport, err error) {
				assert.Equal(t, emptyValidationReport, report)
				assert.NoError(t, err)
			},
		},
		{
			name:               "unsuccessful opengraph payload, edge properties validation error",
			payload:            `{"metadata":{"source_kind":"hellobase"},"graph":{"nodes":[],"edges":[{"start":{"value":"TESTNODE"},"end":{"value":"TESTNODE2"},"kind":"RELATED","properties":{"items":{}}}]}}`,
			expectedParsedData: payload.ParsedData{PayloadType: ingest.DataTypeOpenGraph, OpengraphData: payload.ParsedOpenGraphData{MetadataFound: true, Metadata: ingest.OpengraphMetadata{SourceKind: "hellobase"}, EdgesValidated: 1}},
			errValidationFunc: func(t *testing.T, report payload.ValidationReport, err error) {
				assert.ErrorIs(t, err, payload.ErrValidationErrors)

				assert.ElementsMatch(t, report.ValidationErrors, []payload.ValidationError{
					{
						Location:  "/graph/edges[0]",
						RawObject: `{"start":{"value":"TESTNODE"},"end":{"value":"TESTNODE2"},"kind":"RELATED","properties":{"items":{}}}`,
						Errors:    []payload.ValidationErrorDetail{{Location: "/properties/items", Error: "invalid type"}},
					},
				})
			},
		},
		{
			name:               "successful opengraph payload, null edge properties",
			payload:            `{"metadata":{"source_kind":"hellobase"},"graph":{"nodes":[],"edges":[{"start":{"value":"TESTNODE"},"end":{"value":"TESTNODE2"},"kind":"RELATED","properties":null}]}}`,
			expectedParsedData: payload.ParsedData{PayloadType: ingest.DataTypeOpenGraph, OpengraphData: payload.ParsedOpenGraphData{MetadataFound: true, Metadata: ingest.OpengraphMetadata{SourceKind: "hellobase"}, EdgesValidated: 1}},
			errValidationFunc: func(t *testing.T, report payload.ValidationReport, err error) {
				assert.Equal(t, emptyValidationReport, report)
				assert.NoError(t, err)
			},
		},
		{
			name:               "successful opengraph payload, objectid edge property",
			payload:            `{"metadata":{"source_kind":"hellobase"},"graph":{"nodes":[],"edges":[{"start":{"value":"TESTNODE"},"end":{"value":"TESTNODE2"},"kind":"RELATED","properties":{"objectid":"edge-1"}}]}}`,
			expectedParsedData: payload.ParsedData{PayloadType: ingest.DataTypeOpenGraph, OpengraphData: payload.ParsedOpenGraphData{MetadataFound: true, Metadata: ingest.OpengraphMetadata{SourceKind: "hellobase"}, EdgesValidated: 1}},
			errValidationFunc: func(t *testing.T, report payload.ValidationReport, err error) {
				assert.Equal(t, emptyValidationReport, report)
				assert.NoError(t, err)
			},
		},
		{
			name:               "unsuccessful opengraph payload, edge id validation error",
			payload:            `{"metadata":{"source_kind":"hellobase"},"graph":{"nodes":[],"edges":[{"start":{"value":1},"end":{"value":"TESTNODE2"},"kind":"RELATED","properties":{"items":["hi"]}}]}}`,
			expectedParsedData: payload.ParsedData{PayloadType: ingest.DataTypeOpenGraph, OpengraphData: payload.ParsedOpenGraphData{MetadataFound: true, Metadata: ingest.OpengraphMetadata{SourceKind: "hellobase"}, EdgesValidated: 1}},
			errValidationFunc: func(t *testing.T, report payload.ValidationReport, err error) {
				assert.ErrorIs(t, err, payload.ErrValidationErrors)

				assert.ElementsMatch(t, report.ValidationErrors, []payload.ValidationError{
					{
						Location:  "/graph/edges[0]",
						RawObject: `{"start":{"value":1},"end":{"value":"TESTNODE2"},"kind":"RELATED","properties":{"items":["hi"]}}`,
						Errors:    []payload.ValidationErrorDetail{{Location: "/start/value", Error: "got number, want string"}},
					},
				})
			},
		},
		{
			name:               "unsuccessful opengraph payload, edge kind tag prefix validation error",
			payload:            `{"metadata":{"source_kind":"hellobase"},"graph":{"edges":[{"start":{"value":"TESTNODE"},"end":{"value":"TESTNODE2"},"kind":"TAG_Admin"}]}}`,
			expectedParsedData: payload.ParsedData{PayloadType: ingest.DataTypeOpenGraph, OpengraphData: payload.ParsedOpenGraphData{MetadataFound: true, Metadata: ingest.OpengraphMetadata{SourceKind: "hellobase"}, EdgesValidated: 1}},
			errValidationFunc: func(t *testing.T, report payload.ValidationReport, err error) {
				assert.ErrorIs(t, err, payload.ErrValidationErrors)

				assert.ElementsMatch(t, report.ValidationErrors, []payload.ValidationError{
					{
						Location:  "/graph/edges[0]",
						RawObject: `{"start":{"value":"TESTNODE"},"end":{"value":"TESTNODE2"},"kind":"TAG_Admin"}`,
						Errors:    []payload.ValidationErrorDetail{{Location: "/kind", Error: "'not' failed"}},
					},
				})
			},
		},
		{
			name:               "unsuccessful opengraph payload, edge kind standalone tag validation error",
			payload:            `{"metadata":{"source_kind":"hellobase"},"graph":{"edges":[{"start":{"value":"TESTNODE"},"end":{"value":"TESTNODE2"},"kind":"TaG"}]}}`,
			expectedParsedData: payload.ParsedData{PayloadType: ingest.DataTypeOpenGraph, OpengraphData: payload.ParsedOpenGraphData{MetadataFound: true, Metadata: ingest.OpengraphMetadata{SourceKind: "hellobase"}, EdgesValidated: 1}},
			errValidationFunc: func(t *testing.T, report payload.ValidationReport, err error) {
				assert.ErrorIs(t, err, payload.ErrValidationErrors)

				assert.ElementsMatch(t, report.ValidationErrors, []payload.ValidationError{
					{
						Location:  "/graph/edges[0]",
						RawObject: `{"start":{"value":"TESTNODE"},"end":{"value":"TESTNODE2"},"kind":"TaG"}`,
						Errors:    []payload.ValidationErrorDetail{{Location: "/kind", Error: "'not' failed"}},
					},
				})
			},
		},
		{
			name:               "successful opengraph payload with reserved endpoint kind filters",
			payload:            `{"metadata":{"source_kind":"hellobase"},"graph":{"edges":[{"start":{"value":"TESTNODE","kind":"tag_Admin"},"end":{"value":"TESTNODE2","kind":"TaG"},"kind":"RELATED"}]}}`,
			expectedParsedData: payload.ParsedData{PayloadType: ingest.DataTypeOpenGraph, OpengraphData: payload.ParsedOpenGraphData{MetadataFound: true, Metadata: ingest.OpengraphMetadata{SourceKind: "hellobase"}, EdgesValidated: 1}},
			errValidationFunc: func(t *testing.T, report payload.ValidationReport, err error) {
				assert.Equal(t, emptyValidationReport, report)
				assert.NoError(t, err)
			},
		},
		{
			name:               "unsuccessful opengraph payload, invalid edge property matching",
			payload:            `{"metadata":{"source_kind":"hellobase"},"graph":{"nodes":[],"edges":[{"start":{"match_by":"property","property_matchers":{"key":"prop_1","operator":"equals","value":"ROHAN"}},"end":{"match_by":"property","property_matchers":[{"key":"prop_1","operator":"equals","value":"WES"}]},"kind":"RELATED","properties":{"items":["hi"]}}]}}`,
			expectedParsedData: payload.ParsedData{PayloadType: ingest.DataTypeOpenGraph, OpengraphData: payload.ParsedOpenGraphData{MetadataFound: true, Metadata: ingest.OpengraphMetadata{SourceKind: "hellobase"}, EdgesValidated: 1}},
			errValidationFunc: func(t *testing.T, report payload.ValidationReport, err error) {
				assert.ErrorIs(t, err, payload.ErrValidationErrors)

				assert.ElementsMatch(t, report.ValidationErrors, []payload.ValidationError{
					{
						Location:  "/graph/edges[0]",
						RawObject: `{"start":{"match_by":"property","property_matchers":{"key":"prop_1","operator":"equals","value":"ROHAN"}},"end":{"match_by":"property","property_matchers":[{"key":"prop_1","operator":"equals","value":"WES"}]},"kind":"RELATED","properties":{"items":["hi"]}}`,
						Errors:    []payload.ValidationErrorDetail{{Location: "/start/property_matchers", Error: "got object, want array"}},
					},
				})
			},
		},
	})
}

func Test_ParseAndValidateOriginalPayloads(t *testing.T) {
	runParseAndValidateAssertions(t, []parseAndValidateAssertion{
		{
			name:               "successful original payload",
			payload:            `{"meta":{"methods": 0,"type":"sessions","count": 0,"version": 5},"data":[]}`,
			expectedParsedData: payload.ParsedData{PayloadType: ingest.DataTypeSession, OriginalData: payload.ParsedOriginalData{MetadataFound: true, Metadata: ingest.OriginalMetadata{Type: ingest.DataTypeSession, Methods: 0, Version: 5}}},
			errValidationFunc: func(t *testing.T, report payload.ValidationReport, err error) {
				assert.Equal(t, emptyValidationReport, report)
				assert.NoError(t, err)
			},
		},
		{
			name:               "unsuccessful original payload, no data tag",
			payload:            `{"meta":{"methods": 0,"type":"sessions","count": 0,"version":5}}`,
			expectedParsedData: payload.ParsedData{PayloadType: ingest.DataTypeSession, OriginalData: payload.ParsedOriginalData{MetadataFound: true, Metadata: ingest.OriginalMetadata{Type: ingest.DataTypeSession, Methods: 0, Version: 5}}},
			errValidationFunc: func(t *testing.T, report payload.ValidationReport, err error) {
				assert.ErrorIs(t, err, payload.ErrInvalidFileConfiguration)

				assert.ElementsMatch(t, report.CriticalErrors, []payload.CriticalError{{Message: "no data tag found to match original metadata tag", Error: payload.ErrInvalidFileConfiguration}})
			},
		},
		{
			name:               "unsuccessful original payload, no meta tag",
			payload:            `{"data":[]}`,
			expectedParsedData: payload.ParsedData{},
			errValidationFunc: func(t *testing.T, report payload.ValidationReport, err error) {
				assert.ErrorIs(t, err, payload.ErrInvalidFileConfiguration)

				assert.ElementsMatch(t, report.CriticalErrors, []payload.CriticalError{{Message: "no meta tag found to match original data tag", Error: payload.ErrInvalidFileConfiguration}})
			},
		},
		{
			name:               "unsuccessful original payload, duplicate meta tag",
			payload:            `{"meta":{"methods":0,"type":"sessions","count":0,"version":5},"meta":0,"data":[]}`,
			expectedParsedData: payload.ParsedData{PayloadType: ingest.DataTypeSession, OriginalData: payload.ParsedOriginalData{MetadataFound: true, Metadata: ingest.OriginalMetadata{Type: ingest.DataTypeSession, Methods: 0, Version: 5}}},
			errValidationFunc: func(t *testing.T, report payload.ValidationReport, err error) {
				assert.ErrorIs(t, err, payload.ErrInvalidFileConfiguration)

				assert.ElementsMatch(t, report.CriticalErrors, []payload.CriticalError{{Message: "duplicate top level meta tag found", Error: payload.ErrInvalidFileConfiguration}})
			},
		},
		{
			name:               "unsuccessful original payload, invalid meta",
			payload:            `{"data":[],"meta":0}`,
			expectedParsedData: payload.ParsedData{OriginalData: payload.ParsedOriginalData{MetadataFound: true}},
			errValidationFunc: func(t *testing.T, report payload.ValidationReport, err error) {
				require.Len(t, report.CriticalErrors, 1)
				var (
					criticalError = report.CriticalErrors[0]
					unmarshalErr  = &json.UnmarshalTypeError{}
				)

				assert.Equal(t, "failed to decode original metadata", criticalError.Message)
				assert.ErrorAs(t, criticalError.Error, &unmarshalErr)
				assert.ErrorAs(t, err, &unmarshalErr)
			},
		},
		{
			name:               "swapped order",
			payload:            `{"data":[],"meta":{"methods":0,"type":"sessions","count":0,"version":5}}`,
			expectedParsedData: payload.ParsedData{PayloadType: ingest.DataTypeSession, OriginalData: payload.ParsedOriginalData{MetadataFound: true, Metadata: ingest.OriginalMetadata{Type: ingest.DataTypeSession, Methods: 0, Version: 5}}},
			errValidationFunc: func(t *testing.T, report payload.ValidationReport, err error) {
				assert.Equal(t, emptyValidationReport, report)
				assert.NoError(t, err)
			},
		},
		{
			name:               "unsuccessful original payload, invalid type",
			payload:            `{"data":[],"meta":{"methods":0,"type":"invalid","count":0,"version":5}}`,
			expectedParsedData: payload.ParsedData{OriginalData: payload.ParsedOriginalData{MetadataFound: true}},
			errValidationFunc: func(t *testing.T, report payload.ValidationReport, err error) {
				assert.ErrorIs(t, err, payload.ErrInvalidDataType)

				assert.ElementsMatch(t, report.CriticalErrors, []payload.CriticalError{{Message: "invalid original metadata data type", Error: payload.ErrInvalidDataType}})
			},
		},
	})
}

func Test_ParseAndValidateTopLevelPayloadErrors(t *testing.T) {
	runParseAndValidateAssertions(t, []parseAndValidateAssertion{
		{
			name:               "unsuccessful payload, no valid tags",
			payload:            `{}`,
			expectedParsedData: payload.ParsedData{},
			errValidationFunc: func(t *testing.T, report payload.ValidationReport, err error) {
				assert.ErrorIs(t, err, payload.ErrInvalidFileConfiguration)

				assert.ElementsMatch(t, report.CriticalErrors, []payload.CriticalError{{Message: "no valid payload tags found", Error: payload.ErrInvalidFileConfiguration}})
			},
		},
		{
			name:               "unsuccessful payload, only unrecognized tags",
			payload:            `{"name":"example","value":123}`,
			expectedParsedData: payload.ParsedData{},
			errValidationFunc: func(t *testing.T, report payload.ValidationReport, err error) {
				assert.ErrorIs(t, err, payload.ErrInvalidFileConfiguration)

				assert.ElementsMatch(t, report.CriticalErrors, []payload.CriticalError{{Message: "no valid payload tags found", Error: payload.ErrInvalidFileConfiguration}})
			},
		},
		{
			name:               "enforce mutual exclusivity",
			payload:            `{"data":[],"graph":{}}`,
			expectedParsedData: payload.ParsedData{},
			errValidationFunc: func(t *testing.T, report payload.ValidationReport, err error) {
				assert.ErrorIs(t, err, payload.ErrInvalidFileConfiguration)

				assert.ElementsMatch(t, report.CriticalErrors, []payload.CriticalError{{Message: "cannot have both original data tag and opengraph graph tag", Error: payload.ErrInvalidFileConfiguration}})
			},
		},
		{
			name:               "successful payload, unknown top level scalar",
			payload:            `{"unknown":"meta","graph":{"nodes":[]}}`,
			expectedParsedData: payload.ParsedData{PayloadType: ingest.DataTypeOpenGraph},
			errValidationFunc: func(t *testing.T, report payload.ValidationReport, err error) {
				assert.Equal(t, emptyValidationReport, report)
				assert.NoError(t, err)
			},
		},
		{
			name:               "successful payload, unknown top level nested value",
			payload:            `{"unknown":{"meta":{"graph":["nodes","edges"]}},"graph":{"nodes":[]}}`,
			expectedParsedData: payload.ParsedData{PayloadType: ingest.DataTypeOpenGraph},
			errValidationFunc: func(t *testing.T, report payload.ValidationReport, err error) {
				assert.Equal(t, emptyValidationReport, report)
				assert.NoError(t, err)
			},
		},
		{
			name:               "unsuccessful payload, trailing data after object",
			payload:            `{"graph":{"nodes":[]}}{}`,
			expectedParsedData: payload.ParsedData{PayloadType: ingest.DataTypeOpenGraph},
			errValidationFunc: func(t *testing.T, report payload.ValidationReport, err error) {
				assert.ErrorContains(t, err, "expected EOF, instead got token: {")
				require.Len(t, report.CriticalErrors, 1)
				assert.Equal(t, "expected to hit the end of the file", report.CriticalErrors[0].Message)
				assert.ErrorContains(t, report.CriticalErrors[0].Error, "expected EOF, instead got token: {")
			},
		},
	})
}

func Test_ParseAndValidateConfigurationErrors(t *testing.T) {
	assertions := []struct {
		name             string
		payload          string
		expectedErr      error
		expectedCritical payload.CriticalError
		errContains      string
	}{
		{
			name:        "invalid top level json",
			payload:     `[]`,
			errContains: "expected open bracket",
			expectedCritical: payload.CriticalError{
				Message: "failed to enter json object",
			},
		},
		{
			name:        "empty input",
			payload:     ``,
			errContains: "EOF",
			expectedCritical: payload.CriticalError{
				Message: "failed to enter json object",
			},
		},
		{
			name:        "malformed top level object",
			payload:     `{"graph":{"nodes":[]},`,
			errContains: "EOF",
			expectedCritical: payload.CriticalError{
				Message: "failed parsing top level tag",
			},
		},
		{
			name:        "malformed unknown top level value",
			payload:     `{"unknown":{"nested":`,
			errContains: "EOF",
			expectedCritical: payload.CriticalError{
				Message: "failed to skip unrecognized top level tag: unknown",
			},
		},
		{
			name:        "malformed schema tag value",
			payload:     `{"$schema":`,
			errContains: "EOF",
			expectedCritical: payload.CriticalError{
				Message: "failed to skip unrecognized top level tag: $schema",
			},
		},
		{
			name:        "metadata tag missing value",
			payload:     `{"metadata":`,
			errContains: "EOF",
			expectedCritical: payload.CriticalError{
				Message: "failed decoding opengraph metadata to raw object",
			},
		},
		{
			name:        "data must be an array",
			payload:     `{"data":{},"meta":{"methods":0,"type":"sessions","count":0,"version":5}}`,
			errContains: "expected open square bracket",
			expectedCritical: payload.CriticalError{
				Message: "failed to enter data array",
			},
		},
		{
			name:        "data tag missing value",
			payload:     `{"data":`,
			errContains: "EOF",
			expectedCritical: payload.CriticalError{
				Message: "failed to enter data array",
			},
		},
		{
			name:        "graph must be an object",
			payload:     `{"graph":[]}`,
			errContains: "expected open bracket",
			expectedCritical: payload.CriticalError{
				Message: "failed to enter graph object",
			},
		},
		{
			name:        "graph nodes must be an array",
			payload:     `{"graph":{"nodes":{}}}`,
			errContains: "expected open square bracket",
			expectedCritical: payload.CriticalError{
				Message: "failed to enter graph nodes array",
			},
		},
		{
			name:        "malformed graph node object",
			payload:     `{"graph":{"nodes":[{"id":"node-1"`,
			errContains: "unexpected EOF",
			expectedCritical: payload.CriticalError{
				Message: "failed to decode nodes array object",
			},
		},
		{
			name:        "graph edges must be an array",
			payload:     `{"graph":{"edges":{}}}`,
			errContains: "expected open square bracket",
			expectedCritical: payload.CriticalError{
				Message: "failed to enter graph edges array",
			},
		},
		{
			name:        "unrecognized graph child tag",
			payload:     `{"graph":{"nodes":[],"strays":[]}}`,
			expectedErr: payload.ErrInvalidFileConfiguration,
			expectedCritical: payload.CriticalError{
				Message: "unrecognized graph child tag: strays",
				Error:   payload.ErrInvalidFileConfiguration,
			},
		},
		{
			name:        "duplicate data tag",
			payload:     `{"data":[],"data":[],"meta":{"methods":0,"type":"sessions","count":0,"version":5}}`,
			expectedErr: payload.ErrInvalidFileConfiguration,
			expectedCritical: payload.CriticalError{
				Message: "duplicate top level data tag found",
				Error:   payload.ErrInvalidFileConfiguration,
			},
		},
		{
			name:        "duplicate opengraph metadata tag",
			payload:     `{"metadata":{},"metadata":{},"graph":{"nodes":[]}}`,
			expectedErr: payload.ErrInvalidFileConfiguration,
			expectedCritical: payload.CriticalError{
				Message: "duplicate top level metadata tag found",
				Error:   payload.ErrInvalidFileConfiguration,
			},
		},
		{
			name:        "duplicate graph tag",
			payload:     `{"graph":{"nodes":[]},"graph":{"nodes":[]}}`,
			expectedErr: payload.ErrInvalidFileConfiguration,
			expectedCritical: payload.CriticalError{
				Message: "duplicate top level graph tag found",
				Error:   payload.ErrInvalidFileConfiguration,
			},
		},
		{
			name:        "duplicate graph nodes tag",
			payload:     `{"graph":{"nodes":[],"nodes":[]}}`,
			expectedErr: payload.ErrInvalidFileConfiguration,
			expectedCritical: payload.CriticalError{
				Message: "duplicate graph nodes tag found",
				Error:   payload.ErrInvalidFileConfiguration,
			},
		},
		{
			name:        "duplicate graph edges tag",
			payload:     `{"graph":{"edges":[],"edges":[]}}`,
			expectedErr: payload.ErrInvalidFileConfiguration,
			expectedCritical: payload.CriticalError{
				Message: "duplicate graph edges tag found",
				Error:   payload.ErrInvalidFileConfiguration,
			},
		},
		{
			name:        "legacy meta with opengraph metadata",
			payload:     `{"meta":{"methods":0,"type":"sessions","count":0,"version":5},"metadata":{},"data":[]}`,
			expectedErr: payload.ErrInvalidFileConfiguration,
			expectedCritical: payload.CriticalError{
				Message: "cannot have both original meta tag and opengraph metadata tag",
				Error:   payload.ErrInvalidFileConfiguration,
			},
		},
		{
			name:        "legacy meta with opengraph graph",
			payload:     `{"meta":{"methods":0,"type":"sessions","count":0,"version":5},"graph":{"nodes":[]},"data":[]}`,
			expectedErr: payload.ErrInvalidFileConfiguration,
			expectedCritical: payload.CriticalError{
				Message: "cannot have both original meta tag and opengraph graph tag",
				Error:   payload.ErrInvalidFileConfiguration,
			},
		},
		{
			name:        "legacy data with opengraph metadata",
			payload:     `{"data":[],"metadata":{},"meta":{"methods":0,"type":"sessions","count":0,"version":5}}`,
			expectedErr: payload.ErrInvalidFileConfiguration,
			expectedCritical: payload.CriticalError{
				Message: "cannot have both original data tag and opengraph metadata tag",
				Error:   payload.ErrInvalidFileConfiguration,
			},
		},
		{
			name:        "opengraph metadata without graph",
			payload:     `{"metadata":{"source_kind":"hellobase"}}`,
			expectedErr: payload.ErrInvalidFileConfiguration,
			expectedCritical: payload.CriticalError{
				Message: "no graph tag found to match opengraph metadata tag",
				Error:   payload.ErrInvalidFileConfiguration,
			},
		},
	}

	schema, err := payload.LoadSchema()
	require.NoError(t, err)

	for _, assertion := range assertions {
		t.Run(assertion.name, func(t *testing.T) {
			v := payload.NewValidator(strings.NewReader(assertion.payload), schema)

			_, report, err := v.ParseAndValidate()
			if assertion.expectedErr != nil {
				assert.ErrorIs(t, err, assertion.expectedErr)
			}
			if assertion.errContains != "" {
				assert.ErrorContains(t, err, assertion.errContains)
			}

			require.Len(t, report.CriticalErrors, 1)
			assert.Equal(t, assertion.expectedCritical.Message, report.CriticalErrors[0].Message)
			if assertion.expectedCritical.Error != nil {
				assert.ErrorIs(t, report.CriticalErrors[0].Error, assertion.expectedCritical.Error)
			}
		})
	}
}

type parseMetadataAssertion struct {
	name               string
	payload            string
	expectedParsedData payload.ParsedData
	errValidationFunc  func(t *testing.T, err error)
}

func Test_ParseMetadata(t *testing.T) {
	assertions := []parseMetadataAssertion{
		{
			name:    "legacy metadata",
			payload: `{"meta":{"methods":0,"type":"sessions","count":0,"version":5},"data":[]}`,
			expectedParsedData: payload.ParsedData{
				PayloadType: ingest.DataTypeSession,
				OriginalData: payload.ParsedOriginalData{
					MetadataFound: true,
					Metadata:      ingest.OriginalMetadata{Type: ingest.DataTypeSession, Methods: 0, Version: 5},
				},
			},
			errValidationFunc: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
		},
		{
			name:    "legacy metadata after data",
			payload: `{"data":[],"meta":{"methods":0,"type":"sessions","count":0,"version":5}}`,
			expectedParsedData: payload.ParsedData{
				PayloadType: ingest.DataTypeSession,
				OriginalData: payload.ParsedOriginalData{
					MetadataFound: true,
					Metadata:      ingest.OriginalMetadata{Type: ingest.DataTypeSession, Methods: 0, Version: 5},
				},
			},
			errValidationFunc: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
		},
		{
			name:    "opengraph metadata",
			payload: `{"metadata":{"source_kind":"hellobase"},"graph":{"nodes":[]}}`,
			expectedParsedData: payload.ParsedData{
				PayloadType: ingest.DataTypeOpenGraph,
				OpengraphData: payload.ParsedOpenGraphData{
					MetadataFound: true,
					Metadata:      ingest.OpengraphMetadata{SourceKind: "hellobase"},
				},
			},
			errValidationFunc: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
		},
		{
			name:    "opengraph metadata with null source kind",
			payload: `{"metadata":{"source_kind":null},"graph":{"nodes":[]}}`,
			expectedParsedData: payload.ParsedData{
				PayloadType: ingest.DataTypeOpenGraph,
				OpengraphData: payload.ParsedOpenGraphData{
					MetadataFound: true,
				},
			},
			errValidationFunc: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
		},
		{
			name:    "opengraph metadata with empty source kind",
			payload: `{"metadata":{"source_kind":""},"graph":{"nodes":[]}}`,
			expectedParsedData: payload.ParsedData{
				PayloadType: ingest.DataTypeOpenGraph,
				OpengraphData: payload.ParsedOpenGraphData{
					MetadataFound: true,
				},
			},
			errValidationFunc: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
		},
		{
			name:    "opengraph graph only",
			payload: `{"graph":{"nodes":[]}}`,
			expectedParsedData: payload.ParsedData{
				PayloadType: ingest.DataTypeOpenGraph,
			},
			errValidationFunc: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
		},
		{
			name:    "opengraph metadata after graph",
			payload: `{"graph":{"nodes":[]},"metadata":{"source_kind":"hellobase"}}`,
			expectedParsedData: payload.ParsedData{
				PayloadType: ingest.DataTypeOpenGraph,
				OpengraphData: payload.ParsedOpenGraphData{
					MetadataFound: true,
					Metadata:      ingest.OpengraphMetadata{SourceKind: "hellobase"},
				},
			},
			errValidationFunc: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
		},
		{
			name:    "opengraph metadata after graph string value",
			payload: `{"graph":"metadata","metadata":{"source_kind":"hellobase"}}`,
			expectedParsedData: payload.ParsedData{
				PayloadType: ingest.DataTypeOpenGraph,
				OpengraphData: payload.ParsedOpenGraphData{
					MetadataFound: true,
					Metadata:      ingest.OpengraphMetadata{SourceKind: "hellobase"},
				},
			},
			errValidationFunc: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
		},
		{
			name:    "opengraph metadata after unrecognized string value",
			payload: `{"unknown":"metadata","metadata":{"source_kind":"hellobase"}}`,
			expectedParsedData: payload.ParsedData{
				PayloadType: ingest.DataTypeOpenGraph,
				OpengraphData: payload.ParsedOpenGraphData{
					MetadataFound: true,
					Metadata:      ingest.OpengraphMetadata{SourceKind: "hellobase"},
				},
			},
			errValidationFunc: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
		},
		{
			name:               "no recognizable metadata",
			payload:            `{}`,
			expectedParsedData: payload.ParsedData{},
			errValidationFunc: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
		},
		{
			name:               "invalid legacy metadata",
			payload:            `{"meta":0}`,
			expectedParsedData: payload.ParsedData{OriginalData: payload.ParsedOriginalData{MetadataFound: true}},
			errValidationFunc: func(t *testing.T, err error) {
				var unmarshalErr *json.UnmarshalTypeError

				assert.ErrorAs(t, err, &unmarshalErr)
			},
		},
		{
			name:               "invalid opengraph metadata",
			payload:            `{"metadata":0}`,
			expectedParsedData: payload.ParsedData{PayloadType: ingest.DataTypeOpenGraph, OpengraphData: payload.ParsedOpenGraphData{MetadataFound: true}},
			errValidationFunc: func(t *testing.T, err error) {
				var unmarshalErr *json.UnmarshalTypeError

				assert.ErrorAs(t, err, &unmarshalErr)
			},
		},
		{
			name:    "malformed payload after graph tag",
			payload: `{"graph":`,
			expectedParsedData: payload.ParsedData{
				PayloadType: ingest.DataTypeOpenGraph,
			},
			errValidationFunc: func(t *testing.T, err error) {
				assert.ErrorContains(t, err, "EOF")
			},
		},
		{
			name:               "invalid top level json",
			payload:            `[]`,
			expectedParsedData: payload.ParsedData{},
			errValidationFunc: func(t *testing.T, err error) {
				assert.ErrorContains(t, err, "expected open bracket")
			},
		},
	}

	schema, err := payload.LoadSchema()
	require.NoError(t, err)

	for _, assertion := range assertions {
		t.Run(assertion.name, func(t *testing.T) {
			v := payload.NewValidator(strings.NewReader(assertion.payload), schema)

			parsedData, err := v.ParseMetadata()
			assert.Equal(t, assertion.expectedParsedData, parsedData)
			assertion.errValidationFunc(t, err)
		})
	}
}

func TestValidationError_Error(t *testing.T) {
	assertions := []struct {
		name          string
		validationErr payload.ValidationError
		expected      string
	}{
		{
			name: "location and details",
			validationErr: payload.ValidationError{
				Location: "/graph/nodes[0]",
				Errors: []payload.ValidationErrorDetail{
					{Location: "/id", Error: "got number, want string"},
					{Location: "/properties/items", Error: "invalid type"},
				},
			},
			expected: "validation error at /graph/nodes[0]: /id: got number, want string; /properties/items: invalid type",
		},
		{
			name: "detail without location",
			validationErr: payload.ValidationError{
				Errors: []payload.ValidationErrorDetail{
					{Error: "invalid type"},
				},
			},
			expected: "validation error: invalid type",
		},
		{
			name:          "no details",
			validationErr: payload.ValidationError{},
			expected:      "validation error",
		},
	}

	for _, assertion := range assertions {
		t.Run(assertion.name, func(t *testing.T) {
			assert.Equal(t, assertion.expected, assertion.validationErr.Error())
		})
	}
}
