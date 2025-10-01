import { apiClient } from 'bh-shared-ui';
import { useMutation } from 'react-query';

type Node = any;

export const useAddNodeMutation = () => {
    return useMutation({
        mutationFn: (node: Node) => {
            return apiClient.baseClient.post(`/api/v2/nodes/`, node);
        },
    });
};

export const useEditNodeMutation = () => {
    return useMutation({
        mutationFn: ({ nodeId, node }: { nodeId: string; node: Node }) => {
            return apiClient.baseClient.put(`/api/v2/nodes/${nodeId}`, node);
        },
    });
};

export const useDeleteNodeMutation = () => {
    return useMutation({
        mutationFn: ({ nodeId }: { nodeId: string }) => {
            return apiClient.baseClient.delete(`/api/v2/nodes/${nodeId}`);
        },
    });
};
