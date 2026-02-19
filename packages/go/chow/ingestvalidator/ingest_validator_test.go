package validator

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/specterops/bloodhound/cmd/api/src/model/ingest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var emptyValidationReport = ValidationReport{CriticalErrors: []CriticalError{}, ValidationErrors: []ValidationError{}}

type parseAndValidateAssertion struct {
	name               string
	payload            string
	expectedParsedData ParsedData
	errValidationFunc  func(t *testing.T, report ValidationReport, err error)
}

func Test_ParseAndValidate(t *testing.T) {
	assertions := []parseAndValidateAssertion{
		// OpenGraph payload tests
		{
			name:               "successful opengraph payload",
			payload:            `{"metadata":{},"graph":{"nodes":[]}}`,
			expectedParsedData: ParsedData{PayloadType: ingest.DataTypeOpenGraph},
			errValidationFunc: func(t *testing.T, report ValidationReport, err error) {
				assert.Equal(t, report, emptyValidationReport)
				assert.NoError(t, err)
			},
		},
		{
			name:               "successful opengraph payload with no metadata",
			payload:            `{"graph":{"nodes":[]}}`,
			expectedParsedData: ParsedData{PayloadType: ingest.DataTypeOpenGraph},
			errValidationFunc: func(t *testing.T, report ValidationReport, err error) {
				assert.Equal(t, report, emptyValidationReport)
				assert.NoError(t, err)
			},
		},
		{
			name:               "successful opengraph metadata",
			payload:            `{"metadata":{"source_kind":"hellobase"},"graph":{"nodes":[]}}`,
			expectedParsedData: ParsedData{PayloadType: ingest.DataTypeOpenGraph, OpengraphData: ParsedOpenGraphData{Metadata: ingest.OpengraphMetadata{SourceKind: "hellobase"}}},
			errValidationFunc: func(t *testing.T, report ValidationReport, err error) {
				assert.Equal(t, report, emptyValidationReport)
				assert.NoError(t, err)
			},
		},
		{
			name:               "successful opengraph payload with node",
			payload:            `{"metadata":{"source_kind":"hellobase"},"graph":{"nodes":[{"id":"TESTNODE","kinds":["User"],"properties":{"items":["hi"]}}]}}`,
			expectedParsedData: ParsedData{PayloadType: ingest.DataTypeOpenGraph, OpengraphData: ParsedOpenGraphData{Metadata: ingest.OpengraphMetadata{SourceKind: "hellobase"}, NodesValidated: 1}},
			errValidationFunc: func(t *testing.T, report ValidationReport, err error) {
				assert.Equal(t, report, emptyValidationReport)
				assert.NoError(t, err)
			},
		},
		{
			name:               "unsuccessful opengraph payload, node id validation error",
			payload:            `{"metadata":{"source_kind":"hellobase"},"graph":{"nodes":[{"id":1,"kinds":["User"]}]}}`,
			expectedParsedData: ParsedData{PayloadType: ingest.DataTypeOpenGraph, OpengraphData: ParsedOpenGraphData{Metadata: ingest.OpengraphMetadata{SourceKind: "hellobase"}, NodesValidated: 1}},
			errValidationFunc: func(t *testing.T, report ValidationReport, err error) {
				assert.ErrorIs(t, err, ErrValidationErrors)

				assert.ElementsMatch(t, report.ValidationErrors, []ValidationError{
					{
						Location:  "/graph/nodes[0]",
						RawObject: `{"id":1,"kinds":["User"]}`,
						Errors:    []ValidationErrorDetail{{Location: "/id", Error: "got number, want string"}},
					},
				})
			},
		},
		{
			name:               "unsuccessful opengraph payload, node kinds validation error",
			payload:            `{"metadata":{"source_kind":"hellobase"},"graph":{"nodes":[{"id":"TESTNODE","kinds":["User", 1]}]}}`,
			expectedParsedData: ParsedData{PayloadType: ingest.DataTypeOpenGraph, OpengraphData: ParsedOpenGraphData{Metadata: ingest.OpengraphMetadata{SourceKind: "hellobase"}, NodesValidated: 1}},
			errValidationFunc: func(t *testing.T, report ValidationReport, err error) {
				assert.ErrorIs(t, err, ErrValidationErrors)

				assert.ElementsMatch(t, report.ValidationErrors, []ValidationError{
					{
						Location:  "/graph/nodes[0]",
						RawObject: `{"id":"TESTNODE","kinds":["User", 1]}`,
						Errors:    []ValidationErrorDetail{{Location: "/kinds/1", Error: "got number, want string"}},
					},
				})
			},
		},
		{
			name:               "unsuccessful opengraph payload, node properties validation error",
			payload:            `{"metadata":{"source_kind":"hellobase"},"graph":{"nodes":[{"id":"TESTNODE","kinds":["User"],"properties":{"items":{}}}]}}`,
			expectedParsedData: ParsedData{PayloadType: ingest.DataTypeOpenGraph, OpengraphData: ParsedOpenGraphData{Metadata: ingest.OpengraphMetadata{SourceKind: "hellobase"}, NodesValidated: 1}},
			errValidationFunc: func(t *testing.T, report ValidationReport, err error) {
				assert.ErrorIs(t, err, ErrValidationErrors)

				assert.ElementsMatch(t, report.ValidationErrors, []ValidationError{
					{
						Location:  "/graph/nodes[0]",
						RawObject: `{"id":"TESTNODE","kinds":["User"],"properties":{"items":{}}}`,
						Errors:    []ValidationErrorDetail{{Location: "/properties/items", Error: "invalid type"}},
					},
				})
			},
		},
		{
			name:               "unsuccessful opengraph payload, node multiple validation errors",
			payload:            `{"metadata":{"source_kind":"hellobase"},"graph":{"nodes":[{"id":1,"kinds":["User"],"properties":{"items":{}}}]}}`,
			expectedParsedData: ParsedData{PayloadType: ingest.DataTypeOpenGraph, OpengraphData: ParsedOpenGraphData{Metadata: ingest.OpengraphMetadata{SourceKind: "hellobase"}, NodesValidated: 1}},
			errValidationFunc: func(t *testing.T, report ValidationReport, err error) {
				assert.ErrorIs(t, err, ErrValidationErrors)

				require.Len(t, report.ValidationErrors, 1)
				require.Equal(t, report.ValidationErrors[0].Location, "/graph/nodes[0]")
				require.Equal(t, report.ValidationErrors[0].RawObject, `{"id":1,"kinds":["User"],"properties":{"items":{}}}`)
				assert.ElementsMatch(t, report.ValidationErrors[0].Errors, []ValidationErrorDetail{{Location: "/id", Error: "got number, want string"}, {Location: "/properties/items", Error: "invalid type"}})
			},
		},
		{
			name: "unsuccessful opengraph payload, exceeds max validation errors",
			payload: `{"metadata":{"source_kind":"hellobase"},"graph":{"nodes":[{"id":"1","kinds":["A","A","A","A"]},` +
				`{"id":"1","kinds":["A","A","A","A"]},{"id":"1","kinds":["A","A","A","A"]},{"id":"1","kinds":["A","A","A","A"]},{"id":"1","kinds":["A","A","A","A"]},` +
				`{"id":"1","kinds":["A","A","A","A"]},{"id":"1","kinds":["A","A","A","A"]},{"id":"1","kinds":["A","A","A","A"]},{"id":"1","kinds":["A","A","A","A"]},` +
				`{"id":"1","kinds":["A","A","A","A"]},{"id":"1","kinds":["A","A","A","A"]},{"id":"1","kinds":["A","A","A","A"]},{"id":"1","kinds":["A","A","A","A"]},` +
				`{"id":"1","kinds":["A","A","A","A"]},{"id":"1","kinds":["A","A","A","A"]},{"id":"1","kinds":["A","A","A","A"]},{"id":"1","kinds":["A","A","A","A"]},}]}}`,
			expectedParsedData: ParsedData{PayloadType: ingest.DataTypeOpenGraph, OpengraphData: ParsedOpenGraphData{Metadata: ingest.OpengraphMetadata{SourceKind: "hellobase"}, NodesValidated: 15}},
			errValidationFunc: func(t *testing.T, report ValidationReport, err error) {
				assert.ErrorIs(t, err, ErrMaxValidationErrors)

				assert.ElementsMatch(t, report.ValidationErrors, []ValidationError{
					{Location: "/graph/nodes[0]", RawObject: `{"id":"1","kinds":["A","A","A","A"]}`, Errors: []ValidationErrorDetail{{Location: "/kinds", Error: "maxItems: got 4, want 3"}}},
					{Location: "/graph/nodes[1]", RawObject: `{"id":"1","kinds":["A","A","A","A"]}`, Errors: []ValidationErrorDetail{{Location: "/kinds", Error: "maxItems: got 4, want 3"}}},
					{Location: "/graph/nodes[2]", RawObject: `{"id":"1","kinds":["A","A","A","A"]}`, Errors: []ValidationErrorDetail{{Location: "/kinds", Error: "maxItems: got 4, want 3"}}},
					{Location: "/graph/nodes[3]", RawObject: `{"id":"1","kinds":["A","A","A","A"]}`, Errors: []ValidationErrorDetail{{Location: "/kinds", Error: "maxItems: got 4, want 3"}}},
					{Location: "/graph/nodes[4]", RawObject: `{"id":"1","kinds":["A","A","A","A"]}`, Errors: []ValidationErrorDetail{{Location: "/kinds", Error: "maxItems: got 4, want 3"}}},
					{Location: "/graph/nodes[5]", RawObject: `{"id":"1","kinds":["A","A","A","A"]}`, Errors: []ValidationErrorDetail{{Location: "/kinds", Error: "maxItems: got 4, want 3"}}},
					{Location: "/graph/nodes[6]", RawObject: `{"id":"1","kinds":["A","A","A","A"]}`, Errors: []ValidationErrorDetail{{Location: "/kinds", Error: "maxItems: got 4, want 3"}}},
					{Location: "/graph/nodes[7]", RawObject: `{"id":"1","kinds":["A","A","A","A"]}`, Errors: []ValidationErrorDetail{{Location: "/kinds", Error: "maxItems: got 4, want 3"}}},
					{Location: "/graph/nodes[8]", RawObject: `{"id":"1","kinds":["A","A","A","A"]}`, Errors: []ValidationErrorDetail{{Location: "/kinds", Error: "maxItems: got 4, want 3"}}},
					{Location: "/graph/nodes[9]", RawObject: `{"id":"1","kinds":["A","A","A","A"]}`, Errors: []ValidationErrorDetail{{Location: "/kinds", Error: "maxItems: got 4, want 3"}}},
					{Location: "/graph/nodes[10]", RawObject: `{"id":"1","kinds":["A","A","A","A"]}`, Errors: []ValidationErrorDetail{{Location: "/kinds", Error: "maxItems: got 4, want 3"}}},
					{Location: "/graph/nodes[11]", RawObject: `{"id":"1","kinds":["A","A","A","A"]}`, Errors: []ValidationErrorDetail{{Location: "/kinds", Error: "maxItems: got 4, want 3"}}},
					{Location: "/graph/nodes[12]", RawObject: `{"id":"1","kinds":["A","A","A","A"]}`, Errors: []ValidationErrorDetail{{Location: "/kinds", Error: "maxItems: got 4, want 3"}}},
					{Location: "/graph/nodes[13]", RawObject: `{"id":"1","kinds":["A","A","A","A"]}`, Errors: []ValidationErrorDetail{{Location: "/kinds", Error: "maxItems: got 4, want 3"}}},
					{Location: "/graph/nodes[14]", RawObject: `{"id":"1","kinds":["A","A","A","A"]}`, Errors: []ValidationErrorDetail{{Location: "/kinds", Error: "maxItems: got 4, want 3"}}},
				})
			},
		},
		{
			name:               "successful opengraph payload with edge",
			payload:            `{"metadata":{"source_kind": "hellobase"},"graph": {"nodes":[], "edges": [{"start": {"value": "TESTNODE"},"end": {"value": "TESTNODE2"},"kind": "RELATED", "properties": {"items": ["hi"]}}]}}`,
			expectedParsedData: ParsedData{PayloadType: ingest.DataTypeOpenGraph, OpengraphData: ParsedOpenGraphData{Metadata: ingest.OpengraphMetadata{SourceKind: "hellobase"}, EdgesValidated: 1}},
			errValidationFunc: func(t *testing.T, report ValidationReport, err error) {
				assert.Equal(t, report, emptyValidationReport)
				assert.NoError(t, err)
			},
		},
		{
			name:               "unsuccessful opengraph payload, edge properties validation error",
			payload:            `{"metadata":{"source_kind":"hellobase"},"graph":{"nodes":[],"edges":[{"start":{"value":"TESTNODE"},"end":{"value":"TESTNODE2"},"kind":"RELATED","properties":{"items":{}}}]}}`,
			expectedParsedData: ParsedData{PayloadType: ingest.DataTypeOpenGraph, OpengraphData: ParsedOpenGraphData{Metadata: ingest.OpengraphMetadata{SourceKind: "hellobase"}, EdgesValidated: 1}},
			errValidationFunc: func(t *testing.T, report ValidationReport, err error) {
				assert.ErrorIs(t, err, ErrValidationErrors)

				assert.ElementsMatch(t, report.ValidationErrors, []ValidationError{
					{
						Location:  "/graph/edges[0]",
						RawObject: `{"start":{"value":"TESTNODE"},"end":{"value":"TESTNODE2"},"kind":"RELATED","properties":{"items":{}}}`,
						Errors:    []ValidationErrorDetail{{Location: "/properties/items", Error: "invalid type"}},
					},
				})
			},
		},
		{
			name:               "unsuccessful opengraph payload, edge id validation error",
			payload:            `{"metadata":{"source_kind":"hellobase"},"graph":{"nodes":[],"edges":[{"start":{"value":1},"end":{"value":"TESTNODE2"},"kind":"RELATED","properties":{"items":["hi"]}}]}}`,
			expectedParsedData: ParsedData{PayloadType: ingest.DataTypeOpenGraph, OpengraphData: ParsedOpenGraphData{Metadata: ingest.OpengraphMetadata{SourceKind: "hellobase"}, EdgesValidated: 1}},
			errValidationFunc: func(t *testing.T, report ValidationReport, err error) {
				assert.ErrorIs(t, err, ErrValidationErrors)

				assert.ElementsMatch(t, report.ValidationErrors, []ValidationError{
					{
						Location:  "/graph/edges[0]",
						RawObject: `{"start":{"value":1},"end":{"value":"TESTNODE2"},"kind":"RELATED","properties":{"items":["hi"]}}`,
						Errors:    []ValidationErrorDetail{{Location: "/start/value", Error: "got number, want string"}},
					},
				})
			},
		},
		{
			name:               "unsuccessful opengraph metadata",
			payload:            `{"metadata":{"source_kind":1},"graph":{"nodes":[]}}`,
			expectedParsedData: ParsedData{},
			errValidationFunc: func(t *testing.T, report ValidationReport, err error) {
				assert.ErrorIs(t, err, ErrOpengraphMetadataValidation)

				assert.ElementsMatch(t, report.CriticalErrors, []CriticalError{{Message: "opengraph metadata failed validation", Error: ErrOpengraphMetadataValidation}})
			},
		},
		{
			name:               "unsuccessful opengraph no nodes",
			payload:            `{"graph":{}}`,
			expectedParsedData: ParsedData{PayloadType: ingest.DataTypeOpenGraph},
			errValidationFunc: func(t *testing.T, report ValidationReport, err error) {
				assert.ErrorIs(t, err, ErrInvalidFileConfiguration)

				assert.ElementsMatch(t, report.CriticalErrors, []CriticalError{{Message: "graph tag requires child nodes tag", Error: ErrInvalidFileConfiguration}})
			},
		},
		{
			name:               "unsuccessful opengraph metadata, invalid field",
			payload:            `{"metadata":{"random field":"hello"},"graph":{"nodes":[]}}`,
			expectedParsedData: ParsedData{},
			errValidationFunc: func(t *testing.T, report ValidationReport, err error) {
				assert.ErrorIs(t, err, ErrOpengraphMetadataValidation)

				assert.ElementsMatch(t, report.CriticalErrors, []CriticalError{{Message: "opengraph metadata failed validation", Error: ErrOpengraphMetadataValidation}})
			},
		},
		// Original payload tests
		{
			name:               "successful original payload",
			payload:            `{"meta":{"methods": 0,"type":"sessions","count": 0,"version": 5},"data":[]}`,
			expectedParsedData: ParsedData{PayloadType: ingest.DataTypeSession, LegacyMetadata: ingest.LegacyMetadata{Type: ingest.DataTypeSession, Methods: 0, Version: 5}},
			errValidationFunc: func(t *testing.T, report ValidationReport, err error) {
				assert.Equal(t, report, emptyValidationReport)
				assert.NoError(t, err)
			},
		},
		{
			name:               "unsuccessful original payload, no data tag",
			payload:            `{"meta":{"methods": 0,"type":"sessions","count": 0,"version":5}}`,
			expectedParsedData: ParsedData{PayloadType: ingest.DataTypeSession, LegacyMetadata: ingest.LegacyMetadata{Type: ingest.DataTypeSession, Methods: 0, Version: 5}},
			errValidationFunc: func(t *testing.T, report ValidationReport, err error) {
				assert.ErrorIs(t, err, ErrInvalidFileConfiguration)

				assert.ElementsMatch(t, report.CriticalErrors, []CriticalError{{Message: "no data tag found to match legacy metadata tag", Error: ErrInvalidFileConfiguration}})
			},
		},
		{
			name:               "unsuccesful original payload, no meta tag",
			payload:            `{"data":[]}`,
			expectedParsedData: ParsedData{},
			errValidationFunc: func(t *testing.T, report ValidationReport, err error) {
				assert.ErrorIs(t, err, ErrInvalidFileConfiguration)

				assert.ElementsMatch(t, report.CriticalErrors, []CriticalError{{Message: "no meta tag found to match legacy data tag", Error: ErrInvalidFileConfiguration}})
			},
		},
		{
			name:               "unsuccesful payload, no valid tags",
			payload:            `{}`,
			expectedParsedData: ParsedData{},
			errValidationFunc: func(t *testing.T, report ValidationReport, err error) {
				assert.ErrorIs(t, err, ErrInvalidFileConfiguration)

				assert.ElementsMatch(t, report.CriticalErrors, []CriticalError{{Message: "no tags found", Error: ErrInvalidFileConfiguration}})
			},
		},
		{
			name:               "unsuccessful original payload, duplicate meta tag",
			payload:            `{"meta":{"methods":0,"type":"sessions","count":0,"version":5},"meta":0,"data":[]}`,
			expectedParsedData: ParsedData{PayloadType: ingest.DataTypeSession, LegacyMetadata: ingest.LegacyMetadata{Type: ingest.DataTypeSession, Methods: 0, Version: 5}},
			errValidationFunc: func(t *testing.T, report ValidationReport, err error) {
				assert.ErrorIs(t, err, ErrInvalidFileConfiguration)

				assert.ElementsMatch(t, report.CriticalErrors, []CriticalError{{Message: "duplicate top level meta tag found", Error: ErrInvalidFileConfiguration}})
			},
		},
		{
			name:               "unsuccessful original payload, invalid meta",
			payload:            `{"data":[],"meta":0}`,
			expectedParsedData: ParsedData{},
			errValidationFunc: func(t *testing.T, report ValidationReport, err error) {
				require.Len(t, report.CriticalErrors, 1)
				var (
					criticalError = report.CriticalErrors[0]
					unmarshalErr  = &json.UnmarshalTypeError{}
				)

				assert.Equal(t, "failed to decode legacy metadata", criticalError.Message)
				assert.ErrorAs(t, criticalError.Error, &unmarshalErr)
				assert.ErrorAs(t, err, &unmarshalErr)
			},
		},
		{
			name:               "swapped order",
			payload:            `{"data":[],"meta":{"methods":0,"type":"sessions","count":0,"version":5}}`,
			expectedParsedData: ParsedData{PayloadType: ingest.DataTypeSession, LegacyMetadata: ingest.LegacyMetadata{Type: ingest.DataTypeSession, Methods: 0, Version: 5}},
			errValidationFunc: func(t *testing.T, report ValidationReport, err error) {
				assert.Equal(t, report, emptyValidationReport)
				assert.NoError(t, err)
			},
		},
		{
			name:               "unsuccessful original payload, invalid type",
			payload:            `{"data":[],"meta":{"methods":0,"type":"invalid","count":0,"version":5}}`,
			expectedParsedData: ParsedData{},
			errValidationFunc: func(t *testing.T, report ValidationReport, err error) {
				assert.ErrorIs(t, err, ErrInvalidDataType)

				assert.ElementsMatch(t, report.CriticalErrors, []CriticalError{{Message: "invalid legacy metadata data type", Error: ErrInvalidDataType}})
			},
		},
		// Invalid payload tests
		{
			name:               "enforce mutual exclusivity",
			payload:            `{"data":[],"graph":{}}`,
			expectedParsedData: ParsedData{},
			errValidationFunc: func(t *testing.T, report ValidationReport, err error) {
				assert.ErrorIs(t, err, ErrInvalidFileConfiguration)

				assert.ElementsMatch(t, report.CriticalErrors, []CriticalError{{Message: "cannot have both legacy data tag and opengraph graph tag", Error: ErrInvalidFileConfiguration}})
			},
		},
	}

	schema, err := LoadIngestSchema()
	require.Nil(t, err)

	for _, assertion := range assertions {
		t.Run(assertion.name, func(t *testing.T) {
			v := NewValidator(strings.NewReader(assertion.payload), schema)

			parsedData, validationReport, err := v.ParseAndValidate()
			assert.Equal(t, assertion.expectedParsedData, parsedData)
			assertion.errValidationFunc(t, validationReport, err)
		})
	}
}
