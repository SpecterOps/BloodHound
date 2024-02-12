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

import { useCallback, useEffect, useState } from 'react';
import { useSelector } from 'react-redux';
import { getSelfResponse } from 'src/ducks/auth/types';
import { PermissionsSpec } from 'bh-shared-ui';

const usePermissions: (permissions: PermissionsSpec[]) => boolean = (permissions) => {
    const [hasAllPermissions, setHasAllPermissions] = useState<boolean>(false);
    const auth: { user: getSelfResponse } = useSelector((state: any) => state.auth);

    const doesUserHavePermissions = useCallback(
        (user: getSelfResponse): boolean => {
            const userPermissions = user.roles.map((role) => role.permissions).flat();

            return permissions.every((permission) => {
                return userPermissions.some((userPermission) => {
                    return userPermission.authority === permission.authority && userPermission.name === permission.name;
                });
            });
        },
        [permissions]
    );

    useEffect(() => {
        setHasAllPermissions(auth.user && doesUserHavePermissions(auth.user));
    }, [auth, doesUserHavePermissions]);

    return hasAllPermissions;
};

export default usePermissions;
