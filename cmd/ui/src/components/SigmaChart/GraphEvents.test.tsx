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

import { act, render } from '@testing-library/react';
import { MultiDirectedGraph } from 'graphology';
import { createRef } from 'react';
import { GraphEvents } from './GraphEvents';

const sigmaMocks = vi.hoisted(() => ({
    handlers: {} as Record<string, (event: any) => void>,
    registerEvents: vi.fn((handlers: Record<string, (event: any) => void>) => {
        sigmaMocks.handlers = handlers;
    }),
    setSettings: vi.fn(),
    sigma: undefined as any,
}));

const animationMocks = vi.hoisted(() => ({
    pending: [] as Array<{
        callback?: () => void;
        cancel: ReturnType<typeof vi.fn>;
        cancelled: boolean;
        graph: MultiDirectedGraph;
        targets: Record<string, { x: number; y: number }>;
    }>,
    animateNodes: vi.fn((graph, targets, _options, callback) => {
        const animation = {
            callback,
            cancel: vi.fn(() => {
                animation.cancelled = true;
            }),
            cancelled: false,
            graph,
            targets,
        };
        animationMocks.pending.push(animation);
        return animation.cancel;
    }),
}));

const completeLatestAnimation = () => {
    const animation = animationMocks.pending.at(-1);
    if (!animation || animation.cancelled) return;

    Object.entries(animation.targets).forEach(([id, position]) => animation.graph.mergeNodeAttributes(id, position));
    animation.callback?.();
};

const layoutMocks = vi.hoisted(() => ({
    sequentialLayout: vi.fn((graph: MultiDirectedGraph) => {
        graph.setNodeAttribute('alpha', 'x', 12);
        graph.setNodeAttribute('alpha', 'y', 12);
        graph.setNodeAttribute('bravo', 'x', 18);
        graph.setNodeAttribute('bravo', 'y', 18);
    }),
    standardLayout: vi.fn((graph: MultiDirectedGraph) => {
        graph.setNodeAttribute('alpha', 'x', 22);
        graph.setNodeAttribute('alpha', 'y', 22);
        graph.setNodeAttribute('bravo', 'x', 28);
        graph.setNodeAttribute('bravo', 'y', 28);
    }),
}));

vi.mock('@react-sigma/core', () => ({
    useRegisterEvents: () => sigmaMocks.registerEvents,
    useSetSettings: () => sigmaMocks.setSettings,
    useSigma: () => sigmaMocks.sigma,
}));

vi.mock('bh-shared-ui', () => ({
    useTheme: () => ({ contrast: '#ffffff', neutral: { primary: '#000000' } }),
}));

vi.mock('sigma/utils/animate', () => animationMocks);

vi.mock('src/store', () => ({
    useAppSelector: (selector: (state: any) => unknown) =>
        selector({
            global: {
                view: {
                    darkMode: false,
                    exploreLayout: undefined,
                    isExploreGraphHighlight: false,
                    isExploreGraphLabelClip: false,
                },
            },
        }),
}));

vi.mock('src/ducks/graph/utils', async (importOriginal) => ({
    ...(await importOriginal<typeof import('src/ducks/graph/utils')>()),
    getNodeOffset: () => ({ x: 0, y: 0 }),
    resetCamera: vi.fn(),
}));

vi.mock('src/utils', () => ({ preventAllDefaults: vi.fn() }));

vi.mock('src/views/Explore/utils', () => layoutMocks);

const getPosition = (graph: MultiDirectedGraph, id: string) => {
    const { x, y } = graph.getNodeAttributes(id);
    return { x, y };
};

const expectGridAlignedWithoutCollision = (graph: MultiDirectedGraph) => {
    const alpha = getPosition(graph, 'alpha');
    const bravo = getPosition(graph, 'bravo');

    expect([alpha.x, alpha.y, bravo.x, bravo.y].every((coordinate) => coordinate % 100 === 0)).toBe(true);
    expect(alpha).not.toEqual(bravo);
};

const createGraph = () => {
    const graph = new MultiDirectedGraph();
    graph.addNode('alpha', { x: 10, y: 10 });
    graph.addNode('bravo', { x: 20, y: 20 });
    return graph;
};

const createSigma = (graph: MultiDirectedGraph) => ({
    framedGraphToViewport: vi.fn((position) => position),
    getCamera: vi.fn(() => ({ animate: vi.fn(), ratio: 1 })),
    getGraph: () => graph,
    getNodeDisplayData: vi.fn(),
    refresh: vi.fn(),
    scheduleRefresh: vi.fn(),
    viewportToGraph: vi.fn((position) => position),
});

describe('GraphEvents snap to grid', () => {
    beforeEach(() => {
        sigmaMocks.handlers = {};
        sigmaMocks.registerEvents.mockClear();
        sigmaMocks.setSettings.mockClear();
        animationMocks.animateNodes.mockClear();
        animationMocks.pending = [];
        layoutMocks.sequentialLayout.mockClear();
        layoutMocks.standardLayout.mockClear();
    });

    it('aligns the current graph immediately when enabled', () => {
        const graph = createGraph();
        sigmaMocks.sigma = createSigma(graph);

        render(<GraphEvents highlightedItem={null} snapToGridEnabled />);

        expectGridAlignedWithoutCollision(graph);
        expect(sigmaMocks.sigma.refresh).toHaveBeenCalledOnce();
    });

    it('moves freely while dragging, then snaps around occupied cells on release', () => {
        const graph = createGraph();
        sigmaMocks.sigma = createSigma(graph);
        render(<GraphEvents highlightedItem={null} snapToGridEnabled />);
        const fixedPosition = getPosition(graph, 'bravo');

        act(() => {
            sigmaMocks.handlers.downNode({ event: { original: { button: 0 }, x: 0, y: 0 }, node: 'alpha' });
        });
        act(() => {
            sigmaMocks.handlers.mousemovebody({ ...fixedPosition });
        });

        expect(getPosition(graph, 'alpha')).toEqual(fixedPosition);

        act(() => {
            sigmaMocks.handlers.mouseup({});
            completeLatestAnimation();
        });

        expect(getPosition(graph, 'bravo')).toEqual(fixedPosition);
        expect(getPosition(graph, 'alpha')).not.toEqual(fixedPosition);
        expectGridAlignedWithoutCollision(graph);
        expect(animationMocks.animateNodes).toHaveBeenCalledWith(
            graph,
            { alpha: { x: 0, y: -200 } },
            { duration: 100, easing: 'quadraticOut' },
            expect.any(Function)
        );
        expect(sigmaMocks.sigma.refresh).toHaveBeenCalledOnce();
    });

    it('keeps a rapid re-grab active after the prior drag reset delay', () => {
        vi.useFakeTimers();
        const graph = createGraph();
        sigmaMocks.sigma = createSigma(graph);
        render(<GraphEvents highlightedItem={null} snapToGridEnabled />);

        act(() => {
            sigmaMocks.handlers.downNode({ event: { original: { button: 0 }, x: 0, y: 0 }, node: 'alpha' });
        });
        act(() => {
            sigmaMocks.handlers.mousemovebody({ x: 37, y: 43 });
        });
        act(() => {
            sigmaMocks.handlers.mouseup({});
        });
        const firstAnimation = animationMocks.pending[0];

        act(() => {
            sigmaMocks.handlers.downNode({ event: { original: { button: 0 }, x: 37, y: 43 }, node: 'alpha' });
        });
        act(() => {
            vi.advanceTimersByTime(10);
            sigmaMocks.handlers.mousemovebody({ x: 71, y: 83 });
        });

        expect(firstAnimation.cancel).toHaveBeenCalledOnce();
        expect(getPosition(graph, 'alpha')).toEqual({ x: 71, y: 83 });
        vi.useRealTimers();
    });

    it('cancels settlement before disabling snap or applying another layout', () => {
        const graph = createGraph();
        const ref = createRef<any>();
        sigmaMocks.sigma = createSigma(graph);
        const { rerender } = render(<GraphEvents highlightedItem={null} snapToGridEnabled ref={ref} />);

        act(() => {
            sigmaMocks.handlers.downNode({ event: { original: { button: 0 }, x: 0, y: 0 }, node: 'alpha' });
        });
        act(() => {
            sigmaMocks.handlers.mousemovebody({ x: 37, y: 43 });
        });
        act(() => {
            sigmaMocks.handlers.mouseup({});
        });
        const animationBeforeDisable = animationMocks.pending[0];
        rerender(<GraphEvents highlightedItem={null} snapToGridEnabled={false} ref={ref} />);
        const positionAtDisable = getPosition(graph, 'alpha');
        completeLatestAnimation();

        expect(animationBeforeDisable.cancel).toHaveBeenCalledOnce();
        expect(getPosition(graph, 'alpha')).toEqual(positionAtDisable);

        rerender(<GraphEvents highlightedItem={null} snapToGridEnabled ref={ref} />);
        act(() => {
            sigmaMocks.handlers.downNode({ event: { original: { button: 0 }, x: 0, y: 0 }, node: 'alpha' });
        });
        act(() => {
            sigmaMocks.handlers.mousemovebody({ x: 137, y: 143 });
        });
        act(() => {
            sigmaMocks.handlers.mouseup({});
        });
        act(() => {
            ref.current.runSequentialLayout();
        });
        const animationBeforeLayout = animationMocks.pending[1];
        const positionAfterLayout = getPosition(graph, 'alpha');
        completeLatestAnimation();

        expect(animationBeforeLayout.cancel).toHaveBeenCalledOnce();
        expect(getPosition(graph, 'alpha')).toEqual(positionAfterLayout);
        expectGridAlignedWithoutCollision(graph);
    });

    it('resnaps both imperative layouts while enabled', () => {
        const graph = createGraph();
        const ref = createRef<any>();
        sigmaMocks.sigma = createSigma(graph);
        render(<GraphEvents highlightedItem={null} snapToGridEnabled ref={ref} />);

        act(() => ref.current.runSequentialLayout());
        expect(layoutMocks.sequentialLayout).toHaveBeenCalledWith(graph);
        expectGridAlignedWithoutCollision(graph);

        act(() => ref.current.runStandardLayout());
        expect(layoutMocks.standardLayout).toHaveBeenCalledWith(graph);
        expectGridAlignedWithoutCollision(graph);
    });

    it('avoids graph-wide occupancy work and permits free-form dragging while disabled', () => {
        const graph = createGraph();
        sigmaMocks.sigma = createSigma(graph);
        const { rerender } = render(<GraphEvents highlightedItem={null} snapToGridEnabled />);
        const alignedPosition = getPosition(graph, 'alpha');
        const nodesSpy = vi.spyOn(graph, 'nodes');

        rerender(<GraphEvents highlightedItem={null} snapToGridEnabled={false} />);
        expect(getPosition(graph, 'alpha')).toEqual(alignedPosition);
        nodesSpy.mockClear();

        act(() => {
            sigmaMocks.handlers.downNode({ event: { original: { button: 0 }, x: 0, y: 0 }, node: 'alpha' });
        });
        expect(nodesSpy).not.toHaveBeenCalled();

        act(() => {
            sigmaMocks.handlers.mousemovebody({ x: 37, y: 43 });
        });
        expect(getPosition(graph, 'alpha')).toEqual({ x: 37, y: 43 });
    });
});
