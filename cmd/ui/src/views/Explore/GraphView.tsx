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

import { Box, Grid, Link, Popper, useTheme } from '@mui/material';
import {
    EdgeInfoState,
    GraphProgress,
    NoDataAlert,
    SearchCurrentNodes,
    WebGLDisabledAlert,
    exportToJson,
    isWebGLEnabled,
    setEdgeInfoOpen,
    setSelectedEdge,
    useAvailableDomains,
} from 'bh-shared-ui';
import { MultiDirectedGraph } from 'graphology';
import { Attributes } from 'graphology-types';
import { GraphNodes } from 'js-client-library';
import isEmpty from 'lodash/isEmpty';
import { FC, useEffect, useRef, useState } from 'react';
import { Link as RouterLink } from 'react-router-dom';
import { SigmaNodeEventPayload } from 'sigma/sigma';
import GraphButtons from 'src/components/GraphButtons/GraphButtons';
import SigmaChart from 'src/components/SigmaChart';
import { setEntityInfoOpen, setSelectedNode } from 'src/ducks/entityinfo/actions';
import { GraphState } from 'src/ducks/explore/types';
import { setAssetGroupEdit } from 'src/ducks/global/actions';
import { ROUTE_ADMINISTRATION_FILE_INGEST } from 'src/ducks/global/routes';
import { GlobalOptionsState } from 'src/ducks/global/types';
import { discardChanges } from 'src/ducks/tierzero/actions';
import { useToggle } from 'bh-shared-ui';
import { useAppDispatch, useAppSelector } from 'src/store';
import { transformFlatGraphResponse } from 'src/utils';
import EdgeInfoPane from 'src/views/Explore/EdgeInfo/EdgeInfoPane';
import EntityInfoPanel from 'src/views/Explore/EntityInfo/EntityInfoPanel';
import ExploreSearch from 'src/views/Explore/ExploreSearch';
import usePrompt from 'src/views/Explore/NavigationAlert';
import { initGraph } from 'src/views/Explore/utils';
import ContextMenu from './ContextMenu/ContextMenu';

const columnsDefault = { xs: 6, md: 5, lg: 4, xl: 3 };

const cypherSearchColumns = { xs: 6, md: 6, lg: 6, xl: 4 };

const columnStyles = { height: '100%', overflow: 'hidden', display: 'flex', flexDirection: 'column' };

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

const sampleDataLink = (
    <Link target='_blank' href={'https://github.com/SpecterOps/BloodHound/wiki/Example-Data'}>
        GitHub Sample Collection
    </Link>
);

const GraphView: FC = () => {
    /* Hooks */
    const theme = useTheme();

    const dispatch = useAppDispatch();

    const graphState: GraphState = useAppSelector((state) => state.explore);

    const opts: GlobalOptionsState = useAppSelector((state) => state.global.options);

    const formIsDirty = Object.keys(useAppSelector((state) => state.tierzero).changelog).length > 0;

    const darkMode = useAppSelector((state) => state.global.view.darkMode);

    const [graphologyGraph, setGraphologyGraph] = useState<MultiDirectedGraph<Attributes, Attributes, Attributes>>();

    const [currentNodes, setCurrentNodes] = useState<GraphNodes>({});

    const [currentSearchOpen, toggleCurrentSearch] = useToggle(false);

    const { data, isLoading, isError } = useAvailableDomains();

    const [contextMenu, setContextMenu] = useState<{ mouseX: number; mouseY: number } | null>(null);

    const exportableGraphState = useAppSelector((state) => state.explore.export);

    const sigmaChartRef = useRef<any>(null);

    const currentSearchAnchorElement = useRef(null);

    const selectedNode = useAppSelector((state) => state.entityinfo.selectedNode);

    const edgeInfoState: EdgeInfoState = useAppSelector((state) => state.edgeinfo);

    const [columns, setColumns] = useState(columnsDefault);

    const [showNodeLabels, setShowNodeLabels] = useState(true);

    const [showEdgeLabels, setShowEdgeLabels] = useState(true);

    useEffect(() => {
        let items: any = graphState.chartProps.items;
        if (!items) return;
        // `items` may be empty, or it may contain an empty `nodes` object
        if (isEmpty(items) || isEmpty(items.nodes)) items = transformFlatGraphResponse(items);

        const graph = new MultiDirectedGraph();

        initGraph(graph, items, theme, darkMode);

        setCurrentNodes(items.nodes);

        setGraphologyGraph(graph);
    }, [graphState.chartProps.items, theme, darkMode]);

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

    if (isError) throw new Error();

    if (!isWebGLEnabled()) {
        return <WebGLDisabledAlert />;
    }

    if (!data.length)
        return (
            <Box position={'relative'} height={'100%'} width={'100%'} overflow={'hidden'}>
                <NoDataAlert
                    dataCollectionLink={dataCollectionLink}
                    fileIngestLink={fileIngestLink}
                    sampleDataLink={sampleDataLink}
                />
            </Box>
        );

    /* Event Handlers */
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

    const handleClickNode = (id: string) => {
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

    const handleCypherTab = (tab: string) => {
        tab === 'cypher' ? setColumns(cypherSearchColumns) : setColumns(columnsDefault);
    };

    return (
        <Box
            sx={{
                position: 'relative',
                height: '100%',
                width: '100%',
                overflow: 'hidden',
            }}
            data-testid='explore'>
            <SigmaChart
                graph={graphologyGraph}
                onClickNode={handleClickNode}
                handleContextMenu={handleContextMenu}
                showNodeLabels={showNodeLabels}
                showEdgeLabels={showEdgeLabels}
                ref={sigmaChartRef}
            />

            <Grid
                container
                direction='row'
                justifyContent='space-between'
                alignItems='flex-start'
                sx={{
                    position: 'relative',
                    padding: theme.spacing(2),
                    boxSizing: 'border-box',
                    pointerEvents: 'none',
                    height: '100%',
                }}>
                <Grid
                    item
                    {...columns}
                    sx={{
                        ...columnStyles,
                        justifyContent: 'space-between',
                        height: '100%',
                        maxHeight: '100%',
                        gap: 2,
                    }}
                    key={'exploreSearch'}>
                    <ExploreSearch onTabChange={handleCypherTab} />
                    <Box
                        sx={{
                            pointerEvents: 'auto',
                            width: '100%',
                            position: 'relative',
                        }}
                        ref={currentSearchAnchorElement}>
                        <GraphButtons
                            onExportJson={() => {
                                exportToJson(exportableGraphState);
                            }}
                            onReset={() => {
                                sigmaChartRef.current?.resetCamera();
                            }}
                            onRunSequentialLayout={() => {
                                sigmaChartRef.current?.runSequentialLayout();
                            }}
                            onRunStandardLayout={() => {
                                sigmaChartRef.current?.runStandardLayout();
                            }}
                            onSearchCurrentResults={() => {
                                toggleCurrentSearch();
                            }}
                            onToggleAllLabels={() => {
                                if (!showNodeLabels || !showEdgeLabels) {
                                    setShowNodeLabels(true);
                                    setShowEdgeLabels(true);
                                } else {
                                    setShowNodeLabels(false);
                                    setShowEdgeLabels(false);
                                }
                            }}
                            onToggleNodeLabels={() => {
                                setShowNodeLabels((prev) => !prev);
                            }}
                            onToggleEdgeLabels={() => {
                                setShowEdgeLabels((prev) => !prev);
                            }}
                            showNodeLabels={showNodeLabels}
                            showEdgeLabels={showEdgeLabels}
                            isCurrentSearchOpen={false}
                        />
                        <Popper
                            open={currentSearchOpen}
                            anchorEl={currentSearchAnchorElement.current}
                            placement='top'
                            disablePortal
                            sx={{
                                width: '90%',
                                zIndex: 1,
                            }}>
                            <SearchCurrentNodes
                                sx={{ padding: 1, marginBottom: 1 }}
                                currentNodes={currentNodes || {}}
                                onSelect={(node) => {
                                    handleClickNode?.(node.id);
                                    toggleCurrentSearch?.();
                                }}
                                onClose={toggleCurrentSearch}
                            />
                        </Popper>
                    </Box>
                </Grid>
                <Grid item {...columnsDefault} sx={columnStyles} key={'info'}>
                    {edgeInfoState.open ? (
                        <EdgeInfoPane selectedEdge={edgeInfoState.selectedEdge} />
                    ) : (
                        <EntityInfoPanel selectedNode={selectedNode} />
                    )}
                </Grid>
            </Grid>
            <ContextMenu contextMenu={contextMenu} handleClose={handleCloseContextMenu} />
            <GraphProgress loading={graphState.loading} />
        </Box>
    );
};

export default GraphView;
