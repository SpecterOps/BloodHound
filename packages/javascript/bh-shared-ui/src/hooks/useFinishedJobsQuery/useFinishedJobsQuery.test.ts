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

import { type ScheduledJobDisplay } from 'js-client-library';
import { rest } from 'msw';
import { setupServer } from 'msw/node';

import { renderHook, waitFor } from '../../test-utils';
import { useFinishedJobsQuery } from './useFinishedJobsQuery';

const addNotificationMock = vi.fn();
const dismissNotificationMock = vi.fn();
const checkPermissionMock = vi.fn();

vi.mock('../../providers', async () => {
    const actual = await vi.importActual('../../providers');
    return {
        ...actual,
        useNotifications: () => ({
            addNotification: addNotificationMock,
            dismissNotification: dismissNotificationMock,
        }),
    };
});

vi.mock('../../hooks/usePermissions', async () => {
    const actual = await vi.importActual('../../hooks');
    return {
        ...actual,
        usePermissions: () => ({
            checkPermission: checkPermissionMock,
        }),
    };
});

const MOCK_FINISHED_JOB: ScheduledJobDisplay = {
    id: 22,
    client_id: '718c9b04-9394-42c0-9d53-c87b689e2d92',
    client_name: 'GOAD',
    event_id: 123,
    execution_time: '2024-08-15T21:24:52.366579Z',
    start_time: '2024-08-15T21:25:21.990437Z',
    end_time: '2024-08-15T21:26:43.033448Z',
    status: 2,
    status_message: 'The service collected successfully',
    session_collection: true,
    local_group_collection: true,
    ad_structure_collection: true,
    cert_services_collection: true,
    ca_registry_collection: true,
    dc_registry_collection: true,
    all_trusted_domains: true,
    domain_controller: '',
    ous: [],
    domains: [],
    domain_results: [],
};

const MOCK_FINISHED_JOBS_RESPONSE = {
    count: 20,
    data: new Array(10).fill(MOCK_FINISHED_JOB).map((item, index) => ({
        ...item,
        id: index,
        status: (index % 10) - 1,
    })),
    limit: 10,
    skip: 10,
};

const server = setupServer(
    rest.get('/api/v2/jobs/finished', (req, res, ctx) => {
        return res(ctx.json(MOCK_FINISHED_JOBS_RESPONSE));
    })
);

beforeAll(() => server.listen());
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

describe('useFinishedJobsQuery', () => {
    it('requests finished jobs', async () => {
        checkPermissionMock.mockImplementation(() => true);
        const { result } = renderHook(() => useFinishedJobsQuery({ page: 0, rowsPerPage: 10 }));
        await waitFor(() => expect(result.current.isLoading).toBe(false));

        expect(result.current.data.data.length).toBe(10);
    });

    it('shows "no permission" notification if lacking permission', async () => {
        checkPermissionMock.mockImplementation(() => false);
        const { result } = renderHook(() => useFinishedJobsQuery({ page: 0, rowsPerPage: 10 }));
        await waitFor(() => expect(result.current.isLoading).toBe(false));

        expect(addNotificationMock).toHaveBeenCalledWith(
            expect.stringContaining('does not permit viewing'),
            'finished-jobs-permission',
            expect.objectContaining({ persist: true })
        );
    });

    it('does not request finished jobs if lacking permission', async () => {
        checkPermissionMock.mockImplementation(() => false);
        const { result } = renderHook(() => useFinishedJobsQuery({ page: 0, rowsPerPage: 10 }));
        await waitFor(() => expect(result.current.isLoading).toBe(false));
        expect(result.current.data).toBeUndefined();
    });

    it('shows an error notification if there is an error fetching', async () => {
        const errorSpy = vi.spyOn(console, 'error').mockImplementation(() => {});
        server.use(rest.get('/api/v2/jobs/finished', (req, res, ctx) => res(ctx.status(400))));
        checkPermissionMock.mockImplementation(() => true);
        const { result } = renderHook(() => useFinishedJobsQuery({ page: 0, rowsPerPage: 10 }));
        await waitFor(() => expect(result.current.isLoading).toBe(false));
        errorSpy.mockRestore();

        expect(addNotificationMock).toHaveBeenCalledWith(
            expect.stringContaining('Unable to fetch finished jobs'),
            'finished-jobs-error'
        );
    });

    it('dismisses the "no permission" notification on unmount', async () => {
        checkPermissionMock.mockImplementation(() => false);
        const { unmount } = renderHook(() => useFinishedJobsQuery({ page: 0, rowsPerPage: 10 }));
        unmount();
        expect(dismissNotificationMock).toHaveBeenCalledWith('finished-jobs-permission');
    });
});
