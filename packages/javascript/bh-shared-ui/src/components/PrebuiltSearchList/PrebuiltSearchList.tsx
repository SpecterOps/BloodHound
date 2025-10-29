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
import { Box, Typography } from '@mui/material';
import makeStyles from '@mui/styles/makeStyles';
import groupBy from 'lodash/groupBy';
import { FC, useEffect, useRef } from 'react';
import { QueryListSection } from '../../types';
import { adaptClickHandlerToKeyDown } from '../../utils/AccessibleClickableDiv';
import { useSavedQueriesContext } from '../../views/Explore/providers/SavedQueriesProvider';
import ListItemActionMenu from './ListItemActionMenu';
interface PrebuiltSearchListProps {
    listSections: QueryListSection[];
    showCommonQueries: boolean;
    clickHandler: (query: string, id?: number) => void;
    deleteHandler?: (id: number) => void;
    clearFiltersHandler: () => void;
}

const useStyles = makeStyles((theme) => ({
    subheader: {
        color: theme.palette?.color.primary,
        backgroundColor: theme.palette?.neutral.tertiary,
        paddingLeft: '8px',
        paddingRight: '8px',
        fontWeight: 'bold',
    },
    selected: {
        backgroundColor: theme.palette?.neutral.quaternary,
        '&:hover': {
            backgroundColor: theme.palette?.neutral.quaternary,
        },
    },
}));

const PrebuiltSearchList: FC<PrebuiltSearchListProps> = ({
    listSections,
    showCommonQueries,
    clickHandler,
    deleteHandler,
    clearFiltersHandler,
}) => {
    const { selectedQuery } = useSavedQueriesContext();
    const styles = useStyles();
    const itemRef = useRef<HTMLLIElement>(null);
    const groupedQueries = groupBy(listSections, 'category');

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
                <div data-testid='list-sections'>
                    {Object.entries(groupedQueries).map(([category, queryData]) => (
                        <div key={category} className='relative'>
                            {category && !!queryData[0].queries.length && (
                                <div className={`${styles.subheader} sticky top-0 z-[1] py-2`}>{category}</div>
                            )}
                            {queryData.map((queryItem, i) => {
                                const { subheader, queries } = queryItem;
                                return (
                                    <ul key={i} className='list-none'>
                                        {queries?.map((lineItem, idx) => {
                                            const { id, name, description, query, canEdit = false } = lineItem;
                                            return (
                                                <li
                                                    className={`p-2 rounded rounded-sm flex items-center w-full cursor-pointer hover:bg-neutral-light-3 dark:hover:bg-neutral-dark-3 justify-between pl-4 scroll-my-10 list-none ${
                                                        testMatch(name, id) ? styles.selected : ''
                                                    }`}
                                                    key={`${id}-${idx}`}
                                                    ref={testMatch(name, id) ? itemRef : null}>
                                                    <div
                                                        role='button'
                                                        tabIndex={0}
                                                        key={`${id}-${idx}`}
                                                        onClick={() => clickHandler(query, id)}
                                                        onKeyDown={(e) =>
                                                            adaptClickHandlerToKeyDown(e, () => clickHandler(query, id))
                                                        }>
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
                                                    {canEdit && typeof id === 'number' && (
                                                        <ListItemActionMenu
                                                            id={id}
                                                            query={query}
                                                            deleteQuery={() => {
                                                                if (deleteHandler) deleteHandler(id);
                                                            }}
                                                        />
                                                    )}
                                                </li>
                                            );
                                        })}
                                    </ul>
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
        </>
    );
};

export default PrebuiltSearchList;
