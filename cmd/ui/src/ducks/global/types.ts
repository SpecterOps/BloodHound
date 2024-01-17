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

const GLOBAL_ADD_SNACKBAR = 'app/global/ADDSNACKBAR';
const GLOBAL_CLOSE_SNACKBAR = 'app/global/CLOSESNACKBAR';
const GLOBAL_REMOVE_SNACKBAR = 'app/global/REMOVESNACKBAR';
const GLOBAL_SET_EXPANDED = 'app/global/SETEXPANDED';
const GLOBAL_SET_DOMAIN = 'app/global/SETDOMAIN';
const GLOBAL_FETCH_ASSET_GROUPS = 'app/global/GLOBALFETCHASSETGROUPS';
const GLOBAL_SET_ASSET_GROUPS = 'app/global/GLOBALSETASSETGROUPS';
const GLOBAL_SET_ASSET_GROUP_INDEX = 'app/global/GLOBALSETASSETGROUPINDEX';
const GLOBAL_SET_ASSET_GROUP_EDIT = 'app/global/GLOBALSETASSETGROUPEDIT';

export {
    GLOBAL_ADD_SNACKBAR,
    GLOBAL_CLOSE_SNACKBAR,
    GLOBAL_REMOVE_SNACKBAR,
    GLOBAL_SET_EXPANDED,
    GLOBAL_SET_DOMAIN,
    GLOBAL_FETCH_ASSET_GROUPS,
    GLOBAL_SET_ASSET_GROUPS,
    GLOBAL_SET_ASSET_GROUP_INDEX,
    GLOBAL_SET_ASSET_GROUP_EDIT,
};

export interface GlobalViewState {
    drawerOpen: boolean;
    pageTitle: string;
    notifications: Notification[];
}

export interface Notification {
    message: string;
    key: string;
    dismissed: boolean;
    options: any;
}

export interface DatapipeStatus {
    status: 'idle' | 'ingesting' | 'analyzing';
    updated_at: string;
    last_complete_analysis_at: string;
}

export interface GlobalOptionsState {
    baseUrl: string;
    domain: Domain | null;
    assetGroups: any[];
    assetGroupIndex: number | null;
    assetGroupEdit: number | null;
}

export interface GlobalAccordionsState {
    expanded: { [key: string]: symbol[] };
}

interface AddSnackbarAction {
    type: typeof GLOBAL_ADD_SNACKBAR;
    notification: Notification;
}

interface RemoveSnackbarAction {
    type: typeof GLOBAL_REMOVE_SNACKBAR;
    key: string;
}

interface CloseSnackbarAction {
    type: typeof GLOBAL_CLOSE_SNACKBAR;
    key: string;
}

export type GlobalViewActionTypes = AddSnackbarAction | RemoveSnackbarAction | CloseSnackbarAction;

export interface SetDomainAction {
    type: typeof GLOBAL_SET_DOMAIN;
    domain: Domain | null;
}

export interface FetchAssetGroupsAction {
    type: typeof GLOBAL_FETCH_ASSET_GROUPS;
}

export interface SetAssetGroupsAction {
    type: typeof GLOBAL_SET_ASSET_GROUPS;
    assetGroups: any[];
}
export interface SetAssetGroupIndexAction {
    type: typeof GLOBAL_SET_ASSET_GROUP_INDEX;
    assetGroupIndex: number | null;
}
export interface SetAssetGroupEditAction {
    type: typeof GLOBAL_SET_ASSET_GROUP_EDIT;
    assetGroupId: number | null;
}

export type GlobalOptionsActionTypes =
    | SetDomainAction
    | FetchAssetGroupsAction
    | SetAssetGroupsAction
    | SetAssetGroupIndexAction
    | SetAssetGroupEditAction;

export interface Domain {
    type: string;
    impactValue: number;
    name: string;
    id: string;
    collected: boolean;
}

export interface SetExpandedAction {
    type: typeof GLOBAL_SET_EXPANDED;
    expanded: { [key: string]: symbol[] };
}

export type GlobalAccordionsActionTypes = SetExpandedAction;
