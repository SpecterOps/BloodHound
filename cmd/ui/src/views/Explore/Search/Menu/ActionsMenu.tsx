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

import { faEdit, faUserTimes } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import ListItemIcon from '@mui/material/ListItemIcon';
import ListItemText from '@mui/material/ListItemText';
import Menu, { MenuProps } from '@mui/material/Menu';
import MenuItem from '@mui/material/MenuItem';
import withStyles from '@mui/styles/withStyles';
import makeStyles from '@mui/styles/makeStyles';
import React, { MouseEvent } from 'react';

interface Props {
    anchorEl: null | HTMLElement;
    assetGroup: any;
    handleClose: () => void;
    handleEdit: (e: any) => any;
    handleDelete: (e: any) => any;
}

const useStyles = makeStyles((theme) => {
    return {
        disabled: {
            color: '#ccc',
            '& .MuiListItemIcon-root': {
                color: '#ccc',
            },
            '&:hover': {
                backgroundColor: theme.palette.background.paper,
            },
        },
    };
});

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

const ActionsMenu: React.FC<Props> = ({ anchorEl, assetGroup, handleClose, handleEdit, handleDelete }) => {
    const styles = useStyles();
    return (
        <div>
            <StyledMenu
                id='customized-menu'
                anchorEl={anchorEl}
                keepMounted
                open={Boolean(anchorEl)}
                onClose={handleClose}>
                <MenuItem
                    onClick={(e: MouseEvent<HTMLLIElement>) => {
                        handleEdit(e);
                        handleClose();
                    }}>
                    <ListItemIcon>
                        <FontAwesomeIcon icon={faEdit} />
                    </ListItemIcon>
                    <ListItemText primary='Edit Asset Group' />
                </MenuItem>

                <MenuItem
                    className={assetGroup?.system_group ? styles.disabled : ''}
                    onClick={(e: MouseEvent<HTMLLIElement>) => {
                        if (!assetGroup.system_group) {
                            handleDelete(e);
                            handleClose();
                        }
                    }}>
                    <ListItemIcon>
                        <FontAwesomeIcon icon={faUserTimes} />
                    </ListItemIcon>
                    <ListItemText primary='Delete Asset Group' />
                </MenuItem>
            </StyledMenu>
        </div>
    );
};

export default ActionsMenu;
