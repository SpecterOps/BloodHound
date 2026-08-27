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

package router_test

import (
	"net/http"
	"testing"

	"github.com/specterops/bloodhound/cmd/api/src/api/router"
	"github.com/specterops/bloodhound/cmd/api/src/auth"
	"github.com/specterops/bloodhound/cmd/api/src/config"
	"github.com/stretchr/testify/assert"
)

func newTestRouter() router.Router {
	return router.NewRouter(config.Configuration{}, auth.NewAuthorizer(nil), "")
}

func noopHandler(http.ResponseWriter, *http.Request) {}

// TestRouter_ExcludeFromAudit verifies a route opts itself out of auditing at its
// own registration site and that IsAuditExcluded reflects only the opted-out
// templates, so the audit middleware can consult it instead of a central list.
func TestRouter_ExcludeFromAudit(t *testing.T) {
	tests := []struct {
		name         string
		register     func(routerInst *router.Router)
		template     string
		wantExcluded bool
	}{
		{
			name:         "route marked ExcludeFromAudit is excluded",
			register:     func(routerInst *router.Router) { routerInst.GET("/test/excluded", noopHandler).ExcludeFromAudit() },
			template:     "/test/excluded",
			wantExcluded: true,
		},
		{
			name:         "route not marked ExcludeFromAudit is not excluded",
			register:     func(routerInst *router.Router) { routerInst.GET("/test/audited", noopHandler) },
			template:     "/test/audited",
			wantExcluded: false,
		},
		{
			name:         "unregistered template is not excluded",
			register:     func(routerInst *router.Router) {},
			template:     "/never-registered",
			wantExcluded: false,
		},
		{
			name: "ExcludeFromAudit composes with other route builders",
			register: func(routerInst *router.Router) {
				routerInst.POST("/test/chained", noopHandler).RequireAuth().ExcludeFromAudit()
			},
			template:     "/test/chained",
			wantExcluded: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			routerInst := newTestRouter()
			tt.register(&routerInst)

			assert.Equal(t, tt.wantExcluded, routerInst.IsAuditExcluded(tt.template))
		})
	}
}
