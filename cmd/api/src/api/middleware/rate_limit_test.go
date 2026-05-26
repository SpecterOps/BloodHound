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
	"github.com/specterops/bloodhound/cmd/api/src/config"
	"github.com/specterops/bloodhound/cmd/api/src/database"
	"github.com/specterops/bloodhound/cmd/api/src/database/mocks"
	"github.com/specterops/bloodhound/cmd/api/src/model/appcfg"
	"go.uber.org/mock/gomock"
)

func TestRateLimitMiddleware(t *testing.T) {
	t.Run("enforces configured limit when enabled", func(t *testing.T) {
		const allowedReqsPerSecond = 5

		router, testHandler := newRateLimitedRouter(t, middleware.RateLimitMiddleware, int64(allowedReqsPerSecond))
		req := newTestRequest(t)

		for i := 0; i <= allowedReqsPerSecond; i++ {
			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			if i == allowedReqsPerSecond && rr.Code != http.StatusTooManyRequests {
				t.Fatalf("invalid final response code: got %d want %d", rr.Code, http.StatusTooManyRequests)
			}
		}

		if testHandler.Count != allowedReqsPerSecond {
			t.Errorf("invalid HTTP 200 count: got %v want %v", testHandler.Count, allowedReqsPerSecond)
		}
	})

	t.Run("does not throttle when limit is zero", func(t *testing.T) {
		const requestCount = 3

		router, testHandler := newRateLimitedRouter(t, middleware.RateLimitMiddleware, 0)
		req := newTestRequest(t)

		for i := 0; i < requestCount; i++ {
			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			if rr.Code != http.StatusOK {
				t.Fatalf("invalid response code for request %d: got %d want %d", i+1, rr.Code, http.StatusOK)
			}
		}

		if testHandler.Count != requestCount {
			t.Errorf("invalid HTTP 200 count: got %v want %v", testHandler.Count, requestCount)
		}
	})
}

func TestDefaultRateLimitMiddleware(t *testing.T) {
	t.Run("enforces default limit when enabled", func(t *testing.T) {
		cfg := config.Configuration{APIRateLimitRequestsPerSecond: config.DefaultAPIRateLimit}
		router, testHandler := newConfiguredRateLimitedRouter(t, cfg, middleware.DefaultRateLimitMiddleware, true)
		req := newTestRequest(t)

		for i := 0; i <= config.DefaultAPIRateLimit; i++ {
			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			if i == config.DefaultAPIRateLimit && rr.Code != http.StatusTooManyRequests {
				t.Fatalf("invalid final response code: got %d want %d", rr.Code, http.StatusTooManyRequests)
			}
		}

		if testHandler.Count != config.DefaultAPIRateLimit {
			t.Errorf("invalid HTTP 200 count: got %v want %v", testHandler.Count, config.DefaultAPIRateLimit)
		}
	})

	t.Run("does not throttle requests when API limit is zero", func(t *testing.T) {
		const additionalRequestCount = 1

		cfg := config.Configuration{APIRateLimitRequestsPerSecond: 0}
		router, testHandler := newConfiguredRateLimitedRouter(t, cfg, middleware.DefaultRateLimitMiddleware, false)
		req := newTestRequest(t)

		for i := 0; i <= config.DefaultAPIRateLimit+additionalRequestCount; i++ {
			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			if rr.Code != http.StatusOK {
				t.Fatalf("invalid response code for request %d: got %d want %d", i+1, rr.Code, http.StatusOK)
			}
		}

		expectedRequestCount := config.DefaultAPIRateLimit + additionalRequestCount + 1
		if testHandler.Count != expectedRequestCount {
			t.Errorf("invalid HTTP 200 count: got %v want %v", testHandler.Count, expectedRequestCount)
		}
	})
}

func TestLoginRateLimitMiddleware(t *testing.T) {
	t.Run("enforces login limit by default", func(t *testing.T) {
		router, testHandler := newConfiguredRateLimitedRouter(t, config.Configuration{}, middleware.LoginRateLimitMiddleware, true)
		req := newTestRequest(t)

		for i := 0; i <= middleware.LoginRateLimit; i++ {
			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			if i == middleware.LoginRateLimit && rr.Code != http.StatusTooManyRequests {
				t.Fatalf("invalid final response code: got %d want %d", rr.Code, http.StatusTooManyRequests)
			}
		}

		if testHandler.Count != middleware.LoginRateLimit {
			t.Errorf("invalid HTTP 200 count: got %v want %v", testHandler.Count, middleware.LoginRateLimit)
		}
	})

	t.Run("does not throttle when login protections are disabled", func(t *testing.T) {
		const requestCount = 3

		cfg := config.Configuration{DisableLoginProtections: true}
		router, testHandler := newConfiguredRateLimitedRouter(t, cfg, middleware.LoginRateLimitMiddleware, false)
		req := newTestRequest(t)

		for i := 0; i < requestCount; i++ {
			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			if rr.Code != http.StatusOK {
				t.Fatalf("invalid response code for request %d: got %d want %d", i+1, rr.Code, http.StatusOK)
			}
		}

		if testHandler.Count != requestCount {
			t.Errorf("invalid HTTP 200 count: got %v want %v", testHandler.Count, requestCount)
		}
	})
}

func newRateLimitedRouter(t *testing.T, middlewareFunc func(database.Database, int64) mux.MiddlewareFunc, limit int64) (*mux.Router, *CountingHandler) {
	t.Helper()

	mockCtl := gomock.NewController(t)
	mockDatabase := mocks.NewMockDatabase(mockCtl)

	if limit > 0 {
		mockDatabase.EXPECT().GetConfigurationParameter(gomock.Any(), appcfg.TrustedProxiesConfig).Return(appcfg.Parameter{}, nil).AnyTimes()
	}

	testHandler := &CountingHandler{}
	router := mux.NewRouter()

	router.Use(middlewareFunc(mockDatabase, limit))

	router.Handle("/teapot", testHandler)

	return router, testHandler
}

func newConfiguredRateLimitedRouter(
	t *testing.T,
	cfg config.Configuration,
	middlewareFunc func(config.Configuration, database.Database) mux.MiddlewareFunc,
	expectTrustedProxyLookup bool,
) (*mux.Router, *CountingHandler) {
	t.Helper()

	mockCtl := gomock.NewController(t)
	mockDatabase := mocks.NewMockDatabase(mockCtl)
	if expectTrustedProxyLookup {
		mockDatabase.EXPECT().GetConfigurationParameter(gomock.Any(), appcfg.TrustedProxiesConfig).Return(appcfg.Parameter{}, nil).AnyTimes()
	}

	testHandler := &CountingHandler{}
	router := mux.NewRouter()
	router.Use(middlewareFunc(cfg, mockDatabase))
	router.Handle("/teapot", testHandler)

	return router, testHandler
}

func newTestRequest(t *testing.T) *http.Request {
	t.Helper()

	req, err := http.NewRequest("GET", "/teapot", nil)
	if err != nil {
		t.Fatal(err)
	}

	return req
}

type CountingHandler struct {
	Count int
}

func (s *CountingHandler) ServeHTTP(response http.ResponseWriter, _ *http.Request) {
	s.Count++
	response.Write([]byte("I'm a little teapot"))
}
