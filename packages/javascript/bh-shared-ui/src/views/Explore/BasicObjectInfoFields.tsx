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
import { ActiveDirectoryNodeKind, AzureNodeKind } from '../../graphSchema';
import { EntityKinds } from '../../utils';
import { SearchValue } from './ExploreSearch/types';
import { Field } from './fragments';

interface BasicObjectInfoFieldsProps {
    displayname?: string;
    grouplinkid?: string;
    handleSourceNodeSelected?: (sourceNode: SearchValue) => void;
    isOwned?: boolean;
    isTierZero?: boolean;
    name?: string;
    noderesourcegroupid?: string;
    nodeType?: string;
    objectid: string;
    service_principal_id?: string;
}

const RelatedKindField = (
    onSourceNodeSelected: (sourceNode: SearchValue) => void,
    fieldLabel: string,
    relatedKind: EntityKinds,
    id: string,
    name?: string
) => {
    return (
        <Box padding={1}>
            <Box fontWeight='bold' mr={1}>
                {fieldLabel}
            </Box>
            <br />
            <Box display='flex' flexDirection='row' flexWrap='wrap' justifyContent='flex-start'>
                <NodeIcon nodeType={relatedKind} />
                <Box
                    onClick={() => onSourceNodeSelected({ objectid: id, type: relatedKind, name: name || '' })}
                    style={{ cursor: 'pointer' }}
                    overflow='hidden'
                    textOverflow='ellipsis'
                    title={id}>
                    {id}
                </Box>
            </Box>
        </Box>
    );
};

export const BasicObjectInfoFields: React.FC<BasicObjectInfoFieldsProps> = (props): JSX.Element => {
    return (
        <>
            {props.nodeType && <Field label='Node Type' value={props.nodeType} />}
            {props.isTierZero && <Field label='Tier Zero:' value={true} />}
            {props.isOwned && <Field label='Owned Object:' value={true} />}
            {props.displayname && <Field label='Display Name:' value={props.displayname} />}
            <Field label='Object ID:' value={props.objectid} />
            {props.handleSourceNodeSelected && (
                <>
                    {props.service_principal_id &&
                        RelatedKindField(
                            props.handleSourceNodeSelected,
                            'Service Principal ID:',
                            AzureNodeKind.ServicePrincipal,
                            props.service_principal_id,
                            props.name
                        )}
                    {props.noderesourcegroupid &&
                        RelatedKindField(
                            props.handleSourceNodeSelected,
                            'Node Resource Group ID:',
                            AzureNodeKind.ResourceGroup,
                            props.noderesourcegroupid,
                            props.name
                        )}
                    {props.grouplinkid &&
                        RelatedKindField(
                            props.handleSourceNodeSelected,
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
