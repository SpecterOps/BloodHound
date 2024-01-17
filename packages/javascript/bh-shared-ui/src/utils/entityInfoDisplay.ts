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

import { DateTime } from 'luxon';
import { LuxonFormat } from './datetime';
import { ActiveDirectoryKindProperties } from '..';

export type EntityPropertyKind = 'ad' | 'az' | 'cm' | null;

export type EntityField = {
    label: string;
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

// Map containing all properties that should display as bitwise integers in the entity panel.
// The key is the property string, the value is the amount of significant digits the hex value should display with.
const BitwiseInts = new Map([['certificatemappingmethodsraw', 2]]);

//These times are handled differently specifically for these properties as they are collected and ingested to match these values
//Here is some related documentation:
//https://learn.microsoft.com/en-us/windows/win32/adschema/a-lastlogon
//https://social.technet.microsoft.com/wiki/contents/articles/22461.understanding-the-ad-account-attributes-lastlogon-lastlogontimestamp-and-lastlogondate.aspx
export const formatADSpecificTime = (timeValue: number, keyprop: ADSpecificTimeProperties): string => {
    const unknownValue = 'UNKNOWN';
    const neverValue = 'NEVER';

    switch (keyprop) {
        case ADSpecificTimeProperties.WHEN_CREATED: {
            if (timeValue === 0 || timeValue === -1) return unknownValue;
            return DateTime.fromSeconds(timeValue).toFormat(LuxonFormat.DATETIME);
        }
        case ADSpecificTimeProperties.LAST_LOGON: //fallthrough
        case ADSpecificTimeProperties.LAST_LOGON_TIMESTAMP: {
            if (timeValue === 0) return unknownValue;
            if (timeValue === -1) return neverValue;
            return DateTime.fromSeconds(timeValue).toFormat(LuxonFormat.DATETIME);
        }
        case ADSpecificTimeProperties.PASSWORD_LAST_SET:
            if (timeValue === 0) return 'ACCOUNT CREATED BUT NO PASSWORD SET';
            if (timeValue === -1) return neverValue;
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

export const formatString = (value: string, keyprop?: string) => {
    const potentialDate: any = DateTime.fromISO(value);

    if (potentialDate.invalid === null && keyprop !== 'functionallevel')
        return potentialDate.toFormat(LuxonFormat.DATETIME);

    return value;
};

const formatPrimitive = (value: string | number | boolean, kind?: EntityPropertyKind, keyprop?: string): string => {
    switch (typeof value) {
        case 'number': {
            return formatNumber(value, kind, keyprop);
        }
        case 'boolean': {
            return formatBoolean(value);
        }
        case 'string': //fallthrough
        default:
            return formatString(value, keyprop);
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
