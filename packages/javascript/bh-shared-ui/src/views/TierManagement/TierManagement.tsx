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

import { Tabs, TabsList, TabsTrigger } from '@bloodhoundenterprise/doodleui';
import { CircularProgress } from '@mui/material';
import React, { FC, Suspense, useContext } from 'react';
import { Navigate, Route, Routes, useLocation } from 'react-router-dom';
import {
    DEFAULT_TIER_MANAGEMENT_ROUTE,
    ROUTE_TIER_MANAGEMENT_LABEL_DETAILS,
    ROUTE_TIER_MANAGEMENT_LABEL_OBJECT_DETAILS,
    ROUTE_TIER_MANAGEMENT_LABEL_SELECTOR_DETAILS,
    ROUTE_TIER_MANAGEMENT_TIER_DETAILS,
    ROUTE_TIER_MANAGEMENT_TIER_OBJECT_DETAILS,
    ROUTE_TIER_MANAGEMENT_TIER_SELECTOR_DETAILS,
    Routable,
} from '../../routes';
import { cn, useAppNavigate } from '../../utils';
import { TierManagementContext } from './TierManagementContext';
import { OWNED_ID, TIER_ZERO_ID } from './utils';

const Details = React.lazy(() => import('./Details/Details'));
const Save = React.lazy(() => import('./Save'));

const detailsPaths = [
    ROUTE_TIER_MANAGEMENT_TIER_DETAILS,
    ROUTE_TIER_MANAGEMENT_LABEL_DETAILS,
    ROUTE_TIER_MANAGEMENT_TIER_SELECTOR_DETAILS,
    ROUTE_TIER_MANAGEMENT_LABEL_SELECTOR_DETAILS,
    ROUTE_TIER_MANAGEMENT_TIER_OBJECT_DETAILS,
    ROUTE_TIER_MANAGEMENT_LABEL_OBJECT_DETAILS,
];

const TierManagement: FC = () => {
    const navigate = useAppNavigate();
    const location = useLocation();

    const { savePaths } = useContext(TierManagementContext);

    const childRoutes: Routable[] = [
        ...detailsPaths.map((path) => {
            return { path, component: Details, authenticationRequired: true, navigation: true };
        }),
        ...savePaths.map((path) => {
            return { path, component: Save, authenticationRequired: true, navigation: true };
        }),
    ];

    return (
        <main>
            <div className='h-dvh min-w-full px-8'>
                <h1 className='text-4xl font-bold pt-8'>Tier Management</h1>
                <p className='mt-6'>
                    <span>Define and manage selectors to dynamically gather objects based on criteria.</span>
                    <br />
                    <span>Ensure selectors capture the right assets for groups assignments or review.</span>
                </p>

                <div className='flex flex-col'>
                    <Tabs
                        defaultValue='tier'
                        className={cn('w-full mt-4', { hidden: location.pathname.includes('save') })}
                        value={location.pathname.includes('label') ? 'label' : 'tier'}
                        onValueChange={(value) => {
                            if (value === 'tier') {
                                navigate(`/tier-management/details/${value}/${TIER_ZERO_ID}`);
                            }
                            if (value === 'label') {
                                navigate(`/tier-management/details/${value}/${OWNED_ID}`);
                            }
                        }}>
                        <TabsList className='w-full flex justify-start'>
                            <TabsTrigger value='tier'>Tiers</TabsTrigger>
                            <TabsTrigger value='label'>Labels</TabsTrigger>
                        </TabsList>
                    </Tabs>
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
                            <Route path='*' element={<Navigate to={DEFAULT_TIER_MANAGEMENT_ROUTE} replace />} />
                        </Routes>
                    </Suspense>
                </div>
            </div>
        </main>
    );
};

export default TierManagement;
