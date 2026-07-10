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

package authz

import (
	"context"
	"log/slog"

	"github.com/specterops/bloodhound/cmd/api/src/auth"
	"github.com/specterops/bloodhound/cmd/api/src/bhctx"
	"github.com/specterops/bloodhound/packages/go/bhlog/attr"
	"github.com/specterops/bloodhound/server/etac"
	"github.com/specterops/bloodhound/server/graphdb/internal/services"
)

type NodeAuthorizer struct {
	etacService etac.Service
}

func NewNodeAuthorizer(etacService etac.Service) *NodeAuthorizer {
	return &NodeAuthorizer{
		etacService: etacService,
	}
}

// CanAccessNode returns true if the caller (provided via context) is authorized to access the node.
// Uses the ETAC service to check access permissions.
func (s *NodeAuthorizer) CanAccessNode(ctx context.Context, node services.Node) bool {
	var (
		modelUser, isUser = auth.GetUserFromAuthCtx(bhctx.Get(ctx).AuthCtx)
	)

	if !isUser { // Unauthenticated caller: we should never hit this. User context is populated by middleware but we deny as a precaution
		return false
	} else if environmentID, ok := node.EnvironmentID(); !ok { // Unset/malformed environment ID
		return false
	} else if hasAccess, err := s.etacService.CheckUserAccess(ctx, &modelUser, environmentID); err != nil {
		slog.ErrorContext(ctx, "Failed to check ETAC user access", attr.Error(err))
		return false
	} else {
		return hasAccess
	}
}
