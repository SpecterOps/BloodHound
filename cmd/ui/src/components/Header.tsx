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

import { faCog, faUsersRectangle, faProjectDiagram } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { AppBar, Box, IconButton, Link, Toolbar } from '@mui/material';
import makeStyles from '@mui/styles/makeStyles';
import React from 'react';
import { Link as RouterLink, useLocation, useNavigate } from 'react-router-dom';
import { MenuItem } from 'bh-shared-ui';
import * as routes from 'src/ducks/global/routes';
import SettingsMenu from 'src/components/SettingsMenu';

const useStyles = makeStyles((theme) => ({
    appBar: {
        height: '50px',
        transition: theme.transitions.create(['width', 'margin'], {
            easing: theme.transitions.easing.sharp,
            duration: theme.transitions.duration.leavingScreen,
        }),
        backgroundColor: theme.palette.background.paper,
        color: theme.palette.text.primary,
    },
    menu: {
        display: 'flex',
        marginLeft: theme.spacing(8),
        marginRight: theme.spacing(2),
        '@media print': {
            display: 'none',
        },
    },
    title: {
        display: 'flex',
        justifyContent: 'flex-end',
        alignItems: 'center',
        flexGrow: 1,
        marginRight: theme.spacing(5),
    },
    titleItem: {
        display: 'flex',
        alignItems: 'center',
        marginLeft: theme.spacing(2),
    },
    text: {
        fontSize: theme.typography.fontSize,
        fontWeight: 'normal',
        textAlign: 'right',
        whiteSpace: 'nowrap',
        textTransform: 'none',
        marginLeft: theme.spacing(1),
    },
    print: {
        flexGrow: 1,
        textAlign: 'right',
    },
    toolBar: {
        height: '100%',
    },
}));

const Header: React.FC = () => {
    const navigate = useNavigate();
    const classes = useStyles();

    const location = useLocation();
    const [anchorEl, setAnchorEl] = React.useState<null | HTMLElement>(null);

    const showMenu = (e: React.MouseEvent<Element, MouseEvent>): boolean => {
        setAnchorEl(e.currentTarget as HTMLElement);
        return false;
    };

    const renderMain = () => {
        return (
            <>
                <div className={classes.menu}>
                    <MenuItem
                        title={'Explore'}
                        icon={<FontAwesomeIcon icon={faProjectDiagram} size='sm' />}
                        active={location.pathname === routes.ROUTE_EXPLORE}
                        onClick={() => navigate(routes.ROUTE_EXPLORE)}
                        data-testid='global_header_nav-explore'
                    />
                    <MenuItem
                        title={'Group Management'}
                        icon={<FontAwesomeIcon icon={faUsersRectangle} size='sm' />}
                        active={location.pathname === routes.ROUTE_GROUP_MANAGEMENT}
                        onClick={() => navigate(routes.ROUTE_GROUP_MANAGEMENT)}
                        data-testid='global_header_nav-group-management'
                    />
                </div>

                <div className={classes.title}></div>

                <IconButton
                    onClick={showMenu}
                    className={'settings'}
                    style={{ color: '#1C222E' }}
                    size='small'
                    data-testid='global_header_settings-menu'
                    aria-label='Settings Menu'>
                    <FontAwesomeIcon icon={faCog} />
                </IconButton>

                <SettingsMenu
                    anchorEl={anchorEl}
                    handleClose={() => {
                        setAnchorEl(null);
                    }}
                />
            </>
        );
    };

    return (
        <AppBar position='static' className={classes.appBar} elevation={0} data-testid='global_header'>
            <Toolbar variant='dense' className={classes.toolBar}>
                <Box height='100%' paddingY='6px' boxSizing='border-box'>
                    <Link component={RouterLink} to={routes.ROUTE_HOME} data-testid='global_header_nav-home'>
                        <img
                            src={`${import.meta.env.BASE_URL}/img/logo-transparent-banner.svg`}
                            alt='BloodHound CE Home'
                            style={{
                                height: '100%',
                            }}
                        />
                    </Link>
                </Box>
                {renderMain()}
            </Toolbar>
        </AppBar>
    );
};

export default Header;
