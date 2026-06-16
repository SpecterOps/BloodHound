// Copyright 2026 Specter Ops, Inc.
//
// Licensed under the Apache License, Version 2.0
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0
package middleware_test

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/specterops/bloodhound/cmd/api/src/api/middleware"
	"github.com/specterops/bloodhound/cmd/api/src/auth"
	"github.com/specterops/bloodhound/cmd/api/src/ctx"
	"github.com/stretchr/testify/assert"
)

func TestLoggingMiddleware_QueryParameters(t *testing.T) {
	var (
		testURL1       = "/api/v2/bloodhound-users"
		testURL2       = "/api/v2/search"
		healthURL      = "/health"
		ssoCallbackURL = "/api/v2/sso/73d94be8-a18c-4de2-b63c-75c28d3d4bff/callback"
	)
	testCases := []struct {
		name              string
		url               string
		logContains       []string
		logDoesNotContain []string
	}{
		{
			name:              "non /api/v2 path does not log query parameters",
			url:               healthURL + "?randomparam=123",
			logContains:       []string{"HTTP request"},
			logDoesNotContain: []string{"query_parameters"},
		},
		{
			name:              "similar /api/v20 does not log query parameters",
			url:               "/api/v20/search?foo=bar",
			logContains:       []string{"HTTP request"},
			logDoesNotContain: []string{"query_parameters"},
		},
		{
			name:              "/api/v2 path with no query params does not log query parameters",
			url:               testURL1,
			logContains:       []string{"HTTP request"},
			logDoesNotContain: []string{"query_parameters"},
		},
		{
			name:              "SSO callback path with query params does not add query_parameters field and logs",
			url:               ssoCallbackURL + "?code=abc123&state=xyz789",
			logContains:       []string{"HTTP request"},
			logDoesNotContain: []string{`"query_parameters":"`, `"code":`, `"state":`},
		},
		{
			name:        "/api/v2 path with query params adds a query_parameters field",
			url:         testURL1 + "?first_name=eq:Hubert",
			logContains: []string{`"query_parameters":"`},
		},
		{
			name:        "single query parameter is logged",
			url:         testURL1 + "?first_name=eq:Hubert",
			logContains: []string{"HTTP request", "query_parameters", "first_name", "eq:Hubert"},
		},
		{
			name:        "multiple query parameters are logged",
			url:         testURL2 + "?q=abcd&limit=1000&type=Computer&type=User&type=Group",
			logContains: []string{"HTTP request", "query_parameters", "abcd", "limit", "1000", "type", "Computer", "User", "Group"},
		},
		{
			name:        "query_parameters field contains the full raw query string",
			url:         testURL2 + "?q=abcd&limit=1000&type=Computer&type=User&type=Group",
			logContains: []string{"HTTP request", `"query_parameters":"q=abcd&limit=1000&type=Computer&type=User&type=Group"`},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			previousLogger := slog.Default()
			t.Cleanup(func() { slog.SetDefault(previousLogger) })

			var logBuffer bytes.Buffer
			slog.SetDefault(slog.New(slog.NewJSONHandler(&logBuffer, &slog.HandlerOptions{})))

			nextHandler := http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
				responseWriter.WriteHeader(http.StatusOK)
			})
			handler := middleware.LoggingMiddleware(auth.NewIdentityResolver(), false)(nextHandler)

			request := httptest.NewRequest(http.MethodGet, testCase.url, nil)
			bhCtx := &ctx.Context{
				StartTime: time.Now(),
				RequestID: "123456",
			}
			request = ctx.SetRequestContext(request, bhCtx)
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, request)

			logOutput := logBuffer.String()
			for _, expected := range testCase.logContains {
				assert.Contains(t, logOutput, expected)
			}
			for _, unexpected := range testCase.logDoesNotContain {
				assert.NotContains(t, logOutput, unexpected)
			}
		})
	}
}
