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

export const getDefaultRoleId = (roles?: Role[]): number | undefined => {
    return roles?.find((role) => DEFAULT_USER_ROLE === role.name)?.id;
};
