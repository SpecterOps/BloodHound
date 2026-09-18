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
import { Environment } from 'js-client-library';
import { PathPattern } from 'react-router-dom';
import { useAvailableEnvironments, useSelectedEnvironment } from './useAvailableEnvironments';
import { EnvironmentAggregation } from './useEnvironmentParams';
import { useMatchingPaths } from './useMatchingPaths';

export const getEnvironmentAggregationIds = (
    environmentAggregation: EnvironmentAggregation,
    environments: Environment[],
    collectedOnly: boolean = true
) => {
    const aggregationIds: string[] = [];

    environments.forEach((environment) => {
        //check if we're including the collected count
        if (
            (!collectedOnly || environment.collected) &&
            (environmentAggregation === 'all' || environment.type === environmentAggregation)
        ) {
            aggregationIds.push(environment.id);
        }
    });

    // Sort IDs to guarantee a stable order for cache keys
    return aggregationIds.sort((a, b) => String(a).localeCompare(String(b)));
};

export const useEnvironmentIdList = (
    ENVIRONMENT_AGGREGATION_SUPPORTED_ROUTES: (string | PathPattern)[],
    collectedOnly: boolean = true
): Environment['id'][] => {
    const { data: availableEnvironments } = useAvailableEnvironments();
    const { environment, environmentAggregation } = useSelectedEnvironment();
    const isEnvironmentAggregationSupportedPage = useMatchingPaths(ENVIRONMENT_AGGREGATION_SUPPORTED_ROUTES);

    if (isEnvironmentAggregationSupportedPage && environmentAggregation && availableEnvironments) {
        const aggregatedEnvironmentIds = getEnvironmentAggregationIds(
            environmentAggregation,
            availableEnvironments,
            collectedOnly
        );
        return aggregatedEnvironmentIds;
    }

    if (environment?.id) return [environment.id];

    return [];
};
