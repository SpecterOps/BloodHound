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

package apiclient

import (
	"fmt"
	"net/http"

	v2 "github.com/specterops/bloodhound/src/api/v2"

	"github.com/specterops/bloodhound/src/model"

	"github.com/specterops/bloodhound/src/api"
)

func (s Client) GetAssetGroup(assetGroupID int32) (model.AssetGroup, error) {
	var assetGroup model.AssetGroup

	if response, err := s.Request(http.MethodGet, fmt.Sprintf("api/v2/asset-groups/%d", assetGroupID), nil, nil); err != nil {
		return assetGroup, err
	} else {
		defer response.Body.Close()

		if api.IsErrorResponse(response) {
			return assetGroup, ReadAPIError(response)
		}

		return assetGroup, api.ReadAPIV2ResponsePayload(&assetGroup, response)
	}
}

func (s Client) ListAssetGroups() (v2.ListAssetGroupsResponse, error) {
	var assetGroups v2.ListAssetGroupsResponse

	if response, err := s.Request(http.MethodGet, "api/v2/asset-groups", nil, nil); err != nil {
		return assetGroups, err
	} else {
		defer response.Body.Close()

		if api.IsErrorResponse(response) {
			return assetGroups, ReadAPIError(response)
		}

		return assetGroups, api.ReadAPIV2ResponsePayload(&assetGroups, response)
	}
}

func (s Client) ListAssetGroupCollections(assetGroupID int32) (v2.AssetGroupCollectionsResponse, error) {
	if response, err := s.Request(http.MethodGet, fmt.Sprintf("api/v2/asset-groups/%d/collections", assetGroupID), nil, nil); err != nil {
		return v2.AssetGroupCollectionsResponse{Data: nil}, err
	} else {
		defer response.Body.Close()

		if api.IsErrorResponse(response) {
			return v2.AssetGroupCollectionsResponse{Data: nil}, ReadAPIError(response)
		}

		assetGroups := struct {
			Collections model.AssetGroupCollections `json:"collections"`
		}{}

		err := api.ReadAPIV2ResponsePayload(&assetGroups, response)
		if err != nil {
			return v2.AssetGroupCollectionsResponse{Data: nil}, err
		}

		data := make([]any, 0)
		for _, dataElement := range assetGroups.Collections {
			data = append(data, dataElement)
		}
		return v2.AssetGroupCollectionsResponse{Data: data}, nil
	}
}

func (s Client) ListAllAssetGroupCollections() (v2.AssetGroupCollectionsResponse, error) {
	if response, err := s.Request(http.MethodGet, "api/v2/asset-groups/collections", nil, nil); err != nil {
		return v2.AssetGroupCollectionsResponse{Data: nil}, err
	} else {
		defer response.Body.Close()

		if api.IsErrorResponse(response) {
			return v2.AssetGroupCollectionsResponse{Data: nil}, ReadAPIError(response)
		}

		assetGroups := struct {
			Collections model.AssetGroupCollections `json:"collections"`
		}{}

		err := api.ReadAPIV2ResponsePayload(&assetGroups, response)
		if err != nil {
			return v2.AssetGroupCollectionsResponse{Data: nil}, err
		}

		data := make([]any, 0)
		for _, dataElement := range assetGroups.Collections {
			data = append(data, dataElement)
		}
		return v2.AssetGroupCollectionsResponse{Data: data}, nil
	}
}

func (s Client) DeleteAssetGroup(assetGroupID int32) error {
	if response, err := s.Request(http.MethodDelete, fmt.Sprintf("api/v2/asset-groups/%d", assetGroupID), nil, nil); err != nil {
		return err
	} else {
		defer response.Body.Close()

		if api.IsErrorResponse(response) {
			return ReadAPIError(response)
		}

		return nil
	}
}

func (s Client) DeleteAssetGroupSelector(assetGroupID, assetGroupSelectorID int32) error {
	if response, err := s.Request(http.MethodDelete, fmt.Sprintf("api/v2/asset-groups/%d/selectors/%d", assetGroupID, assetGroupSelectorID), nil, nil); err != nil {
		return err
	} else {
		defer response.Body.Close()

		if api.IsErrorResponse(response) {
			return ReadAPIError(response)
		}

		return nil
	}
}

func (s Client) CreateAssetGroup(request v2.CreateAssetGroupRequest) (model.AssetGroup, error) {
	var assetGroup model.AssetGroup

	if response, err := s.Request(http.MethodPost, "api/v2/asset-groups", nil, request); err != nil {
		return assetGroup, err
	} else {
		defer response.Body.Close()

		if api.IsErrorResponse(response) {
			return assetGroup, ReadAPIError(response)
		}

		return assetGroup, api.ReadAPIV2ResponsePayload(&assetGroup, response)
	}
}

func (s Client) UpdateAssetGroup(assetGroupID int32, request v2.UpdateAssetGroupRequest) (model.AssetGroup, error) {
	var assetGroup model.AssetGroup

	if response, err := s.Request(http.MethodPut, fmt.Sprintf("api/v2/asset-groups/%d", assetGroupID), nil, request); err != nil {
		return assetGroup, err
	} else {
		defer response.Body.Close()

		if api.IsErrorResponse(response) {
			return assetGroup, ReadAPIError(response)
		}

		return assetGroup, api.ReadAPIV2ResponsePayload(&assetGroup, response)
	}
}

func (s Client) CreateAssetGroupSelector(assetGroupID int32, spec model.AssetGroupSelectorSpec) (model.AssetGroupSelector, error) {
	var assetGroupSelector model.AssetGroupSelector

	if response, err := s.Request(http.MethodPost, fmt.Sprintf("api/v2/asset-groups/%d/selectors", assetGroupID), nil, []model.AssetGroupSelectorSpec{spec}); err != nil {
		return assetGroupSelector, err
	} else {
		defer response.Body.Close()

		if api.IsErrorResponse(response) {
			return assetGroupSelector, ReadAPIError(response)
		}

		return assetGroupSelector, api.ReadAPIV2ResponsePayload(&assetGroupSelector, response)
	}
}
