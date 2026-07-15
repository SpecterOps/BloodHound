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
import { Box, Divider } from '@mui/material';
import { Typography } from 'doodle-ui';
import { RelationshipDetails } from 'js-client-library';
import { ElementType, FC, Fragment } from 'react';
import EdgeInfoComponents from '../../../components/HelpTexts';
import ACLInheritance from '../../../components/HelpTexts/shared/ACLInheritance';
import { ActiveDirectoryKindProperties, CommonKindProperties } from '../../../graphSchema';
import { useExploreParams, useExploreSelectedItem, useGetNodeById } from '../../../hooks';
import { usePrimaryKind } from '../../../hooks/usePrimaryKind';
import { EdgeSections } from '../ExploreSearch/EdgeFilter/edgeCategories';
import { FieldsContainer } from '../fragments';
import EdgeInfoCollapsibleSection from './EdgeInfoCollapsibleSection';
import EdgeObjectInformation from './EdgeObjectInformation';

const EdgeInfoContent: FC<{ selectedEdge: NonNullable<RelationshipDetails> }> = ({ selectedEdge }) => {
    const { setExploreParams, expandedPanelSections } = useExploreParams();
    const { isHidden } = useExploreSelectedItem();
    const sections = EdgeInfoComponents[selectedEdge.kind.name as keyof typeof EdgeInfoComponents];
    const { source_node_id, target_node_id } = selectedEdge;

    const { data: sourceNode, ...sourceNodeQuery } = useGetNodeById(source_node_id);
    const { data: targetNode, ...targetNodeQuery } = useGetNodeById(target_node_id);

    const sourcePrimaryKind = usePrimaryKind(sourceNode?.kinds ?? []);
    const targetPrimaryKind = usePrimaryKind(targetNode?.kinds ?? []);

    if (sourceNodeQuery.isLoading || targetNodeQuery.isLoading) return null;

    const removeExpandedPanelSectionParams = () => {
        setExploreParams({
            expandedPanelSections: [],
        });
    };

    const shouldRenderACLInheritance = !!(
        selectedEdge.properties[ActiveDirectoryKindProperties.IsACL] &&
        selectedEdge.properties[CommonKindProperties.IsInherited] &&
        typeof selectedEdge.properties[ActiveDirectoryKindProperties.InheritanceHash] === 'string' &&
        selectedEdge.properties[ActiveDirectoryKindProperties.InheritanceHash].length > 0
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
                    relationshipQueryItemId: `rel_${selectedEdge.relationship_id}`,
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
                        edgeName={selectedEdge.kind.name}
                        sourceDBId={source_node_id}
                        sourceName={sourceNode?.properties.name}
                        sourceType={sourcePrimaryKind}
                        targetDBId={target_node_id}
                        targetName={targetNode?.properties.name}
                        targetType={targetPrimaryKind}
                        targetId={targetNode?.properties.objectid}
                        haslaps={!!targetNode?.properties.haslaps}
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
                searchType: 'aclinheritance',
                relationshipQueryItemId: `rel_${selectedEdge.relationship_id}`,
            });
        };

        const handleOnChange = (isOpen: boolean) => {
            if (isOpen) setExpandedPanelSectionsParam();
            else removeExpandedPanelSectionParams();
        };

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
                        edgeName={selectedEdge.kind.name}
                        sourceDBId={source_node_id}
                        targetDBId={target_node_id}
                        inheritanceHash={
                            selectedEdge.properties[ActiveDirectoryKindProperties.InheritanceHash] as string
                        }
                    />
                </EdgeInfoCollapsibleSection>
            </Fragment>
        );
    };

    return (
        <Box>
            {isHidden ? (
                <FieldsContainer>
                    <div>
                        <p className='text-sm'>
                            This edge's information is not disclosed. Please contact your admin in order to get access.
                        </p>
                    </div>
                </FieldsContainer>
            ) : (
                <EdgeObjectInformation selectedEdge={selectedEdge} sourceNode={sourceNode} targetNode={targetNode} />
            )}
            {sections || shouldRenderACLInheritance ? (
                <>
                    {Object.entries(sections).map(renderDropdownFromSection)}
                    {shouldRenderACLInheritance && renderACLInheritanceDropdown()}
                </>
            ) : (
                <>
                    {!isHidden && (
                        <>
                            <Box padding={1}>
                                <Divider />
                            </Box>

                            <Box paddingLeft={'0.5rem'}>
                                <Typography variant='body1' className='text-xs'>
                                    The edge <strong className='text-xs'>{selectedEdge.kind.name}</strong>
                                    does not have any additional contextual information at this time.
                                </Typography>
                            </Box>
                        </>
                    )}
                </>
            )}
        </Box>
    );
};

export default EdgeInfoContent;
