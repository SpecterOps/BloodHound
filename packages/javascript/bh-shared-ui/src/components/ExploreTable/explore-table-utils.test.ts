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
import { cypherTestResponse } from '../../mocks';
import {
    compareForExploreTableSort,
    DUPLICATED_KNOWN_KEYS,
    getExploreTableData,
    KNOWN_NODE_KEYS,
} from './explore-table-utils';

describe('Compare function for explore table sort', () => {
    test('function should return 1 when first param is larger, no matter the data type', () => {
        const FIRST_PARAM_IS_LARGER = 1;
        expect(compareForExploreTableSort(6, 5)).toBe(FIRST_PARAM_IS_LARGER);
        expect(compareForExploreTableSort('6', '5')).toBe(FIRST_PARAM_IS_LARGER);
        expect(compareForExploreTableSort('6', 5)).toBe(FIRST_PARAM_IS_LARGER);
        expect(compareForExploreTableSort(6, '5')).toBe(FIRST_PARAM_IS_LARGER);
        expect(compareForExploreTableSort(true, false)).toBe(FIRST_PARAM_IS_LARGER);
        expect(compareForExploreTableSort(true, undefined)).toBe(FIRST_PARAM_IS_LARGER);
        expect(compareForExploreTableSort(true, null)).toBe(FIRST_PARAM_IS_LARGER);
        expect(compareForExploreTableSort('banana', 'apple')).toBe(FIRST_PARAM_IS_LARGER);
        expect(compareForExploreTableSort('apple', 3)).toBe(FIRST_PARAM_IS_LARGER);
        expect(compareForExploreTableSort('apple', undefined)).toBe(FIRST_PARAM_IS_LARGER);
        expect(compareForExploreTableSort('apple', null)).toBe(FIRST_PARAM_IS_LARGER);
        expect(compareForExploreTableSort(true, 1)).toBe(FIRST_PARAM_IS_LARGER);
        +expect(compareForExploreTableSort(0, false)).toBe(FIRST_PARAM_IS_LARGER);

        const SECOND_PARAM_IS_LARGER = -1;

        expect(compareForExploreTableSort(5, 6)).toBe(SECOND_PARAM_IS_LARGER);
        expect(compareForExploreTableSort('5', '6')).toBe(SECOND_PARAM_IS_LARGER);
        expect(compareForExploreTableSort('5', 6)).toBe(SECOND_PARAM_IS_LARGER);
        expect(compareForExploreTableSort(false, true)).toBe(SECOND_PARAM_IS_LARGER);
        expect(compareForExploreTableSort(undefined, true)).toBe(SECOND_PARAM_IS_LARGER);
        expect(compareForExploreTableSort(null, true)).toBe(SECOND_PARAM_IS_LARGER);
        expect(compareForExploreTableSort('apple', 'banana')).toBe(SECOND_PARAM_IS_LARGER);
        expect(compareForExploreTableSort(3, 'apple')).toBe(SECOND_PARAM_IS_LARGER);
        expect(compareForExploreTableSort(undefined, 'apple')).toBe(SECOND_PARAM_IS_LARGER);
        expect(compareForExploreTableSort(null, 'apple')).toBe(SECOND_PARAM_IS_LARGER);
        +expect(compareForExploreTableSort('', 'a')).toBe(SECOND_PARAM_IS_LARGER);
        +expect(compareForExploreTableSort(0, 1)).toBe(SECOND_PARAM_IS_LARGER);
        +expect(compareForExploreTableSort(-1, 0)).toBe(SECOND_PARAM_IS_LARGER);
        +expect(compareForExploreTableSort(1, true)).toBe(SECOND_PARAM_IS_LARGER);

        const VALUES_ARE_EQUAL = 0;
        expect(compareForExploreTableSort(5, 5)).toBe(VALUES_ARE_EQUAL);
        expect(compareForExploreTableSort('apple', 'apple')).toBe(VALUES_ARE_EQUAL);
        expect(compareForExploreTableSort(true, true)).toBe(VALUES_ARE_EQUAL);
        expect(compareForExploreTableSort(false, false)).toBe(VALUES_ARE_EQUAL);
        expect(compareForExploreTableSort(null, null)).toBe(VALUES_ARE_EQUAL);
        expect(compareForExploreTableSort(undefined, undefined)).toBe(VALUES_ARE_EQUAL);
    });
});

describe('getExploreTableData', () => {
    it.each([undefined, null, [], {}])('returns undefined if graphData is %s or not a GraphResponse', (graphData) => {
        const expected = undefined;
        const actual = getExploreTableData(graphData as any);
        expect(actual).toBe(expected);
    });
    it('returns the same nodes from the graphData', () => {
        const expected = cypherTestResponse.data.nodes;
        const actual = getExploreTableData(cypherTestResponse);

        expect(actual?.nodes).toEqual(expected);
    });
    it('returns the node_keys from the graphData with duplicate keys removed and known keys added', () => {
        const expected = cypherTestResponse.data.node_keys;
        const actual = getExploreTableData(cypherTestResponse);

        expect(actual?.node_keys).not.toEqual(expect.arrayContaining(DUPLICATED_KNOWN_KEYS));
        expect(actual?.node_keys).toEqual(expect.arrayContaining(KNOWN_NODE_KEYS));
        expect(actual?.node_keys).toEqual(
            expect.arrayContaining(expected.filter((key) => !DUPLICATED_KNOWN_KEYS.includes(key)))
        );
    });
});
