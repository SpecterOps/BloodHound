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

import { Box, Button, Link, Typography } from '@mui/material';
import { useEffect, useState } from 'react';
import FileUploadDialog from '../FileUploadDialog';
import { useListFileIngestJobs } from '../../hooks';
import ContentPage from '../ContentPage';
import FinishedIngestLog from '../FinishedIngestLog';

const fileIngestLink = (
    <Link
        target='_blank'
        data-testid='file-upload-gettingstarted-link'
        href={
            'https://support.bloodhoundenterprise.io/hc/en-us/articles/17715215791899-Getting-started-with-BloodHound-Community-Edition'
        }>
        Getting Started
    </Link>
);

const FileIngest = () => {
    const [fileUploadDialogOpen, setFileUploadDialogOpen] = useState<boolean>(false);
    const [page, setPage] = useState(0);
    const [rowsPerPage, setRowsPerPage] = useState(10);
    const [totalCount, setTotalCount] = useState(0);

    const { data: listFileIngestJobsData } = useListFileIngestJobs(page, rowsPerPage);

    useEffect(() => setTotalCount(listFileIngestJobsData?.count || 0), [listFileIngestJobsData]);

    const handlePageChange: (event: React.MouseEvent<HTMLButtonElement> | null, page: number) => void = (
        _event,
        newPage
    ) => {
        setPage(newPage);
    };

    const handleRowsPerPageChange: React.ChangeEventHandler<HTMLTextAreaElement | HTMLInputElement> = (event) => {
        setRowsPerPage(parseInt(event.target.value, 10));
        setPage(0);
    };

    const toggleFileUploadDialog = () => setFileUploadDialogOpen((prev) => !prev);

    return (
        <>
            <ContentPage title='Manual File Ingest' data-testid='manual-file-ingest'>
                <Typography variant='body1'>
                    <p>
                        Upload data from SharpHound or AzureHound offline collectors. Check out our {fileIngestLink}{' '}
                        documentation for more information.
                    </p>
                </Typography>
                <Box display='flex' justifyContent='flex-end' alignItems='center' minHeight='24px' mb={2}>
                    <Button
                        color='primary'
                        variant='contained'
                        disableElevation
                        onClick={() => toggleFileUploadDialog()}
                        data-testid='file-ingest_button-upload-files'>
                        Upload File(s)
                    </Button>
                </Box>
            </ContentPage>

            <ContentPage title='Finished Ingest Log' data-testid='finished-ingest-log'>
                <FinishedIngestLog
                    ingestJobs={listFileIngestJobsData?.data || []}
                    paginationProps={{
                        page,
                        rowsPerPage,
                        count: totalCount,
                        onPageChange: handlePageChange,
                        onRowsPerPageChange: handleRowsPerPageChange,
                    }}
                />
            </ContentPage>

            <FileUploadDialog open={fileUploadDialogOpen} onClose={toggleFileUploadDialog} />
        </>
    );
};

export default FileIngest;
