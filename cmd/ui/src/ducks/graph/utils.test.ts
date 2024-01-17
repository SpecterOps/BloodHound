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

import { getEdgeDataFromKey } from 'src/ducks/graph/utils';

describe('using the edge key to get its source, target, and edge labels', () => {
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
