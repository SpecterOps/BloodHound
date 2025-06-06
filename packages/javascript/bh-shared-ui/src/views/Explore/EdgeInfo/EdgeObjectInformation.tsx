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
import { Skeleton } from '@mui/material';
import { FC, useEffect } from 'react';
import { useQuery } from 'react-query';
import { usePreviousValue } from '../../../hooks';
import { SelectedEdge } from '../../../store';
import { EntityField, apiClient, formatObjectInfoFields } from '../../../utils';
import { FieldsContainer, ObjectInfoFields } from '../fragments';
import { useObjectInfoPanelContext } from '../providers';
import EdgeInfoCollapsibleSection from './EdgeInfoCollapsibleSection';

const selectedEdgeCypherQuery = (sourceId: string | number, targetId: string | number, edgeKind: string): string =>
    `MATCH (s)-[r:${edgeKind}]->(t) WHERE ID(s) = ${sourceId} AND ID(t) = ${targetId} RETURN r LIMIT 1`;

const EdgeObjectInformation: FC<{ selectedEdge: NonNullable<SelectedEdge> }> = ({ selectedEdge }) => {
    const { isObjectInfoPanelOpen, setIsObjectInfoPanelOpen } = useObjectInfoPanelContext();

    const previousId = usePreviousValue(selectedEdge.id);

    useEffect(() => {
        if (previousId !== selectedEdge.id) {
            setIsObjectInfoPanelOpen(true);
        }
    }, [previousId, selectedEdge.id, setIsObjectInfoPanelOpen]);

    const {
        data: cypherResponse,
        isLoading,
        isError,
    } = useQuery([selectedEdge.id], ({ signal }) => {
        return apiClient
            .cypherSearch(
                selectedEdgeCypherQuery(selectedEdge.sourceNode.id, selectedEdge.targetNode.id, selectedEdge.name),
                { signal },
                true
            )
            .then((result: any) => {
                if (!result.data.data) return { nodes: {}, edges: [] };
                return result.data.data;
            });
    });

    if (isLoading) {
        return <Skeleton variant='rectangular' sx={{}} />;
    }

    const sourceNodeField: EntityField = {
        label: 'Source Node:',
        value: selectedEdge.sourceNode.name,
    };

    const targetNodeField: EntityField = {
        label: 'Target Node:',
        value: selectedEdge.targetNode.name,
    };

    let formattedObjectFields: EntityField[] = [sourceNodeField, targetNodeField];

    if (!isError) {
        formattedObjectFields = [
            ...formattedObjectFields,
            ...formatObjectInfoFields({
                ...(cypherResponse.edges[0]?.properties || {}),
            }),
        ];
    }

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
