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

// Package modules is the central registry of feature wireup modules. The
// startup entrypoint calls RegisterAll with the shared Deps and the ordered
// list of Module functions to invoke; adding a new feature is passing its
// Register function to RegisterAll alongside the existing ones.
package modules

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/specterops/bloodhound/cmd/api/src/api/router"
)

// Deps carries the shared infrastructure that feature modules need in order to
// construct and register their store, service and handler stacks. New cross-
// cutting dependencies (graph database, filesystem, caches, etc.) are added
// here so that every module has a single, consistent place to pull from.
type Deps struct {
	Router *router.Router
	Pool   *pgxpool.Pool
}

// Module is a feature's entry point. It is invoked once during server startup
// and is responsible for building the module's store -> service -> handler
// chain from the supplied Deps and attaching any routes or other entry points
// the module exposes. A function type is sufficient because a module exposes
// no other behaviour beyond registration.
type Module func(deps Deps)

// RegisterAll registers every supplied feature module with the server,
// invoking each one in order with the shared Deps.
func RegisterAll(deps Deps, mods ...Module) {
	register(deps, mods)
}

// register iterates the supplied modules and invokes each in turn. It is the
// unit of work behind RegisterAll, separated so it can be exercised with fake
// modules in tests.
func register(deps Deps, mods []Module) {
	for _, mod := range mods {
		mod(deps)
	}
}
