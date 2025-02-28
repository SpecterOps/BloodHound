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

import { useAvailableEnvironments, useEnvironmentParams, useFeatureFlag, useMatchingPaths } from 'bh-shared-ui';
import { Domain } from 'js-client-library';
import { useCallback } from 'react';
import { fullyAuthenticatedSelector } from 'src/ducks/auth/authSlice';
import { setDomain } from 'src/ducks/global/actions';
import { useAppDispatch, useAppSelector } from 'src/store';

// Future Dev: when we implement deep linking support for selected domain in BHE, move this to shared-ui and rip out the reducer logic (including stateUpdater)
export const useInitialEnvironment = (envSupportedRoutes: string[]) => {
    const authState = useAppSelector((state) => state.auth);
    const isFullyAuthenticated = useAppSelector(fullyAuthenticatedSelector);
    const reduxEnvironment = useAppSelector((state) => state.global.options.domain);

    const { data: flag } = useFeatureFlag('back_button_support', {
        enabled: !!authState.isInitialized && isFullyAuthenticated,
    });
    const dispatch = useAppDispatch();

    const { environmentId, setEnvironmentParams } = useEnvironmentParams();
    const environmentSupportedRoute = useMatchingPaths(envSupportedRoutes);
    const currentEnvironmentId = flag?.enabled ? environmentId : reduxEnvironment?.id;

    const stateUpdater = useCallback(
        (environment: Domain | null) => {
            if (flag?.enabled) {
                setEnvironmentParams({ environmentId: environment?.id });
            } else {
                dispatch(setDomain(environment));
            }
        },
        [dispatch, flag, setEnvironmentParams]
    );

    useAvailableEnvironments({
        queryKey: ['initial-environment', `current-environment-${currentEnvironmentId}`],
        // set initial environment/tenant once user is authenticated
        enabled: isFullyAuthenticated && environmentSupportedRoute,
        onError: () => stateUpdater(null),
        onSuccess: (availableEnvironments) => {
            if (!availableEnvironments?.length || currentEnvironmentId) return;

            const collectedEnvironments = availableEnvironments
                ?.filter((environment: Domain) => environment.collected) // omit uncollected environments
                .sort((a: Domain, b: Domain) => b.impactValue - a.impactValue); // sort by impactValue descending

            if (collectedEnvironments?.length) {
                stateUpdater(collectedEnvironments[0]);
            } else {
                stateUpdater(null);
            }
        },
    });
};
