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

import { produce } from 'immer';
import { combineReducers } from '@reduxjs/toolkit';
import * as types from './types';
import assign from 'lodash/assign';

const initialGlobalState: types.GlobalViewState = {
    drawerOpen: false,
    pageTitle: 'Home',
    notifications: [],
};

const globalViewReducer = (state = initialGlobalState, action: types.GlobalViewActionTypes) => {
    return produce(state, (draft) => {
        if (action.type === types.GLOBAL_ADD_SNACKBAR) {
            draft.notifications = [...draft.notifications, action.notification];
        } else if (action.type === types.GLOBAL_CLOSE_SNACKBAR) {
            draft.notifications = draft.notifications.map((notification) => {
                return action.key === null || action.key === notification.key
                    ? { ...notification, dismissed: true }
                    : { ...notification };
            });
        } else if (action.type === types.GLOBAL_REMOVE_SNACKBAR) {
            draft.notifications = draft.notifications.filter((notification) => notification.key !== action.key);
        }
    });
};

const initialOptionsState: types.GlobalOptionsState = {
    baseUrl: '',
    domain: null,
    assetGroups: [],
    assetGroupIndex: null,
    assetGroupEdit: null,
};

const globalOptionsReducer = (state = initialOptionsState, action: types.GlobalOptionsActionTypes) => {
    return produce(state, (draft) => {
        if (action.type === types.GLOBAL_SET_DOMAIN) {
            draft.domain = action.domain;
        } else if (action.type === types.GLOBAL_SET_ASSET_GROUPS) {
            draft.assetGroups = action.assetGroups;
            draft.assetGroupIndex = null;
        } else if (action.type === types.GLOBAL_SET_ASSET_GROUP_INDEX) {
            draft.assetGroupIndex = action.assetGroupIndex;
        } else if (action.type === types.GLOBAL_SET_ASSET_GROUP_EDIT) {
            draft.assetGroupEdit = action.assetGroupId;
        }
    });
};

const initialAccordionsState: types.GlobalAccordionsState = {
    expanded: {},
};

const globalAccordionsReducer = (state = initialAccordionsState, action: types.GlobalAccordionsActionTypes) => {
    return produce(state, (draft) => {
        if (action.type === types.GLOBAL_SET_EXPANDED) {
            assign(draft.expanded, action.expanded);
        }
    });
};

export default combineReducers({
    view: globalViewReducer,
    options: globalOptionsReducer,
    accordions: globalAccordionsReducer,
});
