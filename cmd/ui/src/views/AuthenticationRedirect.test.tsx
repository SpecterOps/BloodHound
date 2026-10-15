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

import userEvent from '@testing-library/user-event';
import { DeepPartial, apiClient } from 'bh-shared-ui';
import { createAuthStateWithPermissions } from 'bh-shared-ui/testing';
import { Route, Routes } from 'react-router-dom';
import AuthenticatedRoute from 'src/components/AuthenticatedRoute/AuthenticatedRoute';
import { initialState as initialAuthState } from 'src/ducks/auth/authSlice';
import { LOGIN_REDIRECT_STORAGE_KEY } from 'src/hooks/useLoginRedirect';
import { AppState } from 'src/store';
import { render, screen, waitFor } from 'src/test-utils';
import ExpiredPassword from './ExpiredPassword';
import Home from './Home/Home';
import Login from './Login';

const destination = '/explore?searchType=pathfinding&primarySearch=A&secondarySearch=B#graph';
const authenticatedState = {
    ...initialAuthState,
    user: {
        ...createAuthStateWithPermissions([]).user,
        id: 'user-id',
        AuthSecret: { expires_at: '9999-01-01T00:00:00Z' },
    },
    isInitialized: true,
    sessionToken: 'session-token',
};
const TestRoutes = () => (
    <Routes>
        <Route path='/login' element={<Login />} />
        <Route
            path='/'
            element={
                <AuthenticatedRoute>
                    <Home />
                </AuthenticatedRoute>
            }
        />
        <Route
            path='/explore'
            element={
                <AuthenticatedRoute>
                    <div>Graph destination</div>
                </AuthenticatedRoute>
            }
        />
        <Route
            path='/expired-password'
            element={
                <AuthenticatedRoute>
                    <ExpiredPassword />
                </AuthenticatedRoute>
            }
        />
    </Routes>
);

beforeEach(() => {
    sessionStorage.clear();
    vi.spyOn(apiClient, 'listSSOProviders').mockResolvedValue({ data: { data: [] } } as any);
});
afterEach(() => {
    sessionStorage.clear();
    vi.restoreAllMocks();
    vi.unstubAllGlobals();
});

const expectDestination = async () => {
    expect(await screen.findByText('Graph destination')).toBeInTheDocument();
    expect(window.location.pathname + window.location.search + window.location.hash).toBe(destination);
};

const renderRoutes = (auth: DeepPartial<AppState['auth']>, route = destination) =>
    render(<TestRoutes />, { initialState: { auth }, route });

const submitPasswordReset = async () => {
    const user = userEvent.setup();
    await user.type(await screen.findByLabelText('Expired password'), 'OldPassword123!');
    await user.type(screen.getByLabelText('New Password'), 'NewPassword123!');
    await user.type(screen.getByLabelText('New Password Confirmation'), 'NewPassword123!');
    await user.click(screen.getByRole('button', { name: /reset password/i }));
};

describe('authentication deeplink redirects', () => {
    it('preserves a deeplink through username/password login', async () => {
        vi.spyOn(apiClient, 'login').mockResolvedValue({ data: { data: { session_token: 'session-token' } } } as any);
        vi.spyOn(apiClient.baseClient, 'get').mockResolvedValue({ data: { data: authenticatedState.user } });
        renderRoutes(initialAuthState);
        const user = userEvent.setup();
        await user.type(await screen.findByLabelText('Email Address'), 'user@example.com');
        await user.type(screen.getByLabelText('Password'), 'Password123!');
        await user.click(screen.getByRole('button', { name: 'LOGIN' }));
        await expectDestination();
    });

    it('saves a deeplink before leaving for SSO and restores it on the landing page', async () => {
        vi.spyOn(apiClient, 'listSSOProviders').mockResolvedValue({ data: { data: [{ slug: 'provider' }] } } as any);
        const assign = vi.fn();
        const originalWindow = window;
        const { unmount } = renderRoutes(initialAuthState);
        await screen.findByRole('button', { name: 'LOGIN VIA SSO' });
        vi.stubGlobal(
            'window',
            new Proxy(originalWindow, {
                get(target, property) {
                    if (property === 'location') return { ...target.location, assign };
                    return Reflect.get(target, property, target);
                },
            })
        );
        await userEvent.click(screen.getByRole('button', { name: 'LOGIN VIA SSO' }));
        await waitFor(() => expect(assign).toHaveBeenCalledWith('/api/v2/sso/provider/login'));
        expect(sessionStorage.getItem(LOGIN_REDIRECT_STORAGE_KEY)).toBe(destination);
        unmount();
        vi.unstubAllGlobals();

        // A full SSO round trip loses router state and returns an authenticated user to /ui.
        renderRoutes(authenticatedState, '/');
        await expectDestination();
        expect(sessionStorage.getItem(LOGIN_REDIRECT_STORAGE_KEY)).toBeNull();
    });

    it('preserves a deeplink through an expired-password reset', async () => {
        vi.spyOn(apiClient, 'putUserAuthSecret').mockResolvedValue({} as any);
        vi.spyOn(apiClient, 'getSelf').mockResolvedValue({ data: { data: authenticatedState.user } } as any);
        renderRoutes({
            ...authenticatedState,
            user: { ...authenticatedState.user, AuthSecret: { expires_at: '1970-01-01T00:00:00Z' } },
        });
        await submitPasswordReset();
        await expectDestination();
    });

    it('keeps the SSO destination if the landing page requires a password reset', async () => {
        sessionStorage.setItem(LOGIN_REDIRECT_STORAGE_KEY, destination);
        vi.spyOn(apiClient, 'putUserAuthSecret').mockResolvedValue({} as any);
        vi.spyOn(apiClient, 'getSelf').mockResolvedValue({ data: { data: authenticatedState.user } } as any);
        renderRoutes(
            {
                ...authenticatedState,
                user: { ...authenticatedState.user, AuthSecret: { expires_at: '1970-01-01T00:00:00Z' } },
            },
            '/'
        );
        await submitPasswordReset();
        await expectDestination();
        expect(sessionStorage.getItem(LOGIN_REDIRECT_STORAGE_KEY)).toBeNull();
    });

    it.each(['/', '/?searchType=pathfinding', '//example.com', '/\\example.com', '/login', '/expired-password'])(
        'ignores unsafe or looping stored destinations: %s',
        async (savedDestination) => {
            sessionStorage.setItem(LOGIN_REDIRECT_STORAGE_KEY, savedDestination);
            renderRoutes(authenticatedState, '/');
            expect(await screen.findByText('Graph destination')).toBeInTheDocument();
            expect(window.location.pathname).toBe('/explore');
            expect(window.location.search).toBe('');
            expect(sessionStorage.getItem(LOGIN_REDIRECT_STORAGE_KEY)).toBeNull();
        }
    );

    it('uses the usual home destination when no saved destination exists', async () => {
        renderRoutes(authenticatedState, '/login');
        expect(await screen.findByText('Graph destination')).toBeInTheDocument();
        expect(window.location.pathname).toBe('/explore');
    });

    it('still allows authentication when session storage is unavailable', async () => {
        vi.spyOn(Storage.prototype, 'getItem').mockImplementation(() => {
            throw new Error('Storage blocked');
        });
        vi.spyOn(Storage.prototype, 'removeItem').mockImplementation(() => {
            throw new Error('Storage blocked');
        });
        renderRoutes(authenticatedState, '/login');
        expect(await screen.findByText('Graph destination')).toBeInTheDocument();
    });
});
