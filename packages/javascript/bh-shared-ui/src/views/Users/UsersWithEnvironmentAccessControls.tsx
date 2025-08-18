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

import { Button, Dialog, DialogPortal, DialogTrigger } from '@bloodhoundenterprise/doodleui';
import { Box, Paper, Typography } from '@mui/material';
import { CreateUserRequest, PutUserAuthSecretRequest, UpdateUserRequest, User } from 'js-client-library';
import find from 'lodash/find';
import { FC, useState } from 'react';
import { useMutation, useQuery } from 'react-query';
import {
    ConfirmationDialog,
    CreateUserDialog,
    Disable2FADialog,
    DocumentationLinks,
    PageWithTitle,
    PasswordDialog,
    UserTokenManagementDialog,
} from '../../components';
import { useMountEffect, usePermissions, useToggle } from '../../hooks';
import { useNotifications } from '../../providers';
import { Permission, apiClient } from '../../utils';
import UsersTable from './UsersTable';

const UsersWithEnvironmentAccessControls: FC = () => {
    const [selectedUserId, setSelectedUserId] = useState<string | null>(null);
    const [disable2FADialogOpen, setDisable2FADialogOpen] = useState(false);
    const [disable2FAError, setDisable2FAError] = useState('');
    const [disable2FASecret, setDisable2FASecret] = useState('');
    const [needsPasswordReset, setNeedsPasswordReset] = useState(false);

    const [createUserDialogOpen, toggleCreateUserDialog] = useToggle(false);
    const [updateUserDialogOpen, toggleUpdateUserDialog] = useToggle(false);
    const [disableUserDialogOpen, toggleDisableUserDialog] = useToggle(false);
    const [enableUserDialogOpen, toggleEnableUserDialog] = useToggle(false);
    const [deleteUserDialogOpen, toggleDeleteUserDialog] = useToggle(false);
    const [expireUserPasswordDialogOpen, toggleExpireUserPasswordDialog] = useToggle(false);
    const [resetUserPasswordDialogOpen, toggleResetUserPasswordDialog] = useToggle(false);
    const [manageUserTokensDialogOpen, toggleManageUserTokensDialog] = useToggle(false);

    const { checkPermission } = usePermissions();
    const hasPermission = checkPermission(Permission.AUTH_MANAGE_USERS);

    const { addNotification, dismissNotification } = useNotifications();
    const notificationKey = 'manage-users-permission';

    const effect: React.EffectCallback = () => {
        if (!hasPermission) {
            addNotification(
                `Your user role does not grant permission for managing users. Please contact your administrator for details.`,
                notificationKey,
                {
                    persist: true,
                    anchorOrigin: { vertical: 'top', horizontal: 'right' },
                }
            );
        }

        return () => dismissNotification(notificationKey);
    };

    useMountEffect(effect);

    const getSelfQuery = useQuery(['getSelf'], ({ signal }) =>
        apiClient.getSelf({ signal }).then((res) => res.data?.data)
    );

    const hasSelectedSelf = getSelfQuery.data?.id === selectedUserId!;

    const listUsersQuery = useQuery(
        ['listUsers'],
        ({ signal }) => apiClient.listUsers({ signal }).then((res) => res.data?.data?.users),
        { enabled: hasPermission }
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
                    {/* TODO: IMPLEMENT FEATURE FLAG TO DISPLAY IF ON */}
                    {/*
                    <FeatureFlag
                    flagKey='PUT_ETAC_FEATURE_FLAG_HERE'
                        enabled={
                        }
                        disabled={
                        }
                    />
                    */}
                    <Dialog open={createUserDialogOpen} onOpenChange={toggleCreateUserDialog}>
                        <DialogTrigger asChild>
                            <Button
                                disabled={!hasPermission}
                                onClick={() => {
                                    setSelectedUserId(null);
                                    toggleCreateUserDialog();
                                }}
                                data-testid='manage-users_button-create-user'>
                                Create User
                            </Button>
                        </DialogTrigger>
                        <DialogPortal>
                            <CreateUserDialog
                                createUser={true}
                                updateUser={false}
                                error={createUserMutation.error}
                                isLoading={createUserMutation.isLoading}
                                onClose={toggleCreateUserDialog}
                                onExited={createUserMutation.reset}
                                onSave={createUserMutation.mutateAsync}
                                open={createUserDialogOpen}
                                showEnvironmentAccessControls={true}
                            />
                        </DialogPortal>
                    </Dialog>
                </Box>
                <Paper data-testid='manage-users_table'>
                    <UsersTable
                        onDeleteUser={toggleDeleteUserDialog}
                        onDisabledUser={toggleDisableUserDialog}
                        onEnabledUser={toggleEnableUserDialog}
                        onExpiredUserPassword={toggleExpireUserPasswordDialog}
                        onManageUserTokens={toggleManageUserTokensDialog}
                        onUpdateUser={toggleUpdateUserDialog}
                        onUpdateUserPassword={toggleResetUserPasswordDialog}
                        setDisable2FADialogOpen={setDisable2FADialogOpen}
                        setSelectedUserId={(id) => setSelectedUserId(id)}
                    />
                </Paper>
            </PageWithTitle>

            <Dialog open={updateUserDialogOpen} onOpenChange={toggleUpdateUserDialog}>
                <DialogPortal>
                    <CreateUserDialog
                        createUser={false}
                        updateUser={true}
                        error={updateUserMutation.error}
                        isLoading={updateUserMutation.isLoading}
                        onClose={toggleUpdateUserDialog}
                        onExited={updateUserMutation.reset}
                        onSave={updateUserMutation.mutateAsync}
                        open={updateUserDialogOpen}
                        showEnvironmentAccessControls={true}
                        userId={selectedUserId!}
                        hasSelectedSelf={hasSelectedSelf}
                    />
                </DialogPortal>
            </Dialog>
            {/*
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
            */}
            <ConfirmationDialog
                open={enableUserDialogOpen}
                text={'Are you sure you want to enable this user?'}
                title={'Enable User'}
                onConfirm={() => {
                    disableEnableUserMutation.mutate({
                        userId: selectedUserId!,
                        disable: false,
                    });
                    toggleEnableUserDialog();
                }}
                onCancel={toggleEnableUserDialog}
            />
            <ConfirmationDialog
                open={disableUserDialogOpen}
                text={'Are you sure you want to disable this user?'}
                title={'Disable User'}
                onConfirm={() => {
                    disableEnableUserMutation.mutate({
                        userId: selectedUserId!,
                        disable: true,
                    });
                    toggleDisableUserDialog();
                }}
                onCancel={toggleDisableUserDialog}
            />
            <ConfirmationDialog
                open={deleteUserDialogOpen}
                text={'Are you sure you want to delete this user?'}
                title={'Delete User'}
                onConfirm={() => {
                    deleteUserMutation.mutate(selectedUserId!);
                    toggleDeleteUserDialog();
                }}
                onCancel={toggleDeleteUserDialog}
            />
            <ConfirmationDialog
                open={expireUserPasswordDialogOpen}
                text={
                    "Are you sure you want to expire this user's password? This user will be prompted to change their password on next login."
                }
                title={'Force Password Reset'}
                onConfirm={() => {
                    expireUserPasswordMutation.mutate(selectedUserId!);
                    toggleExpireUserPasswordDialog();
                }}
                onCancel={toggleExpireUserPasswordDialog}
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

export default UsersWithEnvironmentAccessControls;
