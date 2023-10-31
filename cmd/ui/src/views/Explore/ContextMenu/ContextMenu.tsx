import { Menu, MenuItem } from '@mui/material';
import { FC, useEffect, useState } from 'react';
import { useSelector } from 'react-redux';
import { setActiveTab, setSearchValue, startSearchAction } from 'src/ducks/searchbar/actions';
import { PRIMARY_SEARCH, SEARCH_TYPE_EXACT, SECONDARY_SEARCH } from 'src/ducks/searchbar/types';
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
        }
    };

    const handleSetEndingNode = () => {
        if (selectedNode) {
            dispatch(setActiveTab('secondary'));
            dispatch(setSearchValue(null, SECONDARY_SEARCH, SEARCH_TYPE_EXACT));
            dispatch(startSearchAction(selectedNode.name, SECONDARY_SEARCH));
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
