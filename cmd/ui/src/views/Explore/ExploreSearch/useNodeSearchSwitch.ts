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
