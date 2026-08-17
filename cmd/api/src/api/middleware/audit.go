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

// AuditService is the narrow port the audit middleware depends on. It is
// satisfied by *audit.Service. Writes are synchronous: a failed intent write
// causes the middleware to reject the underlying request, while result
// (success/failure) writes are best-effort and only logged on error.
type AuditService interface {
	Intent(ctx context.Context, entry audit.Entry) (uuid.UUID, error)
	Success(ctx context.Context, commitID uuid.UUID, entry audit.Entry) error
	Failure(ctx context.Context, commitID uuid.UUID, entry audit.Entry) error
}

// AuditMiddleware records the intent/success/failure lifecycle of every API
// request. It writes an intent row before the handler runs and a success or
// failure row afterward based on the response status. If the intent write fails
// the request is rejected with a 500 and the handler never runs. muxRouter is
// used to resolve the bounded route template for the audited action. Route
// templates listed in excludedRoutes (e.g. /health) are not audited.
func AuditMiddleware(auditService AuditService, muxRouter *mux.Router, excludedRoutes ...string) mux.MiddlewareFunc {
	exclusions := make(map[string]bool, len(excludedRoutes))
	for _, route := range excludedRoutes {
		exclusions[route] = true
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
			auditHandler(auditService, muxRouter, exclusions, next, response, request)
		})
	}
}

// auditIntentWriteTimeout bounds the fail-closed intent write so a slow or
// unavailable database cannot block the request goroutine indefinitely before the
// handler runs. Because the intent write is fail-closed, exceeding this deadline
// rejects the request with a 500 rather than hanging: under sustained database
// pressure the audit table becomes a bounded availability dependency for audited
// endpoints (the request fails fast) instead of an unbounded one (requests pile
// up waiting on the write).
const auditIntentWriteTimeout = 5 * time.Second

// auditOutcomeWriteTimeout bounds the best-effort success/failure write so a slow
// or unavailable database cannot block the request goroutine indefinitely after
// the handler has already returned.
const auditOutcomeWriteTimeout = 30 * time.Second

func auditHandler(auditService AuditService, muxRouter *mux.Router, excludedRoutes map[string]bool, next http.Handler, response http.ResponseWriter, request *http.Request) {
	var (
		ctx           = request.Context()
		routeTemplate = routeTemplateFor(muxRouter, request)
	)

	// Skip routes we cannot name (routeTemplate == unmatchedRouteLabel) and
	// explicitly excluded routes such as /health: they carry no audit value and
	// auditing them would place a synchronous, fail-closed write in front of
	// high-volume traffic.
	if routeTemplate == unmatchedRouteLabel || excludedRoutes[routeTemplate] {
		next.ServeHTTP(response, request)
		return
	}

	// The intent write is bounded by auditIntentWriteTimeout. The context is
	// derived from the request context so a client disconnecting before the
	// handler runs still cancels the write, while the timeout guarantees an upper
	// bound even when the request context has no deadline of its own.
	intentCtx, cancelIntent := context.WithTimeout(ctx, auditIntentWriteTimeout)
	defer cancelIntent()

	var (
		recorder      = &responseRecorder{delegate: response}
		entry         = buildAuditEntry(request, routeTemplate)
		commitID, err = auditService.Intent(intentCtx, entry)
	)
	// A failed intent write rejects the request: without a durable intent row
	// the action would run unaudited, so the handler is never invoked. This is
	// fail-closed, so an intent write that exceeds auditIntentWriteTimeout also
	// rejects the request rather than proceeding unaudited.
	if err != nil {
		slog.ErrorContext(ctx, "Failed to write audit intent row", attr.Error(err))
		api.WriteErrorResponse(ctx, api.BuildErrorResponse(http.StatusInternalServerError, "audit log intent could not be recorded", request), response)
		return
	}

	next.ServeHTTP(recorder, request)

	// The success/failure write is best-effort but must survive the client
	// disconnecting after the handler runs, so it uses a context detached from the
	// request's cancellation, bounded by auditOutcomeWriteTimeout.
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

// anonymousActorName is used as the actor name for unauthenticated requests so
// that the audit record is attributed to an explicit anonymous actor (tracked by
// source IP) rather than being dropped.
const anonymousActorName = "anonymous"

// buildAuditEntry assembles the audit Entry from the request context, resolving
// the actor from the authenticated user when present and falling back to an
// anonymous actor attributed to the source IP when the request is
// unauthenticated.
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
	} else {
		entry.ActorName = anonymousActorName
	}

	return entry
}
