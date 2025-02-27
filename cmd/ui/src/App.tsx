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
import { DefaultTheme } from '@mui/styles';
import makeStyles from '@mui/styles/makeStyles';
import {
    AppNotifications,
    GenericErrorBoundaryFallback,
    MainNav,
    NotificationsProvider,
    components,
    darkPalette,
    lightPalette,
    setRootClass,
    typography,
    useFeatureFlags,
    useShowNavBar,
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
import {
    useMainNavLogoData,
    useMainNavPrimaryListData,
    useMainNavSecondaryListData,
} from './components/MainNav/MainNavData';
import Notifier from './components/Notifier';
import { setDarkMode } from './ducks/global/actions';
import Content from './views/Content';

export const Inner: React.FC = () => {
    const dispatch = useAppDispatch();
    const authState = useAppSelector((state) => state.auth);
    const fullyAuthenticated = useAppSelector(fullyAuthenticatedSelector);
    const darkMode = useAppSelector((state) => state.global.view.darkMode);

    const queryClient = useQueryClient();

    const featureFlagsRes = useFeatureFlags({
        retry: false,
        enabled: !!(authState.isInitialized && fullyAuthenticated),
    });

    const mainNavData = {
        logo: useMainNavLogoData(),
        primaryList: useMainNavPrimaryListData(),
        secondaryList: useMainNavSecondaryListData(),
    };
    const showNavBar = useShowNavBar(ROUTES);

    const useStyles = makeStyles((theme: DefaultTheme) => ({
        applicationContainer: {
            display: 'flex',
            position: 'relative',
            flexDirection: 'column',
            height: '100%',
            overflow: 'hidden',
            '@global': {
                '.api-explorer .swagger-ui': {
                    [`& a.nostyle,
                    & div.renderedMarkdown > p,
                    & .response-col_status,
                    & .col_header,
                    & div.parameter__name,
                    & .parameter__in,
                    & div.opblock-summary-description,
                    & div > small,
                    & li.tabitem,
                    & .response-col_links,
                    & .opblock-description-wrapper > p,
                    & .btn-group > button,
                    & textarea,
                    & select,
                    & .parameter__type,
                    & .prop-format,
                    `]: {
                        color: theme.palette.color.primary,
                    },
                    ['& input, & textarea, & select, & .models, & .filter-container .operation-filter-input']: {
                        backgroundColor: theme.palette.neutral.primary,
                        border: `1px solid ${theme.palette.grey[700]}`,

                        '&:hover': {
                            borderColor: theme.palette.color.links,
                        },
                        '&:focus': {
                            outline: `1px solid ${theme.palette.color.links}`,
                        },
                    },
                    '& .models': {
                        '& h4': {
                            borderBottomColor: theme.palette.grey[700],
                        },
                        '& span, & table': {
                            color: theme.palette.color.primary,
                        },
                        '& svg': {
                            fill: theme.palette.color.primary,
                        },
                        '& model-box': {
                            backgroundColor: theme.palette.neutral.primary,
                        },
                    },
                    '& .parameter__name.required::after': {
                        color: theme.palette.color.error,
                    },
                    '& .responses-inner': {
                        [`& h4, & h5`]: {
                            color: theme.palette.color.primary,
                        },
                    },
                    '& svg': {
                        fill: theme.palette.color.primary,
                    },
                    '& .opblock-deprecated': {
                        '& .opblock-title_normal': {
                            color: theme.palette.color.primary,
                        },
                    },
                    '& .opblock-section-header': {
                        backgroundColor: theme.palette.neutral.primary,
                        '& h4, & .btn': {
                            color: theme.palette.color.primary,
                        },
                    },
                },
            },
        },
        applicationHeader: {
            flexGrow: 0,
            zIndex: theme.zIndex.drawer + 1,
        },
        applicationContent: {
            backgroundColor: theme.palette.neutral.primary,
            flexGrow: 1,
            overflowY: 'auto',
            overflowX: 'hidden',
        },
    }));

    const classes = useStyles();

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
    }, [dispatch, queryClient, featureFlagsRes.data, darkMode]);

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
