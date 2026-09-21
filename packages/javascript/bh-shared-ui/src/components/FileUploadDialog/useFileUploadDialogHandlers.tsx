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
import { ErrorResponse } from 'js-client-library';
import { useCallback, useEffect, useMemo, useState } from 'react';
import {
    useEndFileIngestJob,
    useGetFileUploadAcceptedTypesQuery,
    useStartFileIngestJob,
    useUploadFileToIngestJob,
} from '../../hooks';
import { useNotifications } from '../../providers';
import { FileForIngest, FileStatus, FileUploadStep } from './types';

export const makeProgressCacheKey = (jobId: string, fileName: string) => `job-${jobId}-file-${fileName}`;

export type UploadProgress = {
    loaded: number;
    total?: number;
};

export const calculateUploadProgress = (progressEvent: UploadProgress) => {
    const { loaded, total } = progressEvent;
    const rawPercent = typeof total === 'number' && total > 0 ? (loaded * 100) / total : 0;
    return Math.max(0, Math.min(100, Math.floor(rawPercent)));
};

export const useFileUploadDialogHandlers = ({
    onCloseProp,
    hasPermissionToUpload,
}: {
    onCloseProp: () => void;
    hasPermissionToUpload: boolean;
}) => {
    const [filesForIngest, setFilesForIngest] = useState<FileForIngest[]>([]);
    const [fileUploadStep, setFileUploadStep] = useState<FileUploadStep>(FileUploadStep.ADD_FILES);

    const [submitDialogDisabled, setSubmitDialogDisabled] = useState<boolean>(false);
    const [uploadDialogDisabled, setUploadDialogDisabled] = useState<boolean>(false);
    const [uploadMessage, setUploadMessage] = useState<string>('');
    const [currentIngestJobId, setCurrentIngestJobId] = useState('');
    const [progressCache, setProgressCache] = useState<Record<string, number>>({});

    const { addNotification } = useNotifications();
    const getFileUploadAcceptedTypes = useGetFileUploadAcceptedTypesQuery({ enabled: hasPermissionToUpload });
    const startFileIngestJob = useStartFileIngestJob();
    const uploadFileToIngestJob = useUploadFileToIngestJob();
    const endFileIngestJob = useEndFileIngestJob();

    const resetModal = () => {
        setSubmitDialogDisabled(false);
        setUploadMessage('');
        setProgressCache({});
        setCurrentIngestJobId('');
    };

    const onClose = useCallback(() => {
        resetModal();

        onCloseProp();
    }, [onCloseProp]);

    useEffect(() => {
        const filesHaveErrors = filesForIngest.filter((file) => file.errors).length > 0;
        const filesAreUploading = filesForIngest.filter((file) => file.status === FileStatus.UPLOADING).length > 0;
        const noFilesAreReady = filesForIngest.every((file) => file.status !== FileStatus.READY);

        const shouldDisableSubmit = filesHaveErrors || !filesForIngest.length || noFilesAreReady;
        setSubmitDialogDisabled(shouldDisableSubmit);
        setUploadDialogDisabled(filesAreUploading);
    }, [filesForIngest]);

    const handleRemoveFile = (index: number) => {
        setFilesForIngest((prevFiles) => prevFiles.filter((_file, i) => i !== index));
    };

    const handleAppendFiles = (files: FileForIngest[]) => {
        setFilesForIngest((prevFiles) => {
            const unfinishedPreviousFiles = prevFiles.filter((file) => file.status === FileStatus.READY);
            return [...unfinishedPreviousFiles, ...files];
        });
    };

    const updateStatusOfReadyFiles = (status: FileStatus) => {
        setFilesForIngest((prevFiles) =>
            prevFiles.map((file) => {
                return {
                    ...file,
                    status: file.status === FileStatus.READY ? status : file.status,
                };
            })
        );
    };

    const setNewFileStatus = (name: string, status: FileStatus) => {
        setFilesForIngest((prevFiles) =>
            prevFiles.map((file) => {
                if (file.file.name === name) {
                    return { ...file, status };
                }
                return file;
            })
        );
    };

    const setUploadFailureError = (name: string, error: string) => {
        setNewFileStatus(name, FileStatus.FAILURE);

        setFilesForIngest((prevFiles) =>
            prevFiles.map((file) => {
                if (file.file.name === name) {
                    return { ...file, errors: [error] };
                }
                return file;
            })
        );
    };

    const handleUploadAllFiles = async () => {
        updateStatusOfReadyFiles(FileStatus.UPLOADING);

        try {
            const response = await startUpload();
            const jobId = response?.data?.id?.toString() ?? '';

            setCurrentIngestJobId(jobId);
            // Counting errors manually here to avoid race conditions with react state updates
            let errorCount = 0;

            if (jobId) {
                const uploadPromises = filesForIngest.map((ingestFile) =>
                    uploadFile(jobId, ingestFile)
                        .then(() => {
                            setNewFileStatus(ingestFile.file.name, FileStatus.DONE);
                        })
                        .catch(() => {
                            setNewFileStatus(ingestFile.file.name, FileStatus.FAILURE);
                            // onError handler already sets failure/error state; we just count it here
                            errorCount += 1;
                        })
                );
                await Promise.allSettled(uploadPromises);
            }

            if (jobId) {
                await finishUpload(jobId);
            }

            logFinishedIngestJob(errorCount);
        } catch (error) {
            console.error(error);
        }
    };

    const startUpload = async () => {
        if (hasPermissionToUpload) {
            return startFileIngestJob.mutateAsync(undefined, {
                onError: () => {
                    addNotification('Failed to start ingest process', 'StartFileIngestFail');
                    setFilesForIngest((prevFiles) => prevFiles.map((file) => ({ ...file, status: FileStatus.READY })));
                },
            });
        }
    };

    const retryUploadSingleFile = async (file: FileForIngest) => {
        try {
            await uploadFile(currentIngestJobId, file);
            setNewFileStatus(file.file.name, FileStatus.DONE);
        } catch (e) {
            console.error(e);
        }
    };

    const uploadFile = async (jobId: string, ingestFile: FileForIngest) => {
        if (hasPermissionToUpload) {
            return uploadFileToIngestJob.mutateAsync(
                {
                    jobId,
                    fileContents: ingestFile.file,
                    contentType: ingestFile.file.type,
                    options: {
                        onUploadProgress: (progressEvent) => {
                            setProgressCache((prevProgressCache) => ({
                                ...prevProgressCache,
                                [makeProgressCacheKey(jobId, ingestFile?.file?.name)]:
                                    calculateUploadProgress(progressEvent),
                            }));
                        },
                    },
                },
                {
                    onError: (error: any) => {
                        const apiError = error?.response?.data as ErrorResponse;

                        if (apiError?.errors?.length && apiError.errors[0].message?.length) {
                            const { message } = apiError.errors[0];
                            addNotification(`Upload failed: ${message}`, 'IngestFileUploadFail');
                            setUploadFailureError(ingestFile.file.name, message);
                        } else {
                            addNotification(`File upload failed for ${ingestFile.file.name}`, 'IngestFileUploadFail');
                            setUploadFailureError(ingestFile.file.name, 'Upload Failed');
                        }
                    },
                }
            );
        }
    };

    const finishUpload = async (jobId: string) => {
        if (hasPermissionToUpload) {
            return endFileIngestJob.mutateAsync(
                { jobId },
                {
                    onError: () => {
                        addNotification('Failed to close out ingest job', 'EndFileIngestFail');
                    },
                    onSettled: () => {
                        setFileUploadStep(FileUploadStep.ADD_FILES);
                    },
                }
            );
        }
    };

    const logFinishedIngestJob = (errorCount: number) => {
        const uploadMessage =
            errorCount > 0
                ? 'Some files have failed to upload and have not been included for ingest.'
                : 'All files have successfully been uploaded for ingest.';
        setUploadMessage(uploadMessage);

        addNotification(
            `Successfully uploaded ${filesForIngest.length - errorCount} files for ingest`,
            'FileIngestSuccess'
        );
    };

    const handleFileDrop = (files: FileList | null) => {
        if (files && files.length > 0) {
            const validatedFiles: FileForIngest[] = [...files].map((file) => {
                if (getFileUploadAcceptedTypes.data?.data.includes(file.type)) {
                    return { file, status: FileStatus.READY };
                } else {
                    return { file, errors: ['invalid file type'], status: FileStatus.READY };
                }
            });

            resetModal();
            handleAppendFiles(validatedFiles);
        }
    };

    const handleSubmit = () => {
        setSubmitDialogDisabled(true);
        if (fileUploadStep === FileUploadStep.ADD_FILES) {
            setFileUploadStep(FileUploadStep.UPLOAD);
            handleUploadAllFiles();
        }
    };

    const currentlyUploading = useMemo(
        () => fileUploadStep === FileUploadStep.UPLOAD && !uploadMessage,
        [fileUploadStep, uploadMessage]
    );

    return {
        setFilesForIngest,
        setFileUploadStep,
        handleFileDrop,
        currentlyUploading,
        getFileUploadAcceptedTypes,
        uploadMessage,
        onClose,
        filesForIngest,
        progressCache,
        uploadDialogDisabled,
        submitDialogDisabled,
        handleSubmit,
        handleRemoveFile,
        currentIngestJobId,
        retryUploadSingleFile,
    };
};
