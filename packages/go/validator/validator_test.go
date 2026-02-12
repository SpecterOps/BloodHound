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
			name:               "unsuccessful opengraph metadata",
			payload:            `{"metadata":{"source_kind": 1},"graph": {"nodes":[]}}`,
			expectedParsedData: ParsedData{},
			errValidationFunc: func(t *testing.T, report ValidationReport, err error) {
				require.Len(t, report.CriticalErrors, 1)
				var (
					criticalError = report.CriticalErrors[0]
					unmarshalErr  = &json.UnmarshalTypeError{}
				)

				assert.Equal(t, "failed to decode opengraph metadata", criticalError.Message)
				assert.ErrorAs(t, criticalError.Error, &unmarshalErr)
				assert.ErrorAs(t, err, &unmarshalErr)
			},
		},
		{
			name:               "unsuccessful opengraph metadata, invalid field",
			payload:            `{"metadata":{"random field": "hello"},"graph": {"nodes":[]}}`,
			expectedParsedData: ParsedData{PayloadType: ingest.DataTypeOpenGraph},
			errValidationFunc: func(t *testing.T, report ValidationReport, err error) {
				require.Len(t, report.CriticalErrors, 1)
				var (
					criticalError = report.CriticalErrors[0]
					unmarshalErr  = &json.UnmarshalTypeError{}
				)

				assert.Equal(t, "failed to decode opengraph metadata", criticalError.Message)
				assert.ErrorAs(t, criticalError.Error, &unmarshalErr)
				assert.ErrorAs(t, err, &unmarshalErr)
			},
		},
		// {
		// 	name:         "enforce mutual exclusivity",
		// 	rawString:    `{"data": [], "graph": {}}`,
		// 	err:          ingest.ErrMixedIngestFormat,
		// 	expectedType: ingest.DataTypeOpenGraph,
		// },
		// {
		// 	name:         "valid",
		// 	rawString:    `{"meta": {"methods": 0, "type": "sessions", "count": 0, "version": 5}, "data": []}`,
		// 	err:          nil,
		// 	expectedType: ingest.DataTypeSession,
		// },
		// {
		// 	name:      "No data tag",
		// 	rawString: `{"meta": {"methods": 0, "type": "sessions", "count": 0, "version": 5}}`,
		// 	err:       ingest.ErrDataTagNotFound,
		// },
		// {
		// 	name:      "No meta tag",
		// 	rawString: `{"data": []}`,
		// 	err:       ingest.ErrMetaTagNotFound,
		// },
		// {
		// 	name:      "No valid tags",
		// 	rawString: `{}`,
		// 	err:       ingest.ErrNoTagFound,
		// },
		// {
		// 	name:         "ignore invalid tag but still find correct tag",
		// 	rawString:    `{"meta": 0, "meta": {"methods": 0, "type": "sessions", "count": 0, "version": 5}, "data": []}`,
		// 	err:          nil,
		// 	expectedType: ingest.DataTypeSession,
		// },
		// {
		// 	name:         "swapped order",
		// 	rawString:    `{"data": [],"meta": {"methods": 0, "type": "sessions", "count": 0, "version": 5}}`,
		// 	err:          nil,
		// 	expectedType: ingest.DataTypeSession,
		// },
		// {
		// 	name:      "invalid type",
		// 	rawString: `{"data": [],"meta": {"methods": 0, "type": "invalid", "count": 0, "version": 5}}`,
		// 	err:       ingest.ErrMetaTagNotFound,
		// },
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
