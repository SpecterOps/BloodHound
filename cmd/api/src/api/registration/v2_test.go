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

package registration_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/specterops/bloodhound/cmd/api/src/api/registration"
	"github.com/specterops/bloodhound/cmd/api/src/api/router"
	v2 "github.com/specterops/bloodhound/cmd/api/src/api/v2"
	"github.com/specterops/bloodhound/cmd/api/src/auth"
	"github.com/specterops/bloodhound/cmd/api/src/bhctx"
	"github.com/specterops/bloodhound/cmd/api/src/config"
	dbmocks "github.com/specterops/bloodhound/cmd/api/src/database/mocks"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/cmd/api/src/model/appcfg"
	graphmocks "github.com/specterops/bloodhound/cmd/api/src/queries/mocks"
	"github.com/specterops/bloodhound/cmd/api/src/services/dogtags"
	"github.com/specterops/bloodhound/cmd/api/src/utils/test"
	"github.com/specterops/bloodhound/packages/go/graphschema/ad"
	"github.com/specterops/bloodhound/packages/go/graphschema/common"
	"github.com/specterops/dawgs/graph"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func userContext(userPermissions ...model.Permission) *bhctx.Context {
	return &bhctx.Context{
		AuthCtx: auth.Context{
			Owner: model.User{
				Roles: model.Roles{{Name: "Test Role", Permissions: userPermissions}},
			},
		},
	}
}

func TestNewV2API_PatchDomainRequiresGraphDBWrite(t *testing.T) {
	const domainObjectID = "S-1-5-21-1-2-3"

	var (
		mockCtrl    = gomock.NewController(t)
		mockDB      = dbmocks.NewMockDatabase(mockCtrl)
		mockGraph   = graphmocks.NewMockGraph(mockCtrl)
		permissions = auth.Permissions()
		routerInst  = router.NewRouter(config.Configuration{}, auth.NewAuthorizer(mockDB), "")
		resources   = v2.Resources{
			DB:         mockDB,
			GraphQuery: mockGraph,
			DogTags:    dogtags.NewTestService(dogtags.TestOverrides{}),
		}
	)
	defer mockCtrl.Finish()

	registration.NewV2API(resources, &routerInst)

	t.Run("user with only GraphDBRead is forbidden", func(t *testing.T) {
		mockDB.EXPECT().AppendAuditLog(gomock.Any(), gomock.Any()).Return(nil)

		test.Request(t).
			WithMethod(http.MethodPatch).
			WithURL("/api/v2/domains/%s", domainObjectID).
			WithBody(map[string]any{"collected": true}).
			WithContext(userContext(permissions.GraphDBRead)).
			OnHandler(routerInst.MuxRouter()).
			Require().
			ResponseStatusCode(http.StatusForbidden)
	})

	t.Run("user with GraphDBWrite may update collected", func(t *testing.T) {
		mockDB.EXPECT().GetConfigurationParameter(gomock.Any(), gomock.Any()).Return(appcfg.Parameter{}, nil).AnyTimes()
		mockGraph.EXPECT().GetEntityByObjectId(gomock.Any(), domainObjectID, ad.Domain).Return(graph.NewNode(graph.ID(1), graph.NewProperties()), nil)
		mockGraph.EXPECT().BatchNodeUpdate(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, nodeUpdate graph.NodeUpdate) error {
			collected, err := nodeUpdate.Node.Properties.Get(common.Collected.String()).Bool()
			require.NoError(t, err)
			require.True(t, collected)
			require.Equal(t, ad.Domain, nodeUpdate.IdentityKind)

			return nil
		})

		test.Request(t).
			WithMethod(http.MethodPatch).
			WithURL("/api/v2/domains/%s", domainObjectID).
			WithBody(map[string]any{"collected": true}).
			WithContext(userContext(permissions.GraphDBWrite)).
			OnHandler(routerInst.MuxRouter()).
			Require().
			ResponseStatusCode(http.StatusOK)
	})
}
