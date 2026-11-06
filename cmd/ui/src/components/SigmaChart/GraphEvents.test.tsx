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
import { act, render } from 'src/test-utils';
import { GraphEvents } from './GraphEvents';

const { registerEvents, setSettings, getSigma } = vi.hoisted(() => ({
    registerEvents: vi.fn(),
    setSettings: vi.fn(),
    getSigma: vi.fn(),
}));
vi.mock('@react-sigma/core', () => ({
    useRegisterEvents: () => registerEvents,
    useSetSettings: () => setSettings,
    useSigma: () => getSigma(),
}));
vi.mock('src/ducks/graph/utils', async (importOriginal) => ({
    ...(await importOriginal<object>()),
    resetCamera: vi.fn(),
}));

it('tolerates a dragged representative disappearing before the next pointer event or render', () => {
    const graph = new MultiDirectedGraph();
    graph.addNode('representative', { x: 10, y: 20 });
    getSigma.mockReturnValue({
        getGraph: () => graph,
        setCustomBBox: vi.fn(),
        setSetting: vi.fn<(key: string, value: unknown) => void>(),
        viewportToGraph: () => ({ x: 10, y: 20 }),
    });
    const view = render(<GraphEvents highlightedItem={null} />);
    act(() =>
        registerEvents.mock.calls.at(-1)![0].downNode({
            node: 'representative',
            event: { original: { button: 0 } },
        })
    );
    // Projection mutates the graph in place, so an already registered event can hold stale drag metadata.
    graph.dropNode('representative');
    expect(() => act(() => registerEvents.mock.calls.at(-1)![0].mousemovebody({}))).not.toThrow();
    expect(() => view.rerender(<GraphEvents highlightedItem={null} />)).not.toThrow();
});

it.each(['curved', 'self'])('renders a %s edge from the active graph after Sigma replaces its graph', (type) => {
    let activeGraph = new MultiDirectedGraph();
    const coordinates = (point: { x: number; y: number }) => point;
    const sigma = {
        getGraph: () => activeGraph,
        setCustomBBox: vi.fn(),
        setSetting: vi.fn<(key: string, value: unknown) => void>(),
        getCamera: () => ({ ratio: 1 }),
        getNodeDisplayData: vi.fn((id: string) => activeGraph.getNodeAttributes(id)),
        framedGraphToViewport: coordinates,
        graphToViewport: coordinates,
        viewportToFramedGraph: coordinates,
    };
    getSigma.mockReturnValue(sigma);
    render(<GraphEvents highlightedItem={null} />);
    const reducer = sigma.setSetting.mock.calls.findLast(([key]) => key === 'edgeReducer')![1] as (
        edge: string,
        data: Attributes
    ) => Attributes;

    // Sigma can replace the graph and process its edges before React installs new reducer callbacks.
    activeGraph = new MultiDirectedGraph();
    activeGraph.addNode('source', { x: 0, y: 0, size: 10 });
    activeGraph.addNode('target', { x: 100, y: 100, size: 10 });
    activeGraph.addEdgeWithKey('rel_4', 'source', type === 'self' ? 'source' : 'target');
    const rendered = reducer('rel_4', {
        type,
        color: '#55595c',
        groupSize: 2,
        groupPosition: 0,
        direction: 1,
    });
    expect(sigma.getNodeDisplayData).toHaveBeenCalledWith('source');
    if (type === 'curved') {
        expect(sigma.getNodeDisplayData).toHaveBeenCalledWith('target');
        expect(Number.isFinite(rendered.control.x)).toBe(true);
        expect(Number.isFinite(rendered.control.y)).toBe(true);
    } else {
        expect(rendered.framedGraphNodeRadius).toBeGreaterThan(0);
    }
});
