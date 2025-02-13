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
import { StyledEngineProvider } from '@mui/material';
import { createRoot } from 'react-dom/client';
import { QueryClientProvider } from 'react-query';
import { ReactQueryDevtools } from 'react-query/devtools';
import { Provider } from 'react-redux';
import App from './App';
import { queryClient } from './queryClient';
import { store } from './store';
import './styles/index.css';
import './styles/index.scss';

declare module '@mui/material/styles' {
    interface Palette {
        neutral: { primary: string; secondary: string; tertiary: string; quaternary: string; quinary: string };
        color: { primary: string; links: string; error: string };
        low: string;
        moderate: string;
        high: string;
        critical: string;
    }
    interface PaletteOptions {
        neutral?: { primary: string; secondary: string; tertiary: string; quaternary: string; quinary: string };
        color: { primary: string; links: string; error: string };
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

const main = async () => {
    const rootContainer = document.getElementById('root');
    const root = createRoot(rootContainer!);

    if (import.meta.env.DEV && location.pathname.startsWith('/ui/')) {
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
                <ReactQueryDevtools position='bottom-right' />
                <StyledEngineProvider injectFirst>
                    <App />
                </StyledEngineProvider>
            </QueryClientProvider>
        </Provider>
    );
};

main();
