// Copyright 2024 Specter Ops, Inc.
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
    Skeleton,
    Paper,
    Table,
    TableBody,
    TableCell,
    TableContainer,
    TableHead,
    TableRow,
    TableSortLabel,
    IconButton,
    ListItemIcon,
    ListItemText,
    Menu,
    MenuItem,
    MenuProps,
    useTheme,
} from '@mui/material';
import { Button } from '@bloodhoundenterprise/doodleui';
import { faEllipsisVertical, faTrash } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import withStyles from '@mui/styles/withStyles';
import React, { useCallback } from 'react';
import { SSOProvider } from 'js-client-library';

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

const SSOProviderTableActionsMenu: React.FC<{
    onDeleteSSOProvider: () => void;
}> = ({ onDeleteSSOProvider }) => {
    /* Hooks */

    const [anchorEl, setAnchorEl] = React.useState<HTMLElement | null>(null);

    /* Event Handlers */

    const handleOnOpen: React.MouseEventHandler<HTMLButtonElement> = (event) => {
        setAnchorEl(event.currentTarget);
    };

    /* Implementation */

    const onClickDeleteSSOProvider = useCallback(() => {
        onDeleteSSOProvider();
        setAnchorEl(null);
    }, []);

    return (
        <>
            <IconButton onClick={handleOnOpen} size='small'>
                <FontAwesomeIcon icon={faEllipsisVertical} />
            </IconButton>
            <StyledMenu
                anchorEl={anchorEl}
                keepMounted
                open={Boolean(anchorEl)}
                onClose={() => {
                    setAnchorEl(null);
                }}>
                <MenuItem onClick={onClickDeleteSSOProvider}>
                    <ListItemIcon>
                        <FontAwesomeIcon icon={faTrash} />
                    </ListItemIcon>
                    <ListItemText primary='Delete SSO Provider' />
                </MenuItem>
            </StyledMenu>
        </>
    );
};

const SSOProviderTable: React.FC<{
    ssoProviders: SSOProvider[];
    loading: boolean;
    onDeleteSSOProvider: (ssoProviderId: SSOProvider['id']) => void;
    onClickSSOProvider: (ssoProviderId: SSOProvider['id']) => void;
    onToggleTypeSortOrder: () => void;
    typeSortOrder?: 'asc' | 'desc';
}> = ({ ssoProviders, loading, onDeleteSSOProvider, onClickSSOProvider, typeSortOrder, onToggleTypeSortOrder }) => {
    const theme = useTheme();
    return (
        <Paper>
            <TableContainer sx={{ maxHeight: 777 }}>
                <Table stickyHeader aria-label='sso provider table'>
                    <TableHead>
                        <TableRow>
                            <TableCell sx={{ backgroundColor: theme.palette.background.paper }} />
                            <TableCell
                                sx={{
                                    fontWeight: 'bold',
                                    verticalAlign: 'bottom',
                                    backgroundColor: theme.palette.background.paper,
                                }}>
                                Provider Name
                            </TableCell>
                            <TableCell sx={{ fontWeight: 'bold', backgroundColor: theme.palette.background.paper }}>
                                <TableSortLabel
                                    active={!!typeSortOrder}
                                    direction={typeSortOrder}
                                    onClick={onToggleTypeSortOrder}>
                                    Type
                                </TableSortLabel>
                            </TableCell>
                        </TableRow>
                    </TableHead>
                    <TableBody
                        sx={{
                            '& > :nth-of-type(odd)': {
                                backgroundColor: theme.palette.neutral.tertiary,
                                '&:hover': {
                                    backgroundColor: theme.palette.neutral.quaternary,
                                },
                            },
                            '& > :nth-of-type(even)': {
                                backgroundColor: theme.palette.neutral.secondary,
                                '&:hover': {
                                    backgroundColor: theme.palette.neutral.quaternary,
                                },
                            },
                        }}>
                        {loading ? (
                            <TableRow>
                                <TableCell>
                                    <Skeleton />
                                </TableCell>
                                <TableCell>
                                    <Skeleton />
                                </TableCell>
                                <TableCell>
                                    <Skeleton />
                                </TableCell>
                            </TableRow>
                        ) : ssoProviders.length === 0 ? (
                            <TableRow>
                                <TableCell colSpan={6} align='center'>
                                    No SSO Providers found
                                </TableCell>
                            </TableRow>
                        ) : (
                            ssoProviders.map((ssoProvider, i) => (
                                <TableRow key={i}>
                                    <TableCell align='center' padding='checkbox'>
                                        <SSOProviderTableActionsMenu
                                            onDeleteSSOProvider={() => onDeleteSSOProvider(ssoProvider.id)}
                                        />
                                    </TableCell>
                                    <TableCell onClick={() => onClickSSOProvider(ssoProvider.id)} size='small'>
                                        <Button variant='text' fontColor={'primary'} className='p-0'>
                                            {ssoProvider.name}
                                        </Button>
                                    </TableCell>
                                    <TableCell>{ssoProvider.type.toUpperCase()}</TableCell>
                                </TableRow>
                            ))
                        )}
                    </TableBody>
                </Table>
            </TableContainer>
        </Paper>
    );
};

export default SSOProviderTable;
