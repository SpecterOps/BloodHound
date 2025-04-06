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
	"errors"
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
	name  string
	input string
	err   error
}

func Test_ValidateGenericIngest(t *testing.T) {

	positiveCases := []genericAssertion{
		{
			name: "payload contains realistic node",
			input: `{"graph": {"nodes": [
			{
			"id": "1234",
			"kinds": ["a","b"],
			"properties": {"thing_one": "thing_two","num": 1,"bool": true}
			 }
			]
			 }}`,
			err: nil,
		},
		{
			name:  "payload contains nodes only",
			input: `{"graph": {"nodes": [{"id": "1234"}]}}`,
			err:   nil,
		},
		{
			name:  "payload contains edges only",
			input: `{"graph": {"edges": [{"id": "1234"}]}}`,
			err:   nil,
		},
		{
			name:  "payload specifies edges before nodes",
			input: `{"graph": {"edges": [{"id": "1234"}], "nodes": [{"id": "1234"}]}}`,
			err:   nil,
		},
	}

	negativeCases := []genericAssertion{
		{
			name:  "payload doesn't contain nodes or edges",
			input: `{"graph": {}`,
			err:   ingest.ErrEmptyIngest,
		},
		{
			name:  "payload contains a node that doesn't conform to spec",
			input: `{"graph": {"nodes": [{"id": 1234}]}}`,
			err:   ingest.ErrNodeSchema,
		},
		{
			name:  "payload contains an edge that doesn't conform to spec",
			input: `{"graph": {"edges": [{"source_id": 1234}]}}`,
			err:   ingest.ErrEdgeSchema,
		},
		{
			name:  "payload contains a node that has invalid json",
			input: `{"graph": {"nodes": [{"id": "1234}]}}`,
			err:   errors.New("unexpected EOF"), // TODO
		},
	}

	for _, assertion := range positiveCases {
		err := ingest_service.ValidateGenericIngest(strings.NewReader(assertion.input), true)
		assert.Nil(t, err)
	}

	for _, assertion := range negativeCases {
		err := ingest_service.ValidateGenericIngest(strings.NewReader(assertion.input), true)
		assert.ErrorContains(t, err, assertion.err.Error())
	}
}

// func Test_hellojsonschema(t *testing.T) {
// 	err := ingest_service.ValidateNodeSchema()
// 	// var sb strings.Builder
// 	fmt.Println(err.Error())
// 	if err, ok := err.(*jsonschema.ValidationError); ok {
// 		for _, cause := range err.Causes {
// 			fmt.Println(cause.Error())
// 			fmt.Println(cause.InstanceLocation)
// 			// sb.WriteString(fmt.Sprintf("Field: %s - Error: %s\n", cause.InstanceLocation, cause.))
// 		}
// 	}
// 	assert.Nil(t, err)
// }
