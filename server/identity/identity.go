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

package identity

import (
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/specterops/bloodhound/cmd/api/src/api/router"
	"github.com/specterops/bloodhound/server/identity/internal/appdb"
	"github.com/specterops/bloodhound/server/identity/internal/handlers"
	"github.com/specterops/bloodhound/server/identity/internal/routes"
	"github.com/specterops/bloodhound/server/identity/internal/services"
)

// Register builds the identity store -> service -> handler chain and attaches
// the identity routes to the provided router.
func Register(routerInst *router.Router, pool *pgxpool.Pool, rateLimit func() mux.MiddlewareFunc) {
	var (
		store      = appdb.NewStore(pool)
		svc        = services.NewService(store)
		handlerSet = handlers.NewHandlersContainer(svc)
	)

	routes.Register(routerInst, handlerSet, rateLimit)
}

// RegisterUserListAlias builds the identity store -> service -> handler chain and
// registers the list-users endpoint at the provided path, applying the supplied
// rate-limit middleware factory when it is non-nil. It supports enterprise
// backwards-compatibility aliases (e.g. /api/v2/bhe-users) without leaking
// enterprise-specific paths into the shared identity route table.
func RegisterUserListAlias(routerInst *router.Router, pool *pgxpool.Pool, path string, rateLimitMiddleware func() mux.MiddlewareFunc) {
	var (
		store      = appdb.NewStore(pool)
		svc        = services.NewService(store)
		handlerSet = handlers.NewHandlersContainer(svc)
		route      = routes.RegisterUserListRoute(routerInst, handlerSet, path)
	)

	if rateLimitMiddleware != nil {
		router.With(rateLimitMiddleware, route)
	}
}
