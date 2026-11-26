// Copyright 2026 Specter Ops, Inc.
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

import { renderHook, waitFor } from '../test-utils';
import { useAutomaticGraphActions } from './useAutomaticGraphActions';
import { useExploreParams } from './useExploreParams';

vi.mock('./useGraphItem/useGraphItem', () => ({
    useGraphItem: vi.fn(),
}));

describe('useAutomaticGraphActions', () => {
    it('replaces the provisional Group Management selection after a single name match', async () => {
        const { result, rerender } = renderHook(
            ({ graphData }: { graphData?: Record<string, unknown> }) => {
                useAutomaticGraphActions(graphData);
                return useExploreParams().selectedItem;
            },
            {
                initialProps: {},
                route: '/explore?searchType=node&primarySearch=Shared+Name&selectedItem=GROUP-OBJECT-ID',
            }
        );

        expect(result.current).toBe('GROUP-OBJECT-ID');
        rerender({ graphData: { 456: { data: { objectid: 'GROUP-OBJECT-ID', name: 'Shared Name' } } } });

        await waitFor(() => expect(result.current).toBe('456'));
    });
});
