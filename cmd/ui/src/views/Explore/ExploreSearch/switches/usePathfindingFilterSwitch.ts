import { EdgeCheckboxType, searchbarActions, useFeatureFlag, usePathfindingFilters } from 'bh-shared-ui';
import { useRef } from 'react';
import { useAppDispatch, useAppSelector } from 'src/store';

export const usePathfindingFilterSwitch = () => {
    const { data: flag } = useFeatureFlag('back_button_support');

    const dispatch = useAppDispatch();
    const initialFilterState = useRef<EdgeCheckboxType[]>([]);
    const reduxPathfindingFilters = useAppSelector((state) => state.search.pathFilters);

    const pathfindingFilters = usePathfindingFilters();

    if (flag?.enabled) {
        return pathfindingFilters;
    } else {
        return {
            selectedFilters: reduxPathfindingFilters,
            initialize: () => (initialFilterState.current = reduxPathfindingFilters),
            handleApplyFilters: () => dispatch(searchbarActions.pathfindingSearch()),
            handleUpdateFilters: (checked: EdgeCheckboxType[]) => dispatch(searchbarActions.pathFiltersSaved(checked)),
            handleCancelFilters: () => dispatch(searchbarActions.pathFiltersSaved(initialFilterState.current)),
        };
    }
};
