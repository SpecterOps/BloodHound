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

import { Items } from 'src/utils';
import * as types from './types';

export const setGraphLoading = (isLoading: boolean): types.GraphActionTypes => {
    return {
        type: types.SET_GRAPH_LOADING,
        isLoading,
    };
};

export const startAPIQuery = (endpoint: types.GraphEndpoints): types.GraphRequestType => {
    return {
        type: types.GRAPH_START,
        endpoint: endpoint,
    };
};

export const startNodeInfoQuery = (url: string): types.GraphRequestType => {
    const newURL = new URL(url, window.location.origin);
    newURL.searchParams.set('type', 'graph');
    return {
        type: types.GRAPH_START,
        url: newURL.pathname + newURL.search,
    };
};

export const startSearchQuery = (objectid: string, searchType: string): types.GraphRequestType => {
    return {
        type: types.GRAPH_START,
        objectid,
        searchType,
    };
};

export const startPathfindingQuery = (start: string, end: string, edges: string[]): types.GraphRequestType => {
    return {
        type: types.GRAPH_START,
        start: start,
        end: end,
        edges: edges,
    };
};

export const startCypherQuery = (cypherQuery: string): types.GraphRequestType => {
    return {
        type: types.GRAPH_START,
        cypherQuery,
    };
};

export const startAssetGroupQuery = (
    assetGroupId: string,
    domainId?: string | null,
    domainType?: string
): types.GraphRequestType => {
    return {
        type: types.GRAPH_START,
        assetGroupId,
        domainId,
        domainType,
    };
};

export const putGraphData = (payload: Items): types.GraphActionTypes => {
    return {
        type: types.GRAPH_SUCCESS,
        payload,
    };
};

export const putGraphError = (error: any): types.GraphActionTypes => {
    return {
        type: types.GRAPH_FAILURE,
        error,
    };
};

export const putGraphVars = (payload: any): types.GraphActionTypes => {
    return {
        type: types.GRAPH_SETVARS,
        payload,
    };
};

export const addNodes = (ids: string[]): types.GraphActionTypes => {
    return {
        type: types.ADD_NODES,
        ids,
    };
};

export const removeNodes = (ids: string[]): types.GraphActionTypes => {
    return {
        type: types.REMOVE_NODES,
        ids,
    };
};

export const initGraph = (payload: boolean): types.GraphActionTypes => {
    return {
        type: types.GRAPH_INIT,
        payload,
    };
};

export const saveResponseForExport = (payload: Items): types.GraphActionTypes => {
    return {
        type: types.SAVE_RESPONSE_FOR_EXPORT,
        payload,
    };
};

export const toggleTierZeroNode = (nodeId: string): types.GraphActionTypes => {
    return {
        type: types.TOGGLE_TIER_ZERO_NODE,
        nodeId,
    };
};
