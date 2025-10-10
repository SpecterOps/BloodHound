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
import { LuxonFormat } from '../../../utils';
import { type AssetGroupTagHistoryFilters } from './FilterDialog';

export const PAGE_SIZE = 25;

export const createHistoryParams = (pageParam: number, filters: AssetGroupTagHistoryFilters) => {
    const skip = (pageParam - 1) * PAGE_SIZE;

    const { tagId, madeBy, action } = filters;

    const start = DateTime.fromFormat(filters['start-date'], LuxonFormat.ISO_8601).startOf('day').toISO();
    const end = DateTime.fromFormat(filters['end-date'], LuxonFormat.ISO_8601).endOf('day').toISO();

    const params = new URLSearchParams();

    params.append('limit', PAGE_SIZE.toString());
    params.append('skip', skip.toString());

    if (action) params.append('action', 'eq:' + action);
    if (tagId) params.append('asset_group_tag_id', 'eq:' + tagId.toString());

    if (madeBy) {
        if (madeBy.includes('@')) params.append('email', 'eq:' + madeBy);
        else params.append('actor', 'eq:' + madeBy);
    }

    if (start !== null && end !== null) {
        params.append('created_at', 'gte:' + start);
        params.append('created_at', 'lte:' + end);
    }

    return params;
};

export const measureElement: ((element: Element) => number) | undefined =
    typeof window !== 'undefined' && navigator.userAgent.indexOf('Firefox') === -1
        ? (element) => element?.getBoundingClientRect().height
        : undefined;
