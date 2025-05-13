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

import { Box, CircularProgress } from '@mui/material';
import makeStyles from '@mui/styles/makeStyles';
import { GenericErrorBoundaryFallback, useFeatureFlag } from 'bh-shared-ui';
import React, { Suspense, useEffect } from 'react';
import { ErrorBoundary } from 'react-error-boundary';
import { Route, Routes } from 'react-router-dom';
import AuthenticatedRoute from 'src/components/AuthenticatedRoute';
import { ListAssetGroups } from 'src/ducks/assetgroups/actionCreators';
import { fullyAuthenticatedSelector } from 'src/ducks/auth/authSlice';
import { fetchAssetGroups, fetchCustomNodeInformation } from 'src/ducks/global/actions';
import { ROUTES, TIER_MANAGEMENT_ROUTE } from 'src/routes';
import { useAppDispatch, useAppSelector } from 'src/store';

const useStyles = makeStyles({
    content: {
        position: 'relative',
        width: '100%',
        height: '100%',
        minHeight: '100%',
    },
});

const Content: React.FC = () => {
    const classes = useStyles();
    const dispatch = useAppDispatch();
    const authState = useAppSelector((state) => state.auth);
    const isFullyAuthenticated = useAppSelector(fullyAuthenticatedSelector);
    const tierManagementFlag = useFeatureFlag('tier_management_engine', {
        enabled: !!authState.isInitialized && isFullyAuthenticated,
    }).data;
    const mapTierRoutes = tierManagementFlag?.enabled;

    useEffect(() => {
        if (isFullyAuthenticated) {
            dispatch(fetchAssetGroups());
            dispatch(ListAssetGroups());
            dispatch(fetchCustomNodeInformation());
        }
    }, [authState, isFullyAuthenticated, dispatch]);

    return (
        <Box className={classes.content}>
            <ErrorBoundary fallbackRender={GenericErrorBoundaryFallback}>
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
                        {ROUTES.map((route) => {
                            return route.authenticationRequired ? (
                                <Route
                                    path={route.path}
                                    element={
                                        // Note: We add a left padding value to account for pages that have nav bar, h-full is because when adding the div it collapsed the views
                                        <AuthenticatedRoute>
                                            <div className={`h-full ${route.navigation && 'pl-nav-width'} `}>
                                                <route.component />
                                            </div>
                                        </AuthenticatedRoute>
                                    }
                                    key={route.path}
                                />
                            ) : (
                                <Route path={route.path} element={<route.component />} key={route.path} />
                            );
                        })}
                        {mapTierRoutes && (
                            <Route
                                path={TIER_MANAGEMENT_ROUTE.path}
                                element={
                                    // Note: We add a left padding value to account for pages that have nav bar, h-full is because when adding the div it collapsed the views
                                    <AuthenticatedRoute>
                                        <div
                                            className={`h-full ${TIER_MANAGEMENT_ROUTE.navigation && 'pl-nav-width'} `}>
                                            <TIER_MANAGEMENT_ROUTE.component />
                                        </div>
                                    </AuthenticatedRoute>
                                }
                                key={TIER_MANAGEMENT_ROUTE.path}
                            />
                        )}
                    </Routes>
                </Suspense>
            </ErrorBoundary>
        </Box>
    );
};

export default Content;
