// Copyright 2025 Specter Ops, Inc.
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

import { ErrorBoundary } from 'react-error-boundary';
import { Route } from 'react-router';
import { GenericErrorBoundaryFallback } from '../components';

export const ROUTE_TIER_MANAGEMENT = '/tier-management';
export const ROUTE_TIER_MANAGEMENT_EDIT = '/tier-management/edit';
export const ROUTE_TIER_MANAGEMENT_CREATE = '/tier-management/create';

export type Routable = {
    path: string;
    component: React.LazyExoticComponent<React.FC>;
    authenticationRequired: boolean;
    navigation: boolean;
    exact?: boolean;
};

export const mapRoutes = (routes: Routable[], AuthenticatedRoute: React.FC<{ children: React.ReactElement }>) => {
    return routes.map((route) => {
        return route.authenticationRequired ? (
            <Route
                path={route.path}
                element={
                    // Note: We add a left padding value to account for pages that have nav bar, h-full is because when adding the div it collapsed the views
                    <ErrorBoundary fallbackRender={GenericErrorBoundaryFallback}>
                        <AuthenticatedRoute>
                            <div className={`h-full ${route.navigation && 'pl-nav-width'} `}>
                                <route.component />
                            </div>
                        </AuthenticatedRoute>
                    </ErrorBoundary>
                }
                key={route.path}
            />
        ) : (
            <Route
                path={route.path}
                element={
                    <ErrorBoundary fallbackRender={GenericErrorBoundaryFallback}>
                        <route.component />
                    </ErrorBoundary>
                }
                key={route.path}
            />
        );
    });
};
