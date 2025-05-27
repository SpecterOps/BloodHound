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

import { Routable } from 'bh-shared-ui';
import React from 'react';
import * as routes from 'src/routes/constants';

const Login = React.lazy(() => import('src/views/Login'));
const DisabledUser = React.lazy(() => import('src/views/DisabledUser'));
const ExpiredPassword = React.lazy(() => import('src/views/ExpiredPassword'));
const Home = React.lazy(() => import('src/views/Home/Home'));
const NotFound = React.lazy(() => import('src/views/NotFound'));
const ExploreGraphView = React.lazy(() => import('src/views/Explore/GraphView'));
const UserProfile = React.lazy(() => import('bh-shared-ui').then((module) => ({ default: module.UserProfile })));
const DownloadCollectors = React.lazy(() => import('src/views/DownloadCollectors'));
const Administration = React.lazy(() => import('src/views/Administration'));
const ApiExplorer = React.lazy(() => import('bh-shared-ui').then((module) => ({ default: module.ApiExplorer })));
const GroupManagement = React.lazy(() => import('src/views/GroupManagement/GroupManagement'));

export const ROUTES: Routable[] = [
    {
        path: routes.ROUTE_USER_DISABLED,
        component: DisabledUser,
        authenticationRequired: false,
        navigation: false,
    },
    {
        path: routes.ROUTE_LOGIN,
        component: Login,
        authenticationRequired: false,
        navigation: false,
    },
    {
        path: routes.ROUTE_EXPIRED_PASSWORD,
        component: ExpiredPassword,
        authenticationRequired: true,
        navigation: false,
    },
    {
        path: routes.ROUTE_HOME,
        component: Home,
        authenticationRequired: true,
        navigation: false,
    },
    {
        path: routes.ROUTE_EXPLORE,
        component: ExploreGraphView,
        authenticationRequired: true,
        navigation: true,
    },
    {
        path: routes.ROUTE_GROUP_MANAGEMENT,
        component: GroupManagement,
        authenticationRequired: true,
        navigation: true,
    },
    {
        path: routes.ROUTE_MY_PROFILE,
        component: UserProfile,
        authenticationRequired: true,
        navigation: true,
    },
    {
        path: routes.ROUTE_DOWNLOAD_COLLECTORS,
        component: DownloadCollectors,
        authenticationRequired: true,
        navigation: true,
    },
    {
        path: routes.ROUTE_ADMINISTRATION_ROOT,
        component: Administration,
        authenticationRequired: true,
        navigation: true,
    },
    {
        exact: true,
        path: routes.ROUTE_API_EXPLORER,
        component: ApiExplorer,
        authenticationRequired: true,
        navigation: true,
    },
    {
        exact: false,
        path: '*',
        component: NotFound,
        authenticationRequired: false,
        navigation: false,
    },
];

export const TIER_MANAGEMENT_ROUTE: Routable = {
    exact: true,
    path: routes.ROUTE_TIER_MANAGEMENT_ROOT,
    component: React.lazy(() => import('bh-shared-ui').then((module) => ({ default: module.TierManagement }))),
    authenticationRequired: true,
    navigation: true,
};
