import { ColumnDef, DataTable, Pagination, Skeleton } from '@bloodhoundenterprise/doodleui';
import { ScheduledJobDisplay } from 'js-client-library';
import { FC, useState } from 'react';
import { StatusIndicator } from '../StatusIndicator';
import { toCollected, toFormatted, toMins, useFinishedJobsQuery } from './finishedJobs';

const COLUMNS_BASE: ColumnDef<ScheduledJobDisplay>[] = [
    {
        header: () => <span className='pl-4'>ID / Client / Status</span>,
        id: 'id',
        size: 160,
    },
    {
        header: 'Message',
        id: 'status',
        size: 170,
    },
    {
        header: 'Start Time',
        id: 'start',
        size: 100,
    },
    {
        header: 'Duration',
        id: 'duration',
        size: 75,
    },
    {
        header: 'Data Collected',
        id: 'collected',
        size: 210,
    },
];

const LOADING_CELLS = [
    () => (
        <>
            <Skeleton className='ml-4 mb-1 h-4' />
            <Skeleton className='ml-4 mb-1 h-4' />
            <Skeleton className='ml-4 h-4' />
        </>
    ),
    () => <Skeleton className='h-4' />,
    () => <Skeleton className='h-4' />,
    () => <Skeleton className='h-4' />,
    () => (
        <>
            <Skeleton className='h-4 mb-1' />
            <Skeleton className='h-4 mb-1' />
            <Skeleton className='h-4' />
        </>
    ),
];

const FINISHED_JOB_CELLS = [
    ({ row: { original: job } }) => (
        <div className='pl-4'>
            <div>ID {job.id}</div>
            <div>{job.client_name}</div>
            <div className='flex items-center'>
                <StatusIndicator status={job.status} type='job' />
            </div>
        </div>
    ),
    ({ row: { original: job } }) => job.status_message,
    ({ row: { original: job } }) => {
        const [date, time, tz] = toFormatted(job.start_time).split(' ', 3);
        return (
            <>
                <div>{date}</div>
                <div>
                    {time} {tz}
                </div>
            </>
        );
    },
    ({ row: { original: job } }) => toMins(job.start_time, job.end_time),
    ({ row: { original: job } }) => toCollected(job),
] as ColumnDef<ScheduledJobDisplay>['cell'][];

/** Return columns with either loading state or success state */
const getColumns = (isLoading: boolean) =>
    COLUMNS_BASE.map((item, index) => ({
        ...item,
        cell: isLoading ? LOADING_CELLS[index] : FINISHED_JOB_CELLS[index],
    }));

export const FinishedJobsLogTable: FC = () => {
    const [page, setPage] = useState(0);
    const [rowsPerPage, setRowsPerPage] = useState(10);

    const { data, isFetching } = useFinishedJobsQuery({ page, rowsPerPage });

    const finishedJobs = data?.data ?? [];
    const count = data?.count ?? 0;

    return (
        <>
            <DataTable
                columns={getColumns(isFetching)}
                data={finishedJobs}
                TableProps={{ className: 'table-fixed' }}
                TableCellProps={{ className: 'align-baseline max-w-[240px]' }}
            />

            {/* TO FIX: Why does the DataTable cover this up? */}
            <Pagination
                count={count}
                rowsPerPage={rowsPerPage}
                page={page}
                onPageChange={setPage}
                onRowsPerPageChange={setRowsPerPage}
            />
        </>
    );
};
