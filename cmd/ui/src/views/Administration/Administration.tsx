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

import { Box, CircularProgress, Container } from '@mui/material';
import {
    AdministrationSection,
    AppNavigate,
    GenericErrorBoundaryFallback,
    Permission,
    SubNav,
    addItemToSection,
    filterAdminSections,
    flattenRoutes,
    getSubRoute,
    useFeatureFlag,
    usePermissions,
} from 'bh-shared-ui';
import React, { Suspense, useMemo } from 'react';
import { ErrorBoundary } from 'react-error-boundary';
import { Route, Routes } from 'react-router-dom';
import {
    DEFAULT_ADMINISTRATION_ROUTE,
    ROUTE_ADMINISTRATION,
    ROUTE_ADMINISTRATION_BLOODHOUND_CONFIGURATION,
    ROUTE_ADMINISTRATION_DATA_QUALITY,
    ROUTE_ADMINISTRATION_DB_MANAGEMENT,
    ROUTE_ADMINISTRATION_EARLY_ACCESS_FEATURES,
    ROUTE_ADMINISTRATION_FILE_INGEST,
    ROUTE_ADMINISTRATION_MANAGE_USERS,
    ROUTE_ADMINISTRATION_OPENGRAPH_MANAGEMENT,
    ROUTE_ADMINISTRATION_SSO_CONFIGURATION,
} from 'src/routes/constants';

const DatabaseManagement = React.lazy(() => import('src/views/DatabaseManagement'));
const DataQuality = React.lazy(() => import('src/views/DataQuality'));
const Users = React.lazy(() => import('bh-shared-ui/Users'));
const EarlyAccessFeatures = React.lazy(() => import('src/views/EarlyAccessFeatures'));
const FileIngest = React.lazy(() => import('bh-shared-ui/FileIngest'));
const BloodHoundConfiguration = React.lazy(() => import('src/views/BloodHoundConfiguration'));
const SSOConfiguration = React.lazy(() => import('bh-shared-ui/SSOConfiguration'));
const OpenGraphManagement = React.lazy(() => import('bh-shared-ui/OpenGraphManagement'));

const sections: AdministrationSection[] = [
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
        ],
    },
];

const openGraphManagement = {
    label: 'OpenGraph Management',
    path: ROUTE_ADMINISTRATION_OPENGRAPH_MANAGEMENT,
    component: OpenGraphManagement,
    adminOnly: false,
};

const Administration: React.FC = () => {
    const { data: openGraphFeatureFlag } = useFeatureFlag('opengraph_extension_management');

    // Add opengraph links and routes if the feature flag is enabled
    const sectionsWithFeatureFlag = useMemo(() => {
        if (!openGraphFeatureFlag?.enabled) {
            return sections;
        }
        return sections.map((s) => addItemToSection(s, 'Configuration', openGraphManagement));
    }, [openGraphFeatureFlag?.enabled]);

    // Checking these for now because the only route we are currently hiding is to the configuration page.
    // In practice, this will permit Administrators and Power User roles only.
    const { checkAllPermissions } = usePermissions();
    const hasAdminPermissions = checkAllPermissions([
        Permission.APP_READ_APPLICATION_CONFIGURATION,
        Permission.APP_WRITE_APPLICATION_CONFIGURATION,
    ]);

    // Filter adminOnly links from the data we pass to the sidebar if a user does not have the correct permissions
    const adminFilteredSections = useMemo(() => {
        if (hasAdminPermissions) {
            return sectionsWithFeatureFlag;
        }
        return filterAdminSections(sectionsWithFeatureFlag);
    }, [sectionsWithFeatureFlag, hasAdminPermissions]);

    return (
        <Box className='flex h-full pl-subnav-width'>
            <SubNav sections={adminFilteredSections} />
            <Box flexGrow={1} position='relative' minWidth={0}>
                <main>
                    <Container maxWidth='xl'>
                        <Box py={2}>
                            <Suspense
                                fallback={
                                    <Box
                                        position='absolute'
                                        top='0'
                                        left='0'
                                        right='0'
                                        bottom='0'
                                        display='flex'
                                        alignItems='center'
                                        justifyContent='center'
                                        zIndex={1000}>
                                        <CircularProgress color='primary' size={80} />
                                    </Box>
                                }>
                                <Routes>
                                    {flattenRoutes(adminFilteredSections).map((item) => (
                                        <Route
                                            path={getSubRoute(ROUTE_ADMINISTRATION, item.path)}
                                            key={item.path}
                                            element={
                                                <ErrorBoundary fallbackRender={GenericErrorBoundaryFallback}>
                                                    <item.component />
                                                </ErrorBoundary>
                                            }
                                        />
                                    ))}
                                    <Route
                                        path='*'
                                        element={<AppNavigate to={DEFAULT_ADMINISTRATION_ROUTE} replace />}
                                    />
                                </Routes>
                            </Suspense>
                        </Box>
                    </Container>
                </main>
            </Box>
        </Box>
    );
};

export default Administration;
