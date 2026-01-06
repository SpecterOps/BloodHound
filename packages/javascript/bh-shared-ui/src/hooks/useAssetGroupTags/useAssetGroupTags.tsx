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

import {
    AssetGroupTagTypeLabel,
    AssetGroupTagTypeOwned,
    AssetGroupTagTypeZone,
    HighestPrivilegePosition,
    ObjectKey,
    ObjectsKey,
    RuleKey,
    RulesKey,
    type AssetGroupTag,
    type AssetGroupTagMemberListItem,
    type AssetGroupTagSelector,
    type AssetGroupTagType,
    type RequestOptions,
} from 'js-client-library';
import { useInfiniteQuery, useQuery } from 'react-query';
import { SortOrderAscending, type SortOrder } from '../../types';
import { apiClient, type GenericQueryOptions } from '../../utils';
import { createPaginatedFetcher, type PageParam } from '../../utils/paginatedFetcher';
import { useFeatureFlag } from '../useFeatureFlags';
import { isNode, type ItemResponse } from '../useGraphItem';

export const privilegeZonesKeys = {
    all: ['privilege-zones'] as const,
    tags: () => [...privilegeZonesKeys.all, 'tags'] as const,
    tagDetail: (tagId: string | number) => [...privilegeZonesKeys.tags(), 'tagId', tagId] as const,
    rules: () => [...privilegeZonesKeys.all, 'rules'] as const,
    rulesByTag: (
        tagId: string | number,
        sortOrder: SortOrder = undefined,
        environments: string[] = [],
        disabled?: boolean,
        isDefault?: boolean
    ) => [...privilegeZonesKeys.rules(), 'tag', tagId, sortOrder, ...environments, disabled, isDefault] as const,
    ruleDetail: (tagId: string | number, ruleId: string | number) =>
        [...privilegeZonesKeys.rulesByTag(tagId), 'ruleId', ruleId] as const,
    members: () => [...privilegeZonesKeys.all, 'members'] as const,
    membersByTag: (
        tagId: string | number,
        sortOrder: SortOrder,
        environments: string[] = [],
        primary_kind: string = 'all'
    ) => [...privilegeZonesKeys.members(), 'tag', tagId, primary_kind, sortOrder, ...environments] as const,
    membersByTagAndRule: (
        tagId: string | number,
        ruleId: string | number | undefined,
        sortOrder: SortOrder,
        environments: string[] = [],
        primary_kind: string = 'all'
    ) => ['tag', tagId, 'rule', ruleId, primary_kind, sortOrder, ...environments] as const,
    memberDetail: (tagId: string | number, memberId: string | number) =>
        [...privilegeZonesKeys.tagDetail(tagId), 'memberId', memberId] as const,

    certifications: (filters: any, search?: string, environments: string[] = []) =>
        [...privilegeZonesKeys.all, 'certifications', filters, search, ...environments] as const,
};

export const getIsOwnedTag = (tags: AssetGroupTag[]) => tags.find((tag) => tag.type === AssetGroupTagTypeOwned);

export const getIsTierZeroTag = (tags: AssetGroupTag[]) =>
    tags.find((tag) => tag.position === HighestPrivilegePosition);

export const isOwnedObject = (item: ItemResponse): boolean => {
    if (!isNode(item)) return false;

    return item.isOwnedObject;
};

export const isTierZero = (item: ItemResponse): boolean => {
    if (!isNode(item)) return false;

    return item.isTierZero;
};

const getAssetGroupTags = (options: RequestOptions) =>
    apiClient
        .getAssetGroupTags({
            ...options,
            params: {
                counts: true,
            },
        })
        .then((res) => res.data.data.tags);

type useTagQueryOptions = GenericQueryOptions<AssetGroupTag[]>;

export type TagSelect = useTagQueryOptions['select'];

export const useTagsQuery = (queryOptions?: useTagQueryOptions) => {
    const { data, isLoading, isError } = useFeatureFlag('tier_management_engine');

    const enabled = !isLoading && !isError && data?.enabled;

    return useQuery({
        queryKey: privilegeZonesKeys.tags() as unknown as string[],
        queryFn: ({ signal }) => getAssetGroupTags({ signal }),
        enabled,
        ...queryOptions,
    });
};

const PAGE_SIZE = 25;

const createGetRulesParams = (queryParams: GetRulesQueryParams) => {
    const params = new URLSearchParams();
    params.append('skip', queryParams.skip.toString());
    params.append('limit', queryParams.limit.toString());
    params.append('sort_by', queryParams.sortBy);
    params.append('counts', `${queryParams.counts}`);

    if (queryParams.isDefault !== undefined) params.append('is_default', `eq:${queryParams.isDefault}`);
    if (queryParams.disabled !== undefined) params.append('disabled_by', queryParams.disabled ? 'neq:null' : 'eq:null');

    queryParams.environments.forEach((environment) => {
        params.append('environments', environment);
    });

    return params;
};

export const getAssetGroupTagRules = (tagId: string | number, queryParams: GetRulesQueryParams) => {
    const params = createGetRulesParams(queryParams);

    return createPaginatedFetcher(
        () => apiClient.getAssetGroupTagSelectors(tagId, { params }),
        RulesKey,
        queryParams.skip,
        queryParams.limit
    );
};

interface GetRulesQueryParams {
    skip: number;
    limit: number;
    sortBy: string;
    environments: string[];
    counts: boolean;
    isDefault?: boolean;
    disabled?: boolean;
}

interface GetRulesParams {
    sortOrder: SortOrder;
    disabled?: boolean;
    counts?: boolean;
    environments?: string[];
    isDefault?: boolean;
}

export const useRulesInfiniteQuery = (tagId: string | number | undefined, params: GetRulesParams, enabled?: boolean) =>
    useInfiniteQuery<{
        items: AssetGroupTagSelector[];
        nextPageParam?: PageParam;
    }>({
        queryKey: privilegeZonesKeys.rulesByTag(
            tagId!,
            params.sortOrder,
            params.environments,
            params.disabled,
            params.isDefault
        ),

        queryFn: ({ pageParam = { skip: 0, limit: PAGE_SIZE } }) => {
            if (!tagId) return Promise.reject('No tag ID provided for rules request');

            return getAssetGroupTagRules(tagId, {
                skip: pageParam.skip,
                limit: pageParam.limit,
                sortBy: params.sortOrder === SortOrderAscending ? 'name' : '-name',
                environments: params.environments || [],
                isDefault: params.isDefault,
                disabled: params.disabled,
                counts: params.counts ? params.counts : true,
            });
        },

        getNextPageParam: (lastPage) => lastPage.nextPageParam,

        enabled: tagId !== undefined && enabled !== undefined ? enabled : true,
    });

export const getAssetGroupTagMembers = (
    tagId: number | string,
    skip = 0,
    limit = PAGE_SIZE,
    sortOrder: SortOrder = SortOrderAscending,
    environments?: string[],
    primary_kind?: string
) =>
    createPaginatedFetcher<AssetGroupTagMemberListItem>(
        () =>
            apiClient.getAssetGroupTagMembers(
                tagId,
                skip,
                limit,
                sortOrder === SortOrderAscending ? 'name' : '-name',
                environments,
                primary_kind
            ),
        ObjectsKey,
        skip,
        limit
    );

export const useTagMembersInfiniteQuery = (
    tagId: number | string | undefined,
    sortOrder: SortOrder,
    environments?: string[],
    primary_kind?: string,
    enabled?: boolean
) =>
    useInfiniteQuery<{
        items: AssetGroupTagMemberListItem[];
        nextPageParam?: PageParam;
    }>({
        queryKey: privilegeZonesKeys.membersByTag(tagId!, sortOrder, environments, primary_kind),

        queryFn: ({ pageParam = { skip: 0, limit: PAGE_SIZE } }) => {
            if (!tagId) return Promise.reject('No tag ID provided for tag members request');

            return getAssetGroupTagMembers(
                tagId,
                pageParam.skip,
                pageParam.limit,
                sortOrder,
                environments,
                primary_kind
            );
        },

        getNextPageParam: (lastPage) => lastPage.nextPageParam,

        enabled: tagId !== undefined && (enabled === undefined ? true : enabled),
    });

export const getAssetGroupTagRuleMembers = (
    tagId: number | string,
    ruleId: number | string,
    skip: number = 0,
    limit: number = PAGE_SIZE,
    sortOrder: SortOrder = SortOrderAscending,
    environments?: string[],
    primary_kind?: string
) =>
    createPaginatedFetcher(
        () =>
            apiClient.getAssetGroupTagSelectorMembers(
                tagId,
                ruleId,
                skip,
                limit,
                sortOrder === SortOrderAscending ? 'name' : '-name',
                environments,
                primary_kind
            ),
        ObjectsKey,
        skip,
        limit
    );

export const useRuleMembersInfiniteQuery = (
    tagId: number | string | undefined,
    ruleId: number | string | undefined,
    sortOrder: SortOrder,
    environments?: string[],
    primary_kind?: string,
    enabled?: boolean
) =>
    useInfiniteQuery<{
        items: AssetGroupTagMemberListItem[];
        nextPageParam?: PageParam;
    }>({
        queryKey: privilegeZonesKeys.membersByTagAndRule(tagId!, ruleId, sortOrder, environments, primary_kind),
        queryFn: ({ pageParam = { skip: 0, limit: PAGE_SIZE } }) => {
            if (!tagId) return Promise.reject('No tag ID available to get rule members');
            if (!ruleId) return Promise.reject('No rule ID available to get rule members');

            return getAssetGroupTagRuleMembers(
                tagId,
                ruleId,
                pageParam.skip,
                pageParam.limit,
                sortOrder,
                environments,
                primary_kind
            );
        },

        getNextPageParam: (lastPage) => lastPage.nextPageParam,

        enabled: tagId !== undefined && ruleId !== undefined && (enabled === undefined ? true : enabled),
    });

export const useMemberInfo = (tagId: string = '', memberId: string = '') =>
    useQuery({
        queryKey: privilegeZonesKeys.memberDetail(tagId, memberId),
        queryFn: async ({ signal }) => {
            return apiClient.getAssetGroupTagMemberInfo(tagId, memberId, { signal }).then((res) => {
                return res.data.data[ObjectKey];
            });
        },
        enabled: tagId !== '' && memberId !== '',
    });

export const useRuleInfo = (tagId: string = '', ruleId: string = '') =>
    useQuery({
        queryKey: privilegeZonesKeys.ruleDetail(tagId, ruleId),
        queryFn: async ({ signal }) => {
            const response = await apiClient.getAssetGroupTagSelector(tagId, ruleId, { signal });
            return response.data.data[RuleKey];
        },
        enabled: tagId !== '' && ruleId !== '',
    });

export const useAssetGroupTagInfo = (tagId: string) =>
    useQuery({
        queryKey: privilegeZonesKeys.tagDetail(tagId),
        queryFn: async ({ signal }) => {
            const response = await apiClient.getAssetGroupTag(tagId, { signal });
            return response.data.data.tag;
        },
        enabled: tagId !== '',
    });

export const useOrderedTags = () => {
    const select = (tags: AssetGroupTag[]) =>
        tags
            .filter((tag) => tag.type === AssetGroupTagTypeZone)
            .sort((a, b) => {
                const aPos = a.position ?? 0;
                const bPos = b.position ?? 0;
                return aPos - bPos;
            });

    return useTagsQuery({ select });
};

export const useHighestPrivilegeTag = () => {
    const { data: orderedTags, isLoading, isError } = useOrderedTags();
    const tag = orderedTags?.find((tag) => tag.position === HighestPrivilegePosition);

    return { isLoading, isError, tag };
};

export const useHighestPrivilegeTagId = () => {
    const { data: orderedTags, isLoading, isError } = useOrderedTags();
    const tagId = orderedTags?.find((tag) => tag.position === HighestPrivilegePosition)?.id;

    return { isLoading, isError, tagId };
};

export const useLabels = () => {
    const labelTypes: AssetGroupTagType[] = [AssetGroupTagTypeLabel, AssetGroupTagTypeOwned];
    const select = (tags: AssetGroupTag[]) => tags.filter((tag) => labelTypes.includes(tag.type));

    return useTagsQuery({ select });
};

export const useOwnedTagId = () => {
    const tagsQuery = useTagsQuery();
    return tagsQuery.data?.find((tag) => tag.type === AssetGroupTagTypeOwned)?.id;
};
