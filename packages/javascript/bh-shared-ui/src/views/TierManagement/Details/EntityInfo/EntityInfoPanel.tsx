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

import { Box, Paper, Typography } from '@mui/material';
import { AssetGroupTagMemberInfo } from 'js-client-library';
import React, { useEffect, useState } from 'react';
import { usePreviousValue } from '../../../../hooks';
import { NoEntitySelectedHeader, NoEntitySelectedMessage } from '../../../../utils';
import { usePaneStyles } from '../../../Explore/InfoStyles';
import EntityInfoContent from './EntityInfoContent';
import Header from './EntityInfoHeader';
import { useEntityInfoPanelContext } from './EntityInfoPanelContext';
import { EntityInfoPanelContextProvider } from './EntityInfoPanelContextProvider';

interface EntityInfoPanelProps {
    selectedNode: AssetGroupTagMemberInfo | null;
}

const EntityInfoPanel: React.FC<EntityInfoPanelProps> = ({ selectedNode }) => {
    const styles = usePaneStyles();
    const [expanded, setExpanded] = useState(true);
    const { setExpandedSections } = useEntityInfoPanelContext();
    const previousSelectedNode = usePreviousValue(selectedNode);

    useEffect(() => {
        if (previousSelectedNode?.node_id !== selectedNode?.node_id) {
            setExpandedSections({ 'Object Information': true });
        }
    }, [setExpandedSections, previousSelectedNode, selectedNode]);

    return (
        <Box className={styles.container} data-testid='explore_entity-information-panel'>
            <Paper elevation={0} classes={{ root: styles.headerPaperRoot }}>
                <Header
                    name={selectedNode?.properties?.name || NoEntitySelectedHeader}
                    nodeType={selectedNode?.primary_kind}
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
                        id={selectedNode.node_id}
                        nodeType={selectedNode.primary_kind}
                        properties={selectedNode.properties}
                    />
                ) : (
                    <Typography variant='body2'>{NoEntitySelectedMessage}</Typography>
                )}
            </Paper>
        </Box>
    );
};

const WrappedEntityInfoPanel: React.FC<EntityInfoPanelProps> = (props) => (
    <EntityInfoPanelContextProvider>
        <EntityInfoPanel {...props} />
    </EntityInfoPanelContextProvider>
);

export default WrappedEntityInfoPanel;
