// Copyright 2026 Specter Ops, Inc.
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
import { clampNumber, getCircumference } from './utils';

describe('RadialGauge utils', () => {
    describe('getCircumference', () => {
        it('correctly calculate the circumference given a radius', () => {
            const expected = 100.53;
            const radius = 16;
            const actual = getCircumference(radius);

            // because the decimal points of a circle are infinite, we must truncate the actual circumference for testing
            const truncatedActual = Math.round(actual * 100) / 100;

            expect(truncatedActual).toBe(expected);
        });
    });

    describe('clampNumber', () => {
        it('returns the value if it falls between the lower and upper bounds', () => {
            const expected = 20;
            const actual = clampNumber(expected, 0, 100);

            expect(actual).toBe(expected);
        });
        it('returns the lower or upper bounds if the value falls outside of those numbers', () => {
            const lower = 0;
            const clampLower = clampNumber(-1, lower, 1);

            expect(clampLower).toBe(lower);

            const upper = 2;
            const clampUpper = clampNumber(3, 1, upper);

            expect(clampUpper).toBe(upper);
        });
        it('returns lower bound if the lower bound is greater than the upper', () => {
            const lower = 3;
            const warnMock = vi.spyOn(console, 'warn').mockImplementation(() => {});
            const actual = clampNumber(2, 3, 1);

            expect(actual).toBe(lower);
            expect(warnMock).toBeCalled();
        });
    });
});
