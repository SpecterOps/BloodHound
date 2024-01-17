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

export const primarySearch = () => {
    return {
        type: types.PRIMARY_SEARCH,
    };
};

export const pathfindingSearch = () => {
    return {
        type: types.PATHFINDING_SEARCH,
    };
};

export const cypherSearch = (cypherQuery?: string): types.CypherSearchAction => {
    return {
        type: types.CYPHER_SEARCH,
        searchTerm: cypherQuery,
    };
};

export const cypherQueryEdited = (cypherQuery: string): types.CypherQueryEditedAction => {
    return {
        type: types.CYPHER_QUERY_EDITED,
        searchTerm: cypherQuery,
    };
};

export const resetSearch = (): types.SearchbarActionTypes => {
    return {
        type: types.SEARCH_RESET,
    };
};

export const pathFiltersSaved = (filters: EdgeCheckboxType[]): types.PathFiltersSavedAction => {
    return {
        type: types.PATH_FILTERS_SAVED,
        filters,
    };
};

export const tabChanged = (tabName: types.TabNames) => {
    return {
        type: types.TAB_CHANGED,
        tabName,
    };
};

export const sourceNodeEdited = (searchTerm: string): types.SourceNodeEditedAction => {
    return {
        type: types.SOURCE_NODE_EDITED,
        searchTerm,
    };
};

export const sourceNodeSelected = (
    node: types.SearchNodeType | null,
    doPathfindSearch: boolean = false // sometimes, selecting a source node should trigger a pathfinding search, and other times it should only trigger a single node search
): types.SourceNodeSelectedAction => {
    return {
        type: types.SOURCE_NODE_SELECTED,
        node,
        doPathfindSearch,
    };
};

export const destinationNodeEdited = (searchTerm: string): types.DestinationNodeEditedAction => {
    return {
        type: types.DESTINATION_NODE_EDITED,
        searchTerm,
    };
};

export const destinationNodeSelected = (node: types.SearchNodeType | null): types.DestinationNodeSelectedAction => {
    return {
        type: types.DESTINATION_NODE_SELECTED,
        node,
    };
};
