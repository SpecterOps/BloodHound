// Copyright 2024 Specter Ops, Inc.
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
import { QueryClient } from 'react-query';
import { renderHook, waitFor } from '../../test-utils';
import { Permission } from '../../utils';
import { usePermissions } from './usePermissions';

const testClientsManagePermission = {
    authority: 'clients',
    name: 'Manage',
};

const testAuthCreateTokenPermission = {
    authority: 'auth',
    name: 'CreateToken',
};

const testAppReadAppConfigPermission = {
    authority: 'app',
    name: 'ReadAppConfig',
};

const allPermissions = [
    Permission.CLIENTS_MANAGE,
    Permission.AUTH_CREATE_TOKEN,
    Permission.APP_READ_APPLICATION_CONFIGURATION,
];

const queryClient = new QueryClient();

const server = setupServer(
    rest.get(`/api/v2/self`, async (_req, res, ctx) => {
        return res(
            ctx.json({
                data: {
                    roles: [],
                },
            })
        );
    })
);
beforeAll(() => server.listen());
afterEach(() => {
    queryClient.clear();
    server.resetHandlers();
});
afterAll(() => server.close());

describe('usePermissions', () => {
    it('passes check if the user has a required permission', async () => {
        server.use(
            rest.get(`/api/v2/self`, async (_req, res, ctx) => {
                return res(
                    ctx.json({
                        data: {
                            roles: [
                                {
                                    permissions: [testClientsManagePermission],
                                },
                            ],
                        },
                    })
                );
            })
        );
        const { result } = renderHook(() => usePermissions());
        await waitFor(() => expect(result.current.isLoading).toBe(false));

        const has = result.current.checkPermission(Permission.CLIENTS_MANAGE);
        const hasAll = result.current.checkAllPermissions([Permission.CLIENTS_MANAGE]);
        const hasAtLeastOne = result.current.checkAtLeastOnePermission([Permission.CLIENTS_MANAGE]);

        expect(has).toBe(true);
        expect(hasAll).toBe(true);
        expect(hasAtLeastOne).toBe(true);
    });

    it('passes checks if the user has multiple required permissions', async () => {
        server.use(
            rest.get(`/api/v2/self`, async (_req, res, ctx) => {
                return res(
                    ctx.json({
                        data: {
                            roles: [
                                {
                                    permissions: [
                                        testClientsManagePermission,
                                        testAuthCreateTokenPermission,
                                        testAppReadAppConfigPermission,
                                    ],
                                },
                            ],
                        },
                    })
                );
            })
        );
        const { result } = renderHook(() => usePermissions());
        await waitFor(() => expect(result.current.isLoading).toBe(false));

        const hasAll = result.current.checkAllPermissions(allPermissions);
        const hasAtLeastOne = result.current.checkAtLeastOnePermission(allPermissions);

        expect(hasAll).toBe(true);
        expect(hasAtLeastOne).toBe(true);
    });

    it('fails checks if the user does not have a matching permission', async () => {
        server.use(
            rest.get(`/api/v2/self`, async (_req, res, ctx) => {
                return res(
                    ctx.json({
                        data: {
                            roles: [
                                {
                                    permissions: [testClientsManagePermission],
                                },
                            ],
                        },
                    })
                );
            })
        );
        const { result } = renderHook(() => usePermissions());
        await waitFor(() => expect(result.current.isLoading).toBe(false));

        const has = result.current.checkPermission(Permission.AUTH_CREATE_TOKEN);
        const hasAll = result.current.checkAllPermissions([Permission.AUTH_CREATE_TOKEN]);
        const hasAtLeastOne = result.current.checkAtLeastOnePermission([Permission.AUTH_CREATE_TOKEN]);

        expect(has).toBe(false);
        expect(hasAll).toBe(false);
        expect(hasAtLeastOne).toBe(false);
    });

    it('passes the check for at least one permission if the user is missing one of many required permissions', async () => {
        server.use(
            rest.get(`/api/v2/self`, async (_req, res, ctx) => {
                return res(
                    ctx.json({
                        data: {
                            roles: [
                                {
                                    permissions: [testAuthCreateTokenPermission, testAppReadAppConfigPermission],
                                },
                            ],
                        },
                    })
                );
            })
        );
        const { result } = renderHook(() => usePermissions());
        await waitFor(() => expect(result.current.isLoading).toBe(false));

        const hasAll = result.current.checkAllPermissions(allPermissions);
        const hasAtLeastOne = result.current.checkAtLeastOnePermission(allPermissions);

        expect(hasAll).toBe(false);
        expect(hasAtLeastOne).toBe(true);
    });

    it('returns a list of the users current permissions', async () => {
        server.use(
            rest.get(`/api/v2/self`, async (_req, res, ctx) => {
                return res(
                    ctx.json({
                        data: {
                            roles: [
                                {
                                    permissions: [
                                        testClientsManagePermission,
                                        testAuthCreateTokenPermission,
                                        testAppReadAppConfigPermission,
                                    ],
                                },
                            ],
                        },
                    })
                );
            })
        );
        const { result } = renderHook(() => usePermissions());
        await waitFor(() => expect(result.current.isLoading).toBe(false));

        const userPermissionsResult = result.current.getUserPermissions();

        expect(allPermissions.length).toEqual(userPermissionsResult.length);

        for (const permission of allPermissions) {
            expect(userPermissionsResult).toContain(permission);
        }
    });
});
