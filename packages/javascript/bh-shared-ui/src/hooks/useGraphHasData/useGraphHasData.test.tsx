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
import { useGraphHasData } from './useGraphHasData';

const populatedPayload = {
    data: {
        nodes: {
            test: {
                isOwnedObject: false,
                isTierZero: false,
                kind: 'Group',
                label: 'DOMAIN USERS@WRAITH.CORP',
                lastSeen: '2025-05-20T19:40:47.175300929Z',
                objectId: 'S-1-5-21-3702535222-3822678775-2090119576-513',
            },
        },
    },
};

const emptyPayload = {
    data: {
        nodes: {},
    },
};
const server = setupServer();

beforeAll(() => server.listen());
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

describe('useGraphHasData', () => {
    it('returns false when no nodes are available', async () => {
        server.use(
            rest.post('/api/v2/graphs/cypher', (req, res, ctx) => {
                return res(ctx.json(emptyPayload));
            })
        );

        const { result } = renderHook(() => useGraphHasData());
        await waitFor(() => expect(result.current.isSuccess).toBe(true));

        expect(result.current.data).toEqual(false);
    });

    it('returns true when nodes are available', async () => {
        server.use(
            rest.post('/api/v2/graphs/cypher', (req, res, ctx) => {
                return res(ctx.json(populatedPayload));
            })
        );

        const { result } = renderHook(() => useGraphHasData());
        await waitFor(() => expect(result.current.isSuccess).toBe(true));

        expect(result.current.data).toEqual(true);
    });

    it('returns false on 404', async () => {
        server.use(
            rest.post('/api/v2/graphs/cypher', (req, res, ctx) => {
                return res(ctx.status(404), ctx.json({}));
            })
        );

        const { result } = renderHook(() => useGraphHasData());
        await waitFor(() => expect(result.current.isLoading).toBe(false));

        expect(result.current.isError).toBe(false);
        expect(result.current.data).toEqual(false);
    });
});
