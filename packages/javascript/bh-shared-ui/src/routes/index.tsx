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

export const DEFAULT_PRIVILEGE_ZONES_ROUTE = 'zone/';
export const ROUTE_PRIVILEGE_ZONES_ROOT = '/privilege-zones';

export const ROUTE_PRIVILEGE_ZONES_SUMMARY = '/summary';
export const ROUTE_PRIVILEGE_ZONES_ZONE_SUMMARY = `/${DEFAULT_PRIVILEGE_ZONES_ROUTE}:zoneId${ROUTE_PRIVILEGE_ZONES_SUMMARY}`;
export const ROUTE_PRIVILEGE_ZONES_LABEL_SUMMARY = `/label/:labelId${ROUTE_PRIVILEGE_ZONES_SUMMARY}`;

export const ROUTE_PRIVILEGE_ZONES_DETAILS = '/details';
export const ROUTE_PRIVILEGE_ZONES_ZONE_DETAILS = `/${DEFAULT_PRIVILEGE_ZONES_ROUTE}:zoneId${ROUTE_PRIVILEGE_ZONES_DETAILS}`;
export const ROUTE_PRIVILEGE_ZONES_LABEL_DETAILS = `/label/:labelId${ROUTE_PRIVILEGE_ZONES_DETAILS}`;
export const ROUTE_PRIVILEGE_ZONES_ZONE_SELECTOR_DETAILS = `/${DEFAULT_PRIVILEGE_ZONES_ROUTE}:zoneId${ROUTE_PRIVILEGE_ZONES_DETAILS}/selector/:selectorId`;
export const ROUTE_PRIVILEGE_ZONES_ZONE_OBJECT_DETAILS = `/${DEFAULT_PRIVILEGE_ZONES_ROUTE}:zoneId${ROUTE_PRIVILEGE_ZONES_DETAILS}/member/:memberId`;
export const ROUTE_PRIVILEGE_ZONES_ZONE_SELECTOR_OBJECT_DETAILS = `/${DEFAULT_PRIVILEGE_ZONES_ROUTE}:zoneId${ROUTE_PRIVILEGE_ZONES_DETAILS}/selector/:selectorId/member/:memberId`;
export const ROUTE_PRIVILEGE_ZONES_LABEL_SELECTOR_DETAILS = `/label/:labelId${ROUTE_PRIVILEGE_ZONES_DETAILS}/selector/:selectorId`;
export const ROUTE_PRIVILEGE_ZONES_LABEL_OBJECT_DETAILS = `/label/:labelId${ROUTE_PRIVILEGE_ZONES_DETAILS}/member/:memberId`;
export const ROUTE_PRIVILEGE_ZONES_LABEL_SELECTOR_OBJECT_DETAILS =
    `/label/:labelId${ROUTE_PRIVILEGE_ZONES_DETAILS}/selector/:selectorId/member/:memberId`;

export const ROUTE_PRIVILEGE_ZONES_SAVE = '/save';
export const ROUTE_PRIVILEGE_ZONES_UPDATE_ZONE = `/${DEFAULT_PRIVILEGE_ZONES_ROUTE}:zoneId${ROUTE_PRIVILEGE_ZONES_SAVE}`;
export const ROUTE_PRIVILEGE_ZONES_UPDATE_LABEL = `/label/:labelId${ROUTE_PRIVILEGE_ZONES_SAVE}`;
export const ROUTE_PRIVILEGE_ZONES_ZONE_UPDATE_SELECTOR = `/${DEFAULT_PRIVILEGE_ZONES_ROUTE}:zoneId${ROUTE_PRIVILEGE_ZONES_SAVE}/selector/:selectorId`;
export const ROUTE_PRIVILEGE_ZONES_ZONE_CREATE_SELECTOR = `/${DEFAULT_PRIVILEGE_ZONES_ROUTE}:zoneId${ROUTE_PRIVILEGE_ZONES_SAVE}/selector`;
export const ROUTE_PRIVILEGE_ZONES_LABEL_UPDATE_SELECTOR = `/label/:labelId${ROUTE_PRIVILEGE_ZONES_SAVE}/selector/:selectorId`;
export const ROUTE_PRIVILEGE_ZONES_LABEL_CREATE_SELECTOR = `/label/:labelId${ROUTE_PRIVILEGE_ZONES_SAVE}/selector`;

export type Routable = {
    path: string;
    component: React.LazyExoticComponent<React.FC>;
    authenticationRequired: boolean;
    navigation: boolean;
    exact?: boolean;
};
