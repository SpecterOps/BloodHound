import { FC, MouseEvent } from 'react';

import { Popover, PopoverContent, PopoverTrigger } from '@bloodhoundenterprise/doodleui';
import { VerticalEllipsis } from '../AppIcon/Icons';
interface ListItemActionMenuProps {
    id?: number;
    deleteQuery: (id: number) => void;
    editQuery: (id: number) => void;
}

const ListItemActionMenu: FC<ListItemActionMenuProps> = ({ id, deleteQuery, editQuery }) => {
    const handleDelete = (event: MouseEvent) => {
        event.stopPropagation();
        deleteQuery(id as number);
    };

    const handleEdit = (event: MouseEvent) => {
        event.stopPropagation();
        editQuery(id as number);
    };

    const listItemStyles = 'w-full px-2 py-3 cursor-pointer hover:bg-neutral-light-4 dark:hover:bg-neutral-dark-4';

    return (
        <>
            <Popover>
                <PopoverTrigger
                    className='dark:text-white p-2 rounded rounded-full hover:bg-neutral-light-4 dark:hover:bg-neutral-dark-2'
                    onClick={(event) => event.stopPropagation()}>
                    <VerticalEllipsis size={24} />
                </PopoverTrigger>
                <PopoverContent className='p-0'>
                    <div className={listItemStyles}>Run</div>
                    <div className={listItemStyles} onClick={handleEdit}>
                        Edit/Share
                    </div>
                    <div className={listItemStyles} onClick={handleDelete}>
                        Delete
                    </div>
                </PopoverContent>
            </Popover>
        </>
    );
};

export default ListItemActionMenu;
