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

const SEARCH_START = 'app/search/START';
const SEARCH_SUCCESS = 'app/search/SUCCESS';
const SEARCH_FAILURE = 'app/search/FAILURE';
const SEARCH_SET_VALUE = 'app/search/SETVALUE';
const SEARCH_SELECTED = 'app/search/SELECTED';
const SEARCH_SET_PATHFINDING = 'app/search/SET_PATHFINDING';
const SEARCH_RESET = 'app/search/RESET';
const CYPHER_SEARCH_SET_VALUE = 'app/search/CYPHERSEARCH_SETVALUE';
const SAVE_PATH_FILTERS = 'app/search/SAVE_PATH_FILTERS';

const PRIMARY_SEARCH = 'primary';
const SECONDARY_SEARCH = 'secondary';
const CYPHER_SEARCH = 'cypher';
const TIER_ZERO_SEARCH = 'tierZero';

const SEARCH_ENDPOINT = '/search';

const SEARCH_TYPE_FUZZY = 'fuzzy';
const SEARCH_TYPE_EXACT = 'exact';

export {
    SEARCH_START,
    SEARCH_SUCCESS,
    SEARCH_FAILURE,
    SEARCH_SET_VALUE,
    PRIMARY_SEARCH,
    SECONDARY_SEARCH,
    CYPHER_SEARCH,
    TIER_ZERO_SEARCH,
    SEARCH_ENDPOINT,
    SEARCH_SELECTED,
    SEARCH_SET_PATHFINDING,
    SEARCH_TYPE_EXACT,
    SEARCH_TYPE_FUZZY,
    SEARCH_RESET,
    CYPHER_SEARCH_SET_VALUE,
    SAVE_PATH_FILTERS,
};

export interface SearchBarState {
    options: SearchNodeType[];
    searchTerm: string;
    loading: boolean;
    value: SearchNodeType | null;
}

export interface SearchNodeType {
    objectid: string;
    label: string;
    type: GraphNodeTypes;
    name: string;
}

export interface SearchState {
    primary: SearchBarState;
    secondary: SearchBarState;
    tierZero: SearchBarState;
    cypher: SearchBarState;
    searchType: string;
    pathFilters: EdgeCheckboxType[];
}

export interface SearchStartAction {
    type: typeof SEARCH_START;
    searchTerm: string;
    target: SearchTargetType;
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

interface SearchSetValueAction {
    type: typeof SEARCH_SET_VALUE;
    target: SearchTargetType;
    value: SearchNodeType | null;
    searchType: string;
}

interface SearchResetAction {
    type: typeof SEARCH_RESET;
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
    type: typeof SEARCH_START;
    target: typeof CYPHER_SEARCH;
    searchTerm: string;
}

export interface CypherSearchSetQueryTermAction {
    type: typeof CYPHER_SEARCH_SET_VALUE;
    target: typeof CYPHER_SEARCH;
    searchTerm: string;
}

interface SearchSetPathfindingAction {
    type: typeof SEARCH_SET_PATHFINDING;
    primary: any;
    secondary: any;
    target: SearchTargetType;
}

export enum EndPoints {
    search = '/api/search',
}

export type SearchbarTargetedActionTypes =
    | SearchStartAction
    | SearchFailureAction
    | SearchSuccessAction
    | SearchSetValueAction
    | SearchSetPathfindingAction
    | CypherSearchAction
    | CypherSearchSetQueryTermAction;

export type SearchbarActionTypes =
    | SearchbarTargetedActionTypes
    | StartSearchSelectedAction
    | SearchResetAction
    | SavePathFiltersAction;

export type SearchTargetType =
    | typeof PRIMARY_SEARCH
    | typeof SECONDARY_SEARCH
    | typeof TIER_ZERO_SEARCH
    | typeof CYPHER_SEARCH;
