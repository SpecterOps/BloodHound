import { useMutation, useQueryClient } from 'react-query';
import { apiClient } from '../../utils';

type Node = {
    id: string;
    label?: string[];
    kinds?: string[];
    properties?: Record<string, any>;
};

type Edge = {
    source_object_id: string;
    target_object_id: string;
    edge_kind: string;
    properties?: Record<string, any>;
};

const clearGraphCache = (queryClient: any) => () => {
    queryClient.invalidateQueries({ queryKey: ['explore-graph-query', 'cypher'] });
};

export const useCreateNodeMutation = () => {
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: (node: Node) => {
            return apiClient.baseClient.post(`/api/v2/graph/nodes`, node);
        },
        onSuccess: clearGraphCache(queryClient),
    });
};

export const useCreateEdgeMutation = () => {
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: (edge: Edge) => {
            return apiClient.baseClient.post(`/api/v2/graph/edges`, edge);
        },
        onSuccess: clearGraphCache(queryClient),
    });
};

export const useEditNodeMutation = () => {
    return useMutation({
        mutationFn: ({ nodeId, node }: { nodeId: string; node: Record<string, any> }) => {
            return apiClient.baseClient.put(`/api/v2/graph/nodes/${nodeId}`, node);
        },
    });
};

export const useDeleteNodeMutation = () => {
    return useMutation({
        mutationFn: (nodeId: string) => {
            return apiClient.baseClient.delete(`/api/v2/graph/nodes/${nodeId}`);
        },
    });
};

export const useDeleteEdgeMutation = () => {
    return useMutation({
        mutationFn: (edge: Edge) => {
            return apiClient.baseClient.delete(`/api/v2/graph/edges/`, { data: edge });
        },
    });
};
