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

// organize-imports-ignore
import React from 'react';
import { createTheme } from '@mui/material/styles';
import { CssBaseline, StyledEngineProvider, ThemeProvider } from '@mui/material';
import { render, renderHook } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from 'react-query';
import { Router } from 'react-router-dom';
import { createMemoryHistory } from 'history';
import { NotificationsProvider } from './providers';
import { darkPalette } from './constants';
import { SnackbarProvider } from 'notistack';

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

const createProviders = ({ queryClient, history, theme, children }) => {
    return (
        <QueryClientProvider client={queryClient}>
            <StyledEngineProvider injectFirst>
                <ThemeProvider theme={theme}>
                    <NotificationsProvider>
                        <CssBaseline />
                        <Router location={history.location} navigator={history}>
                            <SnackbarProvider>{children}</SnackbarProvider>
                        </Router>
                    </NotificationsProvider>
                </ThemeProvider>
            </StyledEngineProvider>
        </QueryClientProvider>
    );
};

const customRender = (
    ui,
    {
        theme = defaultTheme,
        history = createMemoryHistory(),
        queryClient = createDefaultQueryClient(),
        ...renderOptions
    } = {}
) => {
    const AllTheProviders = ({ children }) => createProviders({ queryClient, history, theme, children });
    return render(ui, { wrapper: AllTheProviders, ...renderOptions });
};

const customRenderHook = (
    hook,
    {
        queryClient = createDefaultQueryClient(),
        theme = defaultTheme,
        history = createMemoryHistory(),
        ...renderOptions
    } = {}
) => {
    const AllTheProviders = ({ children }) => createProviders({ queryClient, history, theme, children });
    return renderHook(hook, { wrapper: AllTheProviders, ...renderOptions });
};

// re-export everything
export * from '@testing-library/react';
// override render and renderHook methods
export { customRender as render, customRenderHook as renderHook };
