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

import {
    TooltipContent,
    TooltipPortal,
    TooltipProvider,
    TooltipRoot,
    TooltipTrigger,
} from '@bloodhoundenterprise/doodleui';
import { faCropAlt } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { MenuItem, Popper } from '@mui/material';
import { GraphNodes } from 'js-client-library';
import capitalize from 'lodash/capitalize';
import isEmpty from 'lodash/isEmpty';
import { useRef, useState } from 'react';
import { useFeatureFlag } from '../../hooks/useFeatureFlags';
import { exportToJson } from '../../utils/exportGraphData';
import GraphButton from '../GraphButton';
import GraphMenu from '../GraphMenu';
import SearchCurrentNodes, { FlatNode } from '../SearchCurrentNodes';

interface GraphControlsProps<T extends readonly string[]> {
    onReset: () => void;
    onLayoutChange: (layout: T[number]) => void;
    onToggleNodeLabels: () => void;
    onToggleEdgeLabels: () => void;
    onSearchedNodeClick: (node: FlatNode) => void;
    layoutOptions: T;
    selectedLayout: T[number];
    showNodeLabels: boolean;
    showEdgeLabels: boolean;
    jsonData: Record<string, any> | undefined;
    currentNodes: GraphNodes;
}

function GraphControls<T extends readonly string[]>(props: GraphControlsProps<T>) {
    const {
        onReset,
        onLayoutChange,
        onToggleNodeLabels,
        onToggleEdgeLabels,
        onSearchedNodeClick,
        layoutOptions,
        selectedLayout,
        showNodeLabels,
        showEdgeLabels,
        jsonData,
        currentNodes,
    } = props;

    const { data: featureFlag } = useFeatureFlag('explore_table_view');

    const [isCurrentSearchOpen, setIsCurrentSearchOpen] = useState(false);

    const currentSearchAnchorElement = useRef(null);

    const handleToggleAllLabels = () => {
        if (showNodeLabels && showEdgeLabels) {
            // Hide All
            onToggleNodeLabels();
            onToggleEdgeLabels();
        } else {
            // Show All
            if (!showNodeLabels) onToggleNodeLabels();
            if (!showEdgeLabels) onToggleEdgeLabels();
        }
    };

    return (
        <>
            <div
                data-testid='explore_graph-controls'
                className='flex gap-1 pointer-events-auto'
                ref={currentSearchAnchorElement}>
                <TooltipProvider>
                    <TooltipRoot>
                        <TooltipTrigger className='pointer-events-auto'>
                            {/* tooltip won't show without this wrapper div for some reason */}
                            <div>
                                <GraphButton
                                    aria-label='Reset Graph'
                                    onClick={onReset}
                                    displayText={<FontAwesomeIcon aria-label='reset graph view' icon={faCropAlt} />}
                                    data-testid='explore_graph-controls_reset-button'
                                />
                            </div>
                        </TooltipTrigger>
                        <TooltipPortal>
                            <TooltipContent className='dark:bg-neutral-dark-5 border-0'>
                                <span>Reset Graph</span>
                            </TooltipContent>
                        </TooltipPortal>
                    </TooltipRoot>
                </TooltipProvider>

                <GraphMenu label={'Hide Labels'}>
                    <MenuItem
                        aria-label='All Labels Toggle'
                        data-testid='explore_graph-controls_all-labels-toggle'
                        onClick={handleToggleAllLabels}>
                        {!showNodeLabels || !showEdgeLabels ? 'Show' : 'Hide'} All Labels
                    </MenuItem>
                    <MenuItem
                        aria-label='Node Labels Toggle'
                        data-testid='explore_graph-controls_node-labels-toggle'
                        onClick={onToggleNodeLabels}>
                        {showNodeLabels ? 'Hide' : 'Show'} Node Labels
                    </MenuItem>
                    <MenuItem
                        aria-label='Edge Labels Toggle'
                        data-testid='explore_graph-controls_edge-labels-toggle'
                        onClick={onToggleEdgeLabels}>
                        {showEdgeLabels ? 'Hide' : 'Show'} Edge Labels
                    </MenuItem>
                </GraphMenu>

                <GraphMenu label='Layout'>
                    {layoutOptions
                        .filter((layout) => {
                            if (!featureFlag?.enabled) {
                                return layout !== 'table';
                            }
                            return true;
                        })
                        .map((layout) => (
                            <MenuItem
                                data-testid={`explore_graph-controls_${layout}-layout`}
                                key={layout}
                                selected={featureFlag?.enabled ? selectedLayout === layout : undefined}
                                onClick={() => onLayoutChange(layout)}>
                                {capitalize(layout)}
                            </MenuItem>
                        ))}
                </GraphMenu>

                <GraphMenu label='Export'>
                    <MenuItem onClick={() => exportToJson(jsonData)} disabled={isEmpty(jsonData)}>
                        JSON
                    </MenuItem>
                </GraphMenu>

                <GraphButton
                    aria-label='Search Current Results'
                    onClick={() => setIsCurrentSearchOpen(true)}
                    displayText={'Search Current Results'}
                    disabled={isCurrentSearchOpen}
                    data-testid='explore_graph-controls_search-current-results'
                />
            </div>
            <Popper
                open={isCurrentSearchOpen}
                anchorEl={currentSearchAnchorElement.current}
                placement='top'
                disablePortal
                className='w-[90%] z-[1]'>
                <div className='pointer-events-auto' data-testid='explore_graph-controls_search-current-nodes-popper'>
                    <SearchCurrentNodes
                        sx={{
                            padding: 1,
                            marginBottom: 1,
                        }}
                        currentNodes={currentNodes}
                        onSelect={(node) => {
                            onSearchedNodeClick(node);
                            setIsCurrentSearchOpen(false);
                        }}
                        onClose={() => setIsCurrentSearchOpen(false)}
                    />
                </div>
            </Popper>
        </>
    );
}

export default GraphControls;
