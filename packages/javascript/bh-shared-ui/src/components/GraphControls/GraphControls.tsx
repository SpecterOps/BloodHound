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
    faArrowDown,
    faArrowLeft,
    faArrowRight,
    faArrowUp,
    faCropAlt,
    faDiagramProject,
    faDownload,
    faEye,
    faEyeSlash,
    faLeftRight,
    faMagnifyingGlass,
    faUpDown,
} from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { IconButton, MenuItem, Popover, PopoverContent, PopoverTrigger, Slider, Tooltip } from 'doodle-ui';
import capitalize from 'lodash/capitalize';
import isEmpty from 'lodash/isEmpty';
import { useCallback, useRef, useState } from 'react';
import { useExploreParams, useKeybindings } from '../../hooks';
import { cn } from '../../utils';
import { exportToJson } from '../../utils/exportGraphData';
import GraphMenu from '../GraphMenu';
import SearchCurrentNodes, { FlatNode } from '../SearchCurrentNodes';

export interface GraphExportAction {
    id: string;
    label: string;
    onSelect: () => void;
    disabled?: boolean;
}

export const graphDirections = ['down', 'left', 'right', 'up'] as const;
export type GraphDirection = (typeof graphDirections)[number];

export interface GraphLayoutControlOptions {
    direction?: GraphDirection;
    horizontalSpacing: number;
    onDirectionChange?: (direction: GraphDirection) => void;
    onHorizontalSpacingChange: (spacing: number) => void;
    onVerticalSpacingChange: (spacing: number) => void;
    verticalSpacing: number;
}

export const defaultGraphSpacing = 5;
export const maximumGraphSpacing = 10;
export const minimumGraphSpacing = 0;

const directionIcons = {
    down: faArrowDown,
    left: faArrowLeft,
    right: faArrowRight,
    up: faArrowUp,
};

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
    additionalExportActions?: readonly GraphExportAction[];
    layoutControls?: GraphLayoutControlOptions;
}

interface SpacingControlProps {
    axis: 'Horizontal' | 'Vertical';
    icon: typeof faLeftRight;
    onChange: (spacing: number) => void;
    value: number;
}

const SpacingControl = ({ axis, icon, onChange, value }: SpacingControlProps) => {
    const label = `${axis} spacing: ${value}`;

    return (
        <Popover>
            <Tooltip
                tooltip={<span>{label}</span>}
                triggerProps={{ asChild: true, className: 'pointer-events-auto' }}
                contentProps={{ className: 'dark:bg-neutral-4 dark:border-neutral-5 dark:text-white' }}>
                <PopoverTrigger asChild>
                    <IconButton
                        aria-label={label}
                        data-testid={`explore_graph-controls_${axis.toLowerCase()}-spacing-menu`}>
                        <FontAwesomeIcon aria-hidden='true' icon={icon} />
                    </IconButton>
                </PopoverTrigger>
            </Tooltip>
            <PopoverContent side='top' align='start' aria-label={`${axis} spacing`} className='w-auto p-1'>
                <div
                    className='box-border w-56 px-5 py-3'
                    data-testid={`explore_graph-${axis.toLowerCase()}-spacing-slider`}>
                    <Slider
                        thumbAriaLabel={`${axis} spacing`}
                        max={maximumGraphSpacing}
                        min={minimumGraphSpacing}
                        onValueChange={(nextValue) => onChange(nextValue)}
                        step={1}
                        value={value}
                    />
                    <div className='flex justify-between text-xs'>
                        <span>Compact</span>
                        <span>Spacious</span>
                    </div>
                </div>
            </PopoverContent>
        </Popover>
    );
};

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
        additionalExportActions = [],
        layoutControls,
    } = props;
    const { searchType } = useExploreParams();
    const [isCurrentSearchOpen, setIsCurrentSearchOpen] = useState(false);
    const searchButtonRef = useRef<HTMLButtonElement>(null);

    const closeCurrentSearch = useCallback(() => {
        setIsCurrentSearchOpen(false);
        // Restore focus to the search control after the panel closes
        requestAnimationFrame(() => searchButtonRef.current?.focus());
    }, []);

    useKeybindings({
        shift: {
            Slash: () => {
                if (isCurrentSearchOpen) {
                    closeCurrentSearch();
                } else {
                    setIsCurrentSearchOpen(true);
                }
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

    return (
        <div className='relative'>
            <div data-testid='explore_graph-controls' className='flex gap-1 pointer-events-auto'>
                <Tooltip
                    tooltip='Reset Graph'
                    triggerProps={{ className: 'pointer-events-auto' }}
                    contentProps={{ className: 'dark:bg-neutral-4 dark:border-neutral-5 dark:text-white' }}>
                    <div>
                        <IconButton
                            aria-label='Reset Graph'
                            onClick={onReset}
                            data-testid='explore_graph-controls_reset-button'>
                            <FontAwesomeIcon aria-hidden='true' icon={faCropAlt} />
                        </IconButton>
                    </div>
                </Tooltip>

                <GraphMenu
                    label={`${!showNodeLabels || !showEdgeLabels ? 'Show' : 'Hide'} Labels`}
                    icon={!showNodeLabels || !showEdgeLabels ? faEyeSlash : faEye}>
                    <MenuItem
                        aria-label={`${!showEdgeLabels ? 'Show' : 'Hide'} All Labels Toggle`}
                        data-testid='explore_graph-controls_all-labels-toggle'
                        onSelect={handleToggleAllLabels}>
                        {!showNodeLabels || !showEdgeLabels ? 'Show' : 'Hide'} All Labels
                    </MenuItem>
                    <MenuItem
                        aria-label={`${showNodeLabels ? 'Hide' : 'Show'} Node Labels Toggle`}
                        data-testid='explore_graph-controls_node-labels-toggle'
                        onSelect={onToggleNodeLabels}>
                        {showNodeLabels ? 'Hide' : 'Show'} Node Labels
                    </MenuItem>
                    <MenuItem
                        aria-label={`${showEdgeLabels ? 'Hide' : 'Show'} Edge Labels Toggle`}
                        data-testid='explore_graph-controls_edge-labels-toggle'
                        onSelect={onToggleEdgeLabels}>
                        {showEdgeLabels ? 'Hide' : 'Show'} Edge Labels
                    </MenuItem>
                </GraphMenu>

                <GraphMenu label='Layout' icon={faDiagramProject}>
                    {layoutOptions.map((buttonLabel) => {
                        const tableViewIsSelected = isExploreTableSelected && searchType === 'cypher';
                        const isSelected = tableViewIsSelected
                            ? buttonLabel === 'table' && isExploreLayoutSelected
                            : buttonLabel === selectedLayout && isExploreLayoutSelected;

                        return (
                            <MenuItem
                                data-testid={`explore_graph-controls_${buttonLabel}-buttonLabel`}
                                key={buttonLabel}
                                onSelect={() => onLayoutChange(buttonLabel)}
                                className={cn({ '!bg-primary !text-white dark:!text-neutral-1': isSelected })}>
                                {capitalize(buttonLabel)}
                            </MenuItem>
                        );
                    })}
                </GraphMenu>

                <GraphMenu label='Export' icon={faDownload}>
                    {additionalExportActions.map((action) => (
                        <MenuItem key={action.id} onSelect={action.onSelect} disabled={action.disabled}>
                            {action.label}
                        </MenuItem>
                    ))}
                    <MenuItem onSelect={() => exportToJson(jsonData)} disabled={isEmpty(jsonData)}>
                        JSON
                    </MenuItem>
                </GraphMenu>

                {layoutControls?.direction && layoutControls.onDirectionChange && (
                    <GraphMenu
                        label={`Graph direction: ${capitalize(layoutControls.direction)}`}
                        icon={directionIcons[layoutControls.direction]}>
                        {graphDirections.map((direction) => {
                            const isSelected = direction === layoutControls.direction;

                            return (
                                <MenuItem
                                    aria-checked={isSelected}
                                    className={cn({ '!bg-primary !text-white dark:!text-neutral-1': isSelected })}
                                    key={direction}
                                    onSelect={() => layoutControls.onDirectionChange?.(direction)}
                                    role='menuitemradio'>
                                    {capitalize(direction)}
                                </MenuItem>
                            );
                        })}
                    </GraphMenu>
                )}

                {layoutControls && (
                    <>
                        <SpacingControl
                            axis='Horizontal'
                            icon={faLeftRight}
                            onChange={layoutControls.onHorizontalSpacingChange}
                            value={layoutControls.horizontalSpacing}
                        />
                        <SpacingControl
                            axis='Vertical'
                            icon={faUpDown}
                            onChange={layoutControls.onVerticalSpacingChange}
                            value={layoutControls.verticalSpacing}
                        />
                    </>
                )}

                <Tooltip
                    tooltip='Search'
                    triggerProps={{ className: 'pointer-events-auto' }}
                    contentProps={{ className: 'dark:bg-neutral-4 dark:border-neutral-5 dark:text-white' }}>
                    <div>
                        <IconButton
                            ref={searchButtonRef}
                            aria-label='Search'
                            onClick={() => setIsCurrentSearchOpen(true)}
                            disabled={isCurrentSearchOpen}
                            data-testid='explore_graph-controls_search-current-results'>
                            <FontAwesomeIcon icon={faMagnifyingGlass} />
                        </IconButton>
                    </div>
                </Tooltip>
            </div>
            {isCurrentSearchOpen && (
                <div
                    role='search'
                    aria-label='Search Current Nodes'
                    className='absolute bottom-full left-0 w-[90%] z-[1] pointer-events-auto'
                    data-testid='explore_graph-controls_search-current-nodes-panel'>
                    <SearchCurrentNodes
                        className='p-2 mb-2'
                        currentNodes={currentNodes}
                        onSelect={(node) => {
                            onSearchedNodeClick(node);
                            closeCurrentSearch();
                        }}
                        onClose={closeCurrentSearch}
                    />
                </div>
            )}
        </div>
    );
}

export default GraphControls;
