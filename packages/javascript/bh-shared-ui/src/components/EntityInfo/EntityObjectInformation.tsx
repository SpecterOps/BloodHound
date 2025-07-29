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
import { Alert, Skeleton } from '@mui/material';
import React, { useEffect } from 'react';
import { useExploreParams, useFetchEntityProperties, usePreviousValue } from '../../hooks';
import { EntityField, EntityInfoContentProps, formatObjectInfoFields } from '../../utils';
import { BasicObjectInfoFields } from '../../views/Explore/BasicObjectInfoFields';
import { SearchValue } from '../../views/Explore/ExploreSearch';
import { FieldsContainer, ObjectInfoFields } from '../../views/Explore/fragments';
import { useObjectInfoPanelContext } from '../../views/Explore/providers/ObjectInfoPanelProvider';
import EntityInfoCollapsibleSection from './EntityInfoCollapsibleSection';

const EntityObjectInformation: React.FC<EntityInfoContentProps> = ({ id, nodeType, databaseId }) => {
    const { setExploreParams } = useExploreParams();
    const { isObjectInfoPanelOpen, setIsObjectInfoPanelOpen } = useObjectInfoPanelContext();
    const { entityProperties, informationAvailable, isLoading, isError } = useFetchEntityProperties({
        objectId: id,
        nodeType,
        databaseId,
    });

    const previousId = usePreviousValue(id);

    useEffect(() => {
        if (previousId !== id) {
            setIsObjectInfoPanelOpen(true);
        }
    }, [previousId, id, setIsObjectInfoPanelOpen]);

    const sectionLabel = 'Object Information';

    const handleOnChange = () => {
        setIsObjectInfoPanelOpen(!isObjectInfoPanelOpen);
    };

    if (isLoading) return <Skeleton data-testid='entity-object-information-skeleton' variant='text' />;

    if (isError || !informationAvailable)
        return (
            <EntityInfoCollapsibleSection
                onChange={handleOnChange}
                isExpanded={isObjectInfoPanelOpen}
                label={sectionLabel}>
                <FieldsContainer>
                    <Alert severity='error'>Unable to load object information for this node.</Alert>
                </FieldsContainer>
            </EntityInfoCollapsibleSection>
        );

    const formattedObjectFields: EntityField[] = formatObjectInfoFields(entityProperties);

    const handleSourceNodeSelected = (sourceNode: SearchValue) => {
        setExploreParams({ primarySearch: sourceNode.objectid, searchType: 'node' });
    };

    return (
        <EntityInfoCollapsibleSection onChange={handleOnChange} isExpanded={isObjectInfoPanelOpen} label={sectionLabel}>
            <FieldsContainer>
                <BasicObjectInfoFields
                    nodeType={nodeType}
                    handleSourceNodeSelected={handleSourceNodeSelected}
                    {...entityProperties}
                />
                <ObjectInfoFields fields={formattedObjectFields} />
            </FieldsContainer>
        </EntityInfoCollapsibleSection>
    );
};

export default EntityObjectInformation;
