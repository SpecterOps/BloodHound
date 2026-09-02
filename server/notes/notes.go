// Copyright 2026 Specter Ops, Inc.
//
// Licensed under the Apache License, Version 2.0
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

// Package notes is a self-contained red team notes feature module. It owns the
// notes domain (the Note type, the Database port and the Service), the
// PostgreSQL adapter (Store) and the Register entry point that wires them
// together so callers obtain a ready-to-use HTTP surface without reaching into
// the storage layer.
package notes

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/specterops/bloodhound/cmd/api/src/api/router"
	"github.com/specterops/bloodhound/server/notes/internal/appdb"
	"github.com/specterops/bloodhound/server/notes/internal/handlers"
	"github.com/specterops/bloodhound/server/notes/internal/routes"
	"github.com/specterops/bloodhound/server/notes/internal/services"
)

func Register(routerInst *router.Router, pool *pgxpool.Pool) {
	var (
		store      = appdb.NewStore(pool)
		svc        = services.NewService(store)
		handlerSet = handlers.NewHandlersContainer(svc)
	)

	routes.Register(routerInst, handlerSet)
}
