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
import { Box, Paper, SxProps, Typography } from '@mui/material';
import React, { useState } from 'react';
import { SelectedNode } from '../../types';
import { EntityInfoDataTableProps, NoEntitySelectedHeader, NoEntitySelectedMessage } from '../../utils';
import { ObjectInfoPanelContextProvider, usePaneStyles } from '../../views/Explore';
import EntityInfoContent from './EntityInfoContent';
import Header from './EntityInfoHeader';

interface EntityInfoPanelProps {
    DataTable: React.FC<EntityInfoDataTableProps>;
    selectedNode?: SelectedNode | null;
    sx?: SxProps;
    additionalTables?: {
        sectionProps: EntityInfoDataTableProps;
        TableComponent: React.FC<EntityInfoDataTableProps>;
    }[];
}

const EntityInfoPanel: React.FC<EntityInfoPanelProps> = ({ selectedNode, sx, additionalTables, DataTable }) => {
    const styles = usePaneStyles();
    const [expanded, setExpanded] = useState(true);

    return (
        <Box sx={sx} className={styles.container} data-testid='explore_entity-information-panel'>
            <Paper elevation={0} classes={{ root: styles.headerPaperRoot }}>
                <Header
                    name={selectedNode?.name || NoEntitySelectedHeader}
                    nodeType={selectedNode?.type}
                    expanded={expanded}
                    onToggleExpanded={(expanded) => {
                        setExpanded(expanded);
                    }}
                />
            </Paper>
            <Paper
                elevation={0}
                classes={{ root: styles.contentPaperRoot }}
                style={{
                    display: expanded ? 'initial' : 'none',
                }}>
                {selectedNode ? (
                    <EntityInfoContent
                        DataTable={DataTable}
                        id={selectedNode.id}
                        nodeType={selectedNode.type}
                        databaseId={selectedNode.graphId}
                        additionalTables={additionalTables}
                    />
                ) : (
                    <Typography variant='body2'>{NoEntitySelectedMessage}</Typography>
                )}
            </Paper>
        </Box>
    );
};

const WrappedEntityInfoPanel: React.FC<EntityInfoPanelProps> = (props) => (
    <ObjectInfoPanelContextProvider>
        <EntityInfoPanel {...props} />
    </ObjectInfoPanelContextProvider>
);

export default WrappedEntityInfoPanel;
