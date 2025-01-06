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

import { faCropAlt } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { Box, MenuItem } from '@mui/material';
import { GraphMenu, GraphButton } from 'bh-shared-ui';
import isEmpty from 'lodash/isEmpty';
import { FC } from 'react';
import { useAppSelector } from 'src/store';

interface GraphButtonsProps {
    onReset: () => void;
    onRunStandardLayout: () => void;
    onRunSequentialLayout: () => void;
    onExportJson: () => void;
    onSearchCurrentResults: () => void;
    onToggleAllLabels: () => void;
    onToggleNodeLabels: () => void;
    onToggleEdgeLabels: () => void;
    showNodeLabels: boolean;
    showEdgeLabels: boolean;
    isCurrentSearchOpen: boolean;
}

const GraphButtons: FC<GraphButtonsProps> = ({
    onReset,
    onRunStandardLayout,
    onRunSequentialLayout,
    onExportJson,
    onSearchCurrentResults,
    onToggleAllLabels,
    onToggleNodeLabels,
    onToggleEdgeLabels,
    showNodeLabels,
    showEdgeLabels,
    isCurrentSearchOpen,
}) => {
    const exportableGraphState = useAppSelector((state) => state.explore.export);

    return (
        <Box display={'flex'} gap={1}>
            <GraphButton onClick={onReset} displayText={<FontAwesomeIcon icon={faCropAlt} />} />

            <GraphMenu label={'Hide Labels'}>
                <MenuItem onClick={onToggleAllLabels}>
                    {!showNodeLabels || !showEdgeLabels ? 'Show' : 'Hide'} All Labels
                </MenuItem>
                <MenuItem onClick={onToggleNodeLabels}>{showNodeLabels ? 'Hide' : 'Show'} Node Labels</MenuItem>
                <MenuItem onClick={onToggleEdgeLabels}>{showEdgeLabels ? 'Hide' : 'Show'} Edge Labels</MenuItem>
            </GraphMenu>

            <GraphMenu label='Layout'>
                <MenuItem onClick={onRunSequentialLayout}>Sequential</MenuItem>
                <MenuItem onClick={onRunStandardLayout}>Standard</MenuItem>
            </GraphMenu>

            <GraphMenu label='Export'>
                <MenuItem onClick={onExportJson} disabled={isEmpty(exportableGraphState)}>
                    JSON
                </MenuItem>
            </GraphMenu>

            <GraphButton
                onClick={onSearchCurrentResults}
                displayText={'Search Current Results'}
                disabled={isCurrentSearchOpen}
            />
        </Box>
    );
};

export default GraphButtons;
