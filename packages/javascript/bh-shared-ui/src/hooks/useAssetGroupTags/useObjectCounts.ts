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

import { useQuery } from 'react-query';
import { ENVIRONMENT_AGGREGATION_SUPPORTED_ROUTES } from '../../routes';
import { apiClient } from '../../utils/api';
import { useEnvironmentIdList } from '../useEnvironmentIdList';
import { usePZPathParams } from '../usePZParams/usePZPathParams';

export const useTagObjectCounts = (tagId: string | undefined, ruleId: string | undefined, environments: string[]) =>
    useQuery({
        queryKey: ['asset-group-tags-count', tagId, ...(environments ?? [])],
        queryFn: async ({ signal }) => {
            if (!tagId) return Promise.reject('No Tag ID available for tag counts request');

            return apiClient.getAssetGroupTagMembersCount(tagId, environments, { signal }).then((res) => res.data.data);
        },
        enabled: !!tagId && !ruleId,
    });

const useRuleObjectCounts = (tagId: string | undefined, ruleId: string | undefined, environments: string[]) =>
    useQuery({
        queryKey: ['asset-group-tags-count', tagId, 'rule', ruleId, ...environments],
        queryFn: async ({ signal }) => {
            if (!tagId) return Promise.reject('No Tag ID available for Rule counts request');
            if (!ruleId) return Promise.reject('No Rule ID available for Rule counts request');

            return apiClient
                .getAssetGroupTagRuleMembersCount(tagId, ruleId, environments, { signal })
                .then((res) => res.data.data);
        },
        enabled: !!tagId && !!ruleId,
    });

export const useObjectCounts = () => {
    const { ruleId, tagId } = usePZPathParams();

    const environments = useEnvironmentIdList(ENVIRONMENT_AGGREGATION_SUPPORTED_ROUTES, false);
    const tagCounts = useTagObjectCounts(tagId, ruleId, environments);
    const ruleCounts = useRuleObjectCounts(tagId, ruleId, environments);

    if (ruleId) return ruleCounts;
    return tagCounts;
};
