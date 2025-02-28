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

import { DeepPartial } from 'bh-shared-ui';
import { createMemoryHistory } from 'history';
import { Domain } from 'js-client-library';
import { rest } from 'msw';
import { setupServer } from 'msw/node';
import * as authSlice from 'src/ducks/auth/authSlice';
import * as globalActions from 'src/ducks/global/actions';
import { AppState } from 'src/store';
import { renderHook, waitFor } from 'src/test-utils';
import { useInitialEnvironment } from './useInitialEnvironment';

const setDomainSpy = vi.spyOn(globalActions, 'setDomain');

const fakeEnvironmentA = {
    type: 'active-directory',
    impactValue: 10,
    name: 'a',
    collected: true,
    id: 'a',
};
const fakeEnvironmentB = {
    type: 'active-directory',
    impactValue: 5,
    name: 'b',
    collected: true,
    id: 'b',
};
const fakeEnvironments = [fakeEnvironmentA, fakeEnvironmentB];

const backButtonSupportFF = {
    key: 'back_button_support',
    enabled: true,
};

const server = setupServer(
    rest.get('/api/v2/available-domains', (req, res, ctx) => {
        return res(ctx.json({ data: fakeEnvironments }));
    }),
    rest.get('/api/v2/features', (req, res, ctx) => {
        return res(
            ctx.json({
                data: [backButtonSupportFF],
            })
        );
    })
);

beforeAll(() => server.listen());
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

describe('useInitialEnvironment', async () => {
    const envSupportedPage = '/graphview';
    const setup = (initialEntry?: string, initialState?: DeepPartial<AppState>) => {
        const fullyAuthenticatedSelectorSpy = vi.spyOn(authSlice, 'fullyAuthenticatedSelector');
        // hard code user as fully authenticated
        fullyAuthenticatedSelectorSpy.mockReturnValue(true);

        const history = createMemoryHistory({ initialEntries: [initialEntry ?? envSupportedPage] });

        renderHook(() => useInitialEnvironment([envSupportedPage]), {
            history,
            initialState: {
                auth: {
                    ...authSlice.initialState,
                    isInitialized: true,
                },
                ...initialState,
            },
        });

        return { history };
    };
    it('attempts to set the environmentId search param with the highest impactValue as the initial environment', async () => {
        const { history } = setup();

        await waitFor(() => expect(history.location.search).toContain(`environmentId=${fakeEnvironmentA.id}`));
    });
    it('sets redux global.info.domain if back_button_support feature flag is disabled', async () => {
        server.use(
            rest.get(`/api/v2/features`, (req, res, ctx) => {
                return res(ctx.json({ ...backButtonSupportFF, enabled: false }));
            })
        );

        setup();

        await waitFor(() => expect(setDomainSpy).toBeCalledWith(fakeEnvironmentA));
    });
    it('does not set initial environment if current route is not an env supported route', async () => {
        const { history } = setup('/not-graphview');

        await waitFor(() => expect(history.location.search).toBe(''));
    });
    it('does not attempt setting a default environment if environmentId is set already', async () => {
        const { history } = setup(`?environmentId=${fakeEnvironmentB.id}`);

        await waitFor(() => expect(history.location.search).toContain(`environmentId=${fakeEnvironmentB.id}`));
    });
    it('not attempt setting a default environment if domain.id is set already', async () => {
        setup(envSupportedPage, { global: { options: { domain: fakeEnvironmentB as Domain } } });

        await waitFor(() => expect(setDomainSpy).not.toBeCalledWith(fakeEnvironmentA));
    });
});
