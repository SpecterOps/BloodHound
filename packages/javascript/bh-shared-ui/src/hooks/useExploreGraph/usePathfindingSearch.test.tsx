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

import { createMemoryHistory } from 'history';
import { rest } from 'msw';
import { setupServer } from 'msw/node';
import { act, renderHook, waitFor } from '../../test-utils';
import { usePathfindingSearch } from './usePathfindingSearch';

const TEST_STRING_1 = 'Test1';
const TEST_STRING_2 = 'test2';

const server = setupServer(
    rest.get('/api/v2/search', (req, res, ctx) => {
        const url = new URL(req.url);
        const searchTerm = url.searchParams.get('q');

        return res(ctx.json({ data: [{ name: searchTerm, objectid: searchTerm }] }));
    })
);

beforeAll(() => server.listen());
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

describe('usePathfindingSearch', () => {
    it('stores the state of source and destination terms without modifying query params', async () => {
        const history = createMemoryHistory();
        const hook = renderHook(() => usePathfindingSearch(), { history });

        expect(history.location.search).toBe('');

        await act(async () => hook.result.current.handleSourceNodeEdited(TEST_STRING_1));
        await act(async () => hook.result.current.handleDestinationNodeEdited(TEST_STRING_2));

        expect(hook.result.current.sourceSearchTerm).toBe(TEST_STRING_1);
        expect(hook.result.current.destinationSearchTerm).toBe(TEST_STRING_2);
        expect(history.location.search).toBe('');
    });

    it("upon selecting just a source node, updates the URL with a searchType of 'node' and primarySearch of the node's objectid", async () => {
        const history = createMemoryHistory();
        const hook = renderHook(() => usePathfindingSearch(), { history });

        await act(async () =>
            hook.result.current.handleSourceNodeSelected({ name: TEST_STRING_1, objectid: TEST_STRING_2 })
        );

        expect(history.location.search).toContain(`primarySearch=${TEST_STRING_2}`);
        expect(history.location.search).toContain('searchType=node');
    });

    it("upon selecting just a destination node, updates the URL with a searchType of 'node' and secondarySearch of the node's objectid", async () => {
        const history = createMemoryHistory();
        const hook = renderHook(() => usePathfindingSearch(), { history });

        await act(async () =>
            hook.result.current.handleDestinationNodeSelected({ name: TEST_STRING_1, objectid: TEST_STRING_2 })
        );

        expect(history.location.search).toContain(`secondarySearch=${TEST_STRING_2}`);
        expect(history.location.search).toContain('searchType=node');
    });

    it("upon selecting a source and destination node, updates the URL with a searchType of 'pathfinding' and the object ids for the source and destination nodes", async () => {
        const history = createMemoryHistory();
        const hook = renderHook(() => usePathfindingSearch(), { history });

        await act(async () =>
            hook.result.current.handleSourceNodeSelected({ name: TEST_STRING_1, objectid: TEST_STRING_1 })
        );
        await act(async () =>
            hook.result.current.handleDestinationNodeSelected({ name: TEST_STRING_2, objectid: TEST_STRING_2 })
        );

        expect(history.location.search).toContain(`primarySearch=${TEST_STRING_1}`);
        expect(history.location.search).toContain(`secondarySearch=${TEST_STRING_2}`);
        expect(history.location.search).toContain('searchType=pathfinding');
    });

    it('populates the source and destination node fields when query params are passed', async () => {
        const url = `?primarySearch=${TEST_STRING_1}&secondarySearch=${TEST_STRING_2}&searchType=pathfinding`;
        const history = createMemoryHistory({ initialEntries: [url] });

        const hook = renderHook(() => usePathfindingSearch(), { history });

        await waitFor(() => expect(hook.result.current.sourceSearchTerm).toEqual(TEST_STRING_1));
        await waitFor(() => expect(hook.result.current.destinationSearchTerm).toEqual(TEST_STRING_2));

        await waitFor(() =>
            expect(hook.result.current.sourceSelectedItem).toEqual({ name: TEST_STRING_1, objectid: TEST_STRING_1 })
        );
        await waitFor(() =>
            expect(hook.result.current.destinationSelectedItem).toEqual({
                name: TEST_STRING_2,
                objectid: TEST_STRING_2,
            })
        );
    });

    it('allows a consumer to swap the values of the two inputs and update the query params accordingly', async () => {
        const url = `?primarySearch=${TEST_STRING_1}&secondarySearch=${TEST_STRING_2}&searchType=pathfinding`;
        const history = createMemoryHistory({ initialEntries: [url] });

        const hook = renderHook(() => usePathfindingSearch(), { history });

        await waitFor(() => expect(hook.result.current.sourceSearchTerm).toEqual(TEST_STRING_1));
        await waitFor(() => expect(hook.result.current.destinationSearchTerm).toEqual(TEST_STRING_2));

        await act(async () => hook.result.current.handleSwapPathfindingInputs());

        expect(history.location.search).toContain(`primarySearch=${TEST_STRING_2}`);
        expect(history.location.search).toContain(`secondarySearch=${TEST_STRING_1}`);
        expect(history.location.search).toContain('searchType=pathfinding');
    });
});
