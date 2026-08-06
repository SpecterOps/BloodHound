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
// used to resolve the bounded route template for the audited action.
func AuditMiddleware(auditService AuditService, muxRouter *mux.Router) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
			auditHandler(auditService, muxRouter, next, response, request)
		})
	}
}

func auditHandler(auditService AuditService, muxRouter *mux.Router, next http.Handler, response http.ResponseWriter, request *http.Request) {
	var (
		ctx           = request.Context()
		recorder      = &responseRecorder{delegate: response}
		entry         = buildAuditEntry(request, muxRouter)
		commitID, err = auditService.Intent(ctx, entry)
	)
	// A failed intent write rejects the request: without a durable intent row
	// the action would run unaudited, so the handler is never invoked.
	if err != nil {
		slog.ErrorContext(ctx, "Failed to write audit intent row", attr.Error(err))
		api.WriteErrorResponse(ctx, api.BuildErrorResponse(http.StatusInternalServerError, "audit log intent could not be recorded", request), response)
		return
	}

	next.ServeHTTP(recorder, request)

	if recorder.statusCode >= http.StatusBadRequest {
		if failureErr := auditService.Failure(ctx, commitID, entry); failureErr != nil {
			slog.ErrorContext(ctx, "Failed to write audit failure row", attr.Error(failureErr))
		}
	} else if successErr := auditService.Success(ctx, commitID, entry); successErr != nil {
		slog.ErrorContext(ctx, "Failed to write audit success row", attr.Error(successErr))
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
func buildAuditEntry(request *http.Request, muxRouter *mux.Router) audit.Entry {
	var (
		bhCtx = bhctx.FromRequest(request)
		entry = audit.Entry{
			Action:          request.Method + routeTemplateFor(muxRouter, request),
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
