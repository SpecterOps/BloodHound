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
import * as types from 'src/ducks/explore/types';

const initialGraphDataState: types.GraphState = {
    chartProps: {},
    loading: false,
    error: null,
    init: false,
    export: {},
};

const graphDataReducer = (state = initialGraphDataState, action: types.GraphActionTypes) => {
    return produce(state, (draft) => {
        if (action.type === types.SET_GRAPH_LOADING) {
            draft.loading = action.isLoading;
        } else if (action.type === types.GRAPH_START) {
            draft.loading = true;
            draft.error = null;
        } else if (action.type === types.GRAPH_SUCCESS) {
            draft.chartProps = { items: action.payload };
            draft.loading = false;
        } else if (action.type === types.GRAPH_FAILURE) {
            draft.loading = false;
            draft.error = action.error;
        } else if (action.type === types.GRAPH_SETVARS) {
            draft.chartProps = Object.assign({}, draft.chartProps, action.payload);
        } else if (action.type === types.ADD_NODES) {
            // handle add nodes (action.ids)
        } else if (action.type === types.REMOVE_NODES) {
            // handle delete nodes (action.ids)
        } else if (action.type === types.GRAPH_INIT) {
            draft.init = action.payload;
        } else if (action.type === types.SAVE_RESPONSE_FOR_EXPORT) {
            draft.export = action.payload;
        } else if (action.type === types.TOGGLE_TIER_ZERO_NODE) {
            const systemTags = state.chartProps.items[action.nodeId].data.system_tags;
            // remove the tier zero tag from the node
            if (systemTags === 'admin_tier_0') {
                draft.chartProps.items[action.nodeId].data.system_tags = '';
            } else {
                draft.chartProps.items[action.nodeId].data.system_tags = 'admin_tier_0';
            }
        }
        return draft;
    });
};

export default graphDataReducer;
