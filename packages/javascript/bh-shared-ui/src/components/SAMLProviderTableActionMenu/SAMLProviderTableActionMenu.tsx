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

import { faBars, faTrash } from '@fortawesome/free-solid-svg-icons';
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

interface SAMLProviderTableActionsMenuProps {
    SAMLProviderId: string;
    onDeleteSAMLProvider: (SAMLProviderId: string) => any;
}

const SAMLProviderTableActionsMenu: React.FC<SAMLProviderTableActionsMenuProps> = ({
    SAMLProviderId,
    onDeleteSAMLProvider,
}) => {
    /* Hooks */

    const [anchorEl, setAnchorEl] = React.useState<HTMLElement | null>(null);

    /* Event Handlers */

    const handleOnOpen: React.MouseEventHandler<HTMLButtonElement> = (event) => {
        setAnchorEl(event.currentTarget);
    };

    /* Implementation */

    return (
        <div>
            <IconButton onClick={handleOnOpen} size='large'>
                <FontAwesomeIcon icon={faBars} />
            </IconButton>
            <StyledMenu
                anchorEl={anchorEl}
                keepMounted
                open={Boolean(anchorEl)}
                onClose={() => {
                    setAnchorEl(null);
                }}>
                <MenuItem
                    onClick={() => {
                        onDeleteSAMLProvider(SAMLProviderId);
                        setAnchorEl(null);
                    }}>
                    <ListItemIcon>
                        <FontAwesomeIcon icon={faTrash} />
                    </ListItemIcon>
                    <ListItemText primary='Delete SAML Provider' />
                </MenuItem>
            </StyledMenu>
        </div>
    );
};

export default SAMLProviderTableActionsMenu;
