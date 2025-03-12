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

import { searchbarActions, SearchValue, useFeatureFlag, usePathfindingSearch } from 'bh-shared-ui';
import { useAppDispatch, useAppSelector } from 'src/store';

export const usePathfindingSearchSwitch = () => {
    const { data: flag } = useFeatureFlag('back_button_support');

    // New implementation
    const pathfindingSearch = usePathfindingSearch();

    // Old redux implementation
    const dispatch = useAppDispatch();

    const primary = useAppSelector((state) => state.search.primary);
    const secondary = useAppSelector((state) => state.search.secondary);

    const { searchTerm: sourceSearchTerm, value: sourceSelectedItem } = primary;
    const { searchTerm: destinationSearchTerm, value: destinationSelectedItem } = secondary;

    const handleSourceNodeEdited = (edit: string) => {
        dispatch(searchbarActions.sourceNodeEdited(edit));
    };

    const handleDestinationNodeEdited = (edit: string) => {
        dispatch(searchbarActions.destinationNodeEdited(edit));
    };

    const handleSourceNodeSelected = (selected: SearchValue) => {
        const doPathfindSearch = !!destinationSelectedItem;
        dispatch(searchbarActions.sourceNodeSelected(selected, doPathfindSearch));
    };

    const handleDestinationNodeSelected = (selected: SearchValue) => {
        dispatch(searchbarActions.destinationNodeSelected(selected));
    };

    if (flag?.enabled) {
        return pathfindingSearch;
    } else {
        return {
            sourceSearchTerm,
            destinationSearchTerm,
            sourceSelectedItem,
            destinationSelectedItem,
            handleSourceNodeEdited,
            handleDestinationNodeEdited,
            handleSourceNodeSelected,
            handleDestinationNodeSelected,
        };
    }
};
