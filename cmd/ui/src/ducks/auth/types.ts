// Copyright 2023 Specter Ops, Inc.
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

export interface AuthState {
    isInitialized: boolean;
    loginLoading: boolean;
    loginError: any;
    updateExpiredPasswordLoading: boolean;
    updateExpiredPasswordError: any;
    sessionToken: string | null;
    user: getSelfResponse | null;
}

export interface getSelfResponse {
    id: string;
    principal_name: string;
    email_address: string;
    first_name: string;
    last_name: string;
    saml_provider_id: number | null;
    eula_accepted: boolean;
    last_login: string;
    AuthSecret: {
        id: number;
        digest_method: string;
        expires_at: string;
        totp_activated: boolean;
    };
    roles: {
        id: number;
        name: string;
        description: string;
        permissions: {
            id: number;
            authority: string;
            name: string;
        }[];
    }[];
}

export interface AuthToken {
    name: string;
    created_at: string;
    last_access: string;
    id: string;
    key: string;
}

export interface NewUserToken {
    token_name: string;
}

export interface Auditable {
    created_at: Date;
    updated_at: Date;
    deleted_at: {
        Time: Date;
        Valid: boolean;
    };
}

export interface NewUser {
    firstName: string;
    lastName: string;
    emailAddress: string;
    principal: string;
    roles: number[];
    SAMLProviderId?: string;
    password?: string;
    needsPasswordReset?: boolean;
}

export interface UpdatedUser {
    firstName: string;
    lastName: string;
    emailAddress: string;
    principal: string;
    roles: number[];
}

export interface SharpHoundUserType extends Auditable {
    AuthSecret: string;
    roles: SharpHoundRoleType[] | null;
    first_name: string;
    last_name: string;
    email_address: string;
    principal_name: string;
    last_login: Date;
    id: string;
}

export interface SharpHoundPermission extends Auditable {
    id: number;
    authority: string;
    name: string;
}

export interface SharpHoundRoleType extends Auditable {
    id: number;
    name: string;
    description: string;
    permissions: SharpHoundPermission[];
}

const START_FETCH_USERS = 'auth/STARTFETCHUSERS';
const SUCCESS_FETCH_USERS = 'auth/SUCCESSFETCHUSERS';
const START_FETCH_ROLES = 'auth/STARTFETCHROLES';
const SUCCESS_FETCH_ROLES = 'auth/SUCCESSFETCHROLES';
const START_CREATE_USER = 'auth/STARTCREATEUSER';
const START_UPDATE_USER = 'auth/STARTUPDATEUSER';
const START_DELETE_USER = 'auth/STARTDELETEUSER';
const START_SET_PWD = 'auth/STARTSETPWD';
const ERROR = 'auth/ERROR';

export {
    START_FETCH_USERS,
    SUCCESS_FETCH_USERS,
    START_FETCH_ROLES,
    SUCCESS_FETCH_ROLES,
    START_CREATE_USER,
    START_UPDATE_USER,
    START_DELETE_USER,
    START_SET_PWD,
    ERROR,
};

export interface StartFetchUsers {
    type: typeof START_FETCH_USERS;
}

export interface SuccessFetchUsers {
    type: typeof SUCCESS_FETCH_USERS;
    users: SharpHoundUserType[];
}

export interface StartFetchRoles {
    type: typeof START_FETCH_ROLES;
}

export interface SuccessFetchRoles {
    type: typeof SUCCESS_FETCH_ROLES;
    roles: SharpHoundRoleType[];
}

export interface StartCreateUser {
    type: typeof START_CREATE_USER;
    user: NewUser;
}

export interface StartUpdateUser {
    type: typeof START_UPDATE_USER;
    userId: string;
    user: UpdatedUser;
}

export interface StartDeleteUser {
    type: typeof START_DELETE_USER;
    userId: string;
}

export interface StartSetPwd {
    type: typeof START_SET_PWD;
    user: SharpHoundUserType;
    secret: string;
    needsPasswordReset: boolean;
}

export interface Error {
    type: typeof ERROR;
}

export type AuthActions =
    | StartFetchUsers
    | SuccessFetchUsers
    | StartFetchRoles
    | SuccessFetchRoles
    | StartCreateUser
    | StartUpdateUser
    | StartDeleteUser
    | StartSetPwd
    | Error;

export const baseUser: SharpHoundUserType = {
    id: '',
    roles: null,
    first_name: '',
    last_name: '',
    email_address: '',
    principal_name: '',
    AuthSecret: '',
    last_login: new Date(),
    created_at: new Date(),
    updated_at: new Date(),
    deleted_at: {
        Time: new Date(),
        Valid: true,
    },
};
