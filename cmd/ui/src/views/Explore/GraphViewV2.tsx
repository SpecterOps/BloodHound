// Copyright 2025 Specter Ops, Inc.
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

import { Popper, useTheme } from '@mui/material';
import {
    GraphProgress,
    SearchCurrentNodes,
    WebGLDisabledAlert,
    exportToJson,
    isWebGLEnabled,
    transformFlatGraphResponse,
    useAvailableEnvironments,
    useExploreSelectedItem,
    useToggle,
} from 'bh-shared-ui';
import { MultiDirectedGraph } from 'graphology';
import { Attributes } from 'graphology-types';
import { GraphNodes } from 'js-client-library';
import isEmpty from 'lodash/isEmpty';
import { FC, useEffect, useRef, useState } from 'react';
import { SigmaNodeEventPayload } from 'sigma/sigma';
import GraphButtons from 'src/components/GraphButtons/GraphButtons';
import { NoDataDialogWithLinks } from 'src/components/NoDataDialogWithLinks';
import SigmaChart from 'src/components/SigmaChart';
import { useSigmaExploreGraph } from 'src/hooks/useSigmaExploreGraph';
import { useAppSelector } from 'src/store';
import { initGraph } from 'src/views/Explore/utils';
import ContextMenuV2 from './ContextMenu/ContextMenuV2';
import ExploreSearchV2 from './ExploreSearch/ExploreSearchV2';
import GraphItemInformationPanel from './GraphItemInformationPanel';

const GraphViewV2: FC = () => {
    /* Hooks */
    const theme = useTheme();

    const graphQuery = useSigmaExploreGraph();
    const { data, isLoading, isError } = useAvailableEnvironments();
    const { setSelectedItem } = useExploreSelectedItem();

    const darkMode = useAppSelector((state) => state.global.view.darkMode);

    const [graphologyGraph, setGraphologyGraph] = useState<MultiDirectedGraph<Attributes, Attributes, Attributes>>();
    const [currentNodes, setCurrentNodes] = useState<GraphNodes>({});
    const [currentSearchOpen, toggleCurrentSearch] = useToggle(false);
    const [contextMenu, setContextMenu] = useState<{ mouseX: number; mouseY: number } | null>(null);
    const [showNodeLabels, setShowNodeLabels] = useState(true);
    const [showEdgeLabels, setShowEdgeLabels] = useState(true);
    const [exportJsonData, setExportJsonData] = useState();

    const sigmaChartRef = useRef<any>(null);
    const currentSearchAnchorElement = useRef(null);

    useEffect(() => {
        let items: any = graphQuery.data;

        if (!items && !graphQuery.isError) return;
        if (!items) items = {};

        // `items` may be empty, or it may contain an empty `nodes` object
        if (isEmpty(items) || isEmpty(items.nodes)) items = transformFlatGraphResponse(items);

        const graph = new MultiDirectedGraph();

        initGraph(graph, items, theme, darkMode);
        setExportJsonData(items);

        setCurrentNodes(items.nodes);

        setGraphologyGraph(graph);
    }, [graphQuery.data, theme, darkMode, graphQuery.isError]);

    if (isLoading) {
        return (
            <div className='relative h-full w-full overflow-hidden' data-testid='explore'>
                <GraphProgress loading={isLoading} />
            </div>
        );
    }

    if (isError) throw new Error();

    if (!isWebGLEnabled()) {
        return <WebGLDisabledAlert />;
    }

    /* Event Handlers */
    const handleClickNode = (id: string) => {
        setSelectedItem(id);
    };

    const handleContextMenu = (event: SigmaNodeEventPayload) => {
        setContextMenu(contextMenu === null ? { mouseX: event.event.x, mouseY: event.event.y } : null);
        setSelectedItem(event.node);
    };

    const handleCloseContextMenu = () => {
        setContextMenu(null);
    };

    return (
        <div
            className='relative h-full w-full overflow-hidden'
            data-testid='explore'
            onContextMenu={(e) => e.preventDefault()}>
            <SigmaChart
                graph={graphologyGraph}
                onClickNode={handleClickNode}
                handleContextMenu={handleContextMenu}
                showNodeLabels={showNodeLabels}
                showEdgeLabels={showEdgeLabels}
                ref={sigmaChartRef}
            />

            <div className='absolute top-0 h-full p-4 flex gap-2 justify-between flex-col pointer-events-none'>
                <ExploreSearchV2 />
                <div className='flex gap-1 pointer-events-auto' ref={currentSearchAnchorElement}>
                    <GraphButtons
                        onExportJson={() => {
                            exportToJson(exportJsonData);
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
                        isJsonExportDisabled={isEmpty(exportJsonData)}
                    />
                </div>
                <Popper
                    open={currentSearchOpen}
                    anchorEl={currentSearchAnchorElement.current}
                    placement='top'
                    disablePortal
                    className='w-[90%] z-[1]'>
                    <div className='pointer-events-auto' data-testid='explore_graph-controls'>
                        <SearchCurrentNodes
                            sx={{ padding: 1, marginBottom: 1 }}
                            currentNodes={currentNodes || {}}
                            onSelect={(node) => {
                                handleClickNode?.(node.id);
                                toggleCurrentSearch?.();
                            }}
                            onClose={toggleCurrentSearch}
                        />
                    </div>
                </Popper>
            </div>
            <GraphItemInformationPanel />
            <ContextMenuV2 contextMenu={contextMenu} handleClose={handleCloseContextMenu} />
            <GraphProgress loading={graphQuery.isLoading} />
            <NoDataDialogWithLinks open={!data?.length} />
        </div>
    );
};

export default GraphViewV2;
