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
import { NodeDetails, RelationshipDetails } from 'js-client-library';
import { FC, useEffect } from 'react';
import { usePreviousValue } from '../../../hooks';
import { EntityField, formatObjectInfoFields } from '../../../utils';
import { FieldsContainer, ObjectInfoFields } from '../fragments';
import { useObjectInfoPanelContext } from '../providers';
import EdgeInfoCollapsibleSection from './EdgeInfoCollapsibleSection';

type EdgeObjectInformationProps = {
    selectedEdge: NonNullable<RelationshipDetails>;
    sourceNode: NodeDetails | undefined;
    targetNode: NodeDetails | undefined;
};

const EdgeObjectInformation: FC<EdgeObjectInformationProps> = ({ selectedEdge, sourceNode, targetNode }) => {
    const { isObjectInfoPanelOpen, setIsObjectInfoPanelOpen } = useObjectInfoPanelContext();

    const previousId = usePreviousValue(selectedEdge.relationship_id);

    useEffect(() => {
        if (previousId !== selectedEdge.relationship_id) {
            setIsObjectInfoPanelOpen(true);
        }
    }, [previousId, selectedEdge.relationship_id, setIsObjectInfoPanelOpen]);

    const sourceNodeField: EntityField = {
        label: 'Source Node:',
        value: sourceNode?.properties.name || sourceNode?.properties.objectid || '',
    };

    const targetNodeField: EntityField = {
        label: 'Target Node:',
        value: targetNode?.properties.name || targetNode?.properties.objectid || '',
    };

    const formattedObjectFields: EntityField[] = [
        sourceNodeField,
        targetNodeField,
        ...formatObjectInfoFields({
            ...selectedEdge.properties,
        }),
    ];

    const sectionLabel = 'Relationship Information';

    const handleOnChange = () => {
        setIsObjectInfoPanelOpen(!isObjectInfoPanelOpen);
    };

    return (
        <EdgeInfoCollapsibleSection isExpanded={isObjectInfoPanelOpen} onChange={handleOnChange} label={sectionLabel}>
            <FieldsContainer>
                <ObjectInfoFields fields={formattedObjectFields} />
            </FieldsContainer>
        </EdgeInfoCollapsibleSection>
    );
};

export default EdgeObjectInformation;
