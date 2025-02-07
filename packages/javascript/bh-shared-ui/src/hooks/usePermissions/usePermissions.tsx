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

import { RequestOptions } from 'js-client-library';
import { useCallback } from 'react';
import { useQuery } from 'react-query';
import { apiClient, Permission, PERMISSIONS } from '../../utils';

export type PermissionsFns = {
    getUserPermissions: () => Permission[];
    checkPermission: (permission: Permission) => boolean;
    checkAllPermissions: (permissions: Permission[]) => boolean;
    checkAtLeastOnePermission: (permissions: Permission[]) => boolean;
};

const formatKey = (p: { authority: string; name: string }) => `${p.authority}-${p.name}`;

const getSelf = (options?: RequestOptions) => apiClient.getSelf(options).then((res) => res.data.data);

export const usePermissions = () => {
    const getSelfQuery = useQuery(['getSelf'], ({ signal }) => getSelf({ signal }), {
        cacheTime: Number.POSITIVE_INFINITY,
        select: (data) => {
            const userPermissions = data?.roles.map((role: any) => role.permissions).flat() || [];
            const newPermMap: Record<string, boolean> = {};
            userPermissions.forEach((perm: any) => (newPermMap[formatKey(perm)] = true));
            return newPermMap;
        },
    });

    const getUserPermissions = useCallback((): Permission[] => {
        if (getSelfQuery.data === undefined) {
            return [];
        }
        return Object.entries(PERMISSIONS)
            .filter(([, definition]) => getSelfQuery.data[formatKey(definition)])
            .map(([name]) => parseInt(name));
    }, [getSelfQuery.data]);

    const checkPermission = useCallback(
        (permission: Permission): boolean => {
            if (getSelfQuery.data === undefined) {
                return false;
            }
            const definition = PERMISSIONS[permission];
            return definition && !!getSelfQuery.data[formatKey(definition)];
        },
        [getSelfQuery.data]
    );

    const checkAllPermissions = useCallback(
        (permissions: Permission[]): boolean => {
            if (getSelfQuery.data === undefined) {
                return false;
            }
            for (const permission of permissions) {
                if (!checkPermission(permission)) return false;
            }
            return true;
        },
        [checkPermission, getSelfQuery.data]
    );

    const checkAtLeastOnePermission = useCallback(
        (permissions: Permission[]): boolean => {
            if (getSelfQuery.data === undefined) {
                return false;
            }
            for (const permission of permissions) {
                if (checkPermission(permission)) return true;
            }
            return false;
        },
        [checkPermission, getSelfQuery.data]
    );

    return { ...getSelfQuery, getUserPermissions, checkPermission, checkAllPermissions, checkAtLeastOnePermission };
};
