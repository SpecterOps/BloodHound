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

import { Box, Paper, SxProps } from '@mui/material';
import { SelectedEdge } from 'bh-shared-ui';
import React, { useState } from 'react';
import EdgeInfoContent from 'src/views/Explore/EdgeInfo/EdgeInfoContent';
import Header from 'src/views/Explore/EdgeInfo/EdgeInfoHeader';
import { usePaneStyles } from 'bh-shared-ui';

const EdgeInfoPane: React.FC<{ selectedEdge: SelectedEdge; sx?: SxProps }> = ({ selectedEdge, sx }) => {
    const styles = usePaneStyles();
    const [expanded, setExpanded] = useState(true);

    return (
        <Box sx={sx} className={styles.container} data-testid='explore_edge-information-pane'>
            <Paper elevation={0} classes={{ root: styles.headerPaperRoot }}>
                <Header
                    name={selectedEdge?.name || 'None'}
                    expanded={expanded}
                    onToggleExpanded={(expanded) => {
                        setExpanded(expanded);
                    }}
                />
            </Paper>
            <Paper
                elevation={0}
                classes={{ root: styles.contentPaperRoot }}
                sx={{
                    display: expanded ? 'initial' : 'none',
                }}>
                {selectedEdge === null ? 'No information to display.' : <EdgeInfoContent selectedEdge={selectedEdge} />}
            </Paper>
        </Box>
    );
};

export default EdgeInfoPane;
