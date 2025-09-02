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
    GraphViewErrorAlert,
    ManageColumnsComboBoxOption,
    NodeClickInfo,
    WebGLDisabledAlert,
    baseGraphLayouts,
    defaultGraphLayout,
    glyphUtils,
    isNode,
    isWebGLEnabled,
    makeStoreMapFromColumnOptions,
    useCustomNodeKinds,
    useExploreParams,
    useExploreSelectedItem,
    useExploreTableAutoDisplay,
    useGraphHasData,
    useTagGlyphs,
    useToggle,
} from 'bh-shared-ui';
import { MultiDirectedGraph } from 'graphology';
import { Attributes } from 'graphology-types';
import { FC, useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { SigmaNodeEventPayload } from 'sigma/sigma';
import { NoDataFileUploadDialogWithLinks } from 'src/components/NoDataFileUploadDialogWithLinks';
import SigmaChart from 'src/components/SigmaChart';
import { setExploreLayout, setIsExploreTableSelected, setSelectedExploreTableColumns } from 'src/ducks/global/actions';
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

    const graphHasDataQuery = useGraphHasData();
    const graphQuery = useSigmaExploreGraph();

    const { searchType } = useExploreParams();
    const { selectedItem, setSelectedItem, selectedItemQuery, clearSelectedItem } = useExploreSelectedItem();

    const [graphologyGraph, setGraphologyGraph] = useState<MultiDirectedGraph<Attributes, Attributes, Attributes>>();
    const [contextMenu, setContextMenu] = useState<{ mouseX: number; mouseY: number } | null>(null);

    const sigmaChartRef = useRef<any>(null);

    const darkMode = useAppSelector((state) => state.global.view.darkMode);
    const exploreLayout = useAppSelector((state) => state.global.view.exploreLayout);
    const selectedColumns = useAppSelector((state) => state.global.view.selectedExploreTableColumns);
    const isExploreTableSelected = useAppSelector((state) => state.global.view.isExploreTableSelected);

    const customIconsQuery = useCustomNodeKinds({ select: transformIconDictionary });
    const tagGlyphMap = useTagGlyphs(glyphUtils, darkMode);

    const autoDisplayTableEnabled = !exploreLayout && !isExploreTableSelected;
    const [autoDisplayTable, setAutoDisplayTable] = useExploreTableAutoDisplay(autoDisplayTableEnabled);
    // TODO: incorporate into larger hook with auto display table logic
    const displayTable = searchType === 'cypher' && (isExploreTableSelected || autoDisplayTable);

    const [showNodeLabels, toggleShowNodeLabels] = useToggle(true);
    const [showEdgeLabels, toggleShowEdgeLabels] = useToggle(true);

    const isWebGLEnabledMemo = useMemo(() => isWebGLEnabled(), []);

    const graphOptions = useMemo(() => {
        return {
            theme,
            darkMode,
            customIcons: customIconsQuery?.data ?? {},
            hideNodes: displayTable,
            tagGlyphMap,
        };
    }, [theme, darkMode, customIconsQuery.data, displayTable, tagGlyphMap]);

    // Initialize graph data for rendering with sigmajs
    useEffect(() => {
        if (!graphQuery.data) return;

        const graphData = graphQuery.data;

        if (graphOptions.hideNodes) {
            setGraphologyGraph(undefined);
        } else {
            const graph = initGraph(graphData, graphOptions);
            setGraphologyGraph(graph);
        }
    }, [graphQuery.data, graphOptions]);

    /* useCallback Event Handlers must appear before return statement */
    const handleContextMenu = useCallback(
        (event: SigmaNodeEventPayload) => {
            setSelectedItem(event.node);
            setContextMenu(contextMenu === null ? { mouseX: event.event.x, mouseY: event.event.y } : null);
        },
        [contextMenu, setContextMenu, setSelectedItem]
    );

    /* Passthrough function to munge shared component callback shape into a Sigma Node event-shaped object */
    const handleKebabMenuClick = useCallback(
        (nodeInfo: NodeClickInfo) => {
            if (!nodeInfo.id) return;
            handleContextMenu({ event: { x: nodeInfo.x, y: nodeInfo.y }, node: nodeInfo.id } as SigmaNodeEventPayload);
        },
        [handleContextMenu]
    );

    const isLoading = graphHasDataQuery.isLoading || graphQuery.isLoading || customIconsQuery.isLoading;
    const isError = graphHasDataQuery.isError || graphQuery.isError || customIconsQuery.isError;

    if (isLoading) {
        return (
            <div className='relative h-full w-full overflow-hidden' data-testid='explore'>
                <GraphProgress loading={isLoading} />
            </div>
        );
    }

    if (isError) return <GraphViewErrorAlert />;

    if (!isWebGLEnabledMemo) return <WebGLDisabledAlert />;

    /* Event Handlers */
    const handleCloseContextMenu = () => {
        setContextMenu(null);
    };

    const handleManageColumnsChange = (columnOptions: ManageColumnsComboBoxOption[]) => {
        const newItems = makeStoreMapFromColumnOptions(columnOptions);

        dispatch(setSelectedExploreTableColumns(newItems));
    };

    const handleLayoutChange = (layout: BaseExploreLayoutOptions) => {
        if (layout === 'table') {
            dispatch(setIsExploreTableSelected(true));
            return;
        }

        dispatch(setExploreLayout(layout));
        dispatch(setIsExploreTableSelected(false));

        if (layout === 'standard') sigmaChartRef.current?.runStandardLayout();

        if (layout === 'sequential') sigmaChartRef.current?.runSequentialLayout();
    };

    return (
        <div
            className='relative h-full w-full overflow-hidden'
            data-testid='explore'
            onContextMenu={(e) => e.preventDefault()}>
            <SigmaChart
                graph={graphologyGraph}
                highlightedItem={selectedItem}
                onClickEdge={setSelectedItem}
                onClickNode={setSelectedItem}
                onClickStage={clearSelectedItem}
                handleContextMenu={handleContextMenu}
                showNodeLabels={showNodeLabels}
                showEdgeLabels={showEdgeLabels}
                ref={sigmaChartRef}
            />

            <div className='absolute top-0 h-full p-4 flex gap-2 justify-between flex-col pointer-events-none'>
                <ExploreSearch />
                <GraphControls
                    isExploreTableSelected={isExploreTableSelected}
                    layoutOptions={baseGraphLayouts}
                    selectedLayout={exploreLayout ?? defaultGraphLayout}
                    onLayoutChange={handleLayoutChange}
                    showNodeLabels={showNodeLabels}
                    onToggleNodeLabels={toggleShowNodeLabels}
                    showEdgeLabels={showEdgeLabels}
                    onToggleEdgeLabels={toggleShowEdgeLabels}
                    jsonData={graphQuery.data}
                    onReset={() => sigmaChartRef.current?.resetCamera()}
                    currentNodes={graphQuery.data?.nodes}
                    onSearchedNodeClick={(node) => {
                        node.id && setSelectedItem(node.id);
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

            <GraphProgress loading={isLoading} />
            <NoDataFileUploadDialogWithLinks open={!graphHasDataQuery.data} />
            {displayTable && (
                <ExploreTable
                    selectedColumns={selectedColumns}
                    onManageColumnsChange={handleManageColumnsChange}
                    onKebabMenuClick={handleKebabMenuClick}
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
