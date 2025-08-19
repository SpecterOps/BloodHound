// Copyright 2023 Specter Ops, Inc.
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

import { Button } from '@bloodhoundenterprise/doodleui';
import { Box, Dialog, DialogActions, DialogContent } from '@mui/material';
import { ErrorResponse } from 'js-client-library';
import { ReactNode, useCallback, useEffect, useRef, useState } from 'react';
import { Link } from 'react-router-dom';
import {
    useEndFileIngestJob,
    useListFileTypesForIngest,
    useOnClickOutside,
    useStartFileIngestJob,
    useUploadFileToIngestJob,
} from '../../hooks';
import { useNotifications } from '../../providers';
import FileDrop from '../FileDrop';
import FileStatusListItem from '../FileStatusListItem';
import { FileForIngest, FileStatus, FileUploadStep } from './types';

const makeProgressCacheKey = (jobId: string, fileName: string) => `job-${jobId}-file-${fileName}`;

const FileUploadDialog: React.FC<{
    open: boolean;
    onClose: () => void;
    headerText?: ReactNode;
    description?: ReactNode;
}> = ({ open, onClose: onCloseProp, headerText = 'Upload Files', description }) => {
    const [filesForIngest, setFilesForIngest] = useState<FileForIngest[]>([]);
    const [fileUploadStep, setFileUploadStep] = useState<FileUploadStep>(FileUploadStep.ADD_FILES);
    const [submitDialogDisabled, setSubmitDialogDisabled] = useState<boolean>(false);
    const [uploadDialogDisabled, setUploadDialogDisabled] = useState<boolean>(false);
    const [uploadMessage, setUploadMessage] = useState<string>('');
    const [currentIngestJobId, setCurrentIngestJobId] = useState('');
    const [progressCache, setProgressCache] = useState<Record<string, number>>({});

    const { addNotification } = useNotifications();
    const listFileTypesForIngest = useListFileTypesForIngest();
    const startFileIngestJob = useStartFileIngestJob();
    const uploadFileToIngestJob = useUploadFileToIngestJob();
    const endFileIngestJob = useEndFileIngestJob();

    const onClose = useCallback(() => {
        setProgressCache({});
        onCloseProp();
    }, [onCloseProp]);

    const dialogRef = useRef<HTMLDivElement>(null);
    useOnClickOutside(dialogRef, onClose);

    useEffect(() => {
        const filesHaveErrors = filesForIngest.filter((file) => file.errors).length > 0;
        const filesAreUploading = filesForIngest.filter((file) => file.status === FileStatus.UPLOADING).length > 0;

        const shouldDisableSubmit = filesHaveErrors || !filesForIngest.length;
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
            const jobId = response?.data?.id?.toString();

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
        return startFileIngestJob.mutateAsync(undefined, {
            onError: () => {
                addNotification('Failed to start ingest process', 'StartFileIngestFail');
                setFilesForIngest((prevFiles) => prevFiles.map((file) => ({ ...file, status: FileStatus.READY })));
            },
        });
    };

    const uploadFile = async (jobId: string, ingestFile: FileForIngest) => {
        return uploadFileToIngestJob.mutateAsync(
            {
                jobId,
                fileContents: ingestFile.file,
                contentType: ingestFile.file.type,
                options: {
                    onUploadProgress: (progressEvent) => {
                        const { loaded, total = 0 } = progressEvent;
                        const percentCompleted = Math.floor((loaded * 100) / total);
                        setProgressCache((prevProgressCache) => ({
                            ...prevProgressCache,
                            [makeProgressCacheKey(jobId, ingestFile?.file?.name)]: percentCompleted,
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
    };

    const finishUpload = async (jobId: string) => {
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
                if (listFileTypesForIngest.data?.data.includes(file.type)) {
                    return { file, status: FileStatus.READY };
                } else {
                    return { file, errors: ['invalid file type'], status: FileStatus.READY };
                }
            });

            setUploadMessage('');
            setProgressCache({});
            setCurrentIngestJobId('');
            // remove from lists items already attempted
            handleAppendFiles(validatedFiles);
        }
    };

    const handleSubmit = () => {
        if (fileUploadStep === FileUploadStep.ADD_FILES) {
            setFileUploadStep(FileUploadStep.UPLOAD);
            handleUploadAllFiles();
        }
    };

    return (
        <Dialog
            open={open}
            fullWidth={true}
            maxWidth={'sm'}
            TransitionProps={{
                onExited: () => {
                    setFileUploadStep(FileUploadStep.ADD_FILES);
                    setFilesForIngest([]);
                },
            }}>
            <div ref={dialogRef}>
                <DialogContent>
                    <div className='pb-2 font-bold'>{headerText}</div>
                    {description && <div>{description}</div>}
                    <>
                        <FileDrop
                            onDrop={handleFileDrop}
                            disabled={listFileTypesForIngest.isLoading}
                            accept={listFileTypesForIngest.data?.data}
                        />
                        {uploadMessage && <Box className='mt-2 mb-2'>{uploadMessage}</Box>}
                        <Link to='/administration/file-ingest' onClick={onClose}>
                            <div className='text-center m-2 p-2 hover:bg-slate-200 rounded-md'>
                                View File Ingest History
                            </div>
                        </Link>

                        {filesForIngest.length > 0 && (
                            <Box sx={{ my: '8px' }}>
                                {filesForIngest.map((file, index) => {
                                    return (
                                        <FileStatusListItem
                                            file={file}
                                            percentCompleted={
                                                progressCache[
                                                    makeProgressCacheKey(currentIngestJobId, file?.file?.name)
                                                ] || 0
                                            }
                                            key={index}
                                            onRemove={() => handleRemoveFile(index)}
                                        />
                                    );
                                })}
                            </Box>
                        )}
                    </>
                    {fileUploadStep === FileUploadStep.UPLOAD && !uploadMessage && (
                        <div>
                            <p>Upload in progress.</p>
                            <p>
                                You can continue using the platform&mdash;we will alert you once the upload is complete.
                            </p>
                        </div>
                    )}
                </DialogContent>
                <DialogActions>
                    <Button variant='tertiary' onClick={onClose} data-testid='confirmation-dialog_button-no'>
                        {uploadDialogDisabled ? 'Uploading Files' : 'Close'}
                    </Button>
                    <Button
                        disabled={submitDialogDisabled}
                        onClick={handleSubmit}
                        data-testid='confirmation-dialog_button-yes'>
                        Upload
                    </Button>
                </DialogActions>
            </div>
        </Dialog>
    );
};

export default FileUploadDialog;
