import { apiClient } from '../../../utils';
import { ExploreQueryParams } from '../../useExploreParams';
import {
    ExploreGraphQueryKey,
    ExploreGraphQueryOptions,
    INITIAL_FILTER_TYPES,
    Notifier,
    transformToFlatGraphResponse,
} from './utils';

export const pathfindingSearchGraphQuery = (
    paramOptions: Partial<ExploreQueryParams>,
    addNotification: Notifier
): ExploreGraphQueryOptions => {
    const { searchType, primarySearch, secondarySearch, pathFilters } = paramOptions;
    if (!primarySearch || !searchType || !secondarySearch) {
        return {
            enabled: false,
        };
    }

    const filter = pathFilters?.length ? `in:${pathFilters.join(',')}` : `in:${INITIAL_FILTER_TYPES.join(',')}`;

    return {
        queryKey: [ExploreGraphQueryKey, searchType, primarySearch, secondarySearch, filter],
        queryFn: ({ signal }) => {
            return apiClient
                .getShortestPathV2(primarySearch, secondarySearch, filter, { signal })
                .then((res) => transformToFlatGraphResponse(res.data));
        },
        retry: false,
        onError: (error) => handleError(error, addNotification),
        enabled: !!(searchType && primarySearch && secondarySearch),
    };
};

const handleError = (error: any, addNotification: Notifier) => {
    const statusCode = error?.response?.status;
    if (statusCode === 404) {
        addNotification('Path not found.', 'shortestPathNotFound');
    } else if (statusCode === 503) {
        addNotification(
            'Calculating the requested Attack Path exceeded memory limitations due to the complexity of paths involved.',
            'shortestPathOutOfMemory'
        );
    } else if (statusCode === 504) {
        addNotification(
            'The results took too long to compute, possibly due to the complexity of paths involved.',
            'shortestPathTimeout'
        );
    } else {
        addNotification('An unknown error occurred. Please try again.', 'shortestPathUnknown');
    }
};
