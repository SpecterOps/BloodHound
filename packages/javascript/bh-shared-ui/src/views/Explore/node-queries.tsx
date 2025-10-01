import { useMutation } from 'react-query';
import { apiClient } from '../../utils';

type Node = {
    object_id: string;
    label?: string[];
    properties?: Record<string, any>;
};

type Edge = {
    source_object_id: string;
    target_object_id: string;
    edge_kind: string;
    properties?: Record<string, any>;
};

export const useCreateNodeMutation = () => {
    return useMutation({
        mutationFn: (node: Node) => {
            return apiClient.baseClient.post(`/api/v2/graph/nodes/`, node);
        },
    });
};

export const useCreateEdgeMutation = () => {
    return useMutation({
        mutationFn: (edge: Edge) => {
            return apiClient.baseClient.delete(`/api/v2/graph/edges/`, { data: edge });
        },
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
