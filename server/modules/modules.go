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
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/specterops/bloodhound/cmd/api/src/api/router"
	"github.com/specterops/bloodhound/server/analysis"
	"github.com/specterops/bloodhound/server/appcfg"
)

// Deps carries the shared infrastructure that feature modules need in order to
// construct and register their store, service and handler stacks. New cross-
// cutting dependencies (graph database, filesystem, caches, etc.) are added
// here so that every module has a single, consistent place to pull from.
type Deps struct {
	Router *router.Router
	Pool   *pgxpool.Pool
}

// Register wires up all feature modules with the provided infrastructure.
// Each feature module builds its own store → service → handler chain and
// attaches its routes to the shared router.
func Register(deps Deps) {
	if deps.Router == nil {
		panic("modules: Register requires a non-nil Router")
	}
	if deps.Pool == nil {
		panic("modules: Register requires a non-nil Pool")
	}

	analysis.Register(deps.Router, deps.Pool)
	appcfg.Register(deps.Router, deps.Pool)
}
