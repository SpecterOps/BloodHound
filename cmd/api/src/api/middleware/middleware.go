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
	"compress/gzip"
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gofrs/uuid"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/specterops/bloodhound/cmd/api/src/api"
	"github.com/specterops/bloodhound/cmd/api/src/config"
	"github.com/specterops/bloodhound/cmd/api/src/ctx"
	"github.com/specterops/bloodhound/cmd/api/src/database"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/cmd/api/src/utils"
	"github.com/specterops/bloodhound/packages/go/headers"
	"github.com/unrolled/secure"
)

// Wrapper is an iterator for middleware function application that wraps around a http.Handler.
type Wrapper struct {
	middleware []mux.MiddlewareFunc
	handler    http.Handler
}

func NewWrapper(handler http.Handler) *Wrapper {
	return &Wrapper{
		handler: handler,
	}
}

func (s *Wrapper) Use(middlewareFunc ...mux.MiddlewareFunc) {
	s.middleware = append(s.middleware, middlewareFunc...)
}

func (s *Wrapper) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	handler := s.handler

	// Wrap the handler in its middleware
	for idx := len(s.middleware) - 1; idx >= 0; idx-- {
		handler = s.middleware[idx](handler)
	}

	// Route the request
	handler.ServeHTTP(response, request)
}

func getScheme(request *http.Request) string {
	if fwdProto := request.Header.Get("X-Forwarded-Proto"); fwdProto != "" {
		return fwdProto
	}

	if request.TLS != nil {
		return "https"
	} else {
		return "http"
	}
}

func RequestWaitDuration(request *http.Request) (time.Duration, error) {
	var (
		requestedWaitDuration time.Duration
		err                   error
	)

	if preferValue := request.Header.Get(headers.Prefer.String()); len(preferValue) > 0 {
		if requestedWaitDuration, err = parsePreferHeaderWait(preferValue); err != nil {
			return 0, err
		}
	}
	return requestedWaitDuration, nil
}

// ContextMiddleware is a middleware function that sets the BloodHound context per-request. It also sets the request ID.
func ContextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		var (
			startTime = time.Now()
			requestID string
		)

		if newUUID, err := uuid.NewV4(); err != nil {
			slog.ErrorContext(request.Context(), fmt.Sprintf("Failed generating a new request UUID: %v", err))
			requestID = "ERROR"
		} else {
			requestID = newUUID.String()
		}

		if requestedWaitDuration, err := RequestWaitDuration(request); err != nil {
			// If there is a failure or other expectation mismatch with the client, respond right away with the relevant
			// error information
			api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf("Prefer header has an invalid value: %v", err), request), response)
		} else {
			// Set the request ID and applied preferences headers
			response.Header().Set(headers.RequestID.String(), requestID)
			response.Header().Set(headers.StrictTransportSecurity.String(), utils.HSTSSetting)

			var (
				requestCtx = request.Context()
				cancel     context.CancelFunc
			)

			// API requests don't have a timeout set by default. Below, we set a custom timeout to the request only if specified in the prefer header
			if requestedWaitDuration > 0 {
				response.Header().Set(headers.PreferenceApplied.String(), fmt.Sprintf("wait=%.2f", requestedWaitDuration.Seconds()))

				requestCtx, cancel = context.WithTimeout(request.Context(), requestedWaitDuration)
				defer cancel()
			}

			// Insert the bh context
			requestCtx = ctx.Set(requestCtx, &ctx.Context{
				StartTime: startTime,
				RequestID: requestID,
				Timeout:   requestedWaitDuration,
				Host: &url.URL{
					Scheme: getScheme(request),
					Host:   request.Host,
				},
				RequestedURL: model.AuditableURL(request.URL.String()),
				RequestIP:    parseUserIP(request),
				RemoteAddr:   request.RemoteAddr,
			})

			// Route the request with the embedded context
			next.ServeHTTP(response, request.WithContext(requestCtx))
		}
	})
}

func parseUserIP(r *http.Request) string {
	var remoteIp string

	// The point of this code is to strip the port, so we don't need to save it.
	if host, _, err := net.SplitHostPort(r.RemoteAddr); err != nil {
		slog.WarnContext(r.Context(), fmt.Sprintf("Error parsing remoteAddress '%s': %s", r.RemoteAddr, err))
		remoteIp = r.RemoteAddr
	} else {
		remoteIp = host
	}

	if result := r.Header.Get("X-Forwarded-For"); result == "" {
		slog.DebugContext(r.Context(), "No data found in X-Forwarded-For header")
		return remoteIp
	} else {
		result += "," + remoteIp
		return result
	}
}

func ParseHeaderValues(values string) map[string]string {
	parsed := map[string]string{}

	for _, part := range strings.Split(values, ",") {
		var (
			sanitizedPart = strings.TrimSpace(part)
			keyValueSlice = strings.SplitN(sanitizedPart, "=", 2)
			sanitizedKey  = strings.TrimSpace(keyValueSlice[0])
		)

		if sanitizedKey != "" {
			var sanitizedValue string

			if len(keyValueSlice) == 2 {
				sanitizedValue = strings.TrimSpace(keyValueSlice[1])
			}

			parsed[sanitizedKey] = sanitizedValue
		}
	}

	return parsed
}

func parsePreferHeaderWait(value string) (time.Duration, error) {
	if waitStr, hasWaitSpecified := ParseHeaderValues(value)["wait"]; hasWaitSpecified {
		if parsedNumSeconds, err := strconv.Atoi(waitStr); err != nil {
			return 0, err
		} else {
			return time.Second * time.Duration(parsedNumSeconds), nil
		}
	} else {
		return 0, nil
	}
}

// CORSMiddleware is a middleware function that sets our CORS options per-request and answers to client OPTIONS requests
//
// Note: AllowedOrigins is set to "" on purpose; otherwise, the Access-Control-Allow-Origin header (ACAO) will be set to
// "*" which is too permissive. If we want a more permissive CORS policy we should explicitly set the
// "Referrer-Policy" header and mirror the request Origin/Referrer in the ACAO header when they match an allowed
// hostname.
//
// XXX: This is not enough to protect against CSRF attacks
func CORSMiddleware() mux.MiddlewareFunc {
	return handlers.CORS(
		handlers.AllowCredentials(),
		handlers.AllowedMethods([]string{"HEAD", "GET", "POST", "DELETE", "PUT"}),
		handlers.AllowedHeaders([]string{headers.ContentType.String(), headers.Authorization.String()}),
		handlers.AllowedOrigins([]string{""}),
	)
}

func SecureHandlerMiddleware(cfg config.Configuration, contentSecurityPolicy string) mux.MiddlewareFunc {
	const (
		permissionsPolicy = "fullscreen=*, unsized-media=*, unoptimized-images=*, geolocation=(), camera=(), microphone=(), payment=()"
		referrerPolicy    = "strict-origin-when-cross-origin"
	)

	return secure.New(secure.Options{
		// Redirect all requests to HTTPS
		SSLRedirect: true,

		// Content-Security-Policy
		ContentSecurityPolicy: contentSecurityPolicy,

		// X-Content-Type-Options
		ContentTypeNosniff: true,

		// HSTS
		STSSeconds:           31536000,
		STSPreload:           true,
		STSIncludeSubdomains: true,

		//Referrer-Policy
		ReferrerPolicy: referrerPolicy,

		// Permissions Policy
		PermissionsPolicy: permissionsPolicy,

		// X-Frame-Options
		CustomFrameOptionsValue: "SAMEORIGIN",

		// This will cause the AllowedHosts, SSLRedirect, and STSSeconds/STSIncludeSubdomains options to be ignored during development
		IsDevelopment: !cfg.TLS.Enabled(),
	}).Handler
}

// FeatureFlagMiddleware is a middleware that enables or disables a given endpoint based on the status of the passed feature flag.
// It is intended to be attached directly to endpoints that should be affected by the feature flag. The feature flag determining the
// endpoint's availability should be specified in flagKey.
//
// If the flag is enabled, the endpoint will work as intended. If the flag is disabled, a 404 will be returned to the user.
func FeatureFlagMiddleware(db database.Database, flagKey string) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
			if flag, err := db.GetFlagByKey(request.Context(), flagKey); err != nil {
				api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusInternalServerError, fmt.Sprintf("error retrieving %s feature flag: %s", flagKey, err), request), response)
			} else if flag.Enabled {
				next.ServeHTTP(response, request)
			} else {
				api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusNotFound, api.ErrorResponseDetailsResourceNotFound, request), response)
			}
		})
	}
}

func EnsureRequestBodyClosed() mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
			next.ServeHTTP(response, request)

			// This type cast is required because of the way that Go interfaces work. It is possible to have an
			// interface pointed at a pointer that points at nil. This would result in the interface not being nil
			// but still cause a panic while acting on the interface. https://go.dev/doc/faq#nil_error
			switch b := request.Body.(type) {
			case *gzip.Reader:
				if b != nil {
					b.Close()
				}
			default:
				if b != nil {
					b.Close()
				}
			}
		})
	}
}
