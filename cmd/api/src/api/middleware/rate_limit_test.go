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
	"github.com/specterops/bloodhound/cmd/api/src/database/mocks"
	"github.com/specterops/bloodhound/cmd/api/src/model/appcfg"
	"go.uber.org/mock/gomock"
)

func TestRateLimitMiddleware(t *testing.T) {
	t.Run("enforces configured limit when enabled", func(t *testing.T) {
		const allowedReqsPerSecond = 5

		router, testHandler := newRateLimitedRouter(t, config.Configuration{}, false, int64(allowedReqsPerSecond))
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

	t.Run("does not throttle when disabled (custom limit)", func(t *testing.T) {
		const requestCount = 3

		router, testHandler := newRateLimitedRouter(t, config.Configuration{DisableRateLimiting: true}, false, 1)
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
		router, testHandler := newRateLimitedRouter(t, config.Configuration{}, true, 0)
		req := newTestRequest(t)

		for i := 0; i <= middleware.DefaultRateLimit; i++ {
			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			if i == middleware.DefaultRateLimit && rr.Code != http.StatusTooManyRequests {
				t.Fatalf("invalid final response code: got %d want %d", rr.Code, http.StatusTooManyRequests)
			}
		}

		if testHandler.Count != middleware.DefaultRateLimit {
			t.Errorf("invalid HTTP 200 count: got %v want %v", testHandler.Count, middleware.DefaultRateLimit)
		}
	})

	t.Run("does not throttle requests when disabled", func(t *testing.T) {
		const additionalRequestCount = 1

		router, testHandler := newRateLimitedRouter(t, config.Configuration{DisableRateLimiting: true}, true, 0)
		req := newTestRequest(t)

		for i := 0; i <= middleware.DefaultRateLimit+additionalRequestCount; i++ {
			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			if rr.Code != http.StatusOK {
				t.Fatalf("invalid response code for request %d: got %d want %d", i+1, rr.Code, http.StatusOK)
			}
		}

		expectedRequestCount := middleware.DefaultRateLimit + additionalRequestCount + 1
		if testHandler.Count != expectedRequestCount {
			t.Errorf("invalid HTTP 200 count: got %v want %v", testHandler.Count, expectedRequestCount)
		}
	})
}

func newRateLimitedRouter(t *testing.T, cfg config.Configuration, useDefaultRateLimit bool, limit int64) (*mux.Router, *CountingHandler) {
	t.Helper()

	mockCtl := gomock.NewController(t)
	mockDatabase := mocks.NewMockDatabase(mockCtl)

	if !cfg.DisableRateLimiting {
		mockDatabase.EXPECT().GetConfigurationParameter(gomock.Any(), appcfg.TrustedProxiesConfig).Return(appcfg.Parameter{}, nil).AnyTimes()
	}

	testHandler := &CountingHandler{}
	router := mux.NewRouter()

	if useDefaultRateLimit {
		router.Use(middleware.DefaultRateLimitMiddleware(cfg, mockDatabase))
	} else {
		router.Use(middleware.RateLimitMiddleware(cfg, mockDatabase, limit))
	}

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
