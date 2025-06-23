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

import { useState } from 'react';
import { EdgeCheckboxType } from '../../edgeTypes';
import { useExploreParams } from '../useExploreParams';
import { EMPTY_FILTER_VALUE, INITIAL_FILTERS, INITIAL_FILTER_TYPES } from './queries';
import { areArraysSimilar, extractEdgeTypes, mapParamsToFilters } from './utils';

export type PathfindingFilters = ReturnType<typeof usePathfindingFilters>;

export const usePathfindingFilters = () => {
    const [selectedFilters, updateSelectedFilters] = useState<EdgeCheckboxType[]>(INITIAL_FILTERS);
    const { pathFilters, setExploreParams } = useExploreParams();

    // Instead of tracking this in an effect, we want to create a callback to let the consumer decide when to sync down
    // query params. This is useful for our filter form where we only want to sync once when the user opens it
    const resetFilters = () => {
        if (pathFilters?.length) {
            // Since we need to track state in the case of an empty set of filters, check for our 'empty' key here
            const incoming = pathFilters[0] === EMPTY_FILTER_VALUE ? [] : pathFilters;

            const mapped = mapParamsToFilters(incoming, INITIAL_FILTERS);
            updateSelectedFilters(mapped);
        } else {
            updateSelectedFilters(INITIAL_FILTERS);
        }
    };

    const handleApplyFilters = (filters = selectedFilters) => {
        const selectedEdgeTypes = extractEdgeTypes(filters);

        if (selectedEdgeTypes.length === 0) {
            // query string stores a value indicating an empty set if every option is unselected
            setExploreParams({ pathFilters: [EMPTY_FILTER_VALUE] });
        } else if (areArraysSimilar(INITIAL_FILTER_TYPES, selectedEdgeTypes)) {
            // query string is not set if user selects the default
            setExploreParams({ pathFilters: null });
        } else {
            setExploreParams({ pathFilters: selectedEdgeTypes });
        }
    };

    const handleRemoveEdgeType = (edgeType: string) => {
        const filteredTypes = selectedFilters.filter((item) => item.edgeType !== edgeType);
        handleUpdateFilters(filteredTypes);
        handleApplyFilters(filteredTypes);
    };

    const handleUpdateFilters = (checked: EdgeCheckboxType[]) => updateSelectedFilters(checked);

    return {
        handleApplyFilters,
        resetFilters,
        handleRemoveEdgeType,
        handleUpdateFilters,
        selectedFilters,
    };
};
