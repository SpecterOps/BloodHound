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

import { RequestOptions } from 'js-client-library/dist/requests';
import { useMutation, useQuery, useQueryClient } from 'react-query';
import { apiClient } from '../../../../utils';
import { CreateSelectorParams, DeleteSelectorParams, PatchSelectorParams } from './types';

const createSelector = async (params: CreateSelectorParams, options?: RequestOptions) => {
    const { tagId, values } = params;

    const res = await apiClient.createAssetGroupTagSelector(tagId, values, options);

    return res.data.data;
};

export const useCreateSelector = (tagId: string | number | undefined) => {
    const queryClient = useQueryClient();
    return useMutation(createSelector, {
        onSettled: () => {
            queryClient.invalidateQueries(['tier-management', 'tags', tagId, 'selectors']);
        },
    });
};

const patchSelector = async (params: PatchSelectorParams, options?: RequestOptions) => {
    const { tagId, selectorId, updatedValues } = params;

    const res = await apiClient.updateAssetGroupTagSelector(tagId, selectorId, updatedValues, options);

    return res.data.data;
};

export const usePatchSelector = (tagId: string | number | undefined) => {
    const queryClient = useQueryClient();
    return useMutation(patchSelector, {
        onSettled: () => {
            queryClient.invalidateQueries(['tier-management', 'tags', tagId, 'selectors']);
        },
    });
};

const deleteSelector = async (ids: DeleteSelectorParams, options?: RequestOptions) =>
    await apiClient.deleteAssetGroupTagSelector(ids.tagId, ids.selectorId, options).then((res) => res.data.data);

export const useDeleteSelector = () => {
    const queryClient = useQueryClient();
    return useMutation(deleteSelector, {
        onSettled: (_data, _error, variables) => {
            queryClient.invalidateQueries(['tier-management', 'tags', variables.tagId, 'selectors']);
        },
    });
};

export const useSelectorInfo = (tagId: string, selectorId: string) =>
    useQuery({
        queryKey: ['tier-management', 'tags', tagId, 'selectors', selectorId],
        queryFn: async ({ signal }) => {
            const response = await apiClient.getAssetGroupTagSelector(tagId, selectorId, { signal });
            return response.data.data['selector'];
        },
        enabled: tagId !== '' && selectorId !== '',
    });
