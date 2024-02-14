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

    it('permitted if the user has a required permission', () => {
        const permissions = checkPermissions({
            has: [manageClientsPermission],
            needs: [manageClientsPermission],
        });

        expect(permissions.hasAll).toBe(true);
        expect(permissions.hasAtLeastOne).toBe(true);
    });

    it('permitted if the user has multiple required permissions', () => {
        const permissions = checkPermissions({
            has: allPermissions,
            needs: allPermissions,
        });

        expect(permissions.hasAll).toBe(true);
        expect(permissions.hasAtLeastOne).toBe(true);
    });

    it('denied if the user does not have a matching permission', () => {
        const permissions = checkPermissions({
            has: [manageClientsPermission],
            needs: [createTokenPermission],
        });

        expect(permissions.hasAll).toBe(false);
        expect(permissions.hasAtLeastOne).toBe(false);
    });

    it('returns hasAtLeastOne if the user is missing one of many required permissions', () => {
        const permissions = checkPermissions({
            has: [manageClientsPermission, createTokenPermission],
            needs: allPermissions,
        });

        expect(permissions.hasAll).toBe(false);
        expect(permissions.hasAtLeastOne).toBe(true);
    });
});
