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
	"time"

	"github.com/gorilla/mux"
	"github.com/specterops/bloodhound/src/api/middleware"
	"github.com/specterops/bloodhound/src/config"
	"github.com/stretchr/testify/require"
	"github.com/ulule/limiter/v3"
	"github.com/ulule/limiter/v3/drivers/store/memory"
)

func TestRateLimitMiddleware(t *testing.T) {
	allowedReqsPerSecond := 5
	rate := limiter.Rate{
		Period: 1 * time.Second,
		Limit:  int64(allowedReqsPerSecond),
	}

	store := memory.NewStore()

	instance := limiter.New(store, rate)

	cfg, err := config.NewDefaultConfiguration()
	require.Nil(t, err)

	testHandler := &CountingHandler{}
	router := mux.NewRouter()
	router.Use(middleware.RateLimitMiddleware(cfg, instance))
	router.Handle("/teapot", testHandler)

	if req, err := http.NewRequest("GET", "/teapot", nil); err != nil {
		t.Fatal(err)
	} else {
		req.Header.Set("X-Real-IP", "8.8.8.8")

		// simulate exceeding the limit as fast as possible
		for i := 0; i <= allowedReqsPerSecond; i++ {
			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)
		}
	}

	if testHandler.Count != allowedReqsPerSecond {
		t.Errorf("invalid HTTP 200 count: got %v want %v", testHandler.Count, 5)
	}
}

func TestDefaultRateLimitMiddleware(t *testing.T) {
	testHandler := &CountingHandler{}

	cfg, err := config.NewDefaultConfiguration()
	require.Nil(t, err)

	router := mux.NewRouter()
	router.Use(middleware.DefaultRateLimitMiddleware(cfg))
	router.Handle("/teapot", testHandler)

	if req, err := http.NewRequest("GET", "/teapot", nil); err != nil {
		t.Fatal(err)
	} else {
		req.Header.Set("X-Real-IP", "8.8.8.8")

		// simulate exceeding the limit as fast as possible
		for i := 0; i <= middleware.DefaultRateLimit; i++ {
			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)
		}
	}

	if testHandler.Count != middleware.DefaultRateLimit {
		t.Errorf("invalid HTTP 200 count: got %v want %v", testHandler.Count, middleware.DefaultRateLimit)
	}
}

type CountingHandler struct {
	Count int
}

func (s *CountingHandler) ServeHTTP(response http.ResponseWriter, _ *http.Request) {
	s.Count++
	response.Write([]byte("I'm a little teapot"))
}
