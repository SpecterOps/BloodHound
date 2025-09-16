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
export const zonesPath = 'zones';
export const labelsPath = 'labels';
export const DEFAULT_PRIVILEGE_ZONES_ROUTE = `/${zonesPath}`;

export const privilegeZonesPath = 'privilege-zones';
export const ROUTE_PRIVILEGE_ZONES_ROOT = `/${privilegeZonesPath}`;

export const summaryPath = 'summary';
export const ROUTE_PZ_ZONE_SUMMARY = `/${zonesPath}/:zoneId/${summaryPath}`;
export const ROUTE_PZ_LABEL_SUMMARY = `/${labelsPath}/:labelId/${summaryPath}`;

export const detailsPath = 'details';
export const ROUTE_PZ_ZONE_DETAILS = `/${zonesPath}/:zoneId/${detailsPath}`;
export const ROUTE_PZ_LABEL_DETAILS = `/${labelsPath}/:labelId/${detailsPath}`;

export const selectorsPath = 'selectors';
export const ROUTE_PZ_ZONE_SELECTOR_DETAILS = `/${zonesPath}/:zoneId/${selectorsPath}/:selectorId/${detailsPath}`;
export const ROUTE_PZ_LABEL_SELECTOR_DETAILS = `/${labelsPath}/:labelId/${selectorsPath}/:selectorId/${detailsPath}`;

export const membersPath = 'members';
export const ROUTE_PZ_ZONE_OBJECT_DETAILS = `/${zonesPath}/:zoneId/${membersPath}/:memberId/${detailsPath}`;
export const ROUTE_PZ_LABEL_OBJECT_DETAILS = `/${labelsPath}/:labelId/${membersPath}/:memberId/${detailsPath}`;
export const ROUTE_PZ_ZONE_SELECTOR_OBJECT_DETAILS = `/${zonesPath}/:zoneId/${selectorsPath}/:selectorId/${membersPath}/:memberId/${detailsPath}`;
export const ROUTE_PZ_LABEL_SELECTOR_OBJECT_DETAILS = `/${labelsPath}/:labelId/${selectorsPath}/:selectorId/${membersPath}/:memberId/${detailsPath}`;

export const savePath = 'save';
export const ROUTE_PZ_CREATE_ZONE = `/${zonesPath}/${savePath}`;
export const ROUTE_PZ_UPDATE_ZONE = `/${zonesPath}/:zoneId/${savePath}`;
export const ROUTE_PZ_CREATE_LABEL = `/${labelsPath}/${savePath}`;
export const ROUTE_PZ_UPDATE_LABEL = `/${labelsPath}/:labelId/${savePath}`;
export const ROUTE_PZ_ZONE_UPDATE_SELECTOR = `/${zonesPath}/:zoneId/${selectorsPath}/:selectorId/${savePath}`;
export const ROUTE_PZ_ZONE_CREATE_SELECTOR = `/${zonesPath}/:zoneId/${selectorsPath}/${savePath}`;
export const ROUTE_PZ_LABEL_UPDATE_SELECTOR = `/${labelsPath}/:labelId/${selectorsPath}/:selectorId/${savePath}`;
export const ROUTE_PZ_LABEL_CREATE_SELECTOR = `/${labelsPath}/:labelId/${selectorsPath}/${savePath}`;

export type Routable = {
    path: string;
    component: React.LazyExoticComponent<React.FC>;
    authenticationRequired: boolean;
    navigation: boolean;
    exact?: boolean;
};
