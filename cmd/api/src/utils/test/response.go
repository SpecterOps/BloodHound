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
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"slices"
	"strings"
	"testing"

	"github.com/specterops/bloodhound/cmd/api/src/utils"
)

func ProcessResponse(t *testing.T, response *httptest.ResponseRecorder) (int, http.Header, string) {
	t.Helper()
	if response.Code != http.StatusOK && response.Code != http.StatusAccepted {
		responseBytes, err := utils.ReplaceFieldValueInJsonString(response.Body.String(), "timestamp", "0001-01-01T00:00:00Z")
		if err != nil {
			// not every error response contains a timestamp so print output and move along
			fmt.Printf("error replacing field value in json string: %v\n", err)
		}

		response.Body = bytes.NewBuffer([]byte(responseBytes))
	}

	if response.Body != nil {
		res, err := io.ReadAll(response.Body)
		if err != nil {
			t.Fatalf("error reading response body: %v", err)
		}

		return response.Code, response.Header(), string(res)
	}

	return response.Code, response.Header(), ""
}

// ModifyCookieAttribute searches the provided HTTP headers and updates the specified cookie
// attribute (e.g., "Secure", "SameSite", "Path") to a new value.
// If the attribute is found in a cookie string, its value is replaced with the provided value.
// Cookies that do not contain the specified attribute are left unchanged.
func ModifyCookieAttribute(headers http.Header, attrKey, value string) http.Header {
	cookies := headers["Set-Cookie"]
	if len(cookies) == 0 {
		// No cookies to modify, return unchanged.
		return headers
	}

	attrPrefix := attrKey + "="
	var newCookies []string
	modified := false

	for _, cookie := range cookies {
		start := strings.Index(cookie, attrPrefix)
		if start == -1 {
			// Attribute not found, keep original.
			newCookies = append(newCookies, cookie)
			continue
		}

		// Find end of the attribute (next semicolon or end of string)
		end := strings.Index(cookie[start:], ";")
		var newCookie string
		if end == -1 {
			// Attribute is last; replace till end
			newCookie = cookie[:start] + attrPrefix + value
		} else {
			end += start // Adjust to full string index
			newCookie = cookie[:start] + attrPrefix + value + cookie[end:]
		}

		newCookies = append(newCookies, newCookie)
		modified = true
	}

	if modified {
		headers["Set-Cookie"] = newCookies
	}

	return headers
}

// OverwriteQueryParamIfHeaderAndParamExist updates paramKey in the query string value
// of headerKey only if both the header and the parameter exist.
// Otherwise, it leaves the header untouched.
func OverwriteQueryParamIfHeaderAndParamExist(headers http.Header, headerKey, paramKey, paramValue string) http.Header {
	// Check if header exists and has at least one value
	vals := headers.Values(headerKey)
	if len(vals) == 0 {
		return headers // header missing, no change
	}

	// Parse the first header value as query string (remove leading "?")
	q, err := url.ParseQuery(strings.TrimPrefix(vals[0], "?"))
	if err != nil {
		return headers // parse error, no change
	}

	// Check if paramKey exists in the query parameters
	if _, exists := q[paramKey]; !exists {
		return headers // param missing, no change
	}

	// Param exists — overwrite its value
	q.Set(paramKey, paramValue)

	// Rebuild query string, preserve leading "?"
	headers.Set(headerKey, "?"+q.Encode())
	return headers
}

// SortJSONArrayElements parses JSON and recursively sorts all arrays so their order doesn't affect equality.
// Returns the raw string if the input is not valid JSON.
//
// Use this to normalize JSON bodies in tests where slice order is nondeterministic.
func SortJSONArrayElements(t *testing.T, raw string) any {
	var data any
	if err := json.Unmarshal([]byte(raw), &data); err != nil {
		return raw
	}

	sortJSONArrays(data)
	return data
}

func sortJSONArrays(v any) {
	switch val := v.(type) {
	case []any:
		for _, elem := range val {
			sortJSONArrays(elem) // Recursively sort nested arrays
		}
		slices.SortStableFunc(val, func(a, b any) int {
			ba, err1 := json.Marshal(a)
			bb, err2 := json.Marshal(b)
			if err1 != nil || err2 != nil {
				// Fallback to string representation if marshaling fails
				sa := fmt.Sprintf("%#v", a)
				sb := fmt.Sprintf("%#v", b)

				// Check if sa comes before sb in order
				if sa < sb {
					return -1
					// Check if sa comes after sb in order
				} else if sa > sb {
					return 1
				}
				// sa and sb are equal
				return 0
			}

			// Compare JSON-encoded strings in order
			sa, sb := string(ba), string(bb)
			if sa < sb {
				return -1
			} else if sa > sb {
				return 1
			}
			return 0
		})
	case map[string]any:
		for _, vv := range val {
			sortJSONArrays(vv) // Recursively sort nested arrays
		}
	}
}
