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
import { PERSIST_NOTIFICATION, useNotifications } from '../../providers';
import {
    FINISHED_JOBS_FETCH_ERROR_KEY,
    FINISHED_JOBS_FETCH_ERROR_MESSAGE,
    FINISHED_JOBS_NO_PERMISSION_KEY,
    FINISHED_JOBS_NO_PERMISSION_MESSAGE,
    FinishedJobParams,
    Permission,
    apiClient,
} from '../../utils';
import { usePermissions } from '../usePermissions';

/** Makes a paginated request for Finished Jobs, returned as a TanStack Query */
export const useFinishedJobs = ({ filters = {}, page, rowsPerPage }: FinishedJobParams) => {
    const { checkPermission, isSuccess: permissionsLoaded } = usePermissions();
    const hasPermission = permissionsLoaded && checkPermission(Permission.CLIENTS_MANAGE);

    const { addNotification, dismissNotification } = useNotifications();

    useEffect(() => {
        if (permissionsLoaded && !hasPermission) {
            addNotification(FINISHED_JOBS_NO_PERMISSION_MESSAGE, FINISHED_JOBS_NO_PERMISSION_KEY, PERSIST_NOTIFICATION);
        }

        return () => dismissNotification(FINISHED_JOBS_NO_PERMISSION_KEY);
    }, [addNotification, dismissNotification, hasPermission, permissionsLoaded]);

    return useQuery<GetScheduledJobDisplayResponse>({
        enabled: Boolean(permissionsLoaded && hasPermission),
        keepPreviousData: true, // Prevent count from resetting to 0 between page fetches
        onError: () => addNotification(FINISHED_JOBS_FETCH_ERROR_MESSAGE, FINISHED_JOBS_FETCH_ERROR_KEY),
        queryFn: ({ signal }) =>
            apiClient
                .getFinishedJobs(
                    {
                        ...filters,
                        skip: rowsPerPage * page,
                        limit: rowsPerPage,
                        hydrate_domains: false,
                        hydrate_ous: false,
                    },
                    { signal }
                )
                .then((res) => res.data),
        queryKey: ['finished-jobs', { ...filters, page, rowsPerPage }],
    });
};
