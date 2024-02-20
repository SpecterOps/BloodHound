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

import { faFilter } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { Button } from '@mui/material';
import makeStyles from '@mui/styles/makeStyles';
import { useEffect, useRef, useState } from 'react';
import EdgeFilteringDialog, { EdgeCheckboxType } from './EdgeFilteringDialog';
import { useAppDispatch, useAppSelector } from 'src/store';
import { pathfindingSearch, pathFiltersSaved } from 'src/ducks/searchbar/actions';

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

const EdgeFilter = () => {
    const classes = useStyles();
    const dispatch = useAppDispatch();

    const [isOpenDialog, setIsOpenDialog] = useState(false);
    const [isActiveFilters, setIsActiveFilters] = useState(false);

    const initialFilterState = useRef<EdgeCheckboxType[]>([]);
    const pathFilters = useAppSelector((state) => state.search.pathFilters);

    useEffect(() => {
        // if user has applied filters, set active
        if (pathFilters?.some((filter) => !filter.checked)) {
            setIsActiveFilters(true);
        } else {
            setIsActiveFilters(false);
        }
    }, [pathFilters]);

    const handlePathfindingSearch = () => {
        dispatch(pathfindingSearch());
    };

    return (
        <>
            <Button
                className={classes.pathfindingButton}
                variant='outlined'
                onClick={() => {
                    setIsOpenDialog(true);
                    // what is the initial state of edge filters?  save it
                    initialFilterState.current = pathFilters;
                }}>
                <FontAwesomeIcon icon={faFilter} color={isActiveFilters ? '#406F8E' : 'black'} />
            </Button>
            <EdgeFilteringDialog
                isOpen={isOpenDialog}
                handleCancel={() => {
                    setIsOpenDialog(false);

                    // rollback changes made in dialog.
                    dispatch(pathFiltersSaved(initialFilterState.current));
                }}
                handleApply={() => {
                    setIsOpenDialog(false);
                    handlePathfindingSearch();
                }}
            />
        </>
    );
};

export default EdgeFilter;
