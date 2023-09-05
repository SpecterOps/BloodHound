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

import * as types from './types';

export const removeSnackbar = (key: string): types.GlobalViewActionTypes => {
    return {
        type: types.GLOBAL_REMOVE_SNACKBAR,
        key: key,
    };
};

export const addSnackbar = (notification: string, key: string, options: any = {}): types.GlobalViewActionTypes => {
    return {
        type: types.GLOBAL_ADD_SNACKBAR,
        notification: {
            message: notification,
            key: key || (new Date().getTime() + Math.random()).toString(),
            options: {
                ...options,
                autoHideDuration: 5000,
            },
            dismissed: false,
        },
    };
};

export const closeSnackbar = (key: string): types.GlobalViewActionTypes => {
    return {
        type: types.GLOBAL_CLOSE_SNACKBAR,
        key: key,
    };
};

export const setExpanded = (expanded: { [key: string]: symbol[] }): types.GlobalAccordionsActionTypes => {
    return {
        type: types.GLOBAL_SET_EXPANDED,
        expanded: expanded,
    };
};

export const setDomain = (domain: types.Domain | null): types.GlobalOptionsActionTypes => {
    return {
        type: types.GLOBAL_SET_DOMAIN,
        domain,
    };
};

export const fetchAssetGroups = (): types.GlobalOptionsActionTypes => {
    return {
        type: types.GLOBAL_FETCH_ASSET_GROUPS,
    };
};

export const setAssetGroups = (assetGroups: any[]): types.GlobalOptionsActionTypes => {
    return {
        type: types.GLOBAL_SET_ASSET_GROUPS,
        assetGroups,
    };
};

export const setAssetGroupIndex = (assetGroupIndex: number | null): types.GlobalOptionsActionTypes => {
    return {
        type: types.GLOBAL_SET_ASSET_GROUP_INDEX,
        assetGroupIndex,
    };
};

export const setAssetGroupEdit = (assetGroupId: number | null): types.GlobalOptionsActionTypes => {
    return {
        type: types.GLOBAL_SET_ASSET_GROUP_EDIT,
        assetGroupId,
    };
};
