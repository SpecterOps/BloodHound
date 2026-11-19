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

import {
    deriveProjectionPresentation,
    type GroupDefinition,
    type ProjectedGraph,
    type SourceGraph,
} from 'bh-shared-ui';
import { MultiDirectedGraph } from 'graphology';
import { useLayoutEffect, useMemo, useRef, useState } from 'react';
import { initGraph, standardLayout, type GraphOptions } from '../utils';
import { graphBounds, sigmaMemberGlyphs, updateSigmaProjection } from './sigmaProjection';

type SigmaProjectionViewOptions = {
    source?: SourceGraph;
    projected?: ProjectedGraph;
    groups: readonly GroupDefinition[];
    queryKey: unknown;
    options: GraphOptions;
    layout: string | undefined;
};

type SigmaProjectionView = {
    identity: string | undefined;
    graph: MultiDirectedGraph;
    index: ReturnType<typeof updateSigmaProjection>;
};

const toGraphData = (source: SourceGraph) => ({
    nodes: Object.fromEntries(Object.entries(source.nodes).map(([id, node]) => [id, node.node])),
    edges: Object.values(source.edges)
        .filter((edge) => Object.hasOwn(source.nodes, edge.source) && Object.hasOwn(source.nodes, edge.target))
        .map((edge) => ({ ...edge.edge, exploreGraphId: edge.id })),
});

export function useSigmaProjectionView({
    source,
    projected,
    groups,
    queryKey,
    options,
    layout,
}: SigmaProjectionViewOptions) {
    const queryIdentity = JSON.stringify(queryKey);
    const rendererCacheRef = useRef<{ queryIdentity: string | undefined; graph: MultiDirectedGraph }>();
    const [view, setView] = useState<SigmaProjectionView>();
    const data = useMemo(() => (source ? toGraphData(source) : undefined), [source]);
    const styles = useMemo(
        () => (data ? initGraph(data, { ...options, hideNodes: false }, false) : undefined),
        [data, options]
    );

    useLayoutEffect(() => {
        if (!source || !projected || !styles || !data) {
            return;
        }

        // Keep one renderer graph per query so expanding/collapsing retains cached node positions.
        // Rebuild it only when the query changes; theme and style updates refresh attributes below.
        if (!rendererCacheRef.current || rendererCacheRef.current.queryIdentity !== queryIdentity) {
            const initialGraph = initGraph(data, { ...options, hideNodes: false });
            if (layout === 'standard') {
                standardLayout(initialGraph);
            }
            initialGraph.setAttribute('originalBBox', graphBounds(initialGraph));
            rendererCacheRef.current = { queryIdentity, graph: initialGraph };
        }

        const rendererGraph = rendererCacheRef.current.graph;
        const presentation = deriveProjectionPresentation(source, projected, sigmaMemberGlyphs(styles), new Set());
        const index = updateSigmaProjection(rendererGraph, styles, source, projected, presentation, groups);

        if (options.hideNodes) {
            rendererGraph.forEachNode((id) => rendererGraph.setNodeAttribute(id, 'hidden', true));
            rendererGraph.forEachEdge((id) => rendererGraph.setEdgeAttribute(id, 'hidden', true));
        }

        setView({ identity: queryIdentity, graph: rendererGraph, index });
    }, [source, projected, groups, queryIdentity, styles, data, options, layout]);

    return view && view.identity === queryIdentity ? view : { graph: undefined, index: undefined };
}
