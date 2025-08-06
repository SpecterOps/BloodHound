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
import { Box, Dialog, DialogActions, DialogContent, DialogContentText, DialogTitle, Typography } from '@mui/material';
import makeStyles from '@mui/styles/makeStyles';
import { groupBy } from 'lodash';
import { FC, useEffect, useRef, useState } from 'react';
import { QueryListSection } from '../../types';
import ListItemActionMenu from './ListItemActionMenu';
interface PrebuiltSearchListProps {
    listSections: QueryListSection[];
    selectedQuery: any;
    showCommonQueries: boolean;
    clickHandler: (query: string, id?: number) => void;
    deleteHandler?: (id: number) => void;
    editHandler: (id: number) => void;
    clearFiltersHandler: () => void;
}

const useStyles = makeStyles((theme) => ({
    subheader: {
        color: theme.palette.color.primary,
        backgroundColor: theme.palette.neutral.tertiary,
        paddingLeft: '8px',
        paddingRight: '8px',
        fontWeight: 'bold',
    },
    selected: {
        backgroundColor: theme.palette.neutral.quaternary,
        '&:hover': {
            backgroundColor: theme.palette.neutral.quaternary,
        },
    },
}));

const PrebuiltSearchList: FC<PrebuiltSearchListProps> = ({
    listSections,
    selectedQuery,
    showCommonQueries,
    clickHandler,
    deleteHandler,
    editHandler,
    clearFiltersHandler,
}) => {
    const [open, setOpen] = useState(false);
    const [queryId, setQueryId] = useState<number>();

    const styles = useStyles();
    const itemRef = useRef<HTMLDivElement>(null);

    const handleClose = () => {
        setOpen(false);
        setQueryId(undefined);
    };

    const groupedQueries = groupBy(listSections, 'category');
    const handleDelete = (id: number) => {
        setQueryId(id);
        setOpen(true);
    };

    const handleEdit = (id: number) => {
        editHandler(id);
    };

    const testMatch = (name: string, id?: number) => {
        if (!selectedQuery) return false;

        if (id && id === selectedQuery.id) {
            return true;
        } else if (name && name === selectedQuery.name) {
            return true;
        }

        return false;
    };

    const scrollSelectedQuery = () => {
        if (itemRef.current) {
            itemRef.current.scrollIntoView({ behavior: 'smooth', block: 'start' });
        }
    };

    useEffect(() => {
        if (selectedQuery) scrollSelectedQuery();
    }, [selectedQuery, showCommonQueries]);

    return (
        <>
            {listSections && (
                <div>
                    {Object.entries(groupedQueries).map(([category, queryData]) => (
                        <div key={category} className='relative'>
                            {!!queryData[0].queries.length && (
                                <div className={`${styles.subheader} sticky top-0 z-[1] py-2`}>{category}</div>
                            )}
                            {queryData.map((queryItem, i) => {
                                const { subheader, queries } = queryItem;
                                return (
                                    <li key={i} className='list-none'>
                                        {queries?.map((lineItem, idx) => {
                                            const { id, name, description, query, canEdit = false } = lineItem;
                                            return (
                                                <div
                                                    className={`p-2 rounded rounded-sm flex items-center w-full cursor-pointer hover:bg-neutral-light-3 dark:hover:bg-neutral-dark-3 justify-between pl-4 scroll-my-10 ${
                                                        testMatch(name, id) ? styles.selected : ''
                                                    }`}
                                                    key={`${id}-${idx}`}
                                                    onClick={() => clickHandler(query, id)}
                                                    ref={testMatch(name, id) ? itemRef : null}>
                                                    <div>
                                                        {name ? (
                                                            <p className='mb-0 leading-none'>{name}</p>
                                                        ) : (
                                                            <p className='mb-0 leading-none'>{description}</p>
                                                        )}

                                                        {category && <span className='text-xs italic'>{category}</span>}
                                                        {category && subheader && (
                                                            <span className='text-xs italic pr-1'>,</span>
                                                        )}
                                                        {subheader && (
                                                            <span className='text-xs italic'>{subheader}</span>
                                                        )}
                                                    </div>
                                                    {canEdit && (
                                                        <ListItemActionMenu
                                                            id={id as number}
                                                            deleteQuery={() => handleDelete(id as number)}
                                                            editQuery={() => handleEdit(id as number)}
                                                        />
                                                    )}
                                                </div>
                                            );
                                        })}
                                    </li>
                                );
                            })}
                        </div>
                    ))}
                </div>
            )}
            {!listSections.length && (
                <Box className='min-h-40 flex flex-col items-center justify-center'>
                    <Typography variant='h6' className='mb-2'>
                        No Results
                    </Typography>
                    <Button variant='primary' size='small' onClick={clearFiltersHandler}>
                        Reset Filters
                    </Button>
                </Box>
            )}

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
