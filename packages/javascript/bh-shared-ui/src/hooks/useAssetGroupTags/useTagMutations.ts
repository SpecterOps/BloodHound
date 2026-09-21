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
    CreateAssetGroupTagRequest,
    CreateSelectorRequest,
    RequestOptions,
    UpdateAssetGroupTagRequest,
    UpdateSelectorRequest,
} from 'js-client-library';
import { useMutation, useQueryClient } from 'react-query';
import { apiClient } from '../../utils/api';
import { privilegeZonesKeys } from './useAssetGroupTags';

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
