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

import { isDeepEqual } from './object';

describe('isDeepEqual', () => {
    it('returns true when objects have the same recursive value', () => {
        const objA = { a: 1, b: { c: false, d: 'true' } };
        const objB = { b: { d: 'true', c: false }, a: 1 };

        expect(isDeepEqual(objA, objB)).toBeTruthy();
    });

    it('returns false if objects do not have the same recursive value', () => {
        const objA = { a: 1, b: { c: false, d: 'true' } };
        const objB = { b: { d: 'true', c: false }, a: 2 };

        expect(isDeepEqual(objA, objB)).toBeFalsy();
    });

    it('returns false if objects do not have the same number of keys', () => {
        const objA = { a: 1, b: 2 };
        const objB = { a: 1 };

        expect(isDeepEqual(objA, objB)).toBeFalsy();
    });

    it('returns false if objects do not have the same keys', () => {
        const objA = { a: 1, b: 2 };
        const objB = { a: 1, c: 2 };

        expect(isDeepEqual(objA, objB)).toBeFalsy();
    });

    it('returns true if arrays have the same values in the same order', () => {
        const objA = ['a', 'b'];
        const objB = ['a', 'b'];

        expect(isDeepEqual(objA, objB)).toBeTruthy();
    });

    it('returns false if arrays have different lengths', () => {
        const objA = ['a', 'b'];
        const objB = ['a'];

        expect(isDeepEqual(objA, objB)).toBeFalsy();
    });

    it('returns false if arrays have the same values in a different order', () => {
        const objA = ['a', 'b'];
        const objB = ['b', 'a'];

        expect(isDeepEqual(objA, objB)).toBeFalsy();
    });

    it('returns false if objects are not the same type', () => {
        const objA = { '0': 'a' };
        const objB = ['a'];

        expect(isDeepEqual(objA, objB)).toBeFalsy();
    });

    it('returns true if primatives have the same value', () => {
        expect(isDeepEqual('a', 'a')).toBeTruthy();
        expect(isDeepEqual(1, 1)).toBeTruthy();
        expect(isDeepEqual(true, true)).toBeTruthy();
        expect(isDeepEqual(undefined, undefined)).toBeTruthy();
    });
});
