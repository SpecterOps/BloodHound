// Copyright 2025 Specter Ops, Inc.
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

import { searchbarActions, SearchValue, useFeatureFlag, useNodeSearch } from 'bh-shared-ui';
import { useAppDispatch, useAppSelector } from 'src/store';

export const useNodeSearchSwitch = () => {
    const { data: flag } = useFeatureFlag('back_button_support');

    // New implementation
    const nodeSearch = useNodeSearch();

    // Redux resources for old implementation
    const dispatch = useAppDispatch();
    const primary = useAppSelector((state) => state.search.primary);
    const { searchTerm: reduxSearchTerm, value: reduxSelectedItem } = primary;

    if (flag?.enabled) {
        return nodeSearch;
    } else {
        return {
            searchTerm: reduxSearchTerm,
            selectedItem: reduxSelectedItem,
            editSourceNode: (edit: string) => dispatch(searchbarActions.sourceNodeEdited(edit)),
            selectSourceNode: (selected?: SearchValue) => dispatch(searchbarActions.sourceNodeSelected(selected)),
        };
    }
};
