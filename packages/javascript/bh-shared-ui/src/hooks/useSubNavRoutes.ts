// Copyright 2026 Specter Ops, Inc.
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

import { useMemo } from 'react';
import { SubNavSection } from '../types';
import { Permission } from '../utils/permissions';
import { isFeatureFlagEnabled, useFeatureFlags } from './useFeatureFlags';
import { usePermissions } from './usePermissions';

/**
 * Filters and returns sub-navigation routes based on the current user's authentication state,
 * feature flags, and admin permissions.
 *
 * Items with a `featureFlag` are included only if that flag is enabled.
 * Items marked `adminOnly` are included only if the user has the required application
 * configuration read/write permissions.
 * Sections with no visible items are omitted from the result.
 *
 * @param sections - The full list of sub-nav sections and their items to filter.
 * @param isAuthenticated - Whether the current user is authenticated. Controls whether
 *   feature flags are fetched.
 * @returns An object containing the filtered `routes` and an `areRoutesLoading` boolean
 *   indicating whether the feature flags are still being fetched.
 */
export function useSubNavRoutes(sections: SubNavSection[], isAuthenticated: boolean) {
    const { data: featureFlags, isLoading: areRoutesLoading } = useFeatureFlags({ enabled: isAuthenticated });
    const { checkAllPermissions, isLoading } = usePermissions();

    const hasAdminPermissions = checkAllPermissions([
        Permission.APP_READ_APPLICATION_CONFIGURATION,
        Permission.APP_WRITE_APPLICATION_CONFIGURATION,
    ]);

    const routes = useMemo(
        () =>
            sections
                .map(({ title, items }) => ({
                    title,
                    items: items.filter((item) => {
                        // If the item has a feature flag return it if it's enabled
                        if (item.featureFlag) {
                            return isFeatureFlagEnabled(item.featureFlag, featureFlags);
                        }

                        // If the item is admin only return if the user has admin permissions
                        if (item.adminOnly && isLoading) {
                            return false;
                        } else if (item.adminOnly && !isLoading) {
                            return hasAdminPermissions;
                        }

                        // Otherwise return the item
                        return true;
                    }),
                }))
                // Filter out any sections that have no items
                .filter((section) => section.items.length > 0),
        [featureFlags, hasAdminPermissions, isLoading, sections]
    );

    return { routes, areRoutesLoading };
}
