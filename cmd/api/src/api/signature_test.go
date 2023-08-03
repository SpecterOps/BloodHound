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

package api_test

import (
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/specterops/bloodhound/src/api"
	"github.com/specterops/bloodhound/headers"

	"github.com/stretchr/testify/require"
)

func TestSignRequestAtInternalError(t *testing.T) {
	reader := strings.NewReader("Hello world")
	request, err := http.NewRequest("GET", "www.foo.bar", reader)
	require.Nil(t, err)

	now := time.Now()
	err = api.SignRequestAtTime("tokenID", "token", "unsupportedHMACType", now, request)
	require.Error(t, err)
	require.Equal(t, "unsupported HMAC method: unsupportedHMACType", err.Error())
}

func TestSignRequestSuccess(t *testing.T) {
	request, err := http.NewRequest("GET", "www.foo.bar", strings.NewReader("Hello world"))
	require.Nil(t, err)

	err = api.SignRequest("tokenID", "token", request)
	require.Nil(t, err)

	requestDate := request.Header.Get(headers.RequestDate.String())
	require.Equal(t, time.Now().Format(time.RFC3339), requestDate)

	authorization := request.Header[headers.Authorization.String()][0]
	require.Equal(t, "bhesignature tokenID", authorization)

	signature := request.Header[headers.Signature.String()][0]
	require.NotNil(t, signature)
}
