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

import { createMemoryHistory } from 'history';
import { ROUTE_EXPIRED_PASSWORD, ROUTE_HOME, ROUTE_LOGIN } from 'src/ducks/global/routes';
import { render, screen } from 'src/test-utils';
import AuthenticatedRoute from './AuthenticatedRoute';

describe('AuthenticatedRoute', () => {
    it('when session token or user are null, redirects to /login', () => {
        const history = createMemoryHistory({ initialEntries: [ROUTE_HOME] });

        render(
            <AuthenticatedRoute>
                <div>authenticated</div>
            </AuthenticatedRoute>,
            {
                initialState: {
                    auth: {
                        sessionToken: null,
                        user: null,
                    },
                },
                history,
            }
        );

        expect(screen.queryByText('authenticated')).not.toBeInTheDocument();
        expect(history.location.pathname).toBe(ROUTE_LOGIN);
    });

    it('when password is expired and not on password reset page, redirects to password reset page', () => {
        const history = createMemoryHistory({ initialEntries: [ROUTE_HOME] });

        render(
            <AuthenticatedRoute>
                <div>authenticated</div>
            </AuthenticatedRoute>,
            {
                initialState: {
                    auth: {
                        sessionToken: 'validToken',
                        user: {
                            id: 'validUserId',
                            AuthSecret: {
                                expires_at: '1970-01-01T00:00:00Z', // expired
                            },
                        },
                    },
                },
                history,
            }
        );

        expect(screen.queryByText('authenticated')).not.toBeInTheDocument();
        expect(history.location.pathname).toBe(ROUTE_EXPIRED_PASSWORD);
    });

    it('when password is expired and on password reset page, no redirect occurs', () => {
        const history = createMemoryHistory({ initialEntries: [ROUTE_EXPIRED_PASSWORD] });

        render(
            <AuthenticatedRoute>
                <div>expired password page</div>
            </AuthenticatedRoute>,
            {
                initialState: {
                    auth: {
                        sessionToken: 'validToken',
                        user: {
                            id: 'validUserId',
                            AuthSecret: {
                                expires_at: '1970-01-01T00:00:00Z', // expired
                            },
                        },
                    },
                },
                history,
            }
        );

        expect(screen.queryByText('expired password page')).toBeInTheDocument();
        expect(history.location.pathname).toBe(ROUTE_EXPIRED_PASSWORD);
    });

    it('when password is not expired no redirect occurs', () => {
        const history = createMemoryHistory({ initialEntries: [ROUTE_HOME] });

        render(
            <AuthenticatedRoute>
                <div>authenticated</div>
            </AuthenticatedRoute>,
            {
                initialState: {
                    auth: {
                        sessionToken: 'validToken',
                        user: {
                            id: 'validUserId',
                            AuthSecret: {
                                expires_at: '9999-01-01T00:00:00Z', // not expired
                            },
                        },
                    },
                },
                history,
            }
        );

        expect(screen.queryByText('authenticated')).toBeInTheDocument();
        expect(history.location.pathname).toBe(ROUTE_HOME);
    });
});
