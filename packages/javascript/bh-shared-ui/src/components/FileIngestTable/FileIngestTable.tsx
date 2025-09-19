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

import { Button, Card } from '@bloodhoundenterprise/doodleui';
import { FileIngestJob } from 'js-client-library';
import { FC, useState } from 'react';
import { useFileUploadDialogContext, useGetFileUploadsQuery, usePermissions } from '../../hooks';
import { JOB_STATUS_INDICATORS, JOB_STATUS_MAP, Permission, getSimpleDuration, toFormatted } from '../../utils';
import DataTable from '../DataTable';
import { StatusIndicator } from '../StatusIndicator';
import { FileIngestFilterDialog } from './FileIngestFilterDialog';

const HEADERS = ['ID / User / Status', 'Status Message', 'Start Time', 'Duration', 'File Information'];

const getHeaders = (headers: string[]) => headers.map((label) => ({ label, verticalAlign: 'baseline' }));

const getRow = (job: FileIngestJob) => {
    const [date, time, tz] = toFormatted(job.start_time).split(' ', 3);
    const indicator = JOB_STATUS_INDICATORS[job.status];
    const label = JOB_STATUS_MAP[job.status];

    return [
        <div className='min-w-32 space-y-2' key={`status-${job.id}`}>
            <div className='text-primary'>ID {job.id}</div>
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
    const [page, setPage] = useState(0);
    const [rowsPerPage, setRowsPerPage] = useState(10);
    const [filters, setFilters] = useState({});

    const { data, isLoading } = useGetFileUploadsQuery({ page, rowsPerPage, filters });

    const { setShowFileIngestDialog } = useFileUploadDialogContext();

    const { checkPermission } = usePermissions();
    const hasPermission = checkPermission(Permission.GRAPH_DB_INGEST);

    const fileUploadJobs = data?.data ?? [];
    const count = data?.count ?? 0;

    const toggleFileUploadDialog = () => setShowFileIngestDialog((prev) => !prev);

    const handleOnConfirm = (filters: any) => {
        setFilters(filters);
        setPage(0);
    };

    return (
        <>
            <div className='w-full flex justify-end gap-2 my-4'>
                <FileIngestFilterDialog onConfirm={handleOnConfirm} />
                <Button
                    onClick={() => toggleFileUploadDialog()}
                    data-testid='file-ingest_button-upload-files'
                    disabled={!hasPermission}>
                    Upload File(s)
                </Button>
            </div>

            <Card>
                <DataTable
                    data={fileUploadJobs.map(getRow)}
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
        </>
    );
};
