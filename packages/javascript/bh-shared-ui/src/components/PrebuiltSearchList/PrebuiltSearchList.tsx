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

import { Button } from '@bloodhoundenterprise/doodleui';
import { faTrash } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import {
    Box,
    Dialog,
    DialogActions,
    DialogContent,
    DialogContentText,
    DialogTitle,
    List,
    ListItem,
    ListItemButton,
    ListItemText,
    ListSubheader,
} from '@mui/material';
import { FC, useState } from 'react';
import makeStyles from '@mui/styles/makeStyles';

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

const useStyles = makeStyles((theme) => ({
    subheader: {
        color: theme.palette.color.primary,
        backgroundColor: theme.palette.neutral.tertiary,
        borderRadius: '8px',
        fontWeight: 'bold',
    },
}));

const PrebuiltSearchList: FC<PrebuiltSearchListProps> = ({ listSections, clickHandler, deleteHandler }) => {
    const [open, setOpen] = useState(false);
    const [queryId, setQueryId] = useState<number>();
    const styles = useStyles();

    const handleOpen = () => {
        setOpen(true);
    };

    const handleClose = () => {
        setOpen(false);
        setQueryId(undefined);
    };

    return (
        <>
            <List dense disablePadding>
                {listSections.map((section) => {
                    const { subheader, lineItems } = section;

                    return (
                        <Box key={subheader}>
                            <ListSubheader className={styles.subheader}>{subheader} </ListSubheader>

                            {lineItems?.map((lineItem, idx) => {
                                const { id, description, cypher, canEdit = false } = lineItem;

                                return (
                                    <ListItem
                                        disablePadding
                                        key={`${id}-${idx}`}
                                        sx={{ borderRadius: '8px', py: '4px' }}
                                        secondaryAction={
                                            canEdit && (
                                                <Button
                                                    aria-label='Delete Query'
                                                    size='small'
                                                    variant='secondary'
                                                    onClick={() => {
                                                        setQueryId(id);
                                                        handleOpen();
                                                    }}>
                                                    <FontAwesomeIcon icon={faTrash} />
                                                </Button>
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

            <Dialog open={open} onClose={handleClose} maxWidth={'xs'} fullWidth>
                <DialogTitle>Delete Query</DialogTitle>
                <DialogContent>
                    <DialogContentText>Are you sure you want to delete this query?</DialogContentText>
                </DialogContent>
                <DialogActions>
                    <Button variant='tertiary' onClick={handleClose}>
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
