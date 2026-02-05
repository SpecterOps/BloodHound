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

    /**
     * Validates that only one file has been dropped, and if so adds it to the dialog in the "Ready" status
     */
    const handleFileDrop = (files: FileList | null) => {
        if (!files || files.length === 0) return;

        if (files?.length === 1) {
            setFile({ file: files[0], status: FileStatus.READY });
        } else {
            addNotification('Currently only supports single file uploads', 'MultipleFileError');
        }
    };

    /**
     * Clears out any selected file and resets to default state
     */
    const resetDialog = () => setFile(null);

    /**
     * Attempts to upload the current file and manages the file's status through the upload process
     */
    const handleUpload = () => {
        if (!file) return;

        setNewFileStatus(FileStatus.UPLOADING);

        return uploadSchemaFile.mutateAsync(
            {
                file: file.file,
                options: {
                    headers: {
                        'Content-Type': file.file.type,
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
                },
                onSuccess: () => setNewFileStatus(FileStatus.DONE),
            }
        );
    };

    const handleUploadError = (err: unknown) => {
        if (err && isAxiosError(err)) {
            addNotification(err.response?.data?.errors?.[0]?.message, 'SchemaUploadFailure');
        } else {
            addNotification(`An unknown error occurred: ${err}`, 'SchemaUploadFailure');
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
    file?: File;
    options?: RequestOptions;
}

const useUploadSchemaFile = () => {
    return useMutation({
        mutationFn: ({ file, options }: UploadSchemaParams) =>
            apiClient.uploadSchemaFile(file, options).then((res) => res.data),
    });
};
