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

import { useMemo } from 'react';
import { useInfiniteQuery } from 'react-query';
import { useTagsQuery } from '../../..';
import { apiClient } from '../../../utils';
import { HistoryItem, type AssetGroupTagHistoryFilters } from './types';
import { PAGE_SIZE, createHistoryParams } from './utils';

export const useAssetGroupTagHistoryQuery = (filters: AssetGroupTagHistoryFilters, query?: string) => {
    const doSearch = query && query.length >= 3;
    const queryKey = doSearch ? query : 'static';

    const { data: tags = [] } = useTagsQuery();

    const tagMap = useMemo(() => {
        const map: Record<number, string> = {};

        tags.forEach((tag) => {
            map[tag.id] = tag.name;
        });

        return map;
    }, [tags]);

    return useInfiniteQuery({
        queryKey: ['asset-group-tag-history', queryKey, filters],
        queryFn: async ({ pageParam = 1 }) => {
            const params = createHistoryParams(pageParam, filters);

            const result = await (doSearch
                ? apiClient.searchAssetGroupTagHistory(query, { params })
                : apiClient.getAssetGroupTagHistory({ params }));

            const historyItems: HistoryItem[] = result.data.data.records.map((item) => ({
                ...item,
                tagName: tagMap[item.asset_group_tag_id],
            }));

            return { ...result.data, data: { records: historyItems } };
        },
        getNextPageParam: (lastPage) => {
            const nextPage = lastPage.skip / PAGE_SIZE + 2;

            if ((nextPage - 1) * PAGE_SIZE >= lastPage.count) {
                return undefined;
            }

            return nextPage;
        },
        getPreviousPageParam: (firstPage) => {
            if (firstPage.skip === 0) {
                return undefined;
            }

            return firstPage.skip / PAGE_SIZE - 1;
        },
    });
};
