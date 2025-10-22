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
import { Route, Routes, useLocation } from 'react-router-dom';
import { useHighestPrivilegeTagId, useOwnedTagId, usePZPathParams } from '../../hooks';
import {
    ROUTE_PZ_CERTIFICATIONS,
    ROUTE_PZ_HISTORY,
    ROUTE_PZ_LABEL_DETAILS,
    ROUTE_PZ_LABEL_MEMBER_DETAILS,
    ROUTE_PZ_LABEL_SELECTOR_DETAILS,
    ROUTE_PZ_LABEL_SELECTOR_MEMBER_DETAILS,
    ROUTE_PZ_LABEL_SUMMARY,
    ROUTE_PZ_ZONE_DETAILS,
    ROUTE_PZ_ZONE_MEMBER_DETAILS,
    ROUTE_PZ_ZONE_SELECTOR_DETAILS,
    ROUTE_PZ_ZONE_SELECTOR_MEMBER_DETAILS,
    ROUTE_PZ_ZONE_SUMMARY,
    Routable,
    certificationsPath,
    detailsPath,
    historyPath,
    labelsPath,
    privilegeZonesPath,
    savePath,
    summaryPath,
    zonesPath,
} from '../../routes';
import { cn, useAppNavigate } from '../../utils';
import DefaultRoot from './DefaultRoot';
import { PrivilegeZonesContext } from './PrivilegeZonesContext';

const Details = React.lazy(() => import('./Details'));
const Save = React.lazy(() => import('./Save'));
const History = React.lazy(() => import('./History'));
const Certification = React.lazy(() => import('./Certification/Certification'));

const detailsPaths = [
    ROUTE_PZ_ZONE_DETAILS,
    ROUTE_PZ_LABEL_DETAILS,
    ROUTE_PZ_ZONE_SELECTOR_DETAILS,
    ROUTE_PZ_LABEL_SELECTOR_DETAILS,
    ROUTE_PZ_ZONE_MEMBER_DETAILS,
    ROUTE_PZ_ZONE_SELECTOR_MEMBER_DETAILS,
    ROUTE_PZ_LABEL_MEMBER_DETAILS,
    ROUTE_PZ_LABEL_SELECTOR_MEMBER_DETAILS,
];

const summaryPaths = [ROUTE_PZ_ZONE_SUMMARY, ROUTE_PZ_LABEL_SUMMARY];
const historyPaths = [ROUTE_PZ_HISTORY];
const certificationsPaths = [ROUTE_PZ_CERTIFICATIONS];

const PrivilegeZones: FC = () => {
    const navigate = useAppNavigate();
    const location = useLocation();
    const ownedId = useOwnedTagId();
    const { tagId } = useHighestPrivilegeTagId();
    const { isHistoryPage, tagType, isSummaryPage } = usePZPathParams();

    const context = useContext(PrivilegeZonesContext);
    if (!context) {
        throw new Error('PrivilegeZones must be used within a PrivilegeZonesContext.Provider');
    }
    const { savePaths, SupportLink, Summary } = context;

    const childRoutes: Routable[] = [
        ...detailsPaths.map((path) => {
            return { path, component: Details, authenticationRequired: true, navigation: true };
        }),
        ...savePaths.map((path) => {
            return { path, component: Save, authenticationRequired: true, navigation: true };
        }),
        ...historyPaths.map((path) => {
            return { path, component: History, authenticationRequired: true, navigation: true };
        }),
        ...certificationsPaths.map((path) => {
            return { path, component: Certification, authenticationRequired: true, navigation: true };
        }),
    ];

    if (Summary !== undefined) {
        childRoutes.push(
            ...summaryPaths.map((path) => {
                return { path, component: Summary, authenticationRequired: true, navigation: true };
            })
        );
    }

    const tabValue = location.pathname.includes(certificationsPath)
        ? certificationsPath
        : isHistoryPage
          ? historyPath
          : tagType;

    return (
        <main>
            <div className='h-dvh min-w-full px-8'>
                <h1 className='text-4xl font-bold pt-8'>Privilege Zone Management</h1>
                <p className='mt-6'>
                    Use Privilege Zones to segment and organize assets based on sensitivity and access level.
                    <br />
                    {SupportLink && <SupportLink />}
                </p>
                <div className='flex flex-col h-[75vh]'>
                    <Tabs
                        defaultValue={zonesPath}
                        value={tabValue}
                        className={cn('w-full mt-4', { hidden: location.pathname.includes(savePath) })}
                        onValueChange={(value) => {
                            if (value === certificationsPath) {
                                return navigate(`/${privilegeZonesPath}/${certificationsPath}`, {
                                    discardQueryParams: true,
                                });
                            }
                            if (value === historyPath) {
                                return navigate(`/${privilegeZonesPath}/${historyPath}`, { discardQueryParams: true });
                            } else {
                                const path = isSummaryPage ? summaryPath : detailsPath;
                                const id = value === zonesPath ? tagId : ownedId;
                                navigate(`/${privilegeZonesPath}/${value}/${id}/${path}?environmentAggregation=all`);
                            }
                        }}>
                        <TabsList className='w-full flex justify-start'>
                            <TabsTrigger value={zonesPath} data-testid='privilege-zones_tab-list_zones-tab'>
                                Zones
                            </TabsTrigger>
                            <TabsTrigger value={labelsPath} data-testid='privilege-zones_tab-list_labels-tab'>
                                Labels
                            </TabsTrigger>
                            <TabsTrigger value={historyPath} data-testid='privilege-zones_tab-list_history-tab'>
                                History
                            </TabsTrigger>
                            <TabsTrigger
                                value={certificationsPath}
                                data-testid='privilege-zones_tab-list_certifications-tab'>
                                Certifications
                            </TabsTrigger>
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
                            <Route path='*' element={<DefaultRoot defaultPath={context.defaultPath} />} />
                        </Routes>
                    </Suspense>
                </div>
            </div>
        </main>
    );
};

export default PrivilegeZones;
