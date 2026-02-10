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
import { BaseGraphLayoutOptions, Notification } from 'bh-shared-ui';
import { Environment } from 'js-client-library';
import { SnackbarKey } from 'notistack';

const GLOBAL_ADD_SNACKBAR = 'app/global/ADDSNACKBAR';
const GLOBAL_CLOSE_SNACKBAR = 'app/global/CLOSESNACKBAR';
const GLOBAL_REMOVE_SNACKBAR = 'app/global/REMOVESNACKBAR';
const GLOBAL_SET_EXPANDED = 'app/global/SETEXPANDED';
const GLOBAL_SET_DOMAIN = 'app/global/SETDOMAIN';
const GLOBAL_FETCH_ASSET_GROUPS = 'app/global/GLOBALFETCHASSETGROUPS';
const GLOBAL_SET_ASSET_GROUPS = 'app/global/GLOBALSETASSETGROUPS';
const GLOBAL_SET_ASSET_GROUP_INDEX = 'app/global/GLOBALSETASSETGROUPINDEX';
const GLOBAL_SET_ASSET_GROUP_EDIT = 'app/global/GLOBALSETASSETGROUPEDIT';
const GLOBAL_SET_DARK_MODE = 'app/global/GLOBALSETDARKMODE';
const GLOBAL_SET_EXPLORE_LAYOUT = 'app/global/GLOBAL_SET_EXPLORE_LAYOUT';
const GLOBAL_SET_IS_EXPLORE_TABLE_SELECTED = 'app/global/GLOBAL_SET_IS_EXPLORE_TABLE_SELECTED';
const GLOBAL_SET_AUTO_RUN_QUERIES = 'app/global/GLOBALSETAUTORUNQUERIES';
const GLOBAL_SET_SELECTED_EXPLORE_TABLE_COLUMNS = 'app/global/GLOBAL_SET_SELECTED_EXPLORE_TABLE_COLUMNS';

export {
    GLOBAL_ADD_SNACKBAR,
    GLOBAL_CLOSE_SNACKBAR,
    GLOBAL_FETCH_ASSET_GROUPS,
    GLOBAL_REMOVE_SNACKBAR,
    GLOBAL_SET_ASSET_GROUP_EDIT,
    GLOBAL_SET_ASSET_GROUP_INDEX,
    GLOBAL_SET_ASSET_GROUPS,
    GLOBAL_SET_AUTO_RUN_QUERIES,
    GLOBAL_SET_DARK_MODE,
    GLOBAL_SET_DOMAIN,
    GLOBAL_SET_EXPANDED,
    GLOBAL_SET_EXPLORE_LAYOUT,
    GLOBAL_SET_IS_EXPLORE_TABLE_SELECTED,
    GLOBAL_SET_SELECTED_EXPLORE_TABLE_COLUMNS,
};

export interface GlobalViewState {
    notifications: Notification[];
    darkMode: boolean;
    autoRunQueries: boolean;
    // Future dev: exploreLayout and isExploreTableSelected are undefined until a user selects a layout. After that, the layout is persisted in localStorage (until cache clears)
    exploreLayout?: BaseGraphLayoutOptions;
    isExploreTableSelected?: boolean;
    selectedExploreTableColumns?: Record<string, boolean>;
}

export interface GlobalOptionsState {
    domain: Environment | null;
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
    key: SnackbarKey;
}

interface CloseSnackbarAction {
    type: typeof GLOBAL_CLOSE_SNACKBAR;
    key: SnackbarKey;
}

export interface SetDarkModeAction {
    type: typeof GLOBAL_SET_DARK_MODE;
    darkMode: boolean;
}
export interface SetExploreLayoutAction {
    type: typeof GLOBAL_SET_EXPLORE_LAYOUT;
    exploreLayout: BaseGraphLayoutOptions;
}

export interface SetAutoRunQueriesAction {
    type: typeof GLOBAL_SET_AUTO_RUN_QUERIES;
    autoRunQueries: boolean;
}

export type GlobalViewActionTypes =
    | AddSnackbarAction
    | RemoveSnackbarAction
    | CloseSnackbarAction
    | SetDarkModeAction
    | SetExploreLayoutAction
    | SetIsExploreTableSelectedAction
    | SetAutoRunQueriesAction
    | SetSelectedExploreTableColumns;

export interface SetDomainAction {
    type: typeof GLOBAL_SET_DOMAIN;
    domain: Environment | null;
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

export interface SetIsExploreTableSelectedAction {
    type: typeof GLOBAL_SET_IS_EXPLORE_TABLE_SELECTED;
    isExploreTableSelected: boolean;
}
export interface SetSelectedExploreTableColumns {
    type: typeof GLOBAL_SET_SELECTED_EXPLORE_TABLE_COLUMNS;
    selectedExploreTableColumns: Record<string, boolean>;
}

export type GlobalOptionsActionTypes =
    | SetDomainAction
    | FetchAssetGroupsAction
    | SetAssetGroupsAction
    | SetAssetGroupIndexAction
    | SetAssetGroupEditAction;

export interface SetExpandedAction {
    type: typeof GLOBAL_SET_EXPANDED;
    expanded: { [key: string]: symbol[] };
}

export type GlobalAccordionsActionTypes = SetExpandedAction;
