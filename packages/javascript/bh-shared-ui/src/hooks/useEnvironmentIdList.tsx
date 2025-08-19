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
    environments: Environment[]
) => {
    const collectedEnvironments = environments.filter((environment) => environment.collected);

    let aggregationIds;
    if (environmentAggregation === 'all') {
        aggregationIds = collectedEnvironments.map((environment) => environment.id);
    } else {
        aggregationIds = collectedEnvironments
            .filter((environment) => environment.type === environmentAggregation)
            .map((environment) => environment.id);
    }

    return aggregationIds;
};

export const useEnvironmentIdList = (
    ENVIRONMENT_AGGREGATION_SUPPORTED_ROUTES: (string | PathPattern)[]
): Environment['id'][] => {
    const { data: availableEnvironments } = useAvailableEnvironments();
    const { environment, environmentAggregation } = useSelectedEnvironment();
    const isEnvironmentAggregationSupportedPage = useMatchingPaths(ENVIRONMENT_AGGREGATION_SUPPORTED_ROUTES);

    if (isEnvironmentAggregationSupportedPage && environmentAggregation && availableEnvironments) {
        const aggregatedEnvironmentIds = getEnvironmentAggregationIds(environmentAggregation, availableEnvironments);
        return aggregatedEnvironmentIds;
    }

    if (environment?.id) return [environment.id];

    return [];
};
