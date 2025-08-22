// Copyright 2025 Specter Ops, Inc.
//
// Licensed under the Apache License, Version 2.0
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

import type { GetScheduledJobDisplayResponse } from 'js-client-library';
import { useEffect } from 'react';
import { useQuery } from 'react-query';
import { usePermissions } from '../usePermissions';
import { useNotifications } from '../../providers';
import { apiClient } from '../../utils/api';
import { Permission } from '../../utils/permissions';
import { PERSIST_NOTIFICATION } from '../../utils';

interface FinishedJobParams {
    page: number;
    rowsPerPage: number;
}

export const NO_PERMISSION_MESSAGE = `Your user role does not grant permission to view the finished jobs details. Please
    contact your administrator for details.`;
export const NO_PERMISSION_KEY = 'finished-jobs-permission';

export const FETCH_ERROR_MESSAGE = 'Unable to fetch jobs. Please try again.';
export const FETCH_ERROR_KEY = 'finished-jobs-error';

/** Makes a paginated request for Finished Jobs, returned as a TanStack Query */
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
        enabled: hasPermission,
        keepPreviousData: true, // Prevent count from resetting to 0 between page fetches
        onError: () => addNotification(FETCH_ERROR_MESSAGE, FETCH_ERROR_KEY),
        queryFn: () => apiClient.getFinishedJobs(rowsPerPage * page, rowsPerPage, false, false).then((res) => res.data),
        queryKey: ['finished-jobs', { page, rowsPerPage }],
    });
};
