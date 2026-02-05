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
import { Theme, ThemeOptions, createTheme } from '@mui/material/styles';
import { CssBaseline, StyledEngineProvider, ThemeProvider } from '@mui/material';
import { render, renderHook, RenderHookOptions, RenderHookResult, RenderResult } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from 'react-query';
import { BrowserRouter } from 'react-router-dom';
import { NotificationsProvider } from './providers';
import { darkPalette, reactRouterFutureFlags } from './constants';
import { SnackbarProvider } from 'notistack';

export const defaultTheme = createTheme({ palette: darkPalette });

const createDefaultQueryClient = () => {
    return new QueryClient({
        defaultOptions: {
            queries: {
                retry: false,
            },
        },
    });
};

const createProviders = ({
    queryClient,
    route,
    theme,
    children,
}: {
    queryClient: QueryClient;
    route: string;
    theme: ThemeOptions;
    children: React.ReactNode;
}) => {
    window.history.pushState({}, 'Initialize', route);
    return (
        <QueryClientProvider client={queryClient}>
            <StyledEngineProvider injectFirst>
                <ThemeProvider theme={theme}>
                    <NotificationsProvider>
                        <CssBaseline />
                        <BrowserRouter future={reactRouterFutureFlags}>
                            <SnackbarProvider>{children}</SnackbarProvider>
                        </BrowserRouter>
                    </NotificationsProvider>
                </ThemeProvider>
            </StyledEngineProvider>
        </QueryClientProvider>
    );
};

interface CustomRenderOptions {
    queryClient?: QueryClient;
    route?: string;
    theme?: Theme;
}

const customRender: (
    ui: React.ReactElement,
    renderOptions?: Parameters<typeof render>[1] & CustomRenderOptions
) => RenderResult = (ui, renderOptions = {}) => {
    const { queryClient = createDefaultQueryClient(), route = '/', theme = defaultTheme } = renderOptions;

    const AllTheProviders = ({ children }: { children: React.ReactNode }) =>
        createProviders({ queryClient, route, theme, children });

    return render(ui, { wrapper: AllTheProviders, ...renderOptions });
};

function customRenderHook<Result, Props>(
    hook: (initialProps: Props) => Result,
    renderOptions: RenderHookOptions<Props> & CustomRenderOptions = {}
): RenderHookResult<Result, Props> {
    const { queryClient = createDefaultQueryClient(), route = '/', theme = defaultTheme } = renderOptions;

    const AllTheProviders = ({ children }: { children: React.ReactNode }) =>
        createProviders({ queryClient, route, theme, children });

    return renderHook<Result, Props>(hook, { wrapper: AllTheProviders, ...renderOptions });
}

// re-export everything
export * from '@testing-library/react';
// override render and renderHook methods
export { customRender as render, customRenderHook as renderHook };
