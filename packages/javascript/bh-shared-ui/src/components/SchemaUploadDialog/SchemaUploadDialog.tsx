import {
    Button,
    Dialog,
    DialogActions,
    DialogClose,
    DialogContent,
    DialogDescription,
    DialogTitle,
    DialogTrigger,
} from '@bloodhoundenterprise/doodleui';
import { faCubes } from '@fortawesome/free-solid-svg-icons';
import { ReactNode } from 'react';
import { useExecuteOnFileDrag } from '../../hooks';
import FileDrop from '../FileDrop';
import FileStatusListItem from '../FileStatusListItem';
import { FileStatus } from '../FileUploadDialog';
import { useSchemaUploadHandlers } from './useSchemaUploadHandlers';

export const SchemaUploadDialog = ({ children }: { children?: ReactNode }) => {
    const { file, handleFileDrop, handleUpload, resetDialog, dialogOpen, setDialogOpen } = useSchemaUploadHandlers();

    useExecuteOnFileDrag(() => () => setDialogOpen(true), {
        acceptedTypes: ['application/json'],
    });

    return (
        <Dialog
            open={dialogOpen}
            onOpenChange={(open) => {
                if (open) resetDialog();
                setDialogOpen(open);
            }}>
            <DialogTrigger asChild>
                {children ?? (
                    <Button className='self-start' variant='secondary'>
                        Upload File
                    </Button>
                )}
            </DialogTrigger>
            <DialogContent>
                <DialogTitle>Upload Schema Files</DialogTitle>
                <DialogDescription className='sr-only'>
                    An interface for uploading JSON OpenGraph schema files
                </DialogDescription>
                <FileDrop
                    onDrop={handleFileDrop}
                    disabled={false}
                    multiple={false}
                    icon={faCubes}
                    accept={['application/json']}
                />
                <p className='text-xs text-center -mt-2 mb-4'>Only single JSON file upload supported at this time</p>
                {file && (
                    <FileStatusListItem
                        file={file}
                        percentCompleted={0}
                        onRemove={resetDialog}
                        onRefresh={handleUpload}
                    />
                )}
                <DialogActions>
                    <DialogClose asChild>
                        <Button variant='tertiary'>Cancel</Button>
                    </DialogClose>
                    {file?.status === FileStatus.FAILURE || file?.status === FileStatus.DONE ? (
                        <DialogClose asChild>
                            <Button>Complete</Button>
                        </DialogClose>
                    ) : (
                        <Button disabled={!file || file.status === FileStatus.UPLOADING} onClick={handleUpload}>
                            Upload
                        </Button>
                    )}
                </DialogActions>
            </DialogContent>
        </Dialog>
    );
};
