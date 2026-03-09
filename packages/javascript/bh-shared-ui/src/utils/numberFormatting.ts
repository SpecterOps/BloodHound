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

type THRESHOLDS = { abbreviationThreshold: number; decimalDigitThreshold: number };

export const abbreviatedNumber = (
    num: number,
    fractionDigits: number = 1,
    { abbreviationThreshold, decimalDigitThreshold }: THRESHOLDS = {
        abbreviationThreshold: 1000,
        decimalDigitThreshold: 0,
    }
) => {
    // Exit early in case number in response is larger than the max safe integer to avoid doing math on it and getting erronious numbers
    if (!Number.isSafeInteger(num)) return '>9Q';

    const absNum = Math.abs(num);
    if (abbreviationThreshold && absNum < abbreviationThreshold) {
        // If the number is under the abbreviationThreshold, no abbreviation needed
        return commaSeparatedNumber(absNum);
    }

    if (decimalDigitThreshold && absNum < decimalDigitThreshold) {
        // if the number is under the decimalDigitThreshold, add no decimals
        fractionDigits = 0;
    }

    const abbreviations = ['', 'K', 'M', 'B', 'T'];
    const log1000 = Math.floor(Math.log10(absNum) / 3); // appropriate abbreviation index

    // Otherwise, divide the number by the appropriate power of 1000 and add the abbreviation
    const formattedNumber = (absNum / Math.pow(1000, log1000)).toFixed(fractionDigits);
    return formattedNumber + abbreviations[log1000];
};

export const commaSeparatedNumber = (num: number) => {
    return new Intl.NumberFormat('en-US').format(num);
};
