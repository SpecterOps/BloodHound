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

import { FileUploadJob, FileUploadJobStatusToString } from './types';
import { LuxonFormat, calculateJobDuration } from '../../utils/datetime';
import { DateTime } from 'luxon';
import DataTable from '../DataTable';
import { Paper } from '@mui/material';

const ZERO_VALUE_API_DATE = '0001-01-01T00:00:00Z';

const ingestTableHeaders = [
    { label: 'User' },
    { label: 'Start Time' },
    { label: 'End Time' },
    { label: 'Duration' },
    { label: 'Status' },
    { label: 'Status Message' },
];

const FinishedIngestLog: React.FC<{
    ingestJobs: FileUploadJob[];
    paginationProps?: {
        page: number;
        rowsPerPage: number;
        count: number;
        onPageChange: (event: React.MouseEvent<HTMLButtonElement, MouseEvent> | null, page: number) => void;
        onRowsPerPageChange: React.ChangeEventHandler<HTMLInputElement | HTMLTextAreaElement>;
    };
}> = ({ ingestJobs, paginationProps }) => {
    const ingestRows = ingestJobs
        .sort((a, b) => b.id - a.id)
        .map((job: FileUploadJob, index: number) => [
            job.user_email_address,
            DateTime.fromISO(job.start_time).toFormat(LuxonFormat.DATETIME_WITH_LINEBREAKS),
            job.end_time === ZERO_VALUE_API_DATE
                ? ''
                : DateTime.fromISO(job.end_time).toFormat(LuxonFormat.DATETIME_WITH_LINEBREAKS),
            job.end_time === ZERO_VALUE_API_DATE ? '' : calculateJobDuration(job.start_time, job.end_time),
            FileUploadJobStatusToString[job.status],
            job.status_message,
        ]);

    return (
        <Paper>
            <DataTable
                headers={ingestTableHeaders}
                data={ingestRows}
                showPaginationControls={true}
                paginationProps={paginationProps}
            />
        </Paper>
    );
};

export default FinishedIngestLog;
