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

import { Environment } from 'js-client-library';
import { useQuery, UseQueryOptions } from 'react-query';
import { apiClient } from '../../utils/api';
import { useEnvironmentParams } from '../useEnvironmentParams';

export const availableEnvironmentKeys = {
    all: ['available-environments'],
    getKey: (customKey?: string[]) =>
        customKey?.length ? [...availableEnvironmentKeys.all, ...customKey] : availableEnvironmentKeys.all,
};

type QueryOptions<T = Environment[]> = Omit<
    UseQueryOptions<Environment[], unknown, T | undefined, string[]>,
    'queryFn'
>;

export function useAvailableEnvironments<T = Environment[]>(options?: QueryOptions<T>) {
    const { queryKey, ...rest } = options ?? {};

    return useQuery({
        queryKey: availableEnvironmentKeys.getKey(queryKey),
        queryFn: ({ signal }) => apiClient.getAvailableEnvironments({ signal }).then((response) => response.data.data),
        ...rest,
    });
}

export const selectEnvironment = (environmentId: Environment['id']): QueryOptions<Environment>['select'] => {
    return (data) => data.find((domain) => domain.id === environmentId);
};

export const useEnvironment = (
    environmentId?: Environment['id'] | null,
    options?: Omit<QueryOptions<Environment>, 'select'>
) => {
    const { environmentId: environmentIdParam } = useEnvironmentParams();
    const selectedEnvironment = environmentId ?? environmentIdParam;

    return useAvailableEnvironments({
        select: selectEnvironment(selectedEnvironment!),
        refetchOnWindowFocus: false,
        enabled: !!selectedEnvironment,
        ...options,
    });
};
