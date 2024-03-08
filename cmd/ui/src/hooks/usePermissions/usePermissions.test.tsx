// Copyright 2024 Specter Ops, Inc.
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

import { Permission, PERMISSIONS } from 'bh-shared-ui';
import { renderHook } from 'src/test-utils';
import usePermissions, { PermissionsFns } from './usePermissions';

describe('usePermissions', () => {
    const getPermissionsWithUser = (permissions: Permission[]): PermissionsFns => {
        return renderHook(() => usePermissions(), {
            initialState: {
                auth: {
                    user: {
                        roles: [
                            {
                                permissions: permissions.map((p) => PERMISSIONS[p]),
                            },
                        ],
                    },
                },
            },
        }).result.current;
    };

    const allPermissions = [
        Permission.CLIENTS_MANAGE,
        Permission.AUTH_CREATE_TOKEN,
        Permission.APP_READ_APPLICATION_CONFIGURATION,
    ];

    it('passes check if the user has a required permission', () => {
        const permissions = getPermissionsWithUser([Permission.CLIENTS_MANAGE]);

        const has = permissions.checkPermission(Permission.CLIENTS_MANAGE);
        const hasAll = permissions.checkAllPermissions([Permission.CLIENTS_MANAGE]);
        const hasAtLeastOne = permissions.checkAtLeastOnePermission([Permission.CLIENTS_MANAGE]);

        expect(has).toBe(true);
        expect(hasAll).toBe(true);
        expect(hasAtLeastOne).toBe(true);
    });

    it('passes checks if the user has multiple required permissions', () => {
        const permissions = getPermissionsWithUser(allPermissions);

        const hasAll = permissions.checkAllPermissions(allPermissions);
        const hasAtLeastOne = permissions.checkAtLeastOnePermission(allPermissions);

        expect(hasAll).toBe(true);
        expect(hasAtLeastOne).toBe(true);
    });

    it('fails checks if the user does not have a matching permission', () => {
        const permissions = getPermissionsWithUser([Permission.CLIENTS_MANAGE]);

        const has = permissions.checkPermission(Permission.AUTH_CREATE_TOKEN);
        const hasAll = permissions.checkAllPermissions([Permission.AUTH_CREATE_TOKEN]);
        const hasAtLeastOne = permissions.checkAtLeastOnePermission([Permission.AUTH_CREATE_TOKEN]);

        expect(has).toBe(false);
        expect(hasAll).toBe(false);
        expect(hasAtLeastOne).toBe(false);
    });

    it('passes the check for at least one permission if the user is missing one of many required permissions', () => {
        const permissions = getPermissionsWithUser([
            Permission.APP_READ_APPLICATION_CONFIGURATION,
            Permission.AUTH_CREATE_TOKEN,
        ]);

        const hasAll = permissions.checkAllPermissions(allPermissions);
        const hasAtLeastOne = permissions.checkAtLeastOnePermission(allPermissions);

        expect(hasAll).toBe(false);
        expect(hasAtLeastOne).toBe(true);
    });

    it('returns a list of the users current permissions', () => {
        const permissions = getPermissionsWithUser(allPermissions);
        const userPermissionsResult = permissions.getUserPermissions();

        expect(allPermissions.length).toEqual(userPermissionsResult.length);

        for (const permission of allPermissions) {
            expect(userPermissionsResult).toContain(permission);
        }
    });
});
