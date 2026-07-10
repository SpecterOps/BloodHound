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
	"slices"

	v2 "github.com/specterops/bloodhound/cmd/api/src/api/v2"
	"github.com/specterops/bloodhound/cmd/api/src/auth"
	"github.com/specterops/bloodhound/cmd/api/src/bhctx"
	"github.com/specterops/bloodhound/cmd/api/src/services/dogtags"
	"github.com/specterops/bloodhound/server/graphdb/internal/services"
)

type NodeAuthorizer struct {
	dogTags dogtags.Service
}

func NewNodeAuthorizer(dogTags dogtags.Service) *NodeAuthorizer {
	return &NodeAuthorizer{
		dogTags: dogTags,
	}
}

// CanAccessNode returns true if the caller (provided via context) is authorized to access the node.
// Reusing ShouldFilterForETAC to short-circuit if the feature flag is off or if a user has access to all environments.
func (s *NodeAuthorizer) CanAccessNode(ctx context.Context, node services.Node) bool {
	var (
		user, isUser = auth.GetUserFromAuthCtx(bhctx.Get(ctx).AuthCtx)
	)

	if !isUser { // Unauthenticated caller: we should never hit this. User context is populated by middleware but we deny as a precaution
		return false
	} else if !v2.ShouldFilterForETAC(s.dogTags, user) { // ETAC disabled
		return true
	} else if environmentID, ok := node.EnvironmentID(); !ok { // ETAC enabled but unset/malformed environment ID
		return false
	} else {
		return slices.Contains(v2.ExtractEnvironmentIDsFromUser(&user), environmentID)
	}
}
