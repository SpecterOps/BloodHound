// Copyright 2026 Specter Ops, Inc.
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
import { isAxiosError, RequestOptions } from 'js-client-library';
import { useState } from 'react';
import { useMutation } from 'react-query';
import { useNotifications } from '../../providers';
import { apiClient } from '../../utils';
import { FileForIngest, FileStatus } from '../FileUploadDialog';
import { calculateUploadProgress } from '../FileUploadDialog/useFileUploadDialogHandlers';

export const useSchemaUploadHandlers = () => {
    const [file, setFile] = useState<FileForIngest | null>(null);
    const [uploadProgress, setUploadProgress] = useState<number>(0);
    const uploadSchemaFile = useUploadSchemaFile();
    const { addNotification } = useNotifications();

    // Validates that only one file has been dropped, and if so adds it to the dialog in the "Ready" status
    const handleFileDrop = (files: FileList | null) => {
        if (!files || files.length === 0) return;

        if (files?.length === 1) {
            setFile({ file: files[0], status: FileStatus.READY });
            resetFileUploadProgress();
        } else {
            addNotification('Currently only supports single file uploads', 'MultipleFileError');
        }
    };

    const resetFileUploadProgress = () => setUploadProgress(0);

    // Clears out any selected file and resets to default state
    const resetDialog = () => {
        setFile(null);
        resetFileUploadProgress();
    };

    // Attempts to upload the current file and manages the file's status through the upload process
    const handleUpload = () => {
        if (!file) return;

        setUploadProgress(0);
        setNewFileStatus(FileStatus.UPLOADING);

        return uploadSchemaFile.mutate(
            {
                file: file.file,
                options: {
                    headers: {
                        'Content-Type': file.file.type || 'application/json',
                    },
                    onUploadProgress: (progressEvent) => {
                        const percentCompleted = calculateUploadProgress(progressEvent);
                        setUploadProgress(percentCompleted);
                    },
                },
            },
            {
                onError: (err) => {
                    handleUploadError(err);
                    setNewFileStatus(FileStatus.FAILURE);
                    resetFileUploadProgress();
                },
                onSuccess: () => setNewFileStatus(FileStatus.DONE),
            }
        );
    };

    const handleUploadError = (err: unknown) => {
        if (err && isAxiosError(err)) {
            addNotification(err.response?.data?.errors?.[0]?.message ?? 'Schema upload failed', 'SchemaUploadFailure');
        } else {
            addNotification(`An error occurred: ${(err as Error)?.message ?? 'Unknown error'}`, 'SchemaUploadFailure');
        }
    };

    const setNewFileStatus = (status: FileStatus) => {
        setFile((file) => (file ? { ...file, status } : null));
    };

    return {
        file,
        uploadProgress,
        handleFileDrop,
        handleUpload,
        resetDialog,
    };
};

interface UploadSchemaParams {
    file: File;
    options?: RequestOptions;
}

const useUploadSchemaFile = () => {
    return useMutation({
        mutationFn: ({ file, options }: UploadSchemaParams) =>
            apiClient.uploadSchemaFile(file, options).then((res) => res.data),
    });
};
