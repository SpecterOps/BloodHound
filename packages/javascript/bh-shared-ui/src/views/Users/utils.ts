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
import { CreateUserRequest, UpdateUserRequest, User } from 'js-client-library';
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
    explore_enabled: user.explore_enabled || false,
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
