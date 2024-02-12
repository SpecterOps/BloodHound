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

import { PermissionsAuthority, PermissionsName, PermissionsSpec } from 'bh-shared-ui';
import { renderHook } from 'src/test-utils';
import usePermissions from './usePermissions';

describe('usePermissions', () => {
    const checkPermissions = (permissions: { has: PermissionsSpec[]; needs: PermissionsSpec[] }) => {
        return renderHook(() => usePermissions(permissions.needs), {
            initialState: {
                auth: {
                    user: {
                        roles: [
                            {
                                permissions: permissions.has,
                            },
                        ],
                    },
                },
            },
        }).result.current;
    };

    const manageClientsPermission = {
        authority: PermissionsAuthority.CLIENTS,
        name: PermissionsName.MANAGE_CLIENTS,
    };

    const createTokenPermission = {
        authority: PermissionsAuthority.AUTH,
        name: PermissionsName.CREATE_TOKEN,
    };

    const manageAppConfigPermission = {
        authority: PermissionsAuthority.APP,
        name: PermissionsName.MANAGE_APP_CONFIG,
    };

    const allPermissions = [manageClientsPermission, createTokenPermission, manageAppConfigPermission];

    it('returns true if the user has a required permission', () => {
        const hasPermissions = checkPermissions({
            has: [manageClientsPermission],
            needs: [manageClientsPermission],
        });

        expect(hasPermissions).toBe(true);
    });

    it('returns true if the user has multiple required permissions', () => {
        const hasPermissions = checkPermissions({
            has: allPermissions,
            needs: allPermissions,
        });

        expect(hasPermissions).toBe(true);
    });

    it('returns false if the user does not have a matching permission', () => {
        const hasPermissions = checkPermissions({
            has: [manageClientsPermission],
            needs: [createTokenPermission],
        });

        expect(hasPermissions).toBe(false);
    });

    it('returns false if the user is missing one of many required permissions', () => {
        const hasPermissions = checkPermissions({
            has: [manageClientsPermission, createTokenPermission],
            needs: allPermissions,
        });

        expect(hasPermissions).toBe(false);
    });
});
