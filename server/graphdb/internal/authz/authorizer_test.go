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
	"github.com/specterops/bloodhound/packages/go/graphschema"
	"github.com/specterops/bloodhound/packages/go/graphschema/ad"
	"github.com/specterops/bloodhound/packages/go/graphschema/azure"
	"github.com/specterops/bloodhound/packages/go/slicesext"
	etacMocks "github.com/specterops/bloodhound/server/etac/mocks"
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
		name      string
		ctx       context.Context
		node      services.Node
		want      bool
		mockSetup func(*etacMocks.MockService, context.Context, services.Node)
	}{
		{
			name: "no user context should be denied",
			ctx:  ctxWithoutUser(),
			node: newTestNode(ad.Entity.String(), ad.DomainSID.String(), testDomainSID),
			want: false,
			mockSetup: func(m *etacMocks.MockService, ctx context.Context, node services.Node) {
				// No mock calls expected because we short-circuit on missing user
			},
		},
		{
			name: "node missing environment ID should be denied",
			ctx:  ctxWithUser(userWithEnvironments(false)),
			node: newTestNode("Okta", "NoEnvironmentID", "NoID4Me"),
			want: false,
			mockSetup: func(m *etacMocks.MockService, ctx context.Context, node services.Node) {
				user := userWithEnvironments(false)
				m.EXPECT().ShouldFilterForETAC(&user).Return(true)
			},
		},
		{
			name: "node with malformed environment ID should be denied",
			ctx:  ctxWithUser(userWithEnvironments(false, "42")),
			node: newTestNode("Okta", "BadEnvironmentID", nil),
			want: false,
			mockSetup: func(m *etacMocks.MockService, ctx context.Context, node services.Node) {
				user := userWithEnvironments(false, "42")
				m.EXPECT().ShouldFilterForETAC(&user).Return(true)
			},
		},
		{
			name: "ETAC service returns true for allowed AD node",
			ctx:  ctxWithUser(userWithEnvironments(false, testDomainSID)),
			node: newTestNode(ad.Entity.String(), ad.DomainSID.String(), testDomainSID),
			want: true,
			mockSetup: func(m *etacMocks.MockService, ctx context.Context, node services.Node) {
				user := userWithEnvironments(false, testDomainSID)
				m.EXPECT().ShouldFilterForETAC(&user).Return(true)
				m.EXPECT().CheckUserAccess(ctx, &user, []string{testDomainSID}).Return(true, nil)
			},
		},
		{
			name: "ETAC service returns false for denied AD node",
			ctx:  ctxWithUser(userWithEnvironments(false, "S-0-0-00-000")),
			node: newTestNode(ad.Entity.String(), ad.DomainSID.String(), testDomainSID),
			want: false,
			mockSetup: func(m *etacMocks.MockService, ctx context.Context, node services.Node) {
				user := userWithEnvironments(false, "S-0-0-00-000")
				m.EXPECT().ShouldFilterForETAC(&user).Return(true)
				m.EXPECT().CheckUserAccess(ctx, &user, []string{testDomainSID}).Return(false, nil)
			},
		},
		{
			name: "ETAC service returns true for user with all environments",
			ctx:  ctxWithUser(userWithEnvironments(true)),
			node: newTestNode(ad.Entity.String(), ad.DomainSID.String(), testDomainSID),
			want: true,
			mockSetup: func(m *etacMocks.MockService, ctx context.Context, node services.Node) {
				user := userWithEnvironments(true)
				m.EXPECT().ShouldFilterForETAC(&user).Return(false)
			},
		},
		{
			name: "ETAC service returns true for allowed AZ node",
			ctx:  ctxWithUser(userWithEnvironments(false, testTenantID)),
			node: newTestNode(azure.Entity.String(), azure.TenantID.String(), testTenantID),
			want: true,
			mockSetup: func(m *etacMocks.MockService, ctx context.Context, node services.Node) {
				user := userWithEnvironments(false, testTenantID)
				m.EXPECT().ShouldFilterForETAC(&user).Return(true)
				m.EXPECT().CheckUserAccess(ctx, &user, []string{testTenantID}).Return(true, nil)
			},
		},
		{
			name: "ETAC service returns true for allowed OG node",
			ctx:  ctxWithUser(userWithEnvironments(false, testEnvironmentID)),
			node: newTestNode("Okta", graphschema.EnvironmentIDKey, testEnvironmentID),
			want: true,
			mockSetup: func(m *etacMocks.MockService, ctx context.Context, node services.Node) {
				user := userWithEnvironments(false, testEnvironmentID)
				m.EXPECT().ShouldFilterForETAC(&user).Return(true)
				m.EXPECT().CheckUserAccess(ctx, &user, []string{testEnvironmentID}).Return(true, nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				mockETAC   = etacMocks.NewMockService(t)
				authorizer = authz.NewNodeAuthorizer(mockETAC)
			)

			if tt.mockSetup != nil {
				tt.mockSetup(mockETAC, tt.ctx, tt.node)
			}

			assert.Equal(t, tt.want, authorizer.CanAccessNode(tt.ctx, tt.node))
		})
	}
}
