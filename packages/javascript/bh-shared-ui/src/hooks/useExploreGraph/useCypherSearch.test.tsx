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

import { act, renderHook, waitFor } from '../../test-utils';
import { useCypherSearch } from './useCypherSearch';

const TEST_CYPHER = 'match (n) return n limit 10';
const TEST_BASE64 = 'bWF0Y2ggKG4pIHJldHVybiBuIGxpbWl0IDEw';

describe('useCypherSearch', () => {
    it('stores the state of a search term without modifying the query params', () => {
        const hook = renderHook(() => useCypherSearch());

        expect(window.location.search).toBe('');

        act(() => hook.result.current.setCypherQuery(TEST_CYPHER));

        expect(hook.result.current.cypherQuery).toBe(TEST_CYPHER);
        expect(window.location.search).toBe('');
    });

    it("upon performing a search, updates the url params with the base64-encoded current query and sets searchType to 'cypher'", async () => {
        const hook = renderHook(() => useCypherSearch());

        await act(async () => hook.result.current.setCypherQuery(TEST_CYPHER));
        await act(async () => hook.result.current.performSearch());

        expect(window.location.search).toContain('searchType=cypher');
        expect(window.location.search).toContain(`cypherSearch=${TEST_BASE64}`);
    });

    it('optionally allows the consumer to pass a query to performSearch() and adds that query to the url', async () => {
        const hook = renderHook(() => useCypherSearch());

        await act(async () => hook.result.current.performSearch(TEST_CYPHER));

        expect(window.location.search).toContain('searchType=cypher');
        expect(window.location.search).toContain(`cypherSearch=${TEST_BASE64}`);
    });

    it('populates the cypher search field with the decoded query when the associated query params are set', async () => {
        const url = `?searchType=cypher&cypherSearch=${TEST_BASE64}`;
        const hook = renderHook(() => useCypherSearch(), { route: url });

        await waitFor(() => expect(hook.result.current.cypherQuery).toEqual(TEST_CYPHER));
    });
});
