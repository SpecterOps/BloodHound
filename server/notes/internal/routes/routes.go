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
package routes

import (
	"github.com/specterops/bloodhound/cmd/api/src/api/router"
	"github.com/specterops/bloodhound/cmd/api/src/auth"
	"github.com/specterops/bloodhound/server/notes/internal/handlers"
)

func Register(routerInst *router.Router, handlers *handlers.Handlers) {
	var permissions = auth.Permissions()

	routerInst.GET("/api/v2/red-team-notes", handlers.ListNotes).RequirePermissions(permissions.GraphDBRead)
	routerInst.GET("/api/v2/red-team-notes/tags", handlers.ListTags).RequirePermissions(permissions.GraphDBRead)
	routerInst.POST("/api/v2/red-team-notes/attachments", handlers.UploadAttachment).RequirePermissions(permissions.GraphDBWrite)
	routerInst.GET("/api/v2/red-team-notes/attachments/{attachment_id}", handlers.GetAttachment).RequirePermissions(permissions.GraphDBRead)
	routerInst.GET("/api/v2/red-team-notes/media/{attachment_token}", handlers.GetMedia)
	routerInst.GET("/api/v2/red-team-notes/{note_id}", handlers.GetNote).RequirePermissions(permissions.GraphDBRead)
	routerInst.POST("/api/v2/red-team-notes", handlers.CreateNote).RequirePermissions(permissions.GraphDBWrite)
	routerInst.PUT("/api/v2/red-team-notes/{note_id}", handlers.UpdateNote).RequirePermissions(permissions.GraphDBWrite)
	routerInst.DELETE("/api/v2/red-team-notes/{note_id}", handlers.DeleteNote).RequirePermissions(permissions.GraphDBWrite)
}
