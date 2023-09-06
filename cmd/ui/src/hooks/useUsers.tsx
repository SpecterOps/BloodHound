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

import { RequestOptions } from 'js-client-library';
import { useMutation, useQuery, useQueryClient } from 'react-query';
import { apiClient } from 'bh-shared-ui';
import { addSnackbar } from 'src/ducks/global/actions';
import { useAppDispatch } from 'src/store';

export type User = {
    id: string;
    saml_provider_id: number | null;
    AuthSecret: any;
    Roles: Role[];
    first_name: string | null;
    last_name: string | null;
    email_address: string | null;
    principal_name: string;
    last_login: string;
};

export type Permission = {
    id: number;
    name: string;
    authority: string;
};

export type Role = {
    name: string;
    description: string;
    permissions: Permission[];
};

export type CreateUserRequest = {
    firstName: string;
    lastName: string;
    emailAddress: string;
    principal: string;
    roles: number[];
    SAMLProviderId?: string;
    password?: string;
    needsPasswordReset?: boolean;
};

export type UpdateUserRequest = {
    firstName: string;
    lastName: string;
    emailAddress: string;
    principal: string;
    roles: number[];
};

export type UpdateUserPasswordRequest = {
    secret: string;
    needs_password_reset: boolean;
};

export const userKeys = {
    all: ['users'] as const,
    detail: (userId: number) => [...userKeys.all, userId] as const,
};

export const getUsers = (options?: RequestOptions): Promise<User[]> =>
    apiClient.listUsers(options).then((res) => res.data.data.users);

export const createUser = ({ newUser }: { newUser: CreateUserRequest }, options?: RequestOptions) =>
    apiClient.createUser(newUser, options).then((res) => res.data.data);

export const updateUser = (
    { userId, updatedUser }: { userId: string; updatedUser: UpdateUserRequest },
    options?: RequestOptions
) => apiClient.updateUser(userId, updatedUser, options).then((res) => res.data.data);

export const deleteUser = ({ userId }: { userId: string }, options?: RequestOptions) =>
    apiClient.deleteUser(userId, options).then((res) => res.data);

export const expireUserPassword = ({ userId }: { userId: string }, options?: RequestOptions) =>
    apiClient.expireUserAuthSecret(userId, options).then((res) => res.data);

export const updateUserPassword = (
    {
        userId,
        updatedUserPassword,
    }: {
        userId: string;
        updatedUserPassword: UpdateUserPasswordRequest;
    },
    options?: RequestOptions
) => apiClient.putUserAuthSecret(userId, updatedUserPassword, options).then((res) => res.data);

export const useGetUsers = () => useQuery(userKeys.all, ({ signal }) => getUsers({ signal }));

export const useCreateUser = () => {
    const queryClient = useQueryClient();
    const dispatch = useAppDispatch();

    return useMutation(createUser, {
        onSuccess: () => {
            dispatch(addSnackbar('User created successfully!', 'createUserSuccess'));
            queryClient.invalidateQueries(userKeys.all);
        },
    });
};

export const useUpdateUser = () => {
    const queryClient = useQueryClient();
    const dispatch = useAppDispatch();

    return useMutation(updateUser, {
        onSuccess: () => {
            dispatch(addSnackbar('User updated successfully!', 'updateUserSuccess'));
            queryClient.invalidateQueries(userKeys.all);
        },
    });
};

export const useDeleteUser = () => {
    const queryClient = useQueryClient();
    const dispatch = useAppDispatch();

    return useMutation(deleteUser, {
        onSuccess: () => {
            dispatch(addSnackbar('User deleted successfully!', 'deleteUserSuccess'));
            queryClient.invalidateQueries(userKeys.all);
        },
    });
};

export const useExpireUserPassword = () => {
    const queryClient = useQueryClient();
    const dispatch = useAppDispatch();

    return useMutation(expireUserPassword, {
        onSuccess: () => {
            dispatch(addSnackbar('User password expired successfully!', 'expireUserPasswordSuccess'));
            queryClient.invalidateQueries(userKeys.all);
        },
    });
};

export const useUpdateUserPassword = () => {
    const queryClient = useQueryClient();
    const dispatch = useAppDispatch();

    return useMutation(updateUserPassword, {
        onSuccess: () => {
            dispatch(addSnackbar('User password updated successfully!', 'updateUserPasswordSuccess'));
            queryClient.invalidateQueries(userKeys.all);
        },
    });
};
