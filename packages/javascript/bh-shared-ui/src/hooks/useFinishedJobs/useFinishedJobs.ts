import { GetScheduledJobDisplayResponse } from 'js-client-library';
import { useEffect } from 'react';
import { useQuery } from 'react-query';
import { useNotifications } from '../../providers';
import {
    FETCH_ERROR_KEY,
    FETCH_ERROR_MESSAGE,
    FinishedJobParams,
    NO_PERMISSION_KEY,
    NO_PERMISSION_MESSAGE,
    PERSIST_NOTIFICATION,
    Permission,
    apiClient,
} from '../../utils';
import { usePermissions } from '../usePermissions';

/** Makes a paginated request for Finished Jobs, returned as a TanStack Query */
export const useFinishedJobs = ({ page, rowsPerPage }: FinishedJobParams) => {
    const { checkPermission } = usePermissions();
    const hasPermission = checkPermission(Permission.CLIENTS_MANAGE);

    const { addNotification, dismissNotification } = useNotifications();

    useEffect(() => {
        if (!hasPermission) {
            addNotification(NO_PERMISSION_MESSAGE, NO_PERMISSION_KEY, PERSIST_NOTIFICATION);
        }

        return () => dismissNotification(NO_PERMISSION_KEY);
    }, [addNotification, dismissNotification, hasPermission]);

    return useQuery<GetScheduledJobDisplayResponse>({
        enabled: hasPermission,
        keepPreviousData: true, // Prevent count from resetting to 0 between page fetches
        onError: () => addNotification(FETCH_ERROR_MESSAGE, FETCH_ERROR_KEY),
        queryFn: () => apiClient.getFinishedJobs(rowsPerPage * page, rowsPerPage, false, false).then((res) => res.data),
        queryKey: ['finished-jobs', { page, rowsPerPage }],
    });
};
