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

export const useAssetGroupTags = () =>
    useQuery({
        queryKey: ['asset-group-tags'],
        queryFn: async ({ signal }) => {
            const response = await apiClient.getAssetGroupTags({ signal });
            return response.data.data.tags;
        },
    });

export const useOrderedTags = () => {
    const tagsQuery = useAssetGroupTags();

    if (tagsQuery.isLoading || tagsQuery.isError) return [];

    return (tagsQuery.data ?? [])
        ?.filter((tag) => tag.type === AssetGroupTagTypeTier)
        .sort((a, b) => {
            const aPos = a.position ?? 0;
            const bPos = b.position ?? 0;
            return aPos - bPos;
        });
};

const HighestPrivilegePosition = 1 as const;

export const useHighestPrivilegeTag = () => {
    const orderedTags = useOrderedTags();

    return orderedTags?.find((tag) => tag.position === HighestPrivilegePosition);
};

export const useLabels = () => {
    const tagsQuery = useAssetGroupTags();
    const labelTypes: AssetGroupTagTypes[] = [AssetGroupTagTypeLabel, AssetGroupTagTypeOwned];

    if (tagsQuery.isLoading || tagsQuery.isError) return [];

    return tagsQuery.data?.filter((tag) => labelTypes.includes(tag.type));
};
