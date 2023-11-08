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

import { Paper } from '@mui/material';
import { SelectedEdge } from 'bh-shared-ui';
import React, { useState } from 'react';
import EdgeInfoContent from 'src/views/Explore/EdgeInfo/EdgeInfoContent';
import Header from 'src/views/Explore/EdgeInfo/EdgeInfoHeader';
import usePaneStyles from 'src/views/Explore/InfoStyles/Pane';
import { useQuery } from 'react-query';
import {edgeInformationEndpoints} from 'src/views/Explore/EntityInfo/content.ts';

const EdgeInfoPane: React.FC<{ selectedEdge: SelectedEdge }> = ({ selectedEdge }) => {
    const styles = usePaneStyles();
    const [expanded, setExpanded] = useState(true);

    const { data, isLoading, isError } = useQuery(
        ['edge', selectedEdge?.name, selectedEdge?.sourceNode, selectedEdge?.targetNode],
        ({ signal }) => edgeInformationEndpoints[selectedEdge.name]?.(id, { signal }).then((res) => res.data.data),
        { refetchOnWindowFocus: false, retry: false }
    );

    return (
        <div className={styles.container} data-testid='explore_edge-information-pane'>
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
                {isLoading ? (
                    <div>Loading...</div>
                ) : isError || data === undefined ? (
                    <div>Unable to load node information.</div>
                ) : (
                    <EdgeInfoContent selectedEdge={selectedEdge} {...data} />
                )}
                {selectedEdge === null ? 'No information to display.' : <EdgeInfoContent selectedEdge={selectedEdge} />}
            </Paper>
        </div>
    );
};

export default EdgeInfoPane;
