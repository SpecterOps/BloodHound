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

import { Button } from '@bloodhoundenterprise/doodleui';
import { Box, Typography } from '@mui/material';
import { FC } from 'react';
import { useExecuteOnFileDrag, useMountEffect, usePermissions } from '../../hooks';
import { useFileUploadDialogContext } from '../../hooks/useFileUploadDialogContext';
import { useNotifications } from '../../providers';
import { Permission } from '../../utils';
import DocumentationLinks from '../DocumentationLinks';
import FinishedIngestLog from '../FinishedIngestLog';
import PageWithTitle from '../PageWithTitle';

const FileIngest: FC = () => {
    const { setShowFileIngestDialog } = useFileUploadDialogContext();

    const { checkPermission } = usePermissions();
    const hasPermission = checkPermission(Permission.GRAPH_DB_WRITE);

    const { addNotification, dismissNotification } = useNotifications();
    const notificationKey = 'file-upload-permission';

    const effect: React.EffectCallback = () => {
        if (!hasPermission) {
            addNotification(
                `Your user role does not grant permission to upload data. Please contact your administrator for details.`,
                notificationKey,
                {
                    persist: true,
                    anchorOrigin: { vertical: 'top', horizontal: 'right' },
                }
            );
        }

        return () => dismissNotification(notificationKey);
    };

    useMountEffect(effect);

    // Open the file upload dialog when a processable file is dragged into the browser client
    useExecuteOnFileDrag(() => setShowFileIngestDialog(true), {
        acceptedTypes: ['application/json', 'application/zip'],
    });

    const toggleFileUploadDialog = () => setShowFileIngestDialog((prev) => !prev);

    return (
        <>
            <PageWithTitle
                title='File Ingest'
                data-testid='manual-file-ingest'
                pageDescription={
                    <Typography variant='body2'>
                        Upload data from SharpHound or AzureHound offline collectors. Check out our{' '}
                        {DocumentationLinks.fileIngestLink} documentation for more information.
                    </Typography>
                }></PageWithTitle>

            <Box display='flex' justifyContent='flex-end' alignItems='center' minHeight='24px' my={2}>
                <Button
                    onClick={() => toggleFileUploadDialog()}
                    data-testid='file-ingest_button-upload-files'
                    disabled={!hasPermission}>
                    Upload File(s)
                </Button>
            </Box>
            <FinishedIngestLog />
        </>
    );
};

export default FileIngest;
