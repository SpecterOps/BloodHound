import { searchbarActions, useCypherSearch, useFeatureFlag } from 'bh-shared-ui';
import { useAppDispatch, useAppSelector } from 'src/store';

export const useCypherSearchSwitch = () => {
    const { data: flag } = useFeatureFlag('back_button_support');

    const dispatch = useAppDispatch();
    const reduxCypherQuery = useAppSelector((state) => state.search.cypher.searchTerm);

    const cypherSearch = useCypherSearch();

    if (flag?.enabled) {
        return cypherSearch;
    } else {
        return {
            cypherQuery: reduxCypherQuery,
            setCypherQuery: (query: string) => dispatch(searchbarActions.cypherQueryEdited(query)),
            performSearch: () => dispatch(searchbarActions.cypherSearch(reduxCypherQuery)),
        };
    }
};
