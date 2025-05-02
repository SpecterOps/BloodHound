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

import { CircularProgress } from '@mui/material';
import React, { FC, Suspense } from 'react';
import { Navigate, Route, Routes } from 'react-router-dom';
import {
    ROUTE_TIER_MANAGEMENT_CREATE_SELECTOR,
    ROUTE_TIER_MANAGEMENT_EDIT,
    ROUTE_TIER_MANAGEMENT_EDIT_SELECTOR,
    ROUTE_TIER_MANAGEMENT_EDIT_TAG,
    ROUTE_TIER_MANAGEMENT_OBJECT_DETAILS,
    ROUTE_TIER_MANAGEMENT_SELECTOR_DETAILS,
    ROUTE_TIER_MANAGEMENT_TAG_DETAILS,
    Routable,
} from '../../routes';

const Details = React.lazy(() => import('./Details/Details'));
const Save = React.lazy(() => import('./Save'));

const childRoutes: Routable[] = [
    { path: ROUTE_TIER_MANAGEMENT_TAG_DETAILS, component: Details, authenticationRequired: true, navigation: true },
    {
        path: ROUTE_TIER_MANAGEMENT_SELECTOR_DETAILS,
        component: Details,
        authenticationRequired: true,
        navigation: true,
    },
    { path: ROUTE_TIER_MANAGEMENT_OBJECT_DETAILS, component: Details, authenticationRequired: true, navigation: true },
    { path: ROUTE_TIER_MANAGEMENT_EDIT, component: Save, authenticationRequired: true, navigation: true },
    { path: ROUTE_TIER_MANAGEMENT_EDIT_TAG, component: Save, authenticationRequired: true, navigation: true },
    {
        path: ROUTE_TIER_MANAGEMENT_CREATE_SELECTOR,
        component: Save,
        authenticationRequired: true,
        navigation: true,
    },
    {
        path: ROUTE_TIER_MANAGEMENT_EDIT_SELECTOR,
        component: Save,
        authenticationRequired: true,
        navigation: true,
    },
];

const TierManagement: FC = () => {
    return (
        <main className='pl-nav-width'>
            <div className='min-h-full min-w-full px-8'>
                <h1 className='text-4xl font-bold pt-8'>Tier Management</h1>
                <p className='mt-6'>
                    <span>Define and manage selectors to dynamically gather objects based on criteria.</span>
                    <br />
                    <span>Ensure selectors capture the right assets for groups assignments or review.</span>
                </p>

                <div className='flex flex-col'>
                    <div className='flex gap-4 mt-6 invisible'>
                        <div className='text-lg underline'>Tiers</div>
                        <div className='text-lg'>Labels</div>
                        <div className='text-lg'>Certifications</div>
                        <div className='text-lg'>History</div>
                    </div>
                    <Suspense
                        fallback={
                            <div className='absolute inset-0 flex items-center justify-center'>
                                <CircularProgress color='primary' size={80} />
                            </div>
                        }>
                        <Routes>
                            {childRoutes.map((route) => {
                                return <Route path={route.path} element={<route.component />} key={route.path} />;
                            })}
                            <Route path='*' element={<Navigate to='details/tag/1' replace />} />
                        </Routes>
                    </Suspense>
                </div>
            </div>
        </main>
    );
};

export default TierManagement;
