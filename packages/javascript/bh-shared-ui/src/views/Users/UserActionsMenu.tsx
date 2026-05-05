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
import { Menu, MenuContent, MenuItem, MenuTrigger } from 'doodle-ui';
import React from 'react';
import { useAPITokensConfiguration } from '../../hooks';

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

    const getAbleUserComponent = (): JSX.Element | null => {
        if (!showAuthMgmtButtons) return null;
        if (userDisabled)
            return (
                <MenuItem onSelect={() => onEnableUser(null)} className='flex gap-2 items-center'>
                    <FontAwesomeIcon icon={faUserCheck} />
                    <span>Enable User</span>
                </MenuItem>
            );
        else {
            return (
                <MenuItem onSelect={() => onDisableUser(null)} className='flex gap-2 items-center'>
                    <FontAwesomeIcon icon={faUserLock} />
                    <span>Disable User</span>
                </MenuItem>
            );
        }
    };

    /* Implementation */

    return (
        <Menu>
            <MenuTrigger asChild>
                <button
                    data-testid='manage-users_user-row-action-menu-button'
                    aria-label='Show user actions'
                    onClick={(e) => onOpen(e, userId)}
                    className='p-2 rounded hover:bg-neutral-light-3 dark:hover:bg-neutral-dark-3'>
                    <FontAwesomeIcon icon={faBars} />
                </button>
            </MenuTrigger>
            <MenuContent align='end' data-testid={`manage-users_user-row-action-menu-${index}`}>
                <MenuItem
                    data-testid='manage-users_user-row-action-menu-update-user-button'
                    onSelect={() => onUpdateUser(null)}
                    className='flex gap-2 items-center'>
                    <FontAwesomeIcon icon={faEdit} />
                    <span>Update User</span>
                </MenuItem>

                {showPasswordOptions && (
                    <MenuItem onSelect={() => onUpdateUserPassword(null)} className='flex gap-2 items-center'>
                        <FontAwesomeIcon icon={faKey} />
                        <span>Change Password</span>
                    </MenuItem>
                )}

                {showPasswordOptions && showAuthMgmtButtons && (
                    <MenuItem onSelect={() => onExpireUserPassword(null)} className='flex gap-2 items-center'>
                        <FontAwesomeIcon icon={faLock} />
                        <span>Force Password Reset</span>
                    </MenuItem>
                )}

                {apiTokensEnabled && (
                    <MenuItem onSelect={() => onManageUserTokens(null)} className='flex gap-2 items-center'>
                        <FontAwesomeIcon icon={faCogs} />
                        <span>Generate / Revoke API Tokens</span>
                    </MenuItem>
                )}

                {showDisableMfaButton && (
                    <MenuItem onSelect={() => onDisableUserMfa(null)} className='flex gap-2 items-center'>
                        <FontAwesomeIcon icon={faUnlockAlt} />
                        <span>Disable MFA</span>
                    </MenuItem>
                )}

                {showAuthMgmtButtons && getAbleUserComponent()}

                {showAuthMgmtButtons && (
                    <MenuItem onSelect={() => onDeleteUser(null)} className='flex gap-2 items-center'>
                        <FontAwesomeIcon icon={faTrash} />
                        <span>Delete User</span>
                    </MenuItem>
                )}
            </MenuContent>
        </Menu>
    );
};

export default UserActionsMenu;
