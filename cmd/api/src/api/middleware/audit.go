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

// AuditService is the narrow port the audit middleware depends on, satisfied by
// *audit.Service.
type AuditService interface {
	Intent(ctx context.Context, entry audit.Entry) (uuid.UUID, error)
	Success(ctx context.Context, commitID uuid.UUID, entry audit.Entry) error
	Failure(ctx context.Context, commitID uuid.UUID, entry audit.Entry) error
}

// AuditMiddleware records the intent/success/failure lifecycle of every API
// request. It writes an intent row before the handler runs and a success or
// failure row afterward based on the response status; a failed intent write
// rejects the request with a 500. Route templates for which isExcluded returns
// true are skipped; a nil isExcluded audits every matched route.
func AuditMiddleware(auditService AuditService, muxRouter *mux.Router, isExcluded func(routeTemplate string) bool) mux.MiddlewareFunc {
	if isExcluded == nil {
		isExcluded = func(string) bool { return false }
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
			auditHandler(auditService, muxRouter, isExcluded, next, response, request)
		})
	}
}

// auditIntentWriteTimeout bounds the fail-closed intent write so a slow or
// unavailable database rejects the request fast rather than blocking it.
const auditIntentWriteTimeout = 5 * time.Second

// auditOutcomeWriteTimeout bounds the best-effort success/failure write.
const auditOutcomeWriteTimeout = 30 * time.Second

func auditHandler(auditService AuditService, muxRouter *mux.Router, isExcluded func(routeTemplate string) bool, next http.Handler, response http.ResponseWriter, request *http.Request) {
	var (
		ctx           = request.Context()
		routeTemplate = routeTemplateFor(muxRouter, request)
	)

	// Skip routes we cannot name and routes opted out at registration (e.g. /health).
	if routeTemplate == unmatchedRouteLabel || isExcluded(routeTemplate) {
		next.ServeHTTP(response, request)
		return
	}

	// Derived from the request context so a client disconnect cancels the write,
	// with a timeout that bounds it even when the request context has no deadline.
	intentCtx, cancelIntent := context.WithTimeout(ctx, auditIntentWriteTimeout)
	defer cancelIntent()

	var (
		recorder      = &responseRecorder{delegate: response}
		entry         = buildAuditEntry(request, routeTemplate)
		commitID, err = auditService.Intent(intentCtx, entry)
	)
	// Fail closed: without a durable intent row the handler is never invoked.
	if err != nil {
		slog.ErrorContext(ctx, "Failed to write audit intent row", attr.Error(err))
		api.WriteErrorResponse(ctx, api.BuildErrorResponse(http.StatusInternalServerError, "audit log intent could not be recorded", request), response)
		return
	}

	// Trap a handler panic so the intent row still gets a matching failure row,
	// then re-panic so the outer PanicHandler logs and aborts the request.
	defer func() {
		if recovery := recover(); recovery != nil {
			// The handler goroutine is unwinding, so detach from request cancellation.
			panicCtx, cancelPanic := context.WithTimeout(context.WithoutCancel(ctx), auditOutcomeWriteTimeout)
			defer cancelPanic()

			if failureErr := auditService.Failure(panicCtx, commitID, entry); failureErr != nil {
				slog.ErrorContext(panicCtx, "Failed to write audit failure row after handler panic", attr.Error(failureErr))
			}

			panic(recovery)
		}
	}()

	next.ServeHTTP(recorder, request)

	// Best-effort but must survive a client disconnect, so detach from request
	// cancellation.
	outcomeCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), auditOutcomeWriteTimeout)
	defer cancel()

	if recorder.statusCode >= http.StatusBadRequest {
		if failureErr := auditService.Failure(outcomeCtx, commitID, entry); failureErr != nil {
			slog.ErrorContext(outcomeCtx, "Failed to write audit failure row", attr.Error(failureErr))
		}
	} else if successErr := auditService.Success(outcomeCtx, commitID, entry); successErr != nil {
		slog.ErrorContext(outcomeCtx, "Failed to write audit success row", attr.Error(successErr))
	}
}

// buildAuditEntry assembles the audit Entry from the request context, resolving
// the actor from the authenticated user. An unauthenticated request leaves the
// actor fields empty (attributed by source IP); the audit service applies the
// unknown-actor default centrally.
func buildAuditEntry(request *http.Request, routeTemplate string) audit.Entry {
	var (
		bhCtx = bhctx.FromRequest(request)
		entry = audit.Entry{
			Action:          request.Method + routeTemplate,
			RequestID:       bhCtx.RequestID,
			SourceIPAddress: parseUserIP(request),
			Fields:          map[string]any{},
		}
	)

	if user, ok := auth.GetUserFromAuthCtx(bhCtx.AuthCtx); ok {
		entry.ActorID = user.ID.String()
		entry.ActorName = user.PrincipalName
		entry.ActorEmail = user.EmailAddress.ValueOrZero()
	}

	return entry
}
