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

import { Permission } from 'bh-shared-ui';
import { useCallback, useEffect, useState } from 'react';
import { useAppSelector } from 'src/store';

export type PermissionsFns = {
    checkPermission: (permission: Permission) => boolean;
    checkAllPermissions: (permissions: Permission[]) => boolean;
    checkAtLeastOnePermission: (permissions: Permission[]) => boolean;
};

const usePermissions = () => {
    const auth = useAppSelector((state) => state.auth);
    const [userPermMap, setUserPermMap] = useState<Record<string, boolean>>({});

    const formatKey = useCallback((p: { authority: string; name: string }) => `${p.authority}-${p.name}`, []);

    useEffect(() => {
        const userPermissions = auth.user?.roles.map((role) => role.permissions).flat();

        if (userPermissions) {
            const newPermMap: Record<string, boolean> = {};
            userPermissions.forEach((perm) => (newPermMap[formatKey(perm)] = true));
            setUserPermMap(newPermMap);
        }
    }, [auth.user, formatKey]);

    const checkPermission = (permission: Permission): boolean => {
        return !!userPermMap[formatKey(permission.get())];
    };

    const checkAllPermissions = (permissions: Permission[]): boolean => {
        for (const permission of permissions) {
            const match = userPermMap[formatKey(permission.get())];
            if (!match) return false;
        }
        return true;
    };

    const checkAtLeastOnePermission = (permissions: Permission[]): boolean => {
        for (const permission of permissions) {
            const match = userPermMap[formatKey(permission.get())];
            if (match) return true;
        }
        return false;
    };

    return { checkPermission, checkAllPermissions, checkAtLeastOnePermission };
};
export default usePermissions;
