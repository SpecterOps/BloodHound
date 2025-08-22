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
import { useNotifications } from '../../providers';
import { apiClient, Permission, PERSIST_NOTIFICATION } from '../../utils';
import { usePermissions } from '../usePermissions';
import { fileUploadKeys } from './useFileIngest';

interface FileUploadParams {
    page: number;
    rowsPerPage: number;
}

const NO_PERMISSION_MESSAGE = `Your user role does not grant permission to view the file ingest jobs details. Please
    contact your administrator for details.`;
const NO_PERMISSION_KEY = 'file-upload-permission';

const FETCH_ERROR_MESSAGE = 'Unable to fetch file upload jobs. Please try again.';
const FETCH_ERROR_KEY = 'file-upload-error';

/** Makes a paginated request for File Upload Jobs, returned as a TanStack Query */
export const useGetFileUploadsQuery = ({ page, rowsPerPage }: FileUploadParams) => {
    const { checkPermission } = usePermissions();
    const hasPermission = checkPermission(Permission.GRAPH_DB_WRITE);

    const { addNotification, dismissNotification } = useNotifications();

    useEffect(() => {
        if (!hasPermission) {
            addNotification(NO_PERMISSION_MESSAGE, NO_PERMISSION_KEY, PERSIST_NOTIFICATION);
        }

        return () => dismissNotification(NO_PERMISSION_KEY);
    }, [addNotification, dismissNotification, hasPermission]);

    return useQuery<ListFileIngestJobsResponse>({
        onError: () => addNotification(FETCH_ERROR_MESSAGE, FETCH_ERROR_KEY),
        queryFn: () => apiClient.listFileIngestJobs(rowsPerPage * page, rowsPerPage, '-id').then((res) => res.data),
        queryKey: fileUploadKeys.listJobsPaginated(page, rowsPerPage),
        refetchInterval: 5000,
        enabled: hasPermission,
    });
};
