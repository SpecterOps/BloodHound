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
export const zonePath = 'zones';
export const labelPath = 'labels';
export const DEFAULT_PRIVILEGE_ZONES_ROUTE = `/${zonePath}`;

export const privilegeZonesPath = 'privilege-zones';
export const ROUTE_PRIVILEGE_ZONES_ROOT = '/privilege-zones';

export const summaryPath = 'summary';
export const ROUTE_PRIVILEGE_ZONES_SUMMARY = '/summary';
export const ROUTE_PRIVILEGE_ZONES_ZONE_SUMMARY = `/${zonePath}/:zoneId/${summaryPath}`;
export const ROUTE_PRIVILEGE_ZONES_LABEL_SUMMARY = `/${labelPath}/:labelId/${summaryPath}`;

export const detailsPath = 'details';
export const ROUTE_PRIVILEGE_ZONES_DETAILS = '/details';
export const ROUTE_PRIVILEGE_ZONES_ZONE_DETAILS = `/${zonePath}/:zoneId/${detailsPath}`;
export const ROUTE_PRIVILEGE_ZONES_LABEL_DETAILS = `/${labelPath}/:labelId/${detailsPath}`;

export const selectorPath = 'selectors';
export const ROUTE_PRIVILEGE_ZONES_ZONE_SELECTOR_DETAILS = `/${zonePath}/:zoneId/${selectorPath}/:selectorId/${detailsPath}`;
export const ROUTE_PRIVILEGE_ZONES_LABEL_SELECTOR_DETAILS = `/${labelPath}/:labelId/${selectorPath}/:selectorId/${detailsPath}`;

export const memberPath = 'members';
export const ROUTE_PRIVILEGE_ZONES_ZONE_OBJECT_DETAILS = `/${zonePath}/:zoneId/${memberPath}/:memberId/${detailsPath}`;
export const ROUTE_PRIVILEGE_ZONES_LABEL_OBJECT_DETAILS = `/${labelPath}/:labelId/${memberPath}/:memberId/${detailsPath}`;
export const ROUTE_PRIVILEGE_ZONES_ZONE_SELECTOR_OBJECT_DETAILS = `/${zonePath}/:zoneId/${selectorPath}/:selectorId/${memberPath}/:memberId/${detailsPath}`;
export const ROUTE_PRIVILEGE_ZONES_LABEL_SELECTOR_OBJECT_DETAILS = `/${labelPath}/:labelId/${selectorPath}/:selectorId/${memberPath}/:memberId/${detailsPath}`;

export const savePath = 'save';
export const ROUTE_PRIVILEGE_ZONES_SAVE = '/save';
export const ROUTE_PRIVILEGE_ZONES_UPDATE_ZONE = `/${zonePath}/:zoneId/${savePath}`;
export const ROUTE_PRIVILEGE_ZONES_UPDATE_LABEL = `/${labelPath}/:labelId/${savePath}`;
export const ROUTE_PRIVILEGE_ZONES_ZONE_UPDATE_SELECTOR = `/${zonePath}/:zoneId/${selectorPath}/:selectorId/${savePath}`;
export const ROUTE_PRIVILEGE_ZONES_ZONE_CREATE_SELECTOR = `/${zonePath}/:zoneId/${selectorPath}/${savePath}`;
export const ROUTE_PRIVILEGE_ZONES_LABEL_UPDATE_SELECTOR = `/${labelPath}/:labelId/${selectorPath}/:selectorId/${savePath}`;
export const ROUTE_PRIVILEGE_ZONES_LABEL_CREATE_SELECTOR = `/${labelPath}/:labelId/${selectorPath}/${savePath}`;

export type Routable = {
    path: string;
    component: React.LazyExoticComponent<React.FC>;
    authenticationRequired: boolean;
    navigation: boolean;
    exact?: boolean;
};
