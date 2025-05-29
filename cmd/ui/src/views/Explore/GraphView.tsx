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
    GraphControls,
    GraphProgress,
    TableView,
    WebGLDisabledAlert,
    isWebGLEnabled,
    transformFlatGraphResponse,
    useAvailableEnvironments,
    useCustomNodeKinds,
    useExploreSelectedItem,
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
import { useSigmaExploreGraph } from 'src/hooks/useSigmaExploreGraph';
import { useAppSelector } from 'src/store';
import { initGraph } from 'src/views/Explore/utils';
import ContextMenu from './ContextMenu/ContextMenu';
import ExploreSearch from './ExploreSearch/ExploreSearch';
import GraphItemInformationPanel from './GraphItemInformationPanel';
import { transformIconDictionary } from './svgIcons';

const layoutOptions = ['sequential', 'standard', 'table'] as const;
type LayoutOption = (typeof layoutOptions)[number];

const GraphView: FC = () => {
    /* Hooks */
    const theme = useTheme();

    const graphQuery = useSigmaExploreGraph();
    const { data, isLoading, isError } = useAvailableEnvironments();
    const { setSelectedItem } = useExploreSelectedItem();

    const darkMode = useAppSelector((state) => state.global.view.darkMode);

    const [graphologyGraph, setGraphologyGraph] = useState<MultiDirectedGraph<Attributes, Attributes, Attributes>>();
    const [currentNodes, setCurrentNodes] = useState<GraphNodes>({});
    const [selectedLayout, setSelectedLayout] = useState<LayoutOption>('sequential');
    const [contextMenu, setContextMenu] = useState<{ mouseX: number; mouseY: number } | null>(null);
    const [showNodeLabels, toggleShowNodeLabels] = useToggle(true);
    const [showEdgeLabels, toggleShowEdgeLabels] = useToggle(true);
    const [exportJsonData, setExportJsonData] = useState();

    const sigmaChartRef = useRef<any>(null);

    const customIcons = useCustomNodeKinds({ select: transformIconDictionary });

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

    const handleLayoutChange = (layout: LayoutOption) => {
        setSelectedLayout(layout);
        if (layout === 'standard') {
            sigmaChartRef.current?.runStandardLayout();
        } else if (layout === 'sequential') {
            sigmaChartRef.current?.runSequentialLayout();
        } else if (layout === 'table') {
            graphologyGraph?.updateEachNodeAttributes((node, attr) => {
                return { ...attr, hidden: true };
            });
        }
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
                <GraphControls
                    layoutOptions={layoutOptions}
                    selectedLayout={selectedLayout}
                    onLayoutChange={handleLayoutChange}
                    showNodeLabels={showNodeLabels}
                    onToggleNodeLabels={toggleShowNodeLabels}
                    showEdgeLabels={showEdgeLabels}
                    onToggleEdgeLabels={toggleShowEdgeLabels}
                    jsonData={exportJsonData}
                    onReset={sigmaChartRef.current?.resetCamera}
                    currentNodes={currentNodes}
                />
            </div>
            <GraphItemInformationPanel />
            <ContextMenu contextMenu={contextMenu} handleClose={handleCloseContextMenu} />
            <GraphProgress loading={graphQuery.isLoading} />
            <NoDataDialogWithLinks open={!data?.length} />
            <TableView
                open={selectedLayout === 'table'}
                onClose={() => {
                    handleLayoutChange('sequential');

                    graphologyGraph?.updateEachNodeAttributes((node, attr) => {
                        return { ...attr, hidden: false };
                    });
                }}
            />
        </div>
    );
};

/**
 * TODO:
 *
 * show placeholder when no nodes are present and we are viewing the table
 * where do we throw in the empty node check?
 *   same thing for BHE
 */

export default GraphView;
