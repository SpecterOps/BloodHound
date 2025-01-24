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

export const ROUTES = [
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
        navigation: true,
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
