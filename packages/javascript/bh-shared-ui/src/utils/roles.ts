import { Role } from 'js-client-library';

export const Roles = {
    UPLOAD_ONLY: 'Upload-Only',
    READ_ONLY: 'Read-Only',
    AUDITOR: 'Auditor',
    USER: 'User',
    POWER_USER: 'Power User',
    ADMINISTRATOR: 'Administrator',
} as const;

type RoleKeys = keyof typeof Roles;
type RoleValues = (typeof Roles)[RoleKeys];

const ETAC_ROLES: RoleValues[] = [Roles.READ_ONLY, Roles.USER];
const ADMIN_ROLES: RoleValues[] = [Roles.ADMINISTRATOR, Roles.POWER_USER];
const DEFAULT_USER_ROLE: RoleValues = Roles.READ_ONLY;

export const isETACRole = (roleId: number | undefined, roles?: Role[]): boolean => {
    const matchingRole = roles?.find((role) => roleId === role.id)?.name;
    return !!(matchingRole && ETAC_ROLES.includes(matchingRole as RoleValues));
};

export const isAdminRole = (roleId: number | undefined, roles?: Role[]): boolean => {
    const matchingRole = roles?.find((role) => roleId === role.id)?.name;
    return !!(matchingRole && ADMIN_ROLES.includes(matchingRole as RoleValues));
};

export const getDefaultRoleId = (roles?: Role[]): number | undefined => {
    return roles?.find((role) => DEFAULT_USER_ROLE === role.name)?.id;
};
