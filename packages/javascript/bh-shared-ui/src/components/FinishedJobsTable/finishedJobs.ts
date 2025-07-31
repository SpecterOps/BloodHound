import type { GetScheduledJobDisplayResponse, ScheduledJobDisplay } from 'js-client-library';
import { DateTime, Interval } from 'luxon';
import { useEffect } from 'react';
import { useQuery } from 'react-query';
import { usePermissions } from '../../hooks';
import { useNotifications } from '../../providers';
import { LuxonFormat, Permission, apiClient } from '../../utils';

export interface FinishedJobParams {
    page: number;
    rowsPerPage: number;
}

export const FINISHED_JOBS_LOG_HEADERS = [
    { label: 'ID / Client / Status', width: '240px' },
    { label: 'Status Message', width: '240px' },
    { label: 'Start Time', width: '110px' },
    { label: 'Duration', width: '85px' },
    { label: 'Data Collected', width: '240px' },
];

export const COLLECTION_MAP = new Map(
    Object.entries({
        session_collection: 'Sessions',
        local_group_collection: 'Local Groups',
        ad_structure_collection: 'AD Structure',
        cert_services_collection: 'Certificate Services',
        ca_registry_collection: 'CA Registry',
        dc_registry_collection: 'DC Registry',
        all_trusted_domains: 'All Trusted Domains', // Is this supposed to be here?
    })
);

export const PERSIST_NOTIFICATION = {
    persist: true,
    anchorOrigin: { vertical: 'top', horizontal: 'right' },
} as any; // anchorOrigin is not on type but works to position notification

export const NO_PERMISSION_MESSAGE = `Your user role does not grant permission to view the finished jobs details. Please
    contact your administrator for details.`;
export const NO_PERMISSION_KEY = 'finished-jobs-permission';

export const FETCH_ERROR_MESSAGE = 'Unable to fetch jobs. Please try again.';
export const FETCH_ERROR_KEY = 'finished-jobs-error';

export const useFinishedJobsQuery = ({ page, rowsPerPage }: FinishedJobParams) => {
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
        onError: () => addNotification(FETCH_ERROR_MESSAGE, FETCH_ERROR_KEY),
        queryFn: () => apiClient.getFinishedJobs(rowsPerPage * page, rowsPerPage, false, false).then((res) => res.data),
        queryKey: ['finished-jobs', { page, rowsPerPage }],
        enabled: hasPermission,
    });
};

export const toMins = (start: string, end: string) =>
    Math.floor(Interval.fromDateTimes(DateTime.fromISO(start), DateTime.fromISO(end)).length('minutes')) + ' Min';

export const toCollected = (job: ScheduledJobDisplay) =>
    Object.entries(job)
        .filter(([key, value]) => COLLECTION_MAP.has(key) && value)
        .map(([key]) => COLLECTION_MAP.get(key))
        .join(', ');

export const toFormatted = (dateStr: string) => DateTime.fromISO(dateStr).toFormat(LuxonFormat.DATE_WITHOUT_GMT);
