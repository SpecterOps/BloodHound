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
    PrimaryNavItem,
    components,
    darkPalette,
    lightPalette,
    setRootClass,
    typography,
    useFeatureFlags,
    useShowNavBar,
    useStyles,
} from 'bh-shared-ui';
import { createBrowserHistory } from 'history';
import React, { useEffect } from 'react';
import { ErrorBoundary } from 'react-error-boundary';
import { useQueryClient } from 'react-query';
import { unstable_HistoryRouter as BrowserRouter } from 'react-router-dom';
import { fullyAuthenticatedSelector, initialize } from 'src/ducks/auth/authSlice';
import { ROUTES } from 'src/routes';
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

const tierFlagToggle = (primaryNavList: PrimaryNavItem[], tierFlagEnabled?: boolean) => {
    if (tierFlagEnabled) {
        const groupManagementIndex = primaryNavList.findIndex((listItem) => {
            return listItem.label === groupManagementNavItem.label;
        });
        primaryNavList.splice(groupManagementIndex, 1, tierManagementNavItem);
    } else {
        const tierManagementIndex = primaryNavList.findIndex((listItem) => {
            return listItem.label === tierManagementNavItem.label;
        });
        primaryNavList.splice(tierManagementIndex, 1, groupManagementNavItem);
    }
};

export const Inner: React.FC = () => {
    const dispatch = useAppDispatch();
    const queryClient = useQueryClient();
    const classes = useStyles();

    const authState = useAppSelector((state) => state.auth);
    const fullyAuthenticated = useAppSelector(fullyAuthenticatedSelector);

    const featureFlagsRes = useFeatureFlags({
        retry: false,
        enabled: !!(authState.isInitialized && fullyAuthenticated),
    });

    const mainNavData: MainNavData = {
        logo: useMainNavLogoData(),
        primaryList: useMainNavPrimaryListData(),
        secondaryList: useMainNavSecondaryListData(),
    };
    const showNavBar = useShowNavBar(ROUTES);

    // initialize authentication state and BHE client request/response handlers
    useEffect(() => {
        if (!authState.isInitialized) {
            dispatch(initialize());
            initializeBHEClient();
        }
    }, [dispatch, authState.isInitialized]);

    // remove dark_mode if feature flag is disabled
    useEffect(() => {
        // TODO: Consider adding more flexibility/composability to side effects for toggling feature flags on and off
        if (!featureFlagsRes.data) return;
        const darkModeFeatureFlag = featureFlagsRes.data.find((flag) => flag.key === 'dark_mode');

        if (!darkModeFeatureFlag?.enabled) {
            dispatch(setDarkMode(false));
        }

        // Change the nav item routing for group/tier management based on flag value
        const tierManagementFlag = featureFlagsRes.data.find((flag) => flag.key === 'tier_management_engine');
        tierFlagToggle(mainNavData.primaryList, tierManagementFlag?.enabled);
    }, [dispatch, queryClient, featureFlagsRes.data, darkMode, mainNavData.primaryList]);

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
            <BrowserRouter basename='/ui' history={createBrowserHistory()}>
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
