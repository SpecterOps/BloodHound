# Audit Middleware Reference Implementation

**Related RFC:** [bh-rfc-7.md](../../../../rfc/bh-rfc-7.md)
**Status:** Reference Implementation
**Last Updated:** 2026-06-16

## Overview

This document provides a reference implementation of the audit middleware described in RFC 7. This code is illustrative and demonstrates the key patterns; production implementation should follow established middleware patterns in the codebase.

---

## Middleware Constructor

```go
package middleware

import (
    "context"
    "log/slog"
    "net/http"

    "github.com/gorilla/mux"
    "github.com/specterops/bloodhound/cmd/api/src/auth"
    "github.com/specterops/bloodhound/cmd/api/src/ctx"
    "github.com/specterops/bloodhound/cmd/api/src/model"
    "github.com/specterops/bloodhound/server/audit"
)

// ExclusionList is a set of route templates that should not be audited.
type ExclusionList map[string]bool

func (e ExclusionList) Contains(routeTemplate string) bool {
    return e[routeTemplate]
}

// NewAuditMiddleware constructs the audit middleware with the injected audit service.
// The middleware writes an intent row before the handler runs and a success/failure
// row after it completes.
func NewAuditMiddleware(
    auditService audit.Service,
    muxRouter *mux.Router,
    idResolver auth.IdentityResolver,
    exclusions []string,
) mux.MiddlewareFunc {
    exclusionSet := make(ExclusionList, len(exclusions))
    for _, route := range exclusions {
        exclusionSet[route] = true
    }

    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
            auditMiddlewareHandler(
                auditService,
                muxRouter,
                idResolver,
                exclusionSet,
                next,
                response,
                request,
            )
        })
    }
}
```

---

## Middleware Handler

```go
func auditMiddlewareHandler(
    auditService audit.Service,
    muxRouter *mux.Router,
    idResolver auth.IdentityResolver,
    exclusions ExclusionList,
    next http.Handler,
    response http.ResponseWriter,
    request *http.Request,
) {
    var (
        ctx            = request.Context()
        requestContext = ctx.Get(ctx)  // bhctx.Context
        routeTemplate  = routeTemplateFor(muxRouter, request)
        recorder       = &responseRecorder{delegate: response, statusCode: http.StatusOK}
    )

    // Skip excluded routes (e.g., /health)
    if exclusions.Contains(routeTemplate) {
        next.ServeHTTP(response, request)
        return
    }

    // Build the audit entry with default route-template action
    entry := audit.Entry{
        Action:          model.AuditLogAction(request.Method + " " + routeTemplate),
        RequestID:       requestContext.RequestID,
        SourceIpAddress: requestContext.RequestIP,
        Fields:          make(map[string]any),
    }

    // Resolve actor from auth context, or use anonymous if unauthenticated
    if identity, err := idResolver.GetIdentity(requestContext.AuthCtx); err == nil {
        entry.ActorID = identity.ID.String()
        entry.ActorName = identity.Name
        entry.ActorEmail = identity.Email
    }
    // If GetIdentity fails, actor fields remain empty (anonymous actor)

    // INTENT WRITE (synchronous): Pre-execution record
    // The audit service returns the commit_id that links this intent to its result
    commitID, err := auditService.Intent(ctx, entry)
    if err != nil {
        // Log but do not fail the request
        slog.ErrorContext(ctx, "failed to write audit intent", slog.String("err", err.Error()))
    }

    // Execute the handler
    next.ServeHTTP(recorder, request)

    // Check if handler contributed semantic action and/or fields via context
    if contribution := audit.FromContext(ctx); contribution != nil {
        if contribution.Action != "" {
            entry.Action = contribution.Action  // Override route-template with semantic action
        }
        if contribution.Fields != nil {
            entry.Fields = contribution.Fields  // Add handler-contributed fields
        }
    }

    // RESULT WRITE (async via worker): Post-execution record
    // HTTP status -> outcome: <400 = success, >=400 = failure
    if recorder.statusCode >= http.StatusBadRequest {
        if err := auditService.Failure(ctx, commitID, entry); err != nil {
            slog.ErrorContext(ctx, "failed to write audit failure", slog.String("err", err.Error()))
        }
    } else {
        if err := auditService.Success(ctx, commitID, entry); err != nil {
            slog.ErrorContext(ctx, "failed to write audit success", slog.String("err", err.Error()))
        }
    }
}
```

---

## Route Template Helper

```go
// routeTemplateFor extracts the route template from the request using gorilla/mux.
// This is the same helper used for Prometheus metrics in the codebase.
func routeTemplateFor(router *mux.Router, request *http.Request) string {
    var match mux.RouteMatch
    if router.Match(request, &match) && match.Route != nil {
        if template, err := match.Route.GetPathTemplate(); err == nil {
            return template
        }
    }
    return "unknown"
}
```

---

## Response Recorder

```go
// responseRecorder wraps http.ResponseWriter to capture the status code.
type responseRecorder struct {
    delegate   http.ResponseWriter
    statusCode int
}

func (r *responseRecorder) Header() http.Header {
    return r.delegate.Header()
}

func (r *responseRecorder) Write(data []byte) (int, error) {
    return r.delegate.Write(data)
}

func (r *responseRecorder) WriteHeader(statusCode int) {
    r.statusCode = statusCode
    r.delegate.WriteHeader(statusCode)
}
```

---

## Middleware Registration

In the startup entrypoint (`lib/go/services/entrypoint.go` or registration package):

```go
// Register BHCE modules and obtain cross-cutting services
bhceDeps := modules.Deps{
    Router: &routerInst,
    Pool:   connections.RDMS.Pool(),
}

bhceServices, err := modules.Register(bhceDeps)
if err != nil {
    return fmt.Errorf("failed to register BHCE modules: %w", err)
}

// Construct and register audit middleware
auditMiddleware := middleware.NewAuditMiddleware(
    bhceServices.Audit,
    routerInst.MuxRouter(),
    identityResolver,
    []string{"/health"}, // Exclusion list
)

// Register in post-routing stack (after ContextMiddleware, AuthMiddleware)
routerInst.UsePostrouting(auditMiddleware)
```

---

## Handler Contribution Example

Handlers can optionally contribute semantic actions or model-level fields:

```go
// In a handler that creates a user
func (h *Handlers) CreateUser(response http.ResponseWriter, request *http.Request) {
    ctx := request.Context()

    // ... handler logic ...

    // Contribute semantic action and fields to audit log
    ctx = audit.Contribute(ctx, 
        model.AuditLogAction("CreateUser"),  // Semantic action
        map[string]any{
            "user_id": newUser.ID.String(),
            "username": newUser.Name,
        },
    )

    // Update request context so middleware can access contribution
    *request = *request.WithContext(ctx)

    // ... rest of handler logic ...
}
```

---

## Key Patterns

1. **Intent before handler:** Always written synchronously, even if it fails
2. **Result after handler:** Written asynchronously via buffered worker
3. **Error handling:** Log and swallow - never fail the request
4. **Status mapping:** `< 400` = success, `>= 400` = failure
5. **Context enrichment:** Optional, handler-controlled via `audit.Contribute()`
6. **Anonymous actors:** Missing auth context leaves actor fields empty (source IP captured)

---

## References

- **RFC 7 Section 8:** Original reference implementation
- **Module structure:** `module-structure.md`
- **Existing patterns:** `cmd/api/src/api/middleware/middleware.go`
