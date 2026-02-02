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
    Dialog,
    DialogActions,
    DialogClose,
    DialogContent,
    DialogDescription,
    DialogTitle,
    DialogTrigger,
} from '@bloodhoundenterprise/doodleui';
import { useState } from 'react';
import FileDrop from '../../components/FileDrop';
import FileStatusListItem from '../../components/FileStatusListItem';
import { FileForIngest, FileStatus } from '../../components/FileUploadDialog/types';
import { useExecuteOnFileDrag } from '../../hooks';

export const SchemaUploadCard = () => {
    const [file, setFile] = useState<FileForIngest | null>(null);
    const [dialogOpen, setDialogOpen] = useState<boolean>(false);

    useExecuteOnFileDrag(() => () => setDialogOpen(true), {
        acceptedTypes: ['application/json'],
    });

    const handleFileDrop = (files: FileList | null) => {
        if (files?.length === 1) {
            setFile({ file: files[0], status: FileStatus.READY });
        }
    };

    const retryUpload = () => {
        alert('retrying');
    };

    return (
        <div className='p-6 pr-8 rounded-lg bg-neutral-2 mt-4 max-w-4xl'>
            <h2 className='font-bold text-lg mb-5'>Custom Schema Upload</h2>
            <p className='mb-4'>
                Upload custom schema JSON files to introduce new node and edge types. Then apply and validate schema
                updates to tailor the attack graph model to specific environments, workflows, or needs.
            </p>
            <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
                <DialogTrigger asChild>
                    <Button variant='secondary'>Upload File</Button>
                </DialogTrigger>
                <DialogContent>
                    <DialogTitle>Upload Schema Files</DialogTitle>
                    <DialogDescription className='sr-only'>
                        An interface for uploading JSON OpenGraph schema files
                    </DialogDescription>
                    <FileDrop onDrop={handleFileDrop} disabled={false} multiple={false} />
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
                        <Button>Upload</Button>
                    </DialogActions>
                </DialogContent>
            </Dialog>
        </div>
    );
};
