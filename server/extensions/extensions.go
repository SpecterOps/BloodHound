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
// Package extensions is the wireup module for the OpenGraph extensions feature.
// It is the single place where the extensions store, service, handlers and routes
// are composed; the layered subpackages remain unaware of each other.
package extensions

import (
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/specterops/bloodhound/cmd/api/src/api/router"
	"github.com/specterops/bloodhound/server/extensions/internal/appdb"
	"github.com/specterops/bloodhound/server/extensions/internal/handlers"
	"github.com/specterops/bloodhound/server/extensions/internal/routes"
	"github.com/specterops/bloodhound/server/extensions/internal/services"
)

// Register builds the extensions store -> service -> handler chain and attaches the
// extensions routes to the provided router. It is called from the modules registry and
// receives only the infrastructure it directly needs: the router, the pgx pool (for
// schema reads) and the rate limit middleware factory applied to the registered routes.
func Register(routerInst *router.Router, pool *pgxpool.Pool, rateLimit func() mux.MiddlewareFunc) {
	var (
		store      = appdb.NewStore(pool)
		service    = services.NewService(store)
		handlerSet = handlers.NewHandlersContainer(service)
	)

	routes.Register(routerInst, handlerSet, rateLimit)
}
