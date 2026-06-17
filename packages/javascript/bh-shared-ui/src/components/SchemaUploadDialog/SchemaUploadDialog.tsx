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
import { faCubes } from '@fortawesome/free-solid-svg-icons';
import {
    Button,
    Dialog,
    DialogActions,
    DialogClose,
    DialogContent,
    DialogDescription,
    DialogTitle,
    DialogTrigger,
} from 'doodle-ui';
import { useCallback, useState } from 'react';
import { useExecuteOnFileDrag, useFileUploadDialogContext, usePermissions } from '../../hooks';
import { Permission } from '../../utils';
import { ConditionalTooltip } from '../ConditionalTooltip';
import FileDrop from '../FileDrop';
import FileStatusListItem from '../FileStatusListItem';
import { FileStatus } from '../FileUploadDialog';
import { useSchemaUploadHandlers } from './useSchemaUploadHandlers';

export const SchemaUploadDialog = () => {
    const [dialogOpen, setDialogOpen] = useState<boolean>(false);
    const { file, uploadProgress, handleFileDrop, handleUpload, resetDialog } = useSchemaUploadHandlers();

    const { checkPermission } = usePermissions();
    const hasUploadPermission = checkPermission(Permission.OPENGRAPH_WRITE);
    const { showFileIngestDialog } = useFileUploadDialogContext();

    const shouldRespondToDrag = useCallback(
        () => hasUploadPermission && !showFileIngestDialog,
        [hasUploadPermission, showFileIngestDialog]
    );

    useExecuteOnFileDrag(
        () => {
            if (hasUploadPermission) setDialogOpen(true);
        },
        {
            acceptedTypes: ['application/json'],
            condition: shouldRespondToDrag,
        }
    );

    return (
        <Dialog
            open={dialogOpen}
            onOpenChange={(open) => {
                if (!hasUploadPermission) return;
                if (open) resetDialog();
                setDialogOpen(open);
            }}>
            <ConditionalTooltip condition={!hasUploadPermission} tooltip='You do not have permission to upload files.'>
                <DialogTrigger asChild>
                    <span className='self-start' tabIndex={!hasUploadPermission ? 0 : undefined}>
                        <Button variant='secondary' disabled={!hasUploadPermission}>
                            Upload File
                        </Button>
                    </span>
                </DialogTrigger>
            </ConditionalTooltip>
            <DialogContent>
                <DialogTitle>Upload Schema Files</DialogTitle>
                <DialogDescription className='sr-only'>
                    An interface for uploading JSON OpenGraph schema files
                </DialogDescription>
                <FileDrop
                    onDrop={handleFileDrop}
                    disabled={!!file || !hasUploadPermission}
                    multiple={false}
                    icon={faCubes}
                    accept={['application/json']}
                />
                <p className='text-xs text-center -mt-2 mb-4'>Only single JSON file upload supported at this time</p>
                {file && (
                    <FileStatusListItem
                        file={file}
                        percentCompleted={uploadProgress}
                        onRemove={resetDialog}
                        onRefresh={handleUpload}
                    />
                )}
                <DialogActions>
                    {file?.status === FileStatus.FAILURE || file?.status === FileStatus.DONE ? (
                        <>
                            <DialogClose asChild>
                                <Button variant='tertiary'>Close</Button>
                            </DialogClose>
                            <DialogClose asChild>
                                <Button>Complete</Button>
                            </DialogClose>
                        </>
                    ) : (
                        <>
                            <DialogClose asChild>
                                <Button variant='tertiary'>Cancel</Button>
                            </DialogClose>
                            <Button disabled={!file || file.status === FileStatus.UPLOADING} onClick={handleUpload}>
                                Upload
                            </Button>
                        </>
                    )}
                </DialogActions>
            </DialogContent>
        </Dialog>
    );
};
