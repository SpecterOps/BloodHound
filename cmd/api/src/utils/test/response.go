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
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/specterops/bloodhound/src/utils"
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
