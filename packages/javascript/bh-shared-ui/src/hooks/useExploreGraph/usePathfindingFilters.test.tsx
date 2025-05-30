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

import { act, renderHook } from '../../test-utils';
import { INITIAL_FILTERS } from './queries';
import { usePathfindingFilters } from './usePathfindingFilters';
import { extractEdgeTypes } from './utils';

const TEST_FILTER = INITIAL_FILTERS[0];

describe('usePathfindingFilters', () => {
    it('initializes the list with all filters checked by default', () => {
        const hook = renderHook(() => usePathfindingFilters());
        expect(hook.result.current.selectedFilters).toEqual(INITIAL_FILTERS);
    });

    it('will update the selected filters based on the values stored in query params when the initialize function is called', async () => {
        const url = `?pathFilters=${TEST_FILTER.edgeType}`;
        const hook = renderHook(() => usePathfindingFilters(), { route: url });

        expect(hook.result.current.selectedFilters).toEqual(INITIAL_FILTERS);

        await act(() => hook.result.current.initialize());

        const edgeTypesInFilter = extractEdgeTypes(hook.result.current.selectedFilters);
        expect(edgeTypesInFilter).toEqual([TEST_FILTER.edgeType]);
    });

    it('allows you to update the list of selected filters, only updating the url after calling the apply function', async () => {
        const hook = renderHook(() => usePathfindingFilters());

        await act(() => hook.result.current.handleUpdateFilters([TEST_FILTER]));

        expect(hook.result.current.selectedFilters).toEqual([TEST_FILTER]);
        expect(window.location.search).toEqual('');

        await act(() => hook.result.current.handleApplyFilters());

        expect(hook.result.current.selectedFilters).toEqual([TEST_FILTER]);
        expect(window.location.search).toEqual(`?pathFilters=${TEST_FILTER.edgeType}`);
    });

    it('allows you to cancel filter updates before applying them to the url and reset to the default filter state', async () => {
        const hook = renderHook(() => usePathfindingFilters());

        await act(() => hook.result.current.handleUpdateFilters([TEST_FILTER]));
        await act(() => hook.result.current.handleCancelFilters());

        expect(hook.result.current.selectedFilters).toEqual(INITIAL_FILTERS);
        expect(window.location.search).toEqual('');
    });
});
