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
import { Box, Divider, Typography, useTheme } from '@mui/material';
import { ElementType, FC, Fragment } from 'react';
import EdgeInfoComponents from '../../../components/HelpTexts';
import ACLInheritance from '../../../components/HelpTexts/shared/AclInheritance';
import { ActiveDirectoryKindProperties, CommonKindProperties } from '../../../graphSchema';
import { useExploreParams, useFetchEntityProperties } from '../../../hooks';
import { EdgeSections, SelectedEdge } from '../../../store';
import EdgeInfoCollapsibleSection from './EdgeInfoCollapsibleSection';
import EdgeObjectInformation from './EdgeObjectInformation';

const EdgeInfoContent: FC<{ selectedEdge: NonNullable<SelectedEdge> }> = ({ selectedEdge }) => {
    const theme = useTheme();
    const { setExploreParams, expandedPanelSections } = useExploreParams();
    const sections = EdgeInfoComponents[selectedEdge.name as keyof typeof EdgeInfoComponents];
    const { sourceNode, targetNode } = selectedEdge;
    const { objectId, type } = targetNode;
    const { entityProperties: targetNodeProperties } = useFetchEntityProperties({
        objectId,
        nodeType: type,
    });

    const removeExpandedPanelSectionParams = () => {
        setExploreParams({
            expandedPanelSections: [],
        });
    };

    const shouldRenderACLInheritance = !!(
        selectedEdge.data[ActiveDirectoryKindProperties.IsACL] &&
        selectedEdge.data[CommonKindProperties.IsInherited] &&
        typeof selectedEdge.data[ActiveDirectoryKindProperties.InheritanceHash] === 'string' &&
        selectedEdge.data[ActiveDirectoryKindProperties.InheritanceHash].length > 0
    );

    const renderDropdownFromSection = (section: [string, any], index: number) => {
        const Section = section[1] as ElementType;
        const sectionKeyLabel = section[0] as keyof typeof EdgeSections;

        const isExpandedPanelSection = (expandedPanelSections as string[]).includes(sectionKeyLabel);

        const setExpandedPanelSectionsParam = () => {
            setExploreParams({
                expandedPanelSections: [sectionKeyLabel],
                ...(sectionKeyLabel === 'composition' && {
                    searchType: 'composition',
                    relationshipQueryItemId: selectedEdge.id,
                }),
            });
        };

        const handleOnChange = (isOpen: boolean) => {
            if (isOpen) setExpandedPanelSectionsParam();
            else removeExpandedPanelSectionParams();
        };

        return (
            <Fragment key={index}>
                <Box padding={1}>
                    <Divider />
                </Box>
                <EdgeInfoCollapsibleSection
                    label={EdgeSections[sectionKeyLabel]}
                    isExpanded={isExpandedPanelSection}
                    onChange={handleOnChange}>
                    <Section
                        edgeName={selectedEdge.name}
                        sourceDBId={sourceNode.id}
                        sourceName={sourceNode.name}
                        sourceType={sourceNode.type}
                        targetDBId={targetNode.id}
                        targetName={targetNode.name}
                        targetType={targetNode.type}
                        targetId={targetNode.objectId}
                        haslaps={!!targetNodeProperties?.haslaps}
                    />
                </EdgeInfoCollapsibleSection>
            </Fragment>
        );
    };

    const renderACLInheritanceDropdown = () => {
        const isExpandedPanelSection = (expandedPanelSections as string[]).includes('aclinheritance');

        const setExpandedPanelSectionsParam = () => {
            setExploreParams({
                expandedPanelSections: ['aclinheritance'],
                // searchType: 'aclinheritance',
                relationshipQueryItemId: selectedEdge.id,
            });
        };

        const handleOnChange = (isOpen: boolean) => {
            if (isOpen) setExpandedPanelSectionsParam();
            else removeExpandedPanelSectionParams();
        };

        const castIdToInt = (id: string | number) => (typeof id === 'string' ? parseInt(id) : id);

        return (
            <Fragment key={Object.keys(sections).length}>
                <Box padding={1}>
                    <Divider />
                </Box>
                <EdgeInfoCollapsibleSection
                    label={'ACE Inherited From'}
                    isExpanded={isExpandedPanelSection}
                    onChange={handleOnChange}>
                    <ACLInheritance
                        edgeName={selectedEdge.name}
                        sourceDBId={castIdToInt(sourceNode.id)}
                        targetDBId={castIdToInt(targetNode.id)}
                        inheritanceHash={selectedEdge.data[ActiveDirectoryKindProperties.InheritanceHash]}
                    />
                </EdgeInfoCollapsibleSection>
            </Fragment>
        );
    };

    return (
        <Box>
            <EdgeObjectInformation selectedEdge={selectedEdge} />
            {sections || shouldRenderACLInheritance ? (
                <>
                    {Object.entries(sections).map(renderDropdownFromSection)}
                    {shouldRenderACLInheritance && renderACLInheritanceDropdown()}
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
