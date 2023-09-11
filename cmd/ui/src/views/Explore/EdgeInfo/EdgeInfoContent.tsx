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

import { Box, Divider, Typography, useTheme } from '@mui/material';
import { EdgeInfoComponents, EdgeSections, SelectedEdge } from 'bh-shared-ui';
import { FC, Fragment } from 'react';
import EdgeInfoCollapsibleSection from 'src/views/Explore/EdgeInfo/EdgeInfoCollapsibleSection';
import EdgeObjectInformation from 'src/views/Explore/EdgeInfo/EdgeObjectInformation';

const EdgeInfoContent: FC<{ selectedEdge: NonNullable<SelectedEdge> }> = ({ selectedEdge }) => {
    const theme = useTheme();

    const sections = EdgeInfoComponents[selectedEdge.name as keyof typeof EdgeInfoComponents];
    const { sourceNode, targetNode } = selectedEdge;

    return (
        <Box>
            <EdgeObjectInformation selectedEdge={selectedEdge} />
            {sections ? (
                <>
                    {Object.entries(sections).map((section, index) => {
                        const Section = section[1];
                        return (
                            <Fragment key={index}>
                                <Box padding={1}>
                                    <Divider />
                                </Box>
                                <EdgeInfoCollapsibleSection section={section[0] as keyof typeof EdgeSections}>
                                    <Section
                                        sourceName={sourceNode.name}
                                        sourceType={sourceNode.type}
                                        targetName={targetNode.name}
                                        targetType={targetNode.type}
                                        targetId={targetNode.objectId}
                                        haslaps={!!targetNode.haslaps}
                                    />
                                </EdgeInfoCollapsibleSection>
                            </Fragment>
                        );
                    })}
                </>
            ) : (
                <>
                    <Box padding={1}>
                        <Divider />
                    </Box>
                    <Box paddingLeft={theme.spacing(1)}>
                        <Typography variant='body1' fontSize={'0.75rem'}>
                            The edge{' '}
                            <Typography component={'span'} variant='body1' fontWeight={'bold'} fontSize={'0.75rem'}>
                                {selectedEdge.name}
                            </Typography>{' '}
                            does not have any additional contextual information at this time.
                        </Typography>
                    </Box>
                </>
            )}
        </Box>
    );
};

export default EdgeInfoContent;
