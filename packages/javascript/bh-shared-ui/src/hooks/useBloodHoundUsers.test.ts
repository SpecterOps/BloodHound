// Copyright 2026 Specter Ops, Inc.
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
import { renderHook, waitFor } from '../test-utils';
import { apiClient } from '../utils';
import { useGetUser, useGetUserNameById, useUserNamesById } from './useBloodHoundUsers';

const MOCK_USER_ID = '718c9b04-9394-42c0-9d53-c87b689e2d92';

const MOCK_USER = {
    id: MOCK_USER_ID,
    first_name: 'Ada',
    last_name: 'Lovelace',
    email_address: 'ada@example.com',
};

const MOCK_USERS = [
    { id: '00000000-0000-0000-0000-000000000001', principal_name: 'ada@example.com' },
    { id: '00000000-0000-0000-0000-000000000002', principal_name: 'grace@example.com' },
];

const readUsersPermission = { authority: 'auth', name: 'ReadUsers' };

const server = setupServer(
    rest.get(`/api/v2/bloodhound-users/${MOCK_USER_ID}`, (req, res, ctx) => {
        return res(ctx.json({ data: MOCK_USER }));
    })
);

beforeAll(() => server.listen());
afterEach(() => {
    server.resetHandlers();
    vi.restoreAllMocks();
});
afterAll(() => server.close());

describe('useGetUser', () => {
    it('does not request a user when no userId is provided', async () => {
        const getUserSpy = vi.spyOn(apiClient, 'getUser');
        const { result } = renderHook(() => useGetUser(undefined));

        expect(getUserSpy).not.toHaveBeenCalled();
        expect(result.current.isLoading).toBe(false);
        expect(result.current.data).toBeUndefined();
    });

    it('requests the user when a userId is provided', async () => {
        const getUserSpy = vi.spyOn(apiClient, 'getUser');
        const { result } = renderHook(() => useGetUser(MOCK_USER_ID));

        await waitFor(() => expect(result.current.isSuccess).toBe(true));

        expect(getUserSpy).toHaveBeenCalledWith(
            MOCK_USER_ID,
            expect.objectContaining({ signal: expect.any(AbortSignal) })
        );
        expect(result.current.data).toEqual(MOCK_USER);
    });
});

describe('useUserNamesById', () => {
    it('returns a map of user id to principal_name when the user can read users', async () => {
        server.use(
            rest.get(`/api/v2/self`, (req, res, ctx) => {
                return res(ctx.json({ data: { roles: [{ permissions: [readUsersPermission] }] } }));
            }),
            rest.get(`/api/v2/bloodhound-users`, (req, res, ctx) => {
                return res(ctx.json({ data: { users: MOCK_USERS } }));
            })
        );

        const { result } = renderHook(() => useUserNamesById());

        await waitFor(() => expect(result.current.size).toBe(MOCK_USERS.length));

        expect(result.current.get(MOCK_USERS[0].id)).toBe(MOCK_USERS[0].principal_name);
        expect(result.current.get(MOCK_USERS[1].id)).toBe(MOCK_USERS[1].principal_name);
    });

    it('returns an empty map when the user lacks permission to read users', async () => {
        const listUsersSpy = vi.spyOn(apiClient, 'listUsers');
        server.use(
            rest.get(`/api/v2/self`, (req, res, ctx) => {
                return res(ctx.json({ data: { roles: [] } }));
            })
        );

        const { result } = renderHook(() => useUserNamesById());

        await waitFor(() => expect(result.current.size).toBe(0));
        expect(listUsersSpy).not.toHaveBeenCalled();
    });
});

describe('useGetUserNameById', () => {
    it('returns a function that looks up a user display name by id', async () => {
        server.use(
            rest.get(`/api/v2/self`, (req, res, ctx) => {
                return res(ctx.json({ data: { roles: [{ permissions: [readUsersPermission] }] } }));
            }),
            rest.get(`/api/v2/bloodhound-users`, (req, res, ctx) => {
                return res(ctx.json({ data: { users: MOCK_USERS } }));
            })
        );

        const { result } = renderHook(() => useGetUserNameById());

        await waitFor(() => expect(result.current(MOCK_USERS[0].id)).toBe(MOCK_USERS[0].principal_name));

        expect(result.current(MOCK_USERS[1].id)).toBe(MOCK_USERS[1].principal_name);
    });

    it('returns undefined for an unknown id or when no id is provided', async () => {
        server.use(
            rest.get(`/api/v2/self`, (req, res, ctx) => {
                return res(ctx.json({ data: { roles: [{ permissions: [readUsersPermission] }] } }));
            }),
            rest.get(`/api/v2/bloodhound-users`, (req, res, ctx) => {
                return res(ctx.json({ data: { users: MOCK_USERS } }));
            })
        );

        const { result } = renderHook(() => useGetUserNameById());

        await waitFor(() => expect(result.current(MOCK_USERS[0].id)).toBe(MOCK_USERS[0].principal_name));

        expect(result.current('unknown-id')).toBeUndefined();
        expect(result.current(undefined)).toBeUndefined();
    });
});
