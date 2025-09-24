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

import { Card } from '@bloodhoundenterprise/doodleui';
import type { FileIngestJob } from 'js-client-library';
import { FC, useState } from 'react';
import { useGetFileUploadsQuery } from '../../hooks';
import { JOB_STATUS_INDICATORS, JOB_STATUS_MAP, getSimpleDuration, toFormatted } from '../../utils';
import DataTable from '../DataTable';
import { FileIngestUploadButton } from '../FileIngest/FileIngestUploadButton';
import { StatusIndicator } from '../StatusIndicator';
import { FileIngestDetailsPanel } from './FileIngestDetailsPanel';

const HEADERS = ['ID / User / Status', 'Message', 'Start Time', 'Duration', 'File Information'];

const getHeaders = (headers: string[]) => headers.map((label) => ({ label, verticalAlign: 'baseline' }));

const getRow =
    (onSelectJob: React.Dispatch<React.SetStateAction<FileIngestJob | undefined>>) => (job: FileIngestJob) => {
        const [date, time, tz] = toFormatted(job.start_time).split(' ', 3);
        const indicator = JOB_STATUS_INDICATORS[job.status];
        const label = JOB_STATUS_MAP[job.status];

        return [
            <div className='min-w-32 space-y-1' key={`status-${job.id}`}>
                <button
                    type='button'
                    className='text-secondary dark:text-secondary-variant-2 hover:underline'
                    onClick={() => onSelectJob(job)}
                    aria-label={`View ingest ${job.id} details`}>
                    ID {job.id}
                </button>
                <div>{job.user_email_address}</div>
                <div className='flex items-center'>
                    <StatusIndicator {...indicator} label={label} />
                </div>
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
            <div key={`duration-${job.id}`}>{getSimpleDuration(job.start_time, job.end_time)}</div>,
            <div className='min-w-32 max-w-48' key={`collected-${job.id}`}>
                {job.total_files} Files
            </div>,
        ];
    };

export const FileIngestTable: FC = () => {
    const [selectedIngest, setSelectedIngest] = useState<FileIngestJob>();
    const [page, setPage] = useState(0);
    const [rowsPerPage, setRowsPerPage] = useState(10);

    const { data, isLoading } = useGetFileUploadsQuery({ page, rowsPerPage });

    const fileUploadJobs = data?.data ?? [];
    const count = data?.count ?? 0;

    const getRowWithSelect = getRow(setSelectedIngest);

    return (
        <div className='grid h-full grid-cols-[1fr_27rem] grid-rows-[auto_minmax(0,1fr)] pt-4 gap-4'>
            <div className='col-[1] row-[1] flex items-center justify-end gap-2'>
                <FileIngestUploadButton />
            </div>

            <div className='col-[1] row-[2] min-h-0'>
                <Card>
                    <DataTable
                        data={fileUploadJobs.map(getRowWithSelect)}
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
                </Card>
            </div>

            <div className='col-[2] row-[2]'>
                <FileIngestDetailsPanel ingest={selectedIngest} />
            </div>
        </div>
    );
};
