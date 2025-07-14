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
import { areArraysSimilar } from './array';

describe('areArraysSimilar', () => {
    const primeArray = [1, 2, 3, 3, 4];
    const equalArray = [1, 2, 3, 3, 4];
    const reorderedArray = [3, 2, 3, 4, 1];
    const shorterArray = [1, 2, 3, 4];
    const differentArray = ['a', 'b', 'c', 'd'];

    it('returns true if arrays are referentially equal', () => {
        expect(areArraysSimilar(primeArray, primeArray)).toEqual(true);
    });

    it('returns true if arrays are equal', () => {
        expect(areArraysSimilar(primeArray, equalArray)).toEqual(true);
    });

    it('returns true if both arrays are empty', () => {
        expect(areArraysSimilar([], [])).toEqual(true);
    });

    it('returns true if arrays have same content in different order', () => {
        expect(areArraysSimilar(primeArray, reorderedArray)).toEqual(true);
    });

    it('returns false if arrays have different length', () => {
        expect(areArraysSimilar(primeArray, shorterArray)).toEqual(false);
    });

    it('returns false if arrays have different elements', () => {
        // @ts-expect-error: types differ to test negative case
        expect(areArraysSimilar(primeArray, differentArray)).toEqual(false);
    });

    it('returns false for arrays with different object references', () => {
        const objArray1 = [{ id: 1 }, { id: 2 }];
        const objArray2 = [{ id: 1 }, { id: 2 }];
        expect(areArraysSimilar(objArray1, objArray2)).toEqual(false);
    });

    it('accepts a sort and compare function', () => {
        const objArray1 = [{ id: 1 }, { id: 2 }];
        const objArray2 = [{ id: 1 }, { id: 2 }];
        expect(
            areArraysSimilar(
                objArray1,
                objArray2,
                (i, j) => i.id - j.id,
                (i, j) => i.id === j.id
            )
        ).toEqual(true);
    });
});
