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

// Package appcfg is the wireup module for the application configuration feature.
// It is the single place where the appcfg store, service, handlers and routes are
// composed; the layered subpackages themselves remain unaware of each other.
package appcfg

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/specterops/bloodhound/cmd/api/src/api/router"
	"github.com/specterops/bloodhound/server/appcfg/internal/appdb"
	"github.com/specterops/bloodhound/server/appcfg/internal/handlers"
	"github.com/specterops/bloodhound/server/appcfg/internal/routes"
	"github.com/specterops/bloodhound/server/appcfg/internal/services"
)

// Register builds the appcfg store -> service -> handler chain and attaches
// the appcfg routes to the provided router. It is called from the modules
// registry and receives only the infrastructure it directly needs.
func Register(routerInst *router.Router, pool *pgxpool.Pool) {
	var (
		store   = appdb.NewStore(pool)
		service = services.NewService(store)
		hdl     = handlers.NewHandlers(service)
	)

	routes.Register(routerInst, hdl)
}
