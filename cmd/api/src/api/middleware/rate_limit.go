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
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/ulule/limiter/v3"
	"github.com/ulule/limiter/v3/drivers/middleware/stdlib"
	"github.com/ulule/limiter/v3/drivers/store/memory"
)

// DefaultRateLimit is the default number of allowed requests per second
const DefaultRateLimit = 55

// RateLimitMiddleware is a convenience function for creating rate limiting middleware
//
//	router.Use(RateLimitMiddleware(limiter))
func RateLimitMiddleware(limiter *limiter.Limiter) mux.MiddlewareFunc {
	ipGetter := stdlib.WithKeyGetter(
		func(r *http.Request) string {
			if xff := r.Header.Get("X-Forwarded-For"); xff == "" {
				slog.DebugContext(r.Context(), "No data found in X-Forwarded-For header")
				return r.RemoteAddr
			} else {
				ips := strings.Split(xff, ",")
				rightMostIP := strings.TrimSpace(ips[len(ips)-1])

				return rightMostIP
			}
		},
	)

	middleware := stdlib.NewMiddleware(limiter, ipGetter)

	return middleware.Handler

}

// DefaultRateLimitMiddleware is a convenience function for creating the default rate limiting middleware
// for a router/route
//
// Usage:
//
//	router.Use(DefaultRateLimitMiddleware())
func DefaultRateLimitMiddleware() mux.MiddlewareFunc {
	rate := limiter.Rate{
		Period: 1 * time.Second,
		Limit:  DefaultRateLimit,
	}

	store := memory.NewStore()

	instance := limiter.New(store, rate, limiter.WithTrustForwardHeader(false))

	return RateLimitMiddleware(instance)
}
