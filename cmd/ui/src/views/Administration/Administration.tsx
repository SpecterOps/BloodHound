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
import React, { Suspense } from 'react';
import { ErrorBoundary } from 'react-error-boundary';
import { Navigate, Route, Routes } from 'react-router-dom';
import { GenericErrorBoundaryFallback } from 'bh-shared-ui';
import LeftNav from 'src/components/LeftNav';
import {
    ROUTE_ADMINISTRATION_FILE_INGEST,
    ROUTE_ADMINISTRATION_DATA_QUALITY,
    ROUTE_ADMINISTRATION_EARLY_ACCESS_FEATURES,
    ROUTE_ADMINISTRATION_MANAGE_USERS,
    ROUTE_ADMINISTRATION_SAML_CONFIGURATION,
} from 'src/ducks/global/routes';
const FileIngest = React.lazy(() => import('src/views/FileIngest'));
const QA = React.lazy(() => import('src/views/QA'));
const Users = React.lazy(() => import('src/views/Users'));
const SAMLConfiguration = React.lazy(() => import('src/views/SAMLConfiguration'));
const EarlyAccessFeatures = React.lazy(() => import('src/views/EarlyAccessFeatures'));

const Administration: React.FC = () => {
    const sections = [
        {
            title: 'Data Collection',
            items: [
                {
                    label: 'File Ingest',
                    path: ROUTE_ADMINISTRATION_FILE_INGEST,
                    component: FileIngest,
                },
                {
                    label: 'Data Quality',
                    path: ROUTE_ADMINISTRATION_DATA_QUALITY,
                    component: QA,
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
                },
            ],
            order: 0,
        },
        {
            title: 'Authentication',
            items: [
                {
                    label: 'SAML Configuration',
                    path: ROUTE_ADMINISTRATION_SAML_CONFIGURATION,
                    component: SAMLConfiguration,
                },
            ],
            order: 0,
        },
        {
            title: 'Configuration',
            items: [
                {
                    label: 'Early Access Features',
                    path: ROUTE_ADMINISTRATION_EARLY_ACCESS_FEATURES,
                    component: EarlyAccessFeatures,
                },
            ],
            order: 1,
        },
    ];

    return (
        <Box display='flex' minHeight='100%'>
            <LeftNav sections={sections} />
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
                                                path={item.path.slice(16)}
                                                key={item.path}
                                                element={
                                                    <ErrorBoundary fallbackRender={GenericErrorBoundaryFallback}>
                                                        <item.component />
                                                    </ErrorBoundary>
                                                }
                                            />
                                        ))}
                                    <Route path='*' element={<Navigate to='file-ingest' replace />} />
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
