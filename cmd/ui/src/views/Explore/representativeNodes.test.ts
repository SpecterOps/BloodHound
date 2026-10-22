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

import { MultiDirectedGraph } from 'graphology';
import { GlyphLocation } from 'src/rendering/programs/node.glyphs';
import { applyRepresentativeNodes, countGlyph, representativeGlyphs } from './representativeNodes';

const sets = { representative: { members: ['a', 'b'], target: 'group' } };
const glyph = (image: string, location = GlyphLocation.TOP_RIGHT) => ({
    image,
    location,
    backgroundColor: '#fff',
    color: '#000',
});
const build = () => {
    const graph = new MultiDirectedGraph();
    graph.addNode('group', { x: 0, y: 0 });
    for (const [index, id] of ['a', 'b'].entries()) {
        graph.addNode(id, {
            kind: 'User',
            x: 10 + index,
            y: 20,
            borderColor: '#000',
            backgroundColor: '#fff',
            glyphs: [glyph('zone'), glyph(id, GlyphLocation.BOTTOM_RIGHT)],
        });
        graph.addEdgeWithKey(`edge-${id}`, id, 'group', { label: 'MemberOf', kind: 'MemberOf', dashed: true });
    }
    return graph;
};

it('projects leaves with ordinary relationship labels, the complete glyph union and a separate count', () => {
    const graph = build();
    const original = graph.export();
    applyRepresentativeNodes(graph, sets, new Set(), null);
    expect(graph.getNodeAttribute('representative', 'label')).toBe('2 users');
    expect(graph.getNodeAttribute('representative', 'glyphs')).toHaveLength(4);
    expect(graph.getNodeAttribute('a', 'hidden')).toBe(true);
    expect(graph.getEdgeAttribute('edge-a', 'hidden')).toBe(true);
    const displayed = graph.edges('representative');
    expect(displayed).toHaveLength(1);
    expect(graph.getEdgeAttributes(displayed[0])).toMatchObject({
        label: 'MemberOf',
        dashed: true,
        hidden: false,
        type: 'arrow',
    });
    expect(graph.getEdgeAttributes(displayed[0])).not.toHaveProperty('glyphs');
    applyRepresentativeNodes(graph, sets, new Set(['representative']), null);
    expect(graph.getNodeAttribute('representative', 'hidden')).toBe(true);
    expect(graph.getEdgeAttribute(displayed[0], 'hidden')).toBe(true);
    for (const node of original.nodes) {
        expect(graph.getNodeAttributes(node.key)).toMatchObject(node.attributes!);
        expect(graph.getNodeAttribute(node.key, 'hidden')).toBe(false);
    }
    expect(graph.filterEdges((_id, attributes) => !attributes.hidden)).toEqual(['edge-a', 'edge-b']);
    // Repeated toggles reuse cache identities without multiplying synthetic entities.
    const size = graph.size;
    applyRepresentativeNodes(graph, sets, new Set(), null);
    expect(graph.size).toBe(size);
    expect(graph.getNodeAttribute('representative', 'hidden')).toBe(false);
});

it.each(['a', 'edge-a'])('keeps a selected real entity %s visible during collapse', (selected) => {
    const graph = build();
    applyRepresentativeNodes(graph, sets, new Set(), selected);
    expect(graph.hasNode('representative')).toBe(false);
    expect(graph.getNodeAttribute('a', 'hidden')).toBe(false);
});

it('retains distinct same-corner glyphs without duplicates or a four-slot cap', () => {
    const graph = build();
    graph.setNodeAttribute(
        'a',
        'glyphs',
        ['z1', 'z2', 'z3', 'z4', 'z5'].map((id) => glyph(id))
    );
    graph.setNodeAttribute('b', 'glyphs', [glyph('z1')]);
    const inherited = representativeGlyphs(graph, ['a', 'b']);
    expect(inherited).toHaveLength(5);
    expect(inherited[0].location).toBe(GlyphLocation.TOP_RIGHT);
    expect(new Set(inherited.map((item) => item.distance)).size).toBe(5);
    const badge = countGlyph(10100, '#000', '#fff');
    expect(badge.location).toBe(GlyphLocation.TOP_LEFT);
    expect(decodeURIComponent(badge.image)).toContain('10.1k');
    expect(decodeURIComponent(badge.image)).toContain('font-weight="700"');
});
