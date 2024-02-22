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

const SET_GRAPH_LOADING = 'app/explore/SET_GRAPH_LOADING';
const GRAPH_START = 'app/explore/START';
const GRAPH_SUCCESS = 'app/explore/SUCCESS';
const GRAPH_FAILURE = 'app/explore/FAILURE';
const GRAPH_SETVARS = 'app/explore/SETVARS';
const GRAPH_INIT = 'app/explore/INIT';

const ADD_NODES = 'app/explore/ADDNODE';
const REMOVE_NODES = 'app/explore/REMOVENODE';

const SAVE_RESPONSE_FOR_EXPORT = 'app/explore/SAVE_RESPONSE_FOR_EXPORT';

const TOGGLE_TIER_ZERO_NODE = 'app/explore/TOGGLE_TIER_ZERO_NODE';

export {
    SET_GRAPH_LOADING,
    GRAPH_START,
    GRAPH_SUCCESS,
    GRAPH_FAILURE,
    GRAPH_SETVARS,
    ADD_NODES,
    REMOVE_NODES,
    GRAPH_INIT,
    SAVE_RESPONSE_FOR_EXPORT,
    TOGGLE_TIER_ZERO_NODE,
};

export enum GraphEndpoints {}

export interface GraphState {
    chartProps: any;
    loading: boolean;
    error: string | null;
    init: boolean;
    // we save the raw API response in the export field so that a user can export
    // their canvas to a target format, e.g. JSON, CSV
    export: any;
}

interface SetGraphLoadingAction {
    type: typeof SET_GRAPH_LOADING;
    isLoading: boolean;
}

interface GraphStartAction {
    type: typeof GRAPH_START;
    url: string;
}

interface GraphSuccessAction {
    type: typeof GRAPH_SUCCESS;
    payload: Items;
}

interface GraphFailureAction {
    type: typeof GRAPH_FAILURE;
    error: string;
}

interface GraphSetVarsAction {
    type: typeof GRAPH_SETVARS;
    payload: any;
}

interface GraphInitAction {
    type: typeof GRAPH_INIT;
    payload: boolean;
}

interface AddNodeAction {
    type: typeof ADD_NODES;
    ids: string[];
}

interface RemoveNodeAction {
    type: typeof REMOVE_NODES;
    ids: string[];
}

interface SaveResponseForExportAction {
    type: typeof SAVE_RESPONSE_FOR_EXPORT;
    payload: Items;
}

interface ToggleTierZeroNodeAction {
    type: typeof TOGGLE_TIER_ZERO_NODE;
    nodeId: string;
}

export type GraphActionTypes =
    | SetGraphLoadingAction
    | GraphStartAction
    | GraphSuccessAction
    | GraphFailureAction
    | GraphSetVarsAction
    | AddNodeAction
    | RemoveNodeAction
    | GraphInitAction
    | SaveResponseForExportAction
    | ToggleTierZeroNodeAction;

export interface NodeInfoRequest {
    type: typeof GRAPH_START;
    url: string;
}

export interface SearchRequest {
    type: typeof GRAPH_START;
    objectid: string;
    searchType: string;
}

export interface AssetGroupRequest {
    type: typeof GRAPH_START;
    assetGroupId: string;
    domainId: string | null | undefined;
    domainType: string | undefined;
}
export interface PathfindingRequest {
    type: typeof GRAPH_START;
    start: string;
    end: string;
    edges: string[];
}

export interface APIRequest {
    type: typeof GRAPH_START;
    endpoint: GraphEndpoints;
}

export interface CypherQueryRequest {
    type: typeof GRAPH_START;
    cypherQuery: string;
}

export interface ShortestPathRequest extends APIRequest {
    type: typeof GRAPH_START;
    target: string;
    edges: string[];
}

export type GraphRequestType =
    | APIRequest
    | PathfindingRequest
    | ShortestPathRequest
    | NodeInfoRequest
    | SearchRequest
    | AssetGroupRequest
    | CypherQueryRequest;

export enum GraphNodeTypes {
    User = 'User',
    Group = 'Group',
    Computer = 'Computer',
    GPO = 'GPO',
    OU = 'OU',
    Domain = 'Domain',
    AIACA = 'AIACA',
    RootCA = 'RootCA',
    EnterpriseCA = 'EnterpriseCA',
    NTAuthStore = 'NTAuthStore',
    CertTemplate = 'CertTemplate',
    Container = 'Container',
    Meta = 'Meta',
}

export interface GraphNodeData {
    admintier: number;
    count: number;
    date: object;
    nodetype: string;
    objectid: string;
    type: string;
    category?: string;
}

export interface GraphLinkData {
    impact_count: number;
    impact_pct: number;
    source: string;
    target: string;
}

export type GraphItemData = GraphNodeData & GraphLinkData;

export interface GraphGenericResponse {
    Statement: string;
    Nodes: {
        ID: number;
        Properties: {
            objectid: string;
            [key: string]: any;
        };
        Labels: string[];
    }[];
    Relationships: {
        ID: number;
        StartID: number;
        EndID: number;
        Type: string;
        Properties: {
            [key: string]: any;
        };
    }[];
}
