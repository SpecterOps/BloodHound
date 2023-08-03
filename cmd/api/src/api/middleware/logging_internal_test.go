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

	"github.com/specterops/bloodhound/src/test/must"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"github.com/specterops/bloodhound/headers"
	"github.com/specterops/bloodhound/log/mocks"
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

func Test_fetchSignedRequestFields(t *testing.T) {
	var (
		mockCtrl     = gomock.NewController(t)
		mockLogEvent = mocks.NewMockEvent(mockCtrl)

		expectedTime = time.Now()
		expectedID   = must.NewUUIDv4()
		request      = must.NewHTTPRequest(http.MethodGet, "http://example.com/", nil)
	)

	request.Header.Set(headers.Authorization.String(), "bhesignature "+expectedID.String())
	request.Header.Set(headers.RequestDate.String(), expectedTime.Format(time.RFC3339Nano))

	mockLogEvent.EXPECT().Str("signed_request_date", expectedTime.Format(time.RFC3339Nano)).Times(1)
	mockLogEvent.EXPECT().Str("token_id", expectedID.String()).Times(1)

	setSignedRequestFields(request, mockLogEvent)

	// Remove auth header since it is non-fatal if it is missing
	request.Header.Del(headers.Authorization.String())

	mockLogEvent.EXPECT().Str("signed_request_date", expectedTime.Format(time.RFC3339Nano)).Times(1)

	setSignedRequestFields(request, mockLogEvent)
}
