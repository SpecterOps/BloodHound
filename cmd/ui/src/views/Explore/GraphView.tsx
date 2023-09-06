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

import { Grid, useTheme } from '@mui/material';
import { EdgeInfoState, GraphProgress, setEdgeInfoOpen, setSelectedEdge } from 'bh-shared-ui';
import { MultiDirectedGraph } from 'graphology';
import { random } from 'graphology-layout';
import forceAtlas2 from 'graphology-layout-forceatlas2';
import { Attributes } from 'graphology-types';
import { GraphEdges, GraphNodes } from 'js-client-library';
import isEmpty from 'lodash/isEmpty';
import { FC, useEffect, useState } from 'react';
import { useSelector } from 'react-redux';
import { GraphButtonOptions } from 'src/components/GraphButtons/GraphButtons';
import SigmaChart from 'src/components/SigmaChart';
import { setEntityInfoOpen, setSelectedNode } from 'src/ducks/entityinfo/actions';
import { GraphState } from 'src/ducks/explore/types';
import { setAssetGroupEdit } from 'src/ducks/global/actions';
import { GlobalOptionsState } from 'src/ducks/global/types';
import { discardChanges } from 'src/ducks/tierzero/actions';
import { RankDirection } from 'src/hooks/useLayoutDagre/useLayoutDagre';
import { GLYPHS, GlyphKind, NODE_ICON, UNKNOWN_ICON } from 'src/icons';
import { GlyphLocation } from 'src/rendering/programs/node.glyphs';
import { AppState, useAppDispatch } from 'src/store';
import { EdgeDirection, EdgeParams, NodeParams, transformFlatGraphResponse } from 'src/utils';
import EdgeInfoPane from 'src/views/Explore/EdgeInfo/EdgeInfoPane';
import EntityInfoPanel from 'src/views/Explore/EntityInfo/EntityInfoPanel';
import ExploreSearch from 'src/views/Explore/ExploreSearch';
import usePrompt from 'src/views/Explore/NavigationAlert';

const initGraphNodes = (graph: MultiDirectedGraph, nodes: GraphNodes, nodeSize: number) => {
    Object.keys(nodes).forEach((key: string) => {
        const node = nodes[key];
        // Set default node parameters
        const nodeParams: Partial<NodeParams> = {
            color: '#FFFFFF',
            type: 'combined',
            label: node.label,
            forceLabel: true,
        };

        const icon = NODE_ICON[node.kind] || UNKNOWN_ICON;
        nodeParams.color = icon.color;
        nodeParams.image = icon.url || '';

        // Tier zero nodes should be marked with a gem glyph
        if (node.isTierZero) {
            const glyph = GLYPHS[GlyphKind.TIER_ZERO];
            nodeParams.type = 'glyphs';
            nodeParams.glyphs = [
                {
                    location: GlyphLocation.TOP_RIGHT,
                    image: glyph.url || '',
                    backgroundColor: glyph.color,
                },
            ];
        }

        graph.addNode(key, {
            size: nodeSize,
            borderColor: '#000000',
            ...nodeParams,
        });
    });
};

const initGraphEdges = (graph: MultiDirectedGraph, edges: GraphEdges) => {
    // Group edges with the same start and end nodes into arrays. Should be grouped regardless of direction
    const groupedEdges = edges.reduce<Record<string, GraphEdges>>((groups, edge) => {
        const identifiers = [edge.source, edge.target].sort();
        const id = `${identifiers[0]}_${identifiers[1]}`;

        if (!groups[id]) {
            groups[id] = [];
        }
        groups[id].push(edge);

        return groups;
    }, {});

    // Loop through our group arrays
    for (const group in groupedEdges) {
        const groupSize = groupedEdges[group].length;

        for (const [i, edge] of groupedEdges[group].entries()) {
            const key = `${edge.source}_${edge.kind}_${edge.target}`;

            // Set default values for single edges
            const edgeParams: Partial<EdgeParams> = {
                size: 3,
                type: 'arrow',
                label: edge.label,
                color: '#000000C0',
                groupPosition: 0,
                groupSize: 1,
                exploreGraphId: edge.exploreGraphId || key,
                forceLabel: true,
            };

            // Groups with odd-numbered totals should have a straight edge first, then curve the rest
            const edgeShouldBeCurved = groupSize > 1;

            // Handle edge groups that have a mix of directions that edges travel between source and target.
            // We can use the value of the enum to indicate which direction the curve should bend
            const groupStart = group.split('_')[0];
            const edgeDirection = groupStart === edge.source ? EdgeDirection.FORWARDS : EdgeDirection.BACKWARDS;

            if (edgeShouldBeCurved) {
                edgeParams.type = 'curved';
                edgeParams.groupPosition = i;
                edgeParams.groupSize = groupSize;
                edgeParams.direction = edgeDirection;
            }

            graph.addEdgeWithKey(key, edge.source, edge.target, edgeParams);
        }
    }
};

const GraphView: FC = () => {
    /* Hooks */
    const theme = useTheme();
    const dispatch = useAppDispatch();

    const graphState: GraphState = useSelector((state: AppState) => state.explore);
    const opts: GlobalOptionsState = useSelector((state: AppState) => state.global.options);
    const formIsDirty = Object.keys(useSelector((state: AppState) => state.tierzero).changelog).length > 0;
    const [graphologyGraph, setGraphologyGraph] = useState<MultiDirectedGraph<Attributes, Attributes, Attributes>>();

    useEffect(() => {
        let items: any = graphState.chartProps.items;
        if (!items || isEmpty(items)) return;
        if (!items.nodes || isEmpty(items.nodes)) items = transformFlatGraphResponse(items);

        const graph = new MultiDirectedGraph();
        const nodeSize = 25;

        initGraphNodes(graph, items.nodes, nodeSize);
        initGraphEdges(graph, items.edges);

        random.assign(graph, { scale: 1000 });

        forceAtlas2.assign(graph, {
            iterations: 128,
            settings: {
                scalingRatio: 1000,
                barnesHutOptimize: true,
            },
        });
        setGraphologyGraph(graph);
    }, [graphState.chartProps.items]);

    useEffect(() => {
        if (opts.assetGroupEdit !== null) {
            dispatch(setAssetGroupEdit(null));
        }
    }, [opts.assetGroupEdit, dispatch]);

    //Clear the changelog when leaving the explore page.
    useEffect(() => {
        return () => {
            formIsDirty && dispatch(discardChanges());
        };
    }, [dispatch, formIsDirty]);

    usePrompt(
        'You will lose any unsaved changes if you navigate away from this page without saving. Continue?',
        formIsDirty
    );

    /* Event Handlers */
    const onClickNode = (id: string) => {
        dispatch(setEdgeInfoOpen(false));
        dispatch(setEntityInfoOpen(true));

        const selectedItem = graphState.chartProps.items?.[id];
        if (selectedItem?.data?.nodetype) {
            dispatch(setSelectedEdge(null));
            dispatch(
                setSelectedNode({
                    id: selectedItem.data.objectid,
                    type: selectedItem.data.nodetype,
                    name: selectedItem.data.name,
                })
            );
        }
    };

    const options: GraphButtonOptions = { standard: true, sequential: true };

    return (
        <div style={{ position: 'relative', height: '100%', width: '100%', overflow: 'hidden' }} data-testid='explore'>
            <SigmaChart
                rankDirection={RankDirection.LEFT_RIGHT}
                options={options}
                graph={graphologyGraph}
                onClickNode={onClickNode}
            />
            <Grid
                container
                direction='row'
                justifyContent='space-between'
                alignItems='flex-start'
                sx={{
                    position: 'relative',
                    margin: theme.spacing(2, 2, 0),
                    pointerEvents: 'none',
                    height: '100%',
                }}>
                <GridItems />
            </Grid>

            <GraphProgress loading={graphState.loading} />
        </div>
    );
};

const GridItems = () => {
    const columnsDefault = { xs: 6, md: 5, lg: 4, xl: 4 };
    const cypherSearchColumns = { xs: 6, md: 6, lg: 6, xl: 4 };

    const [columns, setColumns] = useState(columnsDefault);

    const handleCypherTab = (isCypherEditorActive: boolean) => {
        isCypherEditorActive ? setColumns(cypherSearchColumns) : setColumns(columnsDefault);
    };

    const edgeInfoState: EdgeInfoState = useSelector((state: AppState) => state.edgeinfo);

    const { xs, md, lg, xl } = columns;

    return [
        <Grid item xs={xs} md={md} lg={lg} xl={xl} sx={{ height: '100%' }} key={'exploreSearch'}>
            <ExploreSearch handleColumns={handleCypherTab} />
        </Grid>,
        <Grid item xs={6} md={5} lg={4} sx={{ height: '100%' }} key={'info'}>
            {edgeInfoState.open ? <EdgeInfoPane selectedEdge={edgeInfoState.selectedEdge} /> : <EntityInfoPanel />}
        </Grid>,
    ];
};

export default GraphView;
