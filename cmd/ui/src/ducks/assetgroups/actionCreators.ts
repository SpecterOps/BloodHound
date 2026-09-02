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

import * as actions from './actions';
import * as types from './types';

export const ListAssetGroups = (): types.ListAssetGroupsAction => ({
    type: actions.LIST_ASSET_GROUPS,
});

export const ListAssetGroupsSuccess = (assetGroups: any[]): types.ListAssetGroupsSuccessAction => ({
    type: actions.LIST_ASSET_GROUPS_SUCCESS,
    assetGroups,
});

export const ListAssetGroupsFailure = (error: any): types.ListAssetGroupsFailureAction => ({
    type: actions.LIST_ASSET_GROUPS_FAILURE,
    error,
});

export const GetAssetGroup = (assetGroupId: string): types.GetAssetGroupAction => ({
    type: actions.GET_ASSET_GROUP,
    assetGroupId,
});

export const GetAssetGroupSuccess = (
    assetGroupId: string,
    assetGroupDetail: any
): types.GetAssetGroupSuccessAction => ({
    type: actions.GET_ASSET_GROUP_SUCCESS,
    assetGroupId,
    assetGroupDetail,
});

export const GetAssetGroupFailure = (error: any): types.GetAssetGroupFailureAction => ({
    type: actions.GET_ASSET_GROUP_FAILURE,
    error,
});
