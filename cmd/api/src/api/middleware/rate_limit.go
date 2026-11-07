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
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/specterops/bloodhound/cmd/api/src/database"
	"github.com/specterops/bloodhound/cmd/api/src/model/appcfg"
	"github.com/specterops/bloodhound/packages/go/bhlog/attr"
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
				slog.WarnContext(
					r.Context(),
					"Error parsing remoteAddress",
					slog.String("remote_addr", r.RemoteAddr),
					attr.Error(err),
				)
				remoteIP = r.RemoteAddr
			} else {
				remoteIP = host
			}

			if trustedProxies <= 0 {
				slog.DebugContext(
					r.Context(),
					"Using direct remote IP Address for rate limiting",
					slog.String("ip_address", remoteIP),
				)
				return remoteIP
			} else if xff := r.Header.Get("X-Forwarded-For"); xff == "" {
				slog.DebugContext(
					r.Context(),
					"Expected X-Forwarded-For header for rate limiting but none found. Defaulted to remote IP Address",
					slog.String("ip_address", remoteIP),
				)
				return remoteIP
			} else {
				ips := strings.Split(xff, ",")

				idxIP := len(ips) - trustedProxies
				if idxIP < 0 {
					slog.WarnContext(
						r.Context(),
						"Not enough IPs in X-Forwarded-For, defaulting to first IP",
						slog.String("x_forwarded_for", xff),
					)
					idxIP = 0
				}

				finalIP := strings.TrimSpace(ips[idxIP])

				slog.DebugContext(
					r.Context(),
					"Found client IP Address for rate limiting in XFF",
					slog.String("ip_address", finalIP),
					slog.String("x_forwarded_for", xff),
				)
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
// with a particular limit for a router/route
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

// MatchedRouteRateLimitMiddleware applies an independent IP rate limit to each
// matched route. It is intended for the router's post-routing middleware chain
// so the limit runs before authentication and route middleware.
func MatchedRouteRateLimitMiddleware(db database.Database, excludedPathPrefixes []string, limitsByPath map[string]int64) mux.MiddlewareFunc {
	var (
		handlerByRoute = make(map[*mux.Route]http.Handler)
		mutex          sync.RWMutex
	)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
			if isRateLimitExcludedPath(request.URL.Path, excludedPathPrefixes) {
				next.ServeHTTP(response, request)
				return
			}

			matchedRoute := mux.CurrentRoute(request)
			if matchedRoute == nil {
				next.ServeHTTP(response, request)
				return
			}

			mutex.RLock()
			routeHandler, found := handlerByRoute[matchedRoute]
			mutex.RUnlock()
			if !found {
				mutex.Lock()
				routeHandler, found = handlerByRoute[matchedRoute]
				if !found {
					routeHandler = RateLimitMiddleware(db, rateLimitForRoute(matchedRoute, limitsByPath))(next)
					handlerByRoute[matchedRoute] = routeHandler
				}
				mutex.Unlock()
			}

			routeHandler.ServeHTTP(response, request)
		})
	}
}

func isRateLimitExcludedPath(requestPath string, excludedPathPrefixes []string) bool {
	for _, excludedPathPrefix := range excludedPathPrefixes {
		if requestPath == excludedPathPrefix || strings.HasPrefix(requestPath, strings.TrimRight(excludedPathPrefix, "/")+"/") {
			return true
		}
	}

	return false
}

func rateLimitForRoute(route *mux.Route, limitsByPath map[string]int64) int64 {
	routePath, err := route.GetPathTemplate()
	if err != nil {
		return DefaultRateLimit
	}

	if limit, found := limitsByPath[routePath]; found {
		return limit
	}

	return DefaultRateLimit
}
