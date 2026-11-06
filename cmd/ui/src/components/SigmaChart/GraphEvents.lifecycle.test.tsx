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
import { Attributes } from 'graphology-types';
import { useEffect, useState } from 'react';
import { render } from 'src/test-utils';
import { GraphEvents } from './GraphEvents';

const { getSigma, registerEvents } = vi.hoisted(() => ({ getSigma: vi.fn(), registerEvents: vi.fn() }));
vi.mock('@react-sigma/core', () => ({
    useRegisterEvents: () => registerEvents,
    useSigma: () => getSigma(),
    // Match useSetSettings retaining settings while its target Sigma instance changes.
    useSetSettings: () => {
        const sigma = getSigma();
        const [settings, setSettings] = useState<Record<string, unknown>>({});
        useEffect(() => {
            Object.entries(settings).forEach(([key, value]) => sigma.setSetting(key, value));
        }, [sigma, settings]);
        return setSettings;
    },
}));
vi.mock('src/ducks/graph/utils', async (importOriginal) => ({
    ...(await importOriginal<object>()),
    resetCamera: vi.fn(),
}));

it.each(['curved', 'self'])('renders %s edges after Sigma instance replacement', (type) => {
    const results: Attributes[] = [];
    const makeSigma = (graph: MultiDirectedGraph) => {
        const coordinates = (point: { x: number; y: number }) => point;
        return {
            getGraph: () => graph,
            setCustomBBox: vi.fn(),
            getCamera: () => ({ ratio: 1 }),
            getNodeDisplayData: vi.fn((id: string) => graph.getNodeAttributes(id)),
            framedGraphToViewport: coordinates,
            graphToViewport: coordinates,
            viewportToFramedGraph: coordinates,
            // Native installation immediately processes valid edges in the receiving Sigma.
            setSetting: vi.fn((key: string, value: unknown) => {
                if (key === 'edgeReducer' && graph.hasEdge('rel_4')) {
                    results.push(
                        (value as (id: string, data: Attributes) => Attributes)('rel_4', {
                            type,
                            color: '#55595c',
                            groupSize: 2,
                            groupPosition: 0,
                            direction: 1,
                        })
                    );
                }
            }),
        };
    };
    getSigma.mockReturnValue(makeSigma(new MultiDirectedGraph()));
    const view = render(<GraphEvents highlightedItem={null} />);
    view.rerender(<GraphEvents highlightedItem='source' />);
    const graph = new MultiDirectedGraph();
    graph.addNode('source', { x: 0, y: 0, size: 10 });
    graph.addNode('target', { x: 100, y: 100, size: 10 });
    graph.addEdgeWithKey('rel_4', 'source', type === 'self' ? 'source' : 'target');
    const sigma = makeSigma(graph);
    getSigma.mockReturnValue(sigma);
    view.rerender(<GraphEvents highlightedItem='source' />);

    expect(sigma.getNodeDisplayData).toHaveBeenCalledWith('source');
    expect(results.length).toBeGreaterThan(0);
    for (const result of results) {
        if (type === 'curved') {
            expect(sigma.getNodeDisplayData).toHaveBeenCalledWith('target');
            expect(Number.isFinite(result.control.x)).toBe(true);
            expect(Number.isFinite(result.control.y)).toBe(true);
        } else {
            expect(result.framedGraphNodeRadius).toBeGreaterThan(0);
        }
    }
});
