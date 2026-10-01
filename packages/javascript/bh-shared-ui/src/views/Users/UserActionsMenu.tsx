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

import {
    faBars,
    faCogs,
    faEdit,
    faKey,
    faLock,
    faTrash,
    faUnlockAlt,
    faUserCheck,
    faUserLock,
} from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { IconButton } from '@mui/material';
import { Menu, MenuContent, MenuItem, MenuTrigger } from 'doodle-ui';
import React from 'react';
import { useAPITokensConfiguration, usePermissions } from '../../hooks';
import { Permission } from '../../utils';

interface UserActionsMenuProps {
    userId: string;
    onOpen: (e: any, userId: string) => any;
    showPasswordOptions: boolean;
    showAuthMgmtButtons: boolean;
    showDisableMfaButton: boolean;
    userDisabled: boolean;
    onUpdateUser: (e: any) => any;
    onDisableUser: (e: any) => any;
    onEnableUser: (e: any) => any;
    onDeleteUser: (e: any) => any;
    onUpdateUserPassword: (e: any) => any;
    onExpireUserPassword: (e: any) => any;
    onManageUserTokens: (e: any) => any;
    onDisableUserMfa: (e: any) => any;
    index: number;
}

const UserActionsMenu: React.FC<UserActionsMenuProps> = ({
    userId,
    onOpen,
    showPasswordOptions,
    showAuthMgmtButtons,
    showDisableMfaButton,
    userDisabled,
    onUpdateUser,
    onDisableUser,
    onEnableUser,
    onDeleteUser,
    onUpdateUserPassword,
    onExpireUserPassword,
    onManageUserTokens,
    onDisableUserMfa,
    index,
}) => {
    /* Hooks */

    const apiTokensEnabled = useAPITokensConfiguration();

    /* Event Handlers */

    const getAbleUserComponent = (): JSX.Element | null => {
        if (!showAuthMgmtButtons) return null;
        if (userDisabled)
            return (
                <MenuItem onSelect={onEnableUser} iconLeft icon={<FontAwesomeIcon icon={faUserCheck} />}>
                    Enable User
                </MenuItem>
            );
        else {
            return (
                <MenuItem onSelect={onDisableUser} iconLeft icon={<FontAwesomeIcon icon={faUserLock} />}>
                    Disable User
                </MenuItem>
            );
        }
    };

    const { checkPermission } = usePermissions();
    const hasManagePermission = checkPermission(Permission.AUTH_MANAGE_USERS);

    /* Implementation */

    return (
        <Menu>
            <MenuTrigger asChild>
                <IconButton
                    disabled={!hasManagePermission}
                    data-testid='manage-users_user-row-action-menu-button'
                    aria-label='Show user actions'
                    onPointerDown={(event) => {
                        // Radix opens on pointer-down, before a click can select the row's user.
                        if (event.button === 0 && !event.ctrlKey) onOpen(event, userId);
                    }}
                    onKeyDown={(event) => {
                        if (['Enter', ' ', 'ArrowDown'].includes(event.key)) onOpen(event, userId);
                    }}
                    size='large'>
                    <FontAwesomeIcon icon={faBars} />
                </IconButton>
            </MenuTrigger>
            <MenuContent
                className='max-h-[var(--radix-dropdown-menu-content-available-height)] overflow-y-auto'
                align='end'
                data-testid={`manage-users_user-row-action-menu-${index}`}>
                <MenuItem
                    data-testid='manage-users_user-row-action-menu-update-user-button'
                    onSelect={onUpdateUser}
                    iconLeft
                    icon={<FontAwesomeIcon icon={faEdit} />}>
                    Update User
                </MenuItem>

                {showPasswordOptions && (
                    <MenuItem onSelect={onUpdateUserPassword} iconLeft icon={<FontAwesomeIcon icon={faKey} />}>
                        Change Password
                    </MenuItem>
                )}

                {showPasswordOptions && showAuthMgmtButtons && (
                    <MenuItem onSelect={onExpireUserPassword} iconLeft icon={<FontAwesomeIcon icon={faLock} />}>
                        Force Password Reset
                    </MenuItem>
                )}

                {apiTokensEnabled && (
                    <MenuItem onSelect={onManageUserTokens} iconLeft icon={<FontAwesomeIcon icon={faCogs} />}>
                        Generate / Revoke API Tokens
                    </MenuItem>
                )}
                {showDisableMfaButton && (
                    <MenuItem onSelect={onDisableUserMfa} iconLeft icon={<FontAwesomeIcon icon={faUnlockAlt} />}>
                        Disable MFA
                    </MenuItem>
                )}

                {showAuthMgmtButtons && getAbleUserComponent()}

                {showAuthMgmtButtons && (
                    <MenuItem onSelect={onDeleteUser} iconLeft icon={<FontAwesomeIcon icon={faTrash} />}>
                        Delete User
                    </MenuItem>
                )}
            </MenuContent>
        </Menu>
    );
};

export default UserActionsMenu;
