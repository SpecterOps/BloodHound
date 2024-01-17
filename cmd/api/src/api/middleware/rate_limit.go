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
	"time"

	"github.com/specterops/bloodhound/headers"

	"github.com/specterops/bloodhound/src/api"
	"github.com/didip/tollbooth/v6"
	"github.com/didip/tollbooth/v6/limiter"
	"github.com/gorilla/mux"
)

// DefaultRateLimit is the default number of allowed requests per second
const DefaultRateLimit = 55

// RateLimitHandler returns a http.Handler that limits the rate of requests
// for a given handler
//
// Usage:
//
//	limiter := tollbooth.NewLimiter(1, nil).
//		SetMethods([]string{"GET", "POST"}).
//		...configure tollbooth Limiter ...
//
//	router.Handle("/teapot", RateLimitHandler(limiter, handler))
func RateLimitHandler(limiter *limiter.Limiter, handler http.Handler) http.Handler {
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {

		ctx := request.Context()
		select {
		case <-ctx.Done():
			api.WriteErrorResponse(request.Context(), &api.ErrorResponse{
				HTTPStatus: http.StatusBadRequest,
				Error:      "client closed the connection before the request could be completed",
			}, response)
		default:
			if err := tollbooth.LimitByRequest(limiter, response, request); err != nil {
				// In case SetOnLimitReached was called
				limiter.ExecOnLimitReached(response, request)

				api.WriteErrorResponse(request.Context(), &api.ErrorWrapper{
					HTTPStatus: err.StatusCode,
					Timestamp:  time.Now(),
					RequestID:  request.Header.Get(headers.RequestID.String()),
					Errors: []api.ErrorDetails{
						{
							Context: "middleware",
							Message: err.Error(),
						},
					},
				}, response)
			} else {
				handler.ServeHTTP(response, request)
			}

		}
	})
}

// RateLimitMiddleware is a convenience function for creating rate limiting middleware
//
// Usage:
//
//	limiter := tollbooth.NewLimiter(1, nil).
//		SetMethods([]string{"GET", "POST"}).
//		...configure tollbooth Limiter ...
//
//	router.Use(RateLimitMiddleware(limiter))
func RateLimitMiddleware(limiter *limiter.Limiter) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return RateLimitHandler(limiter, next)
	}
}

// DefaultRateLimitMiddleware is a convenience function for creating the default rate limiting middleware
// for a router/route
//
// Usage:
//
//	router.Use(DefaultRateLimitMiddleware())
func DefaultRateLimitMiddleware() mux.MiddlewareFunc {
	limiter := tollbooth.NewLimiter(DefaultRateLimit, nil)
	return RateLimitMiddleware(limiter)
}
