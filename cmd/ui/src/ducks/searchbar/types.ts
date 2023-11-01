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

import { GraphNodeTypes } from 'src/ducks/graph/types';
import { EdgeCheckboxType } from 'src/views/Explore/ExploreSearch/EdgeFilteringDialog';

const SEARCH_SUCCESS = 'app/search/SUCCESS';
const SEARCH_FAILURE = 'app/search/FAILURE';
const SEARCH_SELECTED = 'app/search/SELECTED';
const SEARCH_SET_PATHFINDING = 'app/search/SET_PATHFINDING';
const SEARCH_RESET = 'app/search/RESET';
const CYPHER_QUERY_EDITED = 'app/search/CYPHER_QUERY_EDITED';
const SAVE_PATH_FILTERS = 'app/search/SAVE_PATH_FILTERS';

export const TAB_CHANGED = 'app/search/TAB_CHANGED';

export const SOURCE_NODE_EDITED = 'app/search/SOURCE_NODE_EDITED';
export const SOURCE_NODE_SUGGESTED = 'app/search/SOURCE_NODE_SUGGESTED';
export const SOURCE_NODE_SELECTED = 'app/search/SOURCE_NODE_SELECTED';

export const DESTINATION_NODE_EDITED = 'app/search/DESTINATION_NODE_EDITED';
export const DESTINATION_NODE_SUGGESTED = 'app/search/DESTINATION_NODE_SUGGESTED';
export const DESTINATION_NODE_SELECTED = 'app/search/DESTINATION_NODE_SELECTED';

const PRIMARY_SEARCH = 'primary';
const PATHFINDING_SEARCH = 'secondary';
const CYPHER_SEARCH = 'cypher';
const TIER_ZERO_SEARCH = 'tierZero';

const SEARCH_ENDPOINT = '/search';

const SEARCH_TYPE_FUZZY = 'fuzzy';
const SEARCH_TYPE_EXACT = 'exact';

export {
    SEARCH_SUCCESS,
    SEARCH_FAILURE,
    PRIMARY_SEARCH,
    PATHFINDING_SEARCH,
    CYPHER_SEARCH,
    TIER_ZERO_SEARCH,
    SEARCH_ENDPOINT,
    SEARCH_SELECTED,
    SEARCH_SET_PATHFINDING,
    SEARCH_TYPE_EXACT,
    SEARCH_TYPE_FUZZY,
    SEARCH_RESET,
    CYPHER_QUERY_EDITED,
    SAVE_PATH_FILTERS,
};

export interface SearchBarState {
    options: SearchNodeType[];
    searchTerm: string;
    loading: boolean;
    value: SearchNodeType | null;
    openMenu: boolean;
}
export interface SearchNodeType {
    objectid: string;
    type: GraphNodeTypes;
    name: string;
}

export interface CypherSearchState {
    searchTerm: string;
}

export interface SearchState {
    primary: SearchBarState;
    secondary: SearchBarState;
    cypher: CypherSearchState;

    searchType: string;
    pathFilters: EdgeCheckboxType[];
    activeTab: SearchTargetType;
}

interface SearchSuccessAction {
    type: typeof SEARCH_SUCCESS;
    results: SearchNodeType[];
    target: SearchTargetType;
}

interface SearchFailureAction {
    type: typeof SEARCH_FAILURE;
    target: SearchTargetType;
    error: string;
}

interface SearchResetAction {
    type: typeof SEARCH_RESET;
}

interface TabChangedAction {
    type: typeof TAB_CHANGED;
    tabName: SearchTargetType;
}

export interface SavePathFiltersAction {
    type: typeof SAVE_PATH_FILTERS;
    filters: EdgeCheckboxType[];
}

export interface StartSearchSelectedAction {
    type: typeof SEARCH_SELECTED;
    target: SearchTargetType;
}

export interface CypherSearchAction {
    type: typeof CYPHER_SEARCH;
    searchTerm: string;
}

export interface CypherQueryEditedAction {
    type: typeof CYPHER_QUERY_EDITED;
    searchTerm: string;
}

export interface SourceNodeSuggestedAction {
    type: typeof SOURCE_NODE_SUGGESTED;
    name: string;
}

export interface SourceNodeSelectedAction {
    type: typeof SOURCE_NODE_SELECTED;
    node: SearchNodeType | null;
}

export interface SourceNodeEditedAction {
    type: typeof SOURCE_NODE_EDITED;
    searchTerm: string;
}
export interface DestinationNodeSuggestedAction {
    type: typeof DESTINATION_NODE_SUGGESTED;
    name: string;
}

export interface DestinationNodeSelectedAction {
    type: typeof DESTINATION_NODE_SELECTED;
    node: SearchNodeType | null;
}

export interface DestinationNodeEditedAction {
    type: typeof DESTINATION_NODE_EDITED;
    searchTerm: string;
}

export enum EndPoints {
    search = '/api/search',
}

export type SearchbarTargetedActionTypes = SearchFailureAction | SearchSuccessAction;

export type CypherActionTypes = CypherSearchAction | CypherQueryEditedAction;

export type NodeActionTypes =
    | SourceNodeSuggestedAction
    | SourceNodeSelectedAction
    | SourceNodeEditedAction
    | DestinationNodeSuggestedAction
    | DestinationNodeSelectedAction
    | DestinationNodeEditedAction;

export type SearchbarActionTypes =
    | SearchbarTargetedActionTypes
    | StartSearchSelectedAction
    | SearchResetAction
    | SavePathFiltersAction
    | TabChangedAction
    | CypherActionTypes
    | NodeActionTypes;

export type SearchTargetType = typeof PRIMARY_SEARCH | typeof PATHFINDING_SEARCH;

export type TabNames = typeof PRIMARY_SEARCH | typeof PATHFINDING_SEARCH | typeof CYPHER_SEARCH;
