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

import { Alert, Skeleton } from '@mui/material';
import {
    EntityField,
    FieldsContainer,
    ObjectInfoFields,
    formatObjectInfoFields,
    useFetchEntityProperties,
} from 'bh-shared-ui';
import React from 'react';
import { BasicObjectInfoFields } from '../BasicObjectInfoFields';
import EntityInfoCollapsibleSection from './EntityInfoCollapsibleSection';
import { EntityInfoContentProps } from './EntityInfoContent';
import { useEntityInfoPanelContext } from './EntityInfoPanelContext';

const EntityObjectInformation: React.FC<EntityInfoContentProps> = ({ id, nodeType, databaseId }) => {
    const { entityProperties, informationAvailable, isLoading, isError } = useFetchEntityProperties({
        objectId: id,
        nodeType,
        databaseId,
    });
    const { expandedSections, setExpandedSections } = useEntityInfoPanelContext();
    const sectionLabel = 'Object Information';

    if (isLoading) return <Skeleton data-testid='entity-object-information-skeleton' variant='text' />;

    if (isError || !informationAvailable)
        return (
            <EntityInfoCollapsibleSection
                isExpanded={!!expandedSections[sectionLabel]}
                onChange={(label, isOpen) => {
                    setExpandedSections({ ...expandedSections, [label]: !isOpen });
                }}
                label={sectionLabel}>
                <FieldsContainer>
                    <Alert severity='error'>Unable to load object information for this node.</Alert>
                </FieldsContainer>
            </EntityInfoCollapsibleSection>
        );

    const formattedObjectFields: EntityField[] = formatObjectInfoFields(entityProperties);

    return (
        <EntityInfoCollapsibleSection
            isExpanded={!!expandedSections[sectionLabel]}
            onChange={(label, isOpen) => {
                setExpandedSections({ ...expandedSections, [label]: isOpen });
            }}
            label={sectionLabel}>
            <FieldsContainer>
                <BasicObjectInfoFields {...entityProperties} />
                <ObjectInfoFields fields={formattedObjectFields} />
            </FieldsContainer>
        </EntityInfoCollapsibleSection>
    );
};

export default EntityObjectInformation;
