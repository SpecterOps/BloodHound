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

package hybrid

import (
	"testing"

	adSchema "github.com/specterops/bloodhound/packages/go/graphschema/ad"
	"github.com/specterops/bloodhound/packages/go/graphschema/azure"
	"github.com/specterops/bloodhound/packages/go/graphschema/common"
	"github.com/specterops/dawgs/graph"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAADObjectIDPropertyMatchesCollectorContract(t *testing.T) {
	const collectorProperty = "aadobjectid"

	if actual := adSchema.AADObjectID.String(); actual != collectorProperty {
		t.Fatalf("expected AADObjectID property %q, got %q", collectorProperty, actual)
	}
}

func TestNormalizeObjectID(t *testing.T) {
	for name, testCase := range map[string]struct {
		input    string
		expected string
	}{
		"lowercase":          {input: "69e33ede-7272-4893-ba72-18e6a92a0184", expected: "69E33EDE-7272-4893-BA72-18E6A92A0184"},
		"surrounding spaces": {input: " 69e33ede-7272-4893-ba72-18e6a92a0184 ", expected: "69E33EDE-7272-4893-BA72-18E6A92A0184"},
		"empty":              {input: "", expected: ""},
	} {
		t.Run(name, func(t *testing.T) {
			if actual := normalizeObjectID(testCase.input); actual != testCase.expected {
				t.Fatalf("expected normalized object ID %q, got %q", testCase.expected, actual)
			}
		})
	}
}

func TestReverseRelationshipMap(t *testing.T) {
	actual := reverseRelationshipMap(map[graph.ID][]graph.ID{
		1: {10, 11},
		2: {11},
	})

	assert.Equal(t, map[graph.ID][]graph.ID{
		10: {1},
		11: {1, 2},
	}, actual)
}

func TestAddNodeToObjectIDMap(t *testing.T) {
	for name, testCase := range map[string]struct {
		value          any
		setProperty    bool
		expectError    bool
		expectedObject string
	}{
		"missing":    {},
		"empty":      {setProperty: true, value: "  "},
		"normalized": {setProperty: true, value: " object-id ", expectedObject: "OBJECT-ID"},
		"wrong type": {setProperty: true, value: true, expectError: true},
	} {
		t.Run(name, func(t *testing.T) {
			properties := graph.NewProperties()
			if testCase.setProperty {
				properties.Set(common.ObjectID.String(), testCase.value)
			}

			node := &graph.Node{ID: 1, Properties: properties}
			objectIDMap := make(map[string][]graph.ID)
			err := addNodeToObjectIDMap(objectIDMap, node)

			if testCase.expectError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			if testCase.expectedObject == "" {
				assert.Empty(t, objectIDMap)
			} else {
				assert.Equal(t, []graph.ID{node.ID}, objectIDMap[testCase.expectedObject])
			}
		})
	}
}

func TestAddEntraDSAdminGroupTenant(t *testing.T) {
	for name, testCase := range map[string]struct {
		nameValue      any
		tenantValue    any
		setName        bool
		setTenant      bool
		expectError    bool
		expectedTenant string
	}{
		"missing name":      {},
		"empty name":        {setName: true, nameValue: "  "},
		"unrelated group":   {setName: true, nameValue: "OTHER GROUP", setTenant: true, tenantValue: "tenant-id"},
		"missing tenant":    {setName: true, nameValue: "AAD DC ADMINISTRATORS@EXAMPLE.COM"},
		"empty tenant":      {setName: true, nameValue: "AAD DC ADMINISTRATORS@EXAMPLE.COM", setTenant: true, tenantValue: "  "},
		"normalized valid":  {setName: true, nameValue: " aad dc administrators@example.com ", setTenant: true, tenantValue: " tenant-id ", expectedTenant: "TENANT-ID"},
		"wrong name type":   {setName: true, nameValue: true, expectError: true},
		"wrong tenant type": {setName: true, nameValue: "AAD DC ADMINISTRATORS@EXAMPLE.COM", setTenant: true, tenantValue: true, expectError: true},
	} {
		t.Run(name, func(t *testing.T) {
			properties := graph.NewProperties()
			if testCase.setName {
				properties.Set(common.Name.String(), testCase.nameValue)
			}
			if testCase.setTenant {
				properties.Set(azure.TenantID.String(), testCase.tenantValue)
			}

			node := &graph.Node{ID: 1, Properties: properties}
			tenantMap := make(map[graph.ID]string)
			err := addEntraDSAdminGroupTenant(tenantMap, node)

			if testCase.expectError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			if testCase.expectedTenant == "" {
				assert.Empty(t, tenantMap)
			} else {
				assert.Equal(t, testCase.expectedTenant, tenantMap[node.ID])
			}
		})
	}
}
