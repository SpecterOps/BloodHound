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

import { ReactNode } from 'react';
import { QueryClient, QueryClientProvider } from 'react-query';
import { BrowserRouter } from 'react-router-dom';
import { renderHook, waitFor } from '../../test-utils';
import { usePathfindingSearch } from './usePathfindingSearch';

const TEST_STRING_1 = 'Test1';
const TEST_STRING_2 = 'test2';

const queryClient = new QueryClient();

const wrapper = ({ children }: { children: ReactNode }) => (
    <QueryClientProvider client={queryClient}>
        <BrowserRouter>{children}</BrowserRouter>
    </QueryClientProvider>
);

// Skipping these for now since we will be bringing in the history package in another PR to make testing query param changes easier
describe.skip('usePathfindingSearch', () => {
    it('stores the state of source and destination terms without modifying query params', async () => {
        const hook = renderHook(() => usePathfindingSearch(), { wrapper });

        expect(window.location.search).toBe('');

        await waitFor(() => hook.result.current.handleSourceNodeEdited(TEST_STRING_1));
        await waitFor(() => hook.result.current.handleDestinationNodeEdited(TEST_STRING_2));

        expect(hook.result.current.sourceSearchTerm).toBe(TEST_STRING_1);
        expect(hook.result.current.destinationSearchTerm).toBe(TEST_STRING_2);
        expect(window.location.search).toBe('');
    });
    it("upon selecting just a source node, updates the URL with a searchType of 'node' and primarySearch of the node's objectid", async () => {
        const hook = renderHook(() => usePathfindingSearch(), { wrapper });

        await waitFor(() =>
            hook.result.current.handleSourceNodeSelected({ name: TEST_STRING_1, objectid: TEST_STRING_2 })
        );

        expect(window.location.search).toContain(`primarySearch=${TEST_STRING_2}`);
        expect(window.location.search).toContain('searchType=node');
    });
    it("upon selecting just a destinations node, updates the URL with a searchType of 'node' and secondarySearch of the node's objectid", async () => {
        const hook = renderHook(() => usePathfindingSearch(), { wrapper });

        await waitFor(() =>
            hook.result.current.handleDestinationNodeSelected({ name: TEST_STRING_1, objectid: TEST_STRING_2 })
        );

        expect(window.location.search).toContain(`secondarySearch=${TEST_STRING_2}`);
        expect(window.location.search).toContain('searchType=node');
    });
});
