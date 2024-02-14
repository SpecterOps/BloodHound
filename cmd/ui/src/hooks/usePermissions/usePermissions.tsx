import { useCallback, useEffect, useState } from 'react';
import { getSelfResponse } from 'src/ducks/auth/types';
import { PermissionsSpec } from 'bh-shared-ui';
import { useAppSelector } from 'src/store';

type PermissionState = {
    hasAtLeastOne: boolean;
    hasAll: boolean;
};

const usePermissions: (permissions: PermissionsSpec[]) => PermissionState = (permissions) => {
    const [permissionState, setPermissionState] = useState<PermissionState>({ hasAtLeastOne: false, hasAll: false });
    const auth = useAppSelector((state) => state.auth);

    const checkUserPermissions = useCallback(
        (user: getSelfResponse): PermissionState => {
            const userPermissions = user.roles.map((role) => role.permissions).flat();
            let hasAtLeastOne = false;

            const hasAll = permissions.every((permission) => {
                return userPermissions.some((userPermission) => {
                    const matched =
                        userPermission.authority === permission.authority && userPermission.name === permission.name;

                    if (matched && !hasAtLeastOne) hasAtLeastOne = true;
                    return matched;
                });
            });

            return { hasAtLeastOne, hasAll };
        },
        [permissions]
    );

    useEffect(() => {
        if (auth.user) {
            setPermissionState(checkUserPermissions(auth.user));
        }
    }, [auth, checkUserPermissions]);

    return permissionState;
};

export default usePermissions;
