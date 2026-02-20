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

import { GraphNode } from 'js-client-library';
import isEmpty from 'lodash/isEmpty';
import startCase from 'lodash/startCase';
import { DateTime } from 'luxon';
import { isValidElement } from 'react';
import { ZERO_VALUE_API_DATE } from '../constants';
import {
    ActiveDirectoryKindProperties,
    ActiveDirectoryKindPropertiesToDisplay,
    AzureKindProperties,
    AzureKindPropertiesToDisplay,
    CommonKindProperties,
    CommonKindPropertiesToDisplay,
} from '../graphSchema';
import { MappedStringLiteral, SelectedNode } from '../types';
import { LuxonFormat } from './datetime';

export const formatPotentiallyUnknownLabel = (propKey: string) => {
    const { kind, isKnownProperty } = validateProperty(propKey);

    return isKnownProperty ? getFieldLabel(kind!, propKey) : `${startCase(propKey)}`;
};

export const formatObjectInfoFields = (props: any): EntityField[] => {
    let mappedFields: EntityField[] = [];
    const propKeys = Object.keys(props || {});

    for (let i = 0; i < propKeys.length; i++) {
        const key = propKeys[i];
        const value = props[key];
        // Don't display empty fields or fields with zero date values
        if (
            value === undefined ||
            value === '' ||
            value === ZERO_VALUE_API_DATE ||
            (typeof value === 'object' && isEmpty(value))
        )
            continue;

        // prevent rendering the zone property twice if it exists since there is explicit handling for it in EntityObjectInformation
        if (key === 'zone') continue;

        const { kind } = validateProperty(key);

        mappedFields.push({
            kind: kind,
            label: `${formatPotentiallyUnknownLabel(key)}:`,
            value: value,
            keyprop: key,
        });
    }

    mappedFields = mappedFields.sort((a, b) => {
        if (isValidElement(a) || isValidElement(b)) {
            return 0;
        }
        // @ts-ignore
        return a.label!.localeCompare(b.label!);
    });

    return mappedFields;
};

export const makeFormattedObjectInfoFieldsMap = (props: any) => {
    const fieldsData: { keyprop?: string }[] = formatObjectInfoFields(props);

    if (fieldsData.length) {
        return fieldsData.reduce(
            (acc, curr) => {
                if (curr?.keyprop) {
                    acc[curr?.keyprop] = curr;
                }

                return acc;
            },
            {} as Record<string, any>
        );
    }
};

// Convert *KindProperties enums to a map to for quick lookup and defined outside of the typeguard so we perform enumerations once
const activeDirectoryKindPropertiesMap = Object.fromEntries(
    Object.values(ActiveDirectoryKindProperties).map((value) => [value, true])
);
const isActiveDirectoryProperty = (enumValue: ActiveDirectoryKindProperties): boolean => {
    return !!activeDirectoryKindPropertiesMap[enumValue];
};

const azureKindPropertiesMap = Object.fromEntries(Object.values(AzureKindProperties).map((value) => [value, true]));
const isAzureProperty = (enumValue: AzureKindProperties): boolean => {
    return !!azureKindPropertiesMap[enumValue];
};

const commonKindPropertiesMap = Object.fromEntries(Object.values(CommonKindProperties).map((value) => [value, true]));
const isCommonProperty = (enumValue: CommonKindProperties): boolean => {
    return commonKindPropertiesMap[enumValue];
};

export type KnownNodeProperties = keyof Omit<GraphNode, 'properties'> | 'nodeType';
/**
 * The intent is to standardize keys and their display label in the UI.
 * The keys below are either deduped with their property bag counterpart, or are assigned a label for standardization across the UI.
 */
export const KnownNodePropertiesToDisplay = {
    /**
     * nodeType is actually a prop defined on EntityInfoContentProps, but we include it with other node properties in BasicObjectInfoFieldsProps.
     * In theory we could refactor this prop to be "kind", however, that seems out of scope for this refactor.
     */
    nodeType: 'Node Type',
    kind: 'Node Type',
    kinds: 'Node Kinds',
    isTierZero: 'Tier Zero',
    isOwnedObject: 'Is Owned',
    label: CommonKindPropertiesToDisplay(CommonKindProperties.Name)!,
    objectId: CommonKindPropertiesToDisplay(CommonKindProperties.ObjectID)!,
    lastSeen: CommonKindPropertiesToDisplay(CommonKindProperties.LastSeen)!,
} as const satisfies MappedStringLiteral<KnownNodeProperties, string>;

export type ValidatedProperty = {
    isKnownProperty: boolean;
    kind: EntityPropertyKind;
};

export const validateProperty = (enumValue: string): ValidatedProperty => {
    if (isActiveDirectoryProperty(enumValue as ActiveDirectoryKindProperties))
        return { isKnownProperty: true, kind: 'ad' };
    if (isAzureProperty(enumValue as AzureKindProperties)) return { isKnownProperty: true, kind: 'az' };
    if (isCommonProperty(enumValue as CommonKindProperties)) return { isKnownProperty: true, kind: 'cm' };
    if (enumValue in KnownNodePropertiesToDisplay) return { isKnownProperty: true, kind: 'ov' };
    return { isKnownProperty: false, kind: null };
};

const getFieldLabel = (kind: EntityPropertyKind, key: string): string => {
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
        case 'ov':
            label = KnownNodePropertiesToDisplay[key as KnownNodeProperties]!;
            break;
        default:
            label = key;
    }

    return label;
};

export type EntityPropertyKind = 'ad' | 'az' | 'cm' | 'ov' | null;

export type EntityField = {
    label: string | JSX.Element;
    value: string | number | boolean | any[];
    kind?: EntityPropertyKind;
    keyprop?: string;
};

export enum ADSpecificTimeProperties {
    WHEN_CREATED = 'whencreated',
    LAST_LOGON = 'lastlogon',
    LAST_LOGON_TIMESTAMP = 'lastlogontimestamp',
    PASSWORD_LAST_SET = 'pwdlastset',
}

export const NoEntitySelectedMessage = 'Select a node to view the associated information';
export const NoEntitySelectedHeader = 'None Selected';

export const getEntityName = (selectedEntity: SelectedNode | null | undefined) => {
    if (!selectedEntity) return NoEntitySelectedHeader;

    const { name } = selectedEntity;

    if (!name) return 'Name not found';

    return name;
};

export const getNodeByDatabaseIdCypher = (id: string): string => `MATCH (n) WHERE ID(n) = ${id} RETURN n LIMIT 1`;

// Map containing all properties that should display as bitwise integers in the entity panel.
// The key is the property string, the value is the amount of significant digits the hex value should display with.
const BitwiseInts = new Map([['certificatemappingmethodsraw', 2]]);

//These times are handled differently specifically for these properties as they are collected and ingested to match these values
//Here is some related documentation:
//https://learn.microsoft.com/en-us/windows/win32/adschema/a-lastlogon
//https://social.technet.microsoft.com/wiki/contents/articles/22461.understanding-the-ad-account-attributes-lastlogon-lastlogontimestamp-and-lastlogondate.aspx
export const AD_NEVER_VALUE = 'NEVER';
export const AD_UNKNOWN_VALUE = 'UNKNOWN';
export const formatADSpecificTime = (timeValue: number, keyprop: ADSpecificTimeProperties): string => {
    switch (keyprop) {
        case ADSpecificTimeProperties.WHEN_CREATED: {
            if (timeValue === 0 || timeValue === -1) return AD_UNKNOWN_VALUE;
            return DateTime.fromSeconds(timeValue).toFormat(LuxonFormat.DATETIME);
        }
        case ADSpecificTimeProperties.LAST_LOGON: //fallthrough
        case ADSpecificTimeProperties.LAST_LOGON_TIMESTAMP: {
            if (timeValue === 0) return AD_UNKNOWN_VALUE;
            if (timeValue === -1) return AD_NEVER_VALUE;
            return DateTime.fromSeconds(timeValue).toFormat(LuxonFormat.DATETIME);
        }
        case ADSpecificTimeProperties.PASSWORD_LAST_SET:
            if (timeValue === 0) return 'ACCOUNT CREATED BUT NO PASSWORD SET';
            if (timeValue === -1) return AD_NEVER_VALUE;
            return DateTime.fromSeconds(timeValue).toFormat(LuxonFormat.DATETIME);
        default:
            return '';
    }
};

export const formatBitwiseInt = (value: number, padding: number): string => {
    return `${value} (0x${value.toString(16).padStart(padding, '0').toUpperCase()})`;
};

export const formatNumber = (value: number, kind?: EntityPropertyKind, keyprop?: string): string => {
    const isAmbiguousTimeValue =
        kind === 'ad' && Object.values(ADSpecificTimeProperties).includes(keyprop as ADSpecificTimeProperties);
    const isBitwiseInt = BitwiseInts.has(keyprop as ActiveDirectoryKindProperties);

    if (isAmbiguousTimeValue) return formatADSpecificTime(value, keyprop as ADSpecificTimeProperties);
    if (isBitwiseInt) return formatBitwiseInt(value, 2);

    //315536400 = January 1st, 1980. Seems like a safe bet
    const secondsLowerBound = 315536400;
    //Time of now
    const secondsUpperBound = Math.round(new Date().getTime() / 1000);

    if (value > secondsLowerBound && value < secondsUpperBound) {
        //Assume this is unix time
        return DateTime.fromSeconds(value).toFormat(LuxonFormat.DATETIME);
    } else if (value > 0 && value < 1) {
        //Assume this is a percentage
        const percent = (value * 100).toFixed(0);
        return `${percent}%`;
    } else {
        //Return a number formatted with commas or periods depending on the locale
        return `${value}`.toLocaleString();
    }
};

export const formatBoolean = (value: boolean): string => value.toString().toUpperCase();

export const formatDateString = (value: string) => {
    const potentialDate: any = DateTime.fromISO(value);

    if (potentialDate.invalid === null) {
        return potentialDate.toFormat(LuxonFormat.DATETIME);
    }

    return value;
};

export const DATE_FIELDS = [
    'lastseen',
    'whencreated',
    'lastlogontimestamp',
    'lastlogon',
    'pwdlastset',
    'lastsuccessfulsignindatetime',
];

export const formatPrimitive = (
    value: string | number | boolean,
    kind?: EntityPropertyKind,
    keyprop?: string
): string => {
    switch (typeof value) {
        case 'number': {
            return formatNumber(value, kind, keyprop);
        }
        case 'boolean': {
            return formatBoolean(value);
        }
        case 'string':
            if (!keyprop || DATE_FIELDS.includes(keyprop)) {
                return formatDateString(value);
            }

            return value;
        default:
            return value;
    }
};

export const formatList = (field: EntityField) => {
    const list = field.value as any[];
    const fields: string[] = [];
    list.forEach((value) => {
        fields.push(formatPrimitive(value, field.kind, field.keyprop));
    });
    return fields;
};

export const format = (field: EntityField): string | string[] => {
    const { value, kind, keyprop } = field;

    if (Array.isArray(value)) {
        return formatList(field);
    } else {
        return formatPrimitive(value, kind, keyprop);
    }
};
