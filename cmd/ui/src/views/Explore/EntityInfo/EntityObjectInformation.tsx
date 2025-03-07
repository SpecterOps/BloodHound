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

import { Alert } from '@mui/material';
import { EntityField, FieldsContainer, ObjectInfoFields, formatObjectInfoFields } from 'bh-shared-ui';
import React from 'react';
import { BasicObjectInfoFields } from '../BasicObjectInfoFields';
import EntityInfoCollapsibleSection from './EntityInfoCollapsibleSection';
import { EntityInfoContentProps } from './EntityInfoContent';

const EntityObjectInformation: React.FC<EntityInfoContentProps> = ({ selectedNode }) => {
    if (!selectedNode.properties)
        return (
            <EntityInfoCollapsibleSection label='Object Information'>
                <FieldsContainer>
                    <Alert severity='error'>Unable to load object information for this node.</Alert>
                </FieldsContainer>
            </EntityInfoCollapsibleSection>
        );

    const formattedObjectFields: EntityField[] = formatObjectInfoFields(selectedNode.properties);

    return (
        <EntityInfoCollapsibleSection label='Object Information'>
            <FieldsContainer>
                <BasicObjectInfoFields {...selectedNode.properties} />
                <ObjectInfoFields fields={formattedObjectFields} />
            </FieldsContainer>
        </EntityInfoCollapsibleSection>
    );
};

export default EntityObjectInformation;
