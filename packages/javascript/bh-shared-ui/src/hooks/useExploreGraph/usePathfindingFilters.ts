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
import { compareEdgeTypes, extractEdgeTypes, mapParamsToFilters } from './utils';

export const usePathfindingFilters = () => {
    const [selectedFilters, updateSelectedFilters] = useState<EdgeCheckboxType[]>(INITIAL_FILTERS);
    const { pathFilters, pathSearchMode, deepSniffVariants, setExploreParams } = useExploreParams();

    // Local UI state for new path search configuration
    const [localPathSearchMode, setLocalPathSearchMode] = useState<'hybrid' | 'path' | 'deepsniff'>('hybrid');
    const [localDeepSniffVariants, setLocalDeepSniffVariants] = useState<('EnableDCSync' | 'EnableADCSESC3')[]>([
        'EnableDCSync',
        'EnableADCSESC3',
    ]);

    // Instead of tracking this in an effect, we want to create a callback to let the consumer decide when to sync down
    // query params. This is useful for our filter form where we only want to sync once when the user opens it
    const initialize = () => {
        if (pathFilters?.length) {
            // Since we need to track state in the case of an empty set of filters, check for our 'empty' key here
            const incoming = pathFilters[0] === EMPTY_FILTER_VALUE ? [] : pathFilters;

            const mapped = mapParamsToFilters(incoming, INITIAL_FILTERS);
            updateSelectedFilters(mapped);
        } else {
            updateSelectedFilters(INITIAL_FILTERS);
        }

        // Initialize new path search mode settings from URL (fallback to defaults)
        if (pathSearchMode === 'hybrid' || pathSearchMode === 'path' || pathSearchMode === 'deepsniff') {
            setLocalPathSearchMode(pathSearchMode);
        } else {
            setLocalPathSearchMode('hybrid');
        }
        if (deepSniffVariants && deepSniffVariants.length) {
            // ensure only recognized variants
            const sanitized = deepSniffVariants.filter(
                (v): v is 'EnableDCSync' | 'EnableADCSESC3' => v === 'EnableDCSync' || v === 'EnableADCSESC3'
            );
            if (sanitized.length) setLocalDeepSniffVariants(sanitized as any);
        } else {
            setLocalDeepSniffVariants(['EnableDCSync', 'EnableADCSESC3']);
        }
    };

    const handleUpdateFilters = (checked: EdgeCheckboxType[]) => updateSelectedFilters(checked);

    const handleApplyFilters = () => {
        const selectedEdgeTypes = extractEdgeTypes(selectedFilters);

        if (selectedEdgeTypes.length === 0) {
            // query string stores a value indicating an empty set if every option is unselected
            setExploreParams({ pathFilters: [EMPTY_FILTER_VALUE] });
        } else if (compareEdgeTypes(INITIAL_FILTER_TYPES, selectedEdgeTypes)) {
            // query string is not set if user selects the default
            setExploreParams({ pathFilters: null });
        } else {
            setExploreParams({ pathFilters: extractEdgeTypes(selectedFilters) });
        }

        // Persist path search mode (omit if default hybrid)
        if (localPathSearchMode === 'hybrid') {
            setExploreParams({ pathSearchMode: null });
        } else {
            setExploreParams({ pathSearchMode: localPathSearchMode });
        }

        // Persist deep sniff variants only if not selecting both (default)
        const bothSelected =
            localDeepSniffVariants.includes('EnableDCSync') && localDeepSniffVariants.includes('EnableADCSESC3');
        if (bothSelected) {
            setExploreParams({ deepSniffVariants: null });
        } else {
            setExploreParams({ deepSniffVariants: localDeepSniffVariants });
        }
    };

    // In our new implementation, these two functions are equivalent. Once we no longer need to support the old approach,
    // we can consider removing this.
    const handleCancelFilters = () => initialize();

    return {
        selectedFilters,
        initialize,
        handleApplyFilters,
        handleUpdateFilters,
        handleCancelFilters,
        // Expose new config state & setters for dialog UI
        localPathSearchMode,
        setLocalPathSearchMode,
        localDeepSniffVariants,
        setLocalDeepSniffVariants,
    };
};
