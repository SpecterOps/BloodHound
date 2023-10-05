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

import { faTrash } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import {
    Box,
    Button,
    Dialog,
    DialogActions,
    DialogTitle,
    IconButton,
    List,
    ListItem,
    ListItemButton,
    ListItemText,
    ListSubheader,
} from '@mui/material';
import { FC, useState } from 'react';
import makeStyles from '@mui/styles/makeStyles';

const useStyles = makeStyles((theme) => ({
    list: {
        position: 'relative',
        overflow: 'auto',
        maxHeight: 300,
        '& ul': { padding: 0 },
    },
}));

interface PrebuiltSearchListProps {
    listSections: ListSection[];
    clickHandler: (query: string) => void;

    deleteHandler?: any;
}

type ListSection = {
    subheader: string;
    lineItems: LineItem[];
};

export type LineItem = {
    id?: number;

    description: string;
    cypher: string;
    canEdit?: boolean;
};

const PrebuiltSearchList: FC<PrebuiltSearchListProps> = ({ listSections, clickHandler, deleteHandler }) => {
    const classes = useStyles();

    const [open, setOpen] = useState(false);
    const [queryId, setQueryId] = useState<number>();

    const handleOpen = () => {
        setOpen(true);
    };

    const handleClose = () => {
        setOpen(false);
        setQueryId(undefined);
    };

    return (
        <>
            <List dense disablePadding className={classes.list}>
                {listSections.map((section) => {
                    const { subheader, lineItems } = section;

                    return (
                        <Box key={subheader}>
                            <ListSubheader sx={{ fontWeight: 'bold' }}>{subheader} </ListSubheader>

                            {lineItems?.map((lineItem, idx) => {
                                const { id, description, cypher, canEdit = false } = lineItem;

                                return (
                                    <ListItem
                                        disablePadding
                                        key={`${id}-${idx}`}
                                        secondaryAction={
                                            canEdit && (
                                                <IconButton
                                                    size='small'
                                                    onClick={() => {
                                                        setQueryId(id);
                                                        handleOpen();
                                                    }}>
                                                    <FontAwesomeIcon icon={faTrash} />
                                                </IconButton>
                                            )
                                        }>
                                        <ListItemButton onClick={() => clickHandler(cypher)}>
                                            <ListItemText primary={description} />
                                        </ListItemButton>
                                    </ListItem>
                                );
                            })}
                        </Box>
                    );
                })}
            </List>

            <Dialog
                open={open}
                onClose={handleClose}
                maxWidth={'xs'}
                fullWidth
                aria-labelledby='alert-delete-query-confirmation'>
                <DialogTitle>Delete this query?</DialogTitle>
                <DialogActions>
                    <Button onClick={handleClose}>Cancel</Button>
                    <Button
                        onClick={() => {
                            deleteHandler(queryId);
                            handleClose();
                        }}
                        autoFocus>
                        Confirm
                    </Button>
                </DialogActions>
            </Dialog>
        </>
    );
};

export default PrebuiltSearchList;
