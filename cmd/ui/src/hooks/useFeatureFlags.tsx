// Copyright 2023 Specter Ops, Inc.
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

import { RequestOptions } from 'js-client-library';
import { useMutation, useQuery, useQueryClient } from 'react-query';
import { apiClient } from 'bh-shared-ui';

export type Flag = {
    id: number;
    key: string;
    name: string;
    description: string;
    enabled: boolean;
    user_updatable: boolean;
};

export const featureFlagKeys = {
    all: ['featureFlags'] as const,
};

export const getFeatureFlags = (options?: RequestOptions): Promise<Flag[]> => {
    return apiClient.getFeatureFlags(options).then((response) => response.data.data);
};

export const toggleFeatureFlag = (flagId: string | number, options?: RequestOptions) => {
    return apiClient.toggleFeatureFlag(flagId, options).then((response) => response.data);
};

export const useFeatureFlags = () => useQuery(featureFlagKeys.all, ({ signal }) => getFeatureFlags({ signal }));

export const useFeatureFlag = (flagKey: string) =>
    useQuery(featureFlagKeys.all, ({ signal }) => getFeatureFlags({ signal }), {
        select: (data) => data.find((flag) => flag.key === flagKey),
    });

export const useToggleFeatureFlag = () => {
    const queryClient = useQueryClient();

    return useMutation(toggleFeatureFlag, {
        onSuccess: () => {
            queryClient.invalidateQueries(featureFlagKeys.all);
        },
    });
};
