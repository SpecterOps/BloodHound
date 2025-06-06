// Copyright 2024 Specter Ops, Inc.
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

import { CssBaseline, StyledEngineProvider, ThemeProvider } from '@mui/material';
import { createTheme } from '@mui/material/styles';
import { configureStore } from '@reduxjs/toolkit';
import { render, renderHook } from '@testing-library/react';
import { NotificationsProvider, darkPalette } from 'bh-shared-ui';
import { createMemoryHistory } from 'history';
import { SnackbarProvider } from 'notistack';
import { QueryClient, QueryClientProvider } from 'react-query';
import { Provider } from 'react-redux';
import { BrowserRouter } from 'react-router-dom';
import createSagaMiddleware from 'redux-saga';
import { rootReducer } from 'src/store';

const theme = createTheme(darkPalette);
const defaultTheme = {
    ...theme,
    palette: {
        ...theme.palette,
        neutral: { ...darkPalette.neutral },
        color: { ...darkPalette.color },
        tertiary: { ...darkPalette.tertiary },
    },
};

const createDefaultQueryClient = () => {
    return new QueryClient({
        defaultOptions: {
            queries: {
                retry: false,
            },
        },
    });
};

const createDefaultStore = (state) => {
    return configureStore({
        reducer: rootReducer,
        preloadedState: state,
        middleware: (getDefaultMiddleware) => {
            return [...getDefaultMiddleware({ serializableCheck: false }), createSagaMiddleware()];
        },
    });
};

const createProviders = ({ queryClient, route, theme, store, children }) => {
    window.history.pushState({}, 'Test page', route);

    return (
        <Provider store={store}>
            <QueryClientProvider client={queryClient}>
                <StyledEngineProvider injectFirst>
                    <ThemeProvider theme={theme}>
                        <CssBaseline />
                        <NotificationsProvider>
                            <BrowserRouter>
                                <SnackbarProvider>{children}</SnackbarProvider>
                            </BrowserRouter>
                        </NotificationsProvider>
                    </ThemeProvider>
                </StyledEngineProvider>
            </QueryClientProvider>
        </Provider>
    );
};

const customRender = (
    ui,
    {
        initialState = {},
        queryClient = createDefaultQueryClient(),
        route = '/',
        theme = defaultTheme,
        store = createDefaultStore(initialState),
        ...renderOptions
    } = {}
) => {
    const AllTheProviders = ({ children }) => createProviders({ queryClient, route, theme, store, children });
    return render(ui, { wrapper: AllTheProviders, ...renderOptions });
};

const customRenderHook = (
    hook,
    {
        initialState = {},
        queryClient = createDefaultQueryClient(),
        history = createMemoryHistory(),
        theme = defaultTheme,
        store = createDefaultStore(initialState),
        ...renderOptions
    } = {}
) => {
    const AllTheProviders = ({ children }) => createProviders({ queryClient, history, theme, store, children });
    return renderHook(hook, { wrapper: AllTheProviders, ...renderOptions });
};

// re-export everything
// eslint-disable-next-line react-refresh/only-export-components
export * from '@testing-library/react';
// override render method
export { customRender as render, customRenderHook as renderHook };
