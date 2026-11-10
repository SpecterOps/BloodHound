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
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/specterops/bloodhound/cmd/api/src/api/router"
	"github.com/specterops/bloodhound/server/identity/internal/appdb"
	"github.com/specterops/bloodhound/server/identity/internal/handlers"
	"github.com/specterops/bloodhound/server/identity/internal/routes"
	"github.com/specterops/bloodhound/server/identity/internal/services"
)

type AuthToken = services.AuthToken

// AuthTokenReader allows applications to enforce additional token visibility
// rules without replacing identity's requester authorization and scoping.
type AuthTokenReader = services.AuthTokenReader

// Option configures the dependencies used by the identity routes.
type Option = services.Option

// WithAuthTokenReader supplies an application-specific token reader
func WithAuthTokenReader(reader AuthTokenReader) Option {
	return services.WithAuthTokenReader(reader)
}

// Register builds the identity store -> service -> handler chain and attaches
// the identity routes to the provided router.
func Register(routerInst *router.Router, pool *pgxpool.Pool, options ...Option) {
	var (
		store      = appdb.NewStore(pool)
		svc        = services.NewService(store, options...)
		handlerSet = handlers.NewHandlersContainer(svc)
	)

	routes.Register(routerInst, handlerSet)
}

// RegisterUserListAlias registers the list-users endpoint at an enterprise
// backwards-compatibility path (e.g. /api/v2/bhe-users).
func RegisterUserListAlias(routerInst *router.Router, pool *pgxpool.Pool, path string) {
	var (
		store      = appdb.NewStore(pool)
		svc        = services.NewService(store)
		handlerSet = handlers.NewHandlersContainer(svc)
	)
	routes.RegisterUserListRoute(routerInst, handlerSet, path)
}
