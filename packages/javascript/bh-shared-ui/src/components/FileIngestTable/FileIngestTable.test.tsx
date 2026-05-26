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

import userEvent from '@testing-library/user-event';
import type { FileIngestCompletedTasksResponse, FileIngestJob } from 'js-client-library';
import { rest } from 'msw';
import { setupServer } from 'msw/node';
import * as useFileUpload from '../../hooks/useFileUploadQuery/useFileUploadQuery';
import { act, render, screen, waitFor } from '../../test-utils';
import { FileIngestTable } from './FileIngestTable';

const checkPermissionMock = vi.fn();

vi.mock('../../hooks/usePermissions', async () => {
    const actual = await vi.importActual('../../hooks');
    return {
        ...actual,
        usePermissions: () => ({
            checkPermission: checkPermissionMock,
            isSuccess: true,
        }),
    };
});

const MOCK_INGEST_JOB: FileIngestJob = {
    user_id: '1234',
    user_email_address: 'spam@example.com',
    status: 2,
    status_message: 'Completed',
    start_time: '2024-08-15T21:25:21.990437Z',
    end_time: '2024-08-15T21:26:43.033448Z',
    last_ingest: '2024-08-15T21:26:43.033448Z',
    id: 1,
    total_files: 10,
    failed_files: 0,
    created_at: '',
    updated_at: '',
    deleted_at: {
        Time: '',
        Valid: false,
    },
};

const MOCK_PARTIAL_SUCCESS_INGEST_JOB: FileIngestJob = {
    user_id: '1234',
    user_email_address: 'spam@example.com',
    status: 8,
    status_message: 'Partially Completed',
    start_time: '2024-08-15T21:25:21.990437Z',
    end_time: '2024-08-15T21:26:43.033448Z',
    last_ingest: '2024-08-15T21:26:43.033448Z',
    id: 9,
    total_files: 10,
    failed_files: 0,
    created_at: '',
    updated_at: '',
    deleted_at: {
        Time: '',
        Valid: false,
    },
};

const MOCK_INGEST_JOBS_RESPONSE = {
    count: 20,
    // fill the array with data
    data: Array.from({ length: 10 }, (_, index) => {
        if (index % 2 === 0) {
            return MOCK_INGEST_JOB;
        } else return MOCK_PARTIAL_SUCCESS_INGEST_JOB;
    }).map((item, index) => ({
        ...item,
        id: index,
        status: (index % 10) - 1,
    })),
    limit: 10,
    skip: 10,
};

const MOCK_COMPLETED_TASKS_RESPONSE: FileIngestCompletedTasksResponse = {
    data: [
        {
            file_name: 'generic-with-failed-edges.json',
            parent_file_name: '',
            errors: [],
            warnings: [
                'skipping invalid relationship. unable to resolve endpoints. source: NON2@EXISTING.NODE, target: NON1@EXISTING.NODE',
            ],
            id: 9,
            created_at: '2026-01-14T00:17:40.255611Z',
            updated_at: '2026-01-14T00:17:40.255611Z',
            deleted_at: {
                Time: '0001-01-01T00:00:00Z',
                Valid: false,
            },
        },
    ],
};
const server = setupServer(
    rest.get('/api/v2/file-upload', (req, res, ctx) => res(ctx.json(MOCK_INGEST_JOBS_RESPONSE))),
    rest.get('/api/v2/features', (req, res, ctx) => {
        return res(
            ctx.json({
                data: [
                    {
                        key: 'open_graph_phase_2',
                        enabled: true,
                    },
                ],
            })
        );
    })
);

beforeAll(() => {
    server.listen();
});
afterEach(() => server.resetHandlers());
afterAll(() => {
    server.close();
    vi.clearAllMocks();
    server.resetHandlers();
});

const useFileUploadQuerySpy = vi.spyOn(useFileUpload, 'useFileUploadQuery');
useFileUploadQuerySpy.mockReturnValue({ data: MOCK_COMPLETED_TASKS_RESPONSE, isSuccess: true } as any);

describe('FileIngestTable', () => {
    it('shows a loading state', () => {
        checkPermissionMock.mockImplementation(() => true);
        const { container } = render(<FileIngestTable />);

        // 1 loading skeleton for each column
        const EXPECTED_COLUMN_COUNT = 5;
        const children = container.querySelectorAll('.MuiSkeleton-pulse');
        expect(children.length).toBe(EXPECTED_COLUMN_COUNT);
    });

    it('shows a table with finished jobs', async () => {
        checkPermissionMock.mockImplementation(() => true);
        await act(async () => render(<FileIngestTable />));

        const jobStatus = await screen.findByText('Complete');
        expect(jobStatus).toHaveTextContent('Complete');
    });
    it('shows a table with partially completed jobs', async () => {
        checkPermissionMock.mockImplementation(() => true);
        await act(async () => render(<FileIngestTable />));
        const user = userEvent.setup();

        const jobID = await screen.getByRole('button', { name: 'View ingest 9 details' });
        await user.click(jobID);

        const jobsDropdown = await screen.getAllByTestId('0');
        expect(jobsDropdown[0]).toBeInTheDocument();

        await user.click(jobsDropdown[0]);

        const partiallyCompletedJob = await screen.findByText('generic-with-failed-edges.json');
        expect(partiallyCompletedJob).toBeInTheDocument();

        await user.click(partiallyCompletedJob);

        const warningText = await screen.findByText(
            'skipping invalid relationship. unable to resolve endpoints. source: NON2@EXISTING.NODE, target: NON1@EXISTING.NODE'
        );
        expect(warningText).toBeInTheDocument();
    });

    describe('pagination reset', () => {
        // 5 jobs match the Complete (status=2) filter; 10 do not. Total 15 — chosen so that:
        //   - On rowsPerPage=10, skip=10 has data (5 rows), but skip=10 on the filtered pool of 5 would be empty.
        //   - On rowsPerPage=25, skip=25 on the unfiltered pool of 15 would be empty.
        const COMPLETE_STATUS = 2;
        const OTHER_STATUS = 3;
        const ALL_JOBS: FileIngestJob[] = new Array(15).fill(null).map((_, index) => ({
            ...MOCK_INGEST_JOB,
            id: index,
            status: index < 5 ? COMPLETE_STATUS : OTHER_STATUS,
        }));

        const fileUploadRequests: URL[] = [];
        const lastFileUploadRequest = () => fileUploadRequests[fileUploadRequests.length - 1];
        const visibleIngestRows = () => screen.queryAllByRole('button', { name: /view ingest .* details/i });

        beforeEach(() => {
            checkPermissionMock.mockImplementation(() => true);
            fileUploadRequests.length = 0;
            server.use(
                rest.get('/api/v2/file-upload', (req, res, ctx) => {
                    fileUploadRequests.push(req.url);
                    const skip = parseInt(req.url.searchParams.get('skip') ?? '0', 10);
                    const limit = parseInt(req.url.searchParams.get('limit') ?? '10', 10);
                    const statusParam = req.url.searchParams.get('status');
                    let pool = ALL_JOBS;
                    const statusEq = statusParam?.match(/^eq:(\d+)$/);
                    if (statusEq) {
                        const wanted = parseInt(statusEq[1], 10);
                        pool = pool.filter((job) => job.status === wanted);
                    }
                    return res(
                        ctx.json({
                            count: pool.length,
                            data: pool.slice(skip, skip + limit),
                            limit,
                            skip,
                        })
                    );
                })
            );
        });

        it('resets skip to 0 when filters change so rows are visible on the new result set', async () => {
            const user = userEvent.setup();
            render(<FileIngestTable />);

            // Initial unfiltered load: 15 jobs total, page 0 shows 10.
            await waitFor(() => expect(visibleIngestRows()).toHaveLength(10));
            expect(lastFileUploadRequest().searchParams.get('skip')).toBe('0');

            // Page 1 of the unfiltered set shows the remaining 5 rows.
            await user.click(screen.getByRole('button', { name: /go to next page/i }));
            await waitFor(() => expect(lastFileUploadRequest().searchParams.get('skip')).toBe('10'));
            await waitFor(() => expect(visibleIngestRows()).toHaveLength(5));

            // Apply Complete filter. The filtered pool only has 5 jobs, so without resetting page
            // the request would be skip=10 against a 5-row pool and the table would render empty.
            await user.click(screen.getByTestId('file_ingest_log-open_filter_dialog'));
            await user.click(await screen.findByRole('combobox', { name: 'Status Select' }));
            await user.click(await screen.findByRole('option', { name: 'Complete' }));
            await user.click(screen.getByTestId('file_ingest_log-filter_dialog_confirm'));

            await waitFor(() => {
                const last = lastFileUploadRequest();
                expect(last.searchParams.get('skip')).toBe('0');
                expect(last.searchParams.get('status')).toBe('eq:2');
            });
            await waitFor(() => expect(visibleIngestRows()).toHaveLength(5));
        });

        it('resets skip to 0 when rows per page changes so rows are visible on the new page size', async () => {
            const user = userEvent.setup();
            render(<FileIngestTable />);

            await waitFor(() => expect(visibleIngestRows()).toHaveLength(10));
            await user.click(screen.getByRole('button', { name: /go to next page/i }));
            await waitFor(() => expect(lastFileUploadRequest().searchParams.get('skip')).toBe('10'));
            await waitFor(() => expect(visibleIngestRows()).toHaveLength(5));

            // Change rows per page to 25. The full pool is 15 jobs, so without resetting page
            // the request would be skip=25 (page 1 * 25) and the table would render empty.
            await user.click(screen.getByRole('combobox', { name: /rows per page/i }));
            await user.click(await screen.findByRole('option', { name: '25' }));

            await waitFor(() => {
                const last = lastFileUploadRequest();
                expect(last.searchParams.get('skip')).toBe('0');
                expect(last.searchParams.get('limit')).toBe('25');
            });
            await waitFor(() => expect(visibleIngestRows()).toHaveLength(15));
        });
    });
});
