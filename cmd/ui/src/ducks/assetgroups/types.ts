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

export interface AssetGroupsState {
    assetGroups: any[];
    assetGroupDetails: Record<string, any>;
    loading: boolean;
    error: any | null;
}

export interface ListAssetGroupsAction {
    type: typeof actions.LIST_ASSET_GROUPS;
}

export interface ListAssetGroupsSuccessAction {
    type: typeof actions.LIST_ASSET_GROUPS_SUCCESS;
    assetGroups: any[];
}

export interface ListAssetGroupsFailureAction {
    type: typeof actions.LIST_ASSET_GROUPS_FAILURE;
    error: any;
}

export interface GetAssetGroupAction {
    type: typeof actions.GET_ASSET_GROUP;
    assetGroupId: string;
}

export interface GetAssetGroupSuccessAction {
    type: typeof actions.GET_ASSET_GROUP_SUCCESS;
    assetGroupId: string;
    assetGroupDetail: any;
}

export interface GetAssetGroupFailureAction {
    type: typeof actions.GET_ASSET_GROUP_FAILURE;
    error: any;
}

export type AssetGroupsActionTypes =
    | ListAssetGroupsAction
    | ListAssetGroupsSuccessAction
    | ListAssetGroupsFailureAction
    | GetAssetGroupAction
    | GetAssetGroupSuccessAction
    | GetAssetGroupFailureAction;
