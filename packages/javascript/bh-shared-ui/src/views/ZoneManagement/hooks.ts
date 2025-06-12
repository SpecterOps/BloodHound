import { useQuery } from 'react-query';
import { apiClient } from '../../utils';

export const useAssetGroupTags = () =>
    useQuery({
        queryKey: ['zone-management', 'tags'],
        queryFn: async ({ signal }) => {
            const response = await apiClient.getAssetGroupTags({ signal });
            return response.data.data.tags;
        },
    });
