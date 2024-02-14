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

import { Box, Button, Paper } from '@mui/material';
import { DateTime } from 'luxon';
import { useState } from 'react';
import { useMutation, useQuery } from 'react-query';

import {
    ConfirmationDialog,
    DataTable,
    Header,
    ContentPage,
    PasswordDialog,
    LuxonFormat,
    UserTokenManagementDialog,
    apiClient,
    Disable2FADialog,
} from 'bh-shared-ui';
import { NewUser, UpdatedUser } from 'src/ducks/auth/types';
import { addSnackbar } from 'src/ducks/global/actions';
import useToggle from 'src/hooks/useToggle';
import { User } from 'src/hooks/useUsers';
import { useAppDispatch, useAppSelector } from 'src/store';
import CreateUserDialog from 'src/views/Users/CreateUserDialog';
import UpdateUserDialog from 'src/views/Users/UpdateUserDialog';
import UserActionsMenu from 'src/views/Users/UserActionsMenu';

const Users = () => {
    const dispatch = useAppDispatch();
    const [selectedUserId, setSelectedUserId] = useState<string | null>(null);
    const [createUserDialogOpen, toggleCreateUserDialog] = useToggle(false);
    const [updateUserDialogOpen, toggleUpdateUserDialog] = useToggle(false);
    const [disableUserDialogOpen, toggleDisableUserDialog] = useToggle(false);
    const [enableUserDialogOpen, toggleEnableUserDialog] = useToggle(false);
    const [deleteUserDialogOpen, toggleDeleteUserDialog] = useToggle(false);
    const [expireUserPasswordDialogOpen, toggleExpireUserPasswordDialog] = useToggle(false);
    const [resetUserPasswordDialogOpen, toggleResetUserPasswordDialog] = useToggle(false);
    const [manageUserTokensDialogOpen, toggleManageUserTokensDialog] = useToggle(false);
    const [disable2FADialogOpen, setDisable2FADialogOpen] = useState(false);
    const [disable2FAError, setDisable2FAError] = useState('');
    const [disable2FASecret, setDisable2FASecret] = useState('');

    const self = useAppSelector((state) => state.auth.user);

    const getSelfQuery = useQuery(['getSelf'], ({ signal }) =>
        apiClient.getSelf({ signal }).then((res) => res.data.data)
    );

    const listUsersQuery = useQuery(['listUsers'], ({ signal }) =>
        apiClient.listUsers({ signal }).then((res) => res.data.data.users)
    );

    const listSAMLProvidersQuery = useQuery(['listSAMLProviders'], ({ signal }) =>
        apiClient.listSAMLProviders({ signal }).then((res) => res.data.data.saml_providers)
    );

    const createUserMutation = useMutation((newUser: NewUser) => apiClient.createUser(newUser), {
        onSuccess: () => {
            dispatch(addSnackbar('User created successfully!', 'createUserSuccess'));
            listUsersQuery.refetch();
        },
    });

    const updateUserMutation = useMutation(
        (updatedUser: UpdatedUser) => apiClient.updateUser(selectedUserId!, updatedUser),
        {
            onSuccess: () => {
                dispatch(addSnackbar('User updated successfully!', 'updateUserSuccess'));
                listUsersQuery.refetch();
            },
        }
    );

    const disableUserMutation = useMutation(
        async (userId: string) => {
            const user = listUsersQuery.data.find((user: User) => {
                return user.id === userId;
            });

            const updatedUser = {
                emailAddress: user.email_address || '',
                principal: user.principal_name || '',
                firstName: user.first_name || '',
                lastName: user.last_name || '',
                SAMLProviderId: user.saml_provider_id?.toString() || '',
                roles: user.roles?.map((role: any) => role.id) || [],
                is_disabled: true,
            };
            return apiClient.updateUser(selectedUserId!, updatedUser);
        },
        {
            onSuccess: () => {
                dispatch(addSnackbar('User disabled successfully!', 'disableUserSuccess'));
                listUsersQuery.refetch();
            },
        }
    );

    const enableUserMutation = useMutation(
        async (userId: string) => {
            const user = listUsersQuery.data.find((user: User) => {
                return user.id === userId;
            });

            const updatedUser = {
                emailAddress: user.email_address || '',
                principal: user.principal_name || '',
                firstName: user.first_name || '',
                lastName: user.last_name || '',
                SAMLProviderId: user.saml_provider_id?.toString() || '',
                roles: user.roles?.map((role: any) => role.id) || [],
                is_disabled: false,
            };
            return apiClient.updateUser(selectedUserId!, updatedUser);
        },
        {
            onSuccess: () => {
                dispatch(addSnackbar('User enabled successfully!', 'enableUserSuccess'));
                listUsersQuery.refetch();
            },
        }
    );

    const deleteUserMutation = useMutation((userId: string) => apiClient.deleteUser(userId), {
        onSuccess: () => {
            dispatch(addSnackbar('User deleted successfully!', 'deleteUserSuccess'));
            listUsersQuery.refetch();
        },
    });

    const expireUserPasswordMutation = useMutation((userId: string) => apiClient.expireUserAuthSecret(userId), {
        onSuccess: () => {
            dispatch(addSnackbar('User password expired successfully!', 'expireUserPasswordSuccess'));
        },
    });

    const updateUserPasswordMutation = useMutation(
        ({ userId, secret, needsPasswordReset }: { userId: string; secret: string; needsPasswordReset: boolean }) =>
            apiClient.putUserAuthSecret(userId, {
                needs_password_reset: needsPasswordReset,
                secret: secret,
            }),
        {
            onSuccess: () => {
                dispatch(addSnackbar('User password updated successfully!', 'updateUserPasswordSuccess'));
                toggleResetUserPasswordDialog();
            },
        }
    );

    const SAMLProvidersMap =
        listSAMLProvidersQuery.data?.reduce((acc: any, val: any) => {
            acc[val.id] = val;
            return acc;
        }, {}) || {};

    const usersTableHeaders: Header[] = [
        { label: 'Username' },
        { label: 'Email' },
        { label: 'Name' },
        { label: 'Created' },
        { label: 'Role' },
        { label: 'Status' },
        { label: 'Auth Method' },
        { label: '', alignment: 'right' },
    ];

    type UserStatus = 'Deleted' | 'Disabled' | 'Active';

    const getUserStatusText = (user: any): UserStatus => {
        if (user.deleted_at.Valid) {
            return 'Deleted';
        } else if (user.is_disabled) {
            return 'Disabled';
        } else return 'Active';
    };

    const getAuthMethodText = (user: any): JSX.Element => {
        if (user.saml_provider_id)
            return <span>{`SAML: ${SAMLProvidersMap[user.saml_provider_id]?.name || user.saml_provider_id}`}</span>;
        if (user.AuthSecret?.totp_activated)
            return <span style={{ whiteSpace: 'pre-wrap' }}>{'Username / Password\nMFA Enabled'}</span>;
        return <span>Username / Password</span>;
    };

    const usersTableRows = listUsersQuery.data?.map((user: any, index: number) => [
        user.principal_name,
        user.email_address,
        `${user.first_name} ${user.last_name}`,
        <span style={{ whiteSpace: 'pre' }}>{DateTime.fromISO(user.created_at).toFormat(LuxonFormat.DATETIME)}</span>,
        user.roles?.[0]?.name,
        getUserStatusText(user),
        getAuthMethodText(user),
        <UserActionsMenu
            userId={user.id}
            onOpen={(e, userId) => {
                setSelectedUserId(userId);
            }}
            showPasswordOptions={user.saml_provider_id === null || user.saml_provider_id === undefined}
            showAuthMgmtButtons={user.id !== self?.id}
            showDisableMfaButton={user.AuthSecret?.totp_activated}
            userDisabled={user.is_disabled}
            onUpdateUser={toggleUpdateUserDialog}
            onDisableUser={toggleDisableUserDialog}
            onEnableUser={toggleEnableUserDialog}
            onDeleteUser={toggleDeleteUserDialog}
            onUpdateUserPassword={toggleResetUserPasswordDialog}
            onExpireUserPassword={toggleExpireUserPasswordDialog}
            onManageUserTokens={toggleManageUserTokensDialog}
            onDisableUserMfa={setDisable2FADialogOpen}
            index={index}
        />,
    ]);

    return (
        <>
            <ContentPage title='Manage Users' data-testid='manage-users'>
                <Box display='flex' justifyContent='flex-end' alignItems='center' minHeight='24px' mb={2}>
                    <Button
                        color='primary'
                        variant='contained'
                        disableElevation
                        onClick={() => {
                            setSelectedUserId(null);
                            toggleCreateUserDialog();
                        }}
                        data-testid='manage-users_button-create-user'>
                        Create User
                    </Button>
                </Box>
                <Paper data-testid='manage-users_table'>
                    <DataTable
                        headers={usersTableHeaders}
                        data={usersTableRows}
                        isLoading={listUsersQuery.isLoading}
                        showPaginationControls={false}
                    />
                </Paper>
            </ContentPage>

            <CreateUserDialog
                open={createUserDialogOpen}
                onClose={toggleCreateUserDialog}
                onExited={createUserMutation.reset}
                onSave={createUserMutation.mutateAsync}
                isLoading={createUserMutation.isLoading}
                error={createUserMutation.error}
            />
            <UpdateUserDialog
                open={updateUserDialogOpen}
                onClose={toggleUpdateUserDialog}
                onExited={updateUserMutation.reset}
                userId={selectedUserId!}
                onSave={updateUserMutation.mutateAsync}
                isLoading={updateUserMutation.isLoading}
                error={updateUserMutation.error}
            />
            <ConfirmationDialog
                open={enableUserDialogOpen}
                text={'Are you sure you want to enable this user?'}
                title={'Enable User'}
                onClose={(response) => {
                    if (response) {
                        enableUserMutation.mutate(selectedUserId!);
                    }
                    toggleEnableUserDialog();
                }}
            />
            <ConfirmationDialog
                open={disableUserDialogOpen}
                text={'Are you sure you want to disable this user?'}
                title={'Disable User'}
                onClose={(response) => {
                    if (response) {
                        disableUserMutation.mutate(selectedUserId!);
                    }
                    toggleDisableUserDialog();
                }}
            />
            <ConfirmationDialog
                open={deleteUserDialogOpen}
                text={'Are you sure you want to delete this user?'}
                title={'Delete User'}
                onClose={(response) => {
                    if (response) {
                        deleteUserMutation.mutate(selectedUserId!);
                    }
                    toggleDeleteUserDialog();
                }}
            />
            <ConfirmationDialog
                open={expireUserPasswordDialogOpen}
                text={
                    "Are you sure you want to expire this user's password? This user will be prompted to change their password on next login."
                }
                title={'Force Password Reset'}
                onClose={(response) => {
                    if (response) {
                        expireUserPasswordMutation.mutate(selectedUserId!);
                    }
                    toggleExpireUserPasswordDialog();
                }}
            />
            <Disable2FADialog
                open={disable2FADialogOpen}
                onClose={() => {
                    setDisable2FADialogOpen(false);
                    setDisable2FAError('');
                    setDisable2FASecret('');
                    getSelfQuery.refetch();
                }}
                onCancel={() => {
                    setDisable2FADialogOpen(false);
                    setDisable2FAError('');
                    setDisable2FASecret('');
                    getSelfQuery.refetch();
                }}
                onSave={(secret: string) => {
                    setDisable2FAError('');
                    apiClient
                        .disenrollMFA(selectedUserId!, { secret })
                        .then(() => {
                            setDisable2FADialogOpen(false);
                            dispatch(addSnackbar('User MFA disabled successfully!', 'disableUserMfaSuccess'));
                            setDisable2FASecret('');
                            listUsersQuery.refetch();
                        })
                        .catch(() => {
                            setDisable2FAError('Unable to verify password. Please try again.');
                        });
                }}
                error={disable2FAError}
                secret={disable2FASecret}
                onSecretChange={(e: any) => setDisable2FASecret(e.target.value)}
                contentText='Are you sure you want to disable MFA for this user? Please enter your password to confirm.'
            />
            <PasswordDialog
                open={resetUserPasswordDialogOpen}
                onClose={toggleResetUserPasswordDialog}
                userId={selectedUserId!}
                onSave={updateUserPasswordMutation.mutate}
                showNeedsPasswordReset={true}
            />
            <UserTokenManagementDialog
                open={manageUserTokensDialogOpen}
                onClose={toggleManageUserTokensDialog}
                userId={selectedUserId!}
            />
        </>
    );
};

export default Users;
