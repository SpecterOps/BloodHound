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
import { TIER_ZERO_ID } from '../utils/tagUrlValue';

export const DEFAULT_TIER_MANAGEMENT_ROUTE = 'details/tier/' + TIER_ZERO_ID;

export const ROUTE_TIER_MANAGEMENT_DETAILS = '/details';
export const ROUTE_TIER_MANAGEMENT_TIER_DETAILS = '/details/tier/:tierId';
export const ROUTE_TIER_MANAGEMENT_LABEL_DETAILS = '/details/label/:labelId';
export const ROUTE_TIER_MANAGEMENT_TIER_SELECTOR_DETAILS = '/details/tier/:tierId/selector/:selectorId';
export const ROUTE_TIER_MANAGEMENT_TIER_OBJECT_DETAILS = '/details/tier/:tierId/selector/:selectorId/member/:memberId';
export const ROUTE_TIER_MANAGEMENT_LABEL_SELECTOR_DETAILS = '/details/label/:labelId/selector/:selectorId';
export const ROUTE_TIER_MANAGEMENT_LABEL_OBJECT_DETAILS =
    '/details/label/:labelId/selector/:selectorId/member/:memberId';
export const ROUTE_TIER_MANAGEMENT_SAVE = '/save';
export const ROUTE_TIER_MANAGEMENT_CREATE_TIER = '/save/tier';
export const ROUTE_TIER_MANAGEMENT_UPDATE_TIER = '/save/tier/:tierId';
export const ROUTE_TIER_MANAGEMENT_CREATE_LABEL = '/save/label';
export const ROUTE_TIER_MANAGEMENT_UPDATE_LABEL = '/save/label/:labelId';
export const ROUTE_TIER_MANAGEMENT_TIER_UPDATE_SELECTOR = '/save/tier/:tierId/selector/:selectorId';
export const ROUTE_TIER_MANAGEMENT_TIER_CREATE_SELECTOR = '/save/tier/:tierId/selector';
export const ROUTE_TIER_MANAGEMENT_LABEL_UPDATE_SELECTOR = '/save/label/:labelId/selector/:selectorId';
export const ROUTE_TIER_MANAGEMENT_LABEL_CREATE_SELECTOR = '/save/label/:labelId/selector';

export type Routable = {
    path: string;
    component: React.LazyExoticComponent<React.FC>;
    authenticationRequired: boolean;
    navigation: boolean;
    exact?: boolean;
};
