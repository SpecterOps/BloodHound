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

import { Button } from '@bloodhoundenterprise/doodleui';
import { faEllipsisVertical, faTrash } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import {
    IconButton,
    ListItemIcon,
    ListItemText,
    Menu,
    MenuItem,
    MenuProps,
    Paper,
    Skeleton,
    Table,
    TableBody,
    TableCell,
    TableContainer,
    TableHead,
    TableRow,
    TableSortLabel,
    useTheme,
} from '@mui/material';
import withStyles from '@mui/styles/withStyles';
import { SSOProvider } from 'js-client-library';
import { FC, MouseEventHandler, useState } from 'react';
import { SortOrder } from '../../utils';

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

const SSOProviderTableActionsMenu: FC<{
    onDeleteSSOProvider: () => void;
}> = ({ onDeleteSSOProvider }) => {
    /* Hooks */

    const [anchorEl, setAnchorEl] = useState<HTMLElement | null>(null);

    /* Event Handlers */

    const handleOnOpen: MouseEventHandler<HTMLButtonElement> = (event) => {
        setAnchorEl(event.currentTarget);
    };

    /* Implementation */

    const onClickDeleteSSOProvider = () => {
        onDeleteSSOProvider();
        setAnchorEl(null);
    };

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

const SSOProviderTable: FC<{
    ssoProviders: SSOProvider[];
    loading: boolean;
    onDeleteSSOProvider: (ssoProviderId: SSOProvider['id']) => void;
    onClickSSOProvider: (ssoProviderId: SSOProvider['id']) => void;
    onToggleTypeSortOrder: () => void;
    typeSortOrder?: SortOrder;
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
                                {ssoProviders.length > 1 && !loading ? (
                                    <TableSortLabel
                                        active={!!typeSortOrder}
                                        direction={typeSortOrder}
                                        onClick={onToggleTypeSortOrder}>
                                        Type
                                    </TableSortLabel>
                                ) : (
                                    'Type'
                                )}
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
                                <TableCell padding='checkbox'>
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
                            ssoProviders.map((ssoProvider) => (
                                <TableRow key={ssoProvider.id}>
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
