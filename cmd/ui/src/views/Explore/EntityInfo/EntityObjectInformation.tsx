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
    BasicObjectInfoFields,
    EntityField,
    FieldsContainer,
    ObjectInfoFields,
    SearchValue,
    formatObjectInfoFields,
    searchbarActions,
    useExploreParams,
    useFeatureFlag,
    useFetchEntityProperties,
    useObjectInfoPanelContext,
} from 'bh-shared-ui';
import React, { useEffect } from 'react';
import usePreviousValue from 'src/hooks/usePreviousValue';
import { useAppDispatch } from 'src/store';
import EntityInfoCollapsibleSection from './EntityInfoCollapsibleSection';
import { EntityInfoContentProps } from './EntityInfoContent';

const EntityObjectInformation: React.FC<EntityInfoContentProps> = ({ id, nodeType, databaseId }) => {
    const { setExploreParams } = useExploreParams();
    const { data: flag } = useFeatureFlag('back_button_support');
    const dispatch = useAppDispatch();
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
        if (flag?.enabled) {
            setExploreParams({ primarySearch: sourceNode.objectid, searchType: 'node' });
        } else {
            dispatch(searchbarActions.sourceNodeSelected(sourceNode));
        }
    };

    return (
        <EntityInfoCollapsibleSection onChange={handleOnChange} isExpanded={isObjectInfoPanelOpen} label={sectionLabel}>
            <FieldsContainer>
                <BasicObjectInfoFields handleSourceNodeSelected={handleSourceNodeSelected} {...entityProperties} />
                <ObjectInfoFields fields={formattedObjectFields} />
            </FieldsContainer>
        </EntityInfoCollapsibleSection>
    );
};

export default EntityObjectInformation;
