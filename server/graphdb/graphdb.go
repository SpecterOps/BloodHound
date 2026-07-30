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

// Package graphdb is the wireup module for the graph database feature. It is the single
// place where the graphdb store, service, handlers and routes are composed; the layered
// subpackages themselves remain unaware of each other.
package graphdb

import (
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/specterops/bloodhound/cmd/api/src/api/router"
	"github.com/specterops/bloodhound/cmd/api/src/services/dogtags"
	"github.com/specterops/bloodhound/server/etac"
	"github.com/specterops/bloodhound/server/graphdb/internal/appdb"
	"github.com/specterops/bloodhound/server/graphdb/internal/authz"
	"github.com/specterops/bloodhound/server/graphdb/internal/handlers"
	"github.com/specterops/bloodhound/server/graphdb/internal/routes"
	"github.com/specterops/bloodhound/server/graphdb/internal/services"
	"github.com/specterops/dawgs/graph"
)

// Register builds the graphdb store -> service -> handler chain and attaches the graphdb
// routes to the provided router. It is called from the modules registry and receives
// only the infrastructure it directly needs: the router, the pgx pool (for kind
// resolution), the graph database (for graph reads) and the rate limit middleware
// factory applied to the registered routes.
func Register(routerInst *router.Router, pool *pgxpool.Pool, graphDatabase graph.Database, rateLimit func() mux.MiddlewareFunc, dogTags dogtags.Service) {
	var (
		store          = appdb.NewStore(graphDatabase, pool)
		service        = services.NewService(store)
		etacService    = etac.Register(pool, dogTags)
		nodeAuthorizer = authz.NewNodeAuthorizer(etacService)
		handlerSet     = handlers.NewHandlersContainer(service, nodeAuthorizer)
	)

	routes.Register(routerInst, handlerSet, rateLimit)
}
