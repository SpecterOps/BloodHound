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
package validator_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/specterops/bloodhound/cmd/api/src/model/ingest"
	validator "github.com/specterops/bloodhound/packages/go/chow/ingestvalidator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var emptyValidationReport = validator.ValidationReport{CriticalErrors: []validator.CriticalError{}, ValidationErrors: []validator.ValidationError{}}

type parseAndValidateAssertion struct {
	name               string
	payload            string
	expectedParsedData validator.ParsedData
	errValidationFunc  func(t *testing.T, report validator.ValidationReport, err error)
}

func Test_ParseAndValidate(t *testing.T) {
	assertions := []parseAndValidateAssertion{
		// OpenGraph payload tests
		{
			name:               "successful opengraph payload",
			payload:            `{"metadata":{},"graph":{"nodes":[]}}`,
			expectedParsedData: validator.ParsedData{PayloadType: ingest.DataTypeOpenGraph},
			errValidationFunc: func(t *testing.T, report validator.ValidationReport, err error) {
				assert.Equal(t, emptyValidationReport, report)
				assert.NoError(t, err)
			},
		},
		{
			name:               "successful opengraph payload with no metadata",
			payload:            `{"graph":{"nodes":[]}}`,
			expectedParsedData: validator.ParsedData{PayloadType: ingest.DataTypeOpenGraph},
			errValidationFunc: func(t *testing.T, report validator.ValidationReport, err error) {
				assert.Equal(t, emptyValidationReport, report)
				assert.NoError(t, err)
			},
		},
		{
			name:               "successful opengraph metadata",
			payload:            `{"metadata":{"source_kind":"hellobase"},"graph":{"nodes":[]}}`,
			expectedParsedData: validator.ParsedData{PayloadType: ingest.DataTypeOpenGraph, OpengraphData: validator.ParsedOpenGraphData{Metadata: ingest.OpengraphMetadata{SourceKind: "hellobase"}}},
			errValidationFunc: func(t *testing.T, report validator.ValidationReport, err error) {
				assert.Equal(t, emptyValidationReport, report)
				assert.NoError(t, err)
			},
		},
		{
			name:               "successful opengraph payload with node",
			payload:            `{"metadata":{"source_kind":"hellobase"},"graph":{"nodes":[{"id":"TESTNODE","kinds":["User"],"properties":{"items":["hi"]}}]}}`,
			expectedParsedData: validator.ParsedData{PayloadType: ingest.DataTypeOpenGraph, OpengraphData: validator.ParsedOpenGraphData{Metadata: ingest.OpengraphMetadata{SourceKind: "hellobase"}, NodesValidated: 1}},
			errValidationFunc: func(t *testing.T, report validator.ValidationReport, err error) {
				assert.Equal(t, emptyValidationReport, report)
				assert.NoError(t, err)
			},
		},
		{
			name:               "unsuccessful opengraph payload, node id validation error",
			payload:            `{"metadata":{"source_kind":"hellobase"},"graph":{"nodes":[{"id":1,"kinds":["User"]}]}}`,
			expectedParsedData: validator.ParsedData{PayloadType: ingest.DataTypeOpenGraph, OpengraphData: validator.ParsedOpenGraphData{Metadata: ingest.OpengraphMetadata{SourceKind: "hellobase"}, NodesValidated: 1}},
			errValidationFunc: func(t *testing.T, report validator.ValidationReport, err error) {
				assert.ErrorIs(t, err, validator.ErrValidationErrors)

				assert.ElementsMatch(t, report.ValidationErrors, []validator.ValidationError{
					{
						Location:  "/graph/nodes[0]",
						RawObject: `{"id":1,"kinds":["User"]}`,
						Errors:    []validator.ValidationErrorDetail{{Location: "/id", Error: "got number, want string"}},
					},
				})
			},
		},
		{
			name:               "unsuccessful opengraph payload, node kinds validation error",
			payload:            `{"metadata":{"source_kind":"hellobase"},"graph":{"nodes":[{"id":"TESTNODE","kinds":["User", 1]}]}}`,
			expectedParsedData: validator.ParsedData{PayloadType: ingest.DataTypeOpenGraph, OpengraphData: validator.ParsedOpenGraphData{Metadata: ingest.OpengraphMetadata{SourceKind: "hellobase"}, NodesValidated: 1}},
			errValidationFunc: func(t *testing.T, report validator.ValidationReport, err error) {
				assert.ErrorIs(t, err, validator.ErrValidationErrors)

				assert.ElementsMatch(t, report.ValidationErrors, []validator.ValidationError{
					{
						Location:  "/graph/nodes[0]",
						RawObject: `{"id":"TESTNODE","kinds":["User", 1]}`,
						Errors:    []validator.ValidationErrorDetail{{Location: "/kinds/1", Error: "got number, want string"}},
					},
				})
			},
		},
		{
			name:               "unsuccessful opengraph payload, node properties validation error",
			payload:            `{"metadata":{"source_kind":"hellobase"},"graph":{"nodes":[{"id":"TESTNODE","kinds":["User"],"properties":{"items":{}}}]}}`,
			expectedParsedData: validator.ParsedData{PayloadType: ingest.DataTypeOpenGraph, OpengraphData: validator.ParsedOpenGraphData{Metadata: ingest.OpengraphMetadata{SourceKind: "hellobase"}, NodesValidated: 1}},
			errValidationFunc: func(t *testing.T, report validator.ValidationReport, err error) {
				assert.ErrorIs(t, err, validator.ErrValidationErrors)

				assert.ElementsMatch(t, report.ValidationErrors, []validator.ValidationError{
					{
						Location:  "/graph/nodes[0]",
						RawObject: `{"id":"TESTNODE","kinds":["User"],"properties":{"items":{}}}`,
						Errors:    []validator.ValidationErrorDetail{{Location: "/properties/items", Error: "invalid type"}},
					},
				})
			},
		},
		{
			name:               "unsuccessful opengraph payload, node multiple validation errors",
			payload:            `{"metadata":{"source_kind":"hellobase"},"graph":{"nodes":[{"id":1,"kinds":["User"],"properties":{"items":{}}}]}}`,
			expectedParsedData: validator.ParsedData{PayloadType: ingest.DataTypeOpenGraph, OpengraphData: validator.ParsedOpenGraphData{Metadata: ingest.OpengraphMetadata{SourceKind: "hellobase"}, NodesValidated: 1}},
			errValidationFunc: func(t *testing.T, report validator.ValidationReport, err error) {
				assert.ErrorIs(t, err, validator.ErrValidationErrors)

				require.Len(t, report.ValidationErrors, 1)
				require.Equal(t, "/graph/nodes[0]", report.ValidationErrors[0].Location)
				require.Equal(t, `{"id":1,"kinds":["User"],"properties":{"items":{}}}`, report.ValidationErrors[0].RawObject)
				assert.ElementsMatch(t, report.ValidationErrors[0].Errors, []validator.ValidationErrorDetail{{Location: "/id", Error: "got number, want string"}, {Location: "/properties/items", Error: "invalid type"}})
			},
		},
		{
			name: "unsuccessful opengraph payload, exceeds max validation errors",
			payload: `{"metadata":{"source_kind":"hellobase"},"graph":{"nodes":[{"id":"1","kinds":["A","A","A","A"]},` +
				`{"id":"1","kinds":["A","A","A","A"]},{"id":"1","kinds":["A","A","A","A"]},{"id":"1","kinds":["A","A","A","A"]},{"id":"1","kinds":["A","A","A","A"]},` +
				`{"id":"1","kinds":["A","A","A","A"]},{"id":"1","kinds":["A","A","A","A"]},{"id":"1","kinds":["A","A","A","A"]},{"id":"1","kinds":["A","A","A","A"]},` +
				`{"id":"1","kinds":["A","A","A","A"]},{"id":"1","kinds":["A","A","A","A"]},{"id":"1","kinds":["A","A","A","A"]},{"id":"1","kinds":["A","A","A","A"]},` +
				`{"id":"1","kinds":["A","A","A","A"]},{"id":"1","kinds":["A","A","A","A"]},{"id":"1","kinds":["A","A","A","A"]},{"id":"1","kinds":["A","A","A","A"]}}]}}`,
			expectedParsedData: validator.ParsedData{PayloadType: ingest.DataTypeOpenGraph, OpengraphData: validator.ParsedOpenGraphData{Metadata: ingest.OpengraphMetadata{SourceKind: "hellobase"}, NodesValidated: 15}},
			errValidationFunc: func(t *testing.T, report validator.ValidationReport, err error) {
				assert.ErrorIs(t, err, validator.ErrMaxValidationErrors)

				assert.ElementsMatch(t, report.ValidationErrors, []validator.ValidationError{
					{Location: "/graph/nodes[0]", RawObject: `{"id":"1","kinds":["A","A","A","A"]}`, Errors: []validator.ValidationErrorDetail{{Location: "/kinds", Error: "maxItems: got 4, want 3"}}},
					{Location: "/graph/nodes[1]", RawObject: `{"id":"1","kinds":["A","A","A","A"]}`, Errors: []validator.ValidationErrorDetail{{Location: "/kinds", Error: "maxItems: got 4, want 3"}}},
					{Location: "/graph/nodes[2]", RawObject: `{"id":"1","kinds":["A","A","A","A"]}`, Errors: []validator.ValidationErrorDetail{{Location: "/kinds", Error: "maxItems: got 4, want 3"}}},
					{Location: "/graph/nodes[3]", RawObject: `{"id":"1","kinds":["A","A","A","A"]}`, Errors: []validator.ValidationErrorDetail{{Location: "/kinds", Error: "maxItems: got 4, want 3"}}},
					{Location: "/graph/nodes[4]", RawObject: `{"id":"1","kinds":["A","A","A","A"]}`, Errors: []validator.ValidationErrorDetail{{Location: "/kinds", Error: "maxItems: got 4, want 3"}}},
					{Location: "/graph/nodes[5]", RawObject: `{"id":"1","kinds":["A","A","A","A"]}`, Errors: []validator.ValidationErrorDetail{{Location: "/kinds", Error: "maxItems: got 4, want 3"}}},
					{Location: "/graph/nodes[6]", RawObject: `{"id":"1","kinds":["A","A","A","A"]}`, Errors: []validator.ValidationErrorDetail{{Location: "/kinds", Error: "maxItems: got 4, want 3"}}},
					{Location: "/graph/nodes[7]", RawObject: `{"id":"1","kinds":["A","A","A","A"]}`, Errors: []validator.ValidationErrorDetail{{Location: "/kinds", Error: "maxItems: got 4, want 3"}}},
					{Location: "/graph/nodes[8]", RawObject: `{"id":"1","kinds":["A","A","A","A"]}`, Errors: []validator.ValidationErrorDetail{{Location: "/kinds", Error: "maxItems: got 4, want 3"}}},
					{Location: "/graph/nodes[9]", RawObject: `{"id":"1","kinds":["A","A","A","A"]}`, Errors: []validator.ValidationErrorDetail{{Location: "/kinds", Error: "maxItems: got 4, want 3"}}},
					{Location: "/graph/nodes[10]", RawObject: `{"id":"1","kinds":["A","A","A","A"]}`, Errors: []validator.ValidationErrorDetail{{Location: "/kinds", Error: "maxItems: got 4, want 3"}}},
					{Location: "/graph/nodes[11]", RawObject: `{"id":"1","kinds":["A","A","A","A"]}`, Errors: []validator.ValidationErrorDetail{{Location: "/kinds", Error: "maxItems: got 4, want 3"}}},
					{Location: "/graph/nodes[12]", RawObject: `{"id":"1","kinds":["A","A","A","A"]}`, Errors: []validator.ValidationErrorDetail{{Location: "/kinds", Error: "maxItems: got 4, want 3"}}},
					{Location: "/graph/nodes[13]", RawObject: `{"id":"1","kinds":["A","A","A","A"]}`, Errors: []validator.ValidationErrorDetail{{Location: "/kinds", Error: "maxItems: got 4, want 3"}}},
					{Location: "/graph/nodes[14]", RawObject: `{"id":"1","kinds":["A","A","A","A"]}`, Errors: []validator.ValidationErrorDetail{{Location: "/kinds", Error: "maxItems: got 4, want 3"}}},
				})
			},
		},
		{
			name:               "successful opengraph payload with edge",
			payload:            `{"metadata":{"source_kind": "hellobase"},"graph": {"nodes":[], "edges": [{"start": {"value": "TESTNODE"},"end": {"value": "TESTNODE2"},"kind": "RELATED", "properties": {"items": ["hi"]}}]}}`,
			expectedParsedData: validator.ParsedData{PayloadType: ingest.DataTypeOpenGraph, OpengraphData: validator.ParsedOpenGraphData{Metadata: ingest.OpengraphMetadata{SourceKind: "hellobase"}, EdgesValidated: 1}},
			errValidationFunc: func(t *testing.T, report validator.ValidationReport, err error) {
				assert.Equal(t, emptyValidationReport, report)
				assert.NoError(t, err)
			},
		},
		{
			name:               "unsuccessful opengraph payload, edge properties validation error",
			payload:            `{"metadata":{"source_kind":"hellobase"},"graph":{"nodes":[],"edges":[{"start":{"value":"TESTNODE"},"end":{"value":"TESTNODE2"},"kind":"RELATED","properties":{"items":{}}}]}}`,
			expectedParsedData: validator.ParsedData{PayloadType: ingest.DataTypeOpenGraph, OpengraphData: validator.ParsedOpenGraphData{Metadata: ingest.OpengraphMetadata{SourceKind: "hellobase"}, EdgesValidated: 1}},
			errValidationFunc: func(t *testing.T, report validator.ValidationReport, err error) {
				assert.ErrorIs(t, err, validator.ErrValidationErrors)

				assert.ElementsMatch(t, report.ValidationErrors, []validator.ValidationError{
					{
						Location:  "/graph/edges[0]",
						RawObject: `{"start":{"value":"TESTNODE"},"end":{"value":"TESTNODE2"},"kind":"RELATED","properties":{"items":{}}}`,
						Errors:    []validator.ValidationErrorDetail{{Location: "/properties/items", Error: "invalid type"}},
					},
				})
			},
		},
		{
			name:               "unsuccessful opengraph payload, edge id validation error",
			payload:            `{"metadata":{"source_kind":"hellobase"},"graph":{"nodes":[],"edges":[{"start":{"value":1},"end":{"value":"TESTNODE2"},"kind":"RELATED","properties":{"items":["hi"]}}]}}`,
			expectedParsedData: validator.ParsedData{PayloadType: ingest.DataTypeOpenGraph, OpengraphData: validator.ParsedOpenGraphData{Metadata: ingest.OpengraphMetadata{SourceKind: "hellobase"}, EdgesValidated: 1}},
			errValidationFunc: func(t *testing.T, report validator.ValidationReport, err error) {
				assert.ErrorIs(t, err, validator.ErrValidationErrors)

				assert.ElementsMatch(t, report.ValidationErrors, []validator.ValidationError{
					{
						Location:  "/graph/edges[0]",
						RawObject: `{"start":{"value":1},"end":{"value":"TESTNODE2"},"kind":"RELATED","properties":{"items":["hi"]}}`,
						Errors:    []validator.ValidationErrorDetail{{Location: "/start/value", Error: "got number, want string"}},
					},
				})
			},
		},
		{
			name:               "unsuccessful opengraph metadata",
			payload:            `{"metadata":{"source_kind":1},"graph":{"nodes":[]}}`,
			expectedParsedData: validator.ParsedData{},
			errValidationFunc: func(t *testing.T, report validator.ValidationReport, err error) {
				assert.ErrorIs(t, err, validator.ErrOpengraphMetadataValidation)

				assert.ElementsMatch(t, report.CriticalErrors, []validator.CriticalError{{Message: "opengraph metadata failed validation", Error: validator.ErrOpengraphMetadataValidation}})
			},
		},
		{
			name:               "unsuccessful opengraph no child tags",
			payload:            `{"graph":{}}`,
			expectedParsedData: validator.ParsedData{PayloadType: ingest.DataTypeOpenGraph},
			errValidationFunc: func(t *testing.T, report validator.ValidationReport, err error) {
				assert.ErrorIs(t, err, validator.ErrInvalidFileConfiguration)

				assert.ElementsMatch(t, report.CriticalErrors, []validator.CriticalError{{Message: "graph tag requires child nodes or edges tag", Error: validator.ErrInvalidFileConfiguration}})
			},
		},
		{
			name:               "unsuccessful opengraph metadata, invalid field",
			payload:            `{"metadata":{"random field":"hello"},"graph":{"nodes":[]}}`,
			expectedParsedData: validator.ParsedData{},
			errValidationFunc: func(t *testing.T, report validator.ValidationReport, err error) {
				assert.ErrorIs(t, err, validator.ErrOpengraphMetadataValidation)

				assert.ElementsMatch(t, report.CriticalErrors, []validator.CriticalError{{Message: "opengraph metadata failed validation", Error: validator.ErrOpengraphMetadataValidation}})
			},
		},
		// Original payload tests
		{
			name:               "successful original payload",
			payload:            `{"meta":{"methods": 0,"type":"sessions","count": 0,"version": 5},"data":[]}`,
			expectedParsedData: validator.ParsedData{PayloadType: ingest.DataTypeSession, LegacyMetadata: ingest.OriginalMetadata{Type: ingest.DataTypeSession, Methods: 0, Version: 5}},
			errValidationFunc: func(t *testing.T, report validator.ValidationReport, err error) {
				assert.Equal(t, emptyValidationReport, report)
				assert.NoError(t, err)
			},
		},
		{
			name:               "unsuccessful original payload, no data tag",
			payload:            `{"meta":{"methods": 0,"type":"sessions","count": 0,"version":5}}`,
			expectedParsedData: validator.ParsedData{PayloadType: ingest.DataTypeSession, LegacyMetadata: ingest.OriginalMetadata{Type: ingest.DataTypeSession, Methods: 0, Version: 5}},
			errValidationFunc: func(t *testing.T, report validator.ValidationReport, err error) {
				assert.ErrorIs(t, err, validator.ErrInvalidFileConfiguration)

				assert.ElementsMatch(t, report.CriticalErrors, []validator.CriticalError{{Message: "no data tag found to match original metadata tag", Error: validator.ErrInvalidFileConfiguration}})
			},
		},
		{
			name:               "unsuccessful original payload, no meta tag",
			payload:            `{"data":[]}`,
			expectedParsedData: validator.ParsedData{},
			errValidationFunc: func(t *testing.T, report validator.ValidationReport, err error) {
				assert.ErrorIs(t, err, validator.ErrInvalidFileConfiguration)

				assert.ElementsMatch(t, report.CriticalErrors, []validator.CriticalError{{Message: "no meta tag found to match original data tag", Error: validator.ErrInvalidFileConfiguration}})
			},
		},
		{
			name:               "unsuccessful payload, no valid tags",
			payload:            `{}`,
			expectedParsedData: validator.ParsedData{},
			errValidationFunc: func(t *testing.T, report validator.ValidationReport, err error) {
				assert.ErrorIs(t, err, validator.ErrInvalidFileConfiguration)

				assert.ElementsMatch(t, report.CriticalErrors, []validator.CriticalError{{Message: "no tags found", Error: validator.ErrInvalidFileConfiguration}})
			},
		},
		{
			name:               "unsuccessful original payload, duplicate meta tag",
			payload:            `{"meta":{"methods":0,"type":"sessions","count":0,"version":5},"meta":0,"data":[]}`,
			expectedParsedData: validator.ParsedData{PayloadType: ingest.DataTypeSession, LegacyMetadata: ingest.OriginalMetadata{Type: ingest.DataTypeSession, Methods: 0, Version: 5}},
			errValidationFunc: func(t *testing.T, report validator.ValidationReport, err error) {
				assert.ErrorIs(t, err, validator.ErrInvalidFileConfiguration)

				assert.ElementsMatch(t, report.CriticalErrors, []validator.CriticalError{{Message: "duplicate top level meta tag found", Error: validator.ErrInvalidFileConfiguration}})
			},
		},
		{
			name:               "unsuccessful original payload, invalid meta",
			payload:            `{"data":[],"meta":0}`,
			expectedParsedData: validator.ParsedData{},
			errValidationFunc: func(t *testing.T, report validator.ValidationReport, err error) {
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
			expectedParsedData: validator.ParsedData{PayloadType: ingest.DataTypeSession, LegacyMetadata: ingest.OriginalMetadata{Type: ingest.DataTypeSession, Methods: 0, Version: 5}},
			errValidationFunc: func(t *testing.T, report validator.ValidationReport, err error) {
				assert.Equal(t, emptyValidationReport, report)
				assert.NoError(t, err)
			},
		},
		{
			name:               "unsuccessful original payload, invalid type",
			payload:            `{"data":[],"meta":{"methods":0,"type":"invalid","count":0,"version":5}}`,
			expectedParsedData: validator.ParsedData{},
			errValidationFunc: func(t *testing.T, report validator.ValidationReport, err error) {
				assert.ErrorIs(t, err, validator.ErrInvalidDataType)

				assert.ElementsMatch(t, report.CriticalErrors, []validator.CriticalError{{Message: "invalid original metadata data type", Error: validator.ErrInvalidDataType}})
			},
		},
		// Invalid payload tests
		{
			name:               "enforce mutual exclusivity",
			payload:            `{"data":[],"graph":{}}`,
			expectedParsedData: validator.ParsedData{},
			errValidationFunc: func(t *testing.T, report validator.ValidationReport, err error) {
				assert.ErrorIs(t, err, validator.ErrInvalidFileConfiguration)

				assert.ElementsMatch(t, report.CriticalErrors, []validator.CriticalError{{Message: "cannot have both original data tag and opengraph graph tag", Error: validator.ErrInvalidFileConfiguration}})
			},
		},
		{
			name:               "unsuccessful payload, unrecognized top level tag",
			payload:            `{"graph":{"nodes":[]},"pants":{}}`,
			expectedParsedData: validator.ParsedData{PayloadType: ingest.DataTypeOpenGraph},
			errValidationFunc: func(t *testing.T, report validator.ValidationReport, err error) {
				assert.ErrorIs(t, err, validator.ErrInvalidFileConfiguration)

				assert.ElementsMatch(t, report.CriticalErrors, []validator.CriticalError{{Message: "unrecognized top level tag: pants", Error: validator.ErrInvalidFileConfiguration}})
			},
		},
	}

	schema, err := validator.LoadIngestSchema()
	require.NoError(t, err)

	for _, assertion := range assertions {
		t.Run(assertion.name, func(t *testing.T) {
			v := validator.NewValidator(strings.NewReader(assertion.payload), schema)

			parsedData, validationReport, err := v.ParseAndValidate()
			assert.Equal(t, assertion.expectedParsedData, parsedData)
			assertion.errValidationFunc(t, validationReport, err)
		})
	}
}
