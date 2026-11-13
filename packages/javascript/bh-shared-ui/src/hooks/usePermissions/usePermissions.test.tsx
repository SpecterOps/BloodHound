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
import { renderHook, waitFor } from '../../test-utils';
import { Permission, type PermissionDefinition, PERMISSIONS } from '../../utils/permissions';
import { usePermissions } from './usePermissions';

const clientsManage = PERMISSIONS[Permission.CLIENTS_MANAGE];
const authCreateToken = PERMISSIONS[Permission.AUTH_CREATE_TOKEN];
const grantedPermissions = [clientsManage, authCreateToken];
const requiredPermissions = [Permission.CLIENTS_MANAGE, Permission.AUTH_CREATE_TOKEN];
const server = setupServer();

beforeAll(() => server.listen({ onUnhandledRequest: 'error' }));
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

const renderPermissions = async (permissions: PermissionDefinition[]) => {
    server.use(rest.get('/api/v2/self', (_req, res, ctx) => res(ctx.json({ data: { roles: [{ permissions }] } }))));
    const { result } = renderHook(() => usePermissions());
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    return result;
};

describe('usePermissions', () => {
    it.each([
        {
            name: 'no grants',
            granted: [],
            required: requiredPermissions,
            expected: { single: false, all: false, any: false },
        },
        {
            name: 'one matching permission',
            granted: [clientsManage],
            required: [Permission.CLIENTS_MANAGE],
            expected: { single: true, all: true, any: true },
        },
        {
            name: 'all required permissions',
            granted: grantedPermissions,
            required: requiredPermissions,
            expected: { single: true, all: true, any: true },
        },
        {
            name: 'only one of multiple required permissions',
            granted: [authCreateToken],
            required: requiredPermissions,
            expected: { single: false, all: false, any: true },
        },
        {
            name: 'matching name under a different authority',
            granted: [PERMISSIONS[Permission.ALERTS_MANAGE]],
            required: [Permission.CLIENTS_MANAGE],
            expected: { single: false, all: false, any: false },
        },
        {
            name: 'matching authority with a different name',
            granted: [PERMISSIONS[Permission.CLIENTS_READ]],
            required: [Permission.CLIENTS_MANAGE],
            expected: { single: false, all: false, any: false },
        },
    ])('checks single, all, and any permissions with $name', async ({ granted, required, expected }) => {
        const result = await renderPermissions(granted);

        expect(result.current.checkPermission(required[0])).toBe(expected.single);
        expect(result.current.checkAllPermissions(required)).toBe(expected.all);
        expect(result.current.checkAtLeastOnePermission(required)).toBe(expected.any);
    });

    it('returns the users current permissions', async () => {
        const result = await renderPermissions(grantedPermissions);
        expect(result.current.getUserPermissions()).toEqual(expect.arrayContaining(requiredPermissions));
        expect(result.current.getUserPermissions()).toHaveLength(requiredPermissions.length);
    });
});
