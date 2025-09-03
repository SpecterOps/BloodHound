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

export const DEFAULT_ZONE_MANAGEMENT_ROUTE = 'zone/';
export const ROUTE_ZONE_MANAGEMENT_ROOT = '/privilege-zones';

export const ROUTE_ZONE_MANAGEMENT_SUMMARY = '/summary';
export const ROUTE_ZONE_MANAGEMENT_SUMMARY_TIER_DETAILS = `/zone/:tierId/summary`;
export const ROUTE_ZONE_MANAGEMENT_SUMMARY_LABEL_DETAILS = `label/:labelId/summary`;

export const ROUTE_ZONE_MANAGEMENT_DETAILS = '/details';
export const ROUTE_ZONE_MANAGEMENT_TIER_DETAILS = `/zone/:tierId/details`;
export const ROUTE_ZONE_MANAGEMENT_LABEL_DETAILS = `/label/:labelId/details`;
export const ROUTE_ZONE_MANAGEMENT_TIER_SELECTOR_DETAILS = `/zone/:tierId/details/selector/:selectorId`;
export const ROUTE_ZONE_MANAGEMENT_TIER_OBJECT_DETAILS = `/zone/:tierId/details/member/:memberId`;
export const ROUTE_ZONE_MANAGEMENT_TIER_SELECTOR_OBJECT_DETAILS = `/zone/:tierId/details/selector/:selectorId/member/:memberId`;
export const ROUTE_ZONE_MANAGEMENT_LABEL_SELECTOR_DETAILS = `/label/:labelId/details/selector/:selectorId`;
export const ROUTE_ZONE_MANAGEMENT_LABEL_OBJECT_DETAILS = '/label/:labelId/details/member/:memberId';
export const ROUTE_ZONE_MANAGEMENT_LABEL_SELECTOR_OBJECT_DETAILS =
    '/label/:labelId/details/selector/:selectorId/member/:memberId';

export const ROUTE_ZONE_MANAGEMENT_SAVE = '/save';
export const ROUTE_ZONE_MANAGEMENT_UPDATE_TIER = '/zone/:tierId/save';
export const ROUTE_ZONE_MANAGEMENT_UPDATE_LABEL = '/label/:labelId/save';
export const ROUTE_ZONE_MANAGEMENT_TIER_UPDATE_SELECTOR = '/tier/:tierId/save/selector/:selectorId';
export const ROUTE_ZONE_MANAGEMENT_TIER_CREATE_SELECTOR = '/tier/:tierId/save/selector';
export const ROUTE_ZONE_MANAGEMENT_LABEL_UPDATE_SELECTOR = '/label/:labelId/save/selector/:selectorId';
export const ROUTE_ZONE_MANAGEMENT_LABEL_CREATE_SELECTOR = '/label/:labelId/save/selector';

export type Routable = {
    path: string;
    component: React.LazyExoticComponent<React.FC>;
    authenticationRequired: boolean;
    navigation: boolean;
    exact?: boolean;
};
