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

import { Box, Button } from '@mui/material';
import { ContentPage, FileForIngest, FileStatus, FileUploadDialog, FinishedIngestLog } from 'bh-shared-ui';
import { useEffect, useState } from 'react';
import useToggle from 'src/hooks/useToggle';
import { apiClient } from 'bh-shared-ui';
import { useAppDispatch } from 'src/store';
import { addSnackbar } from 'src/ducks/global/actions';
import { useQuery } from 'react-query';

const FileIngest = () => {
    const dispatch = useAppDispatch();
    const [filesForIngest, setFilesForIngest] = useState<FileForIngest[]>([]);
    const [submitDialogDisabled, setSubmitDialogDisabled] = useState<boolean>(false);
    const [uploadMessage, setUploadMessage] = useState<string>('');
    const [fileUploadDialogOpen, toggleFileUploadDialog] = useToggle(false);
    const [page, setPage] = useState(0);
    const [rowsPerPage, setRowsPerPage] = useState(10);
    const [count, setCount] = useState(0);

    const { data: listFileIngestJobsData, refetch: refetchIngestJobs } = useQuery(
        ['listFileIngestJobs', page, rowsPerPage],
        () =>
            apiClient.listFileIngestJobs(page * rowsPerPage, rowsPerPage, '-id').then((res) => {
                setCount(res.data?.count || 0);
                return res.data;
            }),
        {
            refetchInterval: 5000,
        }
    );

    const handlePageChange: (event: React.MouseEvent<HTMLButtonElement> | null, page: number) => void = (
        _event,
        newPage
    ) => {
        setPage(newPage);
    };

    const handleRowsPerPageChange: React.ChangeEventHandler<HTMLTextAreaElement | HTMLInputElement> = (event) => {
        setRowsPerPage(parseInt(event.target.value, 10));
        setPage(0);
    };

    useEffect(() => {
        const filesHaveErrors = filesForIngest.filter((file) => file.errors).length > 0;
        const filesAreUploading = filesForIngest.filter((file) => file.status === FileStatus.UPLOADING).length > 0;

        if (filesHaveErrors || filesAreUploading || !filesForIngest.length) {
            setSubmitDialogDisabled(true);
        } else {
            setSubmitDialogDisabled(false);
        }
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
                    return { ...file, error: [error] };
                }
                return file;
            })
        );
    };

    const handleUploadAllFiles = async () => {
        try {
            updateStatusOfReadyFiles(FileStatus.UPLOADING);

            const jobId = await startUpload();

            for (const ingestFile of filesForIngest) {
                await uploadFile(jobId, ingestFile);
            }

            await finishUpload(jobId);
        } catch (error) {
            console.error(error);
        }
    };

    const startUpload = async () => {
        try {
            const response = await apiClient.startFileIngest();
            return await response.data.data.id;
        } catch (error) {
            dispatch(addSnackbar('Failed to start ingest process', 'StartFileIngestFail'));
            setFilesForIngest((prevFiles) => prevFiles.map((file) => ({ ...file, status: FileStatus.READY })));
            throw new Error('Failed to start file upload');
        }
    };

    const uploadFile = async (jobId: string, ingestFile: FileForIngest) => {
        try {
            const text = await ingestFile.file.text();
            await apiClient.uploadFileToIngestJob(jobId, JSON.parse(text));
            setNewFileStatus(ingestFile.file.name, FileStatus.DONE);
        } catch (error) {
            dispatch(addSnackbar(`File upload failed for ${ingestFile.file.name}`, 'IngestFileUploadFail'));
            setUploadFailureError(ingestFile.file.name, 'Upload Failed');
        }
    };

    const finishUpload = async (jobId: string) => {
        try {
            const filesWithErrors = filesForIngest.filter((file) => file.errors);
            const uploadMessage =
                filesWithErrors.length > 0
                    ? 'Some files have failed to upload and have not been included for ingest.'
                    : 'All files have successfully been uploaded for ingest.';

            await apiClient.endFileIngest(jobId);
            refetchIngestJobs();

            setUploadMessage(uploadMessage);
            dispatch(
                addSnackbar(
                    `Successfully uploaded ${filesForIngest.length - filesWithErrors.length} files for ingest`,
                    'FileIngestSuccess'
                )
            );
        } catch (error) {
            dispatch(addSnackbar('Failed to close out ingest job', 'EndFileIngestFail'));
        }
    };

    const handleExit = () => setFilesForIngest([]);

    return (
        <>
            <ContentPage title='Manual File Ingest' data-testid='manual-file-ingest'>
                <Box display='flex' justifyContent='flex-end' alignItems='center' minHeight='24px' mb={2}>
                    <Button
                        color='primary'
                        variant='contained'
                        disableElevation
                        onClick={() => toggleFileUploadDialog()}
                        data-testid='file-ingest_button-upload-files'>
                        Upload File(s)
                    </Button>
                </Box>
            </ContentPage>

            <ContentPage title='Finished Ingest Log' data-testid='finished-ingest-log'>
                <FinishedIngestLog
                    ingestJobs={listFileIngestJobsData?.data || []}
                    paginationProps={{
                        page,
                        rowsPerPage,
                        count,
                        onPageChange: handlePageChange,
                        onRowsPerPageChange: handleRowsPerPageChange,
                    }}
                />
            </ContentPage>

            <FileUploadDialog
                open={fileUploadDialogOpen}
                onClose={toggleFileUploadDialog}
                files={filesForIngest}
                submitDisabled={submitDialogDisabled}
                onAppendFiles={handleAppendFiles}
                onRemoveFile={handleRemoveFile}
                onUpload={handleUploadAllFiles}
                onExited={handleExit}
                uploadMessage={uploadMessage}
            />
        </>
    );
};

export default FileIngest;
