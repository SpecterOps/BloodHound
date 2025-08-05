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
    AssetGroupTag,
    AssetGroupTagMemberListItem,
    AssetGroupTagSelector,
    AssetGroupTagTypeLabel,
    AssetGroupTagTypeOwned,
    AssetGroupTagTypes,
    AssetGroupTagTypeTier,
    CreateAssetGroupTagRequest,
    CreateSelectorRequest,
    RequestOptions,
    UpdateAssetGroupTagRequest,
    UpdateSelectorRequest,
} from 'js-client-library';
import { useInfiniteQuery, useMutation, useQuery, useQueryClient } from 'react-query';
import { SortOrder } from '../../types';
import { apiClient } from '../../utils';
import { createPaginatedFetcher, PageParam } from '../../utils/paginatedFetcher';
import { useFeatureFlag } from '../useFeatureFlags';

interface CreateAssetGroupTagParams {
    values: CreateAssetGroupTagRequest;
}

export interface UpdateAssetGroupTagParams {
    tagId: number | string;
    updatedValues: UpdateAssetGroupTagRequest;
}

export interface CreateSelectorParams {
    tagId: string | number;
    values: CreateSelectorRequest;
}
export interface DeleteSelectorParams {
    tagId: string | number;
    selectorId: string | number;
}

export interface PatchSelectorParams extends DeleteSelectorParams {
    updatedValues: UpdateSelectorRequest;
}

const PAGE_SIZE = 25;

export const zoneManagementKeys = {
    all: ['zone-management'] as const,
    tags: () => [...zoneManagementKeys.all, 'tags'] as const,
    tagDetail: (tagId: string | number) => [...zoneManagementKeys.tags(), 'tagId', tagId] as const,
    selectors: () => [...zoneManagementKeys.all, 'selectors'] as const,
    selectorsByTag: (tagId: string | number) => [...zoneManagementKeys.selectors(), 'tag', tagId] as const,
    selectorDetail: (tagId: string | number, selectorId: string | number) =>
        [...zoneManagementKeys.selectorsByTag(tagId), 'selectorId', selectorId] as const,
    members: () => [...zoneManagementKeys.all, 'members'] as const,
    membersByTag: (tagId: string | number) => [...zoneManagementKeys.members(), 'tag', tagId] as const,
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

export const useSelectorsInfiniteQuery = (tagId: string | number | undefined) =>
    useInfiniteQuery<{
        items: AssetGroupTagSelector[];
        nextPageParam?: PageParam;
    }>({
        queryKey: zoneManagementKeys.selectorsByTag(tagId!),
        queryFn: ({ pageParam = { skip: 0, limit: PAGE_SIZE } }) =>
            getAssetGroupTagSelectors(tagId!, pageParam.skip, pageParam.limit),
        getNextPageParam: (lastPage) => lastPage.nextPageParam,
        enabled: tagId !== undefined,
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

export const useTagMembersInfiniteQuery = (tagId: number | string | undefined, sortOrder: SortOrder) =>
    useInfiniteQuery<{
        items: AssetGroupTagMemberListItem[];
        nextPageParam?: PageParam;
    }>({
        queryKey: zoneManagementKeys.membersByTag(tagId!),
        queryFn: ({ pageParam = { skip: 0, limit: PAGE_SIZE } }) =>
            getAssetGroupTagMembers(tagId!, pageParam.skip, pageParam.limit, sortOrder),
        getNextPageParam: (lastPage) => lastPage.nextPageParam,
        enabled: tagId !== undefined,
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
    tagId: number | string | undefined,
    selectorId: number | string | undefined,
    sortOrder: SortOrder
) =>
    useInfiniteQuery<{
        items: AssetGroupTagMemberListItem[];
        nextPageParam?: PageParam;
    }>({
        queryKey: zoneManagementKeys.membersByTagAndSelector(tagId!, selectorId),
        queryFn: ({ pageParam = { skip: 0, limit: PAGE_SIZE } }) =>
            getAssetGroupSelectorMembers(tagId!, selectorId, pageParam.skip, pageParam.limit, sortOrder),
        getNextPageParam: (lastPage) => lastPage.nextPageParam,
        enabled: tagId !== undefined && selectorId !== undefined,
    });

export const createSelector = async (params: CreateSelectorParams, options?: RequestOptions) => {
    const { tagId, values } = params;

    const res = await apiClient.createAssetGroupTagSelector(tagId, values, options);

    return res.data.data;
};

export const useCreateSelector = (tagId: string | number | undefined) => {
    const queryClient = useQueryClient();
    return useMutation(createSelector, {
        onSettled: async () => {
            await queryClient.invalidateQueries(zoneManagementKeys.selectorsByTag(tagId!));
        },
    });
};

export const patchSelector = async (params: PatchSelectorParams, options?: RequestOptions) => {
    const { tagId, selectorId, updatedValues } = params;

    const res = await apiClient.updateAssetGroupTagSelector(tagId, selectorId, updatedValues, options);

    return res.data.data;
};

export const usePatchSelector = (tagId: string | number | undefined) => {
    const queryClient = useQueryClient();
    return useMutation(patchSelector, {
        onSettled: async () => {
            await queryClient.invalidateQueries(zoneManagementKeys.selectorsByTag(tagId!));
        },
    });
};

export const deleteSelector = async (ids: DeleteSelectorParams, options?: RequestOptions) =>
    await apiClient.deleteAssetGroupTagSelector(ids.tagId, ids.selectorId, options).then((res) => res.data.data);

export const useDeleteSelector = () => {
    const queryClient = useQueryClient();
    return useMutation(deleteSelector, {
        onSettled: async (_data, _error, variables) => {
            await queryClient.invalidateQueries(zoneManagementKeys.selectorsByTag(variables.tagId));
        },
    });
};

export const useSelectorInfo = (tagId: string, selectorId: string) =>
    useQuery({
        queryKey: zoneManagementKeys.selectorDetail(tagId, selectorId),
        queryFn: async ({ signal }) => {
            const response = await apiClient.getAssetGroupTagSelector(tagId, selectorId, { signal });
            return response.data.data['selector'];
        },
        enabled: tagId !== '' && selectorId !== '',
    });

export const createAssetGroupTag = async (params: CreateAssetGroupTagParams, options?: RequestOptions) => {
    const { values } = params;

    const res = await apiClient.createAssetGroupTag(values, options);

    return res.data.data;
};

export const useCreateAssetGroupTag = () => {
    const queryClient = useQueryClient();
    return useMutation(createAssetGroupTag, {
        onSettled: async () => {
            await queryClient.invalidateQueries(zoneManagementKeys.tags());
        },
    });
};

export const patchAssetGroupTag = async (params: UpdateAssetGroupTagParams, options?: RequestOptions) => {
    const { tagId, updatedValues } = params;

    const res = await apiClient.updateAssetGroupTag(tagId, updatedValues, options);

    return res.data.data;
};

export const usePatchAssetGroupTag = (tagId: string | number) => {
    const queryClient = useQueryClient();
    return useMutation(patchAssetGroupTag, {
        onSettled: async () => {
            await queryClient.invalidateQueries(zoneManagementKeys.tags());
            await queryClient.invalidateQueries(zoneManagementKeys.tagDetail(tagId));
        },
    });
};

export const deleteAssetGroupTag = async (tagId: string | number, options?: RequestOptions) =>
    await apiClient.deleteAssetGroupTag(tagId, options).then((res) => res.data.data);

export const useDeleteAssetGroupTag = () => {
    const queryClient = useQueryClient();
    return useMutation(deleteAssetGroupTag, {
        onSettled: async (_data, _error, tagId) => {
            await queryClient.invalidateQueries(zoneManagementKeys.tags());
            await queryClient.invalidateQueries(zoneManagementKeys.tagDetail(tagId));
        },
    });
};

export const useAssetGroupTagInfo = (tagId: string) =>
    useQuery({
        queryKey: zoneManagementKeys.tagDetail(tagId),
        queryFn: async ({ signal }) => {
            const response = await apiClient.getAssetGroupTag(tagId, { signal });
            return response.data.data.tag;
        },
        enabled: tagId !== '',
    });

export const useAssetGroupTags = () => {
    const { data, isLoading, isError } = useFeatureFlag('tier_management_engine');

    const queryEnabled = !isLoading && !isError && data?.enabled;

    return useQuery({
        queryKey: zoneManagementKeys.tags(),
        queryFn: async ({ signal }) => {
            const response = await apiClient.getAssetGroupTags({ signal });
            return response.data.data.tags;
        },
        enabled: queryEnabled,
    });
};

export const useOrderedTags = () => {
    const { isLoading, isError, data } = useAssetGroupTags();

    const orderedTags = (data ?? [])
        ?.filter((tag) => tag.type === AssetGroupTagTypeTier)
        .sort((a, b) => {
            const aPos = a.position ?? 0;
            const bPos = b.position ?? 0;
            return aPos - bPos;
        });

    return { orderedTags, isLoading, isError };
};

const HighestPrivilegePosition = 1 as const;

export const useHighestPrivilegeTag = () => {
    const { orderedTags, isLoading, isError } = useOrderedTags();
    const tag = orderedTags?.find((tag) => tag.position === HighestPrivilegePosition);

    return { isLoading, isError, tag };
};

export const useHighestPrivilegeTagId = () => {
    const { orderedTags, isLoading, isError } = useOrderedTags();
    const tagId = orderedTags?.find((tag) => tag.position === HighestPrivilegePosition)?.id;

    return { isLoading, isError, tagId };
};

export const useLabels = () => {
    const tagsQuery = useAssetGroupTags();
    const labelTypes: AssetGroupTagTypes[] = [AssetGroupTagTypeLabel, AssetGroupTagTypeOwned];

    if (tagsQuery.isLoading || tagsQuery.isError) return [];

    return tagsQuery.data?.filter((tag) => labelTypes.includes(tag.type));
};

export const useOwnedTagId = () => {
    const tagsQuery = useAssetGroupTags();
    return tagsQuery.data?.find((tag) => tag.type === AssetGroupTagTypeOwned)?.id;
};
