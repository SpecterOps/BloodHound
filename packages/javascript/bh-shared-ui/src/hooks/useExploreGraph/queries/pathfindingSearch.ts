import { apiClient } from '../../../utils';
import { ExploreQueryParams } from '../../useExploreParams';
import {
    ExploreGraphQuery,
    ExploreGraphQueryError,
    ExploreGraphQueryKey,
    ExploreGraphQueryOptions,
    INITIAL_FILTER_TYPES,
    transformToFlatGraphResponse,
} from './utils';

// Only need to create our default filters once
const createPathFilterString = (types: string[]) => `in:${types.join(',')}`;
const DEFAULT_FILTERS = createPathFilterString(INITIAL_FILTER_TYPES);

export const pathfindingSearchGraphQuery = (paramOptions: Partial<ExploreQueryParams>): ExploreGraphQueryOptions => {
    const { searchType, primarySearch, secondarySearch, pathFilters } = paramOptions;

    // Query should occur whether or not pathFilters exist
    if (!primarySearch || !searchType || !secondarySearch) {
        return { enabled: false };
    }

    const filter = pathFilters?.length ? createPathFilterString(pathFilters) : DEFAULT_FILTERS;

    return {
        queryKey: [ExploreGraphQueryKey, searchType, primarySearch, secondarySearch, filter],
        queryFn: ({ signal }) => {
            return apiClient
                .getShortestPathV2(primarySearch, secondarySearch, filter, { signal })
                .then((res) => transformToFlatGraphResponse(res.data));
        },
        retry: false,
        enabled: !!(searchType && primarySearch && secondarySearch),
    };
};

const getPathfindingErrorMessage = (error: any): ExploreGraphQueryError => {
    const statusCode = error?.response?.status;
    if (statusCode === 404) {
        return { message: 'Path not found.', key: 'shortestPathNotFound' };
    } else if (statusCode === 503) {
        return {
            message:
                'Calculating the requested Attack Path exceeded memory limitations due to the complexity of paths involved.',
            key: 'shortestPathOutOfMemory',
        };
    } else if (statusCode === 504) {
        return {
            message: 'The results took too long to compute, possibly due to the complexity of paths involved.',
            key: 'shortestPathTimeout',
        };
    } else {
        return { message: 'An unknown error occurred. Please try again.', key: 'shortestPathUnknown' };
    }
};

export const pathfindingSearchQuery: ExploreGraphQuery = {
    getQueryConfig: pathfindingSearchGraphQuery,
    getErrorMessage: getPathfindingErrorMessage,
};
