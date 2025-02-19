import { useNotifications } from '../../../providers/NotificationProvider/hooks';
import { apiClient } from '../../../utils';
import { ExploreQueryParams } from '../../useExploreParams';
import { ExploreGraphQueryKey, ExploreGraphQueryOptions, GraphItemMutationFn } from './utils';

// here we can mutate request, response, and handle errors
export const nodeSearchGraphQuery = (
    addNotification: ReturnType<typeof useNotifications>['addNotification'],
    paramOptions: Partial<ExploreQueryParams>,
    mutateResponse?: GraphItemMutationFn
): ExploreGraphQueryOptions => {
    const { searchType, primarySearch } = paramOptions;
    if (!primarySearch || !searchType) {
        return {
            enabled: false,
        };
    }

    return {
        queryKey: [ExploreGraphQueryKey, searchType, primarySearch],
        queryFn: ({ signal }) =>
            apiClient.getSearchResult(primarySearch, 'exact', { signal }).then((res) => {
                if (mutateResponse) {
                    const mutated = mutateResponse(res.data.data);
                    return mutated;
                }
                return res;
            }),
        onError: () => addNotification('Something special', 'someother key'),
        enabled: !!(searchType && primarySearch),
    };
};
