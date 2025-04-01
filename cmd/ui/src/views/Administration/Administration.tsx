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
    GenericErrorBoundaryFallback,
    Permission,
    SubNav,
    getSubRoute,
    usePermissions,
} from 'bh-shared-ui';
import React, { Suspense } from 'react';
import { ErrorBoundary } from 'react-error-boundary';
import { Navigate, Route, Routes } from 'react-router-dom';
import {
    DEFAULT_ADMINISTRATION_ROUTE,
    ROUTE_ADMINISTRATION,
    ROUTE_ADMINISTRATION_BLOODHOUND_CONFIGURATION,
    ROUTE_ADMINISTRATION_DATA_QUALITY,
    ROUTE_ADMINISTRATION_DB_MANAGEMENT,
    ROUTE_ADMINISTRATION_EARLY_ACCESS_FEATURES,
    ROUTE_ADMINISTRATION_FILE_INGEST,
    ROUTE_ADMINISTRATION_MANAGE_USERS,
    ROUTE_ADMINISTRATION_SSO_CONFIGURATION,
} from 'src/routes/constants';

const DatabaseManagement = React.lazy(() => import('src/views/DatabaseManagement'));
const QA = React.lazy(() => import('src/views/QA'));
const Users = React.lazy(() => import('bh-shared-ui').then((module) => ({ default: module.Users })));
const EarlyAccessFeatures = React.lazy(() => import('src/views/EarlyAccessFeatures'));
const FileIngest = React.lazy(() => import('bh-shared-ui').then((module) => ({ default: module.FileIngest })));
const BloodHoundConfiguration = React.lazy(() => import('src/views/BloodHoundConfiguration'));
const SSOConfiguration = React.lazy(() =>
    import('bh-shared-ui').then((module) => ({ default: module.SSOConfiguration }))
);

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
                component: QA,
                adminOnly: false,
            },
            {
                label: 'Database Management',
                path: ROUTE_ADMINISTRATION_DB_MANAGEMENT,
                component: DatabaseManagement,
                adminOnly: false,
            },
        ],
        order: 0,
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
        order: 0,
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
        order: 0,
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
        order: 1,
    },
];

const Administration: React.FC = () => {
    const { checkAllPermissions } = usePermissions();

    // Checking these for now because the only route we are currently hiding is to the configuration page.
    // In practice, this will permit Administrators and Power User roles only.
    const hasAdminPermissions = checkAllPermissions([
        Permission.APP_READ_APPLICATION_CONFIGURATION,
        Permission.APP_WRITE_APPLICATION_CONFIGURATION,
    ]);

    // Filter adminOnly links from the data we pass to the sidebar if a user does not have the correct permissions
    const adminFilteredSections = sections
        .map((section) => {
            const filteredItems = section.items.filter((item) => !item.adminOnly || hasAdminPermissions);
            return {
                ...section,
                items: filteredItems,
            };
        })
        .filter((section) => section.items.length !== 0);

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
                                    {sections
                                        .sort((a, b) => a.order - b.order)
                                        .map((section) => section.items)
                                        .reduce((acc, val) => acc.concat(val), [])
                                        .map((item) => (
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
                                        element={
                                            <Navigate
                                                to={getSubRoute(ROUTE_ADMINISTRATION, DEFAULT_ADMINISTRATION_ROUTE)}
                                                replace
                                            />
                                        }
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
