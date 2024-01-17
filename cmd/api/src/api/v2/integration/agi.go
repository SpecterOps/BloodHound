// Copyright 2023 Specter Ops, Inc.
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

package integration

import (
	v2 "github.com/specterops/bloodhound/src/api/v2"
	"github.com/specterops/bloodhound/src/model"
	"github.com/stretchr/testify/require"
)

func (s *Context) ListAssetGroups() model.AssetGroups {
	listAssetGroupsResponse, err := s.AdminClient().ListAssetGroups()

	require.Nilf(s.TestCtrl, err, "Failed listing asset groups: %v", err)
	return listAssetGroupsResponse.AssetGroups
}

func (s *Context) GetAssetGroupByID(assetGroupID int32) model.AssetGroup {
	assetGroup, err := s.AdminClient().GetAssetGroup(assetGroupID)

	require.Nilf(s.TestCtrl, err, "Failed to get asset group %d: %v", assetGroupID, err)
	return assetGroup
}

func (s *Context) CreateAssetGroup(name, tag string) model.AssetGroup {
	assetGroup, err := s.AdminClient().CreateAssetGroup(v2.CreateAssetGroupRequest{
		Name: name,
		Tag:  tag,
	})

	require.Nilf(s.TestCtrl, err, "Failed to create asset group %s: %v", name, err)
	return assetGroup
}

func (s *Context) DeleteAssetGroup(assetGroupID int32) {
	err := s.AdminClient().DeleteAssetGroup(assetGroupID)
	require.Nilf(s.TestCtrl, err, "Failed to create asset group %d: %v", assetGroupID, err)
}

func (s *Context) UpdateAssetGroup(assetGroupID int32, updatedName string) model.AssetGroup {
	assetGroup, err := s.AdminClient().UpdateAssetGroup(assetGroupID, v2.UpdateAssetGroupRequest{
		Name: updatedName,
	})

	require.Nilf(s.TestCtrl, err, "Failed to update asset group %d: %v", assetGroupID, err)
	return assetGroup
}

func (s *Context) CreateAssetGroupSelector(assetGroupID int32, selectorSpec model.AssetGroupSelectorSpec) model.AssetGroupSelector {
	newSelector, err := s.AdminClient().CreateAssetGroupSelector(assetGroupID, selectorSpec)

	require.Nil(s.TestCtrl, err, "Failed to create a new asset group selector: %v", err)
	return newSelector
}

func (s *Context) AssetAssetGroupSelectors(assetGroupID int32, expectedAssetGroupSelectors model.AssetGroupSelectors) {
	actualAssetGroupSelectors := s.GetAssetGroupByID(assetGroupID).Selectors
	require.Equal(s.TestCtrl, len(expectedAssetGroupSelectors), len(actualAssetGroupSelectors))

	for _, expectedSelector := range expectedAssetGroupSelectors {
		found := false

		for _, actualSelector := range actualAssetGroupSelectors {
			if expectedSelector.Name == actualSelector.Name {
				require.Equal(s.TestCtrl, expectedSelector.Selector, actualSelector.Selector)
				require.Equal(s.TestCtrl, expectedSelector.SystemSelector, actualSelector.SystemSelector)

				found = true
				break
			}
		}

		if !found {
			s.TestCtrl.Fatalf("Unable to find asset group selector %s on asset group %d", expectedSelector.Name, assetGroupID)
		}
	}
}

func (s *Context) AssertAssetGroups(expectedAssetGroups model.AssetGroups) {
	actualAssetGroups := s.ListAssetGroups()
	require.Equal(s.TestCtrl, len(expectedAssetGroups), len(actualAssetGroups))

	for _, expectedAssetGroup := range expectedAssetGroups {
		found := false

		for _, actualAssetGroup := range actualAssetGroups {
			if expectedAssetGroup.Name == actualAssetGroup.Name {
				require.Equal(s.TestCtrl, expectedAssetGroup.Tag, actualAssetGroup.Tag)
				require.Equal(s.TestCtrl, expectedAssetGroup.SystemGroup, actualAssetGroup.SystemGroup)
				require.Equal(s.TestCtrl, expectedAssetGroup.Collections, actualAssetGroup.Collections)

				found = true
				break
			}
		}

		require.Truef(s.TestCtrl, found, "Unable to find an asset group by name %s", expectedAssetGroup.Name)
	}
}
