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
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/specterops/bloodhound/src/database"
	"github.com/specterops/bloodhound/src/model/appcfg"
	"github.com/ulule/limiter/v3"
	"github.com/ulule/limiter/v3/drivers/middleware/stdlib"
	"github.com/ulule/limiter/v3/drivers/store/memory"
)

// DefaultRateLimit is the default number of allowed requests per second
const DefaultRateLimit = 55

func rateLimitMiddleware(db database.Database, limiter *limiter.Limiter) mux.MiddlewareFunc {
	ipGetter := stdlib.WithKeyGetter(
		func(r *http.Request) string {
			var remoteIP string
			trustedProxies := appcfg.GetTrustedProxiesParameters(r.Context(), db)

			if host, _, err := net.SplitHostPort(r.RemoteAddr); err != nil {
				slog.WarnContext(r.Context(), fmt.Sprintf("Error parsing remoteAddress '%s': %s", r.RemoteAddr, err))
				remoteIP = r.RemoteAddr
			} else {
				remoteIP = host
			}

			if trustedProxies <= 0 {
				slog.DebugContext(r.Context(), "Using direct remote IP Address for rate limiting", "IP Address", remoteIP)
				return remoteIP
			} else if xff := r.Header.Get("X-Forwarded-For"); xff == "" {
				slog.DebugContext(r.Context(), "Expected X-Forwarded-For header for rate limiting but none found. Defaulted to remote IP Address", "IP Address", remoteIP)
				return remoteIP
			} else {
				ips := strings.Split(xff, ",")

				idxIP := len(ips) - trustedProxies
				if idxIP < 0 {
					slog.WarnContext(r.Context(), "Not enough IPs in X-Forwarded-For, defaulting to first IP", "X-Forwarded-For", xff)
					idxIP = 0
				}

				finalIP := strings.TrimSpace(ips[idxIP])

				slog.DebugContext(r.Context(), "Found client IP Address for rate limiting in XFF", "IP Address", finalIP, "X-Forwarded-For", xff)
				return finalIP
			}
		},
	)

	middleware := stdlib.NewMiddleware(limiter, ipGetter)

	return func(next http.Handler) http.Handler {
		return middleware.Handler(removeRateLimitHeadersMiddleware(next))
	}
}

// RemoveRateLimitHeadersMiddleware removes rate limit headers that we do not want appearing in our responses
func removeRateLimitHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		response.Header().Del("X-RateLimit-Limit")
		response.Header().Del("X-RateLimit-Remaining")
		response.Header().Del("X-RateLimit-Reset")
		next.ServeHTTP(response, request)
	})
}

// DefaultRateLimitMiddleware is a convenience function for creating the default rate limiting middleware
// for a router/route
//
// Usage:
//
//	router.Use(DefaultRateLimitMiddleware(db))
func DefaultRateLimitMiddleware(db database.Database) mux.MiddlewareFunc {
	return RateLimitMiddleware(db, DefaultRateLimit)
}

// RateLimitMiddleware is a function for creating rate limiting middleware
// with a particular time limit for a router/route
//
// Usage:
//
//	router.Use(RateLimitMiddleware(db, 1))
func RateLimitMiddleware(db database.Database, limit int64) mux.MiddlewareFunc {
	rate := limiter.Rate{
		Period: 1 * time.Second,
		Limit:  limit,
	}

	store := memory.NewStore()

	instance := limiter.New(store, rate)

	return rateLimitMiddleware(db, instance)
}
