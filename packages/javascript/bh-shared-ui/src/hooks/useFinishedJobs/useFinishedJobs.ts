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
    FETCH_ERROR_KEY,
    FETCH_ERROR_MESSAGE,
    FinishedJobParams,
    NO_PERMISSION_KEY,
    NO_PERMISSION_MESSAGE,
    Permission,
    apiClient,
} from '../../utils';
import { usePermissions } from '../usePermissions';

/** Makes a paginated request for Finished Jobs, returned as a TanStack Query */
export const useFinishedJobs = ({ page, rowsPerPage }: FinishedJobParams) => {
    const { checkPermission, isSuccess: permissionsLoaded } = usePermissions();
    const hasPermission = permissionsLoaded && checkPermission(Permission.CLIENTS_MANAGE);

    const { addNotification, dismissNotification } = useNotifications();

    useEffect(() => {
        if (permissionsLoaded && !hasPermission) {
            addNotification(NO_PERMISSION_MESSAGE, NO_PERMISSION_KEY, PERSIST_NOTIFICATION);
        }

        return () => dismissNotification(NO_PERMISSION_KEY);
    }, [addNotification, dismissNotification, hasPermission, permissionsLoaded]);

    return useQuery<GetScheduledJobDisplayResponse>({
        enabled: Boolean(permissionsLoaded && hasPermission),
        keepPreviousData: true, // Prevent count from resetting to 0 between page fetches
        onError: () => addNotification(FETCH_ERROR_MESSAGE, FETCH_ERROR_KEY),
        queryFn: () => apiClient.getFinishedJobs(rowsPerPage * page, rowsPerPage, false, false).then((res) => res.data),
        queryKey: ['finished-jobs', { page, rowsPerPage }],
    });
};
