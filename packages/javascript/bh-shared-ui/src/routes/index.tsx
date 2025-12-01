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
export const rulesPath = 'rules';
export const objectsPath = 'objects';

export const privilegeZonesPath = 'privilege-zones';
export const detailsPath = 'details';
export const savePath = 'save';
export const summaryPath = 'summary';
export const historyPath = 'history';
export const certificationsPath = 'certifications';

export const ROUTE_PRIVILEGE_ZONES = `/${privilegeZonesPath}`;

export const ROUTE_PZ_HISTORY = `/${historyPath}`;
export const ROUTE_PZ_CERTIFICATIONS = `/${certificationsPath}`;

export const ROUTE_PZ_ZONE_SUMMARY = `/${zonesPath}/:zoneId/${summaryPath}`;
export const ROUTE_PZ_LABEL_SUMMARY = `/${labelsPath}/:labelId/${summaryPath}`;

export const ROUTE_PZ_ZONE_DETAILS = `/${zonesPath}/:zoneId/${detailsPath}`;
export const ROUTE_PZ_LABEL_DETAILS = `/${labelsPath}/:labelId/${detailsPath}`;

export const ROUTE_PZ_ZONE_SELECTOR_DETAILS = `/${zonesPath}/:zoneId/${rulesPath}/:ruleId/${detailsPath}`;
export const ROUTE_PZ_LABEL_SELECTOR_DETAILS = `/${labelsPath}/:labelId/${rulesPath}/:ruleId/${detailsPath}`;

export const ROUTE_PZ_ZONE_MEMBER_DETAILS = `/${zonesPath}/:zoneId/${objectsPath}/:memberId/${detailsPath}`;
export const ROUTE_PZ_LABEL_MEMBER_DETAILS = `/${labelsPath}/:labelId/${objectsPath}/:memberId/${detailsPath}`;

export const ROUTE_PZ_ZONE_SELECTOR_MEMBER_DETAILS = `/${zonesPath}/:zoneId/${rulesPath}/:ruleId/${objectsPath}/:memberId/${detailsPath}`;
export const ROUTE_PZ_LABEL_SELECTOR_MEMBER_DETAILS = `/${labelsPath}/:labelId/${rulesPath}/:ruleId/${objectsPath}/:memberId/${detailsPath}`;

export const ROUTE_PZ_UPDATE_ZONE = `/${zonesPath}/:zoneId/${savePath}`;
export const ROUTE_PZ_UPDATE_LABEL = `/${labelsPath}/:labelId/${savePath}`;

export const ROUTE_PZ_ZONE_CREATE_SELECTOR = `/${zonesPath}/:zoneId/${rulesPath}/${savePath}`;
export const ROUTE_PZ_LABEL_CREATE_SELECTOR = `/${labelsPath}/:labelId/${rulesPath}/${savePath}`;

export const ROUTE_PZ_ZONE_UPDATE_SELECTOR = `/${zonesPath}/:zoneId/${rulesPath}/:ruleId/${savePath}`;
export const ROUTE_PZ_LABEL_UPDATE_SELECTOR = `/${labelsPath}/:labelId/${rulesPath}/:ruleId/${savePath}`;

export type Routable = {
    path: string;
    component: React.LazyExoticComponent<React.FC>;
    authenticationRequired: boolean;
    navigation: boolean;
    exact?: boolean;
};
