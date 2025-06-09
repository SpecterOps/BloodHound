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

import { Popper, useTheme } from '@mui/material';
import {
    BaseGraphLayoutOptions,
    ExploreTable,
    GraphProgress,
    SearchCurrentNodes,
    WebGLDisabledAlert,
    exportToJson,
    isWebGLEnabled,
    transformFlatGraphResponse,
    useCustomNodeKinds,
    useExploreSelectedItem,
    useGraphHasData,
    useToggle,
} from 'bh-shared-ui';
import { MultiDirectedGraph } from 'graphology';
import { Attributes } from 'graphology-types';
import { GraphNodes } from 'js-client-library';
import isEmpty from 'lodash/isEmpty';
import { FC, useEffect, useRef, useState } from 'react';
import { SigmaNodeEventPayload } from 'sigma/sigma';
import GraphButtons from 'src/components/GraphButtons';
import { NoDataDialogWithLinks } from 'src/components/NoDataDialogWithLinks';
import SigmaChart from 'src/components/SigmaChart';
import { setExploreLayout } from 'src/ducks/global/actions';
import { useSigmaExploreGraph } from 'src/hooks/useSigmaExploreGraph';
import { useAppDispatch, useAppSelector } from 'src/store';
import { initGraph } from 'src/views/Explore/utils';
import ContextMenu from './ContextMenu/ContextMenu';
import ExploreSearch from './ExploreSearch/ExploreSearch';
import GraphItemInformationPanel from './GraphItemInformationPanel';
import { transformIconDictionary } from './svgIcons';

const GraphView: FC = () => {
    /* Hooks */
    const dispatch = useAppDispatch();
    const theme = useTheme();

    const graphQuery = useSigmaExploreGraph();
    const { data: graphHasData, isLoading, isError } = useGraphHasData();
    const { selectedItem, setSelectedItem } = useExploreSelectedItem();
    const [highlightedItem, setHighlightedItem] = useState<string | null>(selectedItem);

    const darkMode = useAppSelector((state) => state.global.view.darkMode);
    const exploreLayout = useAppSelector((state) => state.global.view.exploreLayout);

    const [graphologyGraph, setGraphologyGraph] = useState<MultiDirectedGraph<Attributes, Attributes, Attributes>>();
    const [currentNodes, setCurrentNodes] = useState<GraphNodes>({});
    const [contextMenu, setContextMenu] = useState<{ mouseX: number; mouseY: number } | null>(null);
    const [exportJsonData, setExportJsonData] = useState();
    const [showNodeLabels, setShowNodeLabels] = useState(true);
    const [showEdgeLabels, setShowEdgeLabels] = useState(true);

    const sigmaChartRef = useRef<any>(null);

    const customIcons = useCustomNodeKinds({ select: transformIconDictionary });

    const [currentSearchOpen, toggleCurrentSearch] = useToggle(false);

    const currentSearchAnchorElement = useRef(null);

    useEffect(() => {
        let items: any = graphQuery.data;

        if (!items && !graphQuery.isError) return;
        if (!items) items = {};

        // `items` may be empty, or it may contain an empty `nodes` object
        if (isEmpty(items) || isEmpty(items.nodes)) items = transformFlatGraphResponse(items);

        const graph = new MultiDirectedGraph();

        initGraph(graph, items, theme, darkMode, customIcons.data ?? {});
        setExportJsonData(items);

        setCurrentNodes(items.nodes);

        setGraphologyGraph(graph);
    }, [graphQuery.data, theme, darkMode, graphQuery.isError, customIcons.data]);

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
        setContextMenu(contextMenu === null ? { mouseX: event.event.x, mouseY: event.event.y } : null);
        selectItem(event.node);
    };

    const handleCloseContextMenu = () => {
        setContextMenu(null);
    };

    const setAllNodesToHidden = (hidden: boolean) => {
        graphologyGraph?.updateEachNodeAttributes((node, attr) => {
            return { ...attr, hidden };
        });
    };

    const handleLayoutChange = (layout: BaseGraphLayoutOptions) => {
        dispatch(setExploreLayout(layout));

        if (layout === 'standard') {
            sigmaChartRef.current?.runStandardLayout();
        } else if (layout === 'sequential') {
            sigmaChartRef.current?.runSequentialLayout();
        } else if (layout === 'table') {
            setAllNodesToHidden(true);
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
                <div className='flex gap-1 pointer-events-auto' ref={currentSearchAnchorElement}>
                    <GraphButtons
                        onExportJson={() => {
                            exportToJson(exportJsonData);
                        }}
                        onReset={() => {
                            sigmaChartRef.current?.resetCamera();
                        }}
                        onRunSequentialLayout={() => {
                            dispatch(setExploreLayout('sequential'));
                            sigmaChartRef.current?.runSequentialLayout();
                        }}
                        onRunStandardLayout={() => {
                            dispatch(setExploreLayout('standard'));
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
                                selectItem(node.id);
                                sigmaChartRef?.current?.zoomTo(node.id);
                                toggleCurrentSearch?.();
                            }}
                            onClose={toggleCurrentSearch}
                        />
                    </div>
                </Popper>
            </div>
            <GraphItemInformationPanel />
            <ContextMenu contextMenu={contextMenu} handleClose={handleCloseContextMenu} />
            <GraphProgress loading={graphQuery.isLoading} />
            <NoDataDialogWithLinks open={!graphHasData} />
            <ExploreTable
                open={exploreLayout === 'table'}
                onClose={() => {
                    handleLayoutChange('sequential');
                    setAllNodesToHidden(false);
                }}
            />
        </div>
    );
};

export default GraphView;
