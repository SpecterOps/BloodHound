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
import { DateTime } from 'luxon';
import { LuxonFormat } from '../../..';
import { createHistoryParams, PAGE_SIZE } from './utils';

const formatDateRange = (date: string) => ({
    start: DateTime.fromFormat(date, LuxonFormat.ISO_8601).startOf('day').toISO(),
    end: DateTime.fromFormat(date, LuxonFormat.ISO_8601).endOf('day').toISO(),
});

describe('createHistoryParams', () => {
    const baseFilters = {
        'start-date': '2023-01-01',
        'end-date': '2023-01-10',
        tagId: '42',
        madeBy: 'alice',
        action: 'update',
    };

    it('constructs correct URLSearchParams with all filters', () => {
        const params = createHistoryParams(1, baseFilters);
        const { start, end } = formatDateRange(baseFilters['start-date']);

        expect(params.get('limit')).toBe(PAGE_SIZE.toString());
        expect(params.get('skip')).toBe('0');
        expect(params.get('action')).toBe('eq:update');
        expect(params.get('asset_group_tag_id')).toBe('eq:42');
        expect(params.get('actor')).toBe('eq:alice');
        expect(params.getAll('created_at')).toEqual(['gte:' + start, 'lte:' + end]);
    });

    it('calculates skip correctly for pageParam > 1', () => {
        const params = createHistoryParams(3, baseFilters); // Page 3
        expect(params.get('skip')).toBe(((3 - 1) * PAGE_SIZE).toString());
    });

    it('adds email filter if madeBy contains "@"', () => {
        const filters = { ...baseFilters, madeBy: 'user@example.com' };
        const params = createHistoryParams(1, filters);

        expect(params.has('email')).toBe(true);
        expect(params.get('email')).toBe('eq:user@example.com');
        expect(params.has('actor')).toBe(false);
    });

    it('does not add optional filters if not present', () => {
        const minimalFilters = {
            'start-date': '2023-01-01',
            'end-date': '2023-01-01',
            tagId: '',
            madeBy: '',
            action: '',
        };
        const { start, end } = formatDateRange(minimalFilters['start-date']);
        const params = createHistoryParams(1, minimalFilters);

        expect(params.get('action')).toBeNull();
        expect(params.get('asset_group_tag_id')).toBeNull();
        expect(params.get('email')).toBeNull();
        expect(params.get('actor')).toBeNull();
        expect(params.getAll('created_at')).toEqual(['gte:' + start, 'lte:' + end]);
    });

    it('handles null or empty madeBy', () => {
        const filters = { ...baseFilters, madeBy: '' };
        const params = createHistoryParams(1, filters);

        expect(params.get('email')).toBeNull();
        expect(params.get('actor')).toBeNull();
    });
});
