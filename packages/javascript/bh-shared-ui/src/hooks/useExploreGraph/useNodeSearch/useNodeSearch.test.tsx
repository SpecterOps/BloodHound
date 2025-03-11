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

import { act, renderHook } from '@testing-library/react';
import { ReactNode } from 'react';
import { QueryClient, QueryClientProvider } from 'react-query';
import { BrowserRouter } from 'react-router-dom';
import { useNodeSearch } from './useNodeSearch';

const TEST_STRING_1 = 'Test1';
const TEST_STRING_2 = 'test2';

const queryClient = new QueryClient();

const wrapper = ({ children }: { children: ReactNode }) => (
    <QueryClientProvider client={queryClient}>
        <BrowserRouter>{children}</BrowserRouter>
    </QueryClientProvider>
);

describe('useNodeSearch', () => {
    it('stores the state of a search term without modifying the query params', () => {
        const hook = renderHook(() => useNodeSearch(), { wrapper });

        expect(window.location.search).toBe('');

        act(() => hook.result.current.editSourceNode(TEST_STRING_1));

        expect(hook.result.current.searchTerm).toBe(TEST_STRING_1);
        expect(window.location.search).toBe('');
    });

    it("upon selecting a source node, updates the URL with a searchType of 'node' and primarySearch of the node's objectid", () => {
        const hook = renderHook(() => useNodeSearch(), { wrapper });

        act(() => hook.result.current.selectSourceNode({ name: TEST_STRING_1, objectid: TEST_STRING_2 }));

        expect(window.location.search).toContain(`primarySearch=${TEST_STRING_2}`);
        expect(window.location.search).toContain('searchType=node');
    });

    it('does not add a query param if the search term is empty', () => {
        const hook = renderHook(() => useNodeSearch(), { wrapper });

        act(() => hook.result.current.selectSourceNode({ name: '', objectid: '' }));

        expect(window.location.search).not.toContain('primarySearch');
    });
});
