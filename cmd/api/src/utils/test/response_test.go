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

package test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestModifyCookieAttribute(t *testing.T) {
	t.Parallel()
	type args struct {
		headers http.Header
		attrKey string
		value   string
	}
	tests := []struct {
		name string
		args args
		want http.Header
	}{
		{
			name: "No Cookies, unchanged",
			args: args{
				headers: http.Header{
					"Should-Remain-Unchanged": []string{
						"test",
					},
				},
				attrKey: "sessionid",
				value:   "changedvalue",
			},
			want: http.Header{
				"Should-Remain-Unchanged": []string{
					"test",
				},
			},
		},
		{
			name: "Cookie attribute is not found, unchanged",
			args: args{
				headers: http.Header{
					"Set-Cookie": []string{
						"sessionid=abc123; Path=/; Secure; HttpOnly",
					},
				},
				attrKey: "notpresent",
				value:   "changedvalue",
			},
			want: http.Header{
				"Set-Cookie": []string{
					"sessionid=abc123; Path=/; Secure; HttpOnly",
				},
			},
		},
		{
			name: "Cookie attribute is found, changed",
			args: args{
				headers: http.Header{
					"Set-Cookie": []string{
						"sessionid=abc123; Path=/; Secure; HttpOnly",
					},
				},
				attrKey: "sessionid",
				value:   "changedvalue",
			},
			want: http.Header{
				"Set-Cookie": []string{
					"sessionid=changedvalue; Path=/; Secure; HttpOnly",
				},
			},
		},
		{
			name: "Cookie attribute is found (last in cookie string - no trailing semicolon), changed",
			args: args{
				headers: http.Header{
					"Set-Cookie": []string{
						"sessionid=abc123; Path=/; SameSite=Lax",
					},
				},
				attrKey: "SameSite",
				value:   "Strict",
			},
			want: http.Header{
				"Set-Cookie": []string{
					"sessionid=abc123; Path=/; SameSite=Strict",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := ModifyCookieAttribute(tt.args.headers, tt.args.attrKey, tt.args.value)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestOverwriteQueryParamIfHeaderAndParamExist(t *testing.T) {
	t.Parallel()
	type args struct {
		headers    http.Header
		headerKey  string
		paramKey   string
		paramValue string
	}
	tests := []struct {
		name string
		args args
		want http.Header
	}{
		{
			name: "headers do not exist, no change",
			args: args{
				headers: http.Header{},
			},
			want: http.Header{},
		},
		{
			name: "headers exists, param does not, no change",
			args: args{
				headers: http.Header{
					"Location": []string{"xyz"},
				},
				headerKey:  "Location",
				paramKey:   "key",
				paramValue: "value",
			},
			want: http.Header{
				"Location": []string{"xyz"},
			},
		},
		{
			name: "headers exists, invalid, no change",
			args: args{
				headers: http.Header{
					"Location": []string{"xyz"},
				},
				headerKey:  "Location",
				paramKey:   "key",
				paramValue: "value",
			},
			want: http.Header{
				"Location": []string{"xyz"},
			},
		},
		{
			name: "headers exists, param exists, change updated",
			args: args{
				headers: http.Header{
					"Location": []string{"?key=test"},
				},
				headerKey:  "Location",
				paramKey:   "key",
				paramValue: "value",
			},
			want: http.Header{
				"Location": []string{"?key=value"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := OverwriteQueryParamIfHeaderAndParamExist(tt.args.headers, tt.args.headerKey, tt.args.paramKey, tt.args.paramValue)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestSortJSONArrayElements(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected any
	}{
		{
			name:     "sorts an array of strings alphabetically",
			input:    `["b","a","c"]`,
			expected: []any{"a", "b", "c"},
		},
		{
			name:  "sorts nested arrays inside objects and arrays",
			input: `{"list":["b","a"],"obj":{"arr":[2,1]}}`,
			expected: map[string]any{
				"list": []any{"a", "b"},
				"obj": map[string]any{
					// json numbers are unmarshaled into float64 by default
					"arr": []any{float64(1), float64(2)},
				},
			},
		},
		{
			name:     "returns non-JSON input string unchanged",
			input:    "plain text",
			expected: "plain text",
		},
		{
			name:     "handles empty JSON object",
			input:    `{}`,
			expected: map[string]any{},
		},
		{
			name:     "handles empty JSON array",
			input:    `[]`,
			expected: []any{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := SortJSONArrayElements(t, tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSortJSONArrays(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected string // Expected JSON string after sorting arrays
	}{
		{
			name:     "sorts an array of strings alphabetically",
			input:    `["b", "a", "c"]`,
			expected: `["a", "b", "c"]`,
		},
		{
			name: "sorts nested arrays inside JSON objects",
			input: `{
				"list": ["b", "a", "c"],
				"nested": {
					"arr": [3, 2, 1]
				}
			}`,
			expected: `{
				"list": ["a", "b", "c"],
				"nested": {
					"arr": [1, 2, 3]
				}
			}`,
		},
		{
			name:     "sorts arrays of JSON objects by their JSON string representation",
			input:    `[{"id":2},{"id":1}]`,
			expected: `[{"id":1},{"id":2}]`,
		},
		{
			name:     "does not change JSON without arrays",
			input:    `{"a":1,"b":"test"}`,
			expected: `{"a":1,"b":"test"}`,
		},
		{
			name:     "handles empty array correctly",
			input:    `[]`,
			expected: `[]`,
		},
		{
			name:     "handles nested empty arrays correctly",
			input:    `{"a":[], "b":{"c":[]}}`,
			expected: `{"a":[], "b":{"c":[]}}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var inputData any
			err := json.Unmarshal([]byte(tt.input), &inputData)
			assert.NoError(t, err, "Failed to unmarshal input JSON")

			var expectedData any
			err = json.Unmarshal([]byte(tt.expected), &expectedData)
			assert.NoError(t, err, "Failed to unmarshal expected JSON")

			sortJSONArrays(inputData)

			assert.Equal(t, expectedData, inputData)
		})
	}
}
