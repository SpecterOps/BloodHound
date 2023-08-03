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

import { Box, Divider } from '@mui/material';
import { EdgeInfoComponents } from 'bh-shared-ui';
import React from 'react';
import { EdgeSections, SelectedEdge } from 'src/ducks/edgeinfo/edgeSlice';
import EdgeInfoCollapsibleSection from 'src/views/Explore/EdgeInfo/EdgeInfoCollapsibleSection';
import EdgeObjectInformation from 'src/views/Explore/EdgeInfo/EdgeObjectInformation';

const EdgeInfoContent: React.FC<{ selectedEdge: NonNullable<SelectedEdge> }> = ({ selectedEdge }) => {
    const sections = EdgeInfoComponents[selectedEdge.name as keyof typeof EdgeInfoComponents];
    const { sourceNode, targetNode } = selectedEdge;

    return (
        <Box>
            <EdgeObjectInformation selectedEdge={selectedEdge} />
            {
                <>
                    {Object.entries(sections).map((section, index) => {
                        const Section = section[1];
                        return (
                            <React.Fragment key={index}>
                                <Box padding={1}>
                                    <Divider />
                                </Box>
                                <EdgeInfoCollapsibleSection section={section[0] as keyof typeof EdgeSections}>
                                    <Section
                                        sourceName={sourceNode.data.name}
                                        sourceType={sourceNode.data.nodetype}
                                        targetName={targetNode.data.name}
                                        targetType={targetNode.data.nodetype}
                                        targetId={targetNode.data.objectid}
                                        haslaps={!!targetNode.data.haslaps}
                                    />
                                </EdgeInfoCollapsibleSection>
                            </React.Fragment>
                        );
                    })}
                </>
            }
        </Box>
    );
};

export default EdgeInfoContent;
