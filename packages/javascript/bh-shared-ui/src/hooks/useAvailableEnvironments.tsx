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

import { Domain } from 'js-client-library';
import { useQuery, UseQueryOptions } from 'react-query';
import { apiClient } from '../utils/api';
import { useEnvironmentParams } from './useEnvironmentParams';

export const availableEnvironmentKeys = {
    all: ['available-domains'],
    getKey: (customKey?: string[]) =>
        customKey?.length ? [...availableEnvironmentKeys.all, ...customKey] : availableEnvironmentKeys.all,
};

type QueryOptions<T = Domain[]> = Omit<
    UseQueryOptions<Domain[], unknown, T | undefined, string[]>,
    'queryKey' | 'queryFn'
> & {
    appendQueryKey?: string[];
};

export function useAvailableEnvironments<T = Domain[]>(options?: QueryOptions<T>) {
    return useQuery({
        queryKey: availableEnvironmentKeys.getKey(options?.appendQueryKey),
        queryFn: ({ signal }) => apiClient.getAvailableEnvironments({ signal }).then((response) => response.data.data),
        ...options,
    });
}

export const selectEnvironment = (environmentId: Domain['id']): QueryOptions<Domain>['select'] => {
    return (data) => data.find((domain) => domain.id === environmentId);
};

export const useEnvironment = (environmentId?: Domain['id'], options?: Omit<QueryOptions<Domain>, 'select'>) => {
    const { environmentId: environmentIdParam } = useEnvironmentParams();
    const selectedEnvironment = environmentId ?? environmentIdParam;

    return useAvailableEnvironments({
        select: selectEnvironment(selectedEnvironment!),
        ...options,
        refetchOnWindowFocus: false,
        enabled: !!selectedEnvironment,
    });
};
