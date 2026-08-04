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
import { useCallback, useMemo } from 'react';
import { useQuery } from 'react-query';
import { apiClient, Permission } from '../utils';
import { usePermissions } from './usePermissions';

export const useSelf = () =>
    useQuery({
        queryKey: ['getSelf'],
        queryFn: ({ signal }) => apiClient.getSelf({ signal }).then((res) => res.data?.data),
    });

export const useBloodHoundUsers = () => {
    const { checkPermission } = usePermissions();
    const hasPermission = checkPermission(Permission.AUTH_MANAGE_USERS) || checkPermission(Permission.AUTH_READ_USERS);

    return useQuery({
        queryKey: ['listUsers'],
        queryFn: ({ signal }) => apiClient.listUsers({ signal }).then((res) => res.data?.data?.users),
        enabled: hasPermission,
    });
};

export const useGetUser = (userId?: string) => {
    return useQuery(
        ['getUser', userId],
        ({ signal }) => apiClient.getUser(userId!, { signal }).then((res) => res.data.data),
        { cacheTime: 0, enabled: !!userId }
    );
};

export const useUserNamesById = () => {
    const { data: users } = useBloodHoundUsers();

    return useMemo(() => {
        const map = new Map<string, string>();
        users?.forEach((user) => map.set(user.id, user.principal_name));
        return map;
    }, [users]);
};

export const useGetUserNameById = () => {
    const userNamesById = useUserNamesById();

    return useCallback((userId?: string) => (userId ? userNamesById.get(userId) : undefined), [userNamesById]);
};
