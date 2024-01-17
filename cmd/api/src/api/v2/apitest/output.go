// Copyright 2023 Specter Ops, Inc.
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

package apitest

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

type Output struct {
	t        *testing.T
	response *httptest.ResponseRecorder
}

type OutputFunc func(output Output)

type apiResponseWrapper struct {
	Data json.RawMessage `json:"data"`
}

// StatusCode requires the response status code to match the given code
func StatusCode(output Output, code int) {
	require.Equal(output.t, code, output.response.Code)
}

// BodyContains requires the given string to exist anywhere in the response body
func BodyContains(output Output, message string) {
	require.Contains(output.t, output.response.Body.String(), message)
}

// BodyNotContains requires the given string not to exist anywhere in the response body
func BodyNotContains(output Output, message string) {
	require.NotContains(output.t, output.response.Body.String(), message)
}

func Equal(output Output, expected any, actual any) {
	require.Equal(output.t, expected, actual)
}

// UnmarshalBody requires the entire response body to unmarshal into the given struct pointer
func UnmarshalBody(output Output, modelPtr any) {
	err := json.Unmarshal(output.response.Body.Bytes(), modelPtr)
	require.Nil(output.t, err)
}

// UnmarshalData requires the response body data field to unmarshal into the given struct pointer
func UnmarshalData(output Output, modelPtr any) {
	res := apiResponseWrapper{}
	err := json.Unmarshal(output.response.Body.Bytes(), &res)
	require.Nil(output.t, err)
	err = json.Unmarshal(res.Data, modelPtr)
	require.Nil(output.t, err)
}
