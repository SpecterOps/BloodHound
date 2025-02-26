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

import { DeepPartial, apiClient, createAuthStateWithPermissions } from 'bh-shared-ui';
import { rest } from 'msw';
import { setupServer } from 'msw/node';
import App, { Inner } from 'src/App';
import * as authSlice from 'src/ducks/auth/authSlice';
import { act, render, screen } from 'src/test-utils';
import { AppState } from './store';

const server = setupServer(
    rest.get('/api/v2/saml/sso', (req, res, ctx) => {
        return res(
            ctx.json({
                endpoints: [],
            })
        );
    }),
    rest.get('/api/v2/features', (req, res, ctx) => {
        return res(
            ctx.json({
                data: [],
            })
        );
    }),
    rest.get('/api/v2/available-domains', (req, res, ctx) => {
        return res(
            ctx.json({
                data: [],
            })
        );
    }),
    rest.get('/api/v2/self', (req, res, ctx) => {
        return res(
            ctx.json({
                data: createAuthStateWithPermissions(Object.values(Permissions)).user,
            })
        );
    })
);

beforeAll(() => server.listen());
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

// override manual mock
vi.unmock('react');

describe('app', () => {
    it.skip('renders', async () => {
        render(<App />);
        expect(await screen.findByText('LOGIN')).toBeInTheDocument();
    });

    describe('<Inner />', () => {
        const setup = async () => {
            await act(async () => {
                const initialState: DeepPartial<AppState> = {
                    auth: {
                        ...authSlice.initialState,
                        isInitialized: true,
                    },
                };
                render(<Inner />, { initialState });
            });
        };

        it('does not make feature-flag request if user is not fully authenticated', async () => {
            const featureFlagSpy = vi.spyOn(apiClient, 'getFeatureFlags');
            const fullyAuthenticatedSelectorSpy = vi.spyOn(authSlice, 'fullyAuthenticatedSelector');

            // hard code user as not authenticated
            fullyAuthenticatedSelectorSpy.mockReturnValue(false);

            await setup();

            expect(featureFlagSpy).not.toHaveBeenCalled();
        });

        it('will request feature-flags when the user is fully authenticated', async () => {
            const featureFlagSpy = vi.spyOn(apiClient, 'getFeatureFlags');
            const fullyAuthenticatedSelectorSpy = vi.spyOn(authSlice, 'fullyAuthenticatedSelector');
            // hard code user as fully authenticated
            fullyAuthenticatedSelectorSpy.mockReturnValue(true);

            await setup();

            expect(featureFlagSpy).toHaveBeenCalled();
        });
    });
});
