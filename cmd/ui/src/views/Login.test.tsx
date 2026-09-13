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

import { DeepPartial } from 'bh-shared-ui';
import { configureStore } from '@reduxjs/toolkit';
import { render } from '@testing-library/react';
import { rest } from 'msw';
import { setupServer } from 'msw/node';
import { HelmetProvider } from 'react-helmet-async';
import { QueryClient, QueryClientProvider } from 'react-query';
import { Provider } from 'react-redux';
import { MemoryRouter } from 'react-router-dom';
import { authSlice } from 'src/ducks/auth/authSlice';
import { AppState, rootReducer } from 'src/store';
import Login from 'src/views/Login';

const { mockNavigate } = vi.hoisted(() => ({ mockNavigate: vi.fn() }));

vi.mock('react-router-dom', async (importOriginal) => {
    const actual = await importOriginal<typeof import('react-router-dom')>();
    return {
        ...actual,
        Navigate: (props: any) => {
            mockNavigate(props);
            return null;
        },
    };
});

const server = setupServer(
    rest.get('/api/v2/sso-providers', (req, res, ctx) => {
        return res(
            ctx.json({
                endpoints: [],
            })
        );
    })
);

beforeAll(() => server.listen());
afterEach(() => {
    server.resetHandlers();
    mockNavigate.mockClear();
});
afterAll(() => server.close());

const renderLogin = (locationState?: { from?: { pathname?: string } }) => {
    const initialState: DeepPartial<AppState> = {
        auth: {
            ...authSlice.initialState,
            sessionToken: 'test-token',
            user: {},
        },
    };
    const store = configureStore({ reducer: rootReducer, preloadedState: initialState });
    const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });

    return render(
        <HelmetProvider>
            <Provider store={store}>
                <QueryClientProvider client={queryClient}>
                    <MemoryRouter initialEntries={[{ pathname: '/login', state: locationState }]}>
                        <Login />
                    </MemoryRouter>
                </QueryClientProvider>
            </Provider>
        </HelmetProvider>
    );
};

describe('Login', () => {
    it('redirects to the deep link location stored in navigation state when already authenticated', async () => {
        renderLogin({ from: { pathname: '/explore' } });

        expect(mockNavigate).toHaveBeenCalledWith(
            expect.objectContaining({
                to: '/explore',
                replace: true,
            })
        );
    });

    it('redirects to the home route when no deep link location is stored in navigation state', async () => {
        renderLogin();

        expect(mockNavigate).toHaveBeenCalledWith(
            expect.objectContaining({
                to: '/',
                replace: true,
            })
        );
    });

    it('redirects to the home route when the stored deep link location is the login page itself', async () => {
        renderLogin({ from: { pathname: '/login' } });

        expect(mockNavigate).toHaveBeenCalledWith(
            expect.objectContaining({
                to: '/',
                replace: true,
            })
        );
    });
});
