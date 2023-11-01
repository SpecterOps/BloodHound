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
import cloneDeep from 'lodash/cloneDeep';
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

export const initialSearchState: types.SearchState = {
    searchType: types.SEARCH_TYPE_EXACT,

    primary: {
        searchTerm: '',
        loading: false,
        value: null,
        options: [],
        openMenu: false,
    },
    secondary: {
        searchTerm: '',
        loading: false,
        value: null,
        options: [],
        openMenu: false,
    },
    cypher: {
        searchTerm: '',
    },
    pathFilters: initialPathFilters,
    activeTab: 'primary',
};

function isTargetedActionType(action: types.SearchbarActionTypes): action is types.SearchbarTargetedActionTypes {
    return (action as types.SearchbarTargetedActionTypes).target !== undefined;
}

const searchReducer = (state = initialSearchState, action: types.SearchbarActionTypes) => {
    switch (action.type) {
        case types.SEARCH_RESET: {
            return cloneDeep(initialSearchState);
        }

        case types.SAVE_PATH_FILTERS: {
            return {
                ...state,
                pathFilters: [...action.filters],
            };
        }

        case types.TAB_CHANGED: {
            return {
                ...state,
                activeTab: action.tabName,
            };
        }

        case types.CYPHER_QUERY_EDITED: {
            return {
                ...state,
                cypher: {
                    searchTerm: action.searchTerm,
                },
            };
        }
    }

    return produce(state, (draft) => {
        switch (action.type) {
            case types.SOURCE_NODE_SUGGESTED: {
                draft.activeTab = types.PATHFINDING_SEARCH;

                draft.primary = {
                    searchTerm: action.name,
                    openMenu: true,

                    value: null,
                    loading: true,
                    options: [],
                };

                break;
            }
            case types.SOURCE_NODE_SELECTED: {
                draft.primary.openMenu = false;
                break;
            }
            case types.DESTINATION_NODE_SUGGESTED: {
                draft.activeTab = types.PATHFINDING_SEARCH;

                draft.secondary = {
                    searchTerm: action.name,
                    openMenu: true,

                    value: null,
                    loading: true,
                    options: [],
                };
                break;
            }
            case types.DESTINATION_NODE_SELECTED: {
                draft.secondary.openMenu = false;
                break;
            }
        }

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
                // clear out `searchTerm` if value is undefined or null
                if (!action.value) {
                    draft[target].searchTerm = '';
                } else {
                    draft[target].searchTerm = action.value.name;
                }
            }
        }
    });
};

export default searchReducer;
