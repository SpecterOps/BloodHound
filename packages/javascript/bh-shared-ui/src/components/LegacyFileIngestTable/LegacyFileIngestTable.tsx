// Copyright 2025 Specter Ops, Inc.
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

import { Paper } from '@mui/material';
import { DateTime } from 'luxon';
import { useEffect, useState } from 'react';
import { ZERO_VALUE_API_DATE } from '../../constants';
import { useGetFileUploadsQuery } from '../../hooks';
import { LuxonFormat, getSimpleDuration } from '../../utils/datetime';
import DataTable from '../DataTable';
import { UploadFilesDialog } from '../FileIngest/UploadFilesDialog';
import { FileUploadJob, FileUploadJobStatusToString } from './types';

const ingestTableHeaders = [
    { label: 'User' },
    { label: 'Start Time' },
    { label: 'End Time' },
    { label: 'Duration' },
    { label: 'Status' },
    { label: 'Status Message' },
];

const LegacyFileIngestTable: React.FC = () => {
    const [page, setPage] = useState(0);
    const [rowsPerPage, setRowsPerPage] = useState(10);
    const [totalCount, setTotalCount] = useState(0);

    const { data: listFileIngestJobsData } = useGetFileUploadsQuery({ page, rowsPerPage });

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

    const ingestRows =
        listFileIngestJobsData?.data
            ?.sort((a, b) => b.id - a.id)
            .map((job: FileUploadJob) => [
                job.user_email_address,
                DateTime.fromISO(job.start_time).toFormat(LuxonFormat.DATETIME_WITH_LINEBREAKS),
                job.end_time === ZERO_VALUE_API_DATE
                    ? ''
                    : DateTime.fromISO(job.end_time).toFormat(LuxonFormat.DATETIME_WITH_LINEBREAKS),
                job.end_time === ZERO_VALUE_API_DATE ? '' : getSimpleDuration(job.start_time, job.end_time),
                FileUploadJobStatusToString[job.status],
                job.status_message,
            ]) || [];

    return (
        <>
            <div className='my-4 text-right'>
                <UploadFilesDialog />
            </div>

            <Paper>
                <DataTable
                    headers={ingestTableHeaders}
                    data={ingestRows}
                    showPaginationControls={true}
                    paginationProps={{
                        page,
                        rowsPerPage,
                        count: totalCount,
                        onPageChange: handlePageChange,
                        onRowsPerPageChange: handleRowsPerPageChange,
                    }}
                />
            </Paper>
        </>
    );
};

export default LegacyFileIngestTable;
