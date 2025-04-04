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

import { Switch } from '@bloodhoundenterprise/doodleui';
import {
    faCircleHalfStroke,
    faCompass,
    faDownload,
    faQuestionCircle,
    faSignOutAlt,
    faUser,
    faUserShield,
} from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { Box, Divider, useTheme } from '@mui/material';
import ListItemIcon from '@mui/material/ListItemIcon';
import ListItemText from '@mui/material/ListItemText';
import Menu, { MenuProps } from '@mui/material/Menu';
import MenuItem from '@mui/material/MenuItem';
import withStyles from '@mui/styles/withStyles';
import { EnterpriseIcon, FeatureFlag } from 'bh-shared-ui';
import React from 'react';
import { useNavigate } from 'react-router-dom';
import { logout } from 'src/ducks/auth/authSlice';
import { setDarkMode } from 'src/ducks/global/actions.ts';
import * as routes from 'src/routes/constants';
import { useAppDispatch, useAppSelector } from 'src/store';

interface Props {
    anchorEl: null | HTMLElement;
    handleClose: () => void;
}

const StyledMenu = withStyles({
    paper: {
        border: '1px solid #d3d4d5',
    },
})((props: MenuProps) => (
    <Menu
        elevation={0}
        anchorOrigin={{
            vertical: 'bottom',
            horizontal: 'center',
        }}
        transformOrigin={{
            vertical: 'top',
            horizontal: 'center',
        }}
        {...props}
    />
));

const SettingsMenu: React.FC<Props> = ({ anchorEl, handleClose }) => {
    const dispatch = useAppDispatch();
    const navigate = useNavigate();
    const darkMode = useAppSelector((state) => state.global.view.darkMode);
    const theme = useTheme();

    const navigateTo = (route: string) => {
        handleClose();
        navigate(route);
    };

    const handleLogout: React.MouseEventHandler<HTMLLIElement> = () => {
        handleClose();
        dispatch(logout());
    };

    const toggleDarkMode: React.MouseEventHandler<HTMLLIElement> = () => {
        dispatch(setDarkMode(!darkMode));
    };

    const openInNewTab = (url: string) => {
        window.open(url, '_blank', 'noreferrer');
    };

    return (
        <div>
            <StyledMenu
                id='customized-menu'
                anchorEl={anchorEl}
                keepMounted
                open={Boolean(anchorEl)}
                onClose={handleClose}>
                <MenuItem
                    onClick={() => navigateTo(routes.ROUTE_MY_PROFILE)}
                    data-testid='global_header_settings-menu_nav-my-profile'>
                    <ListItemIcon>
                        <FontAwesomeIcon icon={faUser} />
                    </ListItemIcon>
                    <ListItemText primary='My Profile' />
                </MenuItem>

                <MenuItem
                    onClick={() => navigateTo(routes.ROUTE_DOWNLOAD_COLLECTORS)}
                    data-testid='global_header_settings-menu_nav-download-collectors'>
                    <ListItemIcon>
                        <FontAwesomeIcon icon={faDownload} />
                    </ListItemIcon>
                    <ListItemText primary='Download Collectors' />
                </MenuItem>

                <MenuItem
                    onClick={() => navigateTo(routes.ROUTE_ADMINISTRATION_ROOT)}
                    data-testid='global_header_settings-menu_nav-administration'>
                    <ListItemIcon>
                        <FontAwesomeIcon icon={faUserShield} />
                    </ListItemIcon>
                    <ListItemText primary='Administration' />
                </MenuItem>

                <MenuItem
                    onClick={() => handleClose()}
                    component='a'
                    href='https://bloodhound.specterops.io'
                    target='_blank'
                    rel='noreferrer'
                    data-testid='global_header_settings-menu_nav-support'>
                    <ListItemIcon>
                        <FontAwesomeIcon icon={faQuestionCircle} />
                    </ListItemIcon>
                    <ListItemText primary='Docs and Support' />
                </MenuItem>

                <MenuItem
                    onClick={() => navigateTo(routes.ROUTE_API_EXPLORER)}
                    data-testid='global_header_settings-menu_nav-api-explorer'>
                    <ListItemIcon>
                        <FontAwesomeIcon icon={faCompass} />
                    </ListItemIcon>
                    <ListItemText primary='API Explorer' />
                </MenuItem>

                <MenuItem
                    onClick={() => openInNewTab('https://bloodhoundenterprise.io/our-solution/')}
                    data-testid='global_header_settings-menu_nav-checkout-BHE'>
                    <ListItemIcon>
                        <EnterpriseIcon fill={theme.palette.color.primary} width='1rem' height='1rem' />
                    </ListItemIcon>
                    <ListItemText primary='BloodHound Enterprise' />
                </MenuItem>

                <FeatureFlag
                    flagKey='dark_mode'
                    errorFallback={null}
                    loadingFallback={null}
                    enabled={
                        <MenuItem onClick={toggleDarkMode} data-testid={'global_header_settings-menu_nav-logout'}>
                            <ListItemIcon>
                                <FontAwesomeIcon icon={faCircleHalfStroke} />
                            </ListItemIcon>
                            <ListItemText primary={'Dark Mode'} />
                            <Switch checked={darkMode}>Dark Mode</Switch>
                        </MenuItem>
                    }
                />

                <Box my={1}>
                    <Divider />
                </Box>

                <MenuItem onClick={handleLogout} data-testid='global_header_settings-menu_nav-logout'>
                    <ListItemIcon>
                        <FontAwesomeIcon icon={faSignOutAlt} />
                    </ListItemIcon>
                    <ListItemText primary='Logout' />
                </MenuItem>
            </StyledMenu>
        </div>
    );
};

export default SettingsMenu;
