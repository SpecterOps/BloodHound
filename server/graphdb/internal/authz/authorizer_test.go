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

package authz_test

import (
	"context"
	"testing"

	"github.com/specterops/bloodhound/cmd/api/src/auth"
	"github.com/specterops/bloodhound/cmd/api/src/bhctx"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/cmd/api/src/services/dogtags"
	"github.com/specterops/bloodhound/packages/go/graphschema"
	"github.com/specterops/bloodhound/packages/go/graphschema/ad"
	"github.com/specterops/bloodhound/packages/go/graphschema/azure"
	"github.com/specterops/bloodhound/packages/go/slicesext"
	"github.com/specterops/bloodhound/server/graphdb/internal/authz"
	"github.com/specterops/bloodhound/server/graphdb/internal/services"
	"github.com/stretchr/testify/assert"
)

func ctxWithUser(user model.User) context.Context {
	return bhctx.Set(context.Background(), &bhctx.Context{
		AuthCtx: auth.Context{
			Owner: user,
		},
	})
}

func ctxWithoutUser() context.Context {
	return bhctx.Set(context.Background(), &bhctx.Context{})
}

func userWithEnvironments(all bool, environmentIDs ...string) model.User {
	envs := slicesext.Map(environmentIDs, func(environmentID string) model.EnvironmentTargetedAccessControl {
		return model.EnvironmentTargetedAccessControl{
			EnvironmentID: environmentID,
		}
	})

	return model.User{
		AllEnvironments:                  all,
		EnvironmentTargetedAccessControl: envs,
	}
}

func newTestNode(kind string, envKey string, envID any) services.Node {
	return services.Node{
		Kinds: []services.Kind{
			{
				Name: kind,
			},
		},
		Properties: map[string]any{
			envKey: envID,
		},
	}
}

var (
	testDomainSID     string = "S-1-2-34-567"
	testTenantID      string = "00000000-0000-0000-0000-000000000000"
	testEnvironmentID string = "00ow0o8if0CNwsKmk697"
)

func TestNodeAuthorizer_CanAccessNode(t *testing.T) {
	tests := []struct {
		name        string
		etacEnabled bool
		ctx         context.Context
		node        services.Node
		want        bool
	}{
		{
			name:        "ETAC enabled, no user context should be denied",
			etacEnabled: true,
			ctx:         ctxWithoutUser(),
			node:        newTestNode(ad.Entity.String(), ad.DomainSID.String(), testDomainSID),
			want:        false,
		},
		{
			name:        "ETAC disabled, no user context should be denied",
			etacEnabled: false,
			ctx:         ctxWithoutUser(),
			node:        newTestNode(ad.Entity.String(), ad.DomainSID.String(), testDomainSID),
			want:        false,
		},
		{
			name:        "ETAC disabled, user context should be allowed",
			etacEnabled: false,
			ctx:         ctxWithUser(userWithEnvironments(false)),
			node:        newTestNode(ad.Entity.String(), ad.DomainSID.String(), testDomainSID),
			want:        true,
		},
		{
			name:        "ETAC enabled, user context with all environments access should be allowed",
			etacEnabled: true,
			ctx:         ctxWithUser(userWithEnvironments(true)),
			node:        newTestNode(ad.Entity.String(), ad.DomainSID.String(), testDomainSID),
			want:        true,
		},
		{
			name:        "ETAC enabled, user context with access to AD node should be allowed",
			etacEnabled: true,
			ctx:         ctxWithUser(userWithEnvironments(false, testDomainSID)),
			node:        newTestNode(ad.Entity.String(), ad.DomainSID.String(), testDomainSID),
			want:        true,
		},
		{
			name:        "ETAC enabled, user context without access to AD node should be denied",
			etacEnabled: true,
			ctx:         ctxWithUser(userWithEnvironments(false, "S-0-0-00-000")),
			node:        newTestNode(ad.Entity.String(), ad.DomainSID.String(), testDomainSID),
			want:        false,
		},
		{
			name:        "ETAC enabled, user context with access to AZ node should be allowed",
			etacEnabled: true,
			ctx:         ctxWithUser(userWithEnvironments(false, testTenantID)),
			node:        newTestNode(azure.Entity.String(), azure.TenantID.String(), testTenantID),
			want:        true,
		},
		{
			name:        "ETAC enabled, user context with access to OG node should be allowed",
			etacEnabled: true,
			ctx:         ctxWithUser(userWithEnvironments(false, testEnvironmentID)),
			node:        newTestNode("Okta", graphschema.EnvironmentIDKey, testEnvironmentID),
			want:        true,
		},
		{
			name:        "ETAC enabled, node missing an environment ID should be denied",
			etacEnabled: true,
			ctx:         ctxWithUser(userWithEnvironments(false, "NoID4Me")),
			node:        newTestNode("Okta", "NoEnvironmentID", "NoID4Me"),
			want:        false,
		},
		{
			name:        "ETAC enabled, node with malformed environment ID should be denied",
			etacEnabled: true,
			ctx:         ctxWithUser(userWithEnvironments(false, "42")),
			node:        newTestNode("Okta", "BadEnvironmentID", nil),
			want:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				dogTags = dogtags.NewTestService(dogtags.TestOverrides{
					Bools: map[dogtags.BoolDogTag]bool{
						dogtags.ETAC_ENABLED: tt.etacEnabled,
					},
				})
				authorizer = authz.NewNodeAuthorizer(dogTags)
			)

			assert.Equal(t, tt.want, authorizer.CanAccessNode(tt.ctx, tt.node))
		})
	}
}
