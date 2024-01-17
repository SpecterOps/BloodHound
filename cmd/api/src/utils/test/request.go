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

package test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/specterops/bloodhound/src/ctx"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"
	"github.com/specterops/bloodhound/headers"
	"github.com/specterops/bloodhound/mediatypes"
)

type RequestExecutor struct {
	request *http.Request
	handler http.Handler
	context context.Context
	t       *testing.T
}

func (s *RequestExecutor) Request() *http.Request {
	return s.request
}

func (s *RequestExecutor) WithContext(bhCtx *ctx.Context) *RequestExecutor {
	s.request = ctx.SetRequestContext(s.request, bhCtx)
	return s
}

func (s *RequestExecutor) WithMethod(method string) *RequestExecutor {
	s.request.Method = method
	return s
}

func (s *RequestExecutor) WithURL(rawURL string, args ...any) *RequestExecutor {
	formattedRawURL := rawURL

	if len(args) > 0 {
		formattedRawURL = fmt.Sprintf(rawURL, args...)
	}

	if parsedURL, err := url.Parse(formattedRawURL); err != nil {
		s.t.Fatalf("Failed parsing URL: %s", rawURL)
	} else {
		s.request.URL = parsedURL
	}

	return s
}

func (s *RequestExecutor) WithURLQueryVars(values url.Values) *RequestExecutor {
	s.request.URL.RawQuery = values.Encode()
	return s
}

func (s *RequestExecutor) WithHeader(key, value string) *RequestExecutor {
	s.request.Header.Set(key, value)
	return s
}

func (s *RequestExecutor) WithBody(body any) *RequestExecutor {
	switch typedBody := body.(type) {
	case io.ReadCloser:
		s.request.Body = typedBody

	case io.Reader:
		s.request.Body = io.NopCloser(typedBody)

	default:
		if content, err := json.Marshal(body); err != nil {
			panic(fmt.Sprintf("failed marshalling body to JSON: %v", err))
		} else {
			s.request.Body = io.NopCloser(bytes.NewBuffer(content))
			s.request.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
		}
	}

	return s
}

func (s *RequestExecutor) WithURLPathVars(vars map[string]string) *RequestExecutor {
	s.request = mux.SetURLVars(s.request, vars)
	return s
}

func (s *RequestExecutor) OnHandler(handler http.Handler) *RequestExecutor {
	s.handler = handler
	return s
}

func (s *RequestExecutor) OnHandlerFunc(handler http.HandlerFunc) *RequestExecutor {
	return s.OnHandler(handler)
}

func (s *RequestExecutor) Require() RequestResponseAssertions {
	defer func() {
		if recovery := recover(); recovery != nil {
			s.t.Fatalf("Panic during request execution against the handler: %v", recovery)
		}
	}()

	responseRecorder := httptest.NewRecorder()
	s.handler.ServeHTTP(responseRecorder, s.request)

	return RequestResponseAssertions{
		request:  s.request,
		response: responseRecorder,
		t:        s.t,
	}
}

func Request(t *testing.T) *RequestExecutor {
	var (
		requestContext = context.WithValue(context.Background(), ctx.ValueKey, &ctx.Context{})
		request        = &http.Request{
			Header:     make(http.Header),
			Proto:      "HTTP/1.1",
			ProtoMajor: 1,
			ProtoMinor: 1,
		}
	)

	return &RequestExecutor{
		request: request.WithContext(requestContext),
		context: requestContext,
		t:       t,
	}
}

type RequestResponseAssertions struct {
	request  *http.Request
	response *httptest.ResponseRecorder
	t        *testing.T
}

func (s RequestResponseAssertions) ResponseStatusCode(status int) RequestResponseAssertions {
	require.Equal(s.t, status, s.response.Code, "Body: %s", s.response.Body.String())
	return s
}

func (s RequestResponseAssertions) ResponseJSONBody(body any) RequestResponseAssertions {
	require.Equal(s.t, mediatypes.ApplicationJson.String(), s.response.Header().Get(headers.ContentType.String()))

	content, err := json.Marshal(body)
	if err != nil {
		s.t.Fatalf("Failed marshaling expected body: %v", err)
	}

	var contentAsInterface any = content
	if v1Expected, ok := contentAsInterface.(map[string]any); ok {
		var v1Actual map[string]any

		if err := json.Unmarshal(s.response.Body.Bytes(), &v1Actual); err != nil {
			s.t.Fatalf("Failed unmarshaling response body: %v", err)
		}

		require.Equal(s.t, v1Expected, v1Actual)
	}

	if v2Expected, ok := contentAsInterface.(string); ok {
		var v2Actual string

		if err := json.Unmarshal(s.response.Body.Bytes(), &v2Actual); err != nil {
			s.t.Fatalf("Failed unmarshaling response body: %v", err)
		}

		require.Equal(s.t, v2Expected, v2Actual)
	}

	return s
}
