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

// Package auth is the wireup module for the auth feature. It is the
// single place where the auth store, service, handlers and routes are
// composed; the layered subpackages themselves remain unaware of each other.
package auth

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/specterops/bloodhound/cmd/api/src/api/router"
	"github.com/specterops/bloodhound/server/auth/internal/appdb"
	"github.com/specterops/bloodhound/server/auth/internal/handlers"
	"github.com/specterops/bloodhound/server/auth/internal/routes"
	"github.com/specterops/bloodhound/server/auth/internal/services"
)

// Register builds the auth store -> service -> handler chain and attaches
// the auth routes to the provided router.
func Register(routerInst *router.Router, pool *pgxpool.Pool) {
	var (
		store      = appdb.NewStore(pool)
		svc        = services.NewService(store)
		handlerSet = handlers.NewHandlersContainer(svc)
	)

	routes.Register(routerInst, handlerSet)
}
