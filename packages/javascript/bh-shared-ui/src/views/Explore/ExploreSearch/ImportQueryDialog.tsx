import { useEffect, useState } from 'react';

import {
    Button,
    Dialog,
    DialogActions,
    DialogClose,
    DialogContent,
    DialogPortal,
    DialogTitle,
} from '@bloodhoundenterprise/doodleui';
import FileDrop from '../../../components/FileDrop';

import FileStatusListItem from '../../../components/FileStatusListItem';

import { FileForIngest, FileStatus, FileUploadStep } from '../../../components/FileUploadDialog/types';

import { ErrorResponse } from 'js-client-library';
import { useImportSavedQuery } from '../../../hooks';
import { useNotifications } from '../../../providers';

const allowedFileTypes = ['application/json', 'application/zip'];

const ImportQueryDialog: React.FC<{
    open: boolean;
    onClose: () => void;
}> = ({ open, onClose }) => {
    const [filesForIngest, setFilesForIngest] = useState<FileForIngest[]>([]);
    const [fileUploadStep, setFileUploadStep] = useState<FileUploadStep>(FileUploadStep.ADD_FILES);
    const [submitDialogDisabled, setSubmitDialogDisabled] = useState<boolean>(false);
    const [uploadDialogDisabled, setUploadDialogDisabled] = useState<boolean>(false);
    const [uploadMessage, setUploadMessage] = useState<string>('');

    const { addNotification } = useNotifications();

    const importSavedQueryMutation = useImportSavedQuery();

    useEffect(() => {
        const filesHaveErrors = filesForIngest.filter((file) => file.errors).length > 0;
        const filesAreUploading = filesForIngest.filter((file) => file.status === FileStatus.UPLOADING).length > 0;
        setUploadDialogDisabled(filesAreUploading);
        const shouldDisableSubmit = filesHaveErrors || !filesForIngest.length;
        setSubmitDialogDisabled(shouldDisableSubmit);
    }, [filesForIngest]);

    const handleFileDrop = (files: FileList | null) => {
        if (files && files.length > 0) {
            const validatedFiles: FileForIngest[] = [...files].map((file) => {
                //Consider validating against userQueries as well.
                if (allowedFileTypes.includes(file.type)) {
                    return { file, status: FileStatus.READY };
                } else {
                    return { file, errors: ['invalid file type'], status: FileStatus.READY };
                }
            });
            handleAppendFiles(validatedFiles);
        }
    };

    const handleAppendFiles = (files: FileForIngest[]) => {
        setFilesForIngest((prevFiles) => [...prevFiles, ...files]);
    };

    const handleRemoveFile = (index: number) => {
        setFilesForIngest((prevFiles) => prevFiles.filter((_file, i) => i !== index));
    };

    const handleSubmit = () => {
        if (fileUploadStep === FileUploadStep.ADD_FILES) {
            setFileUploadStep(FileUploadStep.UPLOAD);
            handleUploadAll();
        }
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
    const uploadFile = async (ingestFile: FileForIngest) => {
        await importSavedQueryMutation.mutateAsync(ingestFile.file, {
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
        });
    };

    const handleUploadAll = async () => {
        updateStatusOfReadyFiles(FileStatus.UPLOADING);
        // const fileToUpload = filesForIngest[0];

        let errorCount = 0;

        for (const ingestFile of filesForIngest) {
            // Separate error handling so we can continue on when a file fails

            try {
                await uploadFile(ingestFile);
            } catch (error) {
                errorCount += 1;
            }

            setNewFileStatus(ingestFile.file.name, FileStatus.DONE);
        }

        if (errorCount === filesForIngest.length) {
            //all fail
            addNotification(`${errorCount} files have failed to upload.`, 'EndIngestFail');
        } else {
            addNotification(
                `Successfully uploaded ${filesForIngest.length - errorCount} files for ingest`,
                'FileIngestSuccess'
            );
        }
        const uploadMessage =
            errorCount > 0 ? 'Some files have failed to upload.' : 'All files have successfully been uploaded.';
        setUploadMessage(uploadMessage);
    };

    const handleClose = () => {
        setFileUploadStep(FileUploadStep.ADD_FILES);
        setFilesForIngest([]);
        onClose();
    };

    return (
        <Dialog open={open} onOpenChange={handleClose}>
            <DialogPortal>
                <DialogContent
                    DialogOverlayProps={{
                        blurBackground: false,
                    }}
                    maxWidth='sm'>
                    <DialogTitle>Upload Files</DialogTitle>

                    {fileUploadStep === FileUploadStep.ADD_FILES && (
                        <FileDrop onDrop={handleFileDrop} disabled={false} accept={allowedFileTypes} />
                    )}
                    {fileUploadStep === FileUploadStep.UPLOAD && uploadMessage && (
                        <div className='text-lg mb-4'>{uploadMessage}</div>
                    )}

                    {filesForIngest.length > 0 && (
                        <>
                            <div>Files</div>
                            {filesForIngest.map((file, index) => {
                                return (
                                    <FileStatusListItem
                                        file={file}
                                        key={index}
                                        onRemove={() => handleRemoveFile(index)}
                                    />
                                );
                            })}
                        </>
                    )}

                    <DialogActions className='flex justify-end gap-4'>
                        {fileUploadStep === FileUploadStep.ADD_FILES && (
                            <>
                                <DialogClose asChild>
                                    <Button variant='text'>Cancel</Button>
                                </DialogClose>
                                <Button variant='text' onClick={handleSubmit} disabled={submitDialogDisabled}>
                                    Upload
                                </Button>
                            </>
                        )}
                        {fileUploadStep === FileUploadStep.UPLOAD && (
                            <DialogClose asChild>
                                <Button variant='text' disabled={uploadDialogDisabled}>
                                    Complete
                                </Button>
                            </DialogClose>
                        )}
                    </DialogActions>
                </DialogContent>
            </DialogPortal>
        </Dialog>
    );
};

export default ImportQueryDialog;
