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
import { Box, Link, Paper, SxProps, Typography } from '@mui/material';
import React, { useState } from 'react';
import { usePaneStyles } from '../InfoStyles';
import { ObjectInfoPanelContextProvider } from '../providers';
import Header from './EdgeInfoHeader';

interface DeepSniffInfoPaneProps {
    sx?: SxProps;
}

const DeepSniffInfoPane: React.FC<DeepSniffInfoPaneProps> = ({ sx }) => {
    const styles = usePaneStyles();
    const [expanded, setExpanded] = useState(true);

    return (
        <Box sx={sx} className={styles.container} data-testid='explore_deepsniff-information-pane'>
            <Paper elevation={0} classes={{ root: styles.headerPaperRoot }}>
                <Header
                    name={'Deep Sniff'}
                    expanded={expanded}
                    onToggleExpanded={(isExpanded: boolean) => setExpanded(isExpanded)}
                />
            </Paper>
            <Paper
                elevation={0}
                classes={{ root: styles.contentPaperRoot }}
                style={{ display: expanded ? 'initial' : 'none' }}>
                <Typography variant='body2' sx={{ mb: 2 }}>
                    The principal has attack paths that can grant it DCSync permissions on the target domain.
                </Typography>

                <Typography variant='h6' sx={{ mb: 1, fontWeight: 'bold' }}>
                    Abuse
                </Typography>
                <Typography variant='body2' sx={{ mb: 2 }}>
                    A DCSync attack requires both the GetChanges and GetChangesAll permissions. Execute the attack
                    paths that result in obtaining these permissions on the target domain.
                </Typography>

                <Typography variant='body2'>
                    For details on performing a DCSync attack, refer to the{' '}
                    <Link
                        href='https://bloodhound.readthedocs.io/en/latest/data-analysis/edges.html#dcsync'
                        target='_blank'
                        rel='noopener noreferrer'>
                        DCSync edge documentation
                    </Link>
                    .
                </Typography>
            </Paper>
        </Box>
    );
};

const WrappedDeepSniffInfoPane: React.FC<DeepSniffInfoPaneProps> = (props) => (
    <ObjectInfoPanelContextProvider>
        <DeepSniffInfoPane {...props} />
    </ObjectInfoPanelContextProvider>
);

export default WrappedDeepSniffInfoPane;
