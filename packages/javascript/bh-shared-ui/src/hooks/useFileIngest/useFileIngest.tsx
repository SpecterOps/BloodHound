// Copyright 2024 Specter Ops, Inc.
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

import { useMutation, useQueryClient } from 'react-query';
import { apiClient } from '../../utils/api';

export const fileUploadKeys = {
    base: 'file-upload' as const,
    listJobs: () => [fileUploadKeys.base, 'list-jobs'] as const,
    listJobsPaginated: (page: number, rowsPerPage: number) =>
        [...fileUploadKeys.listJobs(), page, rowsPerPage] as const,
    listFileTypes: () => [fileUploadKeys.base, 'accepted-types'] as const,
};

export const useStartFileIngestJob = () => {
    return useMutation({
        mutationFn: () => apiClient.startFileIngest().then((res) => res.data),
    });
};

export type AcceptedIngestType = 'application/json' | 'application/zip';

interface UploadFileIngestJobParams {
    jobId: string;
    fileContents: any;
    contentType?: string;
    options?: Parameters<typeof apiClient.uploadFileToIngestJob>[3];
}

export const useUploadFileToIngestJob = () => {
    return useMutation({
        mutationFn: ({ jobId, fileContents, contentType = 'application/json', options }: UploadFileIngestJobParams) =>
            apiClient.uploadFileToIngestJob(jobId, fileContents, contentType, options).then((res) => res.data),
    });
};

interface EndFileIngestJobParams {
    jobId: string;
}

export const useEndFileIngestJob = () => {
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: ({ jobId }: EndFileIngestJobParams) => apiClient.endFileIngest(jobId).then((res) => res.data),
        onSettled: () => queryClient.invalidateQueries(fileUploadKeys.listJobs()),
    });
};
