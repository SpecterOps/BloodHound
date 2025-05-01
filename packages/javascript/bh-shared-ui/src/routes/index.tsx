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

export const ROUTE_TIER_MANAGEMENT_ROOT = '/tier-management/*';
export const ROUTE_TIER_MANAGEMENT_DETAILS = '/details';
export const ROUTE_TIER_MANAGEMENT_TAG_DETAILS = '/details/tag/:tagId';
export const ROUTE_TIER_MANAGEMENT_SELECTOR_DETAILS = '/details/tag/:tagId/selector/:selectorId';
export const ROUTE_TIER_MANAGEMENT_OBJECT_DETAILS = '/details/tag/:tagId/selector/:selectorId/member/:memberId';
export const ROUTE_TIER_MANAGEMENT_EDIT = '/edit';
export const ROUTE_TIER_MANAGEMENT_EDIT_TAG = '/edit/tag/:tagId';
export const ROUTE_TIER_MANAGEMENT_EDIT_SELECTOR = '/edit/tag/:tagId/selector/:selectorId';
export const ROUTE_TIER_MANAGEMENT_CREATE_SELECTOR = '/edit/tag/:tagId/selector';
export const ROUTE_TIER_MANAGEMENT_CREATE = '/create';

export type Routable = {
    path: string;
    component: React.LazyExoticComponent<React.FC>;
    authenticationRequired: boolean;
    navigation: boolean;
    exact?: boolean;
};
