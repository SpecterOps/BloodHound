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

package azure

import (
	"testing"

	graph_mocks "github.com/specterops/bloodhound/cmd/api/src/vendormocks/dawgs/graph"
	azschema "github.com/specterops/bloodhound/packages/go/graphschema/azure"
	"github.com/specterops/bloodhound/packages/go/graphschema/common"
	"github.com/specterops/dawgs/graph"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

// TestFetchServicePrincipalApplications_ReturnsApplicationNodes guards against
// regressions of issue #1846: fetching the applications that run as a service
// principal must traverse the RunsAs relationship from its start node (the App),
// not the end node (the Service Principal itself). Returning the service
// principal caused the node entity panel to display the service principal's
// object ID in the App ID property.
func TestFetchServicePrincipalApplications_ReturnsApplicationNodes(t *testing.T) {
	var (
		ctrl   = gomock.NewController(t)
		mockTx = graph_mocks.NewMockTransaction(ctrl)
		app    = &graph.Node{
			ID:         1,
			Kinds:      graph.Kinds{azschema.Entity, azschema.App},
			Properties: graph.NewProperties(),
		}
		servicePrincipal = &graph.Node{
			ID:         2,
			Kinds:      graph.Kinds{azschema.Entity, azschema.ServicePrincipal},
			Properties: graph.NewProperties(),
		}
	)
	defer ctrl.Finish()

	mockRelQuery1 := graph_mocks.NewMockRelationshipQuery(ctrl)
	mockTx.EXPECT().Relationships().Return(mockRelQuery1).Times(1)

	mockRelQuery2 := graph_mocks.NewMockRelationshipQuery(ctrl)
	mockRelQuery1.EXPECT().Filterf(gomock.AssignableToTypeOf(func() graph.Criteria { return nil })).Return(mockRelQuery2)

	mockRelQuery2.EXPECT().
		FetchDirection(graph.DirectionOutbound, gomock.AssignableToTypeOf(func(graph.Cursor[graph.DirectionalResult]) error { return nil })).
		DoAndReturn(func(_ any, delegate func(graph.Cursor[graph.DirectionalResult]) error) error {
			mockCursor := graph_mocks.NewMockCursor[graph.DirectionalResult](ctrl)
			c := make(chan graph.DirectionalResult, 1)
			go func() {
				defer close(c)
				c <- graph.DirectionalResult{
					Relationship: &graph.Relationship{StartID: app.ID, EndID: servicePrincipal.ID, Kind: azschema.RunsAs},
					Node:         app,
				}
			}()
			mockCursor.EXPECT().Chan().Return(c)
			mockCursor.EXPECT().Error().Return(nil)
			return delegate(mockCursor)
		})

	apps, err := FetchServicePrincipalApplications(mockTx, servicePrincipal)
	require.Nil(t, err)
	assert.Equal(t, 1, apps.Len())
	assert.Contains(t, apps.Slice(), app)
	assert.NotContains(t, apps.Slice(), servicePrincipal)
}

// TestGetServicePrincipalAppID_ReturnsApplicationObjectID verifies that the
// value surfaced for the service principal App ID property is read from the
// linked application node, whose object ID carries the application ID
// (see ConvertAZAppToNode), rather than from the service principal itself.
func TestGetServicePrincipalAppID_ReturnsApplicationObjectID(t *testing.T) {
	var (
		ctrl   = gomock.NewController(t)
		mockTx = graph_mocks.NewMockTransaction(ctrl)
		app    = &graph.Node{
			ID:    1,
			Kinds: graph.Kinds{azschema.Entity, azschema.App},
			Properties: graph.AsProperties(graph.PropertyMap{
				common.ObjectID: "application-id-00000000-0000-0000-0000-000000000000",
			}),
		}
		servicePrincipal = &graph.Node{
			ID:         2,
			Kinds:      graph.Kinds{azschema.Entity, azschema.ServicePrincipal},
			Properties: graph.NewProperties(),
		}
	)
	defer ctrl.Finish()

	mockRelQuery1 := graph_mocks.NewMockRelationshipQuery(ctrl)
	mockTx.EXPECT().Relationships().Return(mockRelQuery1).Times(1)

	mockRelQuery2 := graph_mocks.NewMockRelationshipQuery(ctrl)
	mockRelQuery1.EXPECT().Filterf(gomock.AssignableToTypeOf(func() graph.Criteria { return nil })).Return(mockRelQuery2)

	mockRelQuery2.EXPECT().
		FetchDirection(graph.DirectionOutbound, gomock.AssignableToTypeOf(func(graph.Cursor[graph.DirectionalResult]) error { return nil })).
		DoAndReturn(func(_ any, delegate func(graph.Cursor[graph.DirectionalResult]) error) error {
			mockCursor := graph_mocks.NewMockCursor[graph.DirectionalResult](ctrl)
			c := make(chan graph.DirectionalResult, 1)
			go func() {
				defer close(c)
				c <- graph.DirectionalResult{
					Relationship: &graph.Relationship{StartID: app.ID, EndID: servicePrincipal.ID, Kind: azschema.RunsAs},
					Node:         app,
				}
			}()
			mockCursor.EXPECT().Chan().Return(c)
			mockCursor.EXPECT().Error().Return(nil)
			return delegate(mockCursor)
		})

	appID, err := getServicePrincipalAppID(mockTx, servicePrincipal)
	require.Nil(t, err)
	assert.Equal(t, "application-id-00000000-0000-0000-0000-000000000000", appID)
}
