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
    faCropAlt,
    faDiagramProject,
    faDownload,
    faEye,
    faEyeSlash,
    faMagnifyingGlass,
} from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { MenuItem, Popper, Tooltip } from '@mui/material';
import capitalize from 'lodash/capitalize';
import isEmpty from 'lodash/isEmpty';
import { useRef, useState } from 'react';
import { useExploreParams, useKeybindings } from '../../hooks';
import { cn } from '../../utils';
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
    isExploreTableSelected?: boolean;
    isExploreLayoutSelected?: boolean;
    layoutOptions: T;
    selectedLayout?: T[number];
    showNodeLabels: boolean;
    showEdgeLabels: boolean;
    jsonData: Record<string, any> | undefined;
    currentNodes: Record<string, any> | undefined;
}

function GraphControls<T extends readonly string[]>(props: GraphControlsProps<T>) {
    const {
        onReset,
        onLayoutChange,
        onToggleNodeLabels,
        onToggleEdgeLabels,
        onSearchedNodeClick,
        isExploreTableSelected,
        isExploreLayoutSelected,
        layoutOptions,
        selectedLayout,
        showNodeLabels,
        showEdgeLabels,
        jsonData,
        currentNodes = {},
    } = props;
    const { searchType } = useExploreParams();
    const [isCurrentSearchOpen, setIsCurrentSearchOpen] = useState(false);

    const currentSearchAnchorElement = useRef(null);
    useKeybindings({
        shift: {
            Slash: () => {
                setIsCurrentSearchOpen(!isCurrentSearchOpen);
            },
        },
        KeyG: onReset,
    });

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

    const searchButton = (
        <GraphButton
            aria-label='Search'
            onClick={() => setIsCurrentSearchOpen(true)}
            displayText={<FontAwesomeIcon aria-hidden='true' icon={faMagnifyingGlass} />}
            disabled={isCurrentSearchOpen}
            data-testid='explore_graph-controls_search-current-results'
        />
    );

    return (
        <>
            <div
                data-testid='explore_graph-controls'
                className='flex gap-1 pointer-events-auto'
                ref={currentSearchAnchorElement}>
                <Tooltip placement='top' title='Reset Graph'>
                    <GraphButton
                        aria-label='Reset Graph'
                        onClick={onReset}
                        displayText={<FontAwesomeIcon aria-hidden='true' icon={faCropAlt} />}
                        data-testid='explore_graph-controls_reset-button'
                    />
                </Tooltip>

                <GraphMenu
                    controlId='labels'
                    displayText={
                        <FontAwesomeIcon
                            aria-hidden='true'
                            icon={!showNodeLabels || !showEdgeLabels ? faEyeSlash : faEye}
                        />
                    }
                    label={`${!showNodeLabels || !showEdgeLabels ? 'Show' : 'Hide'} Labels`}>
                    <MenuItem
                        aria-label={`${!showEdgeLabels ? 'Show' : 'Hide'} All Labels Toggle`}
                        data-testid='explore_graph-controls_all-labels-toggle'
                        onClick={handleToggleAllLabels}>
                        {!showNodeLabels || !showEdgeLabels ? 'Show' : 'Hide'} All Labels
                    </MenuItem>
                    <MenuItem
                        aria-label={`${showNodeLabels ? 'Hide' : 'Show'} Node Labels Toggle`}
                        data-testid='explore_graph-controls_node-labels-toggle'
                        onClick={onToggleNodeLabels}>
                        {showNodeLabels ? 'Hide' : 'Show'} Node Labels
                    </MenuItem>
                    <MenuItem
                        aria-label={`${showEdgeLabels ? 'Hide' : 'Show'} Edge Labels Toggle`}
                        data-testid='explore_graph-controls_edge-labels-toggle'
                        onClick={onToggleEdgeLabels}>
                        {showEdgeLabels ? 'Hide' : 'Show'} Edge Labels
                    </MenuItem>
                </GraphMenu>

                <GraphMenu
                    controlId='layout'
                    displayText={<FontAwesomeIcon aria-hidden='true' icon={faDiagramProject} />}
                    label='Layout'>
                    {layoutOptions.map((buttonLabel) => {
                        const tableViewIsSelected = isExploreTableSelected && searchType === 'cypher';
                        const isSelected = tableViewIsSelected
                            ? buttonLabel === 'table' && isExploreLayoutSelected
                            : buttonLabel === selectedLayout && isExploreLayoutSelected;

                        return (
                            <MenuItem
                                data-testid={`explore_graph-controls_${buttonLabel}-buttonLabel`}
                                key={buttonLabel}
                                selected={isSelected}
                                onClick={() => onLayoutChange(buttonLabel)}
                                className={cn({ '!bg-primary text-white dark:text-[#121212]': isSelected })}>
                                {capitalize(buttonLabel)}
                            </MenuItem>
                        );
                    })}
                </GraphMenu>

                <GraphMenu
                    controlId='export'
                    displayText={<FontAwesomeIcon aria-hidden='true' icon={faDownload} />}
                    label='Export'>
                    <MenuItem onClick={() => exportToJson(jsonData)} disabled={isEmpty(jsonData)}>
                        JSON
                    </MenuItem>
                </GraphMenu>

                <Tooltip placement='top' title='Search'>
                    {searchButton}
                </Tooltip>
            </div>
            <Popper
                open={isCurrentSearchOpen}
                anchorEl={currentSearchAnchorElement.current}
                placement='top'
                disablePortal
                aria-label='Search Current Nodes'
                className='w-[90%] z-[1]'>
                <div className='pointer-events-auto' data-testid='explore_graph-controls_search-current-nodes-popper'>
                    <SearchCurrentNodes
                        className='p-2 mb-2'
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
