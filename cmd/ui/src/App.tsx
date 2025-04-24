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
import { Box, CssBaseline, ThemeProvider } from '@mui/material';
import { createTheme } from '@mui/material/styles';
import {
    AppNotifications,
    GenericErrorBoundaryFallback,
    MainNav,
    MainNavData,
    NotificationsProvider,
    components,
    darkPalette,
    lightPalette,
    setRootClass,
    typography,
    useFeatureFlags,
    useShowNavBar,
    useStyles,
} from 'bh-shared-ui';
import React, { useEffect } from 'react';
import { ErrorBoundary } from 'react-error-boundary';
import { BrowserRouter } from 'react-router';
import { fullyAuthenticatedSelector, initialize } from 'src/ducks/auth/authSlice';
import { ROUTES, TIER_MANAGEMENT_ROUTES } from 'src/routes';
import { useAppDispatch, useAppSelector } from 'src/store';
import { initializeBHEClient } from 'src/utils';
import Content from 'src/views/Content';
import {
    useMainNavLogoData,
    useMainNavPrimaryListData,
    useMainNavSecondaryListData,
} from './components/MainNav/MainNavData';
import Notifier from './components/Notifier';
import { setDarkMode } from './ducks/global/actions';

export const Inner: React.FC = () => {
    const classes = useStyles();

    const dispatch = useAppDispatch();
    const authState = useAppSelector((state) => state.auth);
    const fullyAuthenticated = useAppSelector(fullyAuthenticatedSelector);
    const darkMode = useAppSelector((state) => state.global.view.darkMode);

    const featureFlagsRes = useFeatureFlags({
        retry: false,
        enabled: !!(authState.isInitialized && fullyAuthenticated),
    });

    const mainNavData: MainNavData = {
        logo: useMainNavLogoData(),
        primaryList: useMainNavPrimaryListData(),
        secondaryList: useMainNavSecondaryListData(),
    };
    const showNavBar = useShowNavBar([...ROUTES, ...TIER_MANAGEMENT_ROUTES]);

    // remove dark_mode if feature flag is disabled
    useEffect(() => {
        // TODO: Consider adding more flexibility/composability to side effects for toggling feature flags on and off
        if (!featureFlagsRes.data) return;
        const darkModeFeatureFlag = featureFlagsRes.data.find((flag) => flag.key === 'dark_mode');

        if (!darkModeFeatureFlag?.enabled) {
            dispatch(setDarkMode(false));
        }
    }, [dispatch, featureFlagsRes.data, darkMode]);

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
        <Box className={`${classes.applicationContainer}`} id='app-root'>
            {showNavBar && <MainNav mainNavData={mainNavData} />}
            <Box className={classes.applicationContent}>
                <Content />
            </Box>
            <AppNotifications />
            <Notifier />
        </Box>
    );
};

const App: React.FC = () => {
    const darkModeEnabled = useAppSelector((state) => state.global.view.darkMode);
    const currentMode = setRootClass(darkModeEnabled ? 'dark' : 'light');

    const palette = darkModeEnabled ? darkPalette : lightPalette;

    let theme = createTheme({
        palette: {
            mode: currentMode,
            ...palette,
        },
        typography,
    });
    // suggested by MUI for defining theme options based on other options. https://mui.com/material-ui/customization/theming/#api
    theme = createTheme(theme, {
        components: components(theme),
    });

    return (
        <ThemeProvider theme={theme}>
            <CssBaseline />
            <BrowserRouter basename='/ui'>
                <NotificationsProvider>
                    <ErrorBoundary fallbackRender={GenericErrorBoundaryFallback}>
                        <Inner />
                    </ErrorBoundary>
                </NotificationsProvider>
            </BrowserRouter>
        </ThemeProvider>
    );
};

export default App;
