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

import { rest } from 'msw';
import { setupServer } from 'msw/node';
import { renderHook, waitFor } from '../../test-utils';
import * as useAvailableEnvironmentHooks from '../useAvailableEnvironments';
import { useInitialEnvironment } from './useInitialEnvironment';
const useAvailableEnvironmentsSpy = vi.spyOn(useAvailableEnvironmentHooks, 'useAvailableEnvironments');

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
    it('calls handleInitialEnvironment with the highest impactValue as the initial environment', async () => {
        const mockHandleInitialEnv = vi.fn();
        renderHook(() => useInitialEnvironment({ orderBy: 'name', handleInitialEnvironment: mockHandleInitialEnv }));

        await waitFor(() => expect(mockHandleInitialEnv).toBeCalledWith(fakeEnvironmentA));
    });
    it('disables useAvailableEnvironment and does not attempt calling handleInitialEnvironment if queryOptions.enabled', async () => {
        const mockHandleInitialEnv = vi.fn();
        renderHook(() =>
            useInitialEnvironment({
                orderBy: 'name',
                handleInitialEnvironment: mockHandleInitialEnv,
                queryOptions: { enabled: false },
            })
        );

        expect(useAvailableEnvironmentsSpy).toBeCalledWith(expect.objectContaining({ enabled: false }));
        await waitFor(() => expect(mockHandleInitialEnv).not.toBeCalled());
    });
});
