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

import { Box, Grid, Link, useTheme } from '@mui/material';
import {
    EdgeInfoState,
    GraphButtonProps,
    GraphProgress,
    NoDataAlert,
    setEdgeInfoOpen,
    setSelectedEdge,
    useAvailableDomains,
} from 'bh-shared-ui';
import { MultiDirectedGraph } from 'graphology';
import { random } from 'graphology-layout';
import forceAtlas2 from 'graphology-layout-forceatlas2';
import { Attributes } from 'graphology-types';
import { GraphNodes } from 'js-client-library';
import isEmpty from 'lodash/isEmpty';
import { FC, useEffect, useState } from 'react';
import { Link as RouterLink } from 'react-router-dom';
import { SigmaNodeEventPayload } from 'sigma/sigma';
import { GraphButtonOptions } from 'src/components/GraphButtons/GraphButtons';
import SigmaChart from 'src/components/SigmaChart';
import { setEntityInfoOpen, setSelectedNode } from 'src/ducks/entityinfo/actions';
import { GraphState } from 'src/ducks/explore/types';
import { setAssetGroupEdit } from 'src/ducks/global/actions';
import { ROUTE_ADMINISTRATION_FILE_INGEST } from 'src/ducks/global/routes';
import { GlobalOptionsState } from 'src/ducks/global/types';
import { discardChanges } from 'src/ducks/tierzero/actions';
import { RankDirection } from 'src/hooks/useLayoutDagre/useLayoutDagre';
import useToggle from 'src/hooks/useToggle';
import { useAppDispatch, useAppSelector } from 'src/store';
import { transformFlatGraphResponse } from 'src/utils';
import EdgeInfoPane from 'src/views/Explore/EdgeInfo/EdgeInfoPane';
import EntityInfoPanel from 'src/views/Explore/EntityInfo/EntityInfoPanel';
import ExploreSearch from 'src/views/Explore/ExploreSearch';
import usePrompt from 'src/views/Explore/NavigationAlert';
import { initGraphEdges, initGraphNodes } from 'src/views/Explore/utils';
import ContextMenu from './ContextMenu/ContextMenu';

const GraphView: FC = () => {
    /* Hooks */
    const theme = useTheme();
    const dispatch = useAppDispatch();

    const graphState: GraphState = useAppSelector((state) => state.explore);
    const opts: GlobalOptionsState = useAppSelector((state) => state.global.options);
    const formIsDirty = Object.keys(useAppSelector((state) => state.tierzero).changelog).length > 0;

    const [graphologyGraph, setGraphologyGraph] = useState<MultiDirectedGraph<Attributes, Attributes, Attributes>>();
    const [currentNodes, setCurrentNodes] = useState<GraphNodes>({});

    const [currentSearchOpen, toggleCurrentSearch] = useToggle(false);
    const { data, isLoading, isError } = useAvailableDomains();

    const [contextMenu, setContextMenu] = useState<{ mouseX: number; mouseY: number } | null>(null);

    useEffect(() => {
        let items: any = graphState.chartProps.items;
        if (!items) return;
        // `items` may be empty, or it may contain an empty `nodes` object
        if (isEmpty(items) || isEmpty(items.nodes)) items = transformFlatGraphResponse(items);

        const graph = new MultiDirectedGraph();
        const nodeSize = 25;

        initGraphNodes(graph, items.nodes, nodeSize);
        initGraphEdges(graph, items.edges);

        setCurrentNodes(items.nodes);

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

    if (isLoading) {
        return (
            <Box sx={{ position: 'relative', height: '100%', width: '100%', overflow: 'hidden' }} data-testid='explore'>
                <GraphProgress loading={isLoading} />
            </Box>
        );
    }

    const dataCollectionLink = (
        <Link
            target='_blank'
            href={'https://support.bloodhoundenterprise.io/hc/en-us/sections/17274904083483-BloodHound-CE-Collection'}>
            Data Collection
        </Link>
    );

    const fileIngestLink = (
        <Link component={RouterLink} to={ROUTE_ADMINISTRATION_FILE_INGEST}>
            File Ingest
        </Link>
    );

    if (isError) throw new Error();

    if (!data.length)
        return (
            <Box position={'relative'} height={'100%'} width={'100%'} overflow={'hidden'}>
                <NoDataAlert dataCollectionLink={dataCollectionLink} fileIngestLink={fileIngestLink} />
            </Box>
        );

    const options: GraphButtonOptions = { standard: true, sequential: true };

    const nonLayoutButtons: GraphButtonProps[] = [
        {
            displayText: 'Search Current Results',
            onClick: toggleCurrentSearch,
            disabled: currentSearchOpen,
        },
    ];

    const findNodeAndSelect = (id: string) => {
        const selectedItem = graphState.chartProps.items?.[id];
        if (selectedItem?.data?.nodetype) {
            dispatch(setSelectedEdge(null));
            dispatch(
                setSelectedNode({
                    id: selectedItem.data.objectid,
                    type: selectedItem.data.nodetype,
                    name: selectedItem.data.name,
                    graphId: id,
                })
            );
        }
    };

    /* Event Handlers */
    const onClickNode = (id: string) => {
        dispatch(setEdgeInfoOpen(false));
        dispatch(setEntityInfoOpen(true));

        findNodeAndSelect(id);
    };

    const handleContextMenu = (event: SigmaNodeEventPayload) => {
        setContextMenu(contextMenu === null ? { mouseX: event.event.x, mouseY: event.event.y } : null);

        const nodeId = event.node;

        findNodeAndSelect(nodeId);
    };

    const handleCloseContextMenu = () => {
        setContextMenu(null);
    };

    return (
        <Box sx={{ position: 'relative', height: '100%', width: '100%', overflow: 'hidden' }} data-testid='explore'>
            <SigmaChart
                rankDirection={RankDirection.LEFT_RIGHT}
                options={options}
                graph={graphologyGraph}
                currentNodes={currentNodes}
                onClickNode={onClickNode}
                nonLayoutButtons={nonLayoutButtons}
                isCurrentSearchOpen={currentSearchOpen}
                toggleCurrentSearch={toggleCurrentSearch}
                handleContextMenu={handleContextMenu}
            />

            <ContextMenu contextMenu={contextMenu} handleClose={handleCloseContextMenu} />

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
        </Box>
    );
};

const GridItems = () => {
    const selectedNode = useAppSelector((state) => state.entityinfo.selectedNode);

    const columnsDefault = { xs: 6, md: 5, lg: 4, xl: 3 };
    const cypherSearchColumns = { xs: 6, md: 6, lg: 6, xl: 4 };

    const edgeInfoState: EdgeInfoState = useAppSelector((state) => state.edgeinfo);
    const [columns, setColumns] = useState(columnsDefault);
    const theme = useTheme();

    const columnStyles = { height: '100%' };

    const infoPanelStyles = {
        margin: theme.spacing(0, 4, 2, 2),
        maxHeight: '95%',
    };

    const handleCypherTab = (isCypherEditorActive: boolean) => {
        isCypherEditorActive ? setColumns(cypherSearchColumns) : setColumns(columnsDefault);
    };

    return [
        <Grid item {...columns} sx={columnStyles} key={'exploreSearch'}>
            <ExploreSearch handleColumns={handleCypherTab} />
        </Grid>,
        <Grid item {...columnsDefault} sx={columnStyles} key={'info'}>
            {edgeInfoState.open ? (
                <EdgeInfoPane sx={infoPanelStyles} selectedEdge={edgeInfoState.selectedEdge} />
            ) : (
                <EntityInfoPanel sx={infoPanelStyles} selectedNode={selectedNode} />
            )}
        </Grid>,
    ];
};

export default GraphView;
