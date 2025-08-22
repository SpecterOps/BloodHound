import { RequestOptions } from 'js-client-library';
import { useQuery } from 'react-query';
import { apiClient } from '../../utils/api';

const getSelf = (options?: RequestOptions) => apiClient.getSelf(options).then((res) => res.data.data);

export const useSelf = () => {
    const getSelfId = useQuery(['getSelf'], ({ signal }) => getSelf({ signal }), {
        cacheTime: Number.POSITIVE_INFINITY,
        select: (data) => {
            return data.id;
        },
    });

    const getSelfRoles = useQuery(['getSelf'], ({ signal }) => getSelf({ signal }), {
        cacheTime: Number.POSITIVE_INFINITY,
        select: (data) => {
            const userRoles = data?.roles.map((role: any) => role.name) || [];
            return userRoles;
        },
    });

    const isAdminOrPowerUser =
        getSelfRoles?.data?.includes('Administrator') || getSelfRoles?.data?.includes('Power User');

    return { getSelfId, getSelfRoles, isAdminOrPowerUser };
};
