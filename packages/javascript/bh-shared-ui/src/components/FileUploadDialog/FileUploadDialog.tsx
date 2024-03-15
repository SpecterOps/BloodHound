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

import { Box, Button, Dialog, DialogActions, DialogContent } from '@mui/material';
import { useEffect, useState } from 'react';
import FileDrop from '../FileDrop';
import FileStatusListItem from '../FileStatusListItem';
import { FileForIngest, FileStatus, FileUploadStep } from './types';
import { ErrorResponse } from 'js-client-library';
import {
    useEndFileIngestJob,
    useListFileTypesForIngest,
    useStartFileIngestJob,
    useUploadFileToIngestJob,
} from '../../hooks';
import { useNotifications } from '../../providers';

const FileUploadDialog: React.FC<{
    open: boolean;
    onClose: () => void;
}> = ({ open, onClose }) => {
    const [filesForIngest, setFilesForIngest] = useState<FileForIngest[]>([]);
    const [fileUploadStep, setFileUploadStep] = useState<FileUploadStep>(FileUploadStep.ADD_FILES);
    const [submitDialogDisabled, setSubmitDialogDisabled] = useState<boolean>(false);
    const [uploadDialogDisabled, setUploadDialogDisabled] = useState<boolean>(false);
    const [uploadMessage, setUploadMessage] = useState<string>('');

    const { addNotification } = useNotifications();
    const listFileTypesForIngest = useListFileTypesForIngest();
    const startFileIngestJob = useStartFileIngestJob();
    const uploadFileToIngestJob = useUploadFileToIngestJob();
    const endFileIngestJob = useEndFileIngestJob();

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
        setFilesForIngest((prevFiles) => [...prevFiles, ...files]);
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
            const currentIngestJobId = response?.data?.id?.toString();

            // Counting errors manually here to avoid race conditions with react state updates
            let errorCount = 0;

            if (currentIngestJobId) {
                for (const ingestFile of filesForIngest) {
                    // Separate error handling so we can continue on when a file fails
                    try {
                        await uploadFile(currentIngestJobId, ingestFile);
                    } catch (error) {
                        errorCount += 1;
                    }
                    setNewFileStatus(ingestFile.file.name, FileStatus.DONE);
                }
            }
            await finishUpload(currentIngestJobId);

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
            { jobId, fileContents: ingestFile.file, contentType: ingestFile.file.type },
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
            handleAppendFiles(validatedFiles);
        }
    };

    const handleSubmit = () => {
        if (fileUploadStep === FileUploadStep.ADD_FILES) {
            setFileUploadStep(FileUploadStep.CONFIRMATION);
        } else if (fileUploadStep === FileUploadStep.CONFIRMATION) {
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
            <DialogContent>
                <>
                    {fileUploadStep === FileUploadStep.ADD_FILES && (
                        <FileDrop
                            onDrop={handleFileDrop}
                            disabled={listFileTypesForIngest.isLoading}
                            accept={listFileTypesForIngest.data?.data}
                        />
                    )}
                    {(fileUploadStep === FileUploadStep.CONFIRMATION || fileUploadStep === FileUploadStep.UPLOAD) && (
                        <Box fontSize={20} marginBottom={5}>
                            {uploadMessage ||
                                'The following files will be uploaded and ingested into BloodHound. This cannot be undone.'}
                        </Box>
                    )}

                    {filesForIngest.length > 0 && (
                        <Box sx={{ marginTop: 1, marginBottom: 1 }}>
                            <Box sx={{ backgroundColor: 'black', color: 'white', fontWeight: 'bold', padding: '4px' }}>
                                Files
                            </Box>
                            {filesForIngest.map((file, index) => {
                                return (
                                    <FileStatusListItem
                                        file={file}
                                        key={index}
                                        onRemove={() => handleRemoveFile(index)}
                                    />
                                );
                            })}
                        </Box>
                    )}

                    {fileUploadStep === FileUploadStep.CONFIRMATION && (
                        <Box fontSize={20} marginTop={3}>
                            Press "Upload" to continue.
                        </Box>
                    )}
                </>
            </DialogContent>
            <DialogActions>
                {(fileUploadStep === FileUploadStep.ADD_FILES || fileUploadStep === FileUploadStep.CONFIRMATION) && (
                    <>
                        <Button autoFocus color='inherit' onClick={onClose} data-testid='confirmation-dialog_button-no'>
                            Cancel
                        </Button>
                        <Button
                            color='primary'
                            disabled={submitDialogDisabled}
                            onClick={handleSubmit}
                            data-testid='confirmation-dialog_button-yes'>
                            Upload
                        </Button>
                    </>
                )}
                {fileUploadStep === FileUploadStep.UPLOAD && (
                    <Button
                        color='primary'
                        onClick={onClose}
                        disabled={uploadDialogDisabled}
                        data-testid='confirmation-dialog_button-yes'>
                        {uploadDialogDisabled ? 'Uploading Files' : 'Close'}
                    </Button>
                )}
            </DialogActions>
        </Dialog>
    );
};

export default FileUploadDialog;
