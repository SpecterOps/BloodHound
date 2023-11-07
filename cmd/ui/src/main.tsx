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
import './styles/index.scss';
import { createRoot } from 'react-dom/client';
import BloodHoundUI from './BloodHoundUI/BloodHoundUI';
import CustomHeader from './components/CustomHeader';

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

    const customRoutes = [
        {
            path: '/attack-paths',
            component: () => <span>Attack Paths Page</span>,
            authenticationRequired: true,
        },
        {
            path: '/posture',
            component: () => <span>Posture Page</span>,
            authenticationRequired: true,
        },
    ];

    const customComponents = {
        Header: CustomHeader,
    };

    const customReducers = {
        exampleReducer: (state: any = { key: 'value' }, action: any) => {
            switch (action.type) {
                case 'UPDATE_VALUE':
                    return { ...state, key: action.value };
                default:
                    return state;
            }
        },
    };

    root.render(
        <BloodHoundUI
            routes={(baseRoutes) => [...baseRoutes, ...customRoutes]}
            components={customComponents}
            reducers={customReducers}
        />
    );
};

main();
