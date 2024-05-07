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
    apiClient,
    entityInformationEndpoints,
    formatObjectInfoFields,
    getNodeByDatabaseIdCypher,
    validateNodeType,
} from 'bh-shared-ui';
import { RequestOptions } from 'js-client-library';
import React from 'react';
import { useQuery } from 'react-query';
import { BasicObjectInfoFields } from '../BasicObjectInfoFields';
import EntityInfoCollapsibleSection from './EntityInfoCollapsibleSection';
import { EntityInfoContentProps } from './EntityInfoContent';

const EntityObjectInformation: React.FC<EntityInfoContentProps> = ({ id, nodeType, databaseId }) => {
    const requestDetails: {
        endpoint: (
            params: string,
            options?: RequestOptions,
            includeProperties?: boolean
        ) => Promise<Record<string, any>>;
        param: string;
    } = {
        endpoint: async function () {
            return {};
        },
        param: '',
    };

    const validatedKind = validateNodeType(nodeType);

    if (validatedKind) {
        requestDetails.endpoint = entityInformationEndpoints[validatedKind];
        requestDetails.param = id;
    } else if (databaseId) {
        requestDetails.endpoint = apiClient.cypherSearch;
        requestDetails.param = getNodeByDatabaseIdCypher(databaseId);
    }

    const informationAvailable = !!validatedKind || !!databaseId;

    const {
        data: objectInformation,
        isLoading,
        isError,
    } = useQuery(
        ['entity', nodeType, id],
        ({ signal }) =>
            requestDetails.endpoint(requestDetails.param, { signal }, true).then((res) => {
                if (validatedKind) return res.data.data.props;
                else if (databaseId) return Object.values(res.data.data.nodes as Record<string, any>)[0].properties;
                else return {};
            }),
        {
            refetchOnWindowFocus: false,
            retry: false,
            enabled: informationAvailable,
        }
    );

    if (isLoading) return <Skeleton data-testid='entity-object-information-skeleton' variant='text' />;

    if (isError || !informationAvailable)
        return (
            <EntityInfoCollapsibleSection label='Object Information'>
                <FieldsContainer>
                    <Alert severity='error'>Unable to load object information for this node.</Alert>
                </FieldsContainer>
            </EntityInfoCollapsibleSection>
        );

    const formattedObjectFields: EntityField[] = formatObjectInfoFields(objectInformation);

    return (
        <EntityInfoCollapsibleSection label='Object Information'>
            <FieldsContainer>
                <BasicObjectInfoFields {...objectInformation} />
                <ObjectInfoFields fields={formattedObjectFields} />
            </FieldsContainer>
        </EntityInfoCollapsibleSection>
    );
};

export default EntityObjectInformation;
