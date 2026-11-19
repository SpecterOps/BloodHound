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

import type { GraphNode, GraphResponse } from 'js-client-library';
import { buildProjection } from './buildProjection';
import { discoverLeafGroups } from './discoverLeafGroups';
import { normalizeSourceGraph } from './normalizeSourceGraph';
import { deriveProjectionPresentation } from './presentation';

const makeNode = (kind: string): GraphNode => ({
    kind,
    kinds: [kind],
    label: kind,
    objectId: kind,
    lastSeen: '',
    isTierZero: false,
    isOwnedObject: false,
});

it('keeps glyph unions and counts independent from ordinary labels and maps member highlighting', () => {
    const source = normalizeSourceGraph({
        data: {
            nodes: { a: makeNode('User'), b: makeNode('User'), t: makeNode('Group') },
            edges: [
                { id: 1, source: 'a', target: 't', kind: 'MemberOf', label: 'MemberOf', lastSeen: '' },
                { id: 2, source: 'b', target: 't', kind: 'MemberOf', label: 'MemberOf', lastSeen: '' },
            ],
        },
    } as GraphResponse);
    const graph = buildProjection(source, discoverLeafGroups(source), { expandedGroupIds: new Set() }, []);
    const result = deriveProjectionPresentation(
        source,
        graph,
        {
            a: [{ id: 'zone-one', position: 'zone' }],
            b: [
                { id: 'zone-two', position: 'zone' },
                { id: 'owned', position: 'owned' },
            ],
        },
        new Set(['a'])
    );
    expect(result[graph.index.sourceNodeToVisible.a]).toMatchObject({
        caption: '2 users',
        count: 2,
        badge: '2',
        highlighted: true,
        glyphs: [
            { id: 'zone-one', position: 'zone' },
            { id: 'zone-two', position: 'zone' },
            { id: 'owned', position: 'owned' },
        ],
    });
});
