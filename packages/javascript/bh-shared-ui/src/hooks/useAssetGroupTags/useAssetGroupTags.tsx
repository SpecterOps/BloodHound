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
    AssetGroupTagTypes,
    AssetGroupTagTypeTier,
} from 'js-client-library';
import { useQuery } from 'react-query';
import { apiClient } from '../../utils';
import { useFeatureFlag } from '../useFeatureFlags';

export const useAssetGroupTags = () => {
    const { data, isLoading, isError } = useFeatureFlag('tier_management_engine');

    const queryEnabled = !isLoading && !isError && data?.enabled;

    return useQuery({
        queryKey: ['asset-group-tags'],
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
