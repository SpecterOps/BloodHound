import { useQuery } from 'react-query';
import { apiClient } from '../utils';

export const useSSOProviders = () => {
    return useQuery(['listSSOProviders'], ({ signal }) =>
        apiClient.listSSOProviders({ signal }).then((res) => res.data.data)
    );
};
