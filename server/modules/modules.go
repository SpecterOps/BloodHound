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
// startup entrypoint asks this package to register every module against the
// supplied wireup.Deps; adding a new feature is a single slice entry here
// plus a new modules/<domain>/module.go file.
package modules

import (
	"github.com/specterops/bloodhound/server/analysis"
	"github.com/specterops/bloodhound/server/wireup"
)

// all is the ordered list of feature modules registered at server startup.
// Order matters only when a later module depends on something a prior module
// has attached to the router (e.g. middleware) — the analysis module has no
// such dependency today.
var all = []wireup.Module{
	analysis.Register,
}

// RegisterAll registers every configured feature module with the server,
// supplying each one the shared wireup.Deps.
func RegisterAll(deps wireup.Deps) {
	register(deps, all)
}

// register iterates the supplied modules and invokes each in turn. It is the
// unit of work behind RegisterAll, separated so it can be exercised with fake
// modules in tests.
func register(deps wireup.Deps, modules []wireup.Module) {
	for _, module := range modules {
		module(deps)
	}
}
