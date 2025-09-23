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
import { createContext, FC } from 'react';
import {
    ROUTE_PZ_LABEL_CREATE_SELECTOR,
    ROUTE_PZ_LABEL_UPDATE_SELECTOR,
    ROUTE_PZ_UPDATE_LABEL,
    ROUTE_PZ_UPDATE_ZONE,
    ROUTE_PZ_ZONE_CREATE_SELECTOR,
    ROUTE_PZ_ZONE_UPDATE_SELECTOR,
} from '../../routes';

const savePaths = [
    ROUTE_PZ_UPDATE_ZONE,
    ROUTE_PZ_UPDATE_LABEL,
    ROUTE_PZ_ZONE_CREATE_SELECTOR,
    ROUTE_PZ_LABEL_CREATE_SELECTOR,
    ROUTE_PZ_ZONE_UPDATE_SELECTOR,
    ROUTE_PZ_LABEL_UPDATE_SELECTOR,
];

const EmptyHeader: React.FC = () => <></>;
export interface PrivilegeZonesContextValue {
    savePaths: string[];
    InfoHeader: FC;
    SupportLink?: FC;
    SalesMessage?: FC;
    ZoneList?: FC<{ zones: AssetGroupTag[]; setPosition: (position: number) => void; name: string }>;
}

export const defaultPrivilegeZoneCtxValue: PrivilegeZonesContextValue = {
    savePaths,
    InfoHeader: EmptyHeader,
};

export const PrivilegeZonesContext = createContext<PrivilegeZonesContextValue>(defaultPrivilegeZoneCtxValue);
