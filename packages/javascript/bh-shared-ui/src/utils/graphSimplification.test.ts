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

import { findRepresentativeNodeSets, formatCompactCount, pluralNodeType } from './graphSimplification';

const edge = (source: string, target = 'target', kind = 'MemberOf') => ({ source, target, kind });
const types = { a: 'User', b: 'User', target: '' };

describe('leaf representative eligibility', () => {
    it('uses stable IDs and sorted real members for a specific neighbor, type, relationship and direction', () => {
        expect(findRepresentativeNodeSets(types, [edge('b'), edge('a')])).toEqual({
            'representative:["User","MemberOf","out","target"]': { members: ['a', 'b'], target: 'target' },
        });
        expect(findRepresentativeNodeSets(types, [edge('a'), edge('target', 'b')])).toEqual({});
        expect(findRepresentativeNodeSets(types, [edge('a'), edge('b', 'target', 'Contains')])).toEqual({});
        expect(findRepresentativeNodeSets({ ...types, b: 'Computer' }, [edge('a'), edge('b')])).toEqual({});
        expect(findRepresentativeNodeSets({ ...types, other: '' }, [edge('a'), edge('b', 'other')])).toEqual({});
    });
    it('does not collapse additional connections, including parallel edges, self edges or missing endpoints/kinds', () => {
        for (const edges of [
            [edge('a'), edge('b'), edge('a', 'target', 'Contains')],
            [edge('a'), edge('b'), edge('a')],
            [edge('a', 'a'), edge('b', 'b')],
            [edge('a', 'missing'), edge('b', 'missing')],
            [edge('a', 'target', ''), edge('b', 'target', '')],
        ])
            expect(findRepresentativeNodeSets(types, edges)).toEqual({});
    });
    it('excludes internal or unknown node types and protects existing IDs', () => {
        for (const type of ['HIDDEN', 'Meta', 'MetaDetail', '']) {
            expect(findRepresentativeNodeSets({ ...types, a: type, b: type }, [edge('a'), edge('b')])).toEqual({});
        }
        expect(
            findRepresentativeNodeSets({ ...types, 'representative:["User","MemberOf","out","target"]': 'Group' }, [
                edge('a'),
                edge('b'),
            ])
        ).toEqual({});
    });
});

it('formats natural plurals and compact counts consistently', () => {
    expect(['User', 'AZServicePrincipal', 'Policy', 'Person'].map(pluralNodeType)).toEqual([
        'users',
        'service principals',
        'policies',
        'people',
    ]);
    expect([5, 10100, 300000].map(formatCompactCount)).toEqual(['5', '10.1k', '300k']);
});
