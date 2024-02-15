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

package datapipe_test

import (
	"github.com/specterops/bloodhound/src/daemons/datapipe"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

type metaTagAssertion struct {
	name         string
	rawString    string
	err          error
	expectedType datapipe.DataType
}

func Test_ValidateMetaTag(t *testing.T) {
	assertions := []metaTagAssertion{
		{
			name:         "valid",
			rawString:    `{"meta": {"methods": 0, "type": "sessions", "count": 0, "version": 5}, "data": []}`,
			err:          nil,
			expectedType: datapipe.DataTypeSession,
		},
		{
			name:      "No data tag",
			rawString: `{"meta": {"methods": 0, "type": "sessions", "count": 0, "version": 5}}`,
			err:       datapipe.ErrDataTagNotFound,
		},
		{
			name:      "No meta tag",
			rawString: `{"data": []}`,
			err:       datapipe.ErrMetaTagNotFound,
		},
		{
			name:      "No valid tags",
			rawString: `{}`,
			err:       datapipe.ErrNoTagFound,
		},
		{
			name:         "ignore invalid tag but still find correct tag",
			rawString:    `{"meta": 0, "meta": {"methods": 0, "type": "sessions", "count": 0, "version": 5}, "data": []}`,
			err:          nil,
			expectedType: datapipe.DataTypeSession,
		},
		{
			name:         "swapped order",
			rawString:    `{"data": [],"meta": {"methods": 0, "type": "sessions", "count": 0, "version": 5}}`,
			err:          nil,
			expectedType: datapipe.DataTypeSession,
		},
		{
			name:      "invalid type",
			rawString: `{"data": [],"meta": {"methods": 0, "type": "invalid", "count": 0, "version": 5}}`,
			err:       datapipe.ErrMetaTagNotFound,
		},
	}

	for _, assertion := range assertions {
		meta, err := datapipe.ValidateMetaTag(strings.NewReader(assertion.rawString))
		assert.ErrorIs(t, err, assertion.err)
		if assertion.err == nil {
			assert.Equal(t, meta.Type, assertion.expectedType)
		}

	}
}
