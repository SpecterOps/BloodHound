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

import type { FileIngestJob } from 'js-client-library';
import { rest } from 'msw';
import { setupServer } from 'msw/node';
import { act, render, screen } from '../../test-utils';
import { FileIngestTable } from './FileIngestTable';

const checkPermissionMock = vi.fn();

vi.mock('../../hooks/usePermissions', async () => {
    const actual = await vi.importActual('../../hooks');
    return {
        ...actual,
        usePermissions: () => ({
            checkPermission: checkPermissionMock,
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

const MOCK_INGEST_JOBS_RESPONSE = {
    count: 20,
    data: new Array(10).fill(MOCK_INGEST_JOB).map((item, index) => ({
        ...item,
        id: index,
        status: (index % 10) - 1,
    })),
    limit: 10,
    skip: 10,
};

const server = setupServer(
    rest.get('/api/v2/file-upload', (req, res, ctx) => res(ctx.json(MOCK_INGEST_JOBS_RESPONSE)))
);

beforeAll(() => server.listen());
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

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
});
