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

package middleware

import (
	"net/http"
	"testing"
	"time"

	"github.com/specterops/bloodhound/cmd/api/src/test/must"
	"github.com/specterops/bloodhound/packages/go/headers"
	"github.com/stretchr/testify/require"
)

func Test_signedRequestDate(t *testing.T) {
	var (
		expectedTime = time.Now()
		expectedID   = must.NewUUIDv4()
		request      = must.NewHTTPRequest(http.MethodGet, "http://example.com/", nil)
	)

	request.Header.Set(headers.Authorization.String(), "bhesignature "+expectedID.String())
	request.Header.Set(headers.RequestDate.String(), expectedTime.Format(time.RFC3339Nano))

	requestDate, hasHeader := getSignedRequestDate(request)

	require.True(t, hasHeader)
	require.Equal(t, expectedTime.Format(time.RFC3339Nano), requestDate)
}
