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

import { useAvailableEnvironments } from 'bh-shared-ui';
import { Environment } from 'js-client-library';
import { orderBy } from 'lodash';
import { UseQueryOptions } from 'react-query';

export interface UseInitialEnvironmentParams {
    /** orderBy alphabetical in CE, we will use impactValue for BHE */
    orderBy: 'impactValue' | 'name';
    handleInitialEnvironment?: (env: Environment | null) => void;
    queryOptions?: Omit<
        UseQueryOptions<Environment[], unknown, Environment | undefined, string[]>,
        'queryFn' | 'onError' | 'onSuccess'
    >;
}

// Future Dev: when we implement deep linking support for selected domain in BHE, move this to shared-ui and rip out the reducer logic (including stateUpdater)
export const useInitialEnvironment = (options: UseInitialEnvironmentParams) => {
    const { orderBy: _orderBy, handleInitialEnvironment, queryOptions = {} } = options ?? { orderBy: 'impactValue' };
    const { queryKey = [], ...restOfQueryOptions } = queryOptions;

    return useAvailableEnvironments({
        queryKey: ['initial-environment', ...queryKey],
        // set initial environment/tenant once user is authenticated
        select: (availableEnvironments) => {
            if (!availableEnvironments?.length) return;

            const collectedEnvironments = availableEnvironments?.filter(
                (environment: Environment) => environment.collected
            );

            const sorted: Environment[] = orderBy(collectedEnvironments, [_orderBy], ['asc']);

            const initialEnvironment = sorted[0];

            if (handleInitialEnvironment) {
                handleInitialEnvironment(initialEnvironment);
            }

            return initialEnvironment;
        },
        refetchOnWindowFocus: false,
        ...restOfQueryOptions,
    });
};
