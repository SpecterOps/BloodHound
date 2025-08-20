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

import { useMutation, useQuery, useQueryClient } from 'react-query';
import { apiClient } from '../utils/api';

export type AcceptedIngestType = 'application/json' | 'application/zip';

export const listFileIngestJobs = (skip?: number, limit?: number, sortBy?: string) =>
    apiClient.listFileIngestJobs(skip, limit, sortBy).then((res) => res.data);

export const listFileTypesForIngest = () => apiClient.listFileTypesForIngest().then((res) => res.data);

export const startFileIngestJob = () => apiClient.startFileIngest().then((res) => res.data);

export const uploadFileToIngestJob = ({
    jobId,
    fileContents,
    contentType = 'application/json',
    options,
}: {
    jobId: string;
    fileContents: any;
    contentType?: string;
    options?: Parameters<typeof apiClient.uploadFileToIngestJob>[3];
}) => {
    const random = Math.random();
    console.log(random);
    if (random > 0.9) {
        return new Promise((res, rej) => rej(new Error('oooohohonoooo')));
    }
    return apiClient.uploadFileToIngestJob(jobId, fileContents, contentType, options).then((res) => res.data);
};

export const endFileIngestJob = ({ jobId }: { jobId: string }) =>
    apiClient.endFileIngest(jobId).then((res) => res.data);

export const fileUploadKeys = {
    all: 'file-upload' as const,
    listJobs: () => [fileUploadKeys.all, 'list-jobs'] as const,
    listJobsPaginated: (page: number, rowsPerPage: number) =>
        [...fileUploadKeys.listJobs(), page, rowsPerPage] as const,
    listFileTypes: () => [fileUploadKeys.all, 'accepted-types'] as const,
};

export const useListFileIngestJobs = (page: number, rowsPerPage: number, enabled: boolean) => {
    return useQuery(
        fileUploadKeys.listJobsPaginated(page, rowsPerPage),
        () => listFileIngestJobs(page * rowsPerPage, rowsPerPage, '-id'),
        { refetchInterval: 5000, enabled: enabled }
    );
};

export const useListFileTypesForIngest = () => {
    return useQuery(fileUploadKeys.listFileTypes(), listFileTypesForIngest, { refetchOnWindowFocus: false });
};

export const useStartFileIngestJob = () => {
    return useMutation(startFileIngestJob);
};

export const useUploadFileToIngestJob = () => {
    return useMutation(uploadFileToIngestJob);
};

export const useEndFileIngestJob = () => {
    const queryClient = useQueryClient();
    return useMutation(endFileIngestJob, {
        onSettled: () => {
            queryClient.invalidateQueries(fileUploadKeys.listJobs());
        },
    });
};
