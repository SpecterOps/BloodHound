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

import { faBars } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { TabContext, TabPanel } from '@mui/lab';
import {
    IconButton,
    List,
    Tab,
    Table,
    TableBody,
    TableCell,
    TableContainer,
    TableHead,
    TableRow,
    Tabs,
} from '@mui/material';
import createStyles from '@mui/styles/createStyles';
import makeStyles from '@mui/styles/makeStyles';
import withStyles from '@mui/styles/withStyles';
import React, { useState } from 'react';
import { ConfirmationDialog } from 'bh-shared-ui';
import { initGraph, startAssetGroupQuery } from 'src/ducks/explore/actions';
import { setAssetGroupEdit, setAssetGroupIndex } from 'src/ducks/global/actions';
import { setTierZeroSelection } from 'src/ducks/tierzero/actions';
import { useAppDispatch, useAppSelector } from 'src/store';
import ActionsMenu from './ActionsMenu';

const useStyles = makeStyles((theme) => ({
    buttons: {
        justifyContent: 'center',
        padding: '5px 0',
    },
    tableRow: {
        cursor: 'pointer',
        '&.Mui-selected, &.Mui-selected:hover': {
            backgroundColor: theme.palette.primary.light,
        },
    },
    tableRowSelected: {
        backgroundColor: 'blue',
    },
}));

enum MenuItems {
    Groups = 'Groups',
    Queries = 'Queries',
}

interface StyledTabProps {
    label: string;
    value: MenuItems;
}

const MenuTabPanel = withStyles(() =>
    createStyles({
        root: {
            padding: 0,
        },
    })
)(TabPanel);

const MenuTabs = withStyles(() =>
    createStyles({
        indicator: {
            backgroundColor: '#6798B9',
        },
    })
)(Tabs);

const MenuTab = withStyles(() =>
    createStyles({
        root: {
            maxWidth: 'none',
        },
        wrapper: {
            whiteSpace: 'nowrap',
            textTransform: 'none',
        },
    })
)((props: StyledTabProps) => <Tab {...props} />);

const MenuContainer: React.FC = () => {
    const styles = useStyles();
    const dispatch = useAppDispatch();
    const assetGroups = useAppSelector((state) => state.global.options.assetGroups);
    const domain = useAppSelector((state) => state.global.options.domain);

    const [selectedAssetGroup, setSelectedAssetGroup] = useState<null | any>(null);
    const [anchorEl, setAnchorEl] = React.useState<null | HTMLElement>(null);
    const [confirmDelete, setConfirmDelete] = useState<boolean>(false);

    const showMenu =
        (assetGroup: any) =>
        (e: React.MouseEvent<Element, MouseEvent>): boolean => {
            setSelectedAssetGroup(assetGroup);
            setAnchorEl(e.currentTarget as HTMLElement);
            return false;
        };

    const handleEdit = () => {
        dispatch(setAssetGroupEdit(selectedAssetGroup.id));
    };

    const handleDelete = () => {
        setConfirmDelete(true);
    };

    return (
        <TabContext value={MenuItems.Groups}>
            <ActionsMenu
                anchorEl={anchorEl}
                assetGroup={selectedAssetGroup}
                handleClose={() => {
                    setAnchorEl(null);
                }}
                handleEdit={handleEdit}
                handleDelete={handleDelete}
            />
            <MenuTabs value={MenuItems.Groups}>
                <MenuTab key={0} label='Asset Groups' value={MenuItems.Groups} />
            </MenuTabs>
            <MenuTabPanel value={MenuItems.Groups}>
                <TableContainer>
                    <Table size='small'>
                        <TableHead>
                            <TableRow>
                                <TableCell>{'Name'}</TableCell>
                                <TableCell>{'Members'}</TableCell>
                                <TableCell />
                            </TableRow>
                        </TableHead>
                        <TableBody>
                            {assetGroups.map((group, index) => (
                                <TableRow
                                    className={styles.tableRow}
                                    onClick={() => {
                                        setSelectedAssetGroup(group);
                                        dispatch(setAssetGroupIndex(index));
                                        dispatch(initGraph(true));
                                        dispatch(startAssetGroupQuery(group.id, domain?.id, domain?.type));
                                        dispatch(setTierZeroSelection(domain?.id || null, domain?.type || null));
                                    }}
                                    selected={selectedAssetGroup !== null && selectedAssetGroup.name === group.name}
                                    hover
                                    key={index}>
                                    <TableCell>Admin Tier Zero</TableCell>
                                    <TableCell>{group.member_count}</TableCell>
                                    <TableCell>
                                        <IconButton
                                            data-testid='explore_search_button-actions-menu'
                                            onClick={(e: React.MouseEvent<HTMLButtonElement, MouseEvent>) => {
                                                showMenu(group)(e);
                                            }}
                                            size='small'>
                                            <FontAwesomeIcon icon={faBars} size='sm' />
                                        </IconButton>
                                    </TableCell>
                                </TableRow>
                            ))}
                        </TableBody>
                    </Table>
                </TableContainer>
            </MenuTabPanel>

            <MenuTabPanel value={MenuItems.Queries}>
                <List></List>
            </MenuTabPanel>

            {confirmDelete && (
                <ConfirmationDialog
                    open={confirmDelete}
                    text={'Are you sure you want to delete this asset group?'}
                    title={'Delete Asset Group'}
                    onClose={(response) => {
                        if (response) {
                            // TODO: delete asset group.
                        }
                        setConfirmDelete(false);
                    }}
                />
            )}
        </TabContext>
    );
};

export default MenuContainer;
