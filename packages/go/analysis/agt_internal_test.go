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
package analysis

import (
	"context"
	"errors"
	"testing"

	dbmocks "github.com/specterops/bloodhound/cmd/api/src/database/mocks"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	graphmocks "github.com/specterops/bloodhound/cmd/api/src/vendormocks/dawgs/graph"
	"github.com/specterops/bloodhound/packages/go/graphschema"
	"github.com/specterops/bloodhound/packages/go/graphschema/common"
	"github.com/specterops/dawgs/graph"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

// expectReadTransaction expects a read transaction and runs its delegate with the mock transaction.
func expectReadTransaction(graphDB *graphmocks.MockDatabase, transaction graph.Transaction) *gomock.Call {
	return graphDB.EXPECT().ReadTransaction(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, delegate graph.TransactionDelegate, _ ...graph.TransactionOption) error {
		return delegate(transaction)
	})
}

// expectZoneNodeFetch returns nodes or the test-provided fetch error.
func expectZoneNodeFetch(ctrl *gomock.Controller, transaction *graphmocks.MockTransaction, zoneNodes []*graph.Node, fetchError error) {
	var (
		nodeQuery         = graphmocks.NewMockNodeQuery(ctrl)
		filteredNodeQuery = graphmocks.NewMockNodeQuery(ctrl)
	)

	transaction.EXPECT().Nodes().Return(nodeQuery)
	nodeQuery.EXPECT().Filter(gomock.Any()).Return(filteredNodeQuery)

	if fetchError != nil {
		filteredNodeQuery.EXPECT().Fetch(gomock.Any()).Return(fetchError)
		return
	}

	var (
		cursor      = graphmocks.NewMockCursor[*graph.Node](ctrl)
		nodeChannel = make(chan *graph.Node, len(zoneNodes))
	)
	for _, zoneNode := range zoneNodes {
		nodeChannel <- zoneNode
	}
	close(nodeChannel)

	cursor.EXPECT().Chan().Return(nodeChannel)
	cursor.EXPECT().Error().Return(nil)
	filteredNodeQuery.EXPECT().Fetch(gomock.Any()).DoAndReturn(func(delegate func(graph.Cursor[*graph.Node]) error, _ ...graph.Criteria) error {
		return delegate(cursor)
	})
}

// expectNodeIDFetch returns IDs or the test-provided fetch error.
func expectNodeIDFetch(ctrl *gomock.Controller, transaction *graphmocks.MockTransaction, nodeIDs []graph.ID, fetchError error) {
	var (
		nodeQuery         = graphmocks.NewMockNodeQuery(ctrl)
		filteredNodeQuery = graphmocks.NewMockNodeQuery(ctrl)
	)

	transaction.EXPECT().Nodes().Return(nodeQuery)
	nodeQuery.EXPECT().Filter(gomock.Any()).Return(filteredNodeQuery)

	if fetchError != nil {
		filteredNodeQuery.EXPECT().FetchIDs(gomock.Any()).Return(fetchError)
		return
	}

	var (
		cursor        = graphmocks.NewMockCursor[graph.ID](ctrl)
		nodeIDChannel = make(chan graph.ID, len(nodeIDs))
	)
	for _, nodeID := range nodeIDs {
		nodeIDChannel <- nodeID
	}
	close(nodeIDChannel)

	cursor.EXPECT().Chan().Return(nodeIDChannel)
	cursor.EXPECT().Error().Return(nil)
	filteredNodeQuery.EXPECT().FetchIDs(gomock.Any()).DoAndReturn(func(delegate func(graph.Cursor[graph.ID]) error) error {
		return delegate(cursor)
	})
}

// expectRelationshipFetch returns relationships or the test-provided fetch error.
func expectRelationshipFetch(ctrl *gomock.Controller, transaction *graphmocks.MockTransaction, relationships []*graph.Relationship, fetchError error) {
	var (
		relationshipQuery         = graphmocks.NewMockRelationshipQuery(ctrl)
		filteredRelationshipQuery = graphmocks.NewMockRelationshipQuery(ctrl)
	)

	transaction.EXPECT().Relationships().Return(relationshipQuery)
	relationshipQuery.EXPECT().Filter(gomock.Any()).Return(filteredRelationshipQuery)

	if fetchError != nil {
		filteredRelationshipQuery.EXPECT().Fetch(gomock.Any()).Return(fetchError)
		return
	}

	var (
		cursor              = graphmocks.NewMockCursor[*graph.Relationship](ctrl)
		relationshipChannel = make(chan *graph.Relationship, len(relationships))
	)
	for _, relationship := range relationships {
		relationshipChannel <- relationship
	}
	close(relationshipChannel)

	cursor.EXPECT().Chan().Return(relationshipChannel)
	cursor.EXPECT().Error().Return(nil)
	filteredRelationshipQuery.EXPECT().Fetch(gomock.Any()).DoAndReturn(func(delegate func(graph.Cursor[*graph.Relationship]) error) error {
		return delegate(cursor)
	})
}

// expectBatchOperation expects a batch operation and runs its delegate with the mock batch.
func expectBatchOperation(graphDB *graphmocks.MockDatabase, batch graph.Batch) *gomock.Call {
	return graphDB.EXPECT().BatchOperation(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, delegate graph.BatchDelegate, _ ...graph.BatchOption) error {
		return delegate(batch)
	})
}

func testZoneNode(id graph.ID, zone model.AssetGroupTag) *graph.Node {
	return graph.NewNode(id, graph.AsProperties(zoneNodePropertyMap(zone)), graphschema.Zone)
}

func TestGetZones(t *testing.T) {
	t.Parallel()

	var (
		testError = errors.New("test error")
		zones     = model.AssetGroupTags{{ID: 1, Name: "Zone"}}
	)
	testCases := []struct {
		name          string
		zones         model.AssetGroupTags
		databaseError error
		expectedError string
	}{
		{
			name:  "Success: returns zones",
			zones: zones,
		},
		{
			name:          "Error: wraps database errors",
			databaseError: testError,
			expectedError: "get zones: test error",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var (
				ctrl         = gomock.NewController(t)
				databaseMock = dbmocks.NewMockDatabase(ctrl)
			)

			databaseMock.EXPECT().GetAssetGroupTags(gomock.Any(), model.SQLFilter{
				SQLString: "type = ?",
				Params:    []any{model.AssetGroupTagTypeTier},
			}).Return(testCase.zones, testCase.databaseError)

			actualZones, err := getZones(context.Background(), databaseMock)

			if testCase.expectedError == "" {
				require.NoError(t, err)
				assert.Equal(t, testCase.zones, actualZones)
			} else {
				assert.EqualError(t, err, testCase.expectedError)
				assert.Nil(t, actualZones)
			}
		})
	}
}

func TestGetZoneNodes(t *testing.T) {
	t.Parallel()

	var (
		testError = errors.New("test error")
		zoneNode  = graph.NewNode(1, graph.NewProperties(), graphschema.Zone)
	)
	testCases := []struct {
		name             string
		transactionError error
		fetchError       error
		expectedNodes    []*graph.Node
		expectedError    string
	}{
		{
			name:          "Success: returns zone nodes",
			expectedNodes: []*graph.Node{zoneNode},
		},
		{
			name:             "Error: wraps transaction errors",
			transactionError: testError,
			expectedError:    "get zone nodes: test error",
		},
		{
			name:          "Error: wraps node fetch errors",
			fetchError:    testError,
			expectedError: "get zone nodes: test error",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var (
				ctrl        = gomock.NewController(t)
				graphDBMock = graphmocks.NewMockDatabase(ctrl)
			)

			if testCase.transactionError != nil {
				graphDBMock.EXPECT().ReadTransaction(gomock.Any(), gomock.Any()).Return(testCase.transactionError)
			} else {
				transaction := graphmocks.NewMockTransaction(ctrl)
				expectZoneNodeFetch(ctrl, transaction, testCase.expectedNodes, testCase.fetchError)
				expectReadTransaction(graphDBMock, transaction)
			}

			actualNodes, err := getZoneNodes(context.Background(), graphDBMock)

			if testCase.expectedError == "" {
				require.NoError(t, err)
				assert.ElementsMatch(t, testCase.expectedNodes, actualNodes)
			} else {
				assert.EqualError(t, err, testCase.expectedError)
				assert.Nil(t, actualNodes)
			}
		})
	}
}

func TestGetZoneKind(t *testing.T) {
	t.Parallel()

	var (
		zoneKind = graph.StringKind("Zone_Test")
	)

	testCases := []struct {
		name          string
		zoneNode      *graph.Node
		expectedKind  graph.Kind
		expectedError string
	}{
		{
			name: "Success: returns the kind named by the zone property",
			zoneNode: graph.NewNode(1, graph.AsProperties(graph.PropertyMap{
				zoneNodeZoneProperty: zoneKind.String(),
			}), graphschema.Zone),
			expectedKind: zoneKind,
		},
		{
			name:          "Error: returns an error when the zone property is missing",
			zoneNode:      graph.NewNode(1, graph.NewProperties(), graphschema.Zone),
			expectedError: "zone node is missing zone property: property zone: property not found",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			actualKind, err := getZoneKind(testCase.zoneNode)

			if testCase.expectedError == "" {
				assert.NoError(t, err)
				assert.Equal(t, testCase.expectedKind, actualKind)
			} else {
				assert.EqualError(t, err, testCase.expectedError)
				assert.Nil(t, actualKind)
			}
		})
	}
}

func TestGetZoneAssetGroupTagID(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		zoneNode      *graph.Node
		expectedID    int
		expectedError string
	}{
		{
			name: "Success: returns the asset group tag ID",
			zoneNode: graph.NewNode(1, graph.AsProperties(graph.PropertyMap{
				zoneNodeAssetGroupTagIDProperty: 42,
			}), graphschema.Zone),
			expectedID: 42,
		},
		{
			name:          "Error: returns an error when the property is missing",
			zoneNode:      graph.NewNode(1, graph.NewProperties(), graphschema.Zone),
			expectedError: "zone node is missing asset group tag ID property: property asset_group_tag_id: property not found",
		},
		{
			name: "Error: returns an error when the property has the wrong type",
			zoneNode: graph.NewNode(1, graph.AsProperties(graph.PropertyMap{
				zoneNodeAssetGroupTagIDProperty: "not an ID",
			}), graphschema.Zone),
			expectedError: "zone node is missing asset group tag ID property: unable to parse numeric value from raw value not an ID",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			actualID, err := getZoneAssetGroupTagID(testCase.zoneNode)

			if testCase.expectedError == "" {
				require.NoError(t, err)
				assert.Equal(t, testCase.expectedID, actualID)
			} else {
				assert.EqualError(t, err, testCase.expectedError)
				assert.Zero(t, actualID)
			}
		})
	}
}

func TestZoneNodePropertiesMatch(t *testing.T) {
	var (
		zone = model.AssetGroupTag{
			ID:   1,
			Name: "Zone",
		}
		zoneNode = graph.NewNode(101, graph.AsProperties(graph.PropertyMap{
			common.Name:                     zone.Name,
			common.DisplayName:              zone.Name,
			common.ObjectID:                 zoneNodeObjectID(zone),
			zoneNodeZoneProperty:            zone.ToKind().String(),
			zoneNodeAssetGroupTagIDProperty: zone.ID,
		}), graphschema.Zone)
	)

	t.Parallel()

	assert.True(t, zoneNodePropertiesMatch(zoneNode, zone))

	zoneNode.Properties.Set(common.Name.String(), "Stale Zone Name")
	assert.False(t, zoneNodePropertiesMatch(zoneNode, zone))
}

func TestIdentifyZoneNodeChanges(t *testing.T) {
	var (
		existingZone = model.AssetGroupTag{
			ID:   1,
			Name: "Existing Zone",
		}
		missingZone = model.AssetGroupTag{
			ID:   2,
			Name: "Missing Zone",
		}
		renamedZone = model.AssetGroupTag{
			ID:   3,
			Name: "Renamed Zone",
		}
		zones = model.AssetGroupTags{existingZone, missingZone, renamedZone}

		existingZoneNode = graph.NewNode(101, graph.AsProperties(graph.PropertyMap{
			common.Name:                     existingZone.Name,
			common.DisplayName:              existingZone.Name,
			common.ObjectID:                 zoneNodeObjectID(existingZone),
			zoneNodeZoneProperty:            existingZone.ToKind().String(),
			zoneNodeAssetGroupTagIDProperty: existingZone.ID,
		}), graphschema.Zone)
		duplicateZoneNode = graph.NewNode(102, graph.AsProperties(graph.PropertyMap{
			zoneNodeAssetGroupTagIDProperty: existingZone.ID,
		}), graphschema.Zone)
		orphanedZoneNode = graph.NewNode(103, graph.AsProperties(graph.PropertyMap{
			zoneNodeAssetGroupTagIDProperty: 4,
		}), graphschema.Zone)
		invalidZoneNode = graph.NewNode(104, graph.NewProperties(), graphschema.Zone)
		renamedZoneNode = graph.NewNode(105, graph.AsProperties(graph.PropertyMap{
			common.Name:                     "Old Zone Name",
			common.DisplayName:              "Old Zone Name",
			common.ObjectID:                 "zone:Old Zone Name",
			zoneNodeZoneProperty:            "Tag_Old_Zone_Name",
			zoneNodeAssetGroupTagIDProperty: renamedZone.ID,
		}), graphschema.Zone)
		existingZoneNodes = []*graph.Node{
			existingZoneNode,
			duplicateZoneNode,
			orphanedZoneNode,
			invalidZoneNode,
			renamedZoneNode,
		}
	)

	t.Parallel()

	zoneNodeIDsToDelete, zoneNodesToCreate, zoneNodesToUpdate := identifyZoneNodeChanges(existingZoneNodes, zones)

	assert.Equal(t, []graph.ID{duplicateZoneNode.ID, orphanedZoneNode.ID, invalidZoneNode.ID}, zoneNodeIDsToDelete)
	if assert.Len(t, zoneNodesToCreate, 1) {
		expectedZoneNode := graph.PrepareNode(graph.AsProperties(graph.PropertyMap{
			common.Name:                     missingZone.Name,
			common.DisplayName:              missingZone.Name,
			common.ObjectID:                 zoneNodeObjectID(missingZone),
			zoneNodeZoneProperty:            missingZone.ToKind().String(),
			zoneNodeAssetGroupTagIDProperty: missingZone.ID,
		}), graphschema.Zone)
		assert.Equal(t, expectedZoneNode, zoneNodesToCreate[0])
	}
	if assert.Equal(t, []*graph.Node{renamedZoneNode}, zoneNodesToUpdate) {
		assert.True(t, zoneNodePropertiesMatch(zoneNodesToUpdate[0], renamedZone))
	}
}

func TestReconcileZoneNodesErrors(t *testing.T) {
	t.Parallel()

	var (
		testError = errors.New("test error")
		zone      = model.AssetGroupTag{ID: 1, Name: "Zone"}
		staleNode = graph.NewNode(1, graph.AsProperties(graph.PropertyMap{
			common.Name:                     "Stale Zone",
			common.DisplayName:              "Stale Zone",
			common.ObjectID:                 "zone:1",
			zoneNodeZoneProperty:            "Tag_Stale_Zone",
			zoneNodeAssetGroupTagIDProperty: zone.ID,
		}), graphschema.Zone)
	)
	testCases := []struct {
		name              string
		existingZoneNodes []*graph.Node
		zones             model.AssetGroupTags
		skipInitialRead   bool
		setup             func(ctrl *gomock.Controller, graphDBMock *graphmocks.MockDatabase)
		expectedError     string
	}{
		{
			name:            "Error: returns get zone node errors",
			skipInitialRead: true,
			setup: func(_ *gomock.Controller, graphDBMock *graphmocks.MockDatabase) {
				graphDBMock.EXPECT().ReadTransaction(gomock.Any(), gomock.Any()).Return(testError)
			},
			expectedError: "get zone nodes: test error",
		},
		{
			name: "Error: wraps delete errors",
			existingZoneNodes: []*graph.Node{graph.NewNode(10, graph.AsProperties(graph.PropertyMap{
				zoneNodeAssetGroupTagIDProperty: 999,
			}), graphschema.Zone)},
			setup: func(ctrl *gomock.Controller, graphDBMock *graphmocks.MockDatabase) {
				batch := graphmocks.NewMockBatch(ctrl)
				batch.EXPECT().DeleteNode(graph.ID(10)).Return(testError)
				expectBatchOperation(graphDBMock, batch)
			},
			expectedError: "creating, updating, and deleting nodes: test error",
		},
		{
			name:  "Error: wraps create errors",
			zones: model.AssetGroupTags{zone},
			setup: func(ctrl *gomock.Controller, graphDBMock *graphmocks.MockDatabase) {
				batch := graphmocks.NewMockBatch(ctrl)
				batch.EXPECT().CreateNode(gomock.Any()).Return(testError)
				expectBatchOperation(graphDBMock, batch)
			},
			expectedError: "creating, updating, and deleting nodes: test error",
		},
		{
			name:              "Error: wraps update errors",
			existingZoneNodes: []*graph.Node{staleNode},
			zones:             model.AssetGroupTags{zone},
			setup: func(ctrl *gomock.Controller, graphDBMock *graphmocks.MockDatabase) {
				batch := graphmocks.NewMockBatch(ctrl)
				batch.EXPECT().UpdateNodes(gomock.Any()).Return(testError)
				expectBatchOperation(graphDBMock, batch)
			},
			expectedError: "creating, updating, and deleting nodes: test error",
		},
		{
			name:  "Error: wraps batch operation errors",
			zones: model.AssetGroupTags{zone},
			setup: func(_ *gomock.Controller, graphDBMock *graphmocks.MockDatabase) {
				graphDBMock.EXPECT().BatchOperation(gomock.Any(), gomock.Any()).Return(testError)
			},
			expectedError: "creating, updating, and deleting nodes: test error",
		},
		{
			name: "Error: returns final zone node read errors",
			setup: func(_ *gomock.Controller, graphDBMock *graphmocks.MockDatabase) {
				graphDBMock.EXPECT().ReadTransaction(gomock.Any(), gomock.Any()).Return(testError)
			},
			expectedError: "get zone nodes: test error",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var (
				ctrl        = gomock.NewController(t)
				graphDBMock = graphmocks.NewMockDatabase(ctrl)
			)

			if !testCase.skipInitialRead {
				transaction := graphmocks.NewMockTransaction(ctrl)
				expectZoneNodeFetch(ctrl, transaction, testCase.existingZoneNodes, nil)
				expectReadTransaction(graphDBMock, transaction)
			}
			testCase.setup(ctrl, graphDBMock)

			actualNodes, err := reconcileZoneNodes(context.Background(), graphDBMock, testCase.zones)

			assert.EqualError(t, err, testCase.expectedError)
			assert.Nil(t, actualNodes)
		})
	}
}

func TestIdentifyMemberOfZoneEdgeChanges(t *testing.T) {
	var (
		existingMembership  = zoneMembership{memberID: 1, zoneNodeID: 10}
		missingMembership   = zoneMembership{memberID: 2, zoneNodeID: 20}
		expectedMemberships = map[zoneMembership]struct{}{
			existingMembership: {},
			missingMembership:  {},
		}

		existingEdge   = graph.NewRelationship(101, existingMembership.memberID, existingMembership.zoneNodeID, graph.NewProperties(), graphschema.MemberOfZone)
		duplicateEdge  = graph.NewRelationship(102, existingMembership.memberID, existingMembership.zoneNodeID, graph.NewProperties(), graphschema.MemberOfZone)
		unexpectedEdge = graph.NewRelationship(103, 3, 30, graph.NewProperties(), graphschema.MemberOfZone)
		existingEdges  = []*graph.Relationship{
			existingEdge,
			duplicateEdge,
			unexpectedEdge,
		}
	)

	t.Parallel()

	edgeIDsToDelete, edgesToCreate := identifyMemberOfZoneEdgeChanges(existingEdges, expectedMemberships)

	assert.Equal(t, []graph.ID{duplicateEdge.ID, unexpectedEdge.ID}, edgeIDsToDelete)
	if assert.Len(t, edgesToCreate, 1) {
		expectedEdge := &graph.Relationship{
			StartID:    missingMembership.memberID,
			EndID:      missingMembership.zoneNodeID,
			Kind:       graphschema.MemberOfZone,
			Properties: graph.NewProperties(),
		}
		assert.Equal(t, expectedEdge, edgesToCreate[0])
	}
}

func TestGetExpectedMembershipsAndExistingEdgesErrors(t *testing.T) {
	t.Parallel()

	var (
		testError = errors.New("test error")
		zone      = model.AssetGroupTag{ID: 1, Name: "Zone"}
		zoneNode  = testZoneNode(10, zone)
	)
	testCases := []struct {
		name          string
		zoneNodes     []*graph.Node
		setup         func(ctrl *gomock.Controller, graphDBMock *graphmocks.MockDatabase)
		expectedError string
	}{
		{
			name: "Error: wraps transaction errors",
			setup: func(_ *gomock.Controller, graphDBMock *graphmocks.MockDatabase) {
				graphDBMock.EXPECT().ReadTransaction(gomock.Any(), gomock.Any()).Return(testError)
			},
			expectedError: "read zone memberships: test error",
		},
		{
			name:      "Error: wraps invalid zone node errors",
			zoneNodes: []*graph.Node{graph.NewNode(10, graph.NewProperties(), graphschema.Zone)},
			setup: func(ctrl *gomock.Controller, graphDBMock *graphmocks.MockDatabase) {
				transaction := graphmocks.NewMockTransaction(ctrl)
				expectReadTransaction(graphDBMock, transaction)
			},
			expectedError: "read zone memberships: zone node is missing zone property: property zone: property not found",
		},
		{
			name:      "Error: wraps member fetch errors",
			zoneNodes: []*graph.Node{zoneNode},
			setup: func(ctrl *gomock.Controller, graphDBMock *graphmocks.MockDatabase) {
				transaction := graphmocks.NewMockTransaction(ctrl)
				expectNodeIDFetch(ctrl, transaction, nil, testError)
				expectReadTransaction(graphDBMock, transaction)
			},
			expectedError: "read zone memberships: test error",
		},
		{
			name: "Error: wraps relationship fetch errors",
			setup: func(ctrl *gomock.Controller, graphDBMock *graphmocks.MockDatabase) {
				transaction := graphmocks.NewMockTransaction(ctrl)
				expectRelationshipFetch(ctrl, transaction, nil, testError)
				expectReadTransaction(graphDBMock, transaction)
			},
			expectedError: "read zone memberships: test error",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var (
				ctrl        = gomock.NewController(t)
				graphDBMock = graphmocks.NewMockDatabase(ctrl)
			)
			testCase.setup(ctrl, graphDBMock)

			expectedMemberships, existingEdges, err := getExpectedMembershipsAndExistingEdges(context.Background(), graphDBMock, testCase.zoneNodes)

			assert.EqualError(t, err, testCase.expectedError)
			assert.Nil(t, expectedMemberships)
			assert.Nil(t, existingEdges)
		})
	}
}

func TestReconcileMemberOfZoneEdgesErrors(t *testing.T) {
	t.Parallel()

	var (
		testError      = errors.New("test error")
		unexpectedEdge = graph.NewRelationship(100, 1, 2, graph.NewProperties(), graphschema.MemberOfZone)
		zone           = model.AssetGroupTag{ID: 1, Name: "Zone"}
		zoneNode       = testZoneNode(10, zone)
	)
	testCases := []struct {
		name          string
		zoneNodes     []*graph.Node
		setup         func(ctrl *gomock.Controller, graphDBMock *graphmocks.MockDatabase)
		expectedError string
	}{
		{
			name: "Error: returns membership read errors",
			setup: func(_ *gomock.Controller, graphDBMock *graphmocks.MockDatabase) {
				graphDBMock.EXPECT().ReadTransaction(gomock.Any(), gomock.Any()).Return(testError)
			},
			expectedError: "read zone memberships: test error",
		},
		{
			name: "Error: wraps relationship delete errors",
			setup: func(ctrl *gomock.Controller, graphDBMock *graphmocks.MockDatabase) {
				transaction := graphmocks.NewMockTransaction(ctrl)
				batch := graphmocks.NewMockBatch(ctrl)
				expectRelationshipFetch(ctrl, transaction, []*graph.Relationship{unexpectedEdge}, nil)
				expectReadTransaction(graphDBMock, transaction)
				batch.EXPECT().DeleteRelationship(unexpectedEdge.ID).Return(testError)
				expectBatchOperation(graphDBMock, batch)
			},
			expectedError: "creating and deleting edges: test error",
		},
		{
			name:      "Error: wraps relationship create errors",
			zoneNodes: []*graph.Node{zoneNode},
			setup: func(ctrl *gomock.Controller, graphDBMock *graphmocks.MockDatabase) {
				transaction := graphmocks.NewMockTransaction(ctrl)
				batch := graphmocks.NewMockBatch(ctrl)
				expectNodeIDFetch(ctrl, transaction, []graph.ID{1}, nil)
				expectRelationshipFetch(ctrl, transaction, nil, nil)
				expectReadTransaction(graphDBMock, transaction)
				batch.EXPECT().CreateRelationship(gomock.Any()).Return(testError)
				expectBatchOperation(graphDBMock, batch)
			},
			expectedError: "creating and deleting edges: test error",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var (
				ctrl        = gomock.NewController(t)
				graphDBMock = graphmocks.NewMockDatabase(ctrl)
			)
			testCase.setup(ctrl, graphDBMock)

			err := reconcileMemberOfZoneEdges(context.Background(), graphDBMock, testCase.zoneNodes)

			assert.EqualError(t, err, testCase.expectedError)
		})
	}
}

func TestGenerateZoneNodesAndMemberEdges_Unit(t *testing.T) {
	t.Parallel()

	var testError = errors.New("test error")
	testCases := []struct {
		name          string
		databaseError error
		setupGraphDB  func(ctrl *gomock.Controller, graphDBMock *graphmocks.MockDatabase)
		expectedError string
	}{
		{
			name:          "Error: returns zone database errors",
			databaseError: testError,
			expectedError: "get zones: test error",
		},
		{
			name: "Error: returns zone node reconciliation errors",
			setupGraphDB: func(_ *gomock.Controller, graphDBMock *graphmocks.MockDatabase) {
				graphDBMock.EXPECT().ReadTransaction(gomock.Any(), gomock.Any()).Return(testError)
			},
			expectedError: "get zone nodes: test error",
		},
		{
			name: "Error: returns membership reconciliation errors",
			setupGraphDB: func(ctrl *gomock.Controller, graphDBMock *graphmocks.MockDatabase) {
				firstTransaction := graphmocks.NewMockTransaction(ctrl)
				secondTransaction := graphmocks.NewMockTransaction(ctrl)
				expectZoneNodeFetch(ctrl, firstTransaction, nil, nil)
				expectZoneNodeFetch(ctrl, secondTransaction, nil, nil)

				gomock.InOrder(
					expectReadTransaction(graphDBMock, firstTransaction),
					expectReadTransaction(graphDBMock, secondTransaction),
					graphDBMock.EXPECT().ReadTransaction(gomock.Any(), gomock.Any()).Return(testError),
				)
			},
			expectedError: "read zone memberships: test error",
		},
		{
			name: "Success: succeeds when no changes are needed",
			setupGraphDB: func(ctrl *gomock.Controller, graphDBMock *graphmocks.MockDatabase) {
				firstTransaction := graphmocks.NewMockTransaction(ctrl)
				secondTransaction := graphmocks.NewMockTransaction(ctrl)
				membershipTransaction := graphmocks.NewMockTransaction(ctrl)
				expectZoneNodeFetch(ctrl, firstTransaction, nil, nil)
				expectZoneNodeFetch(ctrl, secondTransaction, nil, nil)
				expectRelationshipFetch(ctrl, membershipTransaction, nil, nil)

				gomock.InOrder(
					expectReadTransaction(graphDBMock, firstTransaction),
					expectReadTransaction(graphDBMock, secondTransaction),
					expectReadTransaction(graphDBMock, membershipTransaction),
				)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var (
				ctrl         = gomock.NewController(t)
				databaseMock = dbmocks.NewMockDatabase(ctrl)
				graphDBMock  = graphmocks.NewMockDatabase(ctrl)
			)

			databaseMock.EXPECT().GetAssetGroupTags(gomock.Any(), gomock.Any()).Return(nil, testCase.databaseError)
			if testCase.setupGraphDB != nil {
				testCase.setupGraphDB(ctrl, graphDBMock)
			}

			err := generateZoneNodesAndMemberEdges(context.Background(), databaseMock, graphDBMock)

			if testCase.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, testCase.expectedError)
			}
		})
	}
}
