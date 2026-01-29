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
import { CssBaseline, ThemeProvider, useMediaQuery } from '@mui/material';
import { createTheme } from '@mui/material/styles';
import { createRootRoute, createRouter, RouteIds, RouterProvider } from '@tanstack/react-router';
import {
    AppNotifications,
    darkPalette,
    GenericErrorBoundaryFallback,
    lightPalette,
    MainNav,
    MainNavData,
    NotificationsProvider,
    setRootClass,
    themedComponents,
    typography,
    useKeybindings,
    useShowNavBar,
    useStyles,
} from 'bh-shared-ui';
import React, { useEffect } from 'react';
import { ErrorBoundary } from 'react-error-boundary';
import { Helmet } from 'react-helmet';
import { initialize } from 'src/ducks/auth/authSlice';
import { PRIVILEGE_ZONES_ROUTE, ROUTES } from 'src/routes';
import { useAppDispatch, useAppSelector } from 'src/store';
import { initializeBHEClient } from 'src/utils';
import Content from 'src/views/Content';
import NotFound from 'src/views/NotFound';
import {
    useMainNavLogoData,
    useMainNavPrimaryListData,
    useMainNavSecondaryListData,
} from '../components/MainNav/MainNavData';
import Notifier from '../components/Notifier';
import DialogProviders from '../views/Explore/DialogProviders';

const darkTheme = createTheme({
    palette: darkPalette,
    typography,
    components: themedComponents(darkPalette),
});

const lightTheme = createTheme({
    palette: lightPalette,
    typography,
    components: themedComponents(lightPalette),
});

export const RootComponent: React.FC = () => {
    const dispatch = useAppDispatch();
    const classes = useStyles();
    const isOSDarkTheme = useMediaQuery('(prefers-color-scheme: dark)');
    const darkModeEnabled = useAppSelector((state) => state.global.view.darkMode);
    const authState = useAppSelector((state) => state.auth);

    setRootClass(darkModeEnabled ? 'dark' : 'light');

    const theme = darkModeEnabled ? darkTheme : lightTheme;

    const mainNavData: MainNavData = {
        logo: useMainNavLogoData(),
        primaryList: useMainNavPrimaryListData(),
        secondaryList: useMainNavSecondaryListData(),
    };
    const showNavBar = useShowNavBar([...ROUTES, PRIVILEGE_ZONES_ROUTE]);

    useKeybindings({
        KeyD: () => {
            window.open('https://bloodhound.specterops.io/home', '_blank');
        },
    });

    useEffect(() => {
        // exit if already initialized
        if (authState.isInitialized) return;

        // otherwise initialize authentication state and BHE client request/response handlers
        dispatch(initialize());
        initializeBHEClient();
    }, [dispatch, authState.isInitialized]);

    // block rendering until authentication initialization is complete
    if (!authState.isInitialized) {
        return null;
    }

    return (
        <ThemeProvider theme={theme}>
            <Helmet>
                {
                    // dynamically set themed favicon by os/browser theme
                    // Why is this needed and the favicon definition in index.html?
                    // The helmet supports firefox, and index.html ensures a favicon is initially loaded when the tab first renders with the title.
                    isOSDarkTheme ? (
                        <link rel='shortcut icon' href='/ui/favicon-dark.ico' />
                    ) : (
                        <link rel='shortcut icon' href='/ui/favicon-light.ico' />
                    )
                }
            </Helmet>
            <CssBaseline />
            <RouterProvider router={router} />
            <NotificationsProvider>
                <DialogProviders>
                    <ErrorBoundary fallbackRender={GenericErrorBoundaryFallback}>
                        <div className={classes.applicationContainer} id='app-root'>
                            {showNavBar && <MainNav mainNavData={mainNavData} />}
                            <div className='bg-neutral-1 grow overflow-y-auto overflow-x-hidden'>
                                <Content />
                            </div>
                            <AppNotifications />
                            <Notifier />
                        </div>
                    </ErrorBoundary>
                </DialogProviders>
            </NotificationsProvider>
        </ThemeProvider>
    );
};

const routeTree: any = {};
const router = createRouter({
    routeTree,
    defaultPreload: 'intent',
    scrollRestoration: true,
    defaultNotFoundComponent: NotFound,
});

export type RouterType = typeof router;
export type RouterIds = RouteIds<RouterType['routeTree']>;

export const Route = createRootRoute({
    component: RootComponent,
});
