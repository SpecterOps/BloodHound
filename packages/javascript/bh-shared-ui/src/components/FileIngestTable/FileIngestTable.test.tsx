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
import { act, render, screen } from '../../test-utils';
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
});
