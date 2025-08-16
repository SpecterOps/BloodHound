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

import { ScheduledJobDisplay } from 'js-client-library';
import { FC, useState } from 'react';
import { useFinishedJobs } from '../../hooks';
import { JOB_STATUS_MAP, toCollected, toFormatted, toMins } from '../../utils';
import DataTable from '../DataTable';
import { StatusIndicator } from '../StatusIndicator';

const HEADERS = ['ID / Client / Status', 'Message', 'Start Time', 'Duration', 'Data Collected'];

const getHeaders = (headers: string[]) => headers.map((label) => ({ label, verticalAlign: 'baseline' }));

const getRow = (job: ScheduledJobDisplay) => {
    const [date, time, tz] = toFormatted(job.start_time).split(' ', 3);
    const statusProps = JOB_STATUS_MAP[job.status];

    return [
        <div className='min-w-32' key={`status-${job.id}`}>
            <div>ID {job.id}</div>
            <div>{job.client_name}</div>
            <div className='flex items-center'>{statusProps && <StatusIndicator {...statusProps} />}</div>
        </div>,
        <div className='max-w-40' key={`message-${job.id}`}>
            {job.status_message}
        </div>,
        <div className='min-w-20' key={`start-${job.id}`}>
            <div>{date}</div>
            <div>
                {time} {tz}
            </div>
        </div>,
        <div key={`duration-${job.id}`}>{toMins(job.start_time, job.end_time)}</div>,
        <div className='min-w-32 max-w-48' key={`collected-${job.id}`}>
            {toCollected(job)}
        </div>,
    ];
};

export const FinishedJobsTable: FC = () => {
    const [page, setPage] = useState(0);
    const [rowsPerPage, setRowsPerPage] = useState(10);

    const { data, isLoading } = useFinishedJobs({ page, rowsPerPage });

    const finishedJobs = data?.data ?? [];
    const count = data?.count ?? 0;

    return (
        <DataTable
            data={finishedJobs.map(getRow)}
            headers={getHeaders(HEADERS)}
            isLoading={isLoading}
            paginationProps={{
                page,
                rowsPerPage,
                count,
                onPageChange: (_event, page) => setPage(page),
                onRowsPerPageChange: (event) => setRowsPerPage(parseInt(event.target.value)),
            }}
            showPaginationControls
        />
    );
};
