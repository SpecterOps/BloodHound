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
import { AllEdgeTypes, Category, Subcategory } from 'bh-shared-ui';

// by default: all checkboxes are selected
const initialPathFilters: EdgeCheckboxType[] = [];

AllEdgeTypes.forEach((category: Category) => {
    category.subcategories.forEach((subcategory: Subcategory) => {
        subcategory.edgeTypes.forEach((edgeType: string) => {
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
    },
    secondary: {
        searchTerm: '',
        loading: false,
        value: null,
        options: [],
    },
    cypher: {
        searchTerm: '',
    },
    pathFilters: initialPathFilters,
    activeTab: 'primary',
};

const searchReducer = (state = initialSearchState, action: types.SearchbarActionTypes) => {
    switch (action.type) {
        case types.SEARCH_RESET: {
            return cloneDeep(initialSearchState);
        }

        case types.PATH_FILTERS_SAVED: {
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
            case types.SOURCE_NODE_EDITED: {
                draft.primary.searchTerm = action.searchTerm;
                draft.primary.loading = true;
                draft.primary.options = [];

                // any edits to the source node should clear out the previously saved primary.value
                draft.primary.value = null;
                break;
            }

            case types.SOURCE_NODE_SELECTED: {
                draft.searchType = types.SEARCH_TYPE_EXACT;

                draft.primary.value = action.node;

                if (action.node) {
                    draft.primary.searchTerm = action.node.name;
                } else {
                    draft.primary.searchTerm = '';
                }
                break;
            }

            case types.DESTINATION_NODE_EDITED: {
                draft.secondary.searchTerm = action.searchTerm;
                draft.secondary.loading = true;
                draft.secondary.options = [];

                // any edits to the destination node should clear out the previously saved destination.value
                draft.secondary.value = null;

                break;
            }

            case types.DESTINATION_NODE_SELECTED: {
                draft.searchType = types.SEARCH_TYPE_EXACT;

                draft.secondary.value = action.node;

                if (action.node) {
                    draft.secondary.searchTerm = action.node.name;
                } else {
                    draft.secondary.searchTerm = '';
                }
                break;
            }
        }
    });
};

export default searchReducer;
