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
import { useNodeSearch } from './useNodeSearch';

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

describe('useNodeSearch', () => {
    it.skip('stores the state of a search term without modifying the query params', async () => {
        const history = createMemoryHistory();
        const hook = renderHook(() => useNodeSearch(), { history });

        expect(history.location.search).toBe('');

        await act(async () => hook.result.current.editSourceNode(TEST_STRING_1));

        expect(hook.result.current.searchTerm).toBe(TEST_STRING_1);
        expect(history.location.search).toBe('');
    });

    it("upon selecting a source node, updates the URL with a searchType of 'node' and primarySearch of the node's objectid", async () => {
        const history = createMemoryHistory();
        const hook = renderHook(() => useNodeSearch(), { history });

        await act(async () => hook.result.current.selectSourceNode({ name: TEST_STRING_1, objectid: TEST_STRING_2 }));

        expect(history.location.search).toContain(`primarySearch=${TEST_STRING_2}`);
        expect(history.location.search).toContain('searchType=node');
    });

    it('does not add a query param if the search term is empty', async () => {
        const history = createMemoryHistory();
        const hook = renderHook(() => useNodeSearch(), { history });

        await act(async () => hook.result.current.selectSourceNode({ name: '', objectid: '' }));

        expect(history.location.search).not.toContain('primarySearch');
    });

    it("populates the node search field when the 'primarySearch' query param is set", async () => {
        const url = `?primarySearch=${TEST_STRING_1}&searchType=node`;
        const history = createMemoryHistory({ initialEntries: [url] });

        const hook = renderHook(() => useNodeSearch(), { history });

        await waitFor(() => expect(hook.result.current.searchTerm).toEqual(TEST_STRING_1));
        await waitFor(() =>
            expect(hook.result.current.selectedItem).toEqual({ name: TEST_STRING_1, objectid: TEST_STRING_1 })
        );
    });
});
