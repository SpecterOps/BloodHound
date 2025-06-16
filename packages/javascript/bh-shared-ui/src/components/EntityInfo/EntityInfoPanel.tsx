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
import { AssetGroupTagMemberInfo } from 'js-client-library';
import React, { useState } from 'react';
import { useLocation } from 'react-router-dom';
import { SelectedNode } from '../../types';
import { NoEntitySelectedHeader, NoEntitySelectedMessage } from '../../utils';
import { ObjectInfoPanelContextProvider, usePaneStyles } from '../../views';
import EntityInfoContent from './EntityInfoContent';
import Header from './EntityInfoHeader';

interface EntityInfoPanelProps {
    selectedNode: SelectedNode | null;
    selectedZoneManagementNode: AssetGroupTagMemberInfo | null;
    sx?: SxProps;
}

const EntityInfoPanel: React.FC<EntityInfoPanelProps> = ({ selectedNode, sx, selectedZoneManagementNode }) => {
    const styles = usePaneStyles();
    const [expanded, setExpanded] = useState(true);

    const location = useLocation();

    const zoneManagement = location.pathname.includes('zone-management');

    if (zoneManagement) {
        console.log(location.pathname);
    } else {
        console.log('FALSE');
    }

    return (
        <>
            {!zoneManagement ? (
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
                                id={selectedNode.id}
                                nodeType={selectedNode.type}
                                databaseId={selectedNode.graphId}
                            />
                        ) : (
                            <Typography variant='body2'>{NoEntitySelectedMessage}</Typography>
                        )}
                    </Paper>
                </Box>
            ) : (
                <div className='max-w-[400px]'>
                    <Box className={styles.container} data-testid='explore_entity-information-panel'>
                        <Paper elevation={0} classes={{ root: styles.headerPaperRoot }}>
                            <Header
                                name={selectedZoneManagementNode?.properties?.name || NoEntitySelectedHeader}
                                nodeType={selectedZoneManagementNode?.primary_kind}
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
                                    id={selectedNode.id}
                                    nodeType={selectedZoneManagementNode?.primary_kind}
                                    properties={selectedZoneManagementNode?.properties}
                                />
                            ) : (
                                <Typography variant='body2'>{NoEntitySelectedMessage}</Typography>
                            )}
                        </Paper>
                    </Box>
                </div>
            )}
        </>
    );
};

const WrappedEntityInfoPanel: React.FC<EntityInfoPanelProps> = (props) => (
    <ObjectInfoPanelContextProvider>
        <EntityInfoPanel {...props} />
    </ObjectInfoPanelContextProvider>
);

export default WrappedEntityInfoPanel;
