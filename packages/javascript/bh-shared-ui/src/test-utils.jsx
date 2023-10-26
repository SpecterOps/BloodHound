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

import React from 'react';
import { createTheme } from '@mui/material/styles';
import { CssBaseline, StyledEngineProvider, ThemeProvider } from '@mui/material';
import { render } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from 'react-query';
import { NotificationsProvider } from '.';

const customRender = (
    ui,
    queryClient = new QueryClient({
        defaultOptions: {
            queries: {
                retry: false,
            },
        },
    }),
    {
        theme = createTheme({
            palette: {
                primary: {
                    main: '#406f8e',
                    light: '#709dbe',
                    dark: '#064460',
                    contrastText: '#ffffff',
                },
                neutral: {
                    main: '#e0e0e0',
                    light: '#ffffff',
                    dark: '#cccccc',
                    contrastText: '#000000',
                },
                background: {
                    paper: '#fafafa',
                    default: '#e4e9eb',
                },
                low: 'rgb(255, 195, 15)',
                moderate: 'rgb(255, 97, 66)',
                high: 'rgb(205, 0, 117)',
                critical: 'rgb(76, 29, 143)',
            },
        }),

        ...renderOptions
    } = {}
) => {
    const AllTheProviders = ({ children }) => {
        return (
            <QueryClientProvider client={queryClient}>
                <StyledEngineProvider injectFirst>
                    <ThemeProvider theme={theme}>
                        <NotificationsProvider>
                            <CssBaseline />
                            {children}
                        </NotificationsProvider>
                    </ThemeProvider>
                </StyledEngineProvider>
            </QueryClientProvider>
        );
    };
    return render(ui, { wrapper: AllTheProviders, ...renderOptions });
};

// re-export everything
export * from '@testing-library/react';
// override render method
export { customRender as render };
