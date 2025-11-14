import { CreateUserRequest, UpdateUserRequest } from 'js-client-library';
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

    // The all_environments field has different rules based on the user's role
    const parsedAllEnvironments = !!(isAdmin || (isETAC && formFields.all_environments));

    const userRequest = {
        ...filteredValues,
        sso_provider_id: authenticationMethod === 'password' ? undefined : parsedSSO,
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
