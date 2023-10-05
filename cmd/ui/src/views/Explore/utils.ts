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

import {
    ActiveDirectoryKindProperties,
    ActiveDirectoryKindPropertiesToDisplay,
    AzureKindProperties,
    AzureKindPropertiesToDisplay,
    CommonKindProperties,
    CommonKindPropertiesToDisplay,
    EntityField,
    EntityPropertyKind,
} from 'bh-shared-ui';
import isEmpty from 'lodash/isEmpty';
import startCase from 'lodash/startCase';
import { ZERO_VALUE_API_DATE } from 'src/constants';

export let controller = new AbortController();

export const abortRequest = () => {
    controller.abort();
    controller = new AbortController();
};

export const formatObjectInfoFields = (props: any): EntityField[] => {
    let mappedFields: EntityField[] = [];
    const propKeys = Object.keys(props);

    for (let i = 0; i < propKeys.length; i++) {
        const value = props[propKeys[i]];
        // Don't display empty fields or fields with zero date values
        if (
            value === undefined ||
            value === '' ||
            value === ZERO_VALUE_API_DATE ||
            (typeof value === 'object' && isEmpty(value))
        )
            continue;

        const { kind, isKnownProperty } = validateProperty(propKeys[i]);

        if (isKnownProperty) {
            mappedFields.push({
                kind: kind,
                label: getFieldLabel(kind!, propKeys[i]),
                value: value,
                keyprop: propKeys[i],
            });
        } else {
            mappedFields.push({
                kind: kind,
                label: `${startCase(propKeys[i])}:`,
                value: value,
                keyprop: propKeys[i],
            });
        }
    }

    mappedFields = mappedFields.sort((a, b) => {
        return a.label!.localeCompare(b.label!);
    });

    return mappedFields;
};

const isActiveDirectoryProperty = (enumValue: ActiveDirectoryKindProperties): boolean => {
    return Object.values(ActiveDirectoryKindProperties).includes(enumValue);
};

const isAzureProperty = (enumValue: AzureKindProperties): boolean => {
    return Object.values(AzureKindProperties).includes(enumValue);
};

const isCommonProperty = (enumValue: CommonKindProperties): boolean => {
    return Object.values(CommonKindProperties).includes(enumValue);
};

export type ValidatedProperty = {
    isKnownProperty: boolean;
    kind: EntityPropertyKind;
};

export const validateProperty = (enumValue: string): ValidatedProperty => {
    if (isActiveDirectoryProperty(enumValue as ActiveDirectoryKindProperties))
        return { isKnownProperty: true, kind: 'ad' };
    if (isAzureProperty(enumValue as AzureKindProperties)) return { isKnownProperty: true, kind: 'az' };
    if (isCommonProperty(enumValue as CommonKindProperties)) return { isKnownProperty: true, kind: 'cm' };
    return { isKnownProperty: false, kind: null };
};

const getFieldLabel = (kind: string, key: string): string => {
    let label: string;

    switch (kind) {
        case 'ad':
            label = ActiveDirectoryKindPropertiesToDisplay(key as ActiveDirectoryKindProperties)!;
            break;
        case 'az':
            label = AzureKindPropertiesToDisplay(key as AzureKindProperties)!;
            break;
        case 'cm':
            label = CommonKindPropertiesToDisplay(key as CommonKindProperties)!;
            break;
        default:
            label = key;
    }

    return `${label}:`;
};
