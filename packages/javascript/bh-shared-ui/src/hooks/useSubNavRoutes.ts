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
import { SubNavItem, SubNavSection } from '../types';
import { Permission } from '../utils/permissions';
import { isFeatureFlagEnabled, useFeatureFlags } from './useFeatureFlags';
import { usePermissions } from './usePermissions';

/**
 * Filters and returns sub-navigation routes based on the current user's authentication state,
 * feature flags, and admin permissions.
 *
 * Each configured gate on an item must pass for the item to be visible:
 *   - `featureFlag` must be enabled (and feature flags must be loaded)
 *   - `permissions` must all be granted (and permissions must be loaded)
 *   - `adminOnly` requires application configuration read/write permissions
 * Items with no gates are always visible. Sections with no visible items are omitted.
 *
 * @param sections - The full list of sub-nav sections and their items to filter.
 * @param isAuthenticated - Whether the current user is authenticated. Controls whether
 *   feature flags are fetched.
 * @returns An object containing the filtered `routes` and an `areRoutesLoading` boolean
 *   indicating whether the feature flags are still being fetched.
 */
export function useSubNavRoutes(sections: SubNavSection[], isAuthenticated: boolean) {
    const { data: featureFlags, isLoading: areRoutesLoading } = useFeatureFlags({ enabled: isAuthenticated });
    const { checkAllPermissions, isLoading: arePermissionsLoading } = usePermissions();

    const hasAdminPermissions = checkAllPermissions([
        Permission.APP_READ_APPLICATION_CONFIGURATION,
        Permission.APP_WRITE_APPLICATION_CONFIGURATION,
    ]);

    const routes = useMemo(() => {
        const isItemVisible = (item: SubNavItem) => {
            if (item.featureFlag && (areRoutesLoading || !isFeatureFlagEnabled(item.featureFlag, featureFlags))) {
                return false;
            }
            if (item.permissions && (arePermissionsLoading || !checkAllPermissions(item.permissions))) {
                return false;
            }
            if (item.adminOnly && (arePermissionsLoading || !hasAdminPermissions)) {
                return false;
            }
            return true;
        };

        return sections
            .map(({ title, items }) => ({ title, items: items.filter(isItemVisible) }))
            .filter((section) => section.items.length > 0);
    }, [sections, featureFlags, areRoutesLoading, arePermissionsLoading, hasAdminPermissions, checkAllPermissions]);

    return { routes, areRoutesLoading };
}
