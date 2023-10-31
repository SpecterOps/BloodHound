import { CssBaseline, StyledEngineProvider, ThemeProvider } from '@mui/material';
import { createTheme, Theme } from '@mui/material/styles';
import { createBrowserHistory } from 'history';
import { NotificationsProvider, GenericErrorBoundaryFallback } from 'bh-shared-ui';
import { ErrorBoundary } from 'react-error-boundary';
import { QueryClient, QueryClientProvider } from 'react-query';
import { ReactQueryDevtools } from 'react-query/devtools';
import { Provider } from 'react-redux';
import { unstable_HistoryRouter as BrowserRouter } from 'react-router-dom';
import App from '../App';
import { store } from '../store';
import {} from './BloodHoundUIContext';

import { BloodHoundUIContextProvider, BloodHoundUIProps } from '.';

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

const BloodHoundUI = ({ routes, components }: BloodHoundUIProps) => {
    return (
        <BloodHoundUIContextProvider value={{ routes, components }}>
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
        </BloodHoundUIContextProvider>
    );
};

export default BloodHoundUI;
