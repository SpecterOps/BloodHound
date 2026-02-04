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
import {
    Button,
    Card,
    Dialog,
    DialogActions,
    DialogClose,
    DialogContent,
    DialogDescription,
    DialogTitle,
    DialogTrigger,
} from '@bloodhoundenterprise/doodleui';
import { faCubes } from '@fortawesome/free-solid-svg-icons';
import { RequestOptions } from 'js-client-library';
import { useState } from 'react';
import { useMutation } from 'react-query';
import FileDrop from '../../components/FileDrop';
import FileStatusListItem from '../../components/FileStatusListItem';
import { FileForIngest, FileStatus } from '../../components/FileUploadDialog/types';
import { useExecuteOnFileDrag } from '../../hooks';
import { useNotifications } from '../../providers';
import { apiClient } from '../../utils';

export const SchemaUploadCard = () => {
    const [file, setFile] = useState<FileForIngest | null>(null);
    const [dialogOpen, setDialogOpen] = useState<boolean>(false);
    const [isUploading, setIsUploading] = useState<boolean>(false);
    const uploadSchemaFile = useUploadSchemaFile();
    const { addNotification } = useNotifications();

    useExecuteOnFileDrag(() => () => setDialogOpen(true), {
        acceptedTypes: ['application/json'],
    });

    const handleFileDrop = (files: FileList | null) => {
        if (files?.length === 1) {
            setFile({ file: files[0], status: FileStatus.READY });
        }
    };

    const handleUpload = () => {
        setIsUploading(true);

        return uploadSchemaFile.mutateAsync(
            { file: file?.file },
            {
                onError: () => {
                    addNotification(`Schema upload failed for ${file?.file.name}`, 'SchemaUploadFailure');
                    setFile((file) => (file ? { ...file, status: FileStatus.FAILURE } : null));
                },
            }
        );
    };

    const resetDialog = () => {
        setFile(null);
        setIsUploading(false);
    };

    const retryUpload = () => {
        alert('retrying');
    };

    return (
        <Card className='flex flex-col p-6 gap-4'>
            <h2 className='text-xl font-bold'>Custom Schema Upload</h2>
            <p>
                Upload custom schema JSON files to introduce new node and edge types. Then apply and validate schema
                updates to tailor the attack graph model to specific environments, workflows, or needs.
            </p>
            <Dialog
                open={dialogOpen}
                onOpenChange={(open) => {
                    if (open) {
                        resetDialog();
                    }
                    setDialogOpen(open);
                }}>
                <DialogTrigger asChild>
                    <Button className='self-start' variant='secondary'>
                        Upload File
                    </Button>
                </DialogTrigger>
                <DialogContent>
                    <DialogTitle>Upload Schema Files</DialogTitle>
                    <DialogDescription className='sr-only'>
                        An interface for uploading JSON OpenGraph schema files
                    </DialogDescription>
                    <FileDrop onDrop={handleFileDrop} disabled={false} multiple={false} icon={faCubes} />
                    <p className='text-xs text-center -mt-2 mb-4'>
                        Only single JSON file upload supported at this time
                    </p>
                    {file && (
                        <FileStatusListItem
                            file={file}
                            percentCompleted={0}
                            onRemove={() => setFile(null)}
                            onRefresh={retryUpload}
                        />
                    )}
                    <DialogActions>
                        <DialogClose asChild>
                            <Button variant='tertiary'>Close</Button>
                        </DialogClose>
                        <Button disabled={isUploading || !file} onClick={handleUpload}>
                            Upload
                        </Button>
                    </DialogActions>
                </DialogContent>
            </Dialog>
        </Card>
    );
};

interface UploadSchemaParams {
    file: any;
    options?: RequestOptions;
}

export const useUploadSchemaFile = () => {
    return useMutation({
        mutationFn: ({ file, options }: UploadSchemaParams) =>
            apiClient.uploadSchemaFile(file, options).then((res) => res.data),
    });
};
