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
	"github.com/specterops/bloodhound/server/analysis/appdb"
	"github.com/specterops/bloodhound/server/analysis/handlers"
	"github.com/specterops/bloodhound/server/analysis/services"
	"github.com/specterops/bloodhound/server/modules"
)

// Register builds the analysis store -> service -> handler chain and attaches
// the analysis routes to the router supplied via deps. It satisfies the
// modules.Module function type and is referenced from the modules registry.
func Register(deps modules.Deps) {
	var (
		store      = appdb.NewStore(deps.Pool)
		svc        = services.NewService(store)
		handlerSet = handlers.NewHandlersContainer(svc)
	)

	handlers.Register(deps.Router, handlerSet)
}
