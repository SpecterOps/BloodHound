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

// Package modules is the central registry of feature modules. The startup
// entrypoint calls Register with the shared Deps; adding a new feature module
// is a direct call to its Register function from within this package.
package modules

import (
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/specterops/bloodhound/cmd/api/src/api/middleware"
	"github.com/specterops/bloodhound/cmd/api/src/api/router"
	"github.com/specterops/bloodhound/cmd/api/src/services/dogtags"
	alerts "github.com/specterops/bloodhound/server/alerts"
	"github.com/specterops/bloodhound/server/analysis"
	"github.com/specterops/bloodhound/server/appcfg"
	"github.com/specterops/bloodhound/server/audit"
	"github.com/specterops/bloodhound/server/extensions"
	"github.com/specterops/bloodhound/server/featureflags"
	"github.com/specterops/bloodhound/server/graphdb"
	"github.com/specterops/bloodhound/server/identity"
	"github.com/specterops/dawgs/graph"
)

// Deps carries the shared infrastructure that feature modules need in order to
// construct and register their store, service and handler stacks. New cross-
// cutting dependencies (graph database, filesystem, caches, etc.) are added
// here so that every module has a single, consistent place to pull from.
type Deps struct {
	Router              *router.Router
	Pool                *pgxpool.Pool
	Graph               graph.Database
	RateLimitMiddleware func() mux.MiddlewareFunc
	DogTags             dogtags.Service
	AlertPublisher      alerts.Publisher
}

// Services carries handles that the startup entrypoint needs to keep after
// module registration. The audit Maintainer is handed to the GC daemon so it
// can manage the audit_logs range partitions.
type Services struct {
	AuditMaintainer audit.Maintainer
}

// Register wires up all feature modules with the provided infrastructure.
// Each feature module builds its own store → service → handler chain and
// attaches its routes to the shared router. It returns the Services the
// entrypoint must retain after registration.
func Register(deps Deps) Services {
	if deps.Router == nil {
		panic("modules: Register requires a non-nil Router")
	}
	if deps.Pool == nil {
		panic("modules: Register requires a non-nil Pool")
	}
	if deps.Graph == nil {
		panic("modules: Register requires a non-nil Graph")
	}
	if deps.RateLimitMiddleware == nil {
		panic("modules: Register requires a non-nil RateLimitMiddleware")
	}
	if deps.DogTags == nil {
		panic("modules: Register requires a non-nil DogTags")
	}

	if deps.AlertPublisher == nil {
		deps.AlertPublisher = alerts.NewAlertEventPublisher()
	}

	analysis.Register(deps.Router, deps.Pool)
	appcfg.Register(deps.Router, deps.Pool)
	identity.Register(deps.Router, deps.Pool)
	featureflags.Register(deps.Router, deps.Pool)
	graphdb.Register(deps.Router, deps.Pool, deps.Graph, deps.RateLimitMiddleware, deps.DogTags)
	extensions.Register(deps.Router, deps.Pool, deps.RateLimitMiddleware)

	// Audit middleware records the intent/success/failure lifecycle of every
	// request. It is attached post-routing so the authenticated actor set by the
	// auth middleware is available on the request context. Routes opt out of
	// auditing at their own registration site via Route.ExcludeFromAudit (e.g. the
	// health check), which the middleware consults through IsAuditExcluded so no
	// route strings are hardcoded here. The Maintainer is returned so the
	// entrypoint can hand it to the GC daemon to manage the audit_logs partitions.
	auditService, auditMaintainer := audit.Register(deps.Pool)
	deps.Router.UsePostrouting(middleware.AuditMiddleware(auditService, deps.Router.MuxRouter(), deps.Router.IsAuditExcluded))

	return Services{AuditMaintainer: auditMaintainer}
}
