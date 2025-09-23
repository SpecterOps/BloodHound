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

import type { ListFileIngestJobsResponse } from 'js-client-library';
import { useEffect } from 'react';
import { useQuery } from 'react-query';
import { PERSIST_NOTIFICATION, useNotifications } from '../../providers';
import {
    FILE_INGEST_FETCH_ERROR_KEY,
    FILE_INGEST_FETCH_ERROR_MESSAGE,
    FILE_INGEST_NO_PERMISSION_KEY,
    FILE_INGEST_NO_PERMISSION_MESSAGE,
    FileUploadParams,
    Permission,
    apiClient,
} from '../../utils';
import { usePermissions } from '../usePermissions';

/** Makes a paginated request for File Upload Jobs, returned as a TanStack Query */
export const useGetFileUploadsQuery = ({ page, rowsPerPage, filters }: FileUploadParams) => {
    const { checkPermission, isSuccess: permissionsLoaded } = usePermissions();
    const hasPermission = permissionsLoaded && checkPermission(Permission.GRAPH_DB_INGEST);

    const { addNotification, dismissNotification } = useNotifications();

    useEffect(() => {
        if (!hasPermission) {
            addNotification(FILE_INGEST_NO_PERMISSION_MESSAGE, FILE_INGEST_NO_PERMISSION_KEY, PERSIST_NOTIFICATION);
        }

        return () => dismissNotification(FILE_INGEST_NO_PERMISSION_KEY);
    }, [addNotification, dismissNotification, hasPermission]);

    return useQuery<ListFileIngestJobsResponse>({
        enabled: Boolean(permissionsLoaded && hasPermission),
        keepPreviousData: true, // Prevent count from resetting to 0 between page fetches
        onError: () => addNotification(FILE_INGEST_FETCH_ERROR_MESSAGE, FILE_INGEST_FETCH_ERROR_KEY),
        queryFn: ({ signal }) =>
            apiClient
                .listFileIngestJobs(
                    {
                        skip: rowsPerPage * page,
                        limit: rowsPerPage,
                        sortBy: '-id',
                        ...filters,
                    },
                    { signal }
                )
                .then((res) => res.data),

        queryKey: ['file-uploads', { ...filters, page, rowsPerPage }],
    });
};
