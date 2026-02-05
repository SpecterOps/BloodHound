import { RequestOptions } from 'js-client-library';
import { useState } from 'react';
import { useMutation } from 'react-query';
import { useNotifications } from '../../providers';
import { apiClient } from '../../utils';
import { FileForIngest, FileStatus } from '../FileUploadDialog';

export const useSchemaUploadHandlers = () => {
    const [file, setFile] = useState<FileForIngest | null>(null);
    const [dialogOpen, setDialogOpen] = useState<boolean>(false);
    const uploadSchemaFile = useUploadSchemaFile();
    const { addNotification } = useNotifications();

    const handleFileDrop = (files: FileList | null) => {
        if (files?.length === 1) {
            setFile({ file: files[0], status: FileStatus.READY });
        } else {
            addNotification('Currently only supports single file uploads', 'MultipleFileError');
        }
    };

    const handleUpload = () => {
        setNewFileStatus(FileStatus.UPLOADING);
        return uploadSchemaFile.mutateAsync(
            { file: file?.file },
            {
                onError: () => {
                    addNotification(`Schema upload failed for ${file?.file.name}`, 'SchemaUploadFailure');
                    setNewFileStatus(FileStatus.FAILURE);
                },
                onSuccess: () => {
                    setNewFileStatus(FileStatus.DONE);
                },
            }
        );
    };

    const setNewFileStatus = (status: FileStatus) => {
        setFile((file) => (file ? { ...file, status } : null));
    };

    const resetDialog = () => setFile(null);

    return {
        file,
        handleFileDrop,
        handleUpload,
        resetDialog,
        dialogOpen,
        setDialogOpen,
    };
};

interface UploadSchemaParams {
    file?: File;
    options?: RequestOptions;
}

export const useUploadSchemaFile = () => {
    return useMutation({
        mutationFn: ({ file, options }: UploadSchemaParams) =>
            apiClient.uploadSchemaFile(file, options).then((res) => res.data),
    });
};
