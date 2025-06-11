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
    DEFAULT_ZONE_MANAGEMENT_ROUTE,
    ROUTE_ZONE_MANAGEMENT_LABEL_DETAILS,
    ROUTE_ZONE_MANAGEMENT_LABEL_OBJECT_DETAILS,
    ROUTE_ZONE_MANAGEMENT_LABEL_SELECTOR_DETAILS,
    ROUTE_ZONE_MANAGEMENT_SUMMARY,
    ROUTE_ZONE_MANAGEMENT_SUMMARY_LABEL_DETAILS,
    ROUTE_ZONE_MANAGEMENT_SUMMARY_TIER_DETAILS,
    ROUTE_ZONE_MANAGEMENT_TIER_DETAILS,
    ROUTE_ZONE_MANAGEMENT_TIER_OBJECT_DETAILS,
    ROUTE_ZONE_MANAGEMENT_TIER_SELECTOR_DETAILS,
    Routable,
} from '../../routes';
import { cn, useAppNavigate } from '../../utils';
import { ZoneManagementContext } from './ZoneManagementContext';
import { OWNED_ID, TIER_ZERO_ID } from './utils';

const Details = React.lazy(() => import('./Details/Details'));
const Save = React.lazy(() => import('./Save'));
const Summary = React.lazy(() => import('./Summary/Summary'));

const detailsPaths = [
    ROUTE_ZONE_MANAGEMENT_TIER_DETAILS,
    ROUTE_ZONE_MANAGEMENT_LABEL_DETAILS,
    ROUTE_ZONE_MANAGEMENT_TIER_SELECTOR_DETAILS,
    ROUTE_ZONE_MANAGEMENT_LABEL_SELECTOR_DETAILS,
    ROUTE_ZONE_MANAGEMENT_TIER_OBJECT_DETAILS,
    ROUTE_ZONE_MANAGEMENT_LABEL_OBJECT_DETAILS,
];

const summaryPaths = [
    ROUTE_ZONE_MANAGEMENT_SUMMARY,
    ROUTE_ZONE_MANAGEMENT_SUMMARY_TIER_DETAILS,
    ROUTE_ZONE_MANAGEMENT_SUMMARY_LABEL_DETAILS,
];

const ZoneManagement: FC = () => {
    const navigate = useAppNavigate();
    const location = useLocation();

    const context = useContext(ZoneManagementContext);
    if (!context) {
        throw new Error('ZoneManagement must be used within a ZoneManagementContext.Provider');
    }
    const { savePaths, SupportLink } = context;

    const childRoutes: Routable[] = [
        ...detailsPaths.map((path) => {
            return { path, component: Details, authenticationRequired: true, navigation: true };
        }),
        ...savePaths.map((path) => {
            return { path, component: Save, authenticationRequired: true, navigation: true };
        }),
        ...summaryPaths.map((path) => {
            return { path, component: Summary, authenticationRequired: true, navigation: true };
        }),
    ];

    return (
        <main>
            <div className='h-dvh min-w-full px-8'>
                <h1 className='text-4xl font-bold pt-8'>Privilege Zone Management</h1>
                <p className='mt-6'>
                    Use Privilege Zones to segment and organize assets based on sensitivity and access level.
                    <SupportLink />
                </p>

                <div className='flex flex-col'>
                    <Tabs
                        defaultValue='tier'
                        className={cn('w-full mt-4', { hidden: location.pathname.includes('save') })}
                        value={location.pathname.includes('label') ? 'label' : 'tier'}
                        onValueChange={(value) => {
                            const isSummary = location.pathname.includes('summary');
                            const path = isSummary ? 'summary' : 'details';
                            const id = value === 'tier' ? TIER_ZERO_ID : OWNED_ID;
                            navigate(`/zone-management/${path}/${value}/${id}`);
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
                            <Route path='*' element={<Navigate to={DEFAULT_ZONE_MANAGEMENT_ROUTE} replace />} />
                        </Routes>
                    </Suspense>
                </div>
            </div>
        </main>
    );
};

export default ZoneManagement;
