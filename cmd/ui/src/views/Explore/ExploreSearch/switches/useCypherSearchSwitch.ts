import { searchbarActions } from 'bh-shared-ui';
import { useAppDispatch, useAppSelector } from 'src/store';

export const useCypherSearchSwitch = () => {
    const dispatch = useAppDispatch();
    const reduxCypherQuery = useAppSelector((state) => state.search.cypher.searchTerm);

    return {
        cypherQuery: reduxCypherQuery,
        setCypherQuery: (query: string) => dispatch(searchbarActions.cypherQueryEdited(query)),
        performSearch: (query?: string) => dispatch(searchbarActions.cypherSearch(query ?? reduxCypherQuery)),
    };
};
