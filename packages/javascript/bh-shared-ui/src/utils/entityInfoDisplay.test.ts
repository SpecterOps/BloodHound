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
    ActiveDirectoryNodeKind,
    AzureKindProperties,
    CommonKindProperties,
} from '../graphSchema';
import { SelectedNode } from '../types';
import {
    AD_NEVER_VALUE,
    AD_UNKNOWN_VALUE,
    ADSpecificTimeProperties,
    DATE_FIELDS,
    EntityField,
    formatADSpecificTime,
    formatBoolean,
    formatDateString,
    formatList,
    formatNumber,
    formatPrimitive,
    getEntityName,
    NoEntitySelectedHeader,
    validateProperty,
} from './entityInfoDisplay';

describe('Handling value formatting for Active Directory entity properties lastlogon, lastlogontimestamp, whencreated, and pwdlastset', () => {
    test('whencreated', () => {
        expect(formatADSpecificTime(-1, ADSpecificTimeProperties.WHEN_CREATED)).toEqual(AD_UNKNOWN_VALUE);
        expect(formatADSpecificTime(0, ADSpecificTimeProperties.WHEN_CREATED)).toEqual(AD_UNKNOWN_VALUE);
        expect(formatADSpecificTime(1694549003, ADSpecificTimeProperties.WHEN_CREATED)).toEqual(
            '2023-09-12 13:03 PDT (GMT-0700)'
        );
    });
    test('lastlogon, lastlogontimestamp', () => {
        expect(formatADSpecificTime(-1, ADSpecificTimeProperties.LAST_LOGON)).toEqual(AD_NEVER_VALUE);
        expect(formatADSpecificTime(-1, ADSpecificTimeProperties.LAST_LOGON_TIMESTAMP)).toEqual(AD_NEVER_VALUE);
        expect(formatADSpecificTime(0, ADSpecificTimeProperties.LAST_LOGON)).toEqual(AD_UNKNOWN_VALUE);
        expect(formatADSpecificTime(0, ADSpecificTimeProperties.LAST_LOGON_TIMESTAMP)).toEqual(AD_UNKNOWN_VALUE);
        expect(formatADSpecificTime(1694549003, ADSpecificTimeProperties.LAST_LOGON)).toEqual(
            '2023-09-12 13:03 PDT (GMT-0700)'
        );
        expect(formatADSpecificTime(1694549003, ADSpecificTimeProperties.LAST_LOGON_TIMESTAMP)).toEqual(
            '2023-09-12 13:03 PDT (GMT-0700)'
        );
    });
    test('pwdlastset', () => {
        expect(formatADSpecificTime(-1, ADSpecificTimeProperties.PASSWORD_LAST_SET)).toEqual(AD_NEVER_VALUE);
        expect(formatADSpecificTime(0, ADSpecificTimeProperties.PASSWORD_LAST_SET)).toEqual(
            'ACCOUNT CREATED BUT NO PASSWORD SET'
        );
        expect(formatADSpecificTime(1694549003, ADSpecificTimeProperties.PASSWORD_LAST_SET)).toEqual(
            '2023-09-12 13:03 PDT (GMT-0700)'
        );
    });
});

describe('Formatting number properties', () => {
    it('handles unix time values by converting them to our standard date display format', () => {
        expect(formatNumber(1694549003)).toEqual('2023-09-12 13:03 PDT (GMT-0700)');
    });
    it('handles percent values by multiplying by 100, truncating to an integer, and adding a "%" to the end', () => {
        expect(formatNumber(0.23)).toEqual('23%');
        expect(formatNumber(0.234)).toEqual('23%');
        expect(formatNumber(0.2345)).toEqual('23%');
        expect(formatNumber(0.235)).toEqual('24%');
        expect(formatNumber(0.23456789)).toEqual('23%');
    });

    it('handles specific Active Directory properties differently than Azure derived properties', () => {
        expect(formatNumber(0, 'ad', 'whencreated')).toEqual(AD_UNKNOWN_VALUE);
        //A value of 0 will not be held by azure property whencreated but this demonstrated handling the values differently
        expect(formatNumber(0, 'az', 'whencreated')).toEqual('0');
    });

    it('properly displays bitwise integers', () => {
        // Base happy path
        expect(formatNumber(31, 'ad', 'certificatemappingmethodsraw')).toEqual('31 (0x1F)');
        // Honors specified padding
        expect(formatNumber(4, 'ad', 'certificatemappingmethodsraw')).toEqual('4 (0x04)');
    });
});

describe('Formatting boolean properties', () => {
    it('uppercases boolean values', () => {
        expect(formatBoolean(true)).toEqual('TRUE');
        expect(formatBoolean(false)).toEqual('FALSE');
    });
});

describe('Formatting string properties', () => {
    it('handles ISO 8601 formatted date strings and converts it into our standard date display format', () => {
        expect(formatDateString('2011-10-05T14:48:00.000Z')).toEqual('2011-10-05 07:48 PDT (GMT-0700)');
        expect(formatDateString('2016')).not.toEqual('2016');
    });
});

describe('Formatting strings via formatPrimive', () => {
    it('returns a date string only if passed a string with a matching key', () => {
        for (const dateField of DATE_FIELDS) {
            expect(formatPrimitive('2016', null, dateField!)).toEqual('2016-01-01 00:00 PST (GMT-0800)');
            expect(formatPrimitive('2016', null, dateField!)).not.toEqual('2016');
        }

        expect(formatPrimitive('2016', null, 'any_other_field')).toEqual('2016');
        expect(formatPrimitive('2016', null, 'any_other_field')).not.toEqual('2016-01-01 00:00 PST (GMT-0800)');

        // With no field supplied, parse as a date
        expect(formatPrimitive('2016')).toEqual('2016-01-01 00:00 PST (GMT-0800)');
        expect(formatPrimitive('2016')).not.toEqual('2016');
    });
});

describe('Formatting list properties', () => {
    it('will return a list of values', () => {
        const testEntityField: EntityField = {
            value: ['test', 5, false],
            label: 'test',
        };
        expect(formatList(testEntityField)).toEqual(['test', '5', 'FALSE']);
    });
});

describe('validating a node property against the shared generated schema', () => {
    it('should recognize active directory properties', () => {
        Object.values(ActiveDirectoryKindProperties).forEach((property: ActiveDirectoryKindProperties) => {
            expect(validateProperty(property)).toEqual({ isKnownProperty: true, kind: 'ad' });
        });
    });
    it('should recognize azure properties', () => {
        Object.values(AzureKindProperties).forEach((property: AzureKindProperties) => {
            expect(validateProperty(property)).toEqual({ isKnownProperty: true, kind: 'az' });
        });
    });
    it('should recognize a common properties', () => {
        Object.values(CommonKindProperties).forEach((property: CommonKindProperties) => {
            expect(validateProperty(property)).toEqual({ isKnownProperty: true, kind: 'cm' });
        });
    });
    it('should return an object denoting that the property is not in the schema when it is unrecognized', () => {
        expect(validateProperty('notInSchema')).toEqual({ isKnownProperty: false, kind: null });
    });
});

describe('Evaluating the entity display name from a given entity', () => {
    it('should handle an undefined or null entity', () => {
        expect(getEntityName(null)).toBe(NoEntitySelectedHeader);
        expect(getEntityName(undefined)).toBe(NoEntitySelectedHeader);
    });
    it('should handle an entity that has an empty name property', () => {
        expect(getEntityName({ id: '1', type: ActiveDirectoryNodeKind.User, name: '' })).toBe('Name not found');
    });
    it('should handle an entity that has no name property', () => {
        expect(getEntityName({ id: '1', type: ActiveDirectoryNodeKind.User } as SelectedNode)).toBe('Name not found');
    });
    it('should handle the well formed entities', () => {
        expect(getEntityName({ id: '1', type: ActiveDirectoryNodeKind.User, name: 'foo' })).toBe('foo');
    });
});
