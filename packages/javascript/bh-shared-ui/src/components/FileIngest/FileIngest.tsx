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
import { FC, useState } from 'react';
import { useForbiddenNotifier } from '../../hooks';
import { Permission } from '../../utils';
import DocumentationLinks from '../DocumentationLinks';
import FileUploadDialog from '../FileUploadDialog';
import FinishedIngestLog from '../FinishedIngestLog';
import PageWithTitle from '../PageWithTitle';

const FileIngest: FC<{ permissions: Permission[] }> = ({ permissions }) => {
    const [fileUploadDialogOpen, setFileUploadDialogOpen] = useState<boolean>(false);

    const forbidden = useForbiddenNotifier(
        Permission.GRAPH_DB_WRITE,
        permissions,
        'Your user role does not grant permissions to upload data.',
        'file-upload-permission'
    );

    const toggleFileUploadDialog = () => setFileUploadDialogOpen((prev) => !prev);

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
                    disabled={forbidden}>
                    Upload File(s)
                </Button>
            </Box>
            <FinishedIngestLog forbidden={forbidden} />

            <FileUploadDialog open={fileUploadDialogOpen} onClose={toggleFileUploadDialog} />
        </>
    );
};

export default FileIngest;
