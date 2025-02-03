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

import { faWarning } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { Paper, Tooltip, useTheme } from '@mui/material';
import { DateTime } from 'luxon';
import { FC } from 'react';
import { useQuery } from 'react-query';
import { DataTable, Header } from '../../components';
import { LuxonFormat, apiClient } from '../../utils';
import UserActionsMenu from './UserActionsMenu';

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

type UsersTableProps = {
    forbidden: boolean;
    onUpdateUser: (open: boolean) => void;
    onDisabledUser: (open: boolean) => void;
    onEnabledUser: (open: boolean) => void;
    onDeleteUser: (open: boolean) => void;
    onUpdateUserPassword: (open: boolean) => void;
    onExpiredUserPassword: (open: boolean) => void;
    onManageUserTokens: (open: boolean) => void;
    setDisable2FADialogOpen: (open: boolean) => void;
    setSelectedUserId: (userId: string | null) => void;
};

const UsersTable: FC<UsersTableProps> = ({
    forbidden,
    onUpdateUser,
    onDisabledUser,
    onEnabledUser,
    onDeleteUser,
    onUpdateUserPassword,
    onExpiredUserPassword,
    onManageUserTokens,
    setDisable2FADialogOpen,
    setSelectedUserId,
}) => {
    const theme = useTheme();

    const getSelfQuery = useQuery(['getSelf'], ({ signal }) =>
        apiClient.getSelf({ signal }).then((res) => res.data?.data)
    );

    const listUsersQuery = useQuery(
        ['listUsers'],
        ({ signal }) => apiClient.listUsers({ signal }).then((res) => res.data?.data?.users),
        { enabled: !forbidden }
    );

    const listSSOProvidersQuery = useQuery(
        ['listSSOProviders'],
        ({ signal }) => apiClient.listSSOProviders({ signal }).then((res) => res.data?.data),
        { enabled: !forbidden }
    );

    const SSOProvidersMap =
        listSSOProvidersQuery.data?.reduce((acc: any, val: any) => {
            acc[val.id] = val;
            return acc;
        }, {}) || {};

    const getAuthMethodText = (user: any): JSX.Element => {
        if (user.sso_provider_id)
            return <span>{`SSO: ${SSOProvidersMap[user.sso_provider_id]?.name || user.sso_provider_id}`}</span>;
        if (user.AuthSecret?.totp_activated)
            return <span style={{ whiteSpace: 'pre-wrap' }}>{'Username / Password\nMFA Enabled'}</span>;
        return <span>Username / Password</span>;
    };

    const usersTableRows = listUsersQuery.data?.map((user, index) => {
        const isNonUniqueEmail = !!listUsersQuery.data?.find(
            ({ email_address, id }) =>
                user.email_address?.toLowerCase() === email_address?.toLowerCase() && user.id !== id
        );

        return [
            // This linting rule is disabled because the elements in this array do not require a key prop.
            /* eslint-disable react/jsx-key */
            user.principal_name,
            <>
                {user.email_address}
                {isNonUniqueEmail ? (
                    <Tooltip
                        title='Duplicate email detected, unique user emails are required and will be enforced by the database in the following release.'
                        placement='top-start'>
                        <FontAwesomeIcon
                            icon={faWarning}
                            style={{ marginLeft: theme.spacing(1) }}
                            color={theme.palette.warning.main}
                        />
                    </Tooltip>
                ) : null}
            </>,
            `${user.first_name} ${user.last_name}`,
            <span style={{ whiteSpace: 'pre' }}>
                {DateTime.fromISO(user.created_at).toFormat(LuxonFormat.DATETIME)}
            </span>,
            user.roles?.[0]?.name,
            getUserStatusText(user),
            getAuthMethodText(user),
            <UserActionsMenu
                userId={user.id}
                onOpen={(_, userId) => {
                    setSelectedUserId(userId);
                }}
                showPasswordOptions={!user.sso_provider_id}
                showAuthMgmtButtons={user.id !== getSelfQuery.data?.id}
                showDisableMfaButton={user.AuthSecret?.totp_activated}
                userDisabled={user.is_disabled}
                onUpdateUser={onUpdateUser}
                onDisableUser={onDisabledUser}
                onEnableUser={onEnabledUser}
                onDeleteUser={onDeleteUser}
                onUpdateUserPassword={onUpdateUserPassword}
                onExpireUserPassword={onExpiredUserPassword}
                onManageUserTokens={onManageUserTokens}
                onDisableUserMfa={setDisable2FADialogOpen}
                index={index}
            />,
            /* eslint-enable react/jsx-key */
        ];
    });

    return (
        <Paper data-testid='manage-users_table'>
            <DataTable
                headers={usersTableHeaders}
                data={usersTableRows}
                isLoading={listUsersQuery.isLoading}
                showPaginationControls={false}
            />
        </Paper>
    );
};

export default UsersTable;
