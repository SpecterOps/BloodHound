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

import '@fontsource/roboto-mono';
import '@fontsource/roboto/300.css';
import '@fontsource/roboto/400.css';
import '@fontsource/roboto/500.css';
import '@fontsource/roboto/700.css';
import { CssBaseline, StyledEngineProvider, ThemeProvider } from '@mui/material';
import { createTheme, Theme } from '@mui/material/styles';
import { createBrowserHistory } from 'history';
import { NotificationsProvider, GenericErrorBoundaryFallback } from 'bh-shared-ui';
import { createRoot } from 'react-dom/client';
import { ErrorBoundary } from 'react-error-boundary';
import { QueryClient, QueryClientProvider } from 'react-query';
import { ReactQueryDevtools } from 'react-query/devtools';
import { Provider } from 'react-redux';
import { unstable_HistoryRouter as BrowserRouter } from 'react-router-dom';
import App from './App';
import { store } from './store';
import './styles/index.scss';

declare module '@mui/styles/defaultTheme' {
    // eslint-disable-next-line @typescript-eslint/no-empty-interface
    interface DefaultTheme extends Theme {}
}

declare module '@mui/material/styles' {
    interface Palette {
        neutral: Palette['primary'];
        low: string;
        moderate: string;
        high: string;
        critical: string;
    }
    interface PaletteOptions {
        neutral?: PaletteOptions['primary'];
        low: string;
        moderate: string;
        high: string;
        critical: string;
    }
}

declare module '@mui/material/Button' {
    interface ButtonPropsColorOverrides {
        neutral: true;
    }
}

declare module '@mui/material/IconButton' {
    interface IconButtonPropsColorOverrides {
        neutral: true;
    }
}

declare global {
    interface Window {
        Cypress: any;
        graphNodeInfo: {
            data: any;
            positions: {
                x: number;
                y: number;
            };
        };
    }
}

const theme = createTheme({
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
    typography: {
        h1: {
            fontWeight: 400,
            fontSize: '1.8rem',
            lineHeight: 2,
            letterSpacing: 0,
        },
        h2: {
            fontWeight: 500,
            fontSize: '1.5rem',
            lineHeight: 1.5,
            letterSpacing: 0,
        },
        h3: {
            fontWeight: 500,
            fontSize: '1.2rem',
            lineHeight: 1.25,
            letterSpacing: 0,
        },
        h4: {
            fontWeight: 500,
            fontSize: '1.25rem',
            lineHeight: 1.5,
            letterSpacing: 0,
        },
        h5: {
            fontWeight: 700,
            fontSize: '1.125rem',
            lineHeight: 1.5,
            letterSpacing: 0.25,
        },
        h6: {
            fontWeight: 700,
            fontSize: '1.0rem',
            lineHeight: 1.5,
            letterSpacing: 0.25,
        },
    },
    components: {
        MuiButton: {
            styleOverrides: {
                root: {
                    borderRadius: 999, // capsule-shaped buttons
                },
            },
        },
        MuiAccordionSummary: {
            styleOverrides: {
                root: {
                    flexDirection: 'row-reverse',
                },
                content: {
                    marginRight: '4px',
                },
            },
        },
    },
});

const queryClient = new QueryClient();

const main = async () => {
    const rootContainer = document.getElementById('root');
    const root = createRoot(rootContainer!);

    if (import.meta.env.DEV) {
        const { worker } = await import('./mocks/browser');
        await worker.start({
            serviceWorker: {
                url: '/ui/mockServiceWorker.js',
            },
            onUnhandledRequest: 'bypass',
        });
    }

    root.render(
        <Provider store={store}>
            <QueryClientProvider client={queryClient}>
                <ReactQueryDevtools />
                <StyledEngineProvider injectFirst>
                    <ThemeProvider theme={theme}>
                        <CssBaseline />
                        <BrowserRouter basename='/ui' history={createBrowserHistory()}>
                            <NotificationsProvider>
                                <ErrorBoundary fallbackRender={GenericErrorBoundaryFallback}>
                                    <App />
                                </ErrorBoundary>
                            </NotificationsProvider>
                        </BrowserRouter>
                    </ThemeProvider>
                </StyledEngineProvider>
            </QueryClientProvider>
        </Provider>
    );
};

main();
