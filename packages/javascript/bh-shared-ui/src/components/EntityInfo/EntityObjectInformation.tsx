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
import { NodeDetails, NodeDetailsWithInfo } from 'js-client-library';
import { useEffect } from 'react';
import { kindObjectsToKindNames, useExploreParams, usePreviousValue, usePrimaryKind, useTagsQuery } from '../../hooks';
import { getZoneNameFromKinds } from '../../hooks/useAssetGroupTags';
import { EntityField, formatObjectInfoFields } from '../../utils';
import { BasicObjectInfoFields } from '../../views/Explore/BasicObjectInfoFields';
import { SearchValue } from '../../views/Explore/ExploreSearch';
import { FieldsContainer, ObjectInfoFields } from '../../views/Explore/fragments';
import { useObjectInfoPanelContext } from '../../views/Explore/providers/ObjectInfoPanelProvider';
import EntityInfoCollapsibleSection from './EntityInfoCollapsibleSection';

const sectionLabel = 'Object Information';

interface EntityObjectInformationProps {
    selectedNode: NodeDetails | NodeDetailsWithInfo;
}

export default function EntityObjectInformation({ selectedNode }: EntityObjectInformationProps) {
    const { setExploreParams } = useExploreParams();
    const { isObjectInfoPanelOpen, setIsObjectInfoPanelOpen } = useObjectInfoPanelContext();
    const previousEntity = usePreviousValue(selectedNode.node_id);

    const kindNames = kindObjectsToKindNames(selectedNode.kinds);
    const primaryKind = usePrimaryKind(kindNames);

    const tagsQuery = useTagsQuery();
    const zoneName = getZoneNameFromKinds(tagsQuery?.data, kindNames);

    useEffect(() => {
        if (!previousEntity || !selectedNode.node_id || previousEntity !== selectedNode.node_id) {
            setIsObjectInfoPanelOpen(true);
        }
    }, [previousEntity, selectedNode, setIsObjectInfoPanelOpen]);

    const handleOnChange = () => {
        setIsObjectInfoPanelOpen(!isObjectInfoPanelOpen);
    };

    const formattedObjectFields: EntityField[] = formatObjectInfoFields(selectedNode.properties);

    const handleSourceNodeSelected = (sourceNode: SearchValue) => {
        setExploreParams({ primarySearch: sourceNode.objectid, searchType: 'node' });
    };

    return (
        <EntityInfoCollapsibleSection onChange={handleOnChange} isExpanded={isObjectInfoPanelOpen} label={sectionLabel}>
            <FieldsContainer>
                <BasicObjectInfoFields
                    nodeType={primaryKind}
                    handleSourceNodeSelected={handleSourceNodeSelected}
                    {...selectedNode.properties}
                    zone={zoneName}
                />
                <ObjectInfoFields fields={formattedObjectFields} />
            </FieldsContainer>
        </EntityInfoCollapsibleSection>
    );
}
