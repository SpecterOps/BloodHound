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

// Package wireup defines the shared contract used by feature modules to
// register themselves with the server during startup. It carries no feature
// imports of its own so that both the module registry and the individual
// module packages can depend on it without creating an import cycle.
package wireup

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

// Module is implemented by a feature's wireup package. Register is invoked
// once during server startup and is responsible for building the module's
// store -> service -> handler chain from the supplied Deps and attaching any
// routes or other entry points the module exposes.
type Module interface {
	Register(deps Deps)
}
