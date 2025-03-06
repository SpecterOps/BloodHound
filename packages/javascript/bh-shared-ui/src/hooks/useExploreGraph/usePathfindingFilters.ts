import { useState } from 'react';
import { EdgeCheckboxType } from '../../edgeTypes';
import { useExploreParams } from '../useExploreParams';
import { INITIAL_FILTERS, INITIAL_FILTER_TYPES } from './queries';
import { compareEdgeTypes, extractEdgeTypes, mapParamsToFilters } from './utils';

export const usePathfindingFilters = () => {
    const [selectedFilters, updateSelectedFilters] = useState<EdgeCheckboxType[]>(INITIAL_FILTERS);
    const { pathFilters, setExploreParams } = useExploreParams();

    const initialize = () => {
        if (pathFilters?.length) {
            const mapped = mapParamsToFilters(pathFilters, INITIAL_FILTERS);
            updateSelectedFilters(mapped);
        }
    };

    const handleUpdateFilters = (checked: EdgeCheckboxType[]) => updateSelectedFilters(checked);

    const handleApplyFilters = () => {
        const selectedEdgeTypes = extractEdgeTypes(selectedFilters);

        if (compareEdgeTypes(INITIAL_FILTER_TYPES, selectedEdgeTypes)) {
            setExploreParams({ pathFilters: null });
        } else {
            setExploreParams({ pathFilters: extractEdgeTypes(selectedFilters) });
        }
    };

    const handleCancelFilters = () => {
        const previous = pathFilters ? mapParamsToFilters(pathFilters, INITIAL_FILTERS) : INITIAL_FILTERS;
        updateSelectedFilters(previous);
    };

    return {
        selectedFilters,
        initialize,
        handleApplyFilters,
        handleUpdateFilters,
        handleCancelFilters,
    };
};
