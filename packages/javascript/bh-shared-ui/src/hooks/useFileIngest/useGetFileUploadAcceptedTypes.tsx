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

import type { ListFileTypesForIngestResponse } from 'js-client-library';
import { useQuery } from 'react-query';
import { useNotifications } from '../../providers';
import { apiClient } from '../../utils';
import { fileUploadKeys } from './useFileIngest';

const FETCH_ERROR_MESSAGE = 'Unable to fetch file upload accepted types. Please try again.';
const FETCH_ERROR_KEY = 'file-upload-accepted-types-error';

/** Makes a request for File Upload Accepted Types, returned as a TanStack Query */
export const useGetFileUploadAcceptedTypesQuery = () => {
    const { addNotification } = useNotifications();

    return useQuery<ListFileTypesForIngestResponse>({
        onError: () => addNotification(FETCH_ERROR_MESSAGE, FETCH_ERROR_KEY),
        queryFn: () => apiClient.listFileTypesForIngest().then((res) => res.data),
        queryKey: fileUploadKeys.listFileTypes(),
    });
};
