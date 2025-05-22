import { RequestOptions } from 'js-client-library';
import { useQuery, UseQueryResult } from 'react-query';
import { apiClient, GenericQueryOptions } from '../../utils';

export const getDataAvailable = async (options: RequestOptions): Promise<boolean> =>
    apiClient.cypherSearch("MATCH (A) WHERE NOT A:MigrationData RETURN A LIMIT 1", options).then((res) => {
        return (Object.keys(res?.data?.data?.nodes).length > 0)
    });

export const useDataAvailable = (
    queryOptions?: GenericQueryOptions<boolean>
): UseQueryResult<boolean> => {
    return useQuery({
        queryKey: ['getDataAvailable'],
        queryFn: ({ signal }) => getDataAvailable({ signal }),
        ...queryOptions,
    });
};
