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

import { AssetGroupTag } from 'js-client-library';
import { createContext, FC, useContext } from 'react';
import {
    detailsPath,
    ROUTE_PZ_LABEL_CREATE_RULE,
    ROUTE_PZ_LABEL_UPDATE_RULE,
    ROUTE_PZ_UPDATE_LABEL,
    ROUTE_PZ_UPDATE_ZONE,
    ROUTE_PZ_ZONE_CREATE_RULE,
    ROUTE_PZ_ZONE_UPDATE_RULE,
} from '../../routes';

const savePaths = [
    ROUTE_PZ_UPDATE_ZONE,
    ROUTE_PZ_UPDATE_LABEL,
    ROUTE_PZ_ZONE_CREATE_RULE,
    ROUTE_PZ_LABEL_CREATE_RULE,
    ROUTE_PZ_ZONE_UPDATE_RULE,
    ROUTE_PZ_LABEL_UPDATE_RULE,
];

const EmptyFragment: React.FC = () => <></>;

export interface PrivilegeZonesContextValue {
    defaultPath: string;
    savePaths: string[];
    InfoHeader: FC;
    EnvironmentSelector: FC;
    ZoneSelector?: FC<{ onZoneClick?: (zone: AssetGroupTag) => void }>;
    LabelSelector?: FC<{ onLabelClick?: (label: AssetGroupTag) => void }>;
    SupportLink?: FC;
    Summary?: React.LazyExoticComponent<React.FC>;
    Certification?: React.LazyExoticComponent<React.FC>;
    SalesMessage?: FC;
    ZoneList?: FC<{ zones: AssetGroupTag[]; setPosition: (position: number) => void; name: string }>;
}

export const defaultPrivilegeZoneCtxValue: PrivilegeZonesContextValue = {
    defaultPath: detailsPath,
    savePaths,
    InfoHeader: EmptyFragment,
    EnvironmentSelector: EmptyFragment,
};

export const PrivilegeZonesContext = createContext<PrivilegeZonesContextValue>(defaultPrivilegeZoneCtxValue);

export const usePZContext = () => {
    const context = useContext(PrivilegeZonesContext);
    if (!context) {
        throw new Error('usePZContext must be used within a PrivilegeZonesContext.Provider');
    }

    return context;
};
