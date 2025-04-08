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

import { EdgeCheckboxType, searchbarActions } from 'bh-shared-ui';
import { useRef } from 'react';
import { useAppDispatch, useAppSelector } from 'src/store';

export const usePathfindingFilterSwitch = () => {
    const dispatch = useAppDispatch();
    const initialFilterState = useRef<EdgeCheckboxType[]>([]);
    const reduxPathfindingFilters = useAppSelector((state) => state.search.pathFilters);

    return {
        selectedFilters: reduxPathfindingFilters,
        initialize: () => (initialFilterState.current = reduxPathfindingFilters),
        handleApplyFilters: () => dispatch(searchbarActions.pathfindingSearch()),
        handleUpdateFilters: (checked: EdgeCheckboxType[]) => dispatch(searchbarActions.pathFiltersSaved(checked)),
        handleCancelFilters: () => dispatch(searchbarActions.pathFiltersSaved(initialFilterState.current)),
    };
};
