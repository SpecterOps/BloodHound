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

import { Box, Button } from '@mui/material';
import makeStyles from '@mui/styles/makeStyles';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { faBullseye, faCircle, faExchangeAlt, faFilter } from '@fortawesome/free-solid-svg-icons';
import { savePathFilters, setSearchValue, startSearchSelected } from 'src/ducks/searchbar/actions';
import { useDispatch, useSelector } from 'react-redux';
import { useCallback, useEffect, useRef, useState } from 'react';
import { PRIMARY_SEARCH, SEARCH_TYPE_EXACT, SECONDARY_SEARCH, SearchNodeType } from 'src/ducks/searchbar/types';
import NodeSearch from './NodeSearch';
import { AppState } from 'src/store';
import EdgeFilteringDialog, { EdgeCheckboxType } from './EdgeFilteringDialog';

const useStyles = makeStyles((theme) => ({
    pathfindingButton: {
        height: '25px',
        width: '25px',
        minWidth: '25px',
        borderRadius: theme.shape.borderRadius,
        borderColor: 'rgba(0,0,0,0.23)',
        color: 'black',
        padding: 0,
    },
}));

const PathfindingSearch = () => {
    const classes = useStyles();

    const dispatch = useDispatch();

    const [isOpenDialog, setIsOpenDialog] = useState(false);
    const [isActiveFilters, setIsActiveFilters] = useState(false);

    const searchState = useSelector((state: AppState) => state.search);

    const setSourceNode = useCallback(
        (newSource: SearchNodeType | null) => dispatch(setSearchValue(newSource, PRIMARY_SEARCH, SEARCH_TYPE_EXACT)),
        [dispatch]
    );

    const setDestinationNode = useCallback(
        (newDest: SearchNodeType | null) => dispatch(setSearchValue(newDest, SECONDARY_SEARCH, SEARCH_TYPE_EXACT)),
        [dispatch]
    );

    const executeSearch = useCallback(() => dispatch(startSearchSelected(SECONDARY_SEARCH)), [dispatch]);

    useEffect(() => {
        // if user has applied filters, set active
        if (searchState.pathFilters?.some((filter) => !filter.checked)) {
            setIsActiveFilters(true);
        } else {
            setIsActiveFilters(false);
        }
    }, [searchState.pathFilters]);

    const swapPathfindingInputs = useCallback(() => {
        const newSourceItem = searchState.secondary.value;
        const newDestinationItem = searchState.primary.value;

        setSourceNode(newSourceItem);
        setDestinationNode(newDestinationItem);
        executeSearch();
    }, [searchState, setSourceNode, setDestinationNode, executeSearch]);

    const doPathfindingSearch = () => {
        dispatch(startSearchSelected(SECONDARY_SEARCH));
    };

    const initialFilterState = useRef<EdgeCheckboxType[]>([]);

    return (
        <Box display={'flex'} alignItems={'center'} gap={1}>
            <SourceToBullseyeIcon />

            <Box flexGrow={1} gap={1} display={'flex'} flexDirection={'column'}>
                <NodeSearch searchType={PRIMARY_SEARCH} labelText='Start Node' />
                <NodeSearch searchType={SECONDARY_SEARCH} labelText='Destination Node' />
            </Box>

            <Button
                className={classes.pathfindingButton}
                variant='outlined'
                disabled={!searchState.primary.value || !searchState.secondary.value}
                onClick={() => swapPathfindingInputs()}>
                <FontAwesomeIcon icon={faExchangeAlt} className='fa-rotate-90' />
            </Button>

            <Button
                className={classes.pathfindingButton}
                variant='outlined'
                onClick={() => {
                    setIsOpenDialog(true);
                    // what is the initial state of edge filters?  save it
                    initialFilterState.current = searchState.pathFilters;
                }}>
                <FontAwesomeIcon icon={faFilter} color={isActiveFilters ? '#406F8E' : 'black'} />
            </Button>

            <EdgeFilteringDialog
                isOpen={isOpenDialog}
                handleCancel={() => {
                    setIsOpenDialog(false);

                    // rollback changes made in dialog.
                    dispatch(savePathFilters(initialFilterState.current));
                }}
                handleApply={() => {
                    setIsOpenDialog(false);
                    doPathfindingSearch();
                }}
            />
        </Box>
    );
};

const SourceToBullseyeIcon = () => {
    return (
        <Box display={'flex'} flexDirection={'column'} alignItems={'center'}>
            <FontAwesomeIcon icon={faCircle} size='xs' />
            <Box
                border={'none'}
                borderLeft={'1px dotted black'}
                marginTop={'0.5em'}
                marginBottom={'0.5em'}
                height='1em'></Box>
            <FontAwesomeIcon icon={faBullseye} size='xs' />
        </Box>
    );
};

export default PathfindingSearch;
