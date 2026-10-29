// Copyright 2026 Specter Ops, Inc.
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
	"log/slog"
	"net/http"
	"time"

	"github.com/gofrs/uuid"
	"github.com/gorilla/mux"

	"github.com/specterops/bloodhound/cmd/api/src/api"
	"github.com/specterops/bloodhound/cmd/api/src/auth"
	"github.com/specterops/bloodhound/cmd/api/src/bhctx"
	"github.com/specterops/bloodhound/packages/go/bhlog/attr"
	"github.com/specterops/bloodhound/server/audit"
)

const (
	auditIntentWriteTimeout  = 5 * time.Second
	auditOutcomeWriteTimeout = 2 * time.Second
)

// AuditService is the audit lifecycle port used by the middleware.
type AuditService interface {
	Intent(ctx context.Context, entry audit.Entry) (uuid.UUID, error)
	Success(ctx context.Context, commitID uuid.UUID, entry audit.Entry) error
	Failure(ctx context.Context, commitID uuid.UUID, entry audit.Entry) error
}

type endpointAuditMiddleware struct {
	auditService AuditService
	muxRouter    *mux.Router
	isExcluded   func(routeTemplate string) bool
}

// AuditMiddleware records the request's audit lifecycle.
func AuditMiddleware(auditService AuditService, muxRouter *mux.Router, isExcluded func(routeTemplate string) bool) mux.MiddlewareFunc {
	if isExcluded == nil {
		isExcluded = func(string) bool { return false }
	}

	middlewareHandler := endpointAuditMiddleware{
		auditService: auditService,
		muxRouter:    muxRouter,
		isExcluded:   isExcluded,
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
			middlewareHandler.serveHTTP(next, response, request)
		})
	}
}

func (s endpointAuditMiddleware) serveHTTP(next http.Handler, response http.ResponseWriter, request *http.Request) {
	var (
		ctx           = request.Context()
		routeTemplate = routeTemplateFor(s.muxRouter, request)
	)

	if routeTemplate == unmatchedRouteLabel || s.isExcluded(routeTemplate) {
		next.ServeHTTP(response, request)
		return
	}

	var (
		recorder      = &responseRecorder{delegate: response}
		entry         = buildAuditEntry(request, routeTemplate)
		commitID, err = s.writeIntent(ctx, entry)
	)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to write audit intent row", attr.Error(err))
		api.WriteErrorResponse(ctx, api.BuildErrorResponse(http.StatusInternalServerError, "audit log intent could not be recorded", request), response)
		return
	}

	defer func() {
		if panicValue := recover(); panicValue != nil {
			s.recordFailure(ctx, commitID, entry, "Failed to write audit failure row after handler panic")
			panic(panicValue)
		}
	}()

	next.ServeHTTP(recorder, request)
	s.recordOutcome(ctx, recorder.StatusCode(), commitID, entry)
}

func (s endpointAuditMiddleware) writeIntent(ctx context.Context, entry audit.Entry) (uuid.UUID, error) {
	intentCtx, cancel := context.WithTimeout(ctx, auditIntentWriteTimeout)
	defer cancel()

	return s.auditService.Intent(intentCtx, entry)
}

func (s endpointAuditMiddleware) recordOutcome(ctx context.Context, statusCode int, commitID uuid.UUID, entry audit.Entry) {
	if statusCode >= http.StatusBadRequest {
		s.recordFailure(ctx, commitID, entry, "Failed to write audit failure row")
		return
	}

	outcomeCtx, cancel := newAuditOutcomeContext(ctx)
	defer cancel()

	if err := s.auditService.Success(outcomeCtx, commitID, entry); err != nil {
		slog.ErrorContext(outcomeCtx, "Failed to write audit success row", attr.Error(err))
	}
}

func (s endpointAuditMiddleware) recordFailure(ctx context.Context, commitID uuid.UUID, entry audit.Entry, logMessage string) {
	outcomeCtx, cancel := newAuditOutcomeContext(ctx)
	defer cancel()

	if err := s.auditService.Failure(outcomeCtx, commitID, entry); err != nil {
		slog.ErrorContext(outcomeCtx, logMessage, attr.Error(err))
	}
}

func newAuditOutcomeContext(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.WithoutCancel(ctx), auditOutcomeWriteTimeout)
}

func buildAuditEntry(request *http.Request, routeTemplate string) audit.Entry {
	var (
		requestContext = bhctx.FromRequest(request)
		entry          = audit.Entry{
			Action:          request.Method + routeTemplate,
			RequestID:       requestContext.RequestID,
			SourceIPAddress: requestContext.RequestIP,
			Fields:          map[string]any{},
		}
	)

	if user, ok := auth.GetUserFromAuthCtx(requestContext.AuthCtx); ok {
		entry.ActorID = user.ID.String()
		entry.ActorName = user.PrincipalName
		entry.ActorEmail = user.EmailAddress.ValueOrZero()
	}

	return entry
}
