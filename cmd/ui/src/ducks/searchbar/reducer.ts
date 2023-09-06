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

import produce from 'immer';
import cloneDeep from 'lodash/cloneDeep';
import { GraphNodeTypes } from 'src/ducks/graph/types';
import * as types from 'src/ducks/searchbar/types';
import { EdgeCheckboxType } from 'src/views/Explore/ExploreSearch/EdgeFilteringDialog';
import { AllEdgeTypes } from 'src/views/Explore/ExploreSearch/edgeTypes';

// by default: all checkboxes are selected
const initialPathFilters: EdgeCheckboxType[] = [];

AllEdgeTypes.forEach((category) => {
    category.subcategories.forEach((subcategory) => {
        subcategory.edgeTypes.forEach((edgeType) => {
            initialPathFilters.push({
                category: category.categoryName,
                subcategory: subcategory.name,
                edgeType,
                checked: true,
            });
        });
    });
});

const initialSearchState: types.SearchState = {
    searchType: types.SEARCH_TYPE_EXACT,

    primary: {
        searchTerm: '',
        loading: false,
        value: null,
        options: [],
    },
    secondary: {
        searchTerm: '',
        loading: false,
        value: null,
        options: [],
    },
    tierZero: {
        searchTerm: '',
        loading: false,
        value: null,
        options: [],
    },
    cypher: {
        searchTerm: '',
        loading: false,
        value: null,
        options: [],
    },
    pathFilters: initialPathFilters,
};

function isTargetedActionType(action: types.SearchbarActionTypes): action is types.SearchbarTargetedActionTypes {
    return (action as types.SearchbarTargetedActionTypes).target !== undefined;
}

const searchReducer = (state = initialSearchState, action: types.SearchbarActionTypes) => {
    if (action.type === types.SEARCH_RESET) {
        return cloneDeep(initialSearchState);
    }

    if (action.type === types.SAVE_PATH_FILTERS) {
        return {
            ...state,
            pathFilters: [...action.filters],
        };
    }

    return produce(state, (draft) => {
        if (isTargetedActionType(action)) {
            const { target } = action;

            if (action.type === types.SEARCH_START) {
                draft[target].loading = true;
                draft[target].options = [];
                draft[target].searchTerm = action.searchTerm;
            } else if (action.type === types.SEARCH_SUCCESS) {
                draft[target].loading = false;
                draft[target].options = action.results;
            } else if (action.type === types.SEARCH_SET_VALUE) {
                draft.searchType = action.searchType;
                draft[target].value = action.value;
            } else if (action.type === types.CYPHER_SEARCH_SET_VALUE) {
                draft[target].searchTerm = action.searchTerm;
            } else if (action.type === types.SEARCH_SET_PATHFINDING) {
                draft.primary = {
                    searchTerm: action.primary.name,
                    loading: false,
                    value: {
                        label: action.primary.name,
                        name: action.primary.name,
                        objectid: action.primary.objectid,
                        type: GraphNodeTypes.User,
                    },
                    options: [],
                };

                if (action.secondary) {
                    draft.secondary = {
                        searchTerm: action.secondary.name,
                        loading: false,
                        value: {
                            label: action.secondary.name,
                            name: action.secondary.name,
                            objectid: action.secondary.objectid,
                            type: GraphNodeTypes.User,
                        },
                        options: [],
                    };
                } else {
                    draft.secondary = {
                        searchTerm: '',
                        loading: false,
                        value: null,
                        options: [],
                    };
                }
            }
        }
    });
};

export default searchReducer;
