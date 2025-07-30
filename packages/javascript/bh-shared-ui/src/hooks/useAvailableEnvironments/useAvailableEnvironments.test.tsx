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
import * as ReactQuery from 'react-query';
import { renderHook, waitFor } from '../../test-utils';
import { useAvailableEnvironments, useSelectedEnvironment } from './useAvailableEnvironments';

const fakeDomainA = {
    type: 'active-directory',
    impactValue: 12,
    name: 'a',
    collected: true,
    id: 'a',
};
const fakeDomainB = {
    type: 'active-directory',
    impactValue: 12,
    name: 'b',
    collected: true,
    id: 'b',
};
const fakeDomains = [fakeDomainA, fakeDomainB];

const server = setupServer(
    rest.get('/api/v2/available-domains', (req, res, ctx) => {
        return res(ctx.json({ data: fakeDomains }));
    })
);

beforeAll(() => server.listen());
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

describe('useAvailableEnvironments', () => {
    describe('useAvailableEnvironments', () => {
        const useQuerySpy = vi.spyOn(ReactQuery, 'useQuery');
        it('takes the param appendQueryKey and will append that value to the existing queryKey', async () => {
            const testKey = 'test-key';
            renderHook(() => useAvailableEnvironments({ queryKey: [testKey] }));
            expect(useQuerySpy).toBeCalledWith(
                expect.objectContaining({ queryKey: ['available-environments', testKey] })
            );
        });
    });

    describe('useSelectedEnvironment', () => {
        it('returns the full environment for the environmentid passed', async () => {
            const actual = renderHook(() => useSelectedEnvironment(fakeDomainA.id));

            await waitFor(() => expect(actual.result.current.data).toEqual(fakeDomainA));
        });
        it('returns the full environment of the environmentId found in the search params', async () => {
            const url = `/test?environmentId=${fakeDomainA.id}`;
            const actual = renderHook(() => useSelectedEnvironment(fakeDomainA.id), { route: url });

            await waitFor(() => expect(actual.result.current.data).toEqual(fakeDomainA));
        });
        it('returns undefined if environmentId searched for is not found', async () => {
            const fakeEnvId = 'nonexistent';
            const actual = renderHook(() => useSelectedEnvironment(fakeEnvId));

            await waitFor(() => expect(actual.result.current.data).toBeUndefined());

            const url = `/test?environmentId=${fakeEnvId}`;
            const actual2 = renderHook(() => useSelectedEnvironment(), { route: url });

            await waitFor(() => expect(actual2.result.current.data).toBeUndefined());
        });
    });
});
