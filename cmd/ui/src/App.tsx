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
import {
    AppNotifications,
    GenericErrorBoundaryFallback,
    MainNav,
    MainNavData,
    NotificationsProvider,
    darkPalette,
    lightPalette,
    reactRouterFutureFlags,
    setRootClass,
    themedComponents,
    typography,
    useKeybindings,
    useShowNavBar,
    useStyles,
} from 'bh-shared-ui';
import { createBrowserHistory } from 'history';
import React, { useEffect } from 'react';
import { ErrorBoundary } from 'react-error-boundary';
import { Helmet } from 'react-helmet';
import { unstable_HistoryRouter as BrowserRouter } from 'react-router-dom';
import { initialize } from 'src/ducks/auth/authSlice';
import { PRIVILEGE_ZONES_ROUTE, ROUTES } from 'src/routes';
import { useAppDispatch, useAppSelector } from 'src/store';
import { initializeBHEClient } from 'src/utils';
import Content from 'src/views/Content';
import {
    useMainNavLogoData,
    useMainNavPrimaryListData,
    useMainNavSecondaryListData,
} from './components/MainNav/MainNavData';
import Notifier from './components/Notifier';
import DialogProviders from './views/Explore/DialogProviders';

// Create history object for unstable_HistoryRouter
// Type assertion is needed due to incompatibility between history v5 and react-router-dom v6's internal history types
// React Router team has explicitly deprecated custom history support and does not intend to support it in future versions.
// We should migrate from unstable_HistoryRouter to the regular BrowserRouter
const history = createBrowserHistory() as any;

export const Inner: React.FC = () => {
    const classes = useStyles();
    const dispatch = useAppDispatch();
    const authState = useAppSelector((state) => state.auth);
    const isOSDarkTheme = useMediaQuery('(prefers-color-scheme: dark)');

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

    // initialize authentication state and BHE client request/response handlers
    useEffect(() => {
        if (!authState.isInitialized) {
            dispatch(initialize());
            initializeBHEClient();
        }
    }, [dispatch, authState.isInitialized]);

    // block rendering until authentication initialization is complete
    if (!authState.isInitialized) {
        return null;
    }

    return (
        <>
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
            <div className={classes.applicationContainer} id='app-root'>
                {showNavBar && <MainNav mainNavData={mainNavData} />}
                <div className='bg-neutral-1 grow overflow-y-auto overflow-x-hidden'>
                    <Content />
                </div>
                <AppNotifications />
                <Notifier />
            </div>
        </>
    );
};

const App: React.FC = () => {
    const darkModeEnabled = useAppSelector((state) => state.global.view.darkMode);
    setRootClass(darkModeEnabled ? 'dark' : 'light');

    const palette = darkModeEnabled ? darkPalette : lightPalette;

    let theme = createTheme({
        palette,
        typography,
    });
    // suggested by MUI for defining theme options based on other options. https://mui.com/material-ui/customization/theming/#api
    theme = createTheme(theme, {
        components: themedComponents(palette),
    });

    return (
        <ThemeProvider theme={theme}>
            <CssBaseline />
            <BrowserRouter future={reactRouterFutureFlags} basename='/ui' history={history}>
                <NotificationsProvider>
                    <DialogProviders>
                        <ErrorBoundary fallbackRender={GenericErrorBoundaryFallback}>
                            <Inner />
                        </ErrorBoundary>
                    </DialogProviders>
                </NotificationsProvider>
            </BrowserRouter>
        </ThemeProvider>
    );
};

export default App;
