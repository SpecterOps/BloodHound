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

import { Box, Link, Paper, SxProps, Typography } from '@mui/material';
import React, { useState } from 'react';
import { usePaneStyles, useHeaderStyles } from 'bh-shared-ui';

interface DCSyncPanelProps {
    sx?: SxProps;
}

const DCSyncPanelHeader = ({ expanded, onToggleExpanded }: { expanded: boolean; onToggleExpanded: (expanded: boolean) => void }) => {
    const styles = useHeaderStyles();

    return (
        <Box className={styles.header}>
            <Typography variant="h6" className={styles.headerText}>
                Enable DCSync Paths
            </Typography>
        </Box>
    );
};

const DCSyncPanel: React.FC<DCSyncPanelProps> = ({ sx }) => {
    const styles = usePaneStyles();
    const [expanded, setExpanded] = useState(true);

    return (
        <Box sx={sx} className={styles.container} data-testid='dcsync-panel'>
            <Paper elevation={0} classes={{ root: styles.headerPaperRoot }}>
                <DCSyncPanelHeader
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
                <Box sx={{ p: 2 }}>
                    <Typography variant="body2" sx={{ mb: 2 }}>
                        The principal has attack paths that can grant it DCSync permissions on the target domain.
                    </Typography>
                    
                    <Typography variant="h6" sx={{ mb: 1, fontWeight: 'bold' }}>
                        Abuse
                    </Typography>
                    <Typography variant="body2" sx={{ mb: 2 }}>
                        A DCSync attack requires both the GetChanges and GetChangesAll permissions. Execute the attack
                        paths that result in obtaining these permissions on the target domain.
                    </Typography>
                    
                    <Typography variant="body2">
                        For details on performing a DCSync attack, refer to the{' '}
                        <Link 
                            href="https://bloodhound.readthedocs.io/en/latest/data-analysis/edges.html#dcsync" 
                            target="_blank" 
                            rel="noopener noreferrer">
                            DCSync edge documentation
                        </Link>
                        .
                    </Typography>
                </Box>
            </Paper>
        </Box>
    );
};

export default DCSyncPanel;
