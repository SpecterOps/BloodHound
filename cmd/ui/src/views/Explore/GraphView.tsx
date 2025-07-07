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

import { useTheme } from '@mui/material';
import {
    BaseExploreLayoutOptions,
    ExploreTable,
    FeatureFlag,
    GraphControls,
    GraphProgress,
    WebGLDisabledAlert,
    baseGraphLayouts,
    defaultGraphLayout,
    isNode,
    isWebGLEnabled,
    transformFlatGraphResponse,
    useCustomNodeKinds,
    useExploreSelectedItem,
    useExploreTableAutoDisplay,
    useFeatureFlag,
    useGraphHasData,
    useToggle,
} from 'bh-shared-ui';
import { MultiDirectedGraph } from 'graphology';
import { Attributes } from 'graphology-types';
import { GraphNodes } from 'js-client-library';
import isEmpty from 'lodash/isEmpty';
import { FC, useEffect, useRef, useState } from 'react';
import { SigmaNodeEventPayload } from 'sigma/sigma';
import { NoDataDialogWithLinks } from 'src/components/NoDataDialogWithLinks';
import SigmaChart from 'src/components/SigmaChart';
import { setExploreLayout, setIsExploreTableSelected } from 'src/ducks/global/actions';
import { useSigmaExploreGraph } from 'src/hooks/useSigmaExploreGraph';
import { useAppDispatch, useAppSelector } from 'src/store';
import { initGraph } from 'src/views/Explore/utils';
import ContextMenu from './ContextMenu/ContextMenu';
import ContextMenuZoneManagementEnabled from './ContextMenu/ContextMenuZoneManagementEnabled';
import ExploreSearch from './ExploreSearch/ExploreSearch';
import GraphItemInformationPanel from './GraphItemInformationPanel';
import { transformIconDictionary } from './svgIcons';

const GraphView: FC = () => {
    /* Hooks */
    const dispatch = useAppDispatch();

    const theme = useTheme();

    const { data: graphHasData, isLoading, isError } = useGraphHasData();

    const { selectedItem, setSelectedItem, selectedItemQuery } = useExploreSelectedItem();

    const [highlightedItem, setHighlightedItem] = useState<string | null>(selectedItem);
    const { data: tableViewFeatureFlag } = useFeatureFlag('explore_table_view');

    const darkMode = useAppSelector((state) => state.global.view.darkMode);

    const exploreLayout = useAppSelector((state) => state.global.view.exploreLayout);
    let isExploreTableSelected = useAppSelector((state) => state.global.view.isExploreTableSelected);

    if (!tableViewFeatureFlag?.enabled) {
        isExploreTableSelected = false;
    }

    const includeProperties = !!isExploreTableSelected;

    const graphQuery = useSigmaExploreGraph(includeProperties);

    const [graphologyGraph, setGraphologyGraph] = useState<MultiDirectedGraph<Attributes, Attributes, Attributes>>();

    const [currentNodes, setCurrentNodes] = useState<GraphNodes>({});

    const [contextMenu, setContextMenu] = useState<{ mouseX: number; mouseY: number } | null>(null);

    const [showNodeLabels, toggleShowNodeLabels] = useToggle(true);

    const [showEdgeLabels, toggleShowEdgeLabels] = useToggle(true);

    const [exportJsonData, setExportJsonData] = useState();

    const sigmaChartRef = useRef<any>(null);

    const customIcons = useCustomNodeKinds({ select: transformIconDictionary });

    const [autoDisplayTable, setAutoDisplayTable] = useExploreTableAutoDisplay({
        enabled: !exploreLayout,
    });
    const displayTable = autoDisplayTable || !!isExploreTableSelected;

    useEffect(() => {
        let items: any = graphQuery.data;

        if (!items && !graphQuery.isError) return;
        if (!items) items = {};

        // `items` may be empty, or it may contain an empty `nodes` object
        if (isEmpty(items) || isEmpty(items.nodes)) items = transformFlatGraphResponse(items);

        const graph = new MultiDirectedGraph();

        const hideNodes = displayTable;
        if (!hideNodes) initGraph(graph, items, theme, darkMode, customIcons.data ?? {}, hideNodes);
        setExportJsonData(items);

        setCurrentNodes(items.nodes);

        setGraphologyGraph(graph);
    }, [graphQuery.data, theme, darkMode, graphQuery.isError, customIcons.data, displayTable]);

    // Changes highlighted item when browser back/forward is used
    useEffect(() => {
        if (selectedItem && selectedItem !== highlightedItem) {
            setHighlightedItem(selectedItem);
        }
        // NOTE: Do not include `highlightedItem` as it will override ability to unselect highlights
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [selectedItem]);

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
    const selectItem = (id: string) => {
        setSelectedItem(id);
        setHighlightedItem(id);
    };

    const cancelHighlight = () => {
        setHighlightedItem(null);
    };

    const handleContextMenu = (event: SigmaNodeEventPayload) => {
        selectItem(event.node);
        setContextMenu(contextMenu === null ? { mouseX: event.event.x, mouseY: event.event.y } : null);
    };

    const handleCloseContextMenu = () => {
        setContextMenu(null);
    };

    const handleLayoutChange = (layout: BaseExploreLayoutOptions) => {
        if (layout === 'table') {
            dispatch(setIsExploreTableSelected(true));
            return;
        }

        dispatch(setExploreLayout(layout));
        dispatch(setIsExploreTableSelected(false));

        if (layout === 'standard') {
            sigmaChartRef.current?.runStandardLayout();
        } else if (layout === 'sequential') {
            sigmaChartRef.current?.runSequentialLayout();
        }
    };

    return (
        <div
            className='relative h-full w-full overflow-hidden'
            data-testid='explore'
            onContextMenu={(e) => e.preventDefault()}>
            <SigmaChart
                graph={graphologyGraph}
                highlightedItem={highlightedItem}
                onClickEdge={selectItem}
                onClickNode={selectItem}
                onClickStage={cancelHighlight}
                handleContextMenu={handleContextMenu}
                showNodeLabels={showNodeLabels}
                showEdgeLabels={showEdgeLabels}
                ref={sigmaChartRef}
            />

            <div className='absolute top-0 h-full p-4 flex gap-2 justify-between flex-col pointer-events-none'>
                <ExploreSearch />
                <GraphControls
                    layoutOptions={baseGraphLayouts}
                    selectedLayout={exploreLayout ?? defaultGraphLayout}
                    onLayoutChange={handleLayoutChange}
                    showNodeLabels={showNodeLabels}
                    onToggleNodeLabels={toggleShowNodeLabels}
                    showEdgeLabels={showEdgeLabels}
                    onToggleEdgeLabels={toggleShowEdgeLabels}
                    jsonData={exportJsonData}
                    onReset={() => sigmaChartRef.current?.resetCamera()}
                    currentNodes={currentNodes}
                    onSearchedNodeClick={(node) => {
                        selectItem(node.id);
                        sigmaChartRef?.current?.zoomTo(node.id);
                    }}
                />
            </div>
            <GraphItemInformationPanel />
            <FeatureFlag
                flagKey='tier_management_engine'
                enabled={
                    <ContextMenuZoneManagementEnabled
                        contextMenu={isNode(selectedItemQuery.data) ? contextMenu : null}
                        onClose={handleCloseContextMenu}
                    />
                }
                disabled={
                    <ContextMenu
                        contextMenu={isNode(selectedItemQuery.data) ? contextMenu : null}
                        handleClose={handleCloseContextMenu}
                    />
                }
            />

            <GraphProgress loading={graphQuery.isLoading} />
            <NoDataDialogWithLinks open={!graphHasData} />
            {tableViewFeatureFlag?.enabled && (
                <ExploreTable
                    data={graphQuery.data}
                    open={displayTable}
                    onClose={() => {
                        setAutoDisplayTable(false);
                        dispatch(setIsExploreTableSelected(false));
                    }}
                />
            )}
        </div>
    );
};

export default GraphView;
