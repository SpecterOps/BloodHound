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

import { rest } from 'msw';
import { setupServer } from 'msw/node';
import { act, render, screen, waitFor } from '../../test-utils';
import FileIngest from './FileIngest';

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

const server = setupServer(
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
    }),
    rest.get('/api/v2/file-upload', (req, res, ctx) => {
        return res(
            ctx.json({
                data: [
                    {
                        status: 2,
                        status_message: 'Complete',
                        id: 1,
                        start_time: '2023-08-01T22:03:20.245299Z',
                        end_time: '2023-08-01T22:04:23.097927Z',
                        user_email_address: 'test_email@specterops.io',
                    },
                ],

                status: 200,
                statusText: 'OK',
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

describe('FileIngest', () => {
    it('displays a Upload Files button', async () => {
        await act(async () => render(<FileIngest />));
        const uploadButton = screen.getByRole('button', { name: 'Upload File(s)' });
        expect(uploadButton).toBeInTheDocument();
    });
    it('displays a Filters button', async () => {
        await act(async () => render(<FileIngest />));
        const filterButton = screen.getByRole('button', { name: /app-icon-filter-outline/i });
        expect(filterButton).toBeInTheDocument();
    });
    it('displays a table of completed ingest logs', async () => {
        checkPermissionMock.mockImplementation(() => true);
        render(<FileIngest />);
        await waitFor(() => screen.getByText('test_email@specterops.io'));

        expect(screen.getByText('test_email@specterops.io')).toBeInTheDocument();
        expect(screen.getByText('1 min')).toBeInTheDocument();
    });
});
