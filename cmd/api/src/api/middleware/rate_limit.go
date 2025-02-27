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
	"github.com/specterops/bloodhound/src/config"
	"github.com/ulule/limiter/v3"
	"github.com/ulule/limiter/v3/drivers/middleware/stdlib"
	"github.com/ulule/limiter/v3/drivers/store/memory"
)

// DefaultRateLimit is the default number of allowed requests per second
const DefaultRateLimit = 55

// RateLimitMiddleware is a convenience function for creating rate limiting middleware
//
//	router.Use(RateLimitMiddleware(limiter))
func RateLimitMiddleware(cfg config.Configuration, limiter *limiter.Limiter) mux.MiddlewareFunc {
	ipGetter := stdlib.WithKeyGetter(
		func(r *http.Request) string {
			var remoteIP string

			if host, _, err := net.SplitHostPort(r.RemoteAddr); err != nil {
				slog.WarnContext(r.Context(), fmt.Sprintf("Error parsing remoteAddress '%s': %s", r.RemoteAddr, err))
				remoteIP = r.RemoteAddr
			} else {
				remoteIP = host
			}

			if cfg.TrustedProxies == 0 {
				return remoteIP
			} else if xff := r.Header.Get("X-Forwarded-For"); xff == "" {
				slog.WarnContext(r.Context(), "Expected data in X-Forwarded-For header")
				return remoteIP
			} else {
				ips := strings.Split(xff, ",")

				idxIP := len(ips) - cfg.TrustedProxies
				if idxIP < 0 {
					slog.WarnContext(r.Context(), "Not enough IPs in X-Forwarded-For, defaulting to first IP")
					idxIP = 0
				}

				return strings.TrimSpace(ips[idxIP])
			}
		},
	)

	middleware := stdlib.NewMiddleware(limiter, ipGetter)

	return func(next http.Handler) http.Handler {
		return middleware.Handler(RemoveRateLimitHeadersMiddleware(next))
	}
}

// RemoveRateLimitHeadersMiddleware removes rate limit headers that we do not want appearing in our responses
func RemoveRateLimitHeadersMiddleware(next http.Handler) http.Handler {
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
//	router.Use(DefaultRateLimitMiddleware())
func DefaultRateLimitMiddleware() mux.MiddlewareFunc {
	rate := limiter.Rate{
		Period: 1 * time.Second,
		Limit:  DefaultRateLimit,
	}

	store := memory.NewStore()

	instance := limiter.New(store, rate)

	cfg, err := config.NewDefaultConfiguration()
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to create default configuration: %v", err))
	}

	return RateLimitMiddleware(cfg, instance)
}
