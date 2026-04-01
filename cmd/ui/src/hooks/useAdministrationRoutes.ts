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

import { isFeatureFlagEnabled, Permission, SubNavSection, useFeatureFlags, usePermissions } from 'bh-shared-ui';
import { lazy, useMemo } from 'react';
import { fullyAuthenticatedSelector } from 'src/ducks/auth/authSlice';
import {
    ROUTE_ADMINISTRATION_BLOODHOUND_CONFIGURATION,
    ROUTE_ADMINISTRATION_DATA_QUALITY,
    ROUTE_ADMINISTRATION_DB_MANAGEMENT,
    ROUTE_ADMINISTRATION_EARLY_ACCESS_FEATURES,
    ROUTE_ADMINISTRATION_FILE_INGEST,
    ROUTE_ADMINISTRATION_MANAGE_USERS,
    ROUTE_ADMINISTRATION_OPENGRAPH_MANAGEMENT,
    ROUTE_ADMINISTRATION_SSO_CONFIGURATION,
} from 'src/routes/constants';
import { useAppSelector } from 'src/store';

const DatabaseManagement = lazy(() => import('src/views/DatabaseManagement'));
const DataQuality = lazy(() => import('src/views/DataQuality'));
const Users = lazy(() => import('bh-shared-ui/Users'));
const EarlyAccessFeatures = lazy(() => import('src/views/EarlyAccessFeatures'));
const FileIngest = lazy(() => import('bh-shared-ui/FileIngest'));
const BloodHoundConfiguration = lazy(() => import('src/views/BloodHoundConfiguration'));
const SSOConfiguration = lazy(() => import('bh-shared-ui/SSOConfiguration'));
const OpenGraphManagement = lazy(() => import('bh-shared-ui/OpenGraphManagement'));

export const sections: SubNavSection[] = [
    {
        title: 'Data Collection',
        items: [
            {
                label: 'File Ingest',
                path: ROUTE_ADMINISTRATION_FILE_INGEST,
                component: FileIngest,
                adminOnly: false,
            },
            {
                label: 'Data Quality',
                path: ROUTE_ADMINISTRATION_DATA_QUALITY,
                component: DataQuality,
                adminOnly: false,
            },
            {
                label: 'Database Management',
                path: ROUTE_ADMINISTRATION_DB_MANAGEMENT,
                component: DatabaseManagement,
                adminOnly: false,
            },
        ],
    },
    {
        title: 'Users',
        items: [
            {
                label: 'Manage Users',
                path: ROUTE_ADMINISTRATION_MANAGE_USERS,
                component: Users,
                adminOnly: false,
            },
        ],
    },
    {
        title: 'Authentication',
        items: [
            {
                label: 'SSO Configuration',
                path: ROUTE_ADMINISTRATION_SSO_CONFIGURATION,
                component: SSOConfiguration,
                adminOnly: false,
            },
        ],
    },
    {
        title: 'Configuration',
        items: [
            {
                label: 'BloodHound Configuration',
                path: ROUTE_ADMINISTRATION_BLOODHOUND_CONFIGURATION,
                component: BloodHoundConfiguration,
                adminOnly: true,
            },
            {
                label: 'Early Access Features',
                path: ROUTE_ADMINISTRATION_EARLY_ACCESS_FEATURES,
                component: EarlyAccessFeatures,
                adminOnly: false,
            },
            {
                label: 'OpenGraph Management',
                path: ROUTE_ADMINISTRATION_OPENGRAPH_MANAGEMENT,
                component: OpenGraphManagement,
                adminOnly: false,
                featureFlag: 'opengraph_extension_management',
            },
        ],
    },
];

/**
 * Returns the administration routes that are available to the current user
 * based on their permissions and feature flags, along with a loading flag
 * indicating whether feature flags are still being fetched.
 */
export function useAdministrationRoutes() {
    const fullyAuthenticated = useAppSelector(fullyAuthenticatedSelector);
    const { data: featureFlags, isLoading: areRoutesLoading } = useFeatureFlags({ enabled: fullyAuthenticated });
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
        [featureFlags, hasAdminPermissions, isLoading]
    );

    return { routes, areRoutesLoading };
}
