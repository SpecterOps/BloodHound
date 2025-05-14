import { Box, Paper, SxProps, Typography } from '@mui/material';
import React, { useState } from 'react';
import { NoEntitySelectedHeader, NoEntitySelectedMessage } from '../../../utils';
import { usePaneStyles } from '../InfoStyles';
import { ObjectInfoPanelContextProvider } from '../providers/ObjectInfoPanelProvider';

import { SelectedNode } from '../../../utils';
import EntityInfoContent from './EntityInfoContent';
import Header from './EntityInfoHeader';

interface EntityInfoPanelProps {
    selectedNode: SelectedNode | null;
    sx?: SxProps;
}

const EntityInfoPanel: React.FC<EntityInfoPanelProps> = ({ selectedNode, sx }) => {
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
                        id={selectedNode.id}
                        nodeType={selectedNode.type}
                        databaseId={selectedNode.graphId}
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
