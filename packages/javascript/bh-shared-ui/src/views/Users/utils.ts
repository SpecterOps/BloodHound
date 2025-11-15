import { CreateUserRequest, Role, UpdateUserRequest, User } from 'js-client-library';
import { CreateUserRequestForm, UpdateUserRequestForm } from '../..';

type UserRequestForm = UpdateUserRequestForm | CreateUserRequestForm;
type UserRequest = UpdateUserRequest | CreateUserRequest;

// Helper function for converting our form field values in the Create/Update User forms into user request objects.
// Contains extra logic for conditionally adding ETAC fields
export const mapFormFieldsToUserRequest = (
    formFields: UserRequestForm,
    authenticationMethod: string,
    isAdmin: boolean,
    isETAC: boolean
): UserRequest => {
    // Pull ETAC field out so we can format and apply it separately
    const { environment_targeted_access_control, ...filteredValues } = formFields;

    const parsedSSO = formFields.sso_provider_id ? parseInt(formFields.sso_provider_id, 10) : undefined;
    const parsedRoles = formFields.roles ? [formFields.roles] : [];

    // The all_environments field has different rules based on the user's role
    const parsedAllEnvironments = !!(isAdmin || (isETAC && formFields.all_environments));

    const userRequest = {
        ...filteredValues,
        sso_provider_id: authenticationMethod === 'password' ? undefined : parsedSSO,
        roles: parsedRoles,
        all_environments: parsedAllEnvironments,
    };

    if (isETAC) {
        // Add the environments list only for users with ETAC roles
        return {
            ...userRequest,
            environment_targeted_access_control: {
                environments:
                    formFields.all_environments === false
                        ? formFields.environment_targeted_access_control?.environments
                        : null,
            },
        };
    } else {
        return userRequest;
    }
};

// The API responses for Users have a different shape than the expected requests for creating/updating those users
export const mapUserResponseToRequest = (user: User): UserRequest => ({
    email_address: user.email_address || '',
    principal: user.principal_name || '',
    first_name: user.first_name || '',
    last_name: user.last_name || '',
    sso_provider_id: user.sso_provider_id || undefined,
    roles: user.roles?.map((role) => role.id) || [],
    ...(Object.hasOwn(user, 'all_environments') && { all_environments: user.all_environments }),
    ...(Object.hasOwn(user, 'environment_targeted_access_control') && {
        environment_targeted_access_control: {
            environments: user.environment_targeted_access_control || null,
        },
    }),
});

// Extra helper that uses the above function and additionally converts to the required form field types
export const mapUserResponseToForm = (user: User): UserRequestForm => {
    const request = mapUserResponseToRequest(user);
    const roles = request.roles.length ? request.roles[0] : undefined;

    return {
        ...request,
        roles,
        sso_provider_id: user.sso_provider_id?.toString() || undefined,
    };
};

const ETAC_ROLES = ['Read-Only', 'User'];
const ADMIN_ROLES = ['Administrator', 'Power User'];
const DEFAULT_ROLE = 'Read-Only';

export const isETACRole = (roleId: number | undefined, roles?: Role[]): boolean => {
    const matchingRole = roleId && roles?.find((role) => roleId === role.id)?.name;
    return !!(matchingRole && ETAC_ROLES.includes(matchingRole));
};

export const isAdminRole = (roleId: number | undefined, roles?: Role[]): boolean => {
    const matchingRole = roles?.find((role) => roleId === role.id)?.name;
    return !!(matchingRole && ADMIN_ROLES.includes(matchingRole));
};

export const getDefaultRoleId = (roles?: Role[]): number | undefined => {
    return roles?.find((role) => DEFAULT_ROLE === role.name)?.id;
};
