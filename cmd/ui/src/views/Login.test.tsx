// Copyright 2026 Specter Ops, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
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

import { DeepPartial } from 'bh-shared-ui';
import { createAuthStateWithPermissions } from 'bh-shared-ui/testing';
import { rest } from 'msw';
import { setupServer } from 'msw/node';
import { AppState } from 'src/store';
import { render, waitFor } from 'src/test-utils';

import Login from './Login';

let locationState: {
    from?: {
        hash?: string;
        pathname: string;
        search?: string;
    };
} | null = null;

vi.mock('react-router-dom', async () => ({
    ...(await vi.importActual<typeof import('react-router-dom')>('react-router-dom')),
    useLocation: () => ({
        hash: '',
        key: 'default',
        pathname: '/login',
        search: '',
        state: locationState,
    }),
}));

const server = setupServer(
    rest.get('/api/v2/sso-providers', (_req, res, ctx) => {
        return res(ctx.json({ data: { data: [] } }));
    })
);

const authenticatedState: DeepPartial<AppState> = {
    auth: {
        ...createAuthStateWithPermissions([]),
        sessionToken: 'session-token',
    },
};

beforeAll(() => server.listen());
afterEach(() => {
    locationState = null;
    server.resetHandlers();
});
afterAll(() => server.close());

describe('Login', () => {
    it('redirects an authenticated user to the saved destination including its query and hash', async () => {
        locationState = {
            from: {
                hash: '#findings',
                pathname: '/explore',
                search: '?type=computer',
            },
        };

        render(<Login />, { initialState: authenticatedState, route: '/login' });

        await waitFor(() => {
            expect(window.location.pathname).toBe('/explore');
            expect(window.location.search).toBe('?type=computer');
            expect(window.location.hash).toBe('#findings');
        });
    });

    it('redirects an authenticated user home when no saved destination exists', async () => {
        render(<Login />, { initialState: authenticatedState, route: '/login' });

        await waitFor(() => expect(window.location.pathname).toBe('/'));
    });
});
