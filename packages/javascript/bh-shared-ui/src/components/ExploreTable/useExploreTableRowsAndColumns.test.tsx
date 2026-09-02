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
import { renderHook } from '../../test-utils';
import useExploreTableRowsAndColumns from './useExploreTableRowsAndColumns';

describe('useExploreTableRowsAndColumns', () => {
    describe('bhGraphId', () => {
        it('is not overwritten when node properties contain an id field', () => {
            const graphNodeKey = '100';
            const nodeIdProperty = 'property-level-id-value';

            const exploreTableData = {
                nodes: {
                    [graphNodeKey]: {
                        label: 'TestNode',
                        kind: 'User',
                        kinds: ['User'],
                        objectId: 'obj-123',
                        lastSeen: '2026-01-01',
                        isTierZero: false,
                        isOwnedObject: false,
                        properties: { id: nodeIdProperty },
                    },
                },
                node_keys: ['label', 'kind', 'objectId', 'isTierZero', 'isOwnedObject', 'lastSeen'],
            };

            const { result } = renderHook(() =>
                useExploreTableRowsAndColumns({
                    onKebabMenuClick: vi.fn(),
                    searchInput: '',
                    selectedColumns: {},
                    exploreTableData,
                })
            );

            const row = result.current.rows[0];

            expect(row.bhGraphId).toBe(graphNodeKey);
            expect(row.bhGraphId).not.toBe(nodeIdProperty);
        });
    });
});
