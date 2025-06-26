import { ListItemIcon, Menu, MenuItem } from '@mui/material';
import { FC, MouseEvent, useState } from 'react';

import { VerticalEllipsis } from '../AppIcon/Icons';

interface ListItemActionMenuProps {
    canEdit: boolean;
    id?: number;
    deleteQuery: (id: number) => void;
}

const ListItemActionMenu: FC<ListItemActionMenuProps> = ({ canEdit, id, deleteQuery }) => {
    const [anchorEl, setAnchorEl] = useState<null | HTMLElement>(null);
    const open = Boolean(anchorEl);
    const handleClick = (event: MouseEvent<HTMLDivElement>) => {
        event.stopPropagation();
        setAnchorEl(event.currentTarget);
    };
    const handleClose = (event: MouseEvent) => {
        // event.stopPropagation();
        setAnchorEl(null);
    };

    const handleRun = (event: MouseEvent) => {
        handleClose(event);
    };

    const handleDelete = (event: MouseEvent) => {
        // handleClose(event);
        event.stopPropagation();
        deleteQuery(id as number);
        handleClose(event);
    };

    return (
        <>
            <ListItemIcon onClick={handleClick} className='min-w-8'>
                <VerticalEllipsis />
            </ListItemIcon>

            <Menu id='basic-menu' anchorEl={anchorEl} open={open} onClose={handleClose}>
                <MenuItem onClick={handleRun}>Run</MenuItem>
                {canEdit && <MenuItem onClick={handleClose}>Edit/Share</MenuItem>}
                {canEdit && <MenuItem onClick={handleDelete}>Delete</MenuItem>}
            </Menu>
        </>
    );
};

export default ListItemActionMenu;
