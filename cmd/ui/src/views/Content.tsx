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
import React, { Suspense, useEffect } from 'react';
import { ErrorBoundary } from 'react-error-boundary';
import { Route, Routes } from 'react-router-dom';
import { GenericErrorBoundaryFallback, apiClient } from 'bh-shared-ui';
import { ListAssetGroups } from 'src/ducks/assetgroups/actionCreators';
import { fullyAuthenticatedSelector } from 'src/ducks/auth/authSlice';
import { fetchAssetGroups, setDomain } from 'src/ducks/global/actions';
import * as routes from 'src/ducks/global/routes';
import { useAppDispatch, useAppSelector } from 'src/store';
import AuthenticatedRoute from 'src/components/AuthenticatedRoute';

const Login = React.lazy(() => import('src/views/Login'));
const DisabledUser = React.lazy(() => import('src/views/DisabledUser'));
const ExpiredPassword = React.lazy(() => import('src/views/ExpiredPassword'));
const Home = React.lazy(() => import('src/views/Home/Home'));
const NotFound = React.lazy(() => import('src/views/NotFound'));
const ExploreGraphView = React.lazy(() => import('./Explore/GraphView'));
const UserProfile = React.lazy(() => import('bh-shared-ui').then((module) => ({ default: module.UserProfile })));
const DownloadCollectors = React.lazy(() => import('./DownloadCollectors'));
const Administration = React.lazy(() => import('./Administration'));
const ApiExplorer = React.lazy(() => import('./ApiExplorer'));
const GroupManagement = React.lazy(() => import('./GroupManagement/GroupManagement'));

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

    useEffect(() => {
        if (isFullyAuthenticated) {
            dispatch(fetchAssetGroups());
            dispatch(ListAssetGroups());
        }
    }, [authState, isFullyAuthenticated, dispatch]);

    // set inital domain/tenant once user is authenticated
    useEffect(() => {
        if (isFullyAuthenticated) {
            const ctrl = new AbortController();
            apiClient
                .getAvailableDomains({ signal: ctrl.signal })
                .then((result) => {
                    const collectedDomains = result.data.data
                        // omit uncollected domains
                        .filter((domain: any) => domain.collected)
                        // sort by impactValue descending
                        .sort((a: any, b: any) => b.impactValue - a.impactValue);
                    if (collectedDomains.length > 0) {
                        dispatch(setDomain(collectedDomains[0]));
                    } else {
                        dispatch(setDomain(null));
                    }
                })
                .catch(() => {
                    dispatch(setDomain(null));
                });
            return () => ctrl.abort();
        }
    }, [isFullyAuthenticated, dispatch]);

    const ROUTES = [
        {
            path: routes.ROUTE_USER_DISABLED,
            component: DisabledUser,
            authenticationRequired: false,
        },
        {
            path: routes.ROUTE_LOGIN,
            component: Login,
            authenticationRequired: false,
        },
        {
            path: routes.ROUTE_EXPIRED_PASSWORD,
            component: ExpiredPassword,
            authenticationRequired: true,
        },
        {
            path: routes.ROUTE_HOME,
            component: Home,
            authenticationRequired: true,
        },
        {
            path: routes.ROUTE_EXPLORE,
            component: ExploreGraphView,
            authenticationRequired: true,
        },
        {
            path: routes.ROUTE_GROUP_MANAGEMENT,
            component: GroupManagement,
            authenticationRequired: true,
        },
        {
            path: routes.ROUTE_MY_PROFILE,
            component: UserProfile,
            authenticationRequired: true,
        },
        {
            path: routes.ROUTE_DOWNLOAD_COLLECTORS,
            component: DownloadCollectors,
            authenticationRequired: true,
        },
        {
            path: routes.ROUTE_ADMINISTRATION_ROOT,
            component: Administration,
            authenticationRequired: true,
        },
        {
            exact: true,
            path: routes.ROUTE_API_EXPLORER,
            component: ApiExplorer,
            authenticationRequired: true,
        },
        {
            exact: false,
            path: '*',
            component: NotFound,
            authenticationRequired: false,
        },
    ];

    return (
        <Box className={classes.content}>
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
                                    <ErrorBoundary fallbackRender={GenericErrorBoundaryFallback}>
                                        <AuthenticatedRoute>
                                            <route.component />
                                        </AuthenticatedRoute>
                                    </ErrorBoundary>
                                }
                                key={route.path}
                            />
                        ) : (
                            <Route
                                path={route.path}
                                element={
                                    <ErrorBoundary fallbackRender={GenericErrorBoundaryFallback}>
                                        <route.component />
                                    </ErrorBoundary>
                                }
                                key={route.path}
                            />
                        );
                    })}
                </Routes>
            </Suspense>
        </Box>
    );
};

export default Content;
