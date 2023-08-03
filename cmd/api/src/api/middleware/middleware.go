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
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gofrs/uuid"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/specterops/bloodhound/headers"
	"github.com/specterops/bloodhound/log"
	"github.com/specterops/bloodhound/src/api"
	"github.com/specterops/bloodhound/src/config"
	"github.com/specterops/bloodhound/src/ctx"
	"github.com/specterops/bloodhound/src/utils"
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

var ErrorResponseContextDeadlineExceeded = "request took longer than the configured timeout"

func (s *Wrapper) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	doneChannel := make(chan struct{})

	go func(ctx context.Context) {
		defer close(doneChannel)
		handler := s.handler
		for idx := len(s.middleware) - 1; idx >= 0; idx-- {
			handler = s.middleware[idx](handler)
		}

		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return
		}

		handler.ServeHTTP(response, request)
	}(request.Context())

	select {
	case <-request.Context().Done():
		waitDuration, _ := requestWaitDuration(request, defaultTimeout)
		api.WriteErrorResponse(context.Background(), api.BuildErrorResponse(http.StatusGatewayTimeout, fmt.Sprintf("%s: %v second(s)", ErrorResponseContextDeadlineExceeded, waitDuration.Seconds()), request), response)
	case <-doneChannel:
	}
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

const (
	// Default timeout for any request is thirty seconds
	defaultTimeout = 30 * time.Second

	// This is set to 1 minute since intermediate load balancers may terminate requests that sit longer than 1 minute
	maximumTimeout = time.Minute
)

// ContextMiddleware is a middleware function that sets the BloodHound context per-request. It also sets the request ID.
func ContextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		var (
			startTime = time.Now()
			requestID string
		)

		if newUUID, err := uuid.NewV4(); err != nil {
			log.Errorf("Failed generating a new request UUID: %v", err)
			requestID = "ERROR"
		} else {
			requestID = newUUID.String()
		}

		if waitDuration, err := requestWaitDuration(request, defaultTimeout); err != nil {
			// If there is a failure or other expectation mismatch with the client, respond right away with the relevant
			// error information
			api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf("Prefer header has an invalid value: %v", err), request), response)
		} else {
			// Set the request ID and applied preferences headers
			response.Header().Set(headers.RequestID.String(), requestID)
			response.Header().Set(headers.StrictTransportSecurity.String(), utils.HSTSSetting)
			response.Header().Set(headers.PreferenceApplied.String(), fmt.Sprintf("wait=%f", waitDuration.Seconds()))

			// Create a new context with the timeout
			requestCtx, cancel := context.WithTimeout(request.Context(), waitDuration)
			defer cancel()

			// Chain the request context with the BloodHound context and request ID
			requestCtx = context.WithValue(
				requestCtx,
				ctx.ValueKey,
				&ctx.Context{
					StartTime: startTime,
					RequestID: requestID,
					Host: &url.URL{
						Scheme: getScheme(request),
						Host:   request.Host,
					},
				})

			next.ServeHTTP(response, request.WithContext(requestCtx))
		}
	})
}

func requestWaitDuration(request *http.Request, defaultWaitDuration time.Duration) (time.Duration, error) {
	waitDuration := defaultWaitDuration
	if preferValue := request.Header.Get(headers.Prefer.String()); len(preferValue) > 0 {
		if requestedWaitDuration, err := parsePreferHeaderWait(preferValue, defaultWaitDuration); err != nil {
			return 0, err
		} else if requestedWaitDuration <= maximumTimeout {
			waitDuration = requestedWaitDuration
		}
	}

	return waitDuration, nil
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

func parsePreferHeaderWait(value string, defaultWaitDuration time.Duration) (time.Duration, error) {
	var (
		values                    = ParseHeaderValues(value)
		waitStr, hasWaitSpecified = values["wait"]
	)

	if hasWaitSpecified {
		if parsedNumSeconds, err := strconv.Atoi(waitStr); err != nil {
			return 0, err
		} else {
			return time.Second * time.Duration(parsedNumSeconds), nil
		}
	}

	return defaultWaitDuration, nil
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
