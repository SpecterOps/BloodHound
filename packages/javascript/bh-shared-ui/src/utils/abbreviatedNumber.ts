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

export const abbreviatedNumber = (num: number, fractionDigits: number = 1) => {
    if (num < 1000) {
        // If the number is less than 1000, no abbreviation needed
        return num.toString();
    }
    const abbreviations = ['', 'K', 'M', 'B', 'T'];
    const log1000 = Math.floor(Math.log10(Math.abs(num)) / 3); // appropriate abbreviation index

    // Otherwise, divide the number by the appropriate power of 1000 and add the abbreviation
    const formattedNumber = (num / Math.pow(1000, log1000)).toFixed(fractionDigits);
    return formattedNumber + abbreviations[log1000];
};
