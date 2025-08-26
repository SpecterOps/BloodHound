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

import { getSimpleDuration, toFormatted } from './datetime';

describe('getSimpleDuration', () => {
    it('calculates the running time of a job and converts to a human readable format', () => {
        expect(getSimpleDuration('2023-01-01T00:00:00Z', '2023-01-01T00:01:00Z')).toBe('1 min');
        expect(getSimpleDuration('2023-01-01T00:00:00Z', '2023-01-01T00:05:00Z')).toBe('5 mins');
        expect(getSimpleDuration('2023-01-01T00:00:00Z', '2023-01-01T00:30:00Z')).toBe('30 mins');
        expect(getSimpleDuration('2023-01-01T00:00:00Z', '2023-01-02T00:00:00Z')).toBe('a day');
        expect(getSimpleDuration('2023-01-01T00:00:00Z', '2023-01-03T00:00:00Z')).toBe('2 days');
    });

    it('rounds down fractional minutes/days', () => {
        expect(getSimpleDuration('2023-01-01T00:00:00Z', '2023-01-01T00:01:30Z')).toBe('1 min');
        expect(getSimpleDuration('2023-01-01T00:00:00Z', '2023-01-01T00:05:55Z')).toBe('5 mins');
        expect(getSimpleDuration('2023-01-01T00:00:00Z', '2023-01-02T01:20:40Z')).toBe('a day');
        expect(getSimpleDuration('2023-01-01T00:00:00Z', '2023-01-03T12:30:10Z')).toBe('2 days');
    });
});

describe('toFormatted', () => {
    it('formats the date string', () => {
        const result = toFormatted('2024-01-01T15:30:00.500Z');
        // Server TZ might not match local dev TZ
        // Allow 'CST'/'UTC' or 'GMT-06:00' style outputs depending on environment
        expect(result).toMatch(/^\d{4}-\d{2}-\d{2} \d{2}:\d{2} (?:[A-Z]{3,4}|GMT[+-]\d{1,2}:\d{2})$/);
    });
});
