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
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/specterops/bloodhound/src/api/middleware"
	"github.com/didip/tollbooth/v6"
	"github.com/gorilla/mux"
)

func TestRateLimitHandler(t *testing.T) {

	// Limit to 1 req/s
	limiter := tollbooth.NewLimiter(1, nil)

	count_429 := 0
	limiter.SetOnLimitReached(func(response http.ResponseWriter, request *http.Request) {
		count_429++
	})

	if req, err := http.NewRequest("GET", "/teapot", nil); err != nil {
		t.Fatal(err)
	} else {
		req.Header.Set("X-Real-IP", "8.8.8.8")

		handler := middleware.RateLimitHandler(limiter, &CountingHandler{})
		router := mux.NewRouter()
		router.Handle("/teapot", handler)

		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		// Expect a 200
		if status := rr.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
		}

		ch := make(chan int)
		go func() {
			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			// Expect a 429
			if status := rr.Code; status != http.StatusTooManyRequests {
				t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusTooManyRequests)
			}

			if count_429 == 0 {
				t.Error("OnLimitReached callback function should have been called")
			}

			close(ch)
		}()
		<-ch
	}
}

func TestRateLimitMiddleware(t *testing.T) {

	allowedReqsPerSecond := 5
	limiter := tollbooth.NewLimiter(float64(allowedReqsPerSecond), nil)

	count_429 := 0
	limiter.SetOnLimitReached(func(w http.ResponseWriter, r *http.Request) {
		count_429++
	})

	testHandler := &CountingHandler{}
	router := mux.NewRouter()
	router.Use(middleware.RateLimitMiddleware(limiter))
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

	if count_429 != 1 {
		t.Errorf("invalid HTTP 429 count: got %v want %v", count_429, 1)
	}
}

func TestDefaultRateLimitMiddleware(t *testing.T) {
	testHandler := &CountingHandler{}

	router := mux.NewRouter()
	router.Use(middleware.DefaultRateLimitMiddleware())
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

func TestDefaultRateLimitMiddlewareCanceledRequest(t *testing.T) {
	testHandler := &CountingHandler{}

	router := mux.NewRouter()
	router.Use(middleware.DefaultRateLimitMiddleware())
	router.Handle("/teapot", testHandler)

	ctx, cancel := context.WithCancel(context.Background())
	if req, err := http.NewRequestWithContext(ctx, "GET", "/teapot", nil); err != nil {
		cancel()
		t.Fatal(err)
	} else {
		req.Header.Set("X-Real-IP", "8.8.8.8")
		cancel()

		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Errorf("invalid HTTPStatus: got %v want %v", rr.Code, http.StatusBadRequest)
		}
	}
}

type CountingHandler struct {
	Count int
}

func (s *CountingHandler) ServeHTTP(response http.ResponseWriter, _ *http.Request) {
	s.Count++
	response.Write([]byte("I'm a little teapot"))
}
