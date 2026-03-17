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

import { faEdit, faEllipsisVertical, faTrash } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import {
    Menu,
    MenuItem,
    Skeleton,
    Table,
    TableBody,
    TableCell,
    TableContainer,
    TableHead,
    TableRow,
    TableSortLabel,
} from '@mui/material';
import { Button } from 'doodle-ui';
import { SSOProvider } from 'js-client-library';
import { FC, MouseEventHandler, useState } from 'react';
import { usePermissions } from '../../hooks';
import { SortOrder } from '../../types';
import { Permission } from '../../utils';

const SSOProviderTableActionsMenu: FC<{
    onDeleteSSOProvider: () => void;
    onUpdateSSOProvider: () => void;
}> = ({ onDeleteSSOProvider, onUpdateSSOProvider }) => {
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

    const onClickUpdateSSOProvider = () => {
        onUpdateSSOProvider();
        setAnchorEl(null);
    };

    return (
        <>
            <Button variant={'text'} onClick={handleOnOpen} size='small' aria-label='Open provider actions menu'>
                <FontAwesomeIcon icon={faEllipsisVertical} />
            </Button>
            <Menu
                anchorEl={anchorEl}
                elevation={0}
                anchorOrigin={{
                    vertical: 'bottom',
                    horizontal: 'right',
                }}
                transformOrigin={{
                    vertical: 'top',
                    horizontal: 'right',
                }}
                classes={{ paper: 'border border-gray-300' }}
                keepMounted
                open={Boolean(anchorEl)}
                onClose={() => {
                    setAnchorEl(null);
                }}>
                <MenuItem onClick={onClickDeleteSSOProvider} className='flex gap-2'>
                    <FontAwesomeIcon icon={faTrash} className='text-gray-500' />
                    <span>Delete SSO Provider</span>
                </MenuItem>
                <MenuItem onClick={onClickUpdateSSOProvider} className='flex gap-2'>
                    <FontAwesomeIcon icon={faEdit} className='text-gray-500' />
                    <span>Edit SSO Provider</span>
                </MenuItem>
            </Menu>
        </>
    );
};

const SSOProviderTable: FC<{
    ssoProviders: SSOProvider[];
    loading: boolean;
    onDeleteSSOProvider: (ssoProvider: SSOProvider) => void;
    onUpdateSSOProvider: (ssoProvider: SSOProvider) => void;
    onClickSSOProvider: (ssoProviderId: SSOProvider['id']) => void;
    onToggleTypeSortOrder: () => void;
    typeSortOrder?: SortOrder;
}> = ({
    ssoProviders,
    loading,
    onDeleteSSOProvider,
    onUpdateSSOProvider,
    onClickSSOProvider,
    onToggleTypeSortOrder,
    typeSortOrder,
}) => {
    const { checkPermission } = usePermissions();
    const hasPermission = checkPermission(Permission.AUTH_MANAGE_PROVIDERS);

    return (
        <TableContainer className='max-h-[777px]'>
            <Table stickyHeader aria-label='sso provider table'>
                <TableHead>
                    <TableRow className='font-bold *:bg-neutral-2'>
                        <TableCell />
                        <TableCell className='align-bottom'>Provider Name</TableCell>
                        <TableCell>
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
                <TableBody>
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
                    ) : ssoProviders.length === 0 && hasPermission ? (
                        <TableRow>
                            <TableCell colSpan={3} align='center'>
                                No SSO Providers found
                            </TableCell>
                        </TableRow>
                    ) : (
                        ssoProviders.map((ssoProvider) => (
                            <TableRow key={ssoProvider.id} className='odd:bg-neutral-3'>
                                <TableCell align='center' padding='checkbox'>
                                    <SSOProviderTableActionsMenu
                                        onDeleteSSOProvider={() => onDeleteSSOProvider(ssoProvider)}
                                        onUpdateSSOProvider={() => onUpdateSSOProvider(ssoProvider)}
                                    />
                                </TableCell>
                                <TableCell size='small'>
                                    <Button
                                        variant='text'
                                        fontColor={'primary'}
                                        className='p-0'
                                        onClick={(e) => {
                                            e.stopPropagation();
                                            onClickSSOProvider(ssoProvider.id);
                                        }}>
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
    );
};

export default SSOProviderTable;
