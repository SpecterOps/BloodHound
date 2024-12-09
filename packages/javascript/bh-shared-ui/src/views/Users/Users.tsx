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

import { Button } from '@bloodhoundenterprise/doodleui';
import { Box, Paper, Typography } from '@mui/material';
import { DateTime } from 'luxon';
import { useState } from 'react';
import { useMutation, useQuery } from 'react-query';
import {
    ConfirmationDialog,
    DataTable,
    DocumentationLinks,
    Header,
    PasswordDialog,
    UserTokenManagementDialog,
    Disable2FADialog,
    PageWithTitle,
    UpdateUserDialog,
    CreateUserDialog,
} from '../../components';
import { apiClient, LuxonFormat } from '../../utils';
import { CreateUserRequest, PutUserAuthSecretRequest, UpdateUserRequest, User } from 'js-client-library';
import find from 'lodash/find';
import { useToggle } from '../../hooks';
import UserActionsMenu from '../../components/UserActionsMenu';
import { useNotifications } from '../../providers';

const Users = () => {
    const { addNotification } = useNotifications();
    const [selectedUserId, setSelectedUserId] = useState<string | null>(null);
    const [createUserDialogOpen, toggleCreateUserDialog] = useToggle(false);
    const [updateUserDialogOpen, toggleUpdateUserDialog] = useToggle(false);
    const [disableUserDialogOpen, toggleDisableUserDialog] = useToggle(false);
    const [enableUserDialogOpen, toggleEnableUserDialog] = useToggle(false);
    const [deleteUserDialogOpen, toggleDeleteUserDialog] = useToggle(false);
    const [expireUserPasswordDialogOpen, toggleExpireUserPasswordDialog] = useToggle(false);
    const [resetUserPasswordDialogOpen, toggleResetUserPasswordDialog] = useToggle(false);
    const [needsPasswordReset, setNeedsPasswordReset] = useState(false);
    const [manageUserTokensDialogOpen, toggleManageUserTokensDialog] = useToggle(false);
    const [disable2FADialogOpen, setDisable2FADialogOpen] = useState(false);
    const [disable2FAError, setDisable2FAError] = useState('');
    const [disable2FASecret, setDisable2FASecret] = useState('');

    const getSelfQuery = useQuery(['getSelf'], ({ signal }) =>
        apiClient.getSelf({ signal }).then((res) => res.data?.data)
    );

    const hasSelectedSelf = getSelfQuery.data?.id === selectedUserId!;

    const listUsersQuery = useQuery(['listUsers'], ({ signal }) =>
        apiClient.listUsers({ signal }).then((res) => res.data?.data?.users)
    );

    const listSSOProvidersQuery = useQuery(['listSSOProviders'], ({ signal }) =>
        apiClient.listSSOProviders({ signal }).then((res) => res.data?.data)
    );

    const createUserMutation = useMutation((newUser: CreateUserRequest) => apiClient.createUser(newUser), {
        onSuccess: () => {
            addNotification('User created successfully!', 'createUserSuccess');
            listUsersQuery.refetch();
        },
    });

    const updateUserMutation = useMutation(
        (updatedUser: UpdateUserRequest) => apiClient.updateUser(selectedUserId!, updatedUser),
        {
            onSuccess: (response, updatedUser) => {
                addNotification('User updated successfully!', 'updateUserSuccess');
                const selectedUser = find(listUsersQuery.data, (user) => user.id === selectedUserId);
                // if the user previously had a SSO Provider ID but does not have one after the update then show the
                // password reset dialog with the "Force Password Reset?" input defaulted to checked
                if (selectedUser?.sso_provider_id !== null && !updatedUser.SSOProviderId) {
                    setNeedsPasswordReset(true);
                    toggleResetUserPasswordDialog();
                }
                listUsersQuery.refetch();
            },
        }
    );

    const disableEnableUserMutation = useMutation(
        async ({ userId, disable }: { userId: string; disable: boolean }) => {
            const user = listUsersQuery.data?.find((user: User) => {
                return user.id === userId;
            });

            if (!user) {
                return;
            }

            const updatedUser: UpdateUserRequest = {
                emailAddress: user.email_address || '',
                principal: user.principal_name || '',
                firstName: user.first_name || '',
                lastName: user.last_name || '',
                ...(user.sso_provider_id && { SSOProviderId: user.sso_provider_id }),
                roles: user.roles?.map((role: any) => role.id) || [],
                is_disabled: disable,
            };

            return apiClient.updateUser(selectedUserId!, updatedUser);
        },
        {
            onSuccess: (_, { disable }) => {
                addNotification(`User ${disable ? 'disabled' : 'enabled'} successfully!`, 'disableEnableUserSuccess');
                listUsersQuery.refetch();
            },
        }
    );

    const deleteUserMutation = useMutation((userId: string) => apiClient.deleteUser(userId), {
        onSuccess: () => {
            addNotification('User deleted successfully!', 'deleteUserSuccess');
            listUsersQuery.refetch();
        },
    });

    const expireUserPasswordMutation = useMutation((userId: string) => apiClient.expireUserAuthSecret(userId), {
        onSuccess: () => {
            addNotification('User password expired successfully!', 'expireUserPasswordSuccess');
        },
    });

    const updateUserPasswordMutation = useMutation(
        ({ userId, ...payload }: { userId: string } & PutUserAuthSecretRequest) =>
            apiClient.putUserAuthSecret(userId, payload),
        {
            onSuccess: () => {
                addNotification('User password updated successfully!', 'updateUserPasswordSuccess');
                toggleResetUserPasswordDialog();
            },
            onSettled: () => setNeedsPasswordReset(false),
            onError: (error: any) => {
                if (error.response?.status == 403) {
                    addNotification(
                        'Current password invalid. Password update failed.',
                        'UpdateUserPasswordCurrentPasswordInvalidError'
                    );
                } else {
                    addNotification('Password failed to update.', 'UpdateUserPasswordError');
                }
            },
        }
    );

    const SSOProvidersMap =
        listSSOProvidersQuery.data?.reduce((acc: any, val: any) => {
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
        if (user.sso_provider_id)
            return <span>{`SSO: ${SSOProvidersMap[user.sso_provider_id]?.name || user.sso_provider_id}`}</span>;
        if (user.AuthSecret?.totp_activated)
            return <span style={{ whiteSpace: 'pre-wrap' }}>{'Username / Password\nMFA Enabled'}</span>;
        return <span>Username / Password</span>;
    };

    const usersTableRows = listUsersQuery.data?.map((user: any, index: number) => [
        // This linting rule is disabled because the elements in this array do not require a key prop.
        /* eslint-disable react/jsx-key */
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
            showPasswordOptions={user.sso_provider_id === null || user.sso_provider_id === undefined}
            showAuthMgmtButtons={user.id !== getSelfQuery.data?.id}
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
        /* eslint-enable react/jsx-key */
    ]);

    return (
        <>
            <PageWithTitle
                title='Manage Users'
                data-testid='manage-users'
                pageDescription={
                    <Typography variant='body2' paragraph>
                        BloodHound offers multiple roles with degrees of permissions, providing greater security and
                        control of your team.
                        <br />
                        Learn more about {DocumentationLinks.ManageUsersDocLink}.
                    </Typography>
                }>
                <Box display='flex' justifyContent='flex-end' alignItems='center' minHeight='24px' mb={2}>
                    <Button
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
            </PageWithTitle>

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
                hasSelectedSelf={hasSelectedSelf}
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
                        disableEnableUserMutation.mutate({ userId: selectedUserId!, disable: false });
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
                        disableEnableUserMutation.mutate({ userId: selectedUserId!, disable: true });
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
                            addNotification('User MFA disabled successfully!', 'disableUserMfaSuccess');
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
                onClose={() => {
                    toggleResetUserPasswordDialog();
                    setNeedsPasswordReset(false);
                }}
                userId={selectedUserId!}
                onSave={updateUserPasswordMutation.mutate}
                requireCurrentPassword={hasSelectedSelf}
                showNeedsPasswordReset={true}
                initialNeedsPasswordReset={needsPasswordReset}
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
