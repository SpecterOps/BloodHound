// Copyright 2025 Specter Ops, Inc.
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

package analysis

import (
	"testing"

	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/packages/go/graphschema"
	"github.com/specterops/bloodhound/packages/go/graphschema/common"
	"github.com/specterops/dawgs/graph"
	"github.com/stretchr/testify/assert"
)

func TestGetZoneKind(t *testing.T) {
	var (
		zoneKind = graph.StringKind("Zone_Test")
	)

	t.Parallel()

	testCases := []struct {
		name          string
		zoneNode      *graph.Node
		expectedKind  graph.Kind
		expectedError string
	}{
		{
			name: "returns the kind named by the zone property",
			zoneNode: graph.NewNode(1, graph.AsProperties(graph.PropertyMap{
				zoneNodeZoneProperty: zoneKind.String(),
			}), graphschema.Zone),
			expectedKind: zoneKind,
		},
		{
			name:          "returns an error when the zone property is missing",
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

func TestRenconcileZoneNodes(t *testing.T) {
	var (
		existingZone = model.AssetGroupTag{
			ID:   1,
			Name: "Existing Zone",
		}
		missingZone = model.AssetGroupTag{
			ID:   2,
			Name: "Missing Zone",
		}
		zones = model.AssetGroupTags{existingZone, missingZone}

		existingZoneNode = graph.NewNode(101, graph.AsProperties(graph.PropertyMap{
			zoneNodeZoneProperty: existingZone.ToKind().String(),
		}), graphschema.Zone)
		duplicateZoneNode = graph.NewNode(102, graph.AsProperties(graph.PropertyMap{
			zoneNodeZoneProperty: existingZone.ToKind().String(),
		}), graphschema.Zone)
		orphanedZoneNode = graph.NewNode(103, graph.AsProperties(graph.PropertyMap{
			zoneNodeZoneProperty: "Tag_Orphaned_Zone",
		}), graphschema.Zone)
		invalidZoneNode   = graph.NewNode(104, graph.NewProperties(), graphschema.Zone)
		existingZoneNodes = []*graph.Node{
			existingZoneNode,
			duplicateZoneNode,
			orphanedZoneNode,
			invalidZoneNode,
		}
	)

	t.Parallel()

	zoneNodeIDsToDelete, zoneNodesToCreate := renconcileZoneNodes(existingZoneNodes, zones)

	assert.Equal(t, []graph.ID{duplicateZoneNode.ID, orphanedZoneNode.ID, invalidZoneNode.ID}, zoneNodeIDsToDelete)
	if assert.Len(t, zoneNodesToCreate, 1) {
		expectedZoneNode := graph.PrepareNode(graph.AsProperties(graph.PropertyMap{
			common.Name:          missingZone.Name,
			common.DisplayName:   missingZone.Name,
			common.ObjectID:      zoneNodeObjectID(missingZone),
			zoneNodeZoneProperty: missingZone.ToKind().String(),
		}), graphschema.Zone)
		assert.Equal(t, expectedZoneNode, zoneNodesToCreate[0])
	}
}

func TestReconcileMemberOfZoneEdges(t *testing.T) {
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

	edgeIDsToDelete, edgesToCreate := reconcileMemberOfZoneEdges(existingEdges, expectedMemberships)

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
