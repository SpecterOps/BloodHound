import { useMutation, useQuery, useQueryClient } from 'react-query';
import { apiClient } from '../utils/api';

export const listFileIngestJobs = (skip?: number, limit?: number, sortBy?: string) =>
    apiClient.listFileIngestJobs(skip, limit, sortBy).then((res) => res.data);

export const listFileTypesForIngest = () => apiClient.listFileTypesForIngest().then((res) => res.data);

export const startFileIngestJob = () => apiClient.startFileIngest().then((res) => res.data);

export const uploadFileToIngestJob = ({
    jobId,
    fileContents,
    contentType = 'application/json',
}: {
    jobId: string;
    fileContents: any;
    contentType?: string;
}) => {
    return apiClient.uploadFileToIngestJob(jobId, fileContents, contentType).then((res) => res.data);
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

export const useListFileIngestJobs = (page: number, rowsPerPage: number) => {
    return useQuery(
        fileUploadKeys.listJobsPaginated(page, rowsPerPage),
        () => listFileIngestJobs(page * rowsPerPage, rowsPerPage, '-id'),
        { refetchInterval: 5000 }
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
