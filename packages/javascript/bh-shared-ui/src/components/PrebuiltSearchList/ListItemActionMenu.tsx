import { FC, MouseEvent } from 'react';

import { Popover, PopoverContent, PopoverTrigger } from '@bloodhoundenterprise/doodleui';
import { useSavedQueriesContext } from '../../views';
import { VerticalEllipsis } from '../AppIcon/Icons';
interface ListItemActionMenuProps {
    id?: number;
    query?: string;
    deleteQuery: (id: number) => void;
}

const ListItemActionMenu: FC<ListItemActionMenuProps> = ({ id, query, deleteQuery }) => {
    const { runQuery, editQuery } = useSavedQueriesContext();

    const handleRun = (event: MouseEvent) => {
        event.stopPropagation();
        runQuery(query as string, id as number);
    };

    const handleEdit = (event: MouseEvent) => {
        event.stopPropagation();
        editQuery(id as number);
    };

    const handleDelete = (event: MouseEvent) => {
        event.stopPropagation();
        deleteQuery(id as number);
    };

    //on click run - event bubbles up the line item click event which fires Run

    const listItemStyles = 'w-full px-2 py-3 cursor-pointer hover:bg-neutral-light-4 dark:hover:bg-neutral-dark-4';

    return (
        <>
            <Popover>
                <PopoverTrigger
                    data-testid='saved-query-action-menu-trigger'
                    className='dark:text-white p-2 rounded rounded-full hover:bg-neutral-light-4 dark:hover:bg-neutral-dark-2'
                    onClick={(event) => event.stopPropagation()}>
                    <VerticalEllipsis size={24} />
                </PopoverTrigger>
                <PopoverContent className='p-0' data-testid='saved-query-action-menu'>
                    <div className={listItemStyles} onClick={handleRun}>
                        Run
                    </div>
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
