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

import { AssetGroupTag, AssetGroupTagMemberListItem, AssetGroupTagSelector } from 'js-client-library';
import { useInfiniteQuery, useQuery } from 'react-query';
import { SortOrder } from '../../types';
import { apiClient } from '../../utils';
import { PageParam, createPaginatedFetcher } from '../../utils/paginatedFetcher';

const PAGE_SIZE = 25;

export const zoneManagementKeys = {
    all: ['zone-management'] as const,
    tags: () => [...zoneManagementKeys.all, 'tags'] as const,
    selectorsByTag: (tagId: string | number) => [...zoneManagementKeys.all, 'selectors', tagId] as const,
    membersByTag: (tagId: string | number) => [...zoneManagementKeys.all, 'members', 'tag', tagId] as const,
    membersByTagAndSelector: (tagId: string | number, selectorId: string | number | undefined) =>
        [...zoneManagementKeys.membersByTag(tagId), 'selector', selectorId] as const,
};

export const getAssetGroupTags = () =>
    apiClient
        .getAssetGroupTags({
            params: {
                counts: true,
            },
        })
        .then((res) => {
            return res.data.data['tags'];
        });

export const useTagsQuery = (filter?: (value: AssetGroupTag, index: number, array: AssetGroupTag[]) => boolean) =>
    useQuery({
        queryKey: zoneManagementKeys.tags(),
        queryFn: () => getAssetGroupTags(),
        select: (data) => (filter ? data.filter(filter) : data),
    });

export const getAssetGroupTagSelectors = (tagId: string | number, skip: number = 0, limit: number = PAGE_SIZE) =>
    createPaginatedFetcher(
        () =>
            apiClient.getAssetGroupTagSelectors(tagId, {
                params: {
                    skip,
                    limit,
                    counts: true,
                },
            }),
        'selectors',
        skip,
        limit
    );

export const useSelectorsInfiniteQuery = (tagId: string | number) =>
    useInfiniteQuery<{
        items: AssetGroupTagSelector[];
        nextPageParam?: PageParam;
    }>({
        queryKey: zoneManagementKeys.selectorsByTag(tagId),
        queryFn: ({ pageParam = { skip: 0, limit: PAGE_SIZE } }) =>
            getAssetGroupTagSelectors(tagId, pageParam.skip, pageParam.limit),
        getNextPageParam: (lastPage) => lastPage.nextPageParam,
    });

export const getAssetGroupTagMembers = (
    tagId: number | string,
    skip = 0,
    limit = PAGE_SIZE,
    sortOrder: SortOrder = 'asc'
) =>
    createPaginatedFetcher<AssetGroupTagMemberListItem>(
        () => apiClient.getAssetGroupTagMembers(tagId, skip, limit, sortOrder === 'asc' ? 'name' : '-name'),
        'members',
        skip,
        limit
    );

export const useTagMembersInfiniteQuery = (tagId: number | string, sortOrder: SortOrder) =>
    useInfiniteQuery<{
        items: AssetGroupTagMemberListItem[];
        nextPageParam?: PageParam;
    }>({
        queryKey: zoneManagementKeys.membersByTag(tagId),
        queryFn: ({ pageParam = { skip: 0, limit: PAGE_SIZE } }) =>
            getAssetGroupTagMembers(tagId, pageParam.skip, pageParam.limit, sortOrder),
        getNextPageParam: (lastPage) => lastPage.nextPageParam,
    });

export const getAssetGroupSelectorMembers = (
    tagId: number | string,
    selectorId: number | string | undefined = undefined,
    skip: number = 0,
    limit: number = PAGE_SIZE,
    sortOrder: SortOrder = 'asc'
) =>
    createPaginatedFetcher(
        () =>
            apiClient.getAssetGroupTagSelectorMembers(
                tagId,
                selectorId!,
                skip,
                limit,
                sortOrder === 'asc' ? 'name' : '-name'
            ),
        'members',
        skip,
        limit
    );

export const useSelectorMembersInfiniteQuery = (
    tagId: number | string,
    selectorId: number | string | undefined,
    sortOrder: SortOrder
) =>
    useInfiniteQuery<{
        items: AssetGroupTagMemberListItem[];
        nextPageParam?: PageParam;
    }>({
        queryKey: zoneManagementKeys.membersByTagAndSelector(tagId, selectorId),
        queryFn: ({ pageParam = { skip: 0, limit: PAGE_SIZE } }) =>
            getAssetGroupSelectorMembers(tagId, selectorId, pageParam.skip, pageParam.limit, sortOrder),
        getNextPageParam: (lastPage) => lastPage.nextPageParam,
        enabled: selectorId !== undefined,
    });
