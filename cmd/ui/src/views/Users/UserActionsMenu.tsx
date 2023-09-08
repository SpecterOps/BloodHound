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
    faLock,
    faUserCheck,
    faUserLock,
    faKey,
    faTrash,
    faUnlockAlt,
} from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { IconButton, ListItemIcon, ListItemText, Menu, MenuItem, MenuProps } from '@mui/material';
import withStyles from '@mui/styles/withStyles';
import React from 'react';

const StyledMenu = withStyles({
    paper: {
        border: '1px solid #d3d4d5',
    },
})((props: MenuProps) => (
    <Menu
        elevation={0}
        anchorOrigin={{
            vertical: 'bottom',
            horizontal: 'right',
        }}
        transformOrigin={{
            vertical: 'top',
            horizontal: 'right',
        }}
        {...props}
    />
));

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

    const [anchorEl, setAnchorEl] = React.useState<HTMLElement | null>(null);

    /* Event Handlers */

    const handleOnOpen: React.MouseEventHandler<HTMLButtonElement> = (event) => {
        setAnchorEl(event.currentTarget);
        onOpen(event, userId);
    };

    const getAbleUserComponent = (): JSX.Element | null => {
        if (!showAuthMgmtButtons) return null;
        if (userDisabled)
            return (
                <MenuItem
                    onClick={(e: React.MouseEvent<HTMLLIElement>) => {
                        onEnableUser(e);
                        setAnchorEl(null);
                    }}>
                    <ListItemIcon>
                        <FontAwesomeIcon icon={faUserCheck} />
                    </ListItemIcon>
                    <ListItemText primary={'Enable User'} />
                </MenuItem>
            );
        else {
            return (
                <MenuItem
                    onClick={(e: React.MouseEvent<HTMLLIElement>) => {
                        onDisableUser(e);
                        setAnchorEl(null);
                    }}>
                    <ListItemIcon>
                        <FontAwesomeIcon icon={faUserLock} />
                    </ListItemIcon>
                    <ListItemText primary={'Disable User'} />
                </MenuItem>
            );
        }
    };

    /* Implementation */

    return (
        <div>
            <IconButton data-testid='manage-users_user-row-action-menu-button' onClick={handleOnOpen} size='large'>
                <FontAwesomeIcon icon={faBars} />
            </IconButton>
            <StyledMenu
                anchorEl={anchorEl}
                data-testid={`manage-users_user-row-action-menu-${index}`}
                keepMounted
                open={Boolean(anchorEl)}
                onClose={() => {
                    setAnchorEl(null);
                }}>
                <MenuItem
                    onClick={(e: React.MouseEvent<HTMLLIElement>) => {
                        onUpdateUser(e);
                        setAnchorEl(null);
                    }}>
                    <ListItemIcon>
                        <FontAwesomeIcon icon={faEdit} />
                    </ListItemIcon>
                    <ListItemText primary='Update User' />
                </MenuItem>

                {showPasswordOptions && (
                    <MenuItem
                        onClick={(e: React.MouseEvent<HTMLLIElement>) => {
                            onUpdateUserPassword(e);
                            setAnchorEl(null);
                        }}>
                        <ListItemIcon>
                            <FontAwesomeIcon icon={faKey} />
                        </ListItemIcon>
                        <ListItemText primary='Change Password' />
                    </MenuItem>
                )}

                {showPasswordOptions && showAuthMgmtButtons && (
                    <MenuItem
                        onClick={(e: React.MouseEvent<HTMLLIElement>) => {
                            onExpireUserPassword(e);
                            setAnchorEl(null);
                        }}>
                        <ListItemIcon>
                            <FontAwesomeIcon icon={faLock} />
                        </ListItemIcon>
                        <ListItemText primary='Force Password Reset' />
                    </MenuItem>
                )}

                <MenuItem
                    onClick={(e: React.MouseEvent<HTMLLIElement>) => {
                        onManageUserTokens(e);
                        setAnchorEl(null);
                    }}>
                    <ListItemIcon>
                        <FontAwesomeIcon icon={faCogs} />
                    </ListItemIcon>
                    <ListItemText primary='Generate / Revoke API Tokens' />
                </MenuItem>

                {showDisableMfaButton && (
                    <MenuItem
                        onClick={(e: React.MouseEvent<HTMLLIElement>) => {
                            onDisableUserMfa(e);
                            setAnchorEl(null);
                        }}>
                        <ListItemIcon>
                            <FontAwesomeIcon icon={faUnlockAlt} />
                        </ListItemIcon>
                        <ListItemText primary='Disable MFA' />
                    </MenuItem>
                )}

                {showAuthMgmtButtons && getAbleUserComponent()}

                {showAuthMgmtButtons && (
                    <MenuItem
                        onClick={(e: React.MouseEvent<HTMLLIElement>) => {
                            onDeleteUser(e);
                            setAnchorEl(null);
                        }}>
                        <ListItemIcon>
                            <FontAwesomeIcon icon={faTrash} />
                        </ListItemIcon>
                        <ListItemText primary='Delete User' />
                    </MenuItem>
                )}
            </StyledMenu>
        </div>
    );
};

export default UserActionsMenu;
