// Copyright 2026 Specter Ops, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
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
	"net/http/httptest"
	"os"
	"os/exec"
	"testing"

	"github.com/specterops/bloodhound/packages/go/headers"
	"github.com/specterops/bloodhound/packages/go/mediatypes"
	"github.com/stretchr/testify/require"
)

const responseJSONBodyMismatchEnvironmentVariable = "BHE_TEST_RESPONSE_JSON_BODY_MISMATCH"

func TestNormalizeResponseJSONBody(t *testing.T) {
	const (
		errorResponseBody = `{"http_status":400,"timestamp":"2026-10-07T12:34:56Z","errors":[{"message":"invalid request"}]}`
		expectedBody      = `{"http_status":400,"timestamp":"0001-01-01T00:00:00Z","errors":[{"message":"invalid request"}]}`
	)

	actualBody := normalizeResponseJSONBody(errorResponseBody)

	require.JSONEq(t, expectedBody, actualBody)
}

func TestResponseJSONBody(t *testing.T) {
	const actualResponseBody = `{"timestamp":"2026-10-14T12:34:56Z","users":[{"id":1,"roles":["admin","user"]}]}`

	responseRecorder := httptest.NewRecorder()
	responseRecorder.Header().Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
	responseRecorder.Body.WriteString(actualResponseBody)

	assertions := RequestResponseAssertions{
		response: responseRecorder,
		t:        t,
	}

	assertions.ResponseJSONBody(map[string]any{
		"timestamp": "0001-01-01T00:00:00Z",
		"users": []any{map[string]any{
			"id":    1,
			"roles": []string{"admin", "user"},
		}},
	})
}

func TestResponseJSONBodyMismatchSubprocess(t *testing.T) {
	if os.Getenv(responseJSONBodyMismatchEnvironmentVariable) != "1" {
		return
	}

	const mismatchMarker = "running ResponseJSONBody mismatch assertion"
	t.Log(mismatchMarker)

	responseRecorder := httptest.NewRecorder()
	responseRecorder.Header().Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
	responseRecorder.Body.WriteString(`{"id":2}`)

	RequestResponseAssertions{
		response: responseRecorder,
		t:        t,
	}.ResponseJSONBody(map[string]int{"id": 1})
}

func TestResponseJSONBodyRejectsMismatch(t *testing.T) {
	command := exec.Command(os.Args[0], "-test.run=^TestResponseJSONBodyMismatchSubprocess$")
	command.Env = append(os.Environ(), responseJSONBodyMismatchEnvironmentVariable+"=1")

	commandOutput, err := command.CombinedOutput()
	require.Error(t, err, "expected the subprocess test to fail on mismatched response JSON")
	require.Contains(t, string(commandOutput), "running ResponseJSONBody mismatch assertion")
	require.Contains(t, string(commandOutput), "FAIL")
}
