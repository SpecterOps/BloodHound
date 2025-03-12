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
    GraphProgress,
    SearchCurrentNodes,
    WebGLDisabledAlert,
    exportToJson,
    isWebGLEnabled,
    useAvailableDomains,
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
import { GraphState } from 'src/ducks/explore/types';
import { setAssetGroupEdit } from 'src/ducks/global/actions';
import { GlobalOptionsState } from 'src/ducks/global/types';
import { discardChanges } from 'src/ducks/tierzero/actions';
import { useAppDispatch, useAppSelector } from 'src/store';
import { transformFlatGraphResponse } from 'src/utils';
import ExploreSearch from 'src/views/Explore/ExploreSearch';
import usePrompt from 'src/views/Explore/NavigationAlert';
import { initGraph } from 'src/views/Explore/utils';
import ContextMenu from './ContextMenu/ContextMenu';
import GraphItemInformationPanel from './GraphItemInformationPanel';

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

    const [showNodeLabels, setShowNodeLabels] = useState(true);

    const [showEdgeLabels, setShowEdgeLabels] = useState(true);

    const { setSelectedItem } = useExploreSelectedItem();

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
                <ExploreSearch />
                <div className='flex gap-1 pointer-events-auto' ref={currentSearchAnchorElement}>
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
            <ContextMenu contextMenu={contextMenu} handleClose={handleCloseContextMenu} />
            <GraphProgress loading={graphState.loading} />
            <NoDataDialogWithLinks open={!data?.length} />
        </div>
    );
};

export default GraphView;
