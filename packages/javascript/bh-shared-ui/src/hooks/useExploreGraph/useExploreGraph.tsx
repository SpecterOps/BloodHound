import { useQuery } from 'react-query';
import { useNotifications } from '../../providers/NotificationProvider/hooks';
import { ExploreQueryParams, useExploreParams } from '../useExploreParams';
import { ExploreGraphQueryOptions, GraphItemMutationFn, nodeSearchGraphQuery } from './search-modes';

function getExploreGraphQuery(
    addNotification: ReturnType<typeof useNotifications>['addNotification'],
    paramOptions: Partial<ExploreQueryParams>,
    mutateResponse?: GraphItemMutationFn
): ExploreGraphQueryOptions {
    switch (paramOptions.searchType) {
        case 'node':
            return nodeSearchGraphQuery(addNotification, paramOptions, mutateResponse);
        case 'pathfinding':
            return {};
        case 'cypher':
            return {};
        case 'relationship':
            return {};
        case 'composition':
            return {};
        default:
            return { enabled: false };
    }
    // else some unidentified type, display error, set to node-search
}

// Consumer of query params example
export const useExploreGraph = <T,>(mutateResponse?: GraphItemMutationFn) => {
    const { addNotification } = useNotifications();

    const { primarySearch, secondarySearch, cypherSearch, searchType } = useExploreParams();

    const queryConfig = getExploreGraphQuery(
        addNotification,
        {
            primarySearch,
            secondarySearch,
            cypherSearch,
            searchType,
        },
        mutateResponse
    );

    const { data, isLoading, isError } = useQuery(queryConfig);

    return { data: data as T, isLoading, isError };
};
