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
import { Role } from 'js-client-library';

// Some constants and helper functions for working with bloodhound user roles
export const Roles = {
    UPLOAD_ONLY: 'Upload-Only',
    READ_ONLY: 'Read-Only',
    AUDITOR: 'Auditor',
    USER: 'User',
    POWER_USER: 'Power User',
    ADMINISTRATOR: 'Administrator',
} as const;

type RoleValue = (typeof Roles)[keyof typeof Roles];

const ETAC_ROLES = [Roles.READ_ONLY, Roles.USER];
const ADMIN_ROLES = [Roles.ADMINISTRATOR, Roles.POWER_USER];
const DEFAULT_USER_ROLE = Roles.READ_ONLY;

export const isETACRole = (roleId: number | undefined, roles?: Role[]): boolean => {
    const matchingRole = roles?.find((role) => roleId === role.id)?.name;
    return !!(matchingRole && ETAC_ROLES.some((role) => role === matchingRole));
};

export const isAdminRole = (roleId: number | undefined, roles?: Role[]): boolean => {
    const matchingRole = roles?.find((role) => roleId === role.id)?.name;
    return !!(matchingRole && ADMIN_ROLES.some((role) => role === matchingRole));
};

export const getRoleId = (roleName: RoleValue, roles?: Role[]): number | undefined => {
    return roles?.find((role) => role.name === roleName)?.id;
};

export const getDefaultUserRoleId = (roles?: Role[]) => getRoleId(DEFAULT_USER_ROLE, roles);
