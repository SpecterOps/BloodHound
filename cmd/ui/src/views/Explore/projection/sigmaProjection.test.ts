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

import { buildProjection, deriveProjectionPresentation, discoverLeafGroups, normalizeSourceGraph } from 'bh-shared-ui';
import { MultiDirectedGraph } from 'graphology';
import type { GraphNode, GraphResponse } from 'js-client-library';
import { GlyphLocation } from 'src/rendering/programs/node.glyphs';
import { graphBounds, sigmaMemberGlyphs, updateSigmaProjection } from './sigmaProjection';

const makeGraphNode = (kind: string): GraphNode => ({
    kind,
    kinds: [kind],
    label: kind,
    objectId: kind,
    lastSeen: '',
    isTierZero: false,
    isOwnedObject: false,
});

it('applies only the visible projection while retaining original source coordinates in renderer storage', () => {
    const source = normalizeSourceGraph({
        data: {
            nodes: { a: makeGraphNode('User'), b: makeGraphNode('User'), t: makeGraphNode('Group') },
            edges: [
                { id: 1, source: 'a', target: 't', kind: 'MemberOf', label: 'MemberOf', lastSeen: '' },
                { id: 2, source: 'b', target: 't', kind: 'MemberOf', label: 'MemberOf', lastSeen: '' },
            ],
        },
    } as GraphResponse);
    const styles = new MultiDirectedGraph();
    for (const [i, id] of ['a', 'b', 't'].entries())
        styles.addNode(id, {
            x: i * 20,
            y: 10,
            kind: source.nodes[id].kind,
            glyphs: [],
            borderColor: '#000',
            backgroundColor: '#fff',
        });
    styles.addEdgeWithKey('rel_1', 'a', 't', { kind: 'MemberOf', label: 'MemberOf' });
    styles.addEdgeWithKey('rel_2', 'b', 't', { kind: 'MemberOf', label: 'MemberOf' });
    const renderer = styles.copy();
    const groups = discoverLeafGroups(source);
    const projected = buildProjection(source, groups, { expandedGroupIds: new Set() }, []);
    const result = updateSigmaProjection(
        renderer,
        styles,
        source,
        projected,
        deriveProjectionPresentation(source, projected, {}, new Set()),
        groups
    );
    const representative = Object.keys(result.groupIds).find((id) => renderer.hasNode(id))!;
    expect(renderer.getNodeAttributes(representative)).toMatchObject({ label: '2 users', hidden: false, x: 10, y: 10 });
    expect(renderer.getNodeAttribute('a', 'hidden')).toBe(true);
    renderer.setNodeAttribute(representative, 'x', 99);
    const expanded = buildProjection(
        source,
        groups,
        { expandedGroupIds: new Set(groups.map((group) => group.id)) },
        []
    );
    updateSigmaProjection(
        renderer,
        styles,
        source,
        expanded,
        deriveProjectionPresentation(source, expanded, {}, new Set()),
        groups
    );
    expect(renderer.getNodeAttribute('a', 'x')).toBe(0);
    expect(renderer.getNodeAttribute(representative, 'hidden')).toBe(true);
    updateSigmaProjection(
        renderer,
        styles,
        source,
        projected,
        deriveProjectionPresentation(source, projected, {}, new Set()),
        groups
    );
    expect(renderer.getNodeAttribute(representative, 'x')).toBe(99);
});

it('unions same-corner glyphs without a fixed slot cap', () => {
    const source = normalizeSourceGraph({
        data: {
            nodes: { a: makeGraphNode('User'), b: makeGraphNode('User'), t: makeGraphNode('Group') },
            edges: [
                { id: 1, source: 'a', target: 't', kind: 'MemberOf', label: 'MemberOf', lastSeen: '' },
                { id: 2, source: 'b', target: 't', kind: 'MemberOf', label: 'MemberOf', lastSeen: '' },
            ],
        },
    } as GraphResponse);
    const styles = new MultiDirectedGraph();
    for (const id of ['a', 'b', 't'])
        styles.addNode(id, {
            x: 0,
            y: 0,
            kind: source.nodes[id].kind,
            borderColor: '#000',
            backgroundColor: '#fff',
            glyphs:
                id === 't'
                    ? []
                    : ['z1', 'z2', 'z3', 'z4', 'z5'].map((image) => ({
                          image,
                          location: GlyphLocation.TOP_RIGHT,
                          color: '#fff',
                          backgroundColor: '#000',
                      })),
        });
    styles.addEdgeWithKey('rel_1', 'a', 't', {});
    styles.addEdgeWithKey('rel_2', 'b', 't', {});
    const groups = discoverLeafGroups(source);
    const graph = buildProjection(source, groups, { expandedGroupIds: new Set() }, []);
    const renderer = styles.copy();
    const result = updateSigmaProjection(
        renderer,
        styles,
        source,
        graph,
        deriveProjectionPresentation(source, graph, sigmaMemberGlyphs(styles), new Set()),
        groups
    );
    const id = Object.keys(result.groupIds).find((id) => renderer.hasNode(id))!;
    const glyphs = renderer.getNodeAttribute(id, 'glyphs');
    expect(glyphs).toHaveLength(6);
    expect(glyphs.slice(0, 5).every((glyph: { location: number }) => glyph.location === GlyphLocation.TOP_RIGHT)).toBe(
        true
    );
    expect(new Set(glyphs.slice(0, 5).map((glyph: { distance: number }) => glyph.distance)).size).toBe(5);
    expect(renderer.getAttribute('glyphCapacity')).toBe(6);
});

it('fits visible entities while ignoring hidden renderer caches', () => {
    const graph = new MultiDirectedGraph();
    graph.addNode('visible', { x: 10, y: 20 });
    graph.addNode('hidden', { x: -1000, y: 2000, hidden: true });
    expect(graphBounds(graph, true)).toEqual({ x: [10, 10], y: [20, 20] });
});

it('keeps semantically distinct redirected relationships individually visible as parallel curves', () => {
    const source = normalizeSourceGraph({
        data: {
            nodes: { a: makeGraphNode('User'), b: makeGraphNode('User'), t: makeGraphNode('Group') },
            edges: [
                {
                    id: 1,
                    source: 'a',
                    target: 't',
                    kind: 'MemberOf',
                    label: 'MemberOf',
                    lastSeen: '',
                    impactPercent: 10,
                },
                {
                    id: 2,
                    source: 'b',
                    target: 't',
                    kind: 'MemberOf',
                    label: 'MemberOf',
                    lastSeen: '',
                    impactPercent: 20,
                },
            ],
        },
    } as GraphResponse);
    const styles = new MultiDirectedGraph();
    for (const id of ['a', 'b', 't'])
        styles.addNode(id, { x: 0, y: 0, kind: source.nodes[id].kind, borderColor: '#000', backgroundColor: '#fff' });
    styles.addEdgeWithKey('rel_1', 'a', 't', { label: 'MemberOf', type: 'arrow' });
    styles.addEdgeWithKey('rel_2', 'b', 't', { label: 'MemberOf', type: 'arrow' });
    const groups = discoverLeafGroups(source);
    const graph = buildProjection(source, groups, { expandedGroupIds: new Set() }, []);
    const renderer = styles.copy();
    updateSigmaProjection(
        renderer,
        styles,
        source,
        graph,
        deriveProjectionPresentation(source, graph, {}, new Set()),
        groups
    );
    const visible = renderer.filterEdges((_id, edge) => !edge.hidden);
    expect(visible).toHaveLength(2);
    expect(visible.map((id) => renderer.getEdgeAttribute(id, 'type'))).toEqual(['curved', 'curved']);
    expect(new Set(visible.map((id) => renderer.getEdgeAttribute(id, 'groupPosition'))).size).toBe(2);
});
