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

import { EdgeCheckboxType } from 'src/views/Explore/ExploreSearch/EdgeFilteringDialog';
import * as types from './types';

export const startSearchAction = (searchTerm: string, target: types.SearchTargetType): types.SearchbarActionTypes => {
    return {
        type: types.SEARCH_START,
        searchTerm,
        target,
    };
};

export const searchSuccessAction = (
    results: types.SearchNodeType[],
    target: types.SearchTargetType
): types.SearchbarActionTypes => {
    return {
        type: types.SEARCH_SUCCESS,
        results,
        target,
    };
};

export const searchFailAction = (error: string, target: types.SearchTargetType): types.SearchbarActionTypes => {
    return {
        type: types.SEARCH_FAILURE,
        target,
        error,
    };
};

export const setSearchValue = (
    value: types.SearchNodeType | null,
    target: types.SearchTargetType,
    searchType: string
): types.SearchbarActionTypes => {
    return {
        type: types.SEARCH_SET_VALUE,
        target,
        value,
        searchType,
    };
};

export const startSearchSelected = (target: types.SearchTargetType): types.SearchbarActionTypes => {
    return {
        type: types.SEARCH_SELECTED,
        target,
    };
};

export const startCypherSearch = (cypherQuery: string) => {
    return {
        type: types.SEARCH_START,
        target: types.CYPHER_SEARCH,
        searchTerm: cypherQuery,
    };
};

export const setCypherQueryTerm = (cypherQuery: string) => {
    return {
        type: types.CYPHER_SEARCH_SET_VALUE,
        target: types.CYPHER_SEARCH,
        searchTerm: cypherQuery,
    };
};

export const resetSearch = (): types.SearchbarActionTypes => {
    return {
        type: types.SEARCH_RESET,
    };
};

export const savePathFilters = (filters: EdgeCheckboxType[]): types.SavePathFiltersAction => {
    return {
        type: types.SAVE_PATH_FILTERS,
        filters,
    };
};
