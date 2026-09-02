// Copyright 2023 Specter Ops, Inc.
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

import { getDistanceBetween, getEdgeDataFromKey, getNodeOffset } from 'src/ducks/graph/utils';

describe('getEdgeDataFromKey', () => {
    it('should return null if the edge key does not have enough information', () => {
        console.warn = vi.fn();
        expect(getEdgeDataFromKey('')).toBeNull();
        expect(getEdgeDataFromKey('invalid')).toBeNull();
        expect(getEdgeDataFromKey('still_invalid')).toBeNull();
    });

    it('should return the edge parts from the key', () => {
        expect(getEdgeDataFromKey('a_memberOf_b')).toEqual({ source: 'a', target: 'b', label: 'memberOf' });
    });

    it('should appropriately handle edges that contain underscores', () => {
        expect(getEdgeDataFromKey('a_AZMGGroup_ReadWrite_All_b')).toEqual({
            source: 'a',
            target: 'b',
            label: 'AZMGGroup_ReadWrite_All',
        });
    });
});

describe('getDistanceBetween', () => {
    it('returns the distance between two coordinates', () => {
        expect(getDistanceBetween({ x: 3, y: 0 }, { x: 0, y: 4 })).toEqual(5);
    });
});

describe('getNodeOffset', () => {
    it('returns the difference between the node and the position', () => {
        expect(getNodeOffset({ x: 2, y: 1 }, { x: 3, y: 2 })).toEqual({ x: 1, y: 1 });
    });
});
