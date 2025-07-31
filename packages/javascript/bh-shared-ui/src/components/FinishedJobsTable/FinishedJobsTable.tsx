import { ScheduledJobDisplay } from 'js-client-library';
import { FC, useState } from 'react';
import DataTable from '../DataTable';
import { StatusIndicator } from '../StatusIndicator';
import { toCollected, toFormatted, toMins, useFinishedJobsQuery } from './finishedJobs';

const HEADERS = ['ID / Client / Status', 'Message', 'Start Time', 'Duration', 'Data Collected'];

const getHeaders = (headers: string[]) => headers.map((label) => ({ label, verticalAlign: 'baseline' }));

const getRows = (job: ScheduledJobDisplay) => {
    const [date, time, tz] = toFormatted(job.start_time).split(' ', 3);
    return [
        <div className='w-[120px]' key={`status-${job.id}`}>
            <div>ID {job.id}</div>
            <div>{job.client_name}</div>
            <div className='flex items-center'>
                <StatusIndicator status={job.status} type='job' />
            </div>
        </div>,
        <div className='w-[180px]' key={`message-${job.id}`}>
            {job.status_message}
        </div>,
        <div className='w-[80px]' key={`start-${job.id}`}>
            <div>{date}</div>
            <div>
                {time} {tz}
            </div>
        </div>,
        <div className='w-[55px]' key={`duration-${job.id}`}>
            {toMins(job.start_time, job.end_time)}
        </div>,
        <div className='w-[200px]' key={`collected-${job.id}`}>
            {toCollected(job)}
        </div>,
    ];
};

export const FinishedJobsLogTable: FC = () => {
    const [page, setPage] = useState(0);
    const [rowsPerPage, setRowsPerPage] = useState(10);

    const { data, isLoading } = useFinishedJobsQuery({ page, rowsPerPage });

    const finishedJobs = data?.data ?? [];
    const count = data?.count ?? 0;

    return (
        <DataTable
            data={finishedJobs.map(getRows)}
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
