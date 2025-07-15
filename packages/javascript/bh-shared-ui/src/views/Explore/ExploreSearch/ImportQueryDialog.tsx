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

import { useImportSavedQuery } from '../../../hooks';

const allowedFileTypes = ['application/json', 'application/zip'];

const ImportQueryDialog: React.FC<{
    open: boolean;
    onClose: () => void;
}> = ({ open, onClose }) => {
    const [filesForIngest, setFilesForIngest] = useState<FileForIngest[]>([]);
    const [fileUploadStep, setFileUploadStep] = useState<FileUploadStep>(FileUploadStep.ADD_FILES);

    const importSavedQueryMutation = useImportSavedQuery();

    useEffect(() => {
        const filesHaveErrors = filesForIngest.filter((file) => file.errors).length > 0;
        const filesAreUploading = filesForIngest.filter((file) => file.status === FileStatus.UPLOADING).length > 0;

        const shouldDisableSubmit = filesHaveErrors || !filesForIngest.length;
    }, [filesForIngest]);

    const handleFileDrop = (files: FileList | null) => {
        console.log(files);

        if (files && files.length > 0) {
            const validatedFiles: FileForIngest[] = [...files].map((file) => {
                if (allowedFileTypes.includes(file.type)) {
                    return { file, status: FileStatus.READY };
                } else {
                    return { file, errors: ['invalid file type'], status: FileStatus.READY };
                }
            });
            console.log(validatedFiles);
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
            handleUpload();
        }
    };

    const handleUpload = () => {
        // console.log('handleUpload');
        const fileToUpload = filesForIngest[0];
        console.log(fileToUpload.file);
        // console.log(fileToUpload.file.type);
        // const formData = new FormData();
        // formData.append('payload', fileToUpload);

        // const blob = new Blob([fileToUpload.file], { type: fileToUpload.file.type });
        // console.log(blob);
        const fileType = fileToUpload.file.type;
        // const formData = new FormData();
        // formData.append('payload', fileToUpload.file);
        // formData.append('contentType', fileType);

        // importSavedQueryMutation.mutate(formData);
    };

    const handleFileChange = (event: any) => {
        const selectedFile = event.target.files[0];
        // Get the first selected file
        // Perform actions with the selected file, e.g., display its name, read its content, or prepare for upload.
        console.log(selectedFile);
    };

    return (
        <Dialog open={open} onOpenChange={onClose}>
            <DialogPortal>
                <DialogContent
                    DialogOverlayProps={{
                        blurBackground: false,
                    }}
                    maxWidth='sm'>
                    <DialogTitle>Upload Files</DialogTitle>

                    <input type='file' onChange={handleFileChange} />

                    <FileDrop
                        onDrop={handleFileDrop}
                        // disabled={listFileTypesForIngest.isLoading}
                        disabled={false}
                        accept={allowedFileTypes}
                    />
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
                                <Button variant='text' onClick={handleSubmit}>
                                    Upload
                                </Button>
                            </>
                        )}

                        {fileUploadStep === FileUploadStep.UPLOAD && (
                            <DialogClose asChild>
                                <Button variant='text'>Uploading</Button>
                            </DialogClose>
                        )}
                    </DialogActions>
                </DialogContent>
            </DialogPortal>
        </Dialog>
    );
};

export default ImportQueryDialog;
