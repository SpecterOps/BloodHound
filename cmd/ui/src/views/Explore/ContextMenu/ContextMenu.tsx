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

import { Menu, MenuItem } from '@mui/material';
import { FC, useEffect, useState } from 'react';
import { useSelector } from 'react-redux';
import {
    destinationNodeSuggested,
    setActiveTab,
    setSearchValue,
    sourceNodeSuggested,
    startSearchAction,
} from 'src/ducks/searchbar/actions';
import { PRIMARY_SEARCH, SEARCH_TYPE_EXACT, PATHFINDING_SEARCH } from 'src/ducks/searchbar/types';
import { AppState, useAppDispatch } from 'src/store';

const ContextMenu: FC<{ anchorPosition: { x: number; y: number } }> = ({ anchorPosition }) => {
    const dispatch = useAppDispatch();
    const [open, setOpen] = useState(false);

    const selectedNode = useSelector((state: AppState) => state.entityinfo.selectedNode);

    useEffect(() => {
        if (anchorPosition) {
            setOpen(true);
        } else {
            setOpen(false);
        }
    }, [anchorPosition]);

    const handleClick = () => {
        setOpen(false);
    };

    const handleSetStartingNode = () => {
        if (selectedNode) {
            dispatch(setActiveTab('secondary'));
            dispatch(setSearchValue(null, PRIMARY_SEARCH, SEARCH_TYPE_EXACT));
            dispatch(startSearchAction(selectedNode.name, PRIMARY_SEARCH));
            dispatch(sourceNodeSuggested());
        }
    };

    const handleSetEndingNode = () => {
        if (selectedNode) {
            dispatch(setActiveTab('secondary'));
            dispatch(setSearchValue(null, PATHFINDING_SEARCH, SEARCH_TYPE_EXACT));
            dispatch(startSearchAction(selectedNode.name, PATHFINDING_SEARCH));
            dispatch(destinationNodeSuggested());
        }
    };

    return (
        <Menu
            open={open}
            anchorPosition={{ left: anchorPosition?.x || 0, top: anchorPosition?.y || 0 }}
            anchorReference='anchorPosition'
            onClick={handleClick}>
            <MenuItem onClick={handleSetStartingNode}>Set as starting node</MenuItem>
            <MenuItem onClick={handleSetEndingNode}>Set as ending node</MenuItem>
        </Menu>
    );
};

export default ContextMenu;
