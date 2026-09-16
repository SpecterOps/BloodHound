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
