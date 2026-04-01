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

import { Box, CircularProgress, Container } from '@mui/material';
import { AppNavigate, GenericErrorBoundaryFallback, flattenRoutes, getSubRoute } from 'bh-shared-ui';
import React, { Suspense } from 'react';
import { ErrorBoundary } from 'react-error-boundary';
import { Route, Routes } from 'react-router-dom';
import { useAdministrationRoutes } from 'src/hooks/useAdministrationRoutes';
import { DEFAULT_ADMINISTRATION_ROUTE, ROUTE_ADMINISTRATION } from 'src/routes/constants';

const Administration: React.FC = () => {
    const { routes: adminRoutes, areRoutesLoading } = useAdministrationRoutes();

    return (
        <Box className='flex h-full'>
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
                                    {flattenRoutes(adminRoutes).map((item) => (
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
                                    {!areRoutesLoading && (
                                        <Route
                                            path='*'
                                            element={<AppNavigate to={DEFAULT_ADMINISTRATION_ROUTE} replace />}
                                        />
                                    )}
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
