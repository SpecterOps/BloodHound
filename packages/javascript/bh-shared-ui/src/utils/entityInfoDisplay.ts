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

export type EntityPropertyKind = 'ad' | 'az' | 'cm' | null;

export type EntityField = {
    label: string;
    value: string | number | boolean | any[];
    kind?: EntityPropertyKind;
    keyprop?: string;
};

export enum AmbiguousTimeProperties {
    WHEN_CREATED = 'whencreated',
    LAST_LOGON = 'lastlogon',
    LAST_LOGON_TIMESTAMP = 'lastlogontimestamp',
    PASSWORD_LAST_SET = 'pwdlastset',
}

export const formatAmbiguousTime = (ambiguousTime: number, keyprop: AmbiguousTimeProperties): string => {
    const unknownValue = 'UNKNOWN';
    const neverValue = 'NEVER';

    switch (keyprop) {
        case AmbiguousTimeProperties.WHEN_CREATED: {
            if (ambiguousTime === 0 || ambiguousTime === -1) return unknownValue;
            return DateTime.fromSeconds(ambiguousTime).toFormat(LuxonFormat.DATETIME);
        }
        case AmbiguousTimeProperties.LAST_LOGON: //fallthrough
        case AmbiguousTimeProperties.LAST_LOGON_TIMESTAMP: {
            if (ambiguousTime === 0) return unknownValue;
            if (ambiguousTime === -1) return neverValue;
            return DateTime.fromSeconds(ambiguousTime).toFormat(LuxonFormat.DATETIME);
        }
        case AmbiguousTimeProperties.PASSWORD_LAST_SET:
            if (ambiguousTime === 0) return 'ACCOUNT CREATED BUT NO PASSWORD SET';
            if (ambiguousTime === -1) return neverValue;
            return DateTime.fromSeconds(ambiguousTime).toFormat(LuxonFormat.DATETIME);
        default:
            return '';
    }
};

export const formatNumber = (value: number, kind?: EntityPropertyKind, keyprop?: string): string => {
    const isAmbiguousTimeValue =
        kind === 'ad' && Object.values(AmbiguousTimeProperties).includes(keyprop as AmbiguousTimeProperties);

    if (isAmbiguousTimeValue) return formatAmbiguousTime(value, keyprop as AmbiguousTimeProperties);

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
