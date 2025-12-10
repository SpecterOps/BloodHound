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
import { findIconDefinition } from '@fortawesome/fontawesome-svg-core';
import { IconName } from '@fortawesome/free-solid-svg-icons';
import {
    AssetGroupTag,
    AssetGroupTagMemberListItem,
    AssetGroupTagSelector,
    AssetGroupTagType,
    AssetGroupTagTypeLabel,
    AssetGroupTagTypeOwned,
    AssetGroupTagTypeZone,
    CreateAssetGroupTagRequest,
    CreateSelectorRequest,
    RequestOptions,
    UpdateAssetGroupTagRequest,
    UpdateSelectorRequest,
} from 'js-client-library';
import { useEffect, useState } from 'react';
import { useInfiniteQuery, useMutation, useQuery, useQueryClient } from 'react-query';
import { SortOrder } from '../../types';
import {
    DEFAULT_GLYPH_BACKGROUND_COLOR,
    DEFAULT_GLYPH_COLOR,
    GLYPH_SCALE,
    GenericQueryOptions,
    apiClient,
    getModifiedSvgUrlFromIcon,
} from '../../utils';
import { PageParam, createPaginatedFetcher } from '../../utils/paginatedFetcher';
import { useFeatureFlag } from '../useFeatureFlags';

interface CreateAssetGroupTagParams {
    values: CreateAssetGroupTagRequest;
}

export interface UpdateAssetGroupTagParams {
    tagId: number | string;
    updatedValues: UpdateAssetGroupTagRequest;
}

export interface CreateRuleParams {
    tagId: string | number;
    values: CreateSelectorRequest;
}
export interface DeleteRuleParams {
    tagId: string | number;
    ruleId: string | number;
}

export interface PatchRuleParams extends DeleteRuleParams {
    updatedValues: UpdateSelectorRequest;
}

const PAGE_SIZE = 25;

export const privilegeZonesKeys = {
    all: ['privilege-zones'] as const,
    tags: () => [...privilegeZonesKeys.all, 'tags'] as const,
    tagDetail: (tagId: string | number) => [...privilegeZonesKeys.tags(), 'tagId', tagId] as const,
    rules: () => [...privilegeZonesKeys.all, 'rules'] as const,
    rulesByTag: (tagId: string | number, sortOrder: SortOrder = undefined, environments: string[] = []) =>
        [...privilegeZonesKeys.rules(), 'tag', tagId, sortOrder, ...environments] as const,
    ruleDetail: (tagId: string | number, ruleId: string | number) =>
        [...privilegeZonesKeys.rulesByTag(tagId), 'ruleId', ruleId] as const,
    members: () => [...privilegeZonesKeys.all, 'members'] as const,
    membersByTag: (tagId: string | number, sortOrder: SortOrder, environments: string[] = []) =>
        [...privilegeZonesKeys.members(), 'tag', tagId, sortOrder, ...environments] as const,
    membersByTagAndRule: (
        tagId: string | number,
        ruleId: string | number | undefined,
        sortOrder: SortOrder,
        environments: string[] = []
    ) => ['tag', tagId, 'rule', ruleId, sortOrder, ...environments] as const,
    memberDetail: (tagId: string | number, memberId: string | number) =>
        [...privilegeZonesKeys.tagDetail(tagId), 'memberId', memberId] as const,

    certifications: (filters: any, search?: string, environments: string[] = []) =>
        [...privilegeZonesKeys.all, 'certifications', filters, search, ...environments] as const,
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

const glyphQualifier = (glyph: string | null) => !glyph?.includes('http');

const glyphTransformer = (glyph: string, darkMode?: boolean): string => {
    const iconDefiniton = findIconDefinition({ prefix: 'fas', iconName: glyph as IconName });

    if (!iconDefiniton) return '';

    const glyphIconUrl = getModifiedSvgUrlFromIcon(iconDefiniton, {
        styles: {
            'transform-origin': 'center',
            scale: GLYPH_SCALE,
            background: darkMode ? DEFAULT_GLYPH_COLOR : DEFAULT_GLYPH_BACKGROUND_COLOR,
            color: darkMode ? DEFAULT_GLYPH_BACKGROUND_COLOR : DEFAULT_GLYPH_COLOR,
        },
    });

    return glyphIconUrl;
};

export interface GlyphUtils {
    qualifier?: (glyph: string | null) => boolean;
    transformer: (glyph: string, darkMode?: boolean) => string;
}

export const glyphUtils: GlyphUtils = {
    qualifier: glyphQualifier,
    transformer: glyphTransformer,
};

export const TagLabelPrefix = 'Tag_' as const;

export const createGlyphMapFromTags = (
    tags: AssetGroupTag[] | undefined,
    utils: GlyphUtils,
    darkMode?: boolean
): Record<string, string> => {
    const glyphMap: Record<string, string> = {};
    const { qualifier = () => true, transformer } = utils;

    tags?.forEach((tag) => {
        const underscoredTagName = tag.name.split(' ').join('_');

        if (tag.glyph === null) return;
        if (!qualifier(tag.glyph)) return;

        const glyphValue = transformer(tag.glyph, darkMode);

        if (tag.type === AssetGroupTagTypeOwned) {
            glyphMap.owned = `${TagLabelPrefix}${underscoredTagName}`;
            glyphMap.ownedGlyph = glyphValue;
            return;
        }

        if (glyphValue !== '') glyphMap[`${TagLabelPrefix}${underscoredTagName}`] = glyphValue;
    });

    return glyphMap;
};

export const getGlyphFromKinds = (kinds: string[] = [], tagGlyphMap: Record<string, string> = {}): string | null => {
    for (let index = kinds.length - 1; index > -1; index--) {
        const kind = kinds[index];

        if (!kind.includes(TagLabelPrefix)) continue;

        if (tagGlyphMap[kind]) return tagGlyphMap[kind];
    }
    return null;
};

export type TagGlyphs = Record<string, string>;

export const useTagGlyphs = (glyphUtils: GlyphUtils, darkMode?: boolean): TagGlyphs => {
    const [glyphMap, setGlyphMap] = useState<Record<string, string>>({});
    const tagsQuery = useAssetGroupTags();

    useEffect(() => {
        if (!tagsQuery.data) return;

        const newMap = createGlyphMapFromTags(tagsQuery.data, glyphUtils, darkMode);
        setGlyphMap(newMap);
    }, [tagsQuery.data, glyphUtils, darkMode]);

    return glyphMap;
};

type useTagQueryOptions = GenericQueryOptions<AssetGroupTag[]>;
export type TagSelect = useTagQueryOptions['select'];
export const useTagsQuery = (queryOptions?: useTagQueryOptions) =>
    useQuery({
        queryKey: privilegeZonesKeys.tags() as unknown as string[],
        queryFn: ({ signal }) => getAssetGroupTags({ signal }),
        ...queryOptions,
    });

export const getAssetGroupTagRules = (
    tagId: string | number,
    skip: number = 0,
    limit: number = PAGE_SIZE,
    sortOrder: SortOrder = 'asc',
    environments: string[] = []
) =>
    createPaginatedFetcher(
        () =>
            apiClient.getAssetGroupTagSelectors(
                tagId,
                skip,
                limit,
                sortOrder === 'asc' ? 'name' : '-name',
                environments,
                { params: { counts: true } }
            ),
        'selectors', // 'selectors' is the key from the API response so should not be updated to 'rules'
        skip,
        limit
    );

export const useRulesInfiniteQuery = (
    tagId: string | number | undefined,
    sortOrder: SortOrder,
    environments: string[] = []
) =>
    useInfiniteQuery<{
        items: AssetGroupTagSelector[];
        nextPageParam?: PageParam;
    }>({
        queryKey: privilegeZonesKeys.rulesByTag(tagId!, sortOrder, environments),
        queryFn: ({ pageParam = { skip: 0, limit: PAGE_SIZE } }) => {
            if (!tagId) return Promise.reject('No tag ID provided for rules request');
            return getAssetGroupTagRules(tagId, pageParam.skip, pageParam.limit, sortOrder, environments);
        },
        getNextPageParam: (lastPage) => lastPage.nextPageParam,
        enabled: tagId !== undefined,
    });

export const getAssetGroupTagMembers = (
    tagId: number | string,
    skip = 0,
    limit = PAGE_SIZE,
    sortOrder: SortOrder = 'asc',
    environments?: string[]
) =>
    createPaginatedFetcher<AssetGroupTagMemberListItem>(
        () =>
            apiClient.getAssetGroupTagMembers(tagId, skip, limit, sortOrder === 'asc' ? 'name' : '-name', environments),
        'members',
        skip,
        limit
    );

export const useTagMembersInfiniteQuery = (
    tagId: number | string | undefined,
    sortOrder: SortOrder,
    environments?: string[]
) =>
    useInfiniteQuery<{
        items: AssetGroupTagMemberListItem[];
        nextPageParam?: PageParam;
    }>({
        queryKey: privilegeZonesKeys.membersByTag(tagId!, sortOrder, environments),
        queryFn: ({ pageParam = { skip: 0, limit: PAGE_SIZE } }) => {
            if (!tagId) return Promise.reject('No tag ID provided for tag members request');
            return getAssetGroupTagMembers(tagId, pageParam.skip, pageParam.limit, sortOrder, environments);
        },
        getNextPageParam: (lastPage) => lastPage.nextPageParam,
        enabled: tagId !== undefined,
    });

export const getAssetGroupTagRuleMembers = (
    tagId: number | string,
    ruleId: number | string,
    skip: number = 0,
    limit: number = PAGE_SIZE,
    sortOrder: SortOrder = 'asc',
    environments?: string[]
) =>
    createPaginatedFetcher(
        () =>
            apiClient.getAssetGroupTagSelectorMembers(
                tagId,
                ruleId,
                skip,
                limit,
                sortOrder === 'asc' ? 'name' : '-name',
                environments
            ),
        'members',
        skip,
        limit
    );

export const useRuleMembersInfiniteQuery = (
    tagId: number | string | undefined,
    ruleId: number | string | undefined,
    sortOrder: SortOrder,
    environments?: string[]
) =>
    useInfiniteQuery<{
        items: AssetGroupTagMemberListItem[];
        nextPageParam?: PageParam;
    }>({
        queryKey: privilegeZonesKeys.membersByTagAndRule(tagId!, ruleId, sortOrder, environments),
        queryFn: ({ pageParam = { skip: 0, limit: PAGE_SIZE } }) => {
            if (!tagId) return Promise.reject('No tag ID available to get rule members');
            if (!ruleId) return Promise.reject('No rule ID available to get rule members');
            return getAssetGroupTagRuleMembers(tagId, ruleId, pageParam.skip, pageParam.limit, sortOrder, environments);
        },
        getNextPageParam: (lastPage) => lastPage.nextPageParam,
        enabled: tagId !== undefined && ruleId !== undefined,
    });

export const useMemberInfo = (tagId: string = '', memberId: string = '') =>
    useQuery({
        queryKey: privilegeZonesKeys.memberDetail(tagId, memberId),
        queryFn: async ({ signal }) => {
            return apiClient.getAssetGroupTagMemberInfo(tagId, memberId, { signal }).then((res) => {
                return res.data.data['member'];
            });
        },
        enabled: tagId !== '' && memberId !== '',
    });

export const createRule = async (params: CreateRuleParams, options?: RequestOptions) => {
    const { tagId, values } = params;

    const res = await apiClient.createAssetGroupTagSelector(tagId, values, options);

    return res.data.data;
};

export const useCreateRule = (tagId: string | number | undefined) => {
    const queryClient = useQueryClient();
    return useMutation(createRule, {
        onSettled: async () => {
            await queryClient.invalidateQueries(privilegeZonesKeys.rulesByTag(tagId!));
        },
    });
};

export const patchRule = async (params: PatchRuleParams, options?: RequestOptions) => {
    const { tagId, ruleId, updatedValues } = params;

    const res = await apiClient.updateAssetGroupTagSelector(tagId, ruleId, updatedValues, options);

    return res.data.data;
};

export const usePatchRule = (tagId: string | number) => {
    const queryClient = useQueryClient();
    return useMutation(patchRule, {
        onSettled: async () => {
            await queryClient.invalidateQueries(privilegeZonesKeys.rulesByTag(tagId));
        },
    });
};

export const deleteRule = async (ids: DeleteRuleParams, options?: RequestOptions) =>
    await apiClient.deleteAssetGroupTagSelector(ids.tagId, ids.ruleId, options).then((res) => res.data.data);

export const useDeleteRule = () => {
    const queryClient = useQueryClient();
    return useMutation(deleteRule, {
        onSettled: async (_data, _error, variables) => {
            queryClient.invalidateQueries(privilegeZonesKeys.rulesByTag(variables.tagId));
            queryClient.invalidateQueries(privilegeZonesKeys.ruleDetail(variables.tagId, variables.ruleId));
        },
    });
};

export const useRuleInfo = (tagId: string = '', ruleId: string = '') =>
    useQuery({
        queryKey: privilegeZonesKeys.ruleDetail(tagId, ruleId),
        queryFn: async ({ signal }) => {
            const response = await apiClient.getAssetGroupTagSelector(tagId, ruleId, { signal });
            return response.data.data['selector']; // 'selector' is the key from the API response so should not be updated to 'rules'
        },
        enabled: tagId !== '' && ruleId !== '',
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
            await queryClient.invalidateQueries(privilegeZonesKeys.tags());
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
            await queryClient.invalidateQueries(privilegeZonesKeys.tags());
            await queryClient.invalidateQueries(privilegeZonesKeys.tagDetail(tagId));
        },
    });
};

export const deleteAssetGroupTag = async (tagId: string | number, options?: RequestOptions) =>
    await apiClient.deleteAssetGroupTag(tagId, options).then((res) => res.data.data);

export const useDeleteAssetGroupTag = () => {
    const queryClient = useQueryClient();
    return useMutation(deleteAssetGroupTag, {
        onSettled: async (_data, _error, tagId) => {
            queryClient.invalidateQueries(privilegeZonesKeys.tags());
            queryClient.invalidateQueries(privilegeZonesKeys.tagDetail(tagId));
        },
    });
};

export const useAssetGroupTagInfo = (tagId: string) =>
    useQuery({
        queryKey: privilegeZonesKeys.tagDetail(tagId),
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
        queryKey: privilegeZonesKeys.tags(),
        queryFn: getAssetGroupTags,
        enabled: queryEnabled,
    });
};

export const useOrderedTags = () => {
    const { isLoading, isError, data } = useAssetGroupTags();

    const orderedTags = (data ?? [])
        ?.filter((tag) => tag.type === AssetGroupTagTypeZone)
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
    const { isLoading, isError, ...tagsQuery } = useAssetGroupTags();
    const labelTypes: AssetGroupTagType[] = [AssetGroupTagTypeLabel, AssetGroupTagTypeOwned];

    if (isLoading || isError) return [];

    return tagsQuery.data?.filter((tag) => labelTypes.includes(tag.type));
};

export const useOwnedTagId = () => {
    const tagsQuery = useAssetGroupTags();
    return tagsQuery.data?.find((tag) => tag.type === AssetGroupTagTypeOwned)?.id;
};
