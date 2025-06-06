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

import { createContext, FC } from 'react';
import {
    ROUTE_TIER_MANAGEMENT_CREATE_LABEL,
    ROUTE_TIER_MANAGEMENT_CREATE_TIER,
    ROUTE_TIER_MANAGEMENT_LABEL_CREATE_SELECTOR,
    ROUTE_TIER_MANAGEMENT_LABEL_UPDATE_SELECTOR,
    ROUTE_TIER_MANAGEMENT_SAVE,
    ROUTE_TIER_MANAGEMENT_TIER_CREATE_SELECTOR,
    ROUTE_TIER_MANAGEMENT_TIER_UPDATE_SELECTOR,
    ROUTE_TIER_MANAGEMENT_UPDATE_LABEL,
    ROUTE_TIER_MANAGEMENT_UPDATE_TIER,
} from '../../routes';

const savePaths = [
    ROUTE_TIER_MANAGEMENT_SAVE,
    ROUTE_TIER_MANAGEMENT_CREATE_LABEL,
    ROUTE_TIER_MANAGEMENT_CREATE_TIER,
    ROUTE_TIER_MANAGEMENT_UPDATE_TIER,
    ROUTE_TIER_MANAGEMENT_UPDATE_LABEL,
    ROUTE_TIER_MANAGEMENT_TIER_CREATE_SELECTOR,
    ROUTE_TIER_MANAGEMENT_LABEL_CREATE_SELECTOR,
    ROUTE_TIER_MANAGEMENT_TIER_UPDATE_SELECTOR,
    ROUTE_TIER_MANAGEMENT_LABEL_UPDATE_SELECTOR,
];

export interface TierManagementContextValue {
    savePaths: string[];
    InfoHeader: FC;
}

const EmptyHeader: FC = () => <span></span>;

export const defaultTierMgmtCtxValue: TierManagementContextValue = {
    savePaths,
    InfoHeader: EmptyHeader,
};

export const TierManagementContext = createContext<TierManagementContextValue>(defaultTierMgmtCtxValue);
