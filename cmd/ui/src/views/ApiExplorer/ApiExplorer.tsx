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

import React, { Suspense } from 'react';
import { GraphProgress } from 'bh-shared-ui';
import { HideEditionTagsPlugin } from './swagger/HideEditionTagsPlugin';
import { OperationsFilterPlugin } from './swagger/OperationsFilterPlugin';
import { OperationsLayoutPlugin } from './swagger/OperationsLayoutPlugin';
import 'swagger-ui-react/swagger-ui.css';
import { OperationsEditionPlugin } from './swagger/OperationsEditionPlugin';

const SwaggerUI = React.lazy(() => import('swagger-ui-react'));

const authInterceptor = (req: any) => {
    const state = localStorage.getItem('persistedState');
    if (state) {
        try {
            const persistedState = JSON.parse(state);
            const token = persistedState?.auth?.sessionToken;
            if (token) {
                req.headers.Authorization = `Bearer ${token}`;
            }
        } catch (e) {
            // no-op; couldn't parse persistedState
        }
    }
    return req;
};

const ApiExplorer: React.FC = () => {
    return (
        <Suspense fallback={<GraphProgress loading={true} />}>
            <SwaggerUI
                url='/api/v2/swagger/doc.json'
                requestInterceptor={authInterceptor}
                plugins={[
                    HideEditionTagsPlugin,
                    OperationsLayoutPlugin,
                    OperationsFilterPlugin,
                    OperationsEditionPlugin,
                ]}
                layout='OperationsLayout'
                filter={true}
            />
        </Suspense>
    );
};

export default ApiExplorer;
