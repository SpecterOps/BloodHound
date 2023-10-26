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
    DialogContent,
    DialogContentText,
    DialogTitle,
    IconButton,
    List,
    ListItem,
    ListItemButton,
    ListItemText,
    ListSubheader,
    SvgIcon,
} from '@mui/material';
import { FC, useState } from 'react';

interface PrebuiltSearchListProps {
    listSections: ListSection[];
    clickHandler: (query: string) => void;
    deleteHandler?: (id: number) => void;
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
            <Box maxHeight={'300px'} overflow={'auto'}>
                <List dense disablePadding>
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
                                                        aria-label='Delete Query'
                                                        size='small'
                                                        onClick={() => {
                                                            setQueryId(id);
                                                            handleOpen();
                                                        }}>
                                                        <SvgIcon fontSize='small'>
                                                            <FontAwesomeIcon icon={faTrash} />
                                                        </SvgIcon>
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
            </Box>

            <Dialog open={open} onClose={handleClose} maxWidth={'xs'} fullWidth>
                <DialogTitle>Delete Query</DialogTitle>
                <DialogContent>
                    <DialogContentText>Are you sure you want to delete this query?</DialogContentText>
                </DialogContent>
                <DialogActions>
                    <Button color='inherit' onClick={handleClose}>
                        Cancel
                    </Button>
                    <Button
                        onClick={() => {
                            if (deleteHandler) deleteHandler(queryId!);
                            handleClose();
                        }}
                        color='primary'
                        autoFocus>
                        Confirm
                    </Button>
                </DialogActions>
            </Dialog>
        </>
    );
};

export default PrebuiltSearchList;
