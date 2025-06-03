import { QueryClient, QueryClientProvider, QueryKey } from 'react-query';

interface WrapperComponentProps {
    children: React.ReactNode;
    stateMap: { key: QueryKey; data: any }[];
}

export const QueryTestWrapper: React.FC<WrapperComponentProps> = ({ children, stateMap }) => {
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

    stateMap.forEach(({ key, data }) => {
        queryClient.setQueryData(key, data);
    });

    return <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>;
};
