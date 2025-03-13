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
import {
    EdgeSections,
    collapseAllSections,
    edgeSectionToggle,
    useExploreParams,
    useFeatureFlag,
    usePaneStyles,
} from 'bh-shared-ui';
import React, { useEffect, useState } from 'react';
import usePreviousValue from 'src/hooks/usePreviousValue';
import { useAppDispatch } from 'src/store';
import EdgeInfoContent from 'src/views/Explore/EdgeInfo/EdgeInfoContent';
import Header from 'src/views/Explore/EdgeInfo/EdgeInfoHeader';

const EdgeInfoPane: React.FC<{ sx?: SxProps; selectedEdge?: any }> = ({ sx, selectedEdge }) => {
    const styles = usePaneStyles();
    const [expanded, setExpanded] = useState(true);
    const { expandedPanelSections } = useExploreParams();
    const previousSelectedEdge = usePreviousValue(selectedEdge);
    const { data: backButtonFlag } = useFeatureFlag('back_button_support');
    const dispatch = useAppDispatch();

    useEffect(() => {
        if (previousSelectedEdge?.id !== selectedEdge?.id && backButtonFlag?.enabled && expandedPanelSections) {
            dispatch(
                edgeSectionToggle({
                    section: expandedPanelSections?.at(0) as keyof typeof EdgeSections,
                    expanded: true,
                })
            );
        } else if (!backButtonFlag?.enabled && previousSelectedEdge?.id !== selectedEdge?.id) {
            dispatch(collapseAllSections());
        }
    }, [expandedPanelSections, dispatch, backButtonFlag, previousSelectedEdge, selectedEdge]);

    return (
        <Box sx={sx} className={styles.container} data-testid='explore_edge-information-pane'>
            <Paper elevation={0} classes={{ root: styles.headerPaperRoot }}>
                <Header
                    name={selectedEdge.name || 'None'}
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
