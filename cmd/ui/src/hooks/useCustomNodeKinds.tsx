import { apiClient } from 'bh-shared-ui';
import { GetCustomNodeKindsResponse, RequestOptions } from 'js-client-library';
import { UseQueryResult, useQuery } from 'react-query';

export const getCustomNodeKinds = async (options: RequestOptions): Promise<GetCustomNodeKindsResponse> =>
    apiClient.getCustomNodeKinds(options).then((res) => res.data);

export const useGetCustomNodeKinds = (): UseQueryResult<GetCustomNodeKindsResponse> =>
    useQuery<GetCustomNodeKindsResponse>(['getCustomNodeKinds'], ({ signal }) => getCustomNodeKinds({ signal }));
