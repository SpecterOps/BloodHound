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

// Most of these routes are not being used, why do we have them?
export const DEFAULT_PRIVILEGE_ZONES_ROUTE = 'zone/';
export const ROUTE_PRIVILEGE_ZONES_ROOT = '/privilege-zones';

export const ROUTE_PRIVILEGE_ZONES_SUMMARY = '/summary';
export const ROUTE_PRIVILEGE_ZONES_SUMMARY_ZONE_DETAILS = `/zone/:zoneId/summary`;
export const ROUTE_PRIVILEGE_ZONES_SUMMARY_LABEL_DETAILS = `/label/:labelId/summary`;

export const ROUTE_PRIVILEGE_ZONES_DETAILS = '/details';
export const ROUTE_PRIVILEGE_ZONES_ZONE_DETAILS = `/zone/:zoneId/details`;
export const ROUTE_PRIVILEGE_ZONES_LABEL_DETAILS = `/label/:labelId/details`;
export const ROUTE_PRIVILEGE_ZONES_ZONE_SELECTOR_DETAILS = `/zone/:zoneId/details/selector/:selectorId`;
export const ROUTE_PRIVILEGE_ZONES_ZONE_OBJECT_DETAILS = `/zone/:zoneId/details/member/:memberId`;
export const ROUTE_PRIVILEGE_ZONES_ZONE_SELECTOR_OBJECT_DETAILS = `/zone/:zoneId/details/selector/:selectorId/member/:memberId`;
export const ROUTE_PRIVILEGE_ZONES_LABEL_SELECTOR_DETAILS = `/label/:labelId/details/selector/:selectorId`;
export const ROUTE_PRIVILEGE_ZONES_LABEL_OBJECT_DETAILS = '/label/:labelId/details/member/:memberId';
export const ROUTE_PRIVILEGE_ZONES_LABEL_SELECTOR_OBJECT_DETAILS =
    '/label/:labelId/details/selector/:selectorId/member/:memberId';

export const ROUTE_PRIVILEGE_ZONES_SAVE = '/save';
export const ROUTE_PRIVILEGE_ZONES_UPDATE_ZONE = '/zone/:zoneId/save';
export const ROUTE_PRIVILEGE_ZONES_UPDATE_LABEL = '/label/:labelId/save';
export const ROUTE_PRIVILEGE_ZONES_ZONE_UPDATE_SELECTOR = '/zone/:zoneId/save/selector/:selectorId';
export const ROUTE_PRIVILEGE_ZONES_ZONE_CREATE_SELECTOR = '/zone/:zoneId/save/selector';
export const ROUTE_PRIVILEGE_ZONES_LABEL_UPDATE_SELECTOR = '/label/:labelId/save/selector/:selectorId';
export const ROUTE_PRIVILEGE_ZONES_LABEL_CREATE_SELECTOR = '/label/:labelId/save/selector';

export type Routable = {
    path: string;
    component: React.LazyExoticComponent<React.FC>;
    authenticationRequired: boolean;
    navigation: boolean;
    exact?: boolean;
};
