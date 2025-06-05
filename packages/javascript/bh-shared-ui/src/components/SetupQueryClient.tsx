import { QueryClient, QueryKey } from 'react-query';

type StateMap = { key: QueryKey; data: any };
type StateMaps = StateMap[];
export const SetUpQueryClient = (stateMaps: StateMaps) => {
    const queryClient = new QueryClient({
        defaultOptions: {
            queries: {
                retry: false,
                refetchOnMount: false,
                refetchOnWindowFocus: false,
                staleTime: Infinity,
            },
        },
    });

    stateMaps.forEach(({ key, data }) => {
        queryClient.setQueryData(key, data);
    });

    return queryClient;
};
