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
import React, { FC, Suspense } from 'react';
import { Route, Routes, useLocation } from 'react-router-dom';
import { useHighestPrivilegeTagId, useOwnedTagId, usePZPathParams } from '../../hooks';
import {
    ROUTE_PZ_CERTIFICATIONS,
    ROUTE_PZ_HISTORY,
    ROUTE_PZ_LABEL_DETAILS,
    ROUTE_PZ_LABEL_OBJECT_DETAILS,
    ROUTE_PZ_LABEL_RULE_DETAILS,
    ROUTE_PZ_LABEL_RULE_OBJECT_DETAILS,
    ROUTE_PZ_LABEL_SUMMARY,
    ROUTE_PZ_ZONE_DETAILS,
    ROUTE_PZ_ZONE_OBJECT_DETAILS,
    ROUTE_PZ_ZONE_RULE_DETAILS,
    ROUTE_PZ_ZONE_RULE_OBJECT_DETAILS,
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
import { useSelectedDetailsTabsContext } from './Details/SelectedDetailsTabs/SelectedDetailsTabsContext';
import PZDetailsTabsProvider from './Details/SelectedDetailsTabs/SelectedDetailsTabsProvider';
import { usePZContext } from './PrivilegeZonesContext';
import { TagTabValue } from './utils';

const Details = React.lazy(() => import('./Details'));
const Save = React.lazy(() => import('./Save'));
const History = React.lazy(() => import('./History'));

const detailsPaths = [
    ROUTE_PZ_ZONE_DETAILS,
    ROUTE_PZ_LABEL_DETAILS,
    ROUTE_PZ_ZONE_RULE_DETAILS,
    ROUTE_PZ_LABEL_RULE_DETAILS,
    ROUTE_PZ_ZONE_OBJECT_DETAILS,
    ROUTE_PZ_ZONE_RULE_OBJECT_DETAILS,
    ROUTE_PZ_LABEL_OBJECT_DETAILS,
    ROUTE_PZ_LABEL_RULE_OBJECT_DETAILS,
];

const summaryPaths = [ROUTE_PZ_ZONE_SUMMARY, ROUTE_PZ_LABEL_SUMMARY];
const historyPaths = [ROUTE_PZ_HISTORY];
const certificationsPaths = [ROUTE_PZ_CERTIFICATIONS];

const PrivilegeZones: FC = () => {
    const navigate = useAppNavigate();
    const location = useLocation();
    const ownedId = useOwnedTagId();
    const { tagId } = useHighestPrivilegeTagId();
    const { isCertificationsPage, isHistoryPage, tagType, isSummaryPage } = usePZPathParams();

    const { savePaths, Summary, Certification, defaultPath } = usePZContext();
    const { setSelectedDetailsTab } = useSelectedDetailsTabsContext();

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
    ];

    if (Summary !== undefined) {
        childRoutes.push(
            ...summaryPaths.map((path) => {
                return { path, component: Summary, authenticationRequired: true, navigation: true };
            })
        );
    }

    if (Certification !== undefined) {
        childRoutes.push(
            ...certificationsPaths.map((path) => {
                return { path, component: Certification, authenticationRequired: true, navigation: true };
            })
        );
    }

    const tabValue = isCertificationsPage ? certificationsPath : isHistoryPage ? historyPath : tagType;

    return (
        <main>
            <div className='h-dvh min-w-full px-8'>
                <h1 className='text-4xl font-bold pt-8'>Zone Builder</h1>
                <div className='flex flex-col h-[calc(100%-12rem)]'>
                    <Tabs
                        defaultValue={zonesPath}
                        value={tabValue}
                        className={cn('w-full mt-4', { hidden: location.pathname.includes(savePath) })}
                        onValueChange={(value) => {
                            setSelectedDetailsTab(TagTabValue);
                            const path = isSummaryPage ? summaryPath : detailsPath;
                            const id = value === zonesPath ? tagId : ownedId;
                            switch (value) {
                                case certificationsPath:
                                    return navigate(`/${privilegeZonesPath}/${certificationsPath}`, {
                                        discardQueryParams: true,
                                    });
                                case historyPath:
                                    return navigate(`/${privilegeZonesPath}/${historyPath}`, {
                                        discardQueryParams: true,
                                    });
                                case zonesPath:
                                case labelsPath:
                                    return navigate(`/${privilegeZonesPath}/${value}/${id}/${path}`);
                            }
                        }}>
                        <TabsList className='w-full flex justify-start'>
                            <TabsTrigger
                                // per https://github.com/radix-ui/primitives/issues/3013#issuecomment-2453054222
                                // aria-controls is optional, and default radix prop breaks accessibility
                                aria-controls={undefined}
                                value={zonesPath}
                                data-testid='privilege-zones_tab-list_zones-tab'>
                                Zones
                            </TabsTrigger>
                            <TabsTrigger
                                aria-controls={undefined}
                                value={labelsPath}
                                data-testid='privilege-zones_tab-list_labels-tab'>
                                Labels
                            </TabsTrigger>
                            {Certification && (
                                <TabsTrigger
                                    aria-controls={undefined}
                                    value={certificationsPath}
                                    data-testid='privilege-zones_tab-list_certifications-tab'>
                                    Certifications
                                </TabsTrigger>
                            )}
                            <TabsTrigger value={historyPath} data-testid='privilege-zones_tab-list_history-tab'>
                                History
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
                            <Route path='*' element={<DefaultRoot defaultPath={defaultPath} />} />
                        </Routes>
                    </Suspense>
                </div>
            </div>
        </main>
    );
};

const WrappedPrivilegeZones = () => {
    return (
        <PZDetailsTabsProvider>
            <PrivilegeZones />
        </PZDetailsTabsProvider>
    );
};

export default WrappedPrivilegeZones;
