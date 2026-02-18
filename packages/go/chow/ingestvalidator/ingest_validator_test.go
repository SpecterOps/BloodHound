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
			payload:            `{"metadata":{},"graph": {"nodes":[]}}`,
			expectedParsedData: ParsedData{PayloadType: ingest.DataTypeOpenGraph},
			errValidationFunc: func(t *testing.T, report ValidationReport, err error) {
				assert.Equal(t, report, emptyValidationReport)
				assert.NoError(t, err)
			},
		},
		{
			name:               "successful opengraph payload with no metadata",
			payload:            `{"graph": {"nodes":[]}}`,
			expectedParsedData: ParsedData{PayloadType: ingest.DataTypeOpenGraph},
			errValidationFunc: func(t *testing.T, report ValidationReport, err error) {
				assert.Equal(t, report, emptyValidationReport)
				assert.NoError(t, err)
			},
		},
		{
			name:               "successful opengraph metadata",
			payload:            `{"metadata":{"source_kind": "hellobase"},"graph": {"nodes":[]}}`,
			expectedParsedData: ParsedData{PayloadType: ingest.DataTypeOpenGraph, OpengraphData: ParsedOpenGraphData{Metadata: ingest.OpengraphMetadata{SourceKind: "hellobase"}}},
			errValidationFunc: func(t *testing.T, report ValidationReport, err error) {
				assert.Equal(t, report, emptyValidationReport)
				assert.NoError(t, err)
			},
		},
		{
			name:               "successful opengraph payload with node",
			payload:            `{"metadata":{"source_kind": "hellobase"},"graph": {"nodes":[{"id": "TESTNODE","kinds": ["User"],"properties": {"items": ["hi"]}}]}}`,
			expectedParsedData: ParsedData{PayloadType: ingest.DataTypeOpenGraph, OpengraphData: ParsedOpenGraphData{Metadata: ingest.OpengraphMetadata{SourceKind: "hellobase"}, NodesValidated: 1}},
			errValidationFunc: func(t *testing.T, report ValidationReport, err error) {
				assert.Equal(t, report, emptyValidationReport)
				assert.NoError(t, err)
			},
		},
		{
			name:               "unsuccessful opengraph payload, node id validation error",
			payload:            `{"metadata":{"source_kind": "hellobase"},"graph": {"nodes":[{"id": 1,"kinds": ["User"]}]}}`,
			expectedParsedData: ParsedData{PayloadType: ingest.DataTypeOpenGraph, OpengraphData: ParsedOpenGraphData{Metadata: ingest.OpengraphMetadata{SourceKind: "hellobase"}, NodesValidated: 1}},
			errValidationFunc: func(t *testing.T, report ValidationReport, err error) {
				require.Len(t, report.ValidationErrors, 1)
				valErr := report.ValidationErrors[0]

				assert.ErrorIs(t, err, ErrValidationErrors)
				assert.Equal(t, `{"id": 1,"kinds": ["User"]}`, valErr.RawObject)
				assert.Equal(t, "/graph/nodes[0]", valErr.Location)

				require.Len(t, valErr.Errors, 1)
				idErr := valErr.Errors[0]

				assert.Equal(t, "/id", idErr.Location)
				assert.Equal(t, "got number, want string", idErr.Error)
			},
		},
		{
			name:               "unsuccessful opengraph payload, node properties validation error",
			payload:            `{"metadata":{"source_kind": "hellobase"},"graph": {"nodes":[{"id": "TESTNODE","kinds": ["User"],"properties": {"items": {}}}]}}`,
			expectedParsedData: ParsedData{PayloadType: ingest.DataTypeOpenGraph, OpengraphData: ParsedOpenGraphData{Metadata: ingest.OpengraphMetadata{SourceKind: "hellobase"}, NodesValidated: 1}},
			errValidationFunc: func(t *testing.T, report ValidationReport, err error) {
				require.Len(t, report.ValidationErrors, 1)
				valErr := report.ValidationErrors[0]

				assert.ErrorIs(t, err, ErrValidationErrors)
				assert.Equal(t, `{"id": "TESTNODE","kinds": ["User"],"properties": {"items": {}}}`, valErr.RawObject)
				assert.Equal(t, "/graph/nodes[0]", valErr.Location)

				require.Len(t, valErr.Errors, 1)
				schemaErr := valErr.Errors[0]

				assert.Equal(t, "/properties/items", schemaErr.Location)
				assert.Equal(t, "invalid type: object", schemaErr.Error)
			},
		},
		{
			name:               "unsuccessful opengraph payload, node multiple validation errors",
			payload:            `{"metadata":{"source_kind": "hellobase"},"graph": {"nodes":[{"id": 1,"kinds": ["User"], "properties": {"items": {}}}]}}`,
			expectedParsedData: ParsedData{PayloadType: ingest.DataTypeOpenGraph, OpengraphData: ParsedOpenGraphData{Metadata: ingest.OpengraphMetadata{SourceKind: "hellobase"}, NodesValidated: 1}},
			errValidationFunc: func(t *testing.T, report ValidationReport, err error) {
				require.Len(t, report.ValidationErrors, 1)
				valErr := report.ValidationErrors[0]

				assert.ErrorIs(t, err, ErrValidationErrors)
				assert.Equal(t, `{"id": 1,"kinds": ["User"], "properties": {"items": {}}}`, valErr.RawObject)
				assert.Equal(t, "/graph/nodes[0]", valErr.Location)

				require.Len(t, valErr.Errors, 2)
				assert.ElementsMatch(t, []ValidationErrorDetail{{Location: "/id", Error: "got number, want string"}, {Location: "/properties/items", Error: "invalid type: object"}}, valErr.Errors)
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
			payload:            `{"metadata":{"source_kind": "hellobase"},"graph": {"nodes":[], "edges": [{"start": {"value": "TESTNODE"},"end": {"value": "TESTNODE2"},"kind": "RELATED", "properties": {"items": {}}}]}}`,
			expectedParsedData: ParsedData{PayloadType: ingest.DataTypeOpenGraph, OpengraphData: ParsedOpenGraphData{Metadata: ingest.OpengraphMetadata{SourceKind: "hellobase"}, EdgesValidated: 1}},
			errValidationFunc: func(t *testing.T, report ValidationReport, err error) {
				require.Len(t, report.ValidationErrors, 1)
				valErr := report.ValidationErrors[0]

				assert.ErrorIs(t, err, ErrValidationErrors)
				assert.Equal(t, `{"start": {"value": "TESTNODE"},"end": {"value": "TESTNODE2"},"kind": "RELATED", "properties": {"items": {}}}`, valErr.RawObject)
				assert.Equal(t, "/graph/edges[0]", valErr.Location)

				require.Len(t, valErr.Errors, 1)
				schemaErr := valErr.Errors[0]

				assert.Equal(t, "/properties/items", schemaErr.Location)
				assert.Equal(t, "invalid type: object", schemaErr.Error)
			},
		},
		{
			name:               "unsuccessful opengraph payload, edge id validation error",
			payload:            `{"metadata":{"source_kind": "hellobase"},"graph": {"nodes":[], "edges": [{"start": {"value": 1},"end": {"value": "TESTNODE2"},"kind": "RELATED", "properties": {"items": ["hi"]}}]}}`,
			expectedParsedData: ParsedData{PayloadType: ingest.DataTypeOpenGraph, OpengraphData: ParsedOpenGraphData{Metadata: ingest.OpengraphMetadata{SourceKind: "hellobase"}, EdgesValidated: 1}},
			errValidationFunc: func(t *testing.T, report ValidationReport, err error) {
				require.Len(t, report.ValidationErrors, 1)
				valErr := report.ValidationErrors[0]

				assert.ErrorIs(t, err, ErrValidationErrors)
				assert.Equal(t, `{"start": {"value": 1},"end": {"value": "TESTNODE2"},"kind": "RELATED", "properties": {"items": ["hi"]}}`, valErr.RawObject)
				assert.Equal(t, "/graph/edges[0]", valErr.Location)

				require.Len(t, valErr.Errors, 1)
				schemaErr := valErr.Errors[0]

				assert.Equal(t, "/start/value", schemaErr.Location)
				assert.Equal(t, "got number, want string", schemaErr.Error)
			},
		},
		{
			name:               "unsuccessful opengraph metadata",
			payload:            `{"metadata":{"source_kind": 1},"graph": {"nodes":[]}}`,
			expectedParsedData: ParsedData{},
			errValidationFunc: func(t *testing.T, report ValidationReport, err error) {
				require.Len(t, report.CriticalErrors, 1)
				require.Len(t, report.ValidationErrors, 1)
				var (
					criticalError = report.CriticalErrors[0]
				)

				assert.Equal(t, "opengraph metadata failed validation", criticalError.Message)
				assert.ErrorIs(t, criticalError.Error, ErrOpengraphMetadataValidation)
				assert.ErrorIs(t, err, ErrOpengraphMetadataValidation)
			},
		},
		{
			name:               "unsuccessful opengraph no nodes",
			payload:            `{"graph": {}}`,
			expectedParsedData: ParsedData{PayloadType: ingest.DataTypeOpenGraph},
			errValidationFunc: func(t *testing.T, report ValidationReport, err error) {
				require.Len(t, report.CriticalErrors, 1)
				var (
					criticalError = report.CriticalErrors[0]
				)

				assert.Equal(t, "graph tag requires child nodes tag", criticalError.Message)
				assert.ErrorIs(t, criticalError.Error, ErrInvalidFileConfiguration)
				assert.ErrorIs(t, err, ErrInvalidFileConfiguration)
			},
		},
		{
			name:               "unsuccessful opengraph metadata, invalid field",
			payload:            `{"metadata":{"random field": "hello"},"graph": {"nodes":[]}}`,
			expectedParsedData: ParsedData{},
			errValidationFunc: func(t *testing.T, report ValidationReport, err error) {
				require.Len(t, report.CriticalErrors, 1)
				require.Len(t, report.ValidationErrors, 1)
				var (
					criticalError = report.CriticalErrors[0]
				)

				assert.Equal(t, "opengraph metadata failed validation", criticalError.Message)
				assert.ErrorIs(t, criticalError.Error, ErrOpengraphMetadataValidation)
				assert.ErrorIs(t, err, ErrOpengraphMetadataValidation)
			},
		},
		// Original payload tests
		{
			name:               "successful original payload",
			payload:            `{"meta": {"methods": 0, "type": "sessions", "count": 0, "version": 5}, "data": []}`,
			expectedParsedData: ParsedData{PayloadType: ingest.DataTypeSession, LegacyMetadata: ingest.LegacyMetadata{Type: ingest.DataTypeSession, Methods: 0, Version: 5}},
			errValidationFunc: func(t *testing.T, report ValidationReport, err error) {
				assert.Equal(t, report, emptyValidationReport)
				assert.NoError(t, err)
			},
		},
		{
			name:               "unsuccessful original payload, no data tag",
			payload:            `{"meta": {"methods": 0, "type": "sessions", "count": 0, "version": 5}}`,
			expectedParsedData: ParsedData{PayloadType: ingest.DataTypeSession, LegacyMetadata: ingest.LegacyMetadata{Type: ingest.DataTypeSession, Methods: 0, Version: 5}},
			errValidationFunc: func(t *testing.T, report ValidationReport, err error) {
				require.Len(t, report.CriticalErrors, 1)
				var (
					criticalError = report.CriticalErrors[0]
				)

				assert.Equal(t, "no data tag found to match legacy metadata tag", criticalError.Message)
				assert.ErrorIs(t, criticalError.Error, ErrInvalidFileConfiguration)
				assert.ErrorIs(t, err, ErrInvalidFileConfiguration)
			},
		},
		{
			name:               "unsuccesful original payload, no meta tag",
			payload:            `{"data": []}`,
			expectedParsedData: ParsedData{},
			errValidationFunc: func(t *testing.T, report ValidationReport, err error) {
				require.Len(t, report.CriticalErrors, 1)
				var (
					criticalError = report.CriticalErrors[0]
				)

				assert.Equal(t, "no meta tag found to match legacy data tag", criticalError.Message)
				assert.ErrorIs(t, criticalError.Error, ErrInvalidFileConfiguration)
				assert.ErrorIs(t, err, ErrInvalidFileConfiguration)
			},
		},
		{
			name:               "unsuccesful payload, no valid tags",
			payload:            `{}`,
			expectedParsedData: ParsedData{},
			errValidationFunc: func(t *testing.T, report ValidationReport, err error) {
				require.Len(t, report.CriticalErrors, 1)
				var (
					criticalError = report.CriticalErrors[0]
				)

				assert.Equal(t, "no tags found", criticalError.Message)
				assert.ErrorIs(t, criticalError.Error, ErrInvalidFileConfiguration)
				assert.ErrorIs(t, err, ErrInvalidFileConfiguration)
			},
		},
		{
			name:               "unsuccessful original payload, duplicate meta tag",
			payload:            `{"meta": {"methods": 0, "type": "sessions", "count": 0, "version": 5}, "meta": 0, "data": []}`,
			expectedParsedData: ParsedData{PayloadType: ingest.DataTypeSession, LegacyMetadata: ingest.LegacyMetadata{Type: ingest.DataTypeSession, Methods: 0, Version: 5}},
			errValidationFunc: func(t *testing.T, report ValidationReport, err error) {
				require.Len(t, report.CriticalErrors, 1)
				var (
					criticalError = report.CriticalErrors[0]
				)

				assert.Equal(t, "duplicate top level meta tag found", criticalError.Message)
				assert.ErrorIs(t, criticalError.Error, ErrInvalidFileConfiguration)
				assert.ErrorIs(t, err, ErrInvalidFileConfiguration)
			},
		},
		{
			name:               "unsuccessful original payload, invalid meta",
			payload:            `{"data": [],"meta": 0}`,
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
			payload:            `{"data": [],"meta": {"methods": 0, "type": "sessions", "count": 0, "version": 5}}`,
			expectedParsedData: ParsedData{PayloadType: ingest.DataTypeSession, LegacyMetadata: ingest.LegacyMetadata{Type: ingest.DataTypeSession, Methods: 0, Version: 5}},
			errValidationFunc: func(t *testing.T, report ValidationReport, err error) {
				assert.Equal(t, report, emptyValidationReport)
				assert.NoError(t, err)
			},
		},
		{
			name:               "unsuccessful original payload, invalid type",
			payload:            `{"data": [],"meta": {"methods": 0, "type": "invalid", "count": 0, "version": 5}}`,
			expectedParsedData: ParsedData{},
			errValidationFunc: func(t *testing.T, report ValidationReport, err error) {
				require.Len(t, report.CriticalErrors, 1)
				var (
					criticalError = report.CriticalErrors[0]
				)

				assert.Equal(t, "invalid legacy metadata data type", criticalError.Message)
				assert.ErrorIs(t, criticalError.Error, ErrInvalidDataType)
				assert.ErrorIs(t, err, ErrInvalidDataType)
			},
		},
		// Invalid payload tests
		{
			name:               "enforce mutual exclusivity",
			payload:            `{"data": [], "graph": {}}`,
			expectedParsedData: ParsedData{},
			errValidationFunc: func(t *testing.T, report ValidationReport, err error) {
				require.Len(t, report.CriticalErrors, 1)
				var (
					criticalError = report.CriticalErrors[0]
				)

				assert.Equal(t, "cannot have both legacy data tag and opengraph graph tag", criticalError.Message)
				assert.ErrorIs(t, criticalError.Error, ErrInvalidFileConfiguration)
				assert.ErrorIs(t, err, ErrInvalidFileConfiguration)
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
