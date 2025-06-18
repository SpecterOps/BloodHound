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
            <div className='flex gap-1 pointer-events-auto' ref={currentSearchAnchorElement}>
                <GraphButton
                    onClick={onReset}
                    displayText={<FontAwesomeIcon aria-label='reset graph view' icon={faCropAlt} />}
                />

                <GraphMenu label={'Hide Labels'}>
                    <MenuItem onClick={handleToggleAllLabels}>
                        {!showNodeLabels || !showEdgeLabels ? 'Show' : 'Hide'} All Labels
                    </MenuItem>
                    <MenuItem onClick={onToggleNodeLabels}>{showNodeLabels ? 'Hide' : 'Show'} Node Labels</MenuItem>
                    <MenuItem onClick={onToggleEdgeLabels}>{showEdgeLabels ? 'Hide' : 'Show'} Edge Labels</MenuItem>
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
                    onClick={() => setIsCurrentSearchOpen(true)}
                    displayText={'Search Current Results'}
                    disabled={isCurrentSearchOpen}
                />
            </div>
            <Popper
                open={isCurrentSearchOpen}
                anchorEl={currentSearchAnchorElement.current}
                placement='top'
                disablePortal
                className='w-[90%] z-[1]'>
                <div className='pointer-events-auto' data-testid='explore_graph-controls'>
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
