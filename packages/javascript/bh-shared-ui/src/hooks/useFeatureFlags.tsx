// Copyright 2024 Specter Ops, Inc.
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

import type { AxiosError } from 'axios';
import { RequestOptions } from 'js-client-library';
import { UseQueryOptions, useMutation, useQuery, useQueryClient } from 'react-query';
import { apiClient } from '../utils';

export const FORBIDDEN = 403;

export type Flag = {
    id: number;
    key: string;
    name: string;
    description: string;
    enabled: boolean;
    user_updatable: boolean;
};

export const featureFlagKeys = {
    all: ['featureFlags'],
    getKey: (customKey?: string[]) =>
        customKey?.length ? [...featureFlagKeys.all, ...customKey] : featureFlagKeys.all,
};

export const getFeatureFlags = (options?: RequestOptions): Promise<Flag[]> => {
    return apiClient.getFeatureFlags(options).then((response) => response.data.data);
};

export const toggleFeatureFlag = (flagId: string | number, options?: RequestOptions) => {
    return apiClient.toggleFeatureFlag(flagId, options).then((response) => response.data);
};

type QueryOptions<T> = Omit<UseQueryOptions<Flag[], AxiosError, T | undefined, string[]>, 'queryFn'>;

export function useFeatureFlags<T = Flag[]>(options?: QueryOptions<T>) {
    const { queryKey, ...rest } = options ?? {};

    return useQuery({
        ...rest,
        queryKey: featureFlagKeys.getKey(options?.queryKey),
        queryFn: async ({ signal }) => {
            try {
                return await getFeatureFlags({ signal });
            } catch (error) {
                const status = (error as AxiosError).response?.status;
                // Ignore 403s as any retries will still be 403
                if (status === FORBIDDEN) {
                    return [] as Flag[];
                }
                throw error;
            }
        },
    });
}

export function useFeatureFlag(flagKey: string, options?: Omit<QueryOptions<Flag | undefined>, 'select'>) {
    return useFeatureFlags({ select: (data) => data.find((flag) => flag.key === flagKey), ...options });
}

export function useToggleFeatureFlag() {
    const queryClient = useQueryClient();

    return useMutation(toggleFeatureFlag, {
        onSuccess: () => {
            queryClient.invalidateQueries(featureFlagKeys.all);
        },
    });
}
