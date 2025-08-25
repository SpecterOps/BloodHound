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
import { Role } from 'js-client-library';
import { rest } from 'msw';
import { setupServer } from 'msw/node';
import { renderHook, waitFor } from '../../test-utils';
import { useListDisplayRoles } from './useListDisplayRoles';

const MockRoles: Role[] = [
    {
        name: 'Upload-Only',
        description: 'Used for users to ingest files manually',
        permissions: [
            {
                authority: 'graphdb',
                name: 'Ingest',
                id: 30,
            },
        ],
        id: 4,
    },
    {
        name: 'Read-Only',
        description: 'Used for integrations',
        permissions: [
            {
                authority: 'app',
                name: 'ReadAppConfig',
                id: 13,
            },
            {
                authority: 'risks',
                name: 'GenerateReport',
                id: 10,
            },
            {
                authority: 'auth',
                name: 'CreateToken',
                id: 6,
            },
            {
                authority: 'auth',
                name: 'ManageSelf',
                id: 9,
            },
            {
                authority: 'graphdb',
                name: 'Read',
                id: 2,
            },
            {
                authority: 'saved_queries',
                name: 'Read',
                id: 15,
            },
        ],
        id: 3,
    },
    {
        name: 'User',
        description: 'Can read data, modify asset group memberships',
        permissions: [
            {
                authority: 'app',
                name: 'ReadAppConfig',
                id: 13,
            },
            {
                authority: 'risks',
                name: 'GenerateReport',
                id: 10,
            },
            {
                authority: 'auth',
                name: 'CreateToken',
                id: 6,
            },
            {
                authority: 'auth',
                name: 'ManageSelf',
                id: 9,
            },
            {
                authority: 'clients',
                name: 'Read',
                id: 17,
            },
            {
                authority: 'graphdb',
                name: 'Read',
                id: 2,
            },
            {
                authority: 'saved_queries',
                name: 'Read',
                id: 15,
            },
            {
                authority: 'saved_queries',
                name: 'Write',
                id: 16,
            },
        ],
        id: 2,
    },
    {
        name: 'Power User',
        description: 'Can upload data, manage clients, and perform any action a User can',
        permissions: [
            {
                authority: 'graphdb',
                name: 'Ingest',
                id: 30,
            },
            {
                authority: 'app',
                name: 'ReadAppConfig',
                id: 13,
            },
            {
                authority: 'risks',
                name: 'GenerateReport',
                id: 10,
            },
            {
                authority: 'risks',
                name: 'ManageRisks',
                id: 11,
            },
            {
                authority: 'auth',
                name: 'CreateToken',
                id: 6,
            },
            {
                authority: 'auth',
                name: 'ManageSelf',
                id: 9,
            },
            {
                authority: 'clients',
                name: 'Manage',
                id: 4,
            },
            {
                authority: 'clients',
                name: 'Read',
                id: 17,
            },
            {
                authority: 'clients',
                name: 'Tasking',
                id: 5,
            },
            {
                authority: 'collection',
                name: 'ManageJobs',
                id: 3,
            },
            {
                authority: 'graphdb',
                name: 'Mutate',
                id: 19,
            },
            {
                authority: 'graphdb',
                name: 'Read',
                id: 2,
            },
            {
                authority: 'graphdb',
                name: 'Write',
                id: 1,
            },
            {
                authority: 'saved_queries',
                name: 'Read',
                id: 15,
            },
            {
                authority: 'saved_queries',
                name: 'Write',
                id: 16,
            },
        ],
        id: 5,
    },
    {
        name: 'Administrator',
        description: 'Can manage users, clients, and application configuration',
        permissions: [
            {
                authority: 'graphdb',
                name: 'Ingest',
                id: 30,
            },
            {
                authority: 'app',
                name: 'ReadAppConfig',
                id: 13,
            },
            {
                authority: 'app',
                name: 'WriteAppConfig',
                id: 14,
            },
            {
                authority: 'risks',
                name: 'GenerateReport',
                id: 10,
            },
            {
                authority: 'risks',
                name: 'ManageRisks',
                id: 11,
            },
            {
                authority: 'auth',
                name: 'CreateToken',
                id: 6,
            },
            {
                authority: 'auth',
                name: 'ManageAppConfig',
                id: 12,
            },
            {
                authority: 'auth',
                name: 'ManageProviders',
                id: 8,
            },
            {
                authority: 'auth',
                name: 'ManageSelf',
                id: 9,
            },
            {
                authority: 'auth',
                name: 'ManageUsers',
                id: 7,
            },
            {
                authority: 'clients',
                name: 'Manage',
                id: 4,
            },
            {
                authority: 'clients',
                name: 'Read',
                id: 17,
            },
            {
                authority: 'clients',
                name: 'Tasking',
                id: 5,
            },
            {
                authority: 'collection',
                name: 'ManageJobs',
                id: 3,
            },
            {
                authority: 'graphdb',
                name: 'Mutate',
                id: 19,
            },
            {
                authority: 'graphdb',
                name: 'Read',
                id: 2,
            },
            {
                authority: 'graphdb',
                name: 'Write',
                id: 1,
            },
            {
                authority: 'saved_queries',
                name: 'Read',
                id: 15,
            },
            {
                authority: 'saved_queries',
                name: 'Write',
                id: 16,
            },
            {
                authority: 'db',
                name: 'Wipe',
                id: 18,
            },
        ],
        id: 1,
    },
];

const fetchRolesRequest = () => {
    return rest.get('/api/v2/roles', async (_req, rest, ctx) => {
        return rest(
            ctx.json({
                data: {
                    roles: MockRoles,
                },
            })
        );
    });
};

const server = setupServer(fetchRolesRequest());
beforeAll(() => server.listen());
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

describe('useListDisplayRoles', () => {
    it('List Display Roles Hook is Loading', async () => {
        const { result } = renderHook(() => useListDisplayRoles());

        expect(result.current.isLoading).toEqual(true);
    });
    it('Validate Displays all the Roles', async () => {
        const { result } = renderHook(() => useListDisplayRoles());

        await waitFor(() => {
            expect(result.current.data).toHaveLength(5);
        });
    });
});
