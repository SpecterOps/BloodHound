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

package handlers

import (
	"github.com/specterops/bloodhound/cmd/api/src/api/router"
	"github.com/specterops/bloodhound/cmd/api/src/auth"
)

// Register attaches the analysis request endpoints to the given router instance.
func Register(routerInst *router.Router, handlers *Handlers) {
	var permissions = auth.Permissions()

	routerInst.GET("/api/v2/analysis", handlers.GetRequest).RequirePermissions(permissions.AppReadApplicationConfiguration)
	routerInst.PUT("/api/v2/analysis", handlers.CreateRequest).RequirePermissions(permissions.AppWriteApplicationConfiguration)
}
