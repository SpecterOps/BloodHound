import { useMutation, useQuery } from 'react-query';
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

export const useListFileIngestJobs = (page: number, rowsPerPage: number) => {
    return useQuery(
        ['listFileIngestJobs', page, rowsPerPage],
        () => listFileIngestJobs(page * rowsPerPage, rowsPerPage, '-id'),
        { refetchInterval: 5000 }
    );
};

export const useListFileTypesForIngest = () => {
    return useQuery(['listFileTypesForIngest'], listFileTypesForIngest);
};

export const useStartFileIngestJob = () => {
    return useMutation(startFileIngestJob);
};

export const useUploadFileToIngestJob = () => {
    return useMutation(uploadFileToIngestJob);
};

export const useEndFileIngestJob = () => {
    return useMutation(endFileIngestJob);
};
