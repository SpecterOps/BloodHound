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

import { abbreviatedNumber, commaSeparatedNumber } from './numberFormatting';

describe('numberFormatting', () => {
    it('returns numbers as strings', () => {
        const result = abbreviatedNumber(1);
        expect(typeof result).toBe('string');
    });
    it('abbreviates numbers larger than 1000 and should not round the tenth place digit ', () => {
        const result = abbreviatedNumber(9826);
        expect(result).toBe('9.8K');
    });
    it('does not abbreviate numbers < 1000', () => {
        const result = abbreviatedNumber(123);
        expect(result).toBe('123');
    });
    it('abbreviates thousands, millions, billions, trillions as expected', () => {
        const thousands = abbreviatedNumber(1842);
        const millions = abbreviatedNumber(31000000);
        const billions = abbreviatedNumber(220000000000);
        const trillions = abbreviatedNumber(8700000000000);

        expect(thousands).toBe('1.8K');
        expect(millions).toBe('31.0M');
        expect(billions).toBe('220.0B');
        expect(trillions).toBe('8.7T');
    });
    it('returns a comma separated number when provided with a number > 999', () => {
        const formattedThousandNumber = commaSeparatedNumber(1842);
        const formattedMillionNumber = commaSeparatedNumber(1000000);
        const formattedBillionNumber = commaSeparatedNumber(1000000000);

        expect(formattedThousandNumber).toBe('1,842');
        expect(formattedMillionNumber).toBe('1,000,000');
        expect(formattedBillionNumber).toBe('1,000,000,000');
    });

    const THRESHOLDS = {
        abbreviationThreshold: 100000,
        decimalDigitThreshold: 1000000,
    };
    it('EXTENDED ROUNDING when thresholds are set: abbreviates numbers larger than 1000 and should not round the tenth place digit ', () => {
        const result = abbreviatedNumber(9826, 1, THRESHOLDS);
        expect(result).toBe('9,826');
    });
    it('EXTENDED ROUNDING when thresholds are set: does not abbreviate numbers < 10000', () => {
        const resultHundreds = abbreviatedNumber(123, 1, THRESHOLDS);
        expect(resultHundreds).toBe('123');
        const resultThousands = abbreviatedNumber(1234, 1, THRESHOLDS);
        expect(resultThousands).toBe('1,234');
        const resultTenThousands = abbreviatedNumber(12345, 1, THRESHOLDS);
        expect(resultTenThousands).toBe('12,345');
    });
    it('EXTENDED ROUNDING when thresholds are set: abbreviates hundred-thousands, millions, billions, trillions as expected', () => {
        const thousands = abbreviatedNumber(1842, 1, THRESHOLDS);
        const hundredThousands = abbreviatedNumber(593200, 3, THRESHOLDS);
        const millions = abbreviatedNumber(31760000, 3, THRESHOLDS);
        const billions = abbreviatedNumber(226500000000, 3, THRESHOLDS);
        const trillions = abbreviatedNumber(8754000000000, 3, THRESHOLDS);

        expect(thousands).not.toBe('1.8K');
        expect(thousands).toBe('1,842');
        expect(hundredThousands).toBe('593K');
        expect(millions).toBe('31.760M');
        expect(billions).toBe('226.500B');
        expect(trillions).toBe('8.754T');
    });

    it('EXTENDED ROUNDING with different thresholds', () => {
        const ALTERNATIVE_THRESHOLDS = {
            abbreviationThreshold: 1000,
            decimalDigitThreshold: 1000000000000,
        };
        const thousands = abbreviatedNumber(1842, 3, ALTERNATIVE_THRESHOLDS);
        const hundredThousands = abbreviatedNumber(593200, 3, ALTERNATIVE_THRESHOLDS);
        const millions = abbreviatedNumber(31760000, 3, ALTERNATIVE_THRESHOLDS);
        const billions = abbreviatedNumber(226500000000, 3, ALTERNATIVE_THRESHOLDS);
        const trillions = abbreviatedNumber(8754000000000, 3, ALTERNATIVE_THRESHOLDS);

        expect(thousands).not.toBe('1.8K');
        expect(thousands).toBe('2K');
        expect(hundredThousands).toBe('593K');
        expect(millions).toBe('32M');
        expect(billions).toBe('227B');
        expect(trillions).toBe('8.754T');
    });
    it('EXTENDED ROUNDING when thresholds are set: returns a comma separated number when provided with a number > 999', () => {
        const formattedThousandNumber = commaSeparatedNumber(1842);
        const formattedMillionNumber = commaSeparatedNumber(1000000);
        const formattedBillionNumber = commaSeparatedNumber(1000000000);

        expect(formattedThousandNumber).toBe('1,842');
        expect(formattedMillionNumber).toBe('1,000,000');
        expect(formattedBillionNumber).toBe('1,000,000,000');
    });
});
