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

// Package analysis is the wireup module for the analysis feature. It is the
// single place where the analysis store, service, handlers and routes are
// composed; the layered subpackages themselves remain unaware of each other.
package analysis

import (
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/specterops/bloodhound/cmd/api/src/api/router"
	"github.com/specterops/bloodhound/server/analysis/internal/appdb"
	"github.com/specterops/bloodhound/server/analysis/internal/handlers"
	"github.com/specterops/bloodhound/server/analysis/internal/routes"
	"github.com/specterops/bloodhound/server/analysis/internal/services"
)

// AnalysisRequestAdapter is an interface for handling analysis request operations.
// It provides methods for retrieving analysis requests via HTTP.
type AnalysisRequestAdapter interface {
	CreateAnalysisRequest(response http.ResponseWriter, request *http.Request)
}

// NewAnalysisRequestAdapter creates a new AnalysisRequestAdapter by wiring up the
// analysis store, service, and handlers. It accepts a PostgreSQL connection pool
// and returns a fully initialized adapter ready to handle analysis requests.
func NewAnalysisRequestAdapter(pool *pgxpool.Pool) AnalysisRequestAdapter {
	var (
		store      = appdb.NewStore(pool)
		svc        = services.NewService(store)
		handlerSet = handlers.NewHandlersContainer(svc)
	)

	return handlerSet
}

// Register builds the analysis store -> service -> handler chain and attaches
// the analysis routes to the provided router. It is called from the modules
// registry and receives only the infrastructure it directly needs.
func Register(routerInst *router.Router, pool *pgxpool.Pool) {
	var (
		store      = appdb.NewStore(pool)
		svc        = services.NewService(store)
		handlerSet = handlers.NewHandlersContainer(svc)
	)

	routes.Register(routerInst, handlerSet)
}
