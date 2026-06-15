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

import { Dialog, DialogActions, DialogContent, DialogTitle } from '@mui/material';
import { Button } from 'doodle-ui';
import { ReactNode, useRef } from 'react';
import { useOnClickOutside, usePermissions } from '../../hooks';
import { Permission } from '../../utils';
import FileDrop from '../FileDrop';
import FileStatusListItem from '../FileStatusListItem';
import { AppLink } from '../Navigation';
import { FileUploadStep } from './types';
import { makeProgressCacheKey, useFileUploadDialogHandlers } from './useFileUploadDialogHandlers';

const FileUploadDialog: React.FC<{
    open: boolean;
    onClose: () => void;
    headerText?: ReactNode;
    description?: ReactNode;
}> = ({ open, onClose: onCloseProp, headerText = 'Upload Files', description }) => {
    const { checkPermission } = usePermissions();
    const hasPermissionToUpload = checkPermission(Permission.GRAPH_DB_INGEST);

    const {
        currentlyUploading,
        getFileUploadAcceptedTypes,
        progressCache,
        currentIngestJobId,
        filesForIngest,
        setFilesForIngest,
        setFileUploadStep,
        handleFileDrop,
        retryUploadSingleFile,
        uploadMessage,
        uploadDialogDisabled,
        submitDialogDisabled,
        handleSubmit,
        handleRemoveFile,
        onClose,
    } = useFileUploadDialogHandlers({ onCloseProp, hasPermissionToUpload });

    const dialogRef = useRef(null);

    useOnClickOutside(dialogRef, onClose);

    if (!hasPermissionToUpload) return null;

    return (
        <Dialog
            open={open}
            fullWidth={true}
            maxWidth={'sm'}
            scroll='paper'
            onClose={onClose}
            ref={dialogRef}
            className='z-[1600]'
            TransitionProps={{
                onExited: () => {
                    setFileUploadStep(FileUploadStep.ADD_FILES);
                    setFilesForIngest([]);
                },
            }}>
            <DialogTitle>
                <div className='pb-2 font-bold'>{headerText}</div>
                {description && <div>{description}</div>}

                <FileDrop
                    onDrop={handleFileDrop}
                    disabled={currentlyUploading || getFileUploadAcceptedTypes.isLoading}
                    accept={getFileUploadAcceptedTypes.data?.data ?? []}
                />
                {uploadMessage && <div className='mt-2 mb-2 font-normal'>{uploadMessage}</div>}
                <AppLink to='/administration/file-ingest' onClick={onClose}>
                    <div className='text-center font-normal my-2 p-2 hover:bg-neutral-5 rounded-md'>
                        View File Ingest History
                    </div>
                </AppLink>
            </DialogTitle>
            <DialogContent>
                {filesForIngest.length > 0 && (
                    <div className='my-2'>
                        {filesForIngest.map((file, index) => {
                            const key = makeProgressCacheKey(currentIngestJobId, file?.file?.name);
                            return (
                                <FileStatusListItem
                                    file={file}
                                    percentCompleted={progressCache[key] || 0}
                                    key={`${key}${index}`}
                                    onRemove={() => handleRemoveFile(index)}
                                    onRefresh={retryUploadSingleFile}
                                />
                            );
                        })}
                    </div>
                )}

                {currentlyUploading && (
                    <div>
                        <p>Upload in progress.</p>
                        <p>You can continue using the platform&mdash;we will alert you once the upload is complete.</p>
                    </div>
                )}
            </DialogContent>
            <DialogActions>
                <Button variant='tertiary' onClick={onClose} data-testid='confirmation-dialog_button-no'>
                    {uploadDialogDisabled ? 'Uploading Files' : 'Close'}
                </Button>
                <Button
                    disabled={submitDialogDisabled}
                    onClick={handleSubmit}
                    data-testid='confirmation-dialog_button-yes'>
                    Upload
                </Button>
            </DialogActions>
        </Dialog>
    );
};

export default FileUploadDialog;
