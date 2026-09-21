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

import { Box } from '@mui/material';
import NodeIcon from '../../components/NodeIcon';
import { ActiveDirectoryNodeKind, AzureNodeKind, CommonKindProperties } from '../../graphSchema';
import { EntityKinds, KnownNodeProperties, formatPotentiallyUnknownLabel } from '../../utils';
import { SearchValue } from './ExploreSearch/types';
import { Field } from './fragments';

interface BasicObjectInfoFieldsProps {
    properties: Record<string, any> & {
        displayname?: string;
        grouplinkid?: string;
        isOwnedObject?: boolean;
        isTierZero?: boolean;
        name?: string;
        noderesourcegroupid?: string;
        objectid?: string;
        serverreferencecomputer?: string;
        serverreferencecomputername?: string;
        service_principal_id?: string;
        siteservernode?: string;
        siteservernodename?: string;
        federatedidentitycredentialappid?: string;
    };
    handleSourceNodeSelected?: (sourceNode: SearchValue) => void;
    nodeType?: string;
    zone?: string;
}

const RelatedKindField = (
    onSourceNodeSelected: (sourceNode: SearchValue) => void,
    fieldLabel: string,
    relatedKind: EntityKinds,
    id: string,
    name?: string,
    displayValue?: string
) => {
    const value = displayValue || id;

    return (
        <Box padding={1}>
            <Box fontWeight='bold' mr={1}>
                {fieldLabel}
            </Box>
            <br />
            <Box display='flex' flexDirection='row' flexWrap='wrap' justifyContent='flex-start'>
                <NodeIcon nodeType={relatedKind} />
                <Box
                    onClick={() =>
                        onSourceNodeSelected({ objectid: id, type: relatedKind, name: name || displayValue || '' })
                    }
                    style={{ cursor: 'pointer' }}
                    overflow='hidden'
                    textOverflow='ellipsis'
                    title={value}>
                    {value}
                </Box>
            </Box>
        </Box>
    );
};

const basicObjectFields = [
    'zone',
    'nodeType',
    'isTierZero',
    'isOwnedObject',
    CommonKindProperties.DisplayName,
    CommonKindProperties.ObjectID,
] satisfies (KnownNodeProperties | CommonKindProperties | 'zone')[];

export const BasicObjectInfoFields: React.FC<BasicObjectInfoFieldsProps> = ({
    properties: props,
    handleSourceNodeSelected,
    nodeType,
    zone,
}): JSX.Element => {
    const fieldValues = { ...props, nodeType, zone };
    return (
        <>
            {basicObjectFields.map((field) => {
                const value = fieldValues[field];
                if (value === undefined) return null; // <Field /> doesn't support undefined values

                return <Field key={field} label={`${formatPotentiallyUnknownLabel(field) ?? field}:`} value={value} />;
            })}
            {handleSourceNodeSelected && (
                <>
                    {props.service_principal_id &&
                        RelatedKindField(
                            handleSourceNodeSelected,
                            'Service Principal ID:',
                            AzureNodeKind.ServicePrincipal,
                            props.service_principal_id,
                            props.name
                        )}
                    {props.serverreferencecomputer &&
                        RelatedKindField(
                            handleSourceNodeSelected,
                            'Referenced Computer:',
                            ActiveDirectoryNodeKind.Computer,
                            props.serverreferencecomputer,
                            props.serverreferencecomputername,
                            props.serverreferencecomputername
                        )}
                    {props.siteservernode &&
                        RelatedKindField(
                            handleSourceNodeSelected,
                            'Site Server:',
                            ActiveDirectoryNodeKind.SiteServer,
                            props.siteservernode,
                            props.siteservernodename,
                            props.siteservernodename
                        )}
                    {props.federatedidentitycredentialappid &&
                        RelatedKindField(
                            handleSourceNodeSelected,
                            'Federated Identity Credential Application ID:',
                            AzureNodeKind.App,
                            props.federatedidentitycredentialappid,
                            props.name
                        )}
                    {props.noderesourcegroupid &&
                        RelatedKindField(
                            handleSourceNodeSelected,
                            'Node Resource Group ID:',
                            AzureNodeKind.ResourceGroup,
                            props.noderesourcegroupid,
                            props.name
                        )}
                    {props.grouplinkid &&
                        RelatedKindField(
                            handleSourceNodeSelected,
                            'Linked Group ID:',
                            ActiveDirectoryNodeKind.Group,
                            props.grouplinkid,
                            props.name
                        )}
                </>
            )}
        </>
    );
};
