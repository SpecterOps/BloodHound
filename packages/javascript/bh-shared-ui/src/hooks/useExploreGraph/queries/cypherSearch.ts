import { apiClient } from '../../../utils';
import { ExploreQueryParams } from '../../useExploreParams';
import { decodeCypherQuery } from '../utils';
import {
    ExploreGraphQuery,
    ExploreGraphQueryError,
    ExploreGraphQueryKey,
    ExploreGraphQueryOptions,
    transformToFlatGraphResponse,
} from './utils';

export const cypherSearchGraphQuery = (paramOptions: Partial<ExploreQueryParams>): ExploreGraphQueryOptions => {
    const { searchType, cypherSearch } = paramOptions;

    if (!cypherSearch || !searchType) {
        return { enabled: false };
    }

    const decoded = decodeCypherQuery(cypherSearch);

    return {
        queryKey: [ExploreGraphQueryKey, searchType, cypherSearch],
        queryFn: ({ signal }) =>
            apiClient.cypherSearch(decoded, { signal }).then((res) => transformToFlatGraphResponse(res.data)),
        retry: false,
        enabled: !!(searchType && cypherSearch),
    };
};

const getCypherErrorMessage = (error: any): ExploreGraphQueryError => {
    const status = error?.response?.status;
    const message = error?.response?.data?.errors?.[0]?.message;

    if (status === 404) {
        return { message: 'No results match your criteria', key: 'CypherSearchEmptyResponse' };
    } else if (message) {
        return { message, key: 'CypherSearchBadRequest' };
    } else {
        return { message: 'An unknown error occurred.', key: 'CypherSearchUnknown' };
    }
};

export const cypherSearchQuery: ExploreGraphQuery = {
    getQueryConfig: cypherSearchGraphQuery,
    getErrorMessage: getCypherErrorMessage,
};
