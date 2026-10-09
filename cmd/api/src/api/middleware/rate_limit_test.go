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

package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/specterops/bloodhound/cmd/api/src/api/middleware"
	"github.com/specterops/bloodhound/cmd/api/src/api/registration"
	"github.com/specterops/bloodhound/cmd/api/src/api/router"
	"github.com/specterops/bloodhound/cmd/api/src/auth"
	"github.com/specterops/bloodhound/cmd/api/src/config"
	"github.com/specterops/bloodhound/cmd/api/src/database/mocks"
	"github.com/specterops/bloodhound/cmd/api/src/model/appcfg"
	"go.uber.org/mock/gomock"
)

func TestRateLimitMiddleware(t *testing.T) {
	t.Parallel()

	var testCases = []struct {
		name       string
		limit      int64
		useDefault bool
	}{
		{name: "Success: custom limit admits configured requests - 200", limit: 5},
		{name: "Success: default limit admits configured requests - 200", limit: middleware.DefaultRateLimit, useDefault: true},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			mockController := gomock.NewController(t)
			mockDatabase := mocks.NewMockDatabase(mockController)
			mockDatabase.EXPECT().GetConfigurationParameter(gomock.Any(), appcfg.TrustedProxiesConfig).Return(appcfg.Parameter{}, nil).AnyTimes()

			var (
				testHandler = &CountingHandler{}
				testRouter  = mux.NewRouter()
				rateLimiter mux.MiddlewareFunc
			)
			if testCase.useDefault {
				rateLimiter = middleware.DefaultRateLimitMiddleware(mockDatabase)
			} else {
				rateLimiter = middleware.RateLimitMiddleware(mockDatabase, testCase.limit)
			}
			testRouter.Use(rateLimiter)
			testRouter.Handle("/teapot", testHandler)

			request := httptest.NewRequest(http.MethodGet, "/teapot", nil)
			for requestNumber := int64(0); requestNumber <= testCase.limit; requestNumber++ {
				testRouter.ServeHTTP(httptest.NewRecorder(), request)
			}

			if int64(testHandler.Count) != testCase.limit {
				t.Errorf("invalid HTTP 200 count: got %d want %d", testHandler.Count, testCase.limit)
			}
		})
	}
}

func TestMatchedRouteRateLimitMiddleware(t *testing.T) {
	t.Parallel()

	var testCases = []struct {
		name                 string
		path                 string
		secondPath           string
		excludedPathPrefixes []string
		limitsByPath         map[string]int64
		requestCount         int
		wantAllowedRequests  int
		secondRequestCount   int
		secondAllowed        int
		wantDatabaseLookups  int
	}{
		{
			name:                "Error: default route rejects request over limit - 429",
			path:                "/clients",
			secondPath:          "/other",
			limitsByPath:        map[string]int64{"/clients": 2, "/other": 2},
			requestCount:        3,
			wantAllowedRequests: 2,
			secondRequestCount:  3,
			secondAllowed:       2,
			wantDatabaseLookups: 6,
		},
		{
			name:                "Error: login override rejects second request - 429",
			path:                "/api/v2/login",
			limitsByPath:        map[string]int64{"/api/v2/login": 1},
			requestCount:        2,
			wantAllowedRequests: 1,
			wantDatabaseLookups: 2,
		},
		{
			name:                "Success: BHCE does not hardcode BHE support login policy - 204",
			path:                "/api/v2/login/support",
			limitsByPath:        map[string]int64{"/api/v2/login": 1},
			requestCount:        2,
			wantAllowedRequests: 2,
			wantDatabaseLookups: 2,
		},
		{
			name:                "Error: caller supplied BHE support override rejects excess request - 429",
			path:                "/api/v2/login/support",
			limitsByPath:        map[string]int64{"/api/v2/login/support": 1},
			requestCount:        2,
			wantAllowedRequests: 1,
			wantDatabaseLookups: 2,
		},
		{
			name:                 "Success: UI asset prefix is exempt - 204",
			path:                 "/ui/assets/app.js",
			excludedPathPrefixes: []string{"/ui"},
			requestCount:         3,
			wantAllowedRequests:  3,
			wantDatabaseLookups:  0,
		},
		{
			name:                 "Success: BHE remediation asset prefix is exempt - 204",
			path:                 "/api/v2/assets/remediation/image.png",
			excludedPathPrefixes: []string{"/api/v2/assets"},
			requestCount:         3,
			wantAllowedRequests:  3,
		},
		{
			name:                 "Error: exemption respects path boundary - 429",
			path:                 "/ui2/assets/app.js",
			excludedPathPrefixes: []string{"/ui"},
			limitsByPath:         map[string]int64{"/ui2/assets/app.js": 2},
			requestCount:         3,
			wantAllowedRequests:  2,
			wantDatabaseLookups:  3,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			mockController := gomock.NewController(t)
			mockDatabase := mocks.NewMockDatabase(mockController)
			mockDatabase.EXPECT().GetConfigurationParameter(gomock.Any(), appcfg.TrustedProxiesConfig).Return(appcfg.Parameter{}, nil).Times(testCase.wantDatabaseLookups)

			var (
				authCalls    int
				handlerCalls int
				routerInst   = router.NewRouter(config.Configuration{}, auth.NewAuthorizer(nil), "")
			)
			routerInst.UsePanicRecovery(middleware.PanicHandler)
			if err := routerInst.EnsureMatchedRouteRateLimit(func() mux.MiddlewareFunc {
				return middleware.MatchedRouteRateLimitMiddleware(mockDatabase, testCase.excludedPathPrefixes, testCase.limitsByPath)
			}); err != nil {
				t.Fatal(err)
			}
			routerInst.UseAuthenticationMiddleware(func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
					authCalls++
					next.ServeHTTP(response, request)
				})
			})
			routerInst.HandleFunc(testCase.path, func(response http.ResponseWriter, _ *http.Request) {
				handlerCalls++
				response.WriteHeader(http.StatusNoContent)
			})
			if testCase.secondPath != "" {
				routerInst.HandleFunc(testCase.secondPath, func(response http.ResponseWriter, _ *http.Request) {
					handlerCalls++
					response.WriteHeader(http.StatusNoContent)
				})
			}

			for requestNumber := 0; requestNumber < testCase.requestCount; requestNumber++ {
				response := httptest.NewRecorder()
				routerInst.Handler().ServeHTTP(response, httptest.NewRequest(http.MethodGet, testCase.path, nil))
				if requestNumber < testCase.wantAllowedRequests && response.Code != http.StatusNoContent {
					t.Fatalf("request %d returned %d; want %d", requestNumber+1, response.Code, http.StatusNoContent)
				}
				if requestNumber >= testCase.wantAllowedRequests && response.Code != http.StatusTooManyRequests {
					t.Fatalf("request %d returned %d; want %d", requestNumber+1, response.Code, http.StatusTooManyRequests)
				}
			}
			for requestNumber := 0; requestNumber < testCase.secondRequestCount; requestNumber++ {
				response := httptest.NewRecorder()
				routerInst.Handler().ServeHTTP(response, httptest.NewRequest(http.MethodGet, testCase.secondPath, nil))
				if requestNumber < testCase.secondAllowed && response.Code != http.StatusNoContent {
					t.Fatalf("second route request %d returned %d; want %d", requestNumber+1, response.Code, http.StatusNoContent)
				}
				if requestNumber >= testCase.secondAllowed && response.Code != http.StatusTooManyRequests {
					t.Fatalf("second route request %d returned %d; want %d", requestNumber+1, response.Code, http.StatusTooManyRequests)
				}
			}

			wantAdmittedRequests := testCase.wantAllowedRequests + testCase.secondAllowed
			if authCalls != wantAdmittedRequests {
				t.Fatalf("authentication ran %d times; want %d admitted requests", authCalls, wantAdmittedRequests)
			}
			if handlerCalls != wantAdmittedRequests {
				t.Fatalf("handler ran %d times; want %d admitted requests", handlerCalls, wantAdmittedRequests)
			}
		})
	}
}

func TestMatchedRouteRateLimitMiddlewareRunsBetweenPanicRecoveryAndAuthentication(t *testing.T) {
	mockController := gomock.NewController(t)
	mockDatabase := mocks.NewMockDatabase(mockController)
	mockDatabase.EXPECT().GetConfigurationParameter(gomock.Any(), gomock.Any()).Return(appcfg.Parameter{}, nil).AnyTimes()

	var routerInst = router.NewRouter(config.Configuration{}, auth.NewAuthorizer(nil), "")
	registration.RegisterFossGlobalMiddleware(&routerInst, config.Configuration{}, nil, nil, mockDatabase)
	if err := routerInst.EnsureMatchedRouteRateLimit(func() mux.MiddlewareFunc {
		return middleware.MatchedRouteRateLimitMiddleware(mockDatabase, nil, map[string]int64{"/limited": 1})
	}); err != nil {
		t.Fatal(err)
	}
	routerInst.HandleFunc("/limited", func(response http.ResponseWriter, _ *http.Request) {
		response.WriteHeader(http.StatusNoContent)
	})

	for requestNumber, expectedStatus := range []int{http.StatusBadRequest, http.StatusTooManyRequests} {
		request := httptest.NewRequest(http.MethodGet, "/limited", nil)
		request.Header.Set("Authorization", "invalid")
		response := httptest.NewRecorder()
		routerInst.Handler().ServeHTTP(response, request)
		if response.Code != expectedStatus {
			t.Fatalf("request %d returned %d; want %d", requestNumber+1, response.Code, expectedStatus)
		}
	}

	routerInst.UsePostroutingBeforeAuthentication(func(http.Handler) http.Handler {
		return http.HandlerFunc(func(http.ResponseWriter, *http.Request) { panic("limiter downstream panic") })
	})
	routerInst.HandleFunc("/panics", func(http.ResponseWriter, *http.Request) {})
	response := httptest.NewRecorder()
	routerInst.Handler().ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/panics", nil))
	if response.Code != http.StatusOK {
		t.Fatalf("recovered panic returned %d; want the default response status %d", response.Code, http.StatusOK)
	}
}

type CountingHandler struct {
	Count int
}

func (s *CountingHandler) ServeHTTP(response http.ResponseWriter, _ *http.Request) {
	s.Count++
	response.Write([]byte("I'm a little teapot"))
}
