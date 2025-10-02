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

import { faChevronLeft, faChevronRight } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { Button, IconButton, SvgIcon, useTheme } from '@mui/material';
import {
    BaseExploreLayoutOptions,
    ExploreTable,
    GraphControls,
    GraphProgress,
    GraphViewErrorAlert,
    ManageColumnsComboBoxOption,
    NodeClickInfo,
    WebGLDisabledAlert,
    baseGraphLayouts,
    defaultGraphLayout,
    glyphUtils,
    isWebGLEnabled,
    makeStoreMapFromColumnOptions,
    useCustomNodeKinds,
    useExploreParams,
    useExploreSelectedItem,
    useExploreTableAutoDisplay,
    useGraphHasData,
    useRollbackQuery,
    useTagGlyphs,
    useToggle,
} from 'bh-shared-ui';
import { MultiDirectedGraph } from 'graphology';
import { Attributes } from 'graphology-types';
import { FC, useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { SigmaEventPayload, SigmaNodeEventPayload } from 'sigma/sigma';
import { NoDataFileUploadDialogWithLinks } from 'src/components/NoDataFileUploadDialogWithLinks';
import SigmaChart from 'src/components/SigmaChart';
import { setExploreLayout, setIsExploreTableSelected, setSelectedExploreTableColumns } from 'src/ducks/global/actions';
import { useSigmaExploreGraph } from 'src/hooks/useSigmaExploreGraph';
import { useAppDispatch, useAppSelector } from 'src/store';
import { initGraph } from 'src/views/Explore/utils';
import ContextMenu from '../ContextMenu/ContextMenu';
import ExploreSearch from '../ExploreSearch/ExploreSearch';
import GraphItemInformationPanel from '../GraphItemInformationPanel';
import { transformIconDictionary } from '../svgIcons';

type Entry = {
    id: number;
    created_at?: string;
    updated_at?: string;
    deleted_at?: {
        Time: string;
        Valid: boolean;
    };
    change_type: string;
    object_type: string;
    object_id: string;
    source_object_id: string;
    target_object_id: string;
    properties: string;
    rolled_back_at?: string;
};

const GraphView: FC = () => {
    /* Hooks */
    const dispatch = useAppDispatch();

    const theme = useTheme();

    const graphHasDataQuery = useGraphHasData();
    const graphQuery = useSigmaExploreGraph();

    const { searchType } = useExploreParams();
    // const { refetch } = useExploreGraph();
    const { selectedItem, setSelectedItem, clearSelectedItem } = useExploreSelectedItem();

    const [graphologyGraph, setGraphologyGraph] = useState<MultiDirectedGraph<Attributes, Attributes, Attributes>>();
    const [contextMenu, setContextMenu] = useState<{ mouseX: number; mouseY: number } | null>(null);
    const [currentRollbackIndex, setCurrentRollbackIndex] = useState<number>();

    const sigmaChartRef = useRef<any>(null);

    const darkMode = useAppSelector((state) => state.global.view.darkMode);
    const exploreLayout = useAppSelector((state) => state.global.view.exploreLayout);
    const selectedColumns = useAppSelector((state) => state.global.view.selectedExploreTableColumns);
    const isExploreTableSelected = useAppSelector((state) => state.global.view.isExploreTableSelected);

    const customIconsQuery = useCustomNodeKinds({ select: transformIconDictionary });
    const tagGlyphMap = useTagGlyphs(glyphUtils, darkMode);
    const rollbackEnabled = graphQuery.data;
    const { data: _rollbacks } = useRollbackQuery(rollbackEnabled);
    const rollbacks: { count?: number; entries?: Entry[] } = _rollbacks;
    // const { mutateAsync: setRollback } = useRollbackMutation();

    // console.log(rollbacks);
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

    console.log({ rollbacks });
    // useEffect(() => {
    //     if (currentRollbackIndex) {
    //         const rollback = rollbacks[currentRollbackIndex];

    //         setRollback(rollback.id).then(refetch);
    //     }
    // }, [currentRollbackIndex, refetch, rollbacks, setRollback]);

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
        (event: SigmaNodeEventPayload | SigmaEventPayload) => {
            if ('node' in event) {
                setSelectedItem(event.node);
            }
            setContextMenu({ mouseX: event.event.x, mouseY: event.event.y });
        },
        [setContextMenu, setSelectedItem]
    );

    /* Passthrough function to munge shared component callback shape into a Sigma Node event-shaped object */
    const handleKebabMenuClick = useCallback(
        (nodeInfo: NodeClickInfo) => {
            if (!nodeInfo.id) return;
            handleContextMenu({ event: { x: nodeInfo.x, y: nodeInfo.y }, node: nodeInfo.id } as SigmaNodeEventPayload);
        },
        [handleContextMenu]
    );

    if (graphHasDataQuery.isLoading) {
        return (
            <div className='relative h-full w-full overflow-hidden' data-testid='explore'>
                <GraphProgress loading={graphHasDataQuery.isLoading} />
            </div>
        );
    }

    if (graphHasDataQuery.isError) return <GraphViewErrorAlert />;

    if (!isWebGLEnabledMemo) return <WebGLDisabledAlert />;

    /* Event Handlers */
    const handleCloseContextMenu = () => {
        setContextMenu(null);
    };

    const handleBackRollback = () => {
        if (typeof currentRollbackIndex === 'number' && currentRollbackIndex > 0) {
            setCurrentRollbackIndex(currentRollbackIndex - 1);
        }
    };

    const handleForwardRollback = () => {
        if (typeof currentRollbackIndex === 'number' && currentRollbackIndex > 0) {
            setCurrentRollbackIndex(currentRollbackIndex + 1);
        }
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
                <div className='border-neutral-500 bg-neutral-100 p-2 w-fit'>
                    <div>
                        <div className='flex justify-center items-center'>
                            <p className='font-bold'>Time travel</p>
                            <div className='flex'>
                                <IconButton
                                    title={'Tick back'}
                                    onClick={handleBackRollback}
                                    disabled={!currentRollbackIndex}>
                                    <SvgIcon>
                                        <FontAwesomeIcon icon={faChevronLeft} />
                                    </SvgIcon>
                                </IconButton>
                                <IconButton
                                    title={'Tick forward'}
                                    onClick={handleForwardRollback}
                                    disabled={typeof currentRollbackIndex !== 'number'}>
                                    <SvgIcon>
                                        <FontAwesomeIcon icon={faChevronRight} />
                                    </SvgIcon>
                                </IconButton>
                            </div>
                        </div>
                        <div className='flex'>
                            {rollbacks?.entries?.map((entry) => (
                                <Button className='flex flex-col' key={entry.id + entry.object_id}>
                                    <p> {entry.change_type}</p>
                                    <p>{entry.object_id.slice(0, 10)}</p>
                                </Button>
                            ))}
                        </div>
                    </div>
                </div>
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
            <ContextMenu contextMenu={contextMenu} handleClose={handleCloseContextMenu} />
            <GraphProgress loading={graphQuery.isLoading} />
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
