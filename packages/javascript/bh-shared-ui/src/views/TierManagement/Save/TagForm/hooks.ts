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
    RequestOptions,
    UpdateAssetGroupTagRequest,
} from 'js-client-library/dist/requests';
import { useMutation, useQuery, useQueryClient } from 'react-query';
import { apiClient } from '../../../../utils';

interface CreateAssetGroupTagParams {
    values: CreateAssetGroupTagRequest;
}

interface UpdateAssetGroupTagParams {
    tagId: number | string;
    updatedValues: UpdateAssetGroupTagRequest;
}

const createAssetGroupTag = async (params: CreateAssetGroupTagParams, options?: RequestOptions) => {
    const { values } = params;

    const res = await apiClient.createAssetGroupTag(values, options);

    return res.data.data;
};

export const useCreateAssetGroupTag = () => {
    const queryClient = useQueryClient();
    return useMutation(createAssetGroupTag, {
        onSettled: () => {
            queryClient.invalidateQueries(['tier-management', 'tags']);
        },
    });
};

const patchAssetGroupTag = async (params: UpdateAssetGroupTagParams, options?: RequestOptions) => {
    const { tagId, updatedValues } = params;
    console.log('here');

    const res = await apiClient.updateAssetGroupTag(tagId, updatedValues, options);

    return res.data.data;
};

export const usePatchAssetGroupTag = (tagId: string | number) => {
    const queryClient = useQueryClient();
    return useMutation(patchAssetGroupTag, {
        onSettled: () => {
            queryClient.invalidateQueries(['tier-management', 'tags', tagId]);
        },
    });
};

const deleteAssetGroupTag = async (tagId: number | string, options?: RequestOptions) =>
    await apiClient.deleteAssetGroupTag(tagId, options).then((res) => res.data.data);

export const useDeleteAssetGroupTag = () => {
    const queryClient = useQueryClient();
    return useMutation(deleteAssetGroupTag, {
        onSettled: (_data, _error, tagId) => {
            queryClient.invalidateQueries(['tier-management', 'tags', tagId]);
        },
    });
};

export const useAssetGroupTagInfo = (tagId: string) =>
    useQuery({
        queryKey: ['tier-management', 'tags', tagId],
        queryFn: async ({ signal }) => {
            const response = await apiClient.getAssetGroupTag(tagId, { signal });
            return response.data.data.tag;
        },
        enabled: tagId !== '',
    });
