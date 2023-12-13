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

import { Box } from '@mui/material';
import { NodeIcon, Field, AzureNodeKind, EntityKinds } from 'bh-shared-ui';
import { TIER_ZERO_TAG } from 'src/constants';
import { sourceNodeSelected } from 'src/ducks/searchbar/actions';
import { useAppDispatch } from 'src/store';

interface BasicObjectInfoFieldsProps {
    objectid: string;
    displayname?: string;
    system_tags?: string;
    service_principal_id?: string;
    noderesourcegroupid?: string;
    name?: string;
}

const RelatedKindField = (fieldLabel: string, relatedKind: EntityKinds, id: string, name?: string) => {
    const dispatch = useAppDispatch();
    return (
        <Box padding={1}>
            <Box fontWeight='bold' mr={1}>
                {fieldLabel}
            </Box>
            <br />
            <Box display='flex' flexDirection='row' flexWrap='wrap' justifyContent='flex-start'>
                <NodeIcon nodeType={relatedKind} />
                <Box
                    onClick={() => {
                        dispatch(
                            sourceNodeSelected({
                                objectid: id,
                                type: relatedKind,
                                name: name || '',
                            })
                        );
                    }}
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
            {props.system_tags?.includes(TIER_ZERO_TAG) && <Field label='Tier Zero:' value={true} />}
            {props.displayname && <Field label='Display Name:' value={props.displayname} />}
            <Field label='Object ID:' value={props.objectid} />
            {props.service_principal_id &&
                RelatedKindField(
                    'Service Principal ID:',
                    AzureNodeKind.ServicePrincipal,
                    props.service_principal_id,
                    props.name
                )}
            {props.noderesourcegroupid &&
                RelatedKindField(
                    'Node Resource Group ID:',
                    AzureNodeKind.ResourceGroup,
                    props.noderesourcegroupid,
                    props.name
                )}
        </>
    );
};
