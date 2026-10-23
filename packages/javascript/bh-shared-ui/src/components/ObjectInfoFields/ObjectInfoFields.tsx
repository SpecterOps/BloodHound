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

import React from 'react';
import { ActiveDirectoryKindProperties, AzureKindProperties, CommonKindProperties } from '../../graphSchema';
import { EntityField } from '../../utils';
import { Field } from './Field';
import { ObjectInfoFieldLayout } from './types';

export const exclusionList = [
    'gid',
    'admin_rights_count',
    'admin_rights_risk_percent',
    ActiveDirectoryKindProperties.HasSPN,
    CommonKindProperties.SystemTags,
    'isTierZero',
    CommonKindProperties.UserTags,
    'neo4jImportId',
    CommonKindProperties.Name,
    CommonKindProperties.ObjectID,
    CommonKindProperties.DisplayName,
    AzureKindProperties.ServicePrincipalID,
    AzureKindProperties.FederatedIdentityCredentialAppID,
    'serverreferencecomputer',
    'serverreferencecomputername',
    'siteservernode',
    'siteservernodename',
    'highvalue',
    'reconcile',
    ActiveDirectoryKindProperties.InheritanceHashes,
    ActiveDirectoryKindProperties.InheritanceHash,
];

const filterNegatedFields = (fields: EntityField[]): EntityField[] =>
    fields.filter((field: EntityField) => !exclusionList.includes(field.keyprop || ''));

export const ObjectInfoFields: React.FC<{ fields: EntityField[]; layout?: ObjectInfoFieldLayout }> = ({
    fields,
    layout,
}): JSX.Element => {
    const filteredFields = filterNegatedFields(fields);

    return (
        <>
            {filteredFields.map((field: EntityField) => {
                return (
                    <Field
                        kind={field.kind}
                        label={field.label}
                        value={field.value}
                        keyprop={`${field.keyprop}`}
                        key={`${field.keyprop}-${field.label}`}
                        layout={layout}
                    />
                );
            })}
        </>
    );
};
